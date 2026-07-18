package models

import (
	"gorm.io/gorm"
	"time"
)

type Payment struct {
	gorm.Model

	OrderID   string `gorm:"uniqueIndex" json:"order_id"`
	PaymentID string `json:"payment_id,omitempty"`
	Status    string `json:"status"` // created, paid, failed

	StudentID uint `json:"student_id"`
	ExpertID  uint `json:"expert_id"`
	SlotID    uint `json:"slot_id"`

	Amount      uint `json:"amount"`       // in paise
	PlatformFee uint `json:"platform_fee"` // in paise
	ExpertShare uint `json:"expert_share"` // in paise

	Currency string     `json:"currency"` // INR
	Method   string     `json:"method"`   // upi, card, etc.
	PaidAt   *time.Time `json:"paid_at,omitempty"`
}

type paymentRepo struct {
	DB *gorm.DB
}

func InitPaymentRepo(db *gorm.DB) *paymentRepo {
	return &paymentRepo{DB: db}
}

func (r *paymentRepo) Create(payment *Payment) error {
	return r.DB.Create(payment).Error
}

func (r *paymentRepo) GetByOrderID(orderID string) (*Payment, error) {
	var payment Payment
	err := r.DB.Where("order_id = ?", orderID).First(&payment).Error
	return &payment, err
}

func (r *paymentRepo) Update(payment *Payment) error {
	return r.DB.Save(payment).Error
}
