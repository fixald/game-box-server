package apis

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-admin/app/games/models"
	"go-admin/common/apis"
)

type gameInput struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	Category    string `json:"category"`
	GameType    string `json:"gameType"`
	Publisher   string `json:"publisher"`
	VersionTags string `json:"versionTags"`
}

// List returns only published games. code 2001 is reserved for game errors.
func List(c *gin.Context) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(http.StatusInternalServerError, a.Errors, "数据库连接获取失败")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", c.DefaultQuery("pageIndex", "1")))
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
	q := a.Orm.Model(&models.Game{})
	if c.GetBool("publishedOnly") {
		q = q.Where("status = ?", "published")
	}
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}
	if v := c.Query("category"); v != "" {
		q = q.Where("category = ?", v)
	}
	// The admin UI uses `name`, while the public API historically uses
	// `keyword`; accept both so the shared list endpoint stays compatible.
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = c.Query("name")
	}
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if v := c.Query("tag"); v != "" {
		q = q.Where("version_tags LIKE ?", "%"+v+"%")
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		a.Error(500, err, "查询游戏失败")
		return
	}
	var list []models.Game
	order := "updated_at DESC, id DESC"
	switch c.Query("sort") {
	case "rating":
		order = "rating DESC, id DESC"
	case "downloads":
		order = "download_count DESC, id DESC"
	case "name":
		order = "name ASC, id ASC"
	}
	if err := q.Order(order).Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		a.Error(500, err, "查询游戏失败")
		return
	}
	a.PageOK(list, int(count), page, size, "success")
}

func AdminList(c *gin.Context) {
	c.Set("publishedOnly", false)
	List(c)
}

func Get(c *gin.Context) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	var g models.Game
	if err := a.Orm.First(&g, c.Param("id")).Error; err != nil {
		a.Error(20001, err, "游戏不存在")
		return
	}
	a.OK(g, "success")
}
func Create(c *gin.Context) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	var in gameInput
	if err := c.ShouldBindJSON(&in); err != nil {
		a.Error(90001, err, "参数错误")
		return
	}
	b, _ := json.Marshal(in)
	var g models.Game
	_ = json.Unmarshal(b, &g)
	if g.Status == "" {
		g.Status = "draft"
	}
	if err := a.Orm.Create(&g).Error; err != nil {
		a.Error(90001, err, "创建游戏失败")
		return
	}
	a.OK(g, "success")
}
func Update(c *gin.Context) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	var in gameInput
	if err := c.ShouldBindJSON(&in); err != nil {
		a.Error(90001, err, "参数错误")
		return
	}
	var g models.Game
	if err := a.Orm.First(&g, c.Param("id")).Error; err != nil {
		a.Error(20001, err, "游戏不存在")
		return
	}
	if err := a.Orm.Model(&g).Updates(map[string]interface{}{"name": in.Name, "slug": in.Slug, "description": in.Description, "icon_url": in.IconURL, "category": in.Category, "game_type": in.GameType, "publisher": in.Publisher, "version_tags": in.VersionTags}).Error; err != nil {
		a.Error(90001, err, "更新游戏失败")
		return
	}
	a.Orm.First(&g, g.ID)
	a.OK(g, "success")
}
func UpdateStatus(c *gin.Context) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	var in struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || (in.Status != "published" && in.Status != "offline" && in.Status != "draft") {
		a.Error(90001, err, "状态参数错误")
		return
	}
	var g models.Game
	if err := a.Orm.First(&g, c.Param("id")).Error; err != nil {
		a.Error(20001, err, "游戏不存在")
		return
	}
	a.Orm.Model(&g).Update("status", in.Status)
	a.OK(g, "success")
}
func Delete(c *gin.Context) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	var g models.Game
	if err := a.Orm.First(&g, c.Param("id")).Error; err != nil {
		a.Error(20001, err, "游戏不存在")
		return
	}
	if err := a.Orm.Model(&g).Update("status", "offline").Error; err != nil {
		a.Error(90001, err, "删除游戏失败")
		return
	}
	a.OK(nil, "success")
}

// Detail returns a published game and intentionally hides drafts/offline games.
func Detail(c *gin.Context) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(http.StatusInternalServerError, a.Errors, "数据库连接获取失败")
		return
	}
	var game models.Game
	query := a.Orm.Where("id = ?", c.Param("id"))
	if c.GetBool("publishedOnly") {
		query = query.Where("status = ?", "published")
	}
	if err := query.First(&game).Error; err != nil {
		a.Error(20001, err, "游戏不存在")
		return
	}
	a.OK(game, "success")
}

func AdminDetail(c *gin.Context) {
	c.Set("publishedOnly", false)
	Detail(c)
}
