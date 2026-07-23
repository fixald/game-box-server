package version

import (
	"go-admin/app/cauth/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026071900100", migrateAuthRegister) }

func migrateAuthRegister(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&models.InviteCode{}, &models.RegisterTicket{}, &models.RegisterEvent{}); err != nil {
			return err
		}
		var n int64
		tx.Model(&models.InviteCode{}).Where("code = ?", "WELCOME").Count(&n)
		if n == 0 {
			if err := tx.Create(&models.InviteCode{Code: "WELCOME", Status: "active", MaxUses: 0}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
