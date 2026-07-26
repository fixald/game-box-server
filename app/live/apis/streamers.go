package apis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go-admin/app/live/models"
)

type streamerListResponse struct {
	List     interface{} `json:"list"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int64       `json:"total"`
	HasMore  bool        `json:"hasMore"`
}

// StreamerDetail returns a streamer profile.
// @Summary Get client live streamer
// @Tags client-live
// @Produce json
// @Param id path string true "Streamer ID, for example streamer_1001"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/client/live/streamers/{id} [get]
func StreamerDetail(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var row models.Streamer
	if err := a.Orm.First(&row, c.Param("id")).Error; err != nil {
		a.Error(90001, err, "主播不存在")
		return
	}
	a.OK(row, "ok")
}

// StreamerRooms returns live and historical rooms belonging to a streamer.
// @Summary List client streamer rooms
// @Tags client-live
// @Produce json
// @Param id path string true "Streamer ID"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Router /api/v1/client/live/streamers/{id}/rooms [get]
func StreamerRooms(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	page, pageSize, valid := livePagination(c)
	if !valid {
		a.Error(90001, nil, "分页参数无效")
		return
	}

	query := a.Orm.Model(&models.Room{}).Where("streamer_id = ?", c.Param("id"))
	var total int64
	if err := query.Count(&total).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}
	var rows []models.Room
	if err := query.Order("status = 'live' DESC, started_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}
	list := make([]RoomItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, roomItem(row))
	}
	a.OK(streamerListResponse{List: list, Page: page, PageSize: pageSize, Total: total, HasMore: int64(page*pageSize) < total}, "ok")
}

// Streamers returns paginated streamer profiles.
// @Summary List client live streamers
// @Tags client-live
// @Produce json
// @Param sort query string false "popular or latest"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Router /api/v1/client/live/streamers [get]
func Streamers(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	page, pageSize, valid := livePagination(c)
	if !valid {
		a.Error(90001, nil, "分页参数无效")
		return
	}
	query := a.Orm.Model(&models.Streamer{})
	order := "sort ASC, id ASC"
	if c.DefaultQuery("sort", "popular") == "popular" {
		order = "fans DESC, id ASC"
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}
	var rows []models.Streamer
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}
	a.OK(streamerListResponse{List: rows, Page: page, PageSize: pageSize, Total: total, HasMore: int64(page*pageSize) < total}, "ok")
}

func livePagination(c *gin.Context) (int, int, bool) {
	page, err1 := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err2 := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	return page, pageSize, err1 == nil && err2 == nil && page > 0 && pageSize > 0 && pageSize <= 100
}

func roomItem(row models.Room) RoomItem {
	return RoomItem{ID: fmt.Sprintf("live_%d", row.ID), Title: row.Title, StreamerName: row.StreamerName, StreamerAvatar: row.StreamerAvatar, CoverURL: row.CoverURL, Viewers: row.Viewers, GameID: prefixedID("game", row.GameID), GameName: row.GameName, ServerID: prefixedID("server", row.ServerID), ServerName: row.ServerName, Status: row.Status, RoomURL: row.RoomURL, StartedAt: row.StartedAt, EndedAt: row.EndedAt, Sort: row.Sort}
}

func parseStreamerNumber(id string) string { return strings.TrimPrefix(id, "streamer_") }
