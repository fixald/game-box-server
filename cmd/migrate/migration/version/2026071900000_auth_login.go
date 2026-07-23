package version

import (
	"time"

	"go-admin/app/cauth/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026071900000", migrateAuthLogin) }

func migrateAuthLogin(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			&models.User{},
			&models.Agreement{},
			&models.PasswordReset{},
			&models.DeviceAccount{},
		); err != nil {
			return err
		}
		var n int64
		tx.Model(&models.Agreement{}).Where("version = ?", "2026-07-01").Count(&n)
		if n == 0 {
			now := time.Now().UTC()
			if err := tx.Create(&models.Agreement{
				Version:     "2026-07-01",
				Title:       "用户协议与隐私政策",
				ContentURL:  "https://cdn.example.com/agreements/2026-07-01.html",
				Summary:     "登录即表示同意最新用户协议与隐私政策",
				Status:      "published",
				PublishedAt: now,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
