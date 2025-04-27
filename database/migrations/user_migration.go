package migrations

import (
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func MigrateUsersTable(db *gorm.DB) error {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		utils.Log.Error("Failed to migrate users table", zap.Error(err))
		return err
	}

	utils.SLog.Info("Users table migrated successfully")
	return nil
}
