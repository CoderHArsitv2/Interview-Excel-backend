package models

import (
	"time"

	"gorm.io/gorm"
)

type Wallet struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserUUID  string         `gorm:"uniqueIndex;not null"`

	BalanceInPaise int64 `gorm:"not null;default:0"`

	Transactions []WalletTransaction `gorm:"foreignKey:WalletID;references:ID"`
}

type walletRepo struct {
	DB *gorm.DB
}

func (r *walletRepo) GetByUserUUID(userUUID string) (*Wallet, error) {
	var wallet Wallet
	err := r.DB.Where("user_uuid = ?", userUUID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepo) Create(wallet *Wallet) error {
	return r.DB.Create(wallet).Error
}

func (r *walletRepo) UpdateBalance(userUUID string, newBalanceInPaise int64) error {
	return r.DB.Model(&Wallet{}).Where("user_uuid = ?", userUUID).
		Update("balance_in_paise", newBalanceInPaise).Error
}
