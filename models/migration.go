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
	&UserUpload{},
}

func GetMigrationModel() []interface{} {
	return modelsForMigration
}
