package migrations

import (
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func MigrateUsersTable(db *gorm.DB) error {
	err := db.Exec(`DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type') THEN
			CREATE TYPE user_type AS ENUM ('system', 'panel');
		END IF;
	END$$`).Error
	if err != nil {
		utils.Log.Error("Failed to create/check user_type enum", zap.Error(err))
		return err
	}
	utils.SLog.Debug("Checked/created user_type enum")

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		utils.Log.Error("Failed to migrate users table structure", zap.Error(err))
		return err
	}
	utils.SLog.Info("Users table structure migrated successfully")

	return nil
}
