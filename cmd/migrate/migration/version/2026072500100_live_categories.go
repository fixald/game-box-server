package version

import (
	"go-admin/app/live/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072500100", migrateLiveCategories) }

func migrateLiveCategories(db *gorm.DB, version string) error {
	if err := db.AutoMigrate(&models.Category{}); err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		categories := []models.Category{
			{ID: "career_single", Name: "单职业", Type: "career", Sort: 1, Enabled: true},
			{ID: "version_176", Name: "1.76", Type: "version", Sort: 5, Enabled: true},
		}
		for _, category := range categories {
			var existing models.Category
			err := tx.First(&existing, "id = ?", category.ID).Error
			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&category).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
