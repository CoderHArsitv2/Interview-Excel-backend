package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	SessionUUID string         `gorm:"uniqueIndex" json:"session_uuid"`

	ExpertUUID  string `gorm:"index;not null" json:"expert_uuid"`
	StudentUUID string `gorm:"index;not null" json:"student_uuid"`

	SlotID uint `gorm:"index" json:"slot_id"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	MeetLink string `json:"meet_link,omitempty"`

	Status string `gorm:"default:'scheduled';index" json:"status"`
}

type SessionRepo struct {
	db *gorm.DB
}

func (r *SessionRepo) Create(session *Session) error {
	return r.db.Create(session).Error
}

func (r *SessionRepo) GetByUUID(sessionUUID string) (*Session, error) {
	var session Session
	err := r.db.
		Where("session_uuid = ?", sessionUUID).
		First(&session).Error

	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepo) GetByStudentUUID(studentUUID string) ([]Session, error) {
	var sessions []Session
	err := r.db.
		Where("student_uuid = ?", studentUUID).
		Order("start_time ASC").
		Find(&sessions).Error

	return sessions, err
}

func (r *SessionRepo) GetByExpertUUID(expertUUID string) ([]Session, error) {
	var sessions []Session
	err := r.db.
		Where("expert_uuid = ?", expertUUID).
		Order("start_time ASC").
		Find(&sessions).Error

	return sessions, err
}

func (r *SessionRepo) UpdateStatus(sessionUUID string, status string) error {
	result := r.db.
		Model(&Session{}).
		Where("session_uuid = ?", sessionUUID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *SessionRepo) Cancel(sessionUUID string) error {
	return r.UpdateStatus(sessionUUID, "cancelled")
}

func (r *SessionRepo) MarkCompleted(sessionUUID string) error {
	return r.UpdateStatus(sessionUUID, "completed")
}

func (r *SessionRepo) ExistsForSlot(slotID uint) (bool, error) {
	var count int64
	err := r.db.
		Model(&Session{}).
		Where("slot_id = ? AND status IN ?", slotID, []string{"scheduled"}).
		Count(&count).Error

	return count > 0, err
}

func (r *SessionRepo) GetUpcomingForUser(userUUID string) ([]Session, error) {
	var sessions []Session

	err := r.db.
		Where(
			"(student_uuid = ? OR expert_uuid = ?) AND start_time > NOW() AND status = ?",
			userUUID, userUUID, "scheduled",
		).
		Order("start_time ASC").
		Find(&sessions).Error

	return sessions, err
}

func (r *SessionRepo) Delete(sessionUUID string) error {
	result := r.db.
		Where("session_uuid = ?", sessionUUID).
		Delete(&Session{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
