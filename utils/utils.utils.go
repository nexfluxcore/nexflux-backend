package utils

import "github.com/gin-gonic/gin"

func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
	c.Abort()
}

func RespondWithSuccess(c *gin.Context, code int, data interface{}) {
	c.JSON(code, gin.H{"data": data})
}
