package models

var modelsForMigration = []interface{}{
	&User{},
	&Expert{},
	&AvailabilitySlot{},
	&Payment{},
	&Student{},
	&Session{},
	&Wallet{},
	&WalletTransaction{},
}

func GetMigrationModel() []interface{} {
	return modelsForMigration
}
