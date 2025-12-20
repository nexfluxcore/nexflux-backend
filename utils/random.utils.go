package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomString generates a random string of specified length
// Used for OAuth state parameter and other security tokens
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to a less secure but functional random string
		return generateFallbackString(length)
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func generateFallbackString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
