package handlers

import (
	"fmt"
	"net/http"
	"nexfi-backend/pkg/storage"
	"nexfi-backend/utils"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles file upload requests
type UploadHandler struct{}

// NewUploadHandler creates a new UploadHandler
func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadType defines allowed upload categories
type UploadType string

const (
	UploadTypeAvatar    UploadType = "avatar"
	UploadTypeProject   UploadType = "project"
	UploadTypeComponent UploadType = "component"
	UploadTypeDocument  UploadType = "document"
	UploadTypeThumbnail UploadType = "thumbnail"
	UploadTypeGeneral   UploadType = "general"
)

// UploadResponse defines the response structure for uploads
type UploadResponse struct {
	Success  bool   `json:"success"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	Type     string `json:"type"`
}

// Upload godoc
// @Summary Upload a file
// @Description Upload a file to cloud storage and get the URL back
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param type query string false "Upload type: avatar, project, component, document, thumbnail, general" default(general)
// @Security Bearer
// @Success 200 {object} UploadResponse "Upload successful"
// @Failure 400 {object} map[string]string "Invalid file or type"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 413 {object} map[string]string "File too large"
// @Failure 500 {object} map[string]string "Upload failed"
// @Router /upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get upload type
	uploadType := UploadType(c.DefaultQuery("type", string(UploadTypeGeneral)))

	// Validate upload type
	if !isValidUploadType(uploadType) {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid upload type. Allowed: avatar, project, component, document, thumbnail, general")
		return
	}

	// Get the file
	file, err := c.FormFile("file")
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "No file provided")
		return
	}

	// Validate file size based on type
	maxSize := getMaxSizeForType(uploadType)
	if file.Size > maxSize {
		utils.RespondWithError(c, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("File too large. Maximum size for %s is %dMB", uploadType, maxSize/(1024*1024)))
		return
	}

	// Validate file type based on upload type
	contentType := file.Header.Get("Content-Type")
	if !isAllowedContentType(uploadType, contentType) {
		utils.RespondWithError(c, http.StatusBadRequest,
			fmt.Sprintf("Invalid file type '%s' for %s upload", contentType, uploadType))
		return
	}

	// Determine bucket based on type
	bucket := getBucketForType(uploadType)

	// Generate prefix (user ID + optional context)
	prefix := userID.(string)

	// Upload to storage
	url, err := storage.DefaultCloudStorage.UploadFile(bucket, file, prefix)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to upload file: "+err.Error())
		return
	}

	// Return the URL
	c.JSON(http.StatusOK, UploadResponse{
		Success:  true,
		URL:      url,
		Filename: filepath.Base(url),
		Size:     file.Size,
		MimeType: contentType,
		Type:     string(uploadType),
	})
}

// UploadMultiple godoc
// @Summary Upload multiple files
// @Description Upload multiple files to cloud storage and get URLs back
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Files to upload (multiple)"
// @Param type query string false "Upload type" default(general)
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Upload successful"
// @Failure 400 {object} map[string]string "Invalid files"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /upload/multiple [post]
func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	uploadType := UploadType(c.DefaultQuery("type", string(UploadTypeGeneral)))
	if !isValidUploadType(uploadType) {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid upload type")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to parse form")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "No files provided")
		return
	}

	// Limit to 10 files
	if len(files) > 10 {
		utils.RespondWithError(c, http.StatusBadRequest, "Maximum 10 files allowed per request")
		return
	}

	bucket := getBucketForType(uploadType)
	prefix := userID.(string)
	maxSize := getMaxSizeForType(uploadType)

	var results []UploadResponse
	var errors []string

	for _, file := range files {
		// Validate size
		if file.Size > maxSize {
			errors = append(errors, fmt.Sprintf("%s: file too large", file.Filename))
			continue
		}

		// Validate content type
		contentType := file.Header.Get("Content-Type")
		if !isAllowedContentType(uploadType, contentType) {
			errors = append(errors, fmt.Sprintf("%s: invalid file type", file.Filename))
			continue
		}

		// Upload
		url, err := storage.DefaultCloudStorage.UploadFile(bucket, file, prefix)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: upload failed", file.Filename))
			continue
		}

		results = append(results, UploadResponse{
			Success:  true,
			URL:      url,
			Filename: filepath.Base(url),
			Size:     file.Size,
			MimeType: contentType,
			Type:     string(uploadType),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   len(errors) == 0,
		"uploaded":  results,
		"errors":    errors,
		"total":     len(files),
		"succeeded": len(results),
		"failed":    len(errors),
	})
}

// DeleteFile godoc
// @Summary Delete a file
// @Description Delete a previously uploaded file by URL
// @Tags Upload
// @Accept json
// @Produce json
// @Param url query string true "File URL to delete"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Delete successful"
// @Failure 400 {object} map[string]string "Invalid URL"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /upload [delete]
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	fileURL := c.Query("url")
	if fileURL == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "URL is required")
		return
	}

	if err := storage.DefaultCloudStorage.DeleteFile(fileURL); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File deleted successfully",
	})
}

// Helper functions

func isValidUploadType(t UploadType) bool {
	switch t {
	case UploadTypeAvatar, UploadTypeProject, UploadTypeComponent,
		UploadTypeDocument, UploadTypeThumbnail, UploadTypeGeneral:
		return true
	}
	return false
}

func getMaxSizeForType(t UploadType) int64 {
	switch t {
	case UploadTypeAvatar:
		return 5 * 1024 * 1024 // 5MB
	case UploadTypeThumbnail:
		return 10 * 1024 * 1024 // 10MB
	case UploadTypeDocument:
		return 50 * 1024 * 1024 // 50MB
	default:
		return 25 * 1024 * 1024 // 25MB default
	}
}

func getBucketForType(t UploadType) string {
	switch t {
	case UploadTypeAvatar:
		return storage.BucketAvatars
	case UploadTypeProject:
		return storage.BucketProjects
	case UploadTypeComponent:
		return storage.BucketComponents
	case UploadTypeDocument:
		return storage.BucketDocuments
	case UploadTypeThumbnail:
		return storage.BucketThumbnails
	default:
		return storage.BucketDocuments // Default bucket
	}
}

func isAllowedContentType(uploadType UploadType, contentType string) bool {
	// Image types
	imageTypes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/svg+xml": true,
	}

	// Document types
	docTypes := map[string]bool{
		"application/pdf":              true,
		"application/zip":              true,
		"application/x-zip-compressed": true,
		"text/plain":                   true,
		"application/json":             true,
		"application/octet-stream":     true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	}

	switch uploadType {
	case UploadTypeAvatar, UploadTypeThumbnail:
		return imageTypes[contentType]
	case UploadTypeComponent, UploadTypeProject:
		// Allow images and some documents
		return imageTypes[contentType] || strings.HasPrefix(contentType, "application/json")
	case UploadTypeDocument:
		return docTypes[contentType] || imageTypes[contentType]
	case UploadTypeGeneral:
		// Allow most common types
		return imageTypes[contentType] || docTypes[contentType]
	}
	return false
}
