package models

import "gorm.io/gorm"

type IExpert interface {
	Create(s *Expert) error
	CreateWithTx(tx *gorm.DB, s *Expert) error
	GetAllExpert(provider *Expert) (*[]Expert, error)
	GetByID(providerID uint64) (*Expert, error)
	GetWithTx(tx *gorm.DB, where *Expert) (*Expert, error)
	Update(where *Expert, a *Expert) error
	UpdateWithTx(tx *gorm.DB, where *Expert, a *Expert) error
	Delete(where uint64) error
	GetAll() ([]Expert, error)
	GetAllExpertsWithUserDetails() ([]Expert, error)
}

type IAvailabilitySlotRepo interface {
	CreateAvailabilitySlot(availability []AvailabilitySlot) error
	GetAllByExpert(expertID string) ([]AvailabilitySlot, error)
	GetAvailableByExpert(expertID string) ([]AvailabilitySlot, error)
	GetByID(id uint) (*AvailabilitySlot, error)
	MarkAsBooked(id uint) error
	Delete(id uint) error
	Update(slot *AvailabilitySlot) error
	GetBookedByStudent(studentID uint) ([]AvailabilitySlot, error)
	UpdateWithTx(tx *gorm.DB, slot *AvailabilitySlot, where *AvailabilitySlot) error
	GetBookedSlotsByExpert(expertID uint) ([]AvailabilitySlot, error)
	CountAvailableSlotsByExpert(expertID string) (int64, error)
	CountBookedSlotsByExpertUUID(expertID string) (int64, error)
}

type IWalletRepo interface {
	GetByUserUUID(userUUID string) (*Wallet, error)
	Create(wallet *Wallet) error
	UpdateBalance(userUUID string, newBalanceInPaise int64) error
}

type IWalletTransactionRepo interface {
	Create(tx *gorm.DB, wt *WalletTransaction) error
	GetByWalletID(walletID uint) ([]WalletTransaction, error)
	GetByReferenceID(referenceID string) ([]WalletTransaction, error)
}

type IPaymentRepo interface {
	Create(payment *Payment) error
	GetByOrderID(orderID string) (*Payment, error)
}

type IStudent interface {
	Create(student *Student) error
	GetByUserUUID(uuid string) (*Student, error)
	UpdateByUserUUID(uuid string, updates map[string]interface{}) error
	DeleteByUserUUID(uuid string) error
	ListAll() ([]Student, error)
}

type ISession interface {
	Create(session *Session) error
	GetByUUID(sessionUUID string) (*Session, error)
	GetByStudentUUID(studentUUID string) ([]Session, error)
	GetByExpertUUID(expertUUID string) ([]Session, error)
	GetUpcomingForUser(userUUID string) ([]Session, error)
	ExistsForSlot(slotID uint) (bool, error)
	UpdateStatus(sessionUUID string, status string) error
	Cancel(sessionUUID string) error
	MarkCompleted(sessionUUID string) error
	Delete(sessionUUID string) error
}
type IUser interface {
	InitUserRepo(db *gorm.DB) *UserRepo
	Create(user *User) error
	GetByUUID(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
	ListAll() ([]User, error)
}
