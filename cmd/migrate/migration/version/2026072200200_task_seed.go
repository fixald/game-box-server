package version

import (
	"go-admin/app/cauth/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
)

func init() { migration.Migrate.SetVersion("2026072200200", seedTaskCenter) }

func seedTaskCenter(db *gorm.DB, version string) error {
	// DDL 放在事务外：MySQL 的 ALTER 会隐式提交，事务内 AutoMigrate 不可靠。
	// TEXT 不能带 DEFAULT（MySQL Error 1101）。
	if err := db.AutoMigrate(&models.Task{}, &models.CheckinReward{}, &models.CheckinRewardClaim{}); err != nil {
		return err
	}
	if !db.Migrator().HasColumn(&models.Task{}, "Rewards") {
		if err := db.Exec("ALTER TABLE `gb_tasks` ADD COLUMN `rewards` text NULL").Error; err != nil {
			return err
		}
	}
	// emoji 图标需要 utf8mb4；连接与表字符集不一致会触发 Error 3988
	if err := db.Exec("ALTER TABLE `gb_tasks` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE `gb_checkin_rewards` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return err
	}
	_ = db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci").Error

	return db.Transaction(func(tx *gorm.DB) error {
		// Keep this seed idempotent so existing environments can safely run it.
		tasks := []models.Task{
			{Code: "task-1", Category: "daily", Title: "每日登录盒子", Description: "登录客户端即可完成", Icon: "◷", Target: 1, Rewards: `[{"type":"points","name":"积分","amount":20,"icon":"✦"}]`, Status: "active", Sort: 1},
			{Code: "task-2", Category: "daily", Title: "观看直播 10 分钟", Description: "在直播频道观看任意直播", Icon: "▶", Target: 10, Rewards: `[{"type":"points","name":"积分","amount":50,"icon":"✦"}]`, Status: "active", Sort: 2},
			{Code: "task-3", Category: "game", Title: "启动一次游戏", Description: "启动任意已安装的传奇游戏", Icon: "⚔", Target: 1, Rewards: `[{"type":"gift","name":"新手礼包","icon":"🎁"}]`, Status: "active", Sort: 3},
			{Code: "task-4", Category: "newbie", Title: "完善个人资料", Description: "上传头像并设置昵称", Icon: "◇", Target: 1, Rewards: `[{"type":"vip_exp","name":"SVIP经验","amount":100,"icon":"♛"}]`, Status: "active", Sort: 4},
			{Code: "task-5", Category: "social", Title: "关注 3 位主播", Description: "关注你喜欢的传奇主播", Icon: "♡", Target: 3, Rewards: `[{"type":"points","name":"积分","amount":30,"icon":"✦"}]`, Status: "active", Sort: 5},
			{Code: "task-6", Category: "newbie", Title: "完成首次下载", Description: "下载并校验一款游戏", Icon: "↓", Target: 1, Rewards: `[{"type":"coupon","name":"下载加速券","amount":1,"icon":"⚡"}]`, Status: "active", Sort: 6},
			{Code: "task-7", Category: "daily", Title: "浏览新服推荐", Description: "查看今日推荐新区和开服信息", Icon: "◈", Target: 1, Rewards: `[{"type":"points","name":"积分","amount":10,"icon":"✦"}]`, Status: "active", Sort: 7},
			{Code: "task-8", Category: "daily", Title: "完成一次签到", Description: "在任务中心完成每日签到", Icon: "✓", Target: 1, Rewards: `[{"type":"points","name":"积分","amount":20,"icon":"✦"}]`, Status: "active", Sort: 8},
			{Code: "task-9", Category: "game", Title: "查看游戏详情", Description: "浏览任意一款游戏的详情页面", Icon: "◉", Target: 1, Rewards: `[{"type":"points","name":"积分","amount":15,"icon":"✦"}]`, Status: "active", Sort: 9},
			{Code: "task-10", Category: "game", Title: "进入推荐区服", Description: "从新服推荐中选择一个区服", Icon: "⚑", Target: 1, Rewards: `[{"type":"gift","name":"区服礼包","icon":"🎁"}]`, Status: "active", Sort: 10},
			{Code: "task-11", Category: "social", Title: "查看一条资讯", Description: "阅读平台最新游戏资讯", Icon: "▤", Target: 1, Rewards: `[{"type":"points","name":"积分","amount":10,"icon":"✦"}]`, Status: "active", Sort: 11},
			{Code: "task-12", Category: "newbie", Title: "完成首次区服选择", Description: "选择喜欢的游戏区服", Icon: "◇", Target: 1, Rewards: `[{"type":"vip_exp","name":"SVIP经验","amount":50,"icon":"♛"}]`, Status: "active", Sort: 12},
		}
		for _, task := range tasks {
			var existing models.Task
			if err := tx.Where("code = ?", task.Code).First(&existing).Error; err == gorm.ErrRecordNotFound {
				if err := tx.Create(&task).Error; err != nil {
					return err
				}
			}
		}
		rewards := []models.CheckinReward{
			{Level: 1, Name: "初次签到", Reward: "10 积分", Icon: "✦", Status: "active"},
			{Level: 3, Name: "三日礼", Reward: "30 积分", Icon: "✦", Status: "active"},
			{Level: 7, Name: "七日礼包", Reward: "礼包", Icon: "🎁", Status: "active"},
			{Level: 14, Name: "半月奖励", Reward: "SVIP经验", Icon: "♛", Status: "active"},
			{Level: 30, Name: "满月大奖", Reward: "豪华礼包", Icon: "🎁", Status: "active"},
		}
		for _, reward := range rewards {
			var existing models.CheckinReward
			if err := tx.Where("level = ?", reward.Level).First(&existing).Error; err == gorm.ErrRecordNotFound {
				if err := tx.Create(&reward).Error; err != nil {
					return err
				}
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}
