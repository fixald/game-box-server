package version

import (
	"time"

	bannerModels "go-admin/app/banners/models"
	liveModels "go-admin/app/live/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072300200", seedBannerAndLiveRooms) }

// seedBannerAndLiveRooms inserts local/test fixtures only when their IDs do not exist.
// It is intentionally idempotent and does not update existing operational data.
func seedBannerAndLiveRooms(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		startAt := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
		endAt := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)

		var banner bannerModels.Banner
		if err := tx.First(&banner, 3001).Error; err == gorm.ErrRecordNotFound {
			banner = bannerModels.Banner{
				ID: 3001, Title: "星海远征新服开启", ImageURL: "https://cdn.example.com/banner-star.jpg",
				LinkType: "game", LinkValue: "1001", Position: "home_top", Weight: 100,
				Sort: 1, GameID: uintPtr(1001), StartAt: startAt, EndAt: endAt,
				Status: "published", CreatedAt: now, UpdatedAt: now,
			}
			if err := tx.Create(&banner).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		var room liveModels.Room
		if err := tx.First(&room, 1001).Error; err == gorm.ErrRecordNotFound {
			room = liveModels.Room{
				ID: 1001, Title: "冰雪传奇 · 新区冲榜，首充送神装", StreamerName: "小眼睛",
				StreamerAvatar: "https://cdn.example.com/avatar/live_1001.jpg",
				CoverURL:       "https://cdn.example.com/live/live_1001.jpg", Viewers: 1280,
				GameID: uintPtr(1001), GameName: "冰雪传奇", ServerID: uintPtr(2001),
				ServerName: "火龙一区", Status: "live", RoomURL: "/live/live_1001",
				StartedAt: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC), Sort: 1,
				CreatedAt: now, UpdatedAt: now,
			}
			if err := tx.Create(&room).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		return tx.Create(&common.Migration{Version: version}).Error
	})
}

func uintPtr(value uint) *uint { return &value }
