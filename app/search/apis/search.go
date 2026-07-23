package apis

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	gameModels "go-admin/app/games/models"
	searchModels "go-admin/app/search/models"
	serverModels "go-admin/app/servers/models"
	commonapis "go-admin/common/apis"
)

var validTypes = map[string]bool{"all": true, "game": true, "live": true, "server": true, "gift": true, "article": true}

type Item struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Description string   `json:"description"`
	IconURL     string   `json:"iconUrl"`
	CoverURL    string   `json:"coverUrl"`
	Tags        []string `json:"tags"`
	Target      gin.H    `json:"target"`
	Score       float64  `json:"score"`
}

func userID(c *gin.Context) uint {
	claims := jwt.ExtractClaims(c)
	id, _ := strconv.ParseUint(fmt.Sprint(claims["sub"]), 10, 32)
	return uint(id)
}
func pageParam(c *gin.Context, key string, fallback, max int) int {
	n, _ := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if n < 1 {
		n = fallback
	}
	if n > max {
		n = max
	}
	return n
}
func match(text, q string) bool { return strings.Contains(strings.ToLower(text), strings.ToLower(q)) }
func Search(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		c.JSON(500, gin.H{"code": 90001, "message": "服务端异常", "requestId": c.GetHeader("X-Request-ID")})
		return
	}
	q := strings.TrimSpace(c.Query("q"))
	typ := c.DefaultQuery("type", "all")
	if !validTypes[typ] {
		typ = "all"
	}
	page, size := pageParam(c, "page", 1, 1000000), pageParam(c, "pageSize", 20, 100)
	items := make([]Item, 0)
	if typ == "all" || typ == "game" {
		var games []gameModels.Game
		query := a.Orm.Where("status = ?", "published")
		if q != "" {
			query = query.Where("name LIKE ? OR description LIKE ?", "%"+q+"%", "%"+q+"%")
		}
		query.Find(&games)
		for _, g := range games {
			items = append(items, Item{ID: fmt.Sprintf("game_%d", g.ID), Type: "game", Title: g.Name, Subtitle: g.Category, Description: g.Description, IconURL: g.IconURL, Tags: strings.Fields(strings.ReplaceAll(g.VersionTags, ",", " ")), Target: gin.H{"path": "/games/" + strconv.FormatUint(uint64(g.ID), 10)}, Score: g.Rating})
		}
	}
	if typ == "all" || typ == "server" {
		var servers []serverModels.Server
		query := a.Orm.Where("status IN ?", []string{"preview", "open", "published"})
		if q != "" {
			query = query.Where("name LIKE ?", "%"+q+"%")
		}
		query.Find(&servers)
		for _, s := range servers {
			items = append(items, Item{ID: fmt.Sprintf("server_%d", s.ID), Type: "server", Title: s.Name, Subtitle: "新区服", Description: "推荐区服", Target: gin.H{"path": "/servers/" + strconv.FormatUint(uint64(s.ID), 10)}, Score: float64(s.RecommendationWeight)})
		}
	}
	if typ == "all" || typ == "live" || typ == "gift" || typ == "article" {
		var rows []searchModels.SearchItem
		query := a.Orm.Where("type IN ?", []string{"live", "gift", "article"})
		if typ != "all" {
			query = query.Where("type = ?", typ)
		}
		if q != "" {
			query = query.Where("title LIKE ? OR subtitle LIKE ? OR description LIKE ? OR tags LIKE ?", "%"+q+"%", "%"+q+"%", "%"+q+"%", "%"+q+"%")
		}
		query.Order("score DESC, id ASC").Find(&rows)
		for _, row := range rows {
			target := gin.H{}
			_ = json.Unmarshal([]byte(row.Target), &target)
			items = append(items, Item{ID: row.ID, Type: row.Type, Title: row.Title, Subtitle: row.Subtitle, Description: row.Description, IconURL: row.IconURL, CoverURL: row.CoverURL, Tags: strings.Fields(strings.ReplaceAll(row.Tags, ",", " ")), Target: target, Score: row.Score})
		}
	}
	total := len(items)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	facets := []gin.H{}
	for _, t := range []string{"game", "live", "server", "gift", "article"} {
		n := 0
		for _, item := range items {
			if item.Type == t {
				n++
			}
		}
		if n > 0 {
			facets = append(facets, gin.H{"type": t, "label": t, "count": n})
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "requestId": c.GetHeader("X-Request-ID"), "data": gin.H{"query": q, "type": typ, "page": page, "pageSize": size, "total": total, "hasMore": end < total, "items": items[start:end], "facets": facets}})
}

func Suggestions(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	limit := pageParam(c, "limit", 8, 20)
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	var games []gameModels.Game
	a.Orm.Where("status = ? AND name LIKE ?", "published", "%"+q+"%").Limit(limit).Find(&games)
	out := make([]gin.H, 0, len(games))
	for _, g := range games {
		out = append(out, gin.H{"text": g.Name, "type": "game", "highlight": g.Name})
	}
	c.JSON(200, gin.H{"code": 0, "message": "ok", "requestId": c.GetHeader("X-Request-ID"), "data": gin.H{"query": q, "suggestions": out}})
}

func Hot(c *gin.Context) {
	limit := pageParam(c, "limit", 10, 20)
	words := []string{"冰雪传奇", "新服", "首充双倍", "传奇世界", "小眼睛"}
	if limit < len(words) {
		words = words[:limit]
	}
	items := make([]gin.H, 0, len(words))
	for i, w := range words {
		items = append(items, gin.H{"keyword": w, "rank": i + 1, "trend": "stable", "count": 0})
	}
	c.JSON(200, gin.H{"code": 0, "message": "ok", "requestId": c.GetHeader("X-Request-ID"), "data": gin.H{"items": items}})
}

func History(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	uid := userID(c)
	limit := pageParam(c, "limit", 10, 50)
	var rows []searchModels.SearchHistory
	a.Orm.Where("user_id = ?", uid).Order("searched_at DESC").Limit(limit).Find(&rows)
	c.JSON(200, gin.H{"code": 0, "message": "ok", "requestId": c.GetHeader("X-Request-ID"), "data": gin.H{"items": rows}})
}
func AddHistory(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	var in struct {
		Keyword string `json:"keyword"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Keyword) == "" {
		c.JSON(400, gin.H{"code": 400, "message": "keyword 无效", "data": nil})
		return
	}
	uid := userID(c)
	a.Orm.Where("user_id = ? AND keyword = ?", uid, strings.TrimSpace(in.Keyword)).Delete(&searchModels.SearchHistory{})
	row := searchModels.SearchHistory{UserID: uid, Keyword: strings.TrimSpace(in.Keyword), SearchedAt: time.Now()}
	a.Orm.Create(&row)
	c.JSON(200, gin.H{"code": 0, "message": "ok", "requestId": c.GetHeader("X-Request-ID"), "data": gin.H{"item": row}})
}
func ClearHistory(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	a.Orm.Where("user_id = ?", userID(c)).Delete(&searchModels.SearchHistory{})
	c.JSON(200, gin.H{"code": 0, "message": "ok", "requestId": c.GetHeader("X-Request-ID"), "data": nil})
}
func Event(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	var in searchModels.SearchEvent
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.EventType) == "" || strings.TrimSpace(in.Query) == "" {
		c.JSON(400, gin.H{"code": 400, "message": "事件参数无效", "data": nil})
		return
	}
	in.UserID = userID(c)
	if in.OccurredAt.IsZero() {
		in.OccurredAt = time.Now()
	}
	a.Orm.Create(&in)
	c.JSON(200, gin.H{"code": 0, "message": "ok", "requestId": c.GetHeader("X-Request-ID"), "data": nil})
}
