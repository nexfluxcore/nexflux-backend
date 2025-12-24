package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	MQTTClient mqtt.Client
	once       sync.Once
	mu         sync.RWMutex
	handlers   = make(map[string][]MessageHandler)
)

// MessageHandler is a function type for handling MQTT messages
type MessageHandler func(topic string, payload []byte)

// Config holds MQTT configuration
type Config struct {
	Broker   string
	Port     string
	ClientID string
	Username string
	Password string
}

// GetConfig returns MQTT configuration from environment
func GetConfig() Config {
	return Config{
		Broker:   getEnv("MQTT_BROKER", "localhost"),
		Port:     getEnv("MQTT_PORT", "1883"),
		ClientID: getEnv("MQTT_CLIENT_ID", "nexflux-backend"),
		Username: os.Getenv("MQTT_USERNAME"),
		Password: os.Getenv("MQTT_PASSWORD"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitMQTT initializes the MQTT client connection to VerneMQ
func InitMQTT() error {
	var initErr error
	once.Do(func() {
		config := GetConfig()
		broker := fmt.Sprintf("tcp://%s:%s", config.Broker, config.Port)

		opts := mqtt.NewClientOptions()
		opts.AddBroker(broker)
		opts.SetClientID(config.ClientID)
		opts.SetUsername(config.Username)
		opts.SetPassword(config.Password)
		opts.SetAutoReconnect(true)
		opts.SetConnectRetry(true)
		opts.SetConnectRetryInterval(5 * time.Second)
		opts.SetKeepAlive(30 * time.Second)
		opts.SetPingTimeout(10 * time.Second)
		opts.SetCleanSession(false)
		opts.SetOrderMatters(false)

		// Connection handlers
		opts.SetOnConnectHandler(func(client mqtt.Client) {
			log.Printf("📡 Connected to VerneMQ broker at %s", broker)
			// Re-subscribe to all topics
			resubscribeAll()
		})

		opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("❌ Lost connection to VerneMQ: %v", err)
		})

		opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			log.Printf("🔄 Reconnecting to VerneMQ...")
		})

		// Default message handler
		opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
			handleMessage(msg.Topic(), msg.Payload())
		})

		MQTTClient = mqtt.NewClient(opts)
		if token := MQTTClient.Connect(); token.WaitTimeout(5*time.Second) && token.Error() != nil {
			initErr = token.Error()
			log.Printf("❌ Failed to connect to VerneMQ: %v", initErr)
			return
		} else if token.Error() == nil && !IsConnected() {
			initErr = fmt.Errorf("connection timed out")
			log.Printf("❌ Failed to connect to VerneMQ: %v", initErr)
			return
		}

		log.Printf("✅ MQTT client initialized successfully")
	})

	return initErr
}

// IsConnected checks if MQTT client is connected
func IsConnected() bool {
	if MQTTClient == nil {
		return false
	}
	return MQTTClient.IsConnected()
}

// handleMessage routes incoming messages to registered handlers
func handleMessage(topic string, payload []byte) {
	mu.RLock()
	defer mu.RUnlock()

	for pattern, handlerList := range handlers {
		if matchTopic(pattern, topic) {
			for _, handler := range handlerList {
				go handler(topic, payload)
			}
		}
	}
}

// matchTopic checks if a topic matches a pattern (supports + and # wildcards)
func matchTopic(pattern, topic string) bool {
	// Simple implementation - for production, use proper MQTT topic matching
	if pattern == topic {
		return true
	}
	// TODO: Implement full MQTT wildcard matching
	return false
}

// resubscribeAll re-subscribes to all registered topics
func resubscribeAll() {
	mu.RLock()
	defer mu.RUnlock()

	for topic := range handlers {
		if token := MQTTClient.Subscribe(topic, 1, nil); token.Wait() && token.Error() != nil {
			log.Printf("❌ Failed to resubscribe to %s: %v", topic, token.Error())
		}
	}
}

// Subscribe subscribes to a topic with a message handler
func Subscribe(topic string, qos byte, handler MessageHandler) error {
	if MQTTClient == nil || !MQTTClient.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	mu.Lock()
	handlers[topic] = append(handlers[topic], handler)
	mu.Unlock()

	token := MQTTClient.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		handleMessage(msg.Topic(), msg.Payload())
	})

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	log.Printf("📥 Subscribed to topic: %s", topic)
	return nil
}

// Unsubscribe unsubscribes from a topic
func Unsubscribe(topic string) error {
	if MQTTClient == nil || !MQTTClient.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	token := MQTTClient.Unsubscribe(topic)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	mu.Lock()
	delete(handlers, topic)
	mu.Unlock()

	log.Printf("📤 Unsubscribed from topic: %s", topic)
	return nil
}

// Publish publishes a message to a topic
func Publish(topic string, qos byte, retained bool, payload interface{}) error {
	if MQTTClient == nil || !MQTTClient.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	var data []byte
	var err error

	switch v := payload.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %v", err)
		}
	}

	token := MQTTClient.Publish(topic, qos, retained, data)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// PublishJSON publishes a JSON message to a topic
func PublishJSON(topic string, qos byte, retained bool, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	return Publish(topic, qos, retained, data)
}

// Disconnect disconnects the MQTT client
func Disconnect() {
	if MQTTClient != nil && MQTTClient.IsConnected() {
		MQTTClient.Disconnect(250)
		log.Printf("🔌 Disconnected from VerneMQ")
	}
}

// ============================================
// Lab-specific Topic Helpers
// ============================================

// Topic patterns for lab hardware communication
const (
	TopicLabSensors     = "lab/%s/sensors"     // lab/{lab_id}/sensors
	TopicLabActuators   = "lab/%s/actuators"   // lab/{lab_id}/actuators
	TopicLabCode        = "lab/%s/code"        // lab/{lab_id}/code
	TopicLabCompilation = "lab/%s/compilation" // lab/{lab_id}/compilation
	TopicLabHeartbeat   = "lab/%s/heartbeat"   // lab/{lab_id}/heartbeat
	TopicLabStatus      = "lab/%s/status"      // lab/{lab_id}/status
	TopicLabControl     = "lab/%s/control"     // lab/{lab_id}/control
	TopicLabSerial      = "lab/%s/serial"      // lab/{lab_id}/serial
)

// BuildLabTopic builds a lab-specific topic
func BuildLabTopic(pattern, labID string) string {
	return fmt.Sprintf(pattern, labID)
}

// SubscribeToLabSensors subscribes to sensor data from a lab
func SubscribeToLabSensors(labID string, handler MessageHandler) error {
	topic := BuildLabTopic(TopicLabSensors, labID)
	return Subscribe(topic, 1, handler)
}

// SubscribeToLabHeartbeat subscribes to heartbeat from a lab agent
func SubscribeToLabHeartbeat(labID string, handler MessageHandler) error {
	topic := BuildLabTopic(TopicLabHeartbeat, labID)
	return Subscribe(topic, 1, handler)
}

// SubscribeToLabCompilation subscribes to compilation results from a lab
func SubscribeToLabCompilation(labID string, handler MessageHandler) error {
	topic := BuildLabTopic(TopicLabCompilation, labID)
	return Subscribe(topic, 1, handler)
}

// PublishActuatorControl publishes an actuator control command
func PublishActuatorControl(labID string, command interface{}) error {
	topic := BuildLabTopic(TopicLabActuators, labID)
	return PublishJSON(topic, 1, false, command)
}

// PublishCodeUpload publishes a code upload command
func PublishCodeUpload(labID string, codePayload interface{}) error {
	topic := BuildLabTopic(TopicLabCode, labID)
	return PublishJSON(topic, 1, false, codePayload)
}

// PublishLabControl publishes a control command (start/stop session, etc.)
func PublishLabControl(labID string, command interface{}) error {
	topic := BuildLabTopic(TopicLabControl, labID)
	return PublishJSON(topic, 1, false, command)
}

// PublishSerialCommand publishes a serial command
func PublishSerialCommand(labID string, command interface{}) error {
	topic := BuildLabTopic(TopicLabSerial, labID)
	return PublishJSON(topic, 1, false, command)
}

// ============================================
// Message Types for Lab Communication
// ============================================

// SensorMessage represents sensor data from hardware
type SensorMessage struct {
	LabID     string                 `json:"lab_id"`
	Timestamp time.Time              `json:"timestamp"`
	Sensors   map[string]interface{} `json:"sensors"`
}

// ActuatorCommand represents actuator control command
type ActuatorCommand struct {
	LabID     string                 `json:"lab_id"`
	SessionID string                 `json:"session_id"`
	Actuator  string                 `json:"actuator"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// CodeUploadCommand represents code upload command
type CodeUploadCommand struct {
	LabID         string    `json:"lab_id"`
	SessionID     string    `json:"session_id"`
	CompilationID string    `json:"compilation_id"`
	Code          string    `json:"code"`
	Language      string    `json:"language"`
	Filename      string    `json:"filename"`
	Timestamp     time.Time `json:"timestamp"`
}

// CompilationResult represents compilation result from hardware
type CompilationResult struct {
	LabID         string     `json:"lab_id"`
	SessionID     string     `json:"session_id"`
	CompilationID string     `json:"compilation_id"`
	Status        string     `json:"status"`
	Output        string     `json:"output"`
	Errors        string     `json:"errors,omitempty"`
	UploadedAt    *time.Time `json:"uploaded_at,omitempty"`
}

// HeartbeatMessage represents heartbeat from lab agent
type HeartbeatMessage struct {
	LabID     string    `json:"lab_id"`
	AgentID   string    `json:"agent_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
	Metrics   struct {
		CPUUsage    float64 `json:"cpu_usage"`
		MemoryUsage float64 `json:"memory_usage"`
		Temperature float64 `json:"temperature"`
	} `json:"metrics,omitempty"`
}

// LabControlCommand represents lab control commands
type LabControlCommand struct {
	LabID     string `json:"lab_id"`
	SessionID string `json:"session_id,omitempty"`
	Command   string `json:"command"` // start_session, end_session, reset, restart
	UserID    string `json:"user_id,omitempty"`
}

// SerialCommand represents serial command
type SerialCommand struct {
	LabID     string    `json:"lab_id"`
	SessionID string    `json:"session_id"`
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
}

// SerialResponse represents serial response
type SerialResponse struct {
	LabID     string    `json:"lab_id"`
	SessionID string    `json:"session_id"`
	Response  string    `json:"response"`
	Timestamp time.Time `json:"timestamp"`
}
