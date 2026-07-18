package models

import "gorm.io/gorm"

func InitExpertRepo(db *gorm.DB) IExpert {
	return &expertRepo{
		DB: db,
	}
}

func InitStudentRepo(db *gorm.DB) *StudentRepo {
	return &StudentRepo{db: db}
}

func InitAvailabilitySlotRepo(db *gorm.DB) IAvailabilitySlotRepo {
	return &availabilitySlotRepo{
		DB: db,
	}
}

func InitSessionRepo(db *gorm.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func InitWalletRepo(db *gorm.DB) *walletRepo {
	return &walletRepo{DB: db}
}

func InitWalletTransactionRepo(db *gorm.DB) *walletTransactionRepo {
	return &walletTransactionRepo{DB: db}
}
