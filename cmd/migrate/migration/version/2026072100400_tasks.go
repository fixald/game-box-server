package version

import (
	"go-admin/app/cauth/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072100400", migrateTasks) }

func migrateTasks(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable("gb_task_rewards") {
			if err := tx.Migrator().DropTable("gb_task_rewards"); err != nil {
				return err
			}
		}
		if err := tx.AutoMigrate(&models.Task{}, &models.TaskClaim{}); err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
