package initializers

import "github.com/ElissaGunawan/moneywallbackend/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Post{})
	DB.AutoMigrate(&models.Account{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Income{})
	DB.AutoMigrate(&models.Expense{})
	DB.AutoMigrate(&models.Quest{})
	DB.AutoMigrate(&models.Achievement{})
	DB.AutoMigrate(&models.Avatar{})
}
