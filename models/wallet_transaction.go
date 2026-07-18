package models

import (
	"time"

	"gorm.io/gorm"
)

type WalletTransaction struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	WalletID  uint           `gorm:"index;not null"`

	Wallet *Wallet `gorm:"foreignKey:WalletID;references:ID;constraint:OnDelete:CASCADE;"`

	AmountInPaise int64
	Type          string // credit | debit
	Source        string // session | refund | payout
	ReferenceID   string

	Description string
}

type walletTransactionRepo struct {
	DB *gorm.DB
}

func (r *walletTransactionRepo) Create(tx *gorm.DB, wt *WalletTransaction) error {
	return tx.Create(wt).Error
}

func (r *walletTransactionRepo) GetByWalletID(walletID uint) ([]WalletTransaction, error) {
	var transactions []WalletTransaction
	err := r.DB.Where("wallet_id = ?", walletID).Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *walletTransactionRepo) GetByReferenceID(referenceID string) ([]WalletTransaction, error) {
	var transactions []WalletTransaction
	err := r.DB.Where("reference_id = ?", referenceID).Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
