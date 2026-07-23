package apis

import (
	"github.com/gin-gonic/gin"
	"go-admin/app/cauth/models"
)

// VIPLevels returns the active client VIP level definitions.
// @Summary Get client VIP levels
// @Tags client-vip
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/client/vip/levels [get]
func VIPLevels(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	if _, ok := accountUser(c, a); !ok {
		return
	}
	var levels []models.VIPLevel
	if err := a.Orm.Where("status = ?", "active").Order("sort ASC, growth ASC").Find(&levels).Error; err != nil {
		taskWrite(c, 500, "VIP等级加载失败", gin.H{"levels": []gin.H{}})
		return
	}
	result := make([]gin.H, 0, len(levels))
	for _, level := range levels {
		result = append(result, gin.H{"name": level.Name, "requirement": level.Requirement, "growth": level.Growth, "desc": level.Description})
	}
	taskWrite(c, 200, "success", gin.H{"levels": result})
}
