package controllers

import "time"

type SignUpRequest struct {
	FullName        string `json:"full_name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Phone           string `json:"phone"`
	Password        string `json:"password" binding:"required"`
	ConfirrPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	Role            string `json:"role" binding:"required,oneof=student expert"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GoogleAuthRequest struct {
	Role  string `json:"role" binding:"required"`
	Token string `json:"token" binding:"required"`
}

type StudentProfile struct {
	UserID string `json:"user_uuid"`
	Role   string `json:"role"`

	// from UserRepo
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`

	Bio          string    `json:"bio,omitempty"`
	Sessions     string    `json:"sessions"`
	Points       string    `json:"points"`
	PreparingFor string    `json:"preparing_for"`
	DateOfBirth  time.Time `json:"dob"`
	City         string    `json:"city"`
	AboutMe      string    `json:"about_me"`
	Skills       []string  `gorm:"type:json" json:"skills"` // JSON column for skills
}

type ExpertProfile struct {
	// From User
	UserID   uint    `json:"id"`
	UserUUID string  `json:"user_uuid"`
	FullName string  `json:"full_name"`
	Email    string  `json:"email"`
	Picture  string  `json:"picture"`
	Phone    *string `json:"phone"`
	Role     string  `json:"role"`
	City     string  `json:"city"`

	// From Expert
	Bio                string    `json:"bio"`
	DOB                time.Time `json:"dob"`
	Languages          []string  `json:"languages"`
	Specializations    []string  `json:"specializations"`
	Expertise          string    `json:"expertise"`
	Education          string    `json:"education"`
	ExperienceYears    int       `json:"experience_years"`
	ProfilePictureUrl  string    `json:"profile_picture_url"`
	FeesPerSession     int       `json:"fees_per_session"`
	Skills             string    `json:"skills"`
	Achievements       string    `json:"achievements"`
	Rating             float64   `json:"rating"` // if you added
	TotalSessions      int       `json:"total_sessions"`
	VerificationStatus string    `json:"verification_status"`
	IsAvailable        bool      `json:"is_available"`
	StudentMentored    int64     `json:"student_mentored"`
}

type AvailabilityRequest struct {
	ExpertID string   `json:"expert_id"`
	Days     []string `json:"days"`       // ["monday","wednesday","friday"]
	Start    string   `json:"start_time"` // "10:00"
	End      string   `json:"end_time"`   // "14:00"
	SlotSize int      `json:"duration"`   // minutes, e.g. 60
}

type Slot struct {
	ExpertID  string    `json:"expert_id"`
	Date      time.Time `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	IsBooked  bool      `json:"is_booked"`
}

type StudentSessionResponse struct {
	ID                uint      `json:"id"`
	SessionUUID       string    `json:"session_uuid"`
	ExpertUUID        string    `json:"expert_uuid"`
	ExpertName        string    `json:"expert_name"`
	ProfilePictureUrl string    `json:"profile_picture_url"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	MeetLink          string    `json:"meet_link,omitempty"`
	Status            string    `json:"status"`
}

// Expert Dashboard response types

type ExpertDashboardResponse struct {
	Expert           DashboardExpertInfo     `json:"expert"`
	Stats            DashboardStats          `json:"stats"`
	UpcomingSessions []ExpertSessionResponse `json:"upcoming_sessions"`
	SlotOverview     DashboardSlotOverview   `json:"slot_overview"`
}

type DashboardExpertInfo struct {
	FullName           string `json:"full_name"`
	VerificationStatus string `json:"verification_status"`
	IsAvailable        bool   `json:"is_available"`
	ProfilePictureUrl  string `json:"profile_picture_url"`
}

type DashboardStats struct {
	TotalSessions    int     `json:"total_sessions"`
	StudentsMentored int64   `json:"students_mentored"`
	Rating           float64 `json:"rating"`
	Earnings         int64   `json:"earnings"`
}

type DashboardSlotOverview struct {
	AvailableSlots int64 `json:"available_slots"`
	BookedSlots    int64 `json:"booked_slots"`
	SessionFee     int   `json:"session_fee"`
}

type ExpertSessionResponse struct {
	ID          uint      `json:"id"`
	SessionUUID string    `json:"session_uuid"`
	StudentUUID string    `json:"student_uuid"`
	StudentName string    `json:"student_name"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	MeetLink    string    `json:"meet_link,omitempty"`
	Status      string    `json:"status"`
}
