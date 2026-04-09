package migration

import (
	"somewebproject/internal/models"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Product{})
}
