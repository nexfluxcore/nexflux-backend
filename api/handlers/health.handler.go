package handlers

import (
	"net/http"
	"nexfi-backend/database"
	"nexfi-backend/pkg/redis"

	"github.com/gin-gonic/gin"
)

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
