package version

import (
	"go-admin/app/live/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072700100", migrateLiveRoomCategoryFields) }

func migrateLiveRoomCategoryFields(db *gorm.DB, version string) error {
	if err := db.AutoMigrate(&models.Room{}); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		backfills := []struct {
			ID                     uint
			CategoryID, Name, Type string
		}{
			{1001, "career_single", "单职业", "career"}, {1002, "career_single", "单职业", "career"}, {1003, "version_180", "1.80", "version"}, {1004, "career_multi", "多职业", "career"}, {1005, "career_single", "单职业", "career"},
		}
		for _, row := range backfills {
			if err := tx.Model(&models.Room{}).Where("id = ? AND (category_id = '' OR category_id IS NULL)", row.ID).Updates(map[string]interface{}{"category_id": row.CategoryID, "category_name": row.Name, "category_type": row.Type}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Room{}).Where("id = ? AND category_id <> ''", row.ID).Updates(map[string]interface{}{"category_name": row.Name, "category_type": row.Type}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
