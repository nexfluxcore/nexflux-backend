package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nexfi-backend/api/services"
	"nexfi-backend/pkg/mqtt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// LabWSHandler handles WebSocket connections for labs
type LabWSHandler struct {
	service    *services.LabService
	clients    map[string]map[*websocket.Conn]bool // labID -> connections
	clientsMu  sync.RWMutex
	sensorData map[string]interface{} // labID -> latest sensor data
	dataMu     sync.RWMutex
}

// NewLabWSHandler creates a new LabWSHandler
func NewLabWSHandler(db *gorm.DB) *LabWSHandler {
	handler := &LabWSHandler{
		service:    services.NewLabService(db),
		clients:    make(map[string]map[*websocket.Conn]bool),
		sensorData: make(map[string]interface{}),
	}

	// Start background goroutine to listen for MQTT messages
	go handler.startMQTTListener()

	return handler
}

// HandleLabWebSocket handles WebSocket connections for a specific lab
// @Summary WebSocket connection for lab
// @Description Connect to WebSocket to receive real-time lab updates (sensor data, queue updates, etc.)
// @Tags Labs WebSocket
// @Param id path string true "Lab ID (UUID)"
// @Router /ws/labs/{id} [get]
func (h *LabWSHandler) HandleLabWebSocket(c *gin.Context) {
	labID := c.Param("id")

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Get user ID from context (optional, may be empty for public viewing)
	userID, _ := c.Get("userID")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Register client
	h.registerClient(labID, conn)
	defer h.unregisterClient(labID, conn)

	log.Printf("📱 WebSocket client connected to lab %s (user: %s)", labID, userIDStr)

	// Send initial state
	h.sendInitialState(conn, labID)

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse WebSocket message: %v", err)
			continue
		}

		// Handle message based on event type
		h.handleMessage(conn, labID, userIDStr, msg)
	}
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// WSResponse represents a WebSocket response
type WSResponse struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// registerClient registers a WebSocket client
func (h *LabWSHandler) registerClient(labID string, conn *websocket.Conn) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if h.clients[labID] == nil {
		h.clients[labID] = make(map[*websocket.Conn]bool)
	}
	h.clients[labID][conn] = true
}

// unregisterClient unregisters a WebSocket client
func (h *LabWSHandler) unregisterClient(labID string, conn *websocket.Conn) {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	if h.clients[labID] != nil {
		delete(h.clients[labID], conn)
	}
}

// sendInitialState sends initial lab state to a new client
func (h *LabWSHandler) sendInitialState(conn *websocket.Conn, labID string) {
	// Get lab details
	lab, err := h.service.GetLab(labID)
	if err != nil {
		h.sendError(conn, "Failed to get lab details")
		return
	}

	// Send lab status
	h.sendToClient(conn, WSResponse{
		Event: "lab_status",
		Data: map[string]interface{}{
			"status":               lab.Status,
			"queue_count":          lab.QueueCount,
			"current_user":         lab.CurrentUser,
			"is_online":            lab.IsOnline,
			"max_session_duration": lab.MaxSessionDuration,
		},
	})

	// Send latest sensor data if available
	h.dataMu.RLock()
	if data, ok := h.sensorData[labID]; ok {
		h.sendToClient(conn, WSResponse{
			Event: "sensor_data",
			Data:  data,
		})
	}
	h.dataMu.RUnlock()
}

// handleMessage handles incoming WebSocket messages
func (h *LabWSHandler) handleMessage(conn *websocket.Conn, labID, userID string, msg WSMessage) {
	switch msg.Event {
	case "actuator_control":
		h.handleActuatorControl(conn, labID, userID, msg.Data)
	case "serial_command":
		h.handleSerialCommand(conn, labID, userID, msg.Data)
	case "read_sensors":
		h.handleReadSensors(conn, labID, userID)
	case "ping":
		h.sendToClient(conn, WSResponse{Event: "pong", Data: map[string]interface{}{"timestamp": time.Now()}})
	default:
		log.Printf("Unknown WebSocket event: %s", msg.Event)
	}
}

// handleActuatorControl handles actuator control messages
func (h *LabWSHandler) handleActuatorControl(conn *websocket.Conn, labID, userID string, data json.RawMessage) {
	if userID == "" {
		h.sendError(conn, "Authentication required for actuator control")
		return
	}

	var cmd struct {
		Actuator string                 `json:"actuator"`
		Action   string                 `json:"action"`
		Params   map[string]interface{} `json:"params"`
	}
	if err := json.Unmarshal(data, &cmd); err != nil {
		h.sendError(conn, "Invalid actuator command")
		return
	}

	// Send command via MQTT
	mqttCmd := mqtt.ActuatorCommand{
		LabID:     labID,
		Actuator:  cmd.Actuator,
		Action:    cmd.Action,
		Params:    cmd.Params,
		Timestamp: time.Now(),
	}

	if err := mqtt.PublishActuatorControl(labID, mqttCmd); err != nil {
		h.sendError(conn, "Failed to send actuator command")
		return
	}

	h.sendToClient(conn, WSResponse{
		Event: "actuator_ack",
		Data: map[string]interface{}{
			"actuator": cmd.Actuator,
			"action":   cmd.Action,
			"success":  true,
		},
	})
}

// handleSerialCommand handles serial command messages
func (h *LabWSHandler) handleSerialCommand(conn *websocket.Conn, labID, userID string, data json.RawMessage) {
	if userID == "" {
		h.sendError(conn, "Authentication required for serial commands")
		return
	}

	var cmd struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(data, &cmd); err != nil {
		h.sendError(conn, "Invalid serial command")
		return
	}

	// Send command via MQTT
	mqttCmd := mqtt.SerialCommand{
		LabID:     labID,
		Command:   cmd.Command,
		Timestamp: time.Now(),
	}

	if err := mqtt.PublishSerialCommand(labID, mqttCmd); err != nil {
		h.sendError(conn, "Failed to send serial command")
		return
	}

	h.sendToClient(conn, WSResponse{
		Event: "serial_ack",
		Data: map[string]interface{}{
			"command": cmd.Command,
			"success": true,
		},
	})
}

// handleReadSensors handles sensor read requests
func (h *LabWSHandler) handleReadSensors(conn *websocket.Conn, labID, userID string) {
	h.dataMu.RLock()
	data, ok := h.sensorData[labID]
	h.dataMu.RUnlock()

	if ok {
		h.sendToClient(conn, WSResponse{
			Event: "sensor_data",
			Data:  data,
		})
	} else {
		h.sendToClient(conn, WSResponse{
			Event: "sensor_data",
			Data: map[string]interface{}{
				"timestamp": time.Now(),
				"sensors":   map[string]interface{}{},
				"message":   "No sensor data available",
			},
		})
	}
}

// sendToClient sends a message to a specific client
func (h *LabWSHandler) sendToClient(conn *websocket.Conn, msg WSResponse) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
}

// sendError sends an error message to a client
func (h *LabWSHandler) sendError(conn *websocket.Conn, message string) {
	h.sendToClient(conn, WSResponse{
		Event: "error",
		Data: map[string]interface{}{
			"message": message,
		},
	})
}

// BroadcastToLab broadcasts a message to all clients connected to a lab
func (h *LabWSHandler) BroadcastToLab(labID string, msg WSResponse) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	clients := h.clients[labID]
	if clients == nil {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Failed to send WebSocket message: %v", err)
		}
	}
}

// startMQTTListener starts listening for MQTT messages
func (h *LabWSHandler) startMQTTListener() {
	// Wait for MQTT to be initialized
	time.Sleep(5 * time.Second)

	if !mqtt.IsConnected() {
		log.Println("⚠️ MQTT not connected, WebSocket sensor broadcast disabled")
		return
	}

	// Subscribe to sensor data for all labs
	mqtt.Subscribe("lab/+/sensors", 1, func(topic string, payload []byte) {
		// Parse lab ID from topic (lab/{lab_id}/sensors)
		var labID string
		_, err := fmt.Sscanf(topic, "lab/%s/sensors", &labID)
		if err != nil {
			return
		}

		// Parse sensor data
		var sensorData mqtt.SensorMessage
		if err := json.Unmarshal(payload, &sensorData); err != nil {
			log.Printf("Failed to parse sensor data: %v", err)
			return
		}

		// Store latest sensor data
		h.dataMu.Lock()
		h.sensorData[labID] = sensorData
		h.dataMu.Unlock()

		// Broadcast to all connected clients
		h.BroadcastToLab(labID, WSResponse{
			Event: "sensor_data",
			Data:  sensorData,
		})
	})

	// Subscribe to compilation results
	mqtt.Subscribe("lab/+/compilation", 1, func(topic string, payload []byte) {
		var labID string
		_, err := fmt.Sscanf(topic, "lab/%s/compilation", &labID)
		if err != nil {
			return
		}

		var result mqtt.CompilationResult
		if err := json.Unmarshal(payload, &result); err != nil {
			return
		}

		h.BroadcastToLab(labID, WSResponse{
			Event: "compilation",
			Data:  result,
		})
	})

	// Subscribe to lab status updates
	mqtt.Subscribe("lab/+/status", 1, func(topic string, payload []byte) {
		var labID string
		_, err := fmt.Sscanf(topic, "lab/%s/status", &labID)
		if err != nil {
			return
		}

		var status map[string]interface{}
		if err := json.Unmarshal(payload, &status); err != nil {
			return
		}

		h.BroadcastToLab(labID, WSResponse{
			Event: "lab_status",
			Data:  status,
		})
	})

	log.Println("📡 WebSocket MQTT listener started")
}
