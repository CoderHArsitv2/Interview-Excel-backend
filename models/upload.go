package models

import (
	"time"

	"gorm.io/gorm"
)

type FileUploadCategory string

const (
	CategoryProfilePicture FileUploadCategory = "PROFILE_PICTURE"
)

type FileStatus string

const (
	FileStatusActive  FileStatus = "ACTIVE"
	FileStatusDeleted FileStatus = "DELETED"
)

// UserUpload records every asset uploaded to object storage. We persist the
// object key (path) — never a full URL — so the URL can be (re)generated from
// the bucket on demand, e.g. as a presigned URL.
type UserUpload struct {
	ID        uint           `json:"-" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UUID     string             `json:"uuid" gorm:"uniqueIndex;not null"`
	FileKey  string             `json:"file_key" gorm:"not null;index"` // object key/path within the bucket
	Bucket   string             `json:"-"`                              // bucket the object lives in; used to presign/delete
	FileType string             `json:"file_type"`                      // MIME type
	FileName string             `json:"file_name"`                      // original uploaded file name
	Category FileUploadCategory `json:"category" gorm:"index"`
	Status   FileStatus         `json:"status" gorm:"index;default:'ACTIVE'"`

	// Owner is a platform user identified by their UserUUID; OwnerType is the role.
	OwnerUUID string `json:"owner_uuid" gorm:"index;not null"`
	OwnerType string `json:"owner_type"`

	// FileURL is not persisted. It is populated on read with a presigned URL.
	FileURL string `json:"file_url,omitempty" gorm:"-"`
}

type UserUploadRepo struct {
	db *gorm.DB
}

func InitUserUploadRepo(db *gorm.DB) *UserUploadRepo {
	return &UserUploadRepo{db: db}
}

func (r *UserUploadRepo) Create(upload *UserUpload) error {
	return r.db.Create(upload).Error
}

func (r *UserUploadRepo) GetByUUID(uuid string) (*UserUpload, error) {
	var upload UserUpload
	if err := r.db.Where("uuid = ?", uuid).First(&upload).Error; err != nil {
		return nil, err
	}
	return &upload, nil
}

func (r *UserUploadRepo) GetByKey(fileKey string) (*UserUpload, error) {
	var upload UserUpload
	if err := r.db.Where("file_key = ?", fileKey).First(&upload).Error; err != nil {
		return nil, err
	}
	return &upload, nil
}

func (r *UserUploadRepo) ListByOwner(ownerUUID string) ([]UserUpload, error) {
	var uploads []UserUpload
	err := r.db.
		Where("owner_uuid = ? AND status = ?", ownerUUID, string(FileStatusActive)).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
}

// MarkDeleted soft-flags the record as deleted (status only; row is retained).
func (r *UserUploadRepo) MarkDeleted(uuid string) error {
	result := r.db.Model(&UserUpload{}).
		Where("uuid = ?", uuid).
		Update("status", string(FileStatusDeleted))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetLatestByOwnerAndCategory returns the most recent active upload of a given
// category for an owner (e.g. the current profile picture of an expert).
func (r *UserUploadRepo) GetLatestByOwnerAndCategory(ownerUUID string, category FileUploadCategory) (*UserUpload, error) {
	var upload UserUpload
	err := r.db.
		Where("owner_uuid = ? AND category = ? AND status = ?", ownerUUID, string(category), string(FileStatusActive)).
		Order("created_at DESC").
		First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}
