package migrations

import (
	"errors"
	"zatrano/models"
	"zatrano/utils"

	"gorm.io/gorm"
)

func MigrateUsersTable(db *gorm.DB) error {
	utils.SLog.Info("User tablosu için enum tipi kontrol ediliyor...")

	dropEnumQuery := `DROP TYPE IF EXISTS user_type;`
	rawDB, err := db.DB()
	if err != nil {
		return errors.New("DB instance alınamadı: " + err.Error())
	}

	_, err = rawDB.Exec(dropEnumQuery)
	if err != nil {
		return errors.New("user_type enum silinemedi: " + err.Error())
	}
	utils.SLog.Info("user_type enum başarıyla silindi.")

	createEnum := `CREATE TYPE user_type AS ENUM ('system', 'panel');`
	_, err = rawDB.Exec(createEnum)
	if err != nil {
		return errors.New("user_type enum oluşturulamadı: " + err.Error())
	}
	utils.SLog.Info("user_type enum başarıyla oluşturuldu.")

	utils.SLog.Info("User tablosu migrate ediliyor...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return errors.New("User tablosu migrate edilemedi: " + err.Error())
	}

	utils.SLog.Info("User tablosu migrate işlemi tamamlandı.")
	return nil
}
