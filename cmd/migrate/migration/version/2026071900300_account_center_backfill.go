package version

import (
	"go-admin/app/cauth/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

// This migration is intentionally separate from 2026071900200 so existing
// installations that already recorded the account-center migration receive
// the tables added afterwards as well.
func init() { migration.Migrate.SetVersion("2026071900300", migrateAccountCenterBackfill) }

func migrateAccountCenterBackfill(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&models.User{}, &models.RecentGame{}, &models.DownloadRecord{}, &models.CheckinRecord{}, &models.Message{}, &models.UserSettings{}, &models.RewardRecord{}); err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
