package version

import (
	"time"

	"go-admin/app/live/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072600100", seedLiveFilterData) }

// seedLiveFilterData adds local fixtures for live-room filtering verification.
// Existing rows are never updated or overwritten.
func seedLiveFilterData(db *gorm.DB, version string) error {
	now := time.Now().UTC()
	return db.Transaction(func(tx *gorm.DB) error {
		categories := []models.Category{
			{ID: "career_multi", Name: "多职业", Type: "career", Sort: 2, Enabled: true},
			{ID: "version_180", Name: "1.80", Type: "version", Sort: 6, Enabled: true},
			{ID: "career_old", Name: "已下线分类", Type: "career", Sort: 99, Enabled: false},
		}
		for _, row := range categories {
			var existing models.Category
			if err := tx.First(&existing, "id = ?", row.ID).Error; err == gorm.ErrRecordNotFound {
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}

		rooms := []models.Room{
			{ID: 1002, Title: "传奇新区开荒冲级", Announcement: "欢迎进入直播间", StreamerID: "streamer_1002", StreamerName: "老兵阿强", Viewers: 860, GameID: uintPtr(1001), GameName: "冰雪传奇", ServerID: uintPtr(2002), ServerName: "冰雪二区", CategoryID: "career_single", Status: "live", RoomURL: "/live/live_1002", StreamProtocol: "hls", StartedAt: now.Add(-2 * time.Hour), Sort: 2, Recommendation: true, CreatedAt: now, UpdatedAt: now},
			{ID: 1003, Title: "1.80版本装备讲解", Announcement: "新版本玩法详解", StreamerID: "streamer_1003", StreamerName: "装备研究所", Viewers: 420, GameID: uintPtr(1002), GameName: "传奇世界", ServerID: uintPtr(3001), ServerName: "经典一区", CategoryID: "version_180", Status: "live", RoomURL: "/live/live_1003", StreamProtocol: "hls", StartedAt: now.Add(-45 * time.Minute), Sort: 3, Recommendation: false, CreatedAt: now, UpdatedAt: now},
			{ID: 1004, Title: "多职业玩法预告", Announcement: "今晚八点正式开播", StreamerID: "streamer_1004", StreamerName: "职业玩家", Viewers: 0, GameID: uintPtr(1003), GameName: "热血传奇", ServerID: uintPtr(4001), ServerName: "新区一服", CategoryID: "career_multi", Status: "upcoming", RoomURL: "/live/live_1004", StreamProtocol: "hls", StartedAt: now.Add(2 * time.Hour), Sort: 4, Recommendation: true, CreatedAt: now, UpdatedAt: now},
			{ID: 1005, Title: "传奇副本回放：全服首通", Announcement: "精彩录像回放", StreamerID: "streamer_1005", StreamerName: "副本攻略君", Viewers: 120, GameID: uintPtr(1001), GameName: "冰雪传奇", ServerID: uintPtr(2001), ServerName: "火龙一区", CategoryID: "career_single", Status: "replay", RoomURL: "/live/live_1005", StreamProtocol: "hls", StartedAt: now.Add(-24 * time.Hour), Sort: 5, Recommendation: false, CreatedAt: now, UpdatedAt: now},
		}
		for _, row := range rooms {
			var existing models.Room
			if err := tx.First(&existing, row.ID).Error; err == gorm.ErrRecordNotFound {
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
