package apis

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go-admin/app/banners/models"
	"go-admin/common/apis"
)

type input struct {
	Title     string    `json:"title" binding:"required"`
	ImageURL  string    `json:"imageUrl" binding:"required"`
	LinkType  string    `json:"linkType"`
	LinkValue string    `json:"linkValue"`
	Position  string    `json:"position" binding:"required"`
	Weight    int       `json:"weight"`
	GameID    *uint     `json:"gameId"`
	StartAt   time.Time `json:"startAt"`
	EndAt     time.Time `json:"endAt"`
	Sort      int       `json:"sort"`
}

func db(c *gin.Context) (*apis.Api, bool) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return a, false
	}
	return a, true
}
func List(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	q := a.Orm.Model(&models.Banner{})
	if v := c.Query("position"); v != "" {
		q = q.Where("position = ?", v)
	}
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}
	p, _ := strconv.Atoi(c.DefaultQuery("page", c.DefaultQuery("pageIndex", "1")))
	s, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if p < 1 {
		p = 1
	}
	if s < 1 {
		s = 20
	}
	if s > 100 {
		s = 100
	}
	var n int64
	if q.Count(&n).Error != nil {
		a.Error(90002, nil, "查询失败")
		return
	}
	var rows []models.Banner
	if q.Order("weight DESC, sort ASC, id DESC").Offset((p-1)*s).Limit(s).Find(&rows).Error != nil {
		a.Error(90002, nil, "查询失败")
		return
	}
	a.PageOK(rows, int(n), p, s, "success")
}
func Active(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	now := time.Now().UTC()
	position := c.Query("position")
	if position == "" {
		a.Error(90001, nil, "运营位不能为空")
		return
	}
	var rows []models.Banner
	if a.Orm.Where("position = ? AND status = ? AND start_at <= ? AND end_at > ?", position, "published", now, now).
		Order("weight DESC, sort ASC, id DESC").Find(&rows).Error != nil {
		a.Error(90002, nil, "查询失败")
		return
	}
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		result = append(result, gin.H{
			"id": row.ID, "title": row.Title, "imageUrl": row.ImageURL,
			"linkType": row.LinkType, "linkValue": row.LinkValue, "position": row.Position,
			"weight": row.Weight, "sort": row.Sort, "gameId": row.GameID,
			"startAt": row.StartAt, "endAt": row.EndAt,
		})
	}
	a.OK(result, "success")
}
func Get(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var row models.Banner
	if a.Orm.First(&row, c.Param("id")).Error != nil {
		a.Error(90001, nil, "Banner 不存在")
		return
	}
	a.OK(row, "success")
}
func Create(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var in input
	if c.ShouldBindJSON(&in) != nil || in.StartAt.IsZero() || in.EndAt.IsZero() || !in.StartAt.Before(in.EndAt) {
		a.Error(90001, nil, "参数错误")
		return
	}
	if in.LinkType == "" {
		in.LinkType = "none"
	}
	row := models.Banner{Title: in.Title, ImageURL: in.ImageURL, LinkType: in.LinkType, LinkValue: in.LinkValue, Position: in.Position, Weight: in.Weight, GameID: in.GameID, StartAt: in.StartAt.UTC(), EndAt: in.EndAt.UTC(), Sort: in.Sort}
	if a.Orm.Create(&row).Error != nil {
		a.Error(90001, nil, "创建失败")
		return
	}
	a.OK(row, "success")
}
func Update(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var in input
	if c.ShouldBindJSON(&in) != nil || !in.StartAt.Before(in.EndAt) {
		a.Error(90001, nil, "参数错误")
		return
	}
	var row models.Banner
	if a.Orm.First(&row, c.Param("id")).Error != nil {
		a.Error(90001, nil, "Banner 不存在")
		return
	}
	updates := map[string]interface{}{"title": in.Title, "image_url": in.ImageURL, "link_type": in.LinkType, "link_value": in.LinkValue, "position": in.Position, "weight": in.Weight, "game_id": in.GameID, "start_at": in.StartAt.UTC(), "end_at": in.EndAt.UTC(), "sort": in.Sort}
	if a.Orm.Model(&row).Updates(updates).Error != nil {
		a.Error(90001, nil, "更新失败")
		return
	}
	a.Orm.First(&row, row.ID)
	a.OK(row, "success")
}
func Delete(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	if a.Orm.Delete(&models.Banner{}, c.Param("id")).RowsAffected == 0 {
		a.Error(90001, nil, "Banner 不存在")
		return
	}
	a.OK(nil, "success")
}
func Publish(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var row models.Banner
	if a.Orm.First(&row, c.Param("id")).Error != nil {
		a.Error(90001, nil, "Banner 不存在")
		return
	}
	var conflict int64
	a.Orm.Model(&models.Banner{}).Where("id <> ? AND position = ? AND status = ? AND start_at < ? AND end_at > ?", row.ID, row.Position, "published", row.EndAt, row.StartAt).Count(&conflict)
	if conflict > 0 {
		a.Error(90001, nil, "同一运营位时间窗冲突")
		return
	}
	a.Orm.Model(&row).Update("status", "published")
	a.OK(nil, "success")
}
func Recall(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	if a.Orm.Model(&models.Banner{}).Where("id = ?", c.Param("id")).Update("status", "offline").RowsAffected == 0 {
		a.Error(90001, nil, "Banner 不存在")
		return
	}
	a.OK(nil, "success")
}
