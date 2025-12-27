package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
	once    sync.Once
	mu      sync.RWMutex
)

// Queue names for different job types
const (
	QueueCodeCompilation   = "lab.code.compilation"
	QueueSessionExpiration = "lab.session.expiration"
	QueueQueueExpiration   = "lab.queue.expiration"
	QueueXPReward          = "lab.xp.reward"
	QueueNotification      = "lab.notification"
	QueueHardwareLog       = "lab.hardware.log"
)

// Exchange names
const (
	ExchangeLabEvents = "lab.events"
	ExchangeLabDirect = "lab.direct"
)

// Routing keys
const (
	RoutingKeyCompilation   = "compilation"
	RoutingKeySessionExpire = "session.expire"
	RoutingKeyQueueExpire   = "queue.expire"
	RoutingKeyXPReward      = "xp.reward"
	RoutingKeyNotification  = "notification"
	RoutingKeyHardwareLog   = "hardware.log"
)

// Config holds RabbitMQ configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
}

// GetConfig returns RabbitMQ configuration from environment
func GetConfig() Config {
	return Config{
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnv("RABBITMQ_PORT", "5672"),
		User:     getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		VHost:    getEnv("RABBITMQ_VHOST", "/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitRabbitMQ initializes the RabbitMQ connection
func InitRabbitMQ() error {
	var initErr error
	once.Do(func() {
		config := GetConfig()

		// Check if full URL is provided
		var url string
		if envURL := os.Getenv("RABBITMQ_URL"); envURL != "" {
			url = envURL
			log.Printf("📡 Using RABBITMQ_URL from environment")
		} else {
			// Build URL from components
			// VHost "/" needs to be URL-encoded as "%2F"
			vhost := config.VHost
			if vhost == "/" {
				vhost = "" // Empty vhost in URL means default "/"
			} else if vhost != "" && vhost[0] != '/' {
				// URL encode the vhost if it's not already
				vhost = "/" + vhost
			}

			url = fmt.Sprintf("amqp://%s:%s@%s:%s%s",
				config.User, config.Password, config.Host, config.Port, vhost)
		}

		log.Printf("📡 Connecting to RabbitMQ at %s:%s...", config.Host, config.Port)

		var err error
		conn, err = amqp.DialConfig(url, amqp.Config{
			Dial: amqp.DefaultDial(10 * time.Second), // Increased timeout
		})
		if err != nil {
			initErr = fmt.Errorf("failed to connect to RabbitMQ: %v", err)
			log.Printf("❌ %v", initErr)
			return
		}

		channel, err = conn.Channel()
		if err != nil {
			initErr = fmt.Errorf("failed to open channel: %v", err)
			log.Printf("❌ %v", initErr)
			return
		}

		// Setup exchanges
		if err := setupExchanges(); err != nil {
			initErr = err
			return
		}

		// Setup queues
		if err := setupQueues(); err != nil {
			initErr = err
			return
		}

		log.Printf("✅ RabbitMQ connected successfully")
	})

	return initErr
}

// setupExchanges creates necessary exchanges
func setupExchanges() error {
	// Events exchange (fanout for broadcasting)
	err := channel.ExchangeDeclare(
		ExchangeLabEvents,
		"fanout",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare events exchange: %v", err)
	}

	// Direct exchange for job queues
	err = channel.ExchangeDeclare(
		ExchangeLabDirect,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare direct exchange: %v", err)
	}

	return nil
}

// setupQueues creates necessary queues
func setupQueues() error {
	queues := []struct {
		name       string
		routingKey string
	}{
		{QueueCodeCompilation, RoutingKeyCompilation},
		{QueueSessionExpiration, RoutingKeySessionExpire},
		{QueueQueueExpiration, RoutingKeyQueueExpire},
		{QueueXPReward, RoutingKeyXPReward},
		{QueueNotification, RoutingKeyNotification},
		{QueueHardwareLog, RoutingKeyHardwareLog},
	}

	for _, q := range queues {
		_, err := channel.QueueDeclare(
			q.name,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			amqp.Table{
				"x-message-ttl": int32(86400000), // 24 hours TTL
			},
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %v", q.name, err)
		}

		err = channel.QueueBind(
			q.name,
			q.routingKey,
			ExchangeLabDirect,
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %v", q.name, err)
		}
	}

	return nil
}

// IsConnected checks if RabbitMQ is connected
func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()
	return conn != nil && !conn.IsClosed()
}

// Close closes the RabbitMQ connection
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if channel != nil {
		channel.Close()
	}
	if conn != nil {
		conn.Close()
	}
	log.Printf("🔌 RabbitMQ connection closed")
}

// ============================================
// Publishing Functions
// ============================================

// Publish publishes a message to a queue
func Publish(queueName string, routingKey string, message interface{}) error {
	mu.RLock()
	defer mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(ctx,
		ExchangeLabDirect,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

// PublishDelayed publishes a message with a delay
func PublishDelayed(queueName string, routingKey string, message interface{}, delay time.Duration) error {
	mu.RLock()
	defer mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use x-delay header for delayed message plugin
	err = channel.PublishWithContext(ctx,
		ExchangeLabDirect,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers: amqp.Table{
				"x-delay": int32(delay.Milliseconds()),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish delayed message: %v", err)
	}

	return nil
}

// ============================================
// Consumer Functions
// ============================================

// MessageHandler is a function type for handling messages
type MessageHandler func(body []byte) error

// Consume starts consuming messages from a queue
func Consume(queueName string, handler MessageHandler) error {
	mu.RLock()
	defer mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}

	msgs, err := channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("❌ Error processing message: %v", err)
				msg.Nack(false, true) // Requeue on error
			} else {
				msg.Ack(false)
			}
		}
	}()

	log.Printf("📥 Started consuming from queue: %s", queueName)
	return nil
}

// ConsumeWithContext starts consuming with context for graceful shutdown
func ConsumeWithContext(ctx context.Context, queueName string, handler MessageHandler) error {
	mu.RLock()
	defer mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}

	msgs, err := channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %v", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("🛑 Consumer stopped for queue: %s", queueName)
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				if err := handler(msg.Body); err != nil {
					log.Printf("❌ Error processing message: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	log.Printf("📥 Started consuming from queue: %s", queueName)
	return nil
}

// ============================================
// Lab-specific Job Publishers
// ============================================

// CodeCompilationJob represents a code compilation job
type CodeCompilationJob struct {
	CompilationID string    `json:"compilation_id"`
	LabID         string    `json:"lab_id"`
	SessionID     string    `json:"session_id"`
	UserID        string    `json:"user_id"`
	Code          string    `json:"code"`
	Language      string    `json:"language"`
	Filename      string    `json:"filename"`
	CreatedAt     time.Time `json:"created_at"`
}

// PublishCodeCompilation publishes a code compilation job
func PublishCodeCompilation(job CodeCompilationJob) error {
	job.CreatedAt = time.Now()
	return Publish(QueueCodeCompilation, RoutingKeyCompilation, job)
}

// SessionExpirationJob represents a session expiration job
type SessionExpirationJob struct {
	SessionID string    `json:"session_id"`
	LabID     string    `json:"lab_id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PublishSessionExpiration publishes a session expiration job
func PublishSessionExpiration(job SessionExpirationJob) error {
	delay := time.Until(job.ExpiresAt)
	if delay < 0 {
		delay = 0
	}
	return PublishDelayed(QueueSessionExpiration, RoutingKeySessionExpire, job, delay)
}

// QueueExpirationJob represents a queue expiration job
type QueueExpirationJob struct {
	QueueID   string    `json:"queue_id"`
	LabID     string    `json:"lab_id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PublishQueueExpiration publishes a queue expiration job
func PublishQueueExpiration(job QueueExpirationJob) error {
	delay := time.Until(job.ExpiresAt)
	if delay < 0 {
		delay = 0
	}
	return PublishDelayed(QueueQueueExpiration, RoutingKeyQueueExpire, job, delay)
}

// XPRewardJob represents an XP reward job
type XPRewardJob struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	Action    string    `json:"action"`
	XPAmount  int       `json:"xp_amount"`
	CreatedAt time.Time `json:"created_at"`
}

// PublishXPReward publishes an XP reward job
func PublishXPReward(job XPRewardJob) error {
	job.CreatedAt = time.Now()
	return Publish(QueueXPReward, RoutingKeyXPReward, job)
}

// NotificationJob represents a notification job
type NotificationJob struct {
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// PublishNotification publishes a notification job
func PublishNotification(job NotificationJob) error {
	job.CreatedAt = time.Now()
	return Publish(QueueNotification, RoutingKeyNotification, job)
}

// HardwareLogJob represents a hardware log job
type HardwareLogJob struct {
	LabID     string                 `json:"lab_id"`
	SessionID string                 `json:"session_id,omitempty"`
	EventType string                 `json:"event_type"`
	EventData map[string]interface{} `json:"event_data"`
	CreatedAt time.Time              `json:"created_at"`
}

// PublishHardwareLog publishes a hardware log job
func PublishHardwareLog(job HardwareLogJob) error {
	job.CreatedAt = time.Now()
	return Publish(QueueHardwareLog, RoutingKeyHardwareLog, job)
}
