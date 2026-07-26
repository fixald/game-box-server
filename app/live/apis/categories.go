package apis

import (
	"github.com/gin-gonic/gin"
	"go-admin/app/live/models"
)

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

	var rows []models.Category
	if err := a.Orm.Where("enabled = ?", true).
		Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}

	a.OK(gin.H{"list": rows}, "ok")
}
