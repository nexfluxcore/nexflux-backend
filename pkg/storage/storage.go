package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// UploadDir is the base directory for uploads
	UploadDir = "uploads"
	// AvatarDir is the subdirectory for avatar images
	AvatarDir = "avatars"
	// MaxAvatarSize is 5MB
	MaxAvatarSize = 5 * 1024 * 1024
)

var (
	// AllowedImageTypes are the allowed MIME types for avatar uploads
	AllowedImageTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}

	ErrFileTooLarge    = errors.New("file size exceeds maximum allowed (5MB)")
	ErrInvalidFileType = errors.New("invalid file type, only JPEG, PNG, GIF, and WebP are allowed")
	ErrUploadFailed    = errors.New("failed to upload file")
)

// Storage handles file storage operations
type Storage struct {
	baseDir string
	baseURL string
}

// NewStorage creates a new Storage instance
func NewStorage() *Storage {
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &Storage{
		baseDir: UploadDir,
		baseURL: baseURL,
	}
}

// Init initializes the storage directories
func (s *Storage) Init() error {
	dirs := []string{
		filepath.Join(s.baseDir, AvatarDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// UploadAvatar uploads an avatar image and returns the URL
func (s *Storage) UploadAvatar(file *multipart.FileHeader, userID string) (string, error) {
	// Validate file size
	if file.Size > MaxAvatarSize {
		return "", ErrFileTooLarge
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return "", ErrInvalidFileType
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", ErrUploadFailed
	}
	defer src.Close()

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s%s", userID, generateUniqueID(), ext)

	// Create destination path
	destPath := filepath.Join(s.baseDir, AvatarDir, filename)

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return "", ErrUploadFailed
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		// Clean up partial file
		os.Remove(destPath)
		return "", ErrUploadFailed
	}

	// Return the public URL
	return s.GetAvatarURL(filename), nil
}

// DeleteAvatar deletes an avatar file by URL
func (s *Storage) DeleteAvatar(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}

	// Extract filename from URL
	filename := filepath.Base(avatarURL)
	if filename == "" || filename == "." {
		return nil
	}

	path := filepath.Join(s.baseDir, AvatarDir, filename)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(path)
}

// GetAvatarURL returns the full URL for an avatar filename
func (s *Storage) GetAvatarURL(filename string) string {
	return fmt.Sprintf("%s/%s/%s/%s", s.baseURL, s.baseDir, AvatarDir, filename)
}

// DeleteOldAvatar deletes the old avatar when a new one is uploaded
func (s *Storage) DeleteOldAvatar(oldAvatarURL string) {
	if oldAvatarURL == "" {
		return
	}

	// Only delete if it's a local file (contains our upload path)
	if strings.Contains(oldAvatarURL, "/uploads/avatars/") {
		s.DeleteAvatar(oldAvatarURL)
	}
}

// generateUniqueID generates a unique ID for filenames
func generateUniqueID() string {
	return fmt.Sprintf("%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}

// Global storage instance
var DefaultStorage *Storage

// InitStorage initializes the default storage
func InitStorage() error {
	DefaultStorage = NewStorage()
	return DefaultStorage.Init()
}

// UploadAvatar is a convenience function using the default storage
func UploadAvatar(file *multipart.FileHeader, userID string) (string, error) {
	if DefaultStorage == nil {
		if err := InitStorage(); err != nil {
			return "", err
		}
	}
	return DefaultStorage.UploadAvatar(file, userID)
}

// DeleteOldAvatar is a convenience function using the default storage
func DeleteOldAvatar(oldAvatarURL string) {
	if DefaultStorage == nil {
		return
	}
	DefaultStorage.DeleteOldAvatar(oldAvatarURL)
}
