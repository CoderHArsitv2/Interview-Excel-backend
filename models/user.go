package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserUUID  string         `gorm:"uniqueIndex" json:"user_uuid"` // make it unique if used for relation
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	FullName string  `gorm:"not null" json:"full_name"`
	Email    string  `gorm:"not null;uniqueIndex" json:"email"`
	Picture  string  `json:"picture"`
	Password *string `json:"-"`
	Phone    *string `gorm:"uniqueIndex" json:"phone,omitempty"`
	Role     string  `gorm:"not null" json:"role"` // "student" | "expert" | "admin"

	// Relations
	Student *Student `gorm:"foreignKey:UserID;references:UserUUID" json:"student,omitempty"`
	Expert  *Expert  `gorm:"foreignKey:UserID;references:UserUUID" json:"expert,omitempty"`
}
type UserRepo struct {
	db *gorm.DB
}

func InitUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create a new user
func (r *UserRepo) Create(user *User) error {
	return r.db.Create(user).Error
}

// Find user by ID
func (r *UserRepo) GetByUUID(id string) (*User, error) {
	var user User
	if err := r.db.Where("user_uuid=?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Find user by email
func (r *UserRepo) GetByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Find user by phone
func (r *UserRepo) GetByPhone(phone string) (*User, error) {
	var user User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update user
func (r *UserRepo) UpdateByUserUUID(uuid string, updates *User) error {
	return r.db.Model(&User{}).
		Where("user_uuid = ?", uuid).
		Updates(updates).Error
}

// Delete user
func (r *UserRepo) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

// List all users (optionally by role)
func (r *UserRepo) List(role *string) ([]User, error) {
	var users []User
	query := r.db
	if role != nil {
		query = query.Where("role = ?", *role)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Exists check (by email or phone)
func (r *UserRepo) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepo) ExistsByPhone(phone string) (bool, error) {
	var count int64
	if err := r.db.Model(&User{}).Where("phone = ?", phone).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// WithPreload loads student/expert details along with user
func (r *UserRepo) GetWithRelations(id uint) (*User, error) {
	var user User
	if err := r.db.Preload("Student").Preload("Expert").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Utility for login: find by email + load password
func (r *UserRepo) GetForAuth(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
