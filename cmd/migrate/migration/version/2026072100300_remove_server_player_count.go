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
		// The column may already be absent on databases created from a newer
		// schema snapshot. Keep this migration idempotent so it can still be
		// recorded and the remaining migrations can continue.
		if tx.Migrator().HasColumn(&models.Server{}, "player_count") {
			if err := tx.Migrator().DropColumn(&models.Server{}, "player_count"); err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
