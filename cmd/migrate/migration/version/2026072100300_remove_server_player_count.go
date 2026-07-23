package version

import (
	"go-admin/app/servers/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072100300", migrateRemoveServerPlayerCount) }

func migrateRemoveServerPlayerCount(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Migrator().DropColumn(&models.Server{}, "player_count"); err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
