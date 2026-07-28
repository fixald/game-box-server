package apis

import (
	"github.com/gin-gonic/gin"
	"go-admin/app/live/models"
)

type liveCategory struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Sort    int    `json:"sort"`
	Enabled bool   `json:"enabled"`
}

// Categories returns enabled live-room categories ordered for client display.
// @Summary List live-room categories
// @Tags client-live
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/client/live/categories [get]
func Categories(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}

	var rows []liveCategory
	if err := a.Orm.Model(&models.Room{}).
		Select("category_id AS id, category_name AS name, category_type AS type, MIN(sort) AS sort, TRUE AS enabled").
		Where("status = ? AND category_id <> '' AND category_name <> ''", "live").
		Group("category_id, category_name, category_type").Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}

	a.OK(gin.H{"list": rows}, "ok")
}
