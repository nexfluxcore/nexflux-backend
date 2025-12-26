package handlers

import (
	"context"
	"net/http"
	"nexfi-backend/database"
	"nexfi-backend/pkg/livekit"
	"nexfi-backend/pkg/mqtt"
	"nexfi-backend/pkg/rabbitmq"
	"nexfi-backend/pkg/redis"
	"nexfi-backend/pkg/storage"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ServiceStatus represents the status of a single service
type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "healthy", "unhealthy", "degraded"
	Latency string `json:"latency,omitempty"`
	Message string `json:"message,omitempty"`
}

// HealthResponse represents the overall health status
type HealthResponse struct {
	Status    string          `json:"status"` // "healthy", "unhealthy", "degraded"
	Timestamp string          `json:"timestamp"`
	Version   string          `json:"version"`
	Services  []ServiceStatus `json:"services"`
}

// FullHealthCheck godoc
// @Summary Full health check endpoint
// @Description Get the health status of all services including Database, Redis, MQTT, RabbitMQ, LiveKit, and Storage
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse "All services healthy"
// @Success 503 {object} HealthResponse "One or more services unhealthy"
// @Router /health [get]
func (h *HealthHandler) FullHealthCheck(c *gin.Context) {
	services := make([]ServiceStatus, 0)
	allHealthy := true
	hasDegraded := false

	// Check Database (PostgreSQL)
	dbStatus := h.checkDatabase()
	services = append(services, dbStatus)
	if dbStatus.Status == "unhealthy" {
		allHealthy = false
	} else if dbStatus.Status == "degraded" {
		hasDegraded = true
	}

	// Check Redis
	redisStatus := h.checkRedis()
	services = append(services, redisStatus)
	if redisStatus.Status == "unhealthy" {
		allHealthy = false
	} else if redisStatus.Status == "degraded" {
		hasDegraded = true
	}

	// Check MQTT (VerneMQ)
	mqttStatus := h.checkMQTT()
	services = append(services, mqttStatus)
	if mqttStatus.Status == "unhealthy" {
		// MQTT is optional, mark as degraded instead of unhealthy
		hasDegraded = true
	}

	// Check RabbitMQ
	rabbitStatus := h.checkRabbitMQ()
	services = append(services, rabbitStatus)
	if rabbitStatus.Status == "unhealthy" {
		// RabbitMQ is optional for basic ops, mark as degraded
		hasDegraded = true
	}

	// Check LiveKit
	livekitStatus := h.checkLiveKit()
	services = append(services, livekitStatus)
	if livekitStatus.Status == "unhealthy" {
		// LiveKit is optional for basic ops, mark as degraded
		hasDegraded = true
	}

	// Check Storage (MinIO/Local)
	storageStatus := h.checkStorage()
	services = append(services, storageStatus)
	if storageStatus.Status == "unhealthy" {
		hasDegraded = true
	}

	// Determine overall status
	overallStatus := "healthy"
	httpStatus := http.StatusOK
	if !allHealthy {
		overallStatus = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	} else if hasDegraded {
		overallStatus = "degraded"
		httpStatus = http.StatusOK // Still return 200 for degraded
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  services,
	}

	c.JSON(httpStatus, response)
}

// checkDatabase checks PostgreSQL connection
func (h *HealthHandler) checkDatabase() ServiceStatus {
	status := ServiceStatus{
		Name:   "PostgreSQL",
		Status: "healthy",
	}

	start := time.Now()
	sqlDB, err := database.DB.DB()
	if err != nil {
		status.Status = "unhealthy"
		status.Message = "Failed to get database connection: " + err.Error()
		return status
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		status.Status = "unhealthy"
		status.Message = "Database ping failed: " + err.Error()
		return status
	}

	status.Latency = time.Since(start).String()
	status.Message = "Connected"
	return status
}

// checkRedis checks Redis connection
func (h *HealthHandler) checkRedis() ServiceStatus {
	status := ServiceStatus{
		Name:   "Redis",
		Status: "healthy",
	}

	start := time.Now()
	if !redis.IsConnected() {
		status.Status = "unhealthy"
		status.Message = "Redis not connected"
		return status
	}

	// Try a ping
	if err := redis.Ping(); err != nil {
		status.Status = "unhealthy"
		status.Message = "Redis ping failed: " + err.Error()
		return status
	}

	status.Latency = time.Since(start).String()
	status.Message = "Connected"
	return status
}

// checkMQTT checks VerneMQ/MQTT connection
func (h *HealthHandler) checkMQTT() ServiceStatus {
	status := ServiceStatus{
		Name:   "MQTT (VerneMQ)",
		Status: "healthy",
	}

	if !mqtt.IsConnected() {
		status.Status = "unhealthy"
		status.Message = "MQTT broker not connected"
		return status
	}

	status.Message = "Connected"
	return status
}

// checkRabbitMQ checks RabbitMQ connection
func (h *HealthHandler) checkRabbitMQ() ServiceStatus {
	status := ServiceStatus{
		Name:   "RabbitMQ",
		Status: "healthy",
	}

	if !rabbitmq.IsConnected() {
		status.Status = "unhealthy"
		status.Message = "RabbitMQ not connected"
		return status
	}

	status.Message = "Connected"
	return status
}

// checkLiveKit checks LiveKit connection
func (h *HealthHandler) checkLiveKit() ServiceStatus {
	status := ServiceStatus{
		Name:   "LiveKit",
		Status: "healthy",
	}

	if !livekit.IsInitialized() {
		status.Status = "unhealthy"
		status.Message = "LiveKit not initialized"
		return status
	}

	status.Message = "Initialized"
	return status
}

// checkStorage checks MinIO/Local storage
func (h *HealthHandler) checkStorage() ServiceStatus {
	status := ServiceStatus{
		Name:   "Storage",
		Status: "healthy",
	}

	start := time.Now()
	if !storage.IsInitialized() {
		status.Status = "unhealthy"
		status.Message = "Storage not initialized"
		return status
	}

	// Get storage type info
	storageType := storage.GetStorageType()
	status.Message = "Type: " + storageType

	if storageType == "minio" {
		// Check MinIO bucket accessibility
		if err := storage.CheckHealth(); err != nil {
			status.Status = "degraded"
			status.Message = "MinIO health check failed: " + err.Error()
			return status
		}
	}

	status.Latency = time.Since(start).String()
	return status
}

// SimpleHealth godoc
// @Summary Simple health check
// @Description Quick health check that only returns OK
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "OK"
// @Router /ping [get]
func (h *HealthHandler) SimpleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "NexFlux Virtual Lab API is running",
	})
}

// ReadyCheck godoc
// @Summary Readiness check
// @Description Check if the service is ready to accept traffic (all critical services must be available)
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Ready"
// @Success 503 {object} map[string]string "Not Ready"
// @Router /ready [get]
func (h *HealthHandler) ReadyCheck(c *gin.Context) {
	// Check critical services only (Database and Redis)
	sqlDB, err := database.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not_ready",
			"message": "Database connection unavailable",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not_ready",
			"message": "Database ping failed",
		})
		return
	}

	if !redis.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not_ready",
			"message": "Redis not connected",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "All critical services are available",
	})
}

// ============================================
// Legacy handler for backward compatibility
// ============================================

type HealthCheckHandler struct{}

func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

func (h *HealthCheckHandler) HealthCheck(c *gin.Context) {
	dbStatus := "ok"
	db, err := database.DB.DB()
	if err != nil {
		dbStatus = "error"
	}
	if err := db.Ping(); err != nil {
		dbStatus = "error"
	}

	redisStatus := "ok"
	if redis.RedisClient == nil {
		redisStatus = "disabled"
	} else {
		if _, err := redis.RedisClient.Ping(c.Request.Context()).Result(); err != nil {
			redisStatus = "error"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"database": dbStatus,
		"redis":    redisStatus,
	})
}
