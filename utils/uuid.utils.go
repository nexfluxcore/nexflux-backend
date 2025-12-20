package utils

import (
	"github.com/google/uuid"
)

// GenerateUUIDv4 menghasilkan UUID versi 4 (random)
func GenerateUUIDv4() string {
	return uuid.New().String()
}

// GenerateUUIDv4NoHyphen menghasilkan UUID tanpa tanda hyphen
func GenerateUUIDv4NoHyphen() string {
	return uuid.New().String()
}

// IsValidUUID memvalidasi string UUID
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
