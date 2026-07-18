package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Expert struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID    string         `gorm:"uniqueIndex" json:"user_uuid"` // ✅ string identifier
	User      *User          `gorm:"foreignKey:UserID;references:UserUUID" json:"user,omitempty"`

	FullName          string         `json:"full_name"`
	Bio               string         `json:"bio,omitempty"`
	Expertise         string         `json:"expertise"`
	Specializations   pq.StringArray `gorm:"type:text[]" json:"specializations,omitempty"`
	ExperienceYears   int            `json:"experience_years"`
	Education         string         `json:"education,omitempty"`
	Languages         pq.StringArray `gorm:"type:text[]" json:"languages,omitempty"`
	ProfilePictureUrl string         `json:"profile_picture_url,omitempty"`
	FeesPerSession    int            `json:"fees_per_session"`
	City              string         `json:"city"`
	Achievement       string         `json:"achievement"`
	DOB               time.Time      `json:"dob"`

	Rating             float64 `gorm:"default:0" json:"rating"`
	TotalSessions      int     `gorm:"default:0" json:"total_sessions"`
	VerificationStatus string  `gorm:"default:'pending'" json:"verification_status"`
	StudentMentored    int64   `gorm:"default:0" json:"student_mentored"`
	IsAvailable        bool    `gorm:"default:true" json:"is_available"`

	AvailabilitySlots []AvailabilitySlot `gorm:"foreignKey:ExpertID;references:UserID" json:"availability_slots,omitempty"`
}

type expertRepo struct {
	DB *gorm.DB
}

func (e *expertRepo) Create(s *Expert) error {

	return e.CreateWithTx(e.DB, s)
}
func (e *expertRepo) CreateWithTx(tx *gorm.DB, s *Expert) error {

	err := tx.Create(s).Error
	if err != nil {
		return err
	}
	return nil
}
func (e *expertRepo) GetAllExpert(ex *Expert) (*[]Expert, error) {
	experts := []Expert{}
	err := e.DB.Where(ex).Find(&experts).Error
	if err != nil {
		return nil, err
	}
	return &experts, nil
}

func (e *expertRepo) GetByID(ID uint64) (*Expert, error) {
	var expert Expert
	err := e.DB.First(&expert, ID).Error
	if err != nil {
		return nil, err
	}
	return &expert, nil
}

func (e *expertRepo) GetWithTx(tx *gorm.DB, where *Expert) (*Expert, error) {
	var expert Expert
	err := tx.Where(where).First(&expert).Error
	if err != nil {
		return nil, err
	}
	return &expert, nil
}

func (e *expertRepo) Update(where *Expert, update *Expert) error {
	return e.UpdateWithTx(e.DB, where, update)
}

func (e *expertRepo) UpdateWithTx(tx *gorm.DB, where *Expert, update *Expert) error {
	err := tx.Model(&Expert{}).Where(&where).Updates(&update).Error
	if err != nil {
		return err
	}
	return nil
}

func (e *expertRepo) Delete(id uint64) error {
	err := e.DB.Delete(&Expert{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *expertRepo) GetAll() ([]Expert, error) {
	var experts []Expert
	err := r.DB.Find(&experts).Error
	return experts, err
}

func (r *expertRepo) GetAllExpertsWithUserDetails() ([]Expert, error) {
	var experts []Expert
	err := r.DB.
		Preload("User").
		Find(&experts).Error
	return experts, err
}
