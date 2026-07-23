package apis

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-admin/app/live/models"
	commonapis "go-admin/common/apis"
)

func db(c *gin.Context) (*commonapis.Api, bool) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return a, false
	}
	return a, true
}

// List returns the paginated client live-room catalogue.
// @Summary List client live rooms
// @Tags client-live
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/client/live/rooms [get]
func List(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 || pageSize < 1 || pageSize > 100 {
		a.Error(90001, nil, "分页参数无效")
		return
	}

	query := a.Orm.Model(&models.Room{}).Where("status = ?", "live")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}
	var rows []models.Room
	if err := query.Order("viewers DESC, sort ASC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}

	list := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		item := gin.H{
			"id": fmt.Sprintf("live_%d", row.ID), "title": row.Title,
			"streamerName": row.StreamerName, "streamerAvatar": row.StreamerAvatar,
			"coverUrl": row.CoverURL, "viewers": row.Viewers,
			"gameId": numericID(row.GameID), "gameName": row.GameName,
			"serverId": numericID(row.ServerID), "serverName": row.ServerName,
			"status": row.Status, "roomUrl": row.RoomURL,
			"startedAt": row.StartedAt, "endedAt": row.EndedAt,
		}
		list = append(list, item)
	}

	data := gin.H{
		"requestId": c.GetHeader("X-Request-ID"), "list": list, "page": page,
		"pageSize": pageSize, "total": total, "hasMore": int64(page*pageSize) < total,
	}
	a.OK(data, "success")
}

func numericID(id *uint) interface{} {
	if id == nil {
		return nil
	}
	return strconv.FormatUint(uint64(*id), 10)
}
