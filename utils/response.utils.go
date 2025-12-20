// utils/response.go
package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DebugInfo struct {
	RequestID string    `json:"requestId"`
	Version   string    `json:"version"`
	Error     string    `json:"error"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	RuntimeMs int64     `json:"runtimeMs"`
}

type Response struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Debug   DebugInfo   `json:"debug"`
}

func CreateResponse(c *gin.Context, startTime time.Time, message string, data interface{}, err error) {
	// Hitung runtime
	endTime := time.Now()
	runtimeMs := endTime.Sub(startTime).Milliseconds()

	// Buat debug info
	debugInfo := DebugInfo{
		RequestID: uuid.New().String(), // Generate unique request ID
		Version:   "1.0.0",             // Versi aplikasi
		Error:     "",                  // Default error kosong
		StartTime: startTime,
		EndTime:   endTime,
		RuntimeMs: runtimeMs,
	}

	// Jika ada error, tambahkan ke debug info
	if err != nil {
		debugInfo.Error = err.Error()
	}

	// Buat response
	response := Response{
		Data:    data,
		Message: message,
		Debug:   debugInfo,
	}

	// Set header dan kirim response
	c.JSON(http.StatusOK, response)
}
