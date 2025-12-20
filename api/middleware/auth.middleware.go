package middleware

import (
	"net/http"
	"nexfi-backend/pkg/redis"
	"strings"

	"nexfi-backend/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the authorization header has the correct format
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		token := bearerToken[1]

		// Verify the token
		claims, err := utils.VerifyJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Verify token in Redis
		storedToken, err := redis.GetToken(claims.Subject)
		if err != nil || storedToken != token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been invalidated"})
			c.Abort()
			return
		}

		// Set the user ID in the context
		c.Set("userId", claims.Subject)
		c.Next()
	}
}
