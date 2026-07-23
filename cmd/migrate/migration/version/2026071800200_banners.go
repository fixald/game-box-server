package version

import (
	"go-admin/app/banners/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026071800200", migrateBanners) }
func migrateBanners(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&models.Banner{}); err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
