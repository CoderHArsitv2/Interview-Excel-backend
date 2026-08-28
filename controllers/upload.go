package controllers

import (
	"errors"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"interviewexcel-backend-go/config"
	"interviewexcel-backend-go/models"
	logger "interviewexcel-backend-go/pkg/errors"
	"interviewexcel-backend-go/pkg/storage"
)

// UploadImage handles multipart image uploads to Cloudflare R2 and returns the
// public URL. The returned URL can then be sent as profile_picture_url in the
// update-profile calls.
//
// The object key is namespaced by environment and role, e.g. "local/expert/<uuid>.jpg".
func UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error("failed to read upload file: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required (multipart field 'file')"})
		return
	}
	defer file.Close()

	// Build folder as {environment}/{role}, e.g. "local/expert".
	ownerUUID, _ := c.Get("user_uuid")
	ownerUUIDStr, _ := ownerUUID.(string)
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr == "" {
		roleStr = "unknown"
	}
	folder := path.Join(config.RuntimeConfig().StorageEnv, roleStr)

	result, err := storage.UploadImage(c.Request.Context(), folder, file, header)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, storage.ErrFileTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		case errors.Is(err, storage.ErrNotConfigured), errors.Is(err, storage.ErrPublicURLUnset):
			logger.Error("R2 not configured: ", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image upload is not available"})
		default:
			logger.Error("failed to upload image: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
		}
		return
	}

	// Persist the upload record. We store the object key (path), never a URL.
	upload := &models.UserUpload{
		UUID:      "upl_" + uuid.NewString(),
		FileKey:   result.Key,
		Bucket:    result.Bucket,
		FileType:  result.ContentType,
		FileName:  header.Filename,
		Category:  models.CategoryProfilePicture,
		Status:    models.FileStatusActive,
		OwnerUUID: ownerUUIDStr,
		OwnerType: roleStr,
	}
	if err := models.InitUserUploadRepo(config.DB).Create(upload); err != nil {
		logger.Error("failed to persist upload record: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save upload"})
		return
	}

	// Return a fresh presigned URL for immediate display.
	upload.FileURL = storage.PresignStored(c.Request.Context(), result.Key)

	c.JSON(http.StatusOK, upload)
}
