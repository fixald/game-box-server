package version

import (
	"encoding/json"
	"fmt"
	searchModels "go-admin/app/search/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
	"gorm.io/gorm"
	"strings"
)

func init() { migration.Migrate.SetVersion("2026072000100", migrateSearch) }
func migrateSearch(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&searchModels.SearchItem{}, &searchModels.SearchHistory{}, &searchModels.SearchEvent{}); err != nil {
			return err
		}
		for _, typ := range []string{"all", "game", "live", "server", "gift", "article"} {
			if typ == "all" {
				continue
			}
			for i := 1; i <= 20; i++ {
				id := fmt.Sprintf("%s-test-%02d", typ, i)
				row := searchModels.SearchItem{ID: id, Type: typ, Title: fmt.Sprintf("%s 测试数据 %02d", typeLabel(typ), i), Subtitle: fmt.Sprintf("搜索契约 %s 类型测试数据", typ), Description: "用于验证搜索、建议、分页和结果映射的测试数据", Tags: strings.Join([]string{typ, "测试", "gamebox"}, ","), Target: jsonTarget(typ, id), Score: float64(100 - i)}
				if err := tx.Where("id = ?", id).FirstOrCreate(&row).Error; err != nil {
					return err
				}
			}
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}

func typeLabel(typ string) string {
	return map[string]string{"game": "游戏", "live": "直播", "server": "区服", "gift": "礼包", "article": "资讯"}[typ]
}
func jsonTarget(typ, id string) string {
	b, _ := json.Marshal(map[string]string{"type": typ, "id": id})
	return string(b)
}
