package version

import (
	"go-admin/app/cauth/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026071900200", migrateAccountCenter) }
func migrateAccountCenter(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&models.User{}, &models.FavoriteGame{}, &models.RecentGame{}, &models.RecentServer{}, &models.DownloadRecord{}, &models.CheckinRecord{}, &models.Message{}, &models.LoginRecord{}, &models.BindCode{}, &models.UserSettings{}, &models.RewardRecord{}, &models.TaskClaim{}); err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
