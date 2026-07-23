package version

import (
	"go-admin/app/cauth/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072200300", migrateVIPLevels) }

func migrateVIPLevels(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&models.VIPLevel{}); err != nil {
			return err
		}
		levels := []models.VIPLevel{
			{Name: "V1", Requirement: "成长值达到 1", Growth: 1, Description: "基础 VIP 权益", Status: "active", Sort: 1},
			{Name: "V2", Requirement: "成长值达到 100", Growth: 100, Description: "V2 专属权益", Status: "active", Sort: 2},
			{Name: "V3", Requirement: "成长值达到 500", Growth: 500, Description: "V3 专属权益", Status: "active", Sort: 3},
			{Name: "V4", Requirement: "成长值达到 1000", Growth: 1000, Description: "V4 专属权益", Status: "active", Sort: 4},
			{Name: "V5", Requirement: "成长值达到 3000", Growth: 3000, Description: "V5 专属权益", Status: "active", Sort: 5},
		}
		for _, level := range levels {
			var existing models.VIPLevel
			if tx.Where("name = ?", level.Name).First(&existing).Error == gorm.ErrRecordNotFound {
				if err := tx.Create(&level).Error; err != nil {
					return err
				}
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
