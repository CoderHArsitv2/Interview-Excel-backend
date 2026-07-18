package models

import (
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Student struct {
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       string         `gorm:"uniqueIndex" json:"user_uuid"` // references User.UserUUID
	Bio          string         `json:"bio,omitempty"`
	Sessions     string         `json:"sessions"`
	Points       string         `json:"points"`
	PreparingFor string         `json:"preparing_for"`
	DateOfBirth  time.Time      `json:"dob"`
	City         string         `json:"city"`
	AboutMe      string         `json:"about_me"`
	Skills       datatypes.JSON `json:"skills"` // JSON column for skills
}

type StudentRepo struct {
	db *gorm.DB
}

// Create a new student
func (r *StudentRepo) Create(student *Student) error {
	return r.db.Create(student).Error
}

// Get by user UUID
func (r *StudentRepo) GetByUserUUID(uuid string) (*Student, error) {
	var student Student
	err := r.db.Where("user_id = ?", uuid).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}

var ErrStudentNotFound = errors.New("student not found")

func (r *StudentRepo) UpdateByUserUUID(userUUID string, updates map[string]interface{}) error {
	result := r.db.
		Model(&Student{}).
		Where("user_id = ?", userUUID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrStudentNotFound
	}

	return nil
}

// Delete student (by user UUID)
func (r *StudentRepo) DeleteByUserUUID(uuid string) error {
	return r.db.Where("user_id = ?", uuid).Delete(&Student{}).Error
}

// List all students
func (r *StudentRepo) ListAll() ([]Student, error) {
	var students []Student
	err := r.db.Find(&students).Error
	return students, err
}
