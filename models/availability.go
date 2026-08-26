package models

import (
	"time"

	logger "interviewexcel-backend-go/pkg/errors"

	"gorm.io/gorm"
)

type AvailabilitySlotStatus string

const (
	SlotAvailable AvailabilitySlotStatus = "AVAILABLE"
	SlotBooked    AvailabilitySlotStatus = "BOOKED"
	SlotCancelled AvailabilitySlotStatus = "CANCELLED"
)

type AvailabilitySlot struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ExpertID  string         `gorm:"not null;index" json:"expert_id"` // Expert.UserID
	Expert    Expert         `gorm:"foreignKey:ExpertID;references:UserID" json:"-"`

	Date      time.Time `gorm:"not null;index" json:"date"`
	StartTime time.Time `gorm:"not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`

	Status    string `gorm:"type:varchar(20);default:'AVAILABLE';index" json:"status"`
	StudentID *uint  `gorm:"index" json:"student_id,omitempty"`
}

type availabilitySlotRepo struct {
	DB *gorm.DB
}

func (r *availabilitySlotRepo) CreateAvailabilitySlot(availability []AvailabilitySlot) error {
	return r.DB.Create(&availability).Error
}

// Get all slots for an expert
func (r *availabilitySlotRepo) GetAllByExpert(expertID string) ([]AvailabilitySlot, error) {
	var slots []AvailabilitySlot
	err := r.DB.Where("expert_id = ?", expertID).Find(&slots).Error
	return slots, err
}

// Get slots for an expert filtered by time window.
// filter: "upcoming" (start_time >= now), "past" (start_time < now), anything else -> all.
func (r *availabilitySlotRepo) GetSlotsByExpertFiltered(expertID, filter string) ([]AvailabilitySlot, error) {
	query := r.DB.Where("expert_id = ?", expertID)

	switch filter {
	case "upcoming":
		query = query.Where("start_time >= ?", time.Now())
	case "past":
		query = query.Where("start_time < ?", time.Now())
	}

	var slots []AvailabilitySlot
	err := query.Order("start_time ASC").Find(&slots).Error
	return slots, err
}

// Get all available (not booked) slots
func (r *availabilitySlotRepo) GetAvailableByExpert(expertID string) ([]AvailabilitySlot, error) {
	var slots []AvailabilitySlot
	err := r.DB.Where("expert_id = ? AND status = ? AND start_time >= ?", expertID, string(SlotAvailable), time.Now()).
		Order("date ASC, start_time ASC").
		Find(&slots).Error
	return slots, err
}

// Get slot by ID
func (r *availabilitySlotRepo) GetByID(id uint) (*AvailabilitySlot, error) {
	var slot AvailabilitySlot
	err := r.DB.First(&slot, id).Error
	return &slot, err
}

// Mark slot as booked
func (r *availabilitySlotRepo) MarkAsBooked(id uint) error {
	return r.DB.Model(&AvailabilitySlot{}).Where("id = ?", id).Update("status", string(SlotBooked)).Error
}

// Delete a slot
func (r *availabilitySlotRepo) Delete(id uint) error {
	return r.DB.Delete(&AvailabilitySlot{}, id).Error
}

// Update a slot (useful for admin panel or expert updates)
func (r *availabilitySlotRepo) Update(slot *AvailabilitySlot) error {
	return r.DB.Save(slot).Error
}
func (r *availabilitySlotRepo) UpdateWithTx(tx *gorm.DB, slot *AvailabilitySlot, where *AvailabilitySlot) error {

	err := tx.Model(&AvailabilitySlot{}).
		Where(where).Updates(slot).Error
	if err != nil {
		logger.Error("error in updating availability slot: ", err)
		return err
	}
	return nil
}
func (r *availabilitySlotRepo) GetBookedByStudent(studentID uint) ([]AvailabilitySlot, error) {
	var slots []AvailabilitySlot
	err := r.DB.
		Where("student_id = ? AND status = ? AND date >= ?", studentID, string(SlotBooked), time.Now()).
		Order("date ASC").
		Find(&slots).Error
	return slots, err
}

func (r *availabilitySlotRepo) GetBookedSlotsByExpert(expertID uint) ([]AvailabilitySlot, error) {
	var slots []AvailabilitySlot
	err := r.DB.
		Where("expert_id = ? AND status = ? AND date >= ?", expertID, string(SlotBooked), time.Now()).
		Order("date ASC, start_time ASC").
		Find(&slots).Error
	return slots, err
}

func (r *availabilitySlotRepo) CountAvailableSlotsByExpert(expertID string) (int64, error) {
	var count int64
	err := r.DB.Model(&AvailabilitySlot{}).
		Where("expert_id = ? AND status = ?", expertID, string(SlotAvailable)).
		Count(&count).Error
	return count, err
}

func (r *availabilitySlotRepo) CountBookedSlotsByExpertUUID(expertID string) (int64, error) {
	var count int64
	err := r.DB.Model(&AvailabilitySlot{}).
		Where("expert_id = ? AND status = ?", expertID, string(SlotBooked)).
		Count(&count).Error
	return count, err
}
