package version

import (
	"go-admin/app/live/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072500300", migrateLiveStreamers) }

func migrateLiveStreamers(db *gorm.DB, version string) error {
	if err := db.AutoMigrate(&models.Streamer{}, &models.Room{}); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		streamer := models.Streamer{ID: "streamer_1001", Name: "小眼睛", AvatarURL: "https://cdn.example.com/avatar.jpg", CoverURL: "https://cdn.example.com/streamer-cover.jpg", Description: "传奇老玩家，专注新区冲榜", Fans: 25800, IsLive: true, CurrentRoomID: "live_1001", Sort: 1}
		var existing models.Streamer
		if err := tx.First(&existing, "id = ?", streamer.ID).Error; err == gorm.ErrRecordNotFound {
			if err := tx.Create(&streamer).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := tx.Model(&models.Room{}).Where("id = ? AND (streamer_id = '' OR streamer_id IS NULL)", 1001).Update("streamer_id", streamer.ID).Error; err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
