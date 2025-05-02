package migrations

import (
	"strings"
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func MigrateUsersTable(db *gorm.DB) error {
	// Enum tipini oluştur
	err := db.Exec(`CREATE TYPE user_type AS ENUM ('system', 'panel')`).Error
	if err != nil && !isAlreadyExistsError(err) {
		utils.Log.Error("Failed to create user_type enum", zap.Error(err))
		return err
	}
	utils.SLog.Debug("Checked/created user_type enum")

	// Users tablosunu migrate et
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		utils.Log.Error("Failed to migrate users table structure", zap.Error(err))
		return err
	}
	utils.SLog.Info("Users table structure migrated successfully")

	return nil
}

// PostgreSQL enum tipi zaten varsa kontrol eder
func isAlreadyExistsError(err error) bool {
	return strings.Contains(err.Error(), "already exists")
}
