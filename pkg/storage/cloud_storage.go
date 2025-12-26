package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage type constants
const (
	StorageTypeLocal = "local"
	StorageTypeMinio = "minio"
)

// Bucket names
const (
	BucketAvatars    = "avatars"
	BucketProjects   = "projects"
	BucketComponents = "components"
	BucketDocuments  = "documents"
	BucketThumbnails = "thumbnails"
)

// File size limits
const (
	MaxAvatarSize   = 5 * 1024 * 1024  // 5MB
	MaxImageSize    = 10 * 1024 * 1024 // 10MB
	MaxDocumentSize = 50 * 1024 * 1024 // 50MB
)

var (
	// AllowedImageTypes for avatar/image uploads
	AllowedImageTypes = map[string]string{
		"image/jpeg":    ".jpg",
		"image/png":     ".png",
		"image/gif":     ".gif",
		"image/webp":    ".webp",
		"image/svg+xml": ".svg",
	}

	// AllowedDocumentTypes for document uploads
	AllowedDocumentTypes = map[string]string{
		"application/pdf":  ".pdf",
		"application/zip":  ".zip",
		"text/plain":       ".txt",
		"application/json": ".json",
	}

	// Custom errors
	ErrFileTooLarge    = errors.New("file size exceeds maximum allowed")
	ErrInvalidFileType = errors.New("invalid file type")
	ErrUploadFailed    = errors.New("failed to upload file")
	ErrDeleteFailed    = errors.New("failed to delete file")
	ErrStorageNotInit  = errors.New("storage not initialized")
	ErrBucketNotFound  = errors.New("bucket not found")
)

// MinioConfig holds MinIO configuration
type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	BucketPrefix    string // Prefix for bucket names (e.g., "nexflux-")
	PublicURL       string // Public URL for accessing files (CDN or direct)
}

// CloudStorage handles MinIO/S3 storage operations
type CloudStorage struct {
	client       *minio.Client
	config       MinioConfig
	ctx          context.Context
	storageType  string
	localBaseDir string
	localBaseURL string
}

// NewCloudStorage creates a new CloudStorage instance
func NewCloudStorage() (*CloudStorage, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = StorageTypeLocal // Default to local storage
	}

	cs := &CloudStorage{
		ctx:         context.Background(),
		storageType: storageType,
	}

	if storageType == StorageTypeMinio {
		return cs.initMinIO()
	}

	return cs.initLocal()
}

// initMinIO initializes MinIO client
func (cs *CloudStorage) initMinIO() (*CloudStorage, error) {
	cs.config = MinioConfig{
		Endpoint:        os.Getenv("MINIO_ENDPOINT"),
		AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("MINIO_SECRET_KEY"),
		UseSSL:          os.Getenv("MINIO_USE_SSL") == "true",
		Region:          getEnvOrDefault("MINIO_REGION", "us-east-1"),
		BucketPrefix:    getEnvOrDefault("MINIO_BUCKET_PREFIX", "nexflux-"),
		PublicURL:       os.Getenv("MINIO_PUBLIC_URL"),
	}

	// Validate required config
	if cs.config.Endpoint == "" || cs.config.AccessKeyID == "" || cs.config.SecretAccessKey == "" {
		return nil, errors.New("MinIO configuration incomplete: MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY are required")
	}

	// Initialize MinIO client
	client, err := minio.New(cs.config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cs.config.AccessKeyID, cs.config.SecretAccessKey, ""),
		Secure: cs.config.UseSSL,
		Region: cs.config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	cs.client = client

	// Create default buckets
	if err := cs.ensureBuckets(); err != nil {
		return nil, err
	}

	log.Printf("✅ MinIO storage initialized: %s", cs.config.Endpoint)
	return cs, nil
}

// initLocal initializes local file storage
func (cs *CloudStorage) initLocal() (*CloudStorage, error) {
	cs.localBaseDir = getEnvOrDefault("UPLOAD_DIR", "uploads")

	port := getEnvOrDefault("PORT", "8080")
	cs.localBaseURL = getEnvOrDefault("APP_URL", fmt.Sprintf("http://localhost:%s", port))

	// Create local directories
	dirs := []string{
		filepath.Join(cs.localBaseDir, BucketAvatars),
		filepath.Join(cs.localBaseDir, BucketProjects),
		filepath.Join(cs.localBaseDir, BucketComponents),
		filepath.Join(cs.localBaseDir, BucketDocuments),
		filepath.Join(cs.localBaseDir, BucketThumbnails),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	log.Printf("✅ Local storage initialized: %s", cs.localBaseDir)
	return cs, nil
}

// ensureBuckets creates the required buckets if they don't exist
func (cs *CloudStorage) ensureBuckets() error {
	buckets := []string{
		BucketAvatars,
		BucketProjects,
		BucketComponents,
		BucketDocuments,
		BucketThumbnails,
	}

	for _, bucket := range buckets {
		bucketName := cs.getBucketName(bucket)

		exists, err := cs.client.BucketExists(cs.ctx, bucketName)
		if err != nil {
			return fmt.Errorf("failed to check bucket %s: %w", bucketName, err)
		}

		if !exists {
			err = cs.client.MakeBucket(cs.ctx, bucketName, minio.MakeBucketOptions{
				Region: cs.config.Region,
			})
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
			}

			// Set bucket policy to public for read access
			policy := fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}]
			}`, bucketName)

			if err := cs.client.SetBucketPolicy(cs.ctx, bucketName, policy); err != nil {
				log.Printf("Warning: Failed to set public policy for bucket %s: %v", bucketName, err)
			}

			log.Printf("  → Created bucket: %s", bucketName)
		}
	}

	return nil
}

// getBucketName returns the full bucket name with prefix
func (cs *CloudStorage) getBucketName(bucket string) string {
	return cs.config.BucketPrefix + bucket
}

// UploadFile uploads a file to the specified bucket
func (cs *CloudStorage) UploadFile(bucket string, file *multipart.FileHeader, prefix string) (string, error) {
	if cs.storageType == StorageTypeMinio {
		return cs.uploadToMinio(bucket, file, prefix)
	}
	return cs.uploadToLocal(bucket, file, prefix)
}

// uploadToMinio uploads file to MinIO
func (cs *CloudStorage) uploadToMinio(bucket string, file *multipart.FileHeader, prefix string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", ErrUploadFailed
	}
	defer src.Close()

	// Generate object name
	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s/%s%s", prefix, generateUniqueID(), ext)

	// Get content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to MinIO
	bucketName := cs.getBucketName(bucket)
	_, err = cs.client.PutObject(cs.ctx, bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	return cs.GetFileURL(bucket, objectName), nil
}

// uploadToLocal uploads file to local storage
func (cs *CloudStorage) uploadToLocal(bucket string, file *multipart.FileHeader, prefix string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", ErrUploadFailed
	}
	defer src.Close()

	// Generate filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%s%s", prefix, generateUniqueID(), ext)

	// Create destination path
	destPath := filepath.Join(cs.localBaseDir, bucket, filename)

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return "", ErrUploadFailed
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(destPath)
		return "", ErrUploadFailed
	}

	return fmt.Sprintf("%s/%s/%s/%s", cs.localBaseURL, cs.localBaseDir, bucket, filename), nil
}

// UploadAvatar uploads an avatar image
func (cs *CloudStorage) UploadAvatar(file *multipart.FileHeader, userID string) (string, error) {
	// Validate file size
	if file.Size > MaxAvatarSize {
		return "", ErrFileTooLarge
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if _, ok := AllowedImageTypes[contentType]; !ok {
		return "", ErrInvalidFileType
	}

	return cs.UploadFile(BucketAvatars, file, userID)
}

// UploadProjectThumbnail uploads a project thumbnail
func (cs *CloudStorage) UploadProjectThumbnail(file *multipart.FileHeader, projectID string) (string, error) {
	if file.Size > MaxImageSize {
		return "", ErrFileTooLarge
	}

	contentType := file.Header.Get("Content-Type")
	if _, ok := AllowedImageTypes[contentType]; !ok {
		return "", ErrInvalidFileType
	}

	return cs.UploadFile(BucketThumbnails, file, fmt.Sprintf("project_%s", projectID))
}

// UploadComponentImage uploads a component image
func (cs *CloudStorage) UploadComponentImage(file *multipart.FileHeader, componentID string) (string, error) {
	if file.Size > MaxImageSize {
		return "", ErrFileTooLarge
	}

	contentType := file.Header.Get("Content-Type")
	if _, ok := AllowedImageTypes[contentType]; !ok {
		return "", ErrInvalidFileType
	}

	return cs.UploadFile(BucketComponents, file, fmt.Sprintf("component_%s", componentID))
}

// GetFileURL returns the public URL for a file
func (cs *CloudStorage) GetFileURL(bucket, objectName string) string {
	if cs.storageType == StorageTypeMinio {
		// Use public URL if configured (CDN), otherwise construct from endpoint
		if cs.config.PublicURL != "" {
			return fmt.Sprintf("%s/%s/%s", cs.config.PublicURL, cs.getBucketName(bucket), objectName)
		}

		protocol := "http"
		if cs.config.UseSSL {
			protocol = "https"
		}
		return fmt.Sprintf("%s://%s/%s/%s", protocol, cs.config.Endpoint, cs.getBucketName(bucket), objectName)
	}

	// Local storage URL
	return fmt.Sprintf("%s/%s/%s/%s", cs.localBaseURL, cs.localBaseDir, bucket, objectName)
}

// DeleteFile deletes a file from storage
func (cs *CloudStorage) DeleteFile(fileURL string) error {
	if fileURL == "" {
		return nil
	}

	if cs.storageType == StorageTypeMinio {
		return cs.deleteFromMinio(fileURL)
	}
	return cs.deleteFromLocal(fileURL)
}

// deleteFromMinio deletes a file from MinIO
func (cs *CloudStorage) deleteFromMinio(fileURL string) error {
	// Parse URL to get bucket and object name
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return nil
	}

	// Extract path parts
	pathParts := strings.SplitN(strings.TrimPrefix(parsedURL.Path, "/"), "/", 2)
	if len(pathParts) < 2 {
		return nil
	}

	bucketName := pathParts[0]
	objectName := pathParts[1]

	err = cs.client.RemoveObject(cs.ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from MinIO: %w", err)
	}

	return nil
}

// deleteFromLocal deletes a file from local storage
func (cs *CloudStorage) deleteFromLocal(fileURL string) error {
	// Extract filename from URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return nil
	}

	// Remove base path to get relative path
	path := strings.TrimPrefix(parsedURL.Path, "/")

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

// DeleteOldFile safely deletes an old file (checks if it's from our storage)
func (cs *CloudStorage) DeleteOldFile(oldURL string) {
	if oldURL == "" {
		return
	}

	// Check if it's from our storage
	if cs.storageType == StorageTypeMinio {
		if strings.Contains(oldURL, cs.config.Endpoint) ||
			(cs.config.PublicURL != "" && strings.Contains(oldURL, cs.config.PublicURL)) {
			cs.DeleteFile(oldURL)
		}
	} else {
		if strings.Contains(oldURL, cs.localBaseDir) {
			cs.DeleteFile(oldURL)
		}
	}
}

// GetPresignedURL returns a presigned URL for private file access
func (cs *CloudStorage) GetPresignedURL(bucket, objectName string, expiry time.Duration) (string, error) {
	if cs.storageType != StorageTypeMinio {
		return cs.GetFileURL(bucket, objectName), nil
	}

	bucketName := cs.getBucketName(bucket)
	presignedURL, err := cs.client.PresignedGetObject(cs.ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// generateUniqueID generates a unique ID for filenames
func generateUniqueID() string {
	return fmt.Sprintf("%d_%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

// getEnvOrDefault returns env value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ============================================
// GLOBAL INSTANCE & CONVENIENCE FUNCTIONS
// ============================================

var DefaultCloudStorage *CloudStorage

// InitCloudStorage initializes the default cloud storage
func InitCloudStorage() error {
	storage, err := NewCloudStorage()
	if err != nil {
		return err
	}
	DefaultCloudStorage = storage
	return nil
}

// UploadAvatar convenience function
func UploadAvatar(file *multipart.FileHeader, userID string) (string, error) {
	if DefaultCloudStorage == nil {
		return "", ErrStorageNotInit
	}
	return DefaultCloudStorage.UploadAvatar(file, userID)
}

// UploadProjectThumbnail convenience function
func UploadProjectThumbnail(file *multipart.FileHeader, projectID string) (string, error) {
	if DefaultCloudStorage == nil {
		return "", ErrStorageNotInit
	}
	return DefaultCloudStorage.UploadProjectThumbnail(file, projectID)
}

// DeleteOldAvatar convenience function
func DeleteOldAvatar(oldURL string) {
	if DefaultCloudStorage == nil {
		return
	}
	DefaultCloudStorage.DeleteOldFile(oldURL)
}

// GetStorageType returns current storage type
func GetStorageType() string {
	if DefaultCloudStorage == nil {
		return StorageTypeLocal
	}
	return DefaultCloudStorage.storageType
}

// IsInitialized checks if storage is initialized
func IsInitialized() bool {
	return DefaultCloudStorage != nil
}

// CheckHealth performs a health check on the storage
func CheckHealth() error {
	if DefaultCloudStorage == nil {
		return ErrStorageNotInit
	}

	if DefaultCloudStorage.storageType == StorageTypeMinio {
		// Try to list buckets as health check
		_, err := DefaultCloudStorage.client.ListBuckets(DefaultCloudStorage.ctx)
		if err != nil {
			return fmt.Errorf("MinIO health check failed: %w", err)
		}
	} else {
		// Check if local directory is accessible
		if _, err := os.Stat(DefaultCloudStorage.localBaseDir); os.IsNotExist(err) {
			return fmt.Errorf("local storage directory not accessible: %w", err)
		}
	}

	return nil
}

// InitStorage is an alias for InitCloudStorage (backward compatibility)
func InitStorage() error {
	return InitCloudStorage()
}
