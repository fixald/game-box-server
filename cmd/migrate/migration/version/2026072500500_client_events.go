package version

import (
	"go-admin/app/live/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072500500", migrateClientEvents) }

func migrateClientEvents(db *gorm.DB, version string) error {
	if err := db.AutoMigrate(&models.ClientEvent{}); err != nil {
		return err
	}
	return db.Create(&common.Migration{Version: version}).Error
}
