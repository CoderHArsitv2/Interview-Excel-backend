package storage

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"interviewexcel-backend-go/config"
)

// MaxImageSize is the maximum accepted upload size (5 MB).
const MaxImageSize = 5 << 20

// PresignTTL is the default lifetime of a presigned profile-picture URL.
const PresignTTL = 24 * time.Hour

// allowedImageTypes maps accepted MIME types to their file extension.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

var (
	ErrNotConfigured  = errors.New("R2 storage is not configured")
	ErrFileTooLarge   = fmt.Errorf("file exceeds max size of %d bytes", MaxImageSize)
	ErrInvalidType    = errors.New("unsupported image type (allowed: jpeg, png, webp, gif)")
	ErrPublicURLUnset = errors.New("R2 public base URL is not configured")
)

// UploadResult describes a successfully uploaded object.
type UploadResult struct {
	Key         string // object key/path within the bucket
	Bucket      string
	URL         string // public URL (R2PublicBaseURL + key)
	ContentType string
}

// UploadImage uploads an image file to R2 under the given folder. It validates
// the content type and size and returns the object key plus its public URL.
func UploadImage(ctx context.Context, folder string, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	if config.R2Client == nil {
		return nil, ErrNotConfigured
	}
	if config.R2PublicBaseURL == "" {
		return nil, ErrPublicURLUnset
	}
	if header.Size > MaxImageSize {
		return nil, ErrFileTooLarge
	}

	contentType := header.Header.Get("Content-Type")
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return nil, ErrInvalidType
	}

	key := path.Join(strings.Trim(folder, "/"), uuid.NewString()+ext)

	if _, err := config.R2Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.R2Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	}); err != nil {
		return nil, fmt.Errorf("failed to upload to R2: %w", err)
	}

	return &UploadResult{
		Key:         key,
		Bucket:      config.R2Bucket,
		URL:         config.R2PublicBaseURL + "/" + key,
		ContentType: contentType,
	}, nil
}

// KeyFromStored derives the R2 object key from a stored value. It accepts either
// a full public URL (e.g. https://pub-xxx.r2.dev/local/expert/x.jpg) or a bare
// key. Returns "" if the value is empty or is an external URL we don't own.
func KeyFromStored(stored string) string {
	if stored == "" {
		return ""
	}
	if config.R2PublicBaseURL != "" && strings.HasPrefix(stored, config.R2PublicBaseURL+"/") {
		return strings.TrimPrefix(stored, config.R2PublicBaseURL+"/")
	}
	// Not a public URL of ours: treat a non-URL value as an already-bare key.
	if strings.Contains(stored, "://") {
		return ""
	}
	return strings.TrimLeft(stored, "/")
}

// PresignGetURL returns a time-limited signed GET URL for the given object key.
func PresignGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if config.R2Client == nil {
		return "", ErrNotConfigured
	}
	presign := s3.NewPresignClient(config.R2Client)
	req, err := presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(config.R2Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign R2 object: %w", err)
	}
	return req.URL, nil
}

// PresignStored converts a stored profile-picture value into a presigned URL.
// It is best-effort: if the value is empty, not one of our objects, or storage
// is unconfigured, the original value is returned unchanged.
func PresignStored(ctx context.Context, stored string) string {
	key := KeyFromStored(stored)
	if key == "" {
		return stored
	}
	url, err := PresignGetURL(ctx, key, PresignTTL)
	if err != nil {
		return stored
	}
	return url
}
