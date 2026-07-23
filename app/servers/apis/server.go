package apis

import (
	"encoding/csv"
	"github.com/gin-gonic/gin"
	gamemodels "go-admin/app/games/models"
	"go-admin/app/servers/models"
	"go-admin/common/apis"
	"io"
	"strconv"
	"strings"
	"time"
)

type ClientServer struct {
	ID            uint      `json:"id"`
	GameID        uint      `json:"gameId"`
	GameName      string    `json:"gameName"`
	Name          string    `json:"name"`
	ImageURL      string    `json:"imageUrl"`
	OpenTime      time.Time `json:"openTime"`
	Status        string    `json:"status"`
	IsRecommended bool      `json:"isRecommended"`
	OnlineLabel   string    `json:"onlineLabel"`
	Tags          []string  `json:"tags"`
}

type input struct {
	GameID               uint       `json:"gameId" binding:"required"`
	Name                 string     `json:"name" binding:"required"`
	ImageURL             string     `json:"imageUrl"`
	OpenTime             time.Time  `json:"openTime"`
	Status               string     `json:"status"`
	MergeTime            *time.Time `json:"mergeTime"`
	MinClientVersion     string     `json:"minClientVersion"`
	IsRecommended        bool       `json:"isRecommended"`
	RecommendationWeight int        `json:"recommendationWeight"`
}

func db(c *gin.Context) (*apis.Api, bool) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接失败")
		return a, false
	}
	return a, true
}
func List(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	s, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if p < 1 {
		p = 1
	}
	if s > 100 {
		s = 100
	}
	q := a.Orm.Model(&models.Server{})
	if v := c.Query("gameId"); v != "" {
		q = q.Where("game_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}
	if c.Query("recommended") == "true" {
		q = q.Where("is_recommended = ?", true)
	}
	var n int64
	q.Count(&n)
	var rows []models.Server
	if q.Order("id DESC").Offset((p-1)*s).Limit(s).Find(&rows).Error != nil {
		a.Error(90002, nil, "查询失败")
		return
	}
	a.PageOK(rows, int(n), p, s, "success")
}

// RecommendedList returns the public, homepage new-server recommendation feed.
// It deliberately has no authentication requirement; the client homepage is guest-readable.
func RecommendedList(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	q := a.Orm.Table(models.Server{}.TableName()+" AS s").
		Joins("LEFT JOIN "+gamemodels.Game{}.TableName()+" AS g ON g.id = s.game_id").
		Where("s.is_recommended = ?", true).
		Where("s.status NOT IN ?", []string{"maintenance", "closed"})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		a.Error(500, err, "查询推荐区服失败")
		return
	}
	rows := make([]ClientServer, 0)
	query := q.Select("s.id, s.game_id, g.name AS game_name, s.name, s.image_url, s.open_time, s.status, s.is_recommended, s.recommendation_weight").
		Order("s.recommendation_weight DESC, CASE s.status WHEN 'opening_soon' THEN 0 WHEN 'preview' THEN 1 WHEN 'hot' THEN 2 WHEN 'normal' THEN 3 WHEN 'full' THEN 4 ELSE 5 END, s.open_time ASC, s.id DESC").Offset((page - 1) * size).Limit(size)
	if err := query.Scan(&rows).Error; err != nil {
		a.Error(500, err, "查询推荐区服失败")
		return
	}
	for i := range rows {
		rows[i].OnlineLabel = "预约中"
		rows[i].Tags = []string{"新服"}
		if rows[i].IsRecommended {
			rows[i].Tags = append(rows[i].Tags, "推荐区服")
		}
	}
	a.PageOK(rows, int(total), page, size, "success")
}
func Get(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var row models.Server
	if a.Orm.First(&row, c.Param("id")).Error != nil {
		a.Error(30001, nil, "区服不存在")
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
	if c.ShouldBindJSON(&in) != nil {
		a.Error(90001, nil, "参数错误")
		return
	}
	if in.Status == "" {
		in.Status = "preview"
	}
	row := models.Server{GameID: in.GameID, Name: in.Name, ImageURL: in.ImageURL, OpenTime: in.OpenTime, Status: in.Status, MergeTime: in.MergeTime, MinClientVersion: in.MinClientVersion, IsRecommended: in.IsRecommended, RecommendationWeight: in.RecommendationWeight}
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
	if c.ShouldBindJSON(&in) != nil {
		a.Error(90001, nil, "参数错误")
		return
	}
	var row models.Server
	if a.Orm.First(&row, c.Param("id")).Error != nil {
		a.Error(30001, nil, "区服不存在")
		return
	}
	if a.Orm.Model(&row).Updates(map[string]interface{}{"game_id": in.GameID, "name": in.Name, "image_url": in.ImageURL, "open_time": in.OpenTime, "status": in.Status, "merge_time": in.MergeTime, "min_client_version": in.MinClientVersion, "is_recommended": in.IsRecommended, "recommendation_weight": in.RecommendationWeight}).Error != nil {
		a.Error(90001, nil, "更新失败")
		return
	}
	a.OK(row, "success")
}
func BatchMaintain(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var in struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status string `json:"status" binding:"required"`
	}
	if c.ShouldBindJSON(&in) != nil || in.Status != "maintenance" && in.Status != "normal" {
		a.Error(90001, nil, "参数错误")
		return
	}
	if err := a.Orm.Model(&models.Server{}).Where("id IN ?", in.IDs).Update("status", in.Status).Error; err != nil {
		a.Error(90001, err, "批量维护失败")
		return
	}
	a.OK(nil, "success")
}
func Import(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	f, _, err := c.Request.FormFile("file")
	if err != nil {
		a.Error(90001, err, "CSV 文件不能为空")
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	_, _ = r.Read()
	failed := []gin.H{}
	created := 0
	line := 1
	for {
		line++
		rec, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil || len(rec) < 3 {
			failed = append(failed, gin.H{"line": line, "reason": "字段不足"})
			continue
		}
		gid, e := strconv.ParseUint(strings.TrimSpace(rec[0]), 10, 64)
		if e != nil {
			failed = append(failed, gin.H{"line": line, "reason": "gameId 无效"})
			continue
		}
		row := models.Server{GameID: uint(gid), Name: strings.TrimSpace(rec[1]), Status: strings.TrimSpace(rec[2])}
		if row.Name == "" {
			failed = append(failed, gin.H{"line": line, "reason": "名称为空"})
			continue
		}
		if a.Orm.Create(&row).Error != nil {
			failed = append(failed, gin.H{"line": line, "reason": "写入失败"})
		} else {
			created++
		}
	}
	a.OK(gin.H{"created": created, "failed": failed}, "success")
}
