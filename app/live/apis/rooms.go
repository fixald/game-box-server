package apis

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-admin/app/live/models"
	"go-admin/common/alilive"
	commonapis "go-admin/common/apis"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"github.com/google/uuid"
)

// RoomItem is the public representation of a live room. Keep this separate
// from the persistence model so database column names and internal IDs never
// become part of the client contract.
type RoomItem struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	StreamerName   string     `json:"streamerName"`
	StreamerAvatar string     `json:"streamerAvatar"`
	CoverURL       string     `json:"coverUrl"`
	Viewers        int        `json:"viewers"`
	GameID         *string    `json:"gameId"`
	GameName       string     `json:"gameName"`
	ServerID       *string    `json:"serverId"`
	ServerName     string     `json:"serverName"`
	Status         string     `json:"status"`
	RoomURL        string     `json:"roomUrl"`
	StartedAt      time.Time  `json:"startedAt"`
	EndedAt        *time.Time `json:"endedAt"`
	Sort           int        `json:"sort"`
}

type RoomListResponse struct {
	List     []RoomItem `json:"list"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Total    int64      `json:"total"`
	HasMore  bool       `json:"hasMore"`
}

type createRoomRequest struct {
	Title          string `json:"title" binding:"required"`
	StreamerName   string `json:"streamerName" binding:"required"`
	StreamerAvatar string `json:"streamerAvatar"`
	CoverURL       string `json:"coverUrl"`
	GameID         *uint  `json:"gameId"`
	GameName       string `json:"gameName"`
	ServerID       *uint  `json:"serverId"`
	ServerName     string `json:"serverName"`
}

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

	query := a.Orm.Model(&models.Room{})
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", "live")
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(title LIKE ? OR streamer_name LIKE ? OR game_name LIKE ? OR server_name LIKE ?)", like, like, like, like)
	}
	if categoryID := strings.TrimSpace(c.Query("categoryId")); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if gameID := strings.TrimPrefix(strings.TrimSpace(c.Query("gameId")), "game_"); gameID != "" {
		id, err := strconv.ParseUint(gameID, 10, 32)
		if err != nil {
			a.Error(90001, err, "游戏 ID 无效")
			return
		}
		query = query.Where("game_id = ?", uint(id))
	}
	if viewers := strings.TrimSpace(c.Query("viewers")); viewers != "" {
		min, err := strconv.Atoi(viewers)
		if err != nil || min < 0 {
			a.Error(90001, nil, "观看人数无效")
			return
		}
		query = query.Where("viewers >= ?", min)
	}
	if startedAt := strings.TrimSpace(c.Query("startedAt")); startedAt != "" {
		started, err := time.Parse(time.RFC3339, startedAt)
		if err != nil {
			a.Error(90001, err, "开始时间无效")
			return
		}
		query = query.Where("started_at >= ?", started.UTC())
	}
	if recommendation := strings.TrimSpace(c.Query("recommendation")); recommendation != "" {
		value, err := strconv.ParseBool(recommendation)
		if err != nil {
			a.Error(90001, err, "推荐参数无效")
			return
		}
		query = query.Where("recommendation = ?", value)
	}
	order := "viewers DESC, sort ASC, id DESC"
	switch strings.ToLower(strings.TrimSpace(c.Query("sort"))) {
	case "latest", "newest":
		order = "started_at DESC, id DESC"
	case "recommended":
		order = "recommendation DESC, sort ASC, viewers DESC, id DESC"
	case "viewers", "popular":
		order = "viewers DESC, sort ASC, id DESC"
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}
	var rows []models.Room
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}

	list := make([]RoomItem, 0, len(rows))
	for _, row := range rows {
		item := RoomItem{
			ID: fmt.Sprintf("live_%d", row.ID), Title: row.Title,
			StreamerName: row.StreamerName, StreamerAvatar: row.StreamerAvatar,
			CoverURL: row.CoverURL, Viewers: row.Viewers,
			GameID: prefixedID("game", row.GameID), GameName: row.GameName,
			ServerID: prefixedID("server", row.ServerID), ServerName: row.ServerName,
			Status: row.Status, RoomURL: row.RoomURL,
			StartedAt: row.StartedAt, EndedAt: row.EndedAt, Sort: row.Sort,
		}
		list = append(list, item)
	}

	data := RoomListResponse{
		List: list, Page: page, PageSize: pageSize, Total: total,
		HasMore: int64(page*pageSize) < total,
	}
	a.OK(data, "ok")
}

// Detail returns one live room and its playback metadata.
// @Summary Get client live room
// @Tags client-live
// @Produce json
// @Param id path string true "Live room ID, for example live_1001"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/client/live/rooms/{id} [get]
func Detail(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	id, err := parseLiveID(c.Param("id"))
	if err != nil {
		a.Error(90001, err, "直播间不存在")
		return
	}

	var row models.Room
	if err := a.Orm.First(&row, id).Error; err != nil {
		a.Error(90001, err, "直播间不存在")
		return
	}

	streamerID := row.StreamerID
	if streamerID == "" {
		streamerID = fmt.Sprintf("streamer_%d", row.ID)
	}
	protocol := row.StreamProtocol
	if protocol == "" {
		protocol = "hls"
	}
	serverStatus := row.ServerStatus
	if serverStatus == "" {
		serverStatus = "hot"
	}
	qualities := row.StreamQualities
	if qualities == nil {
		qualities = []models.StreamQuality{{Name: "高清", URL: row.RoomURL}}
	}

	a.OK(gin.H{
		"id": rowID("live", row.ID), "title": row.Title, "announcement": row.Announcement,
		"streamer": gin.H{"id": streamerID, "name": row.StreamerName, "avatarUrl": row.StreamerAvatar, "fans": row.StreamerFans, "isFollowed": false},
		"stream":   gin.H{"playUrl": row.RoomURL, "protocol": protocol, "expiresAt": row.StreamExpiresAt, "qualities": qualities},
		"game":     optionalEntity("game", row.GameID, row.GameName),
		"server":   optionalEntityWithStatus("server", row.ServerID, row.ServerName, serverStatus),
		"viewers":  row.Viewers, "status": row.Status, "startedAt": row.StartedAt,
	}, "ok")
}

func parseLiveID(value string) (uint, error) {
	id, err := strconv.ParseUint(strings.TrimPrefix(value, "live_"), 10, 32)
	return uint(id), err
}

func rowID(prefix string, id uint) string { return fmt.Sprintf("%s_%d", prefix, id) }

// Create creates a new live room with push URL.
// @Summary Create a new live room
// @Tags client-live
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body createRoomRequest true "Room information"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 90001 {object} map[string]interface{}
// @Router /api/v1/client/live/room [post]
func Create(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.Error(90001, err, "请求参数错误")
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		a.Error(401, nil, "未登录")
		return
	}

	streamName := fmt.Sprintf("stream_%d_%s", userID, uuid.New().String()[:8])

	pushURL, err := alilive.GeneratePushURL(streamName)
	if err != nil {
		a.Error(90002, err, "生成推流URL失败")
		return
	}

	playURL := alilive.GenerateHLSURL(streamName)

	room := models.Room{
		UserID:         userID,
		Title:          req.Title,
		StreamerName:   req.StreamerName,
		StreamerAvatar: req.StreamerAvatar,
		CoverURL:       req.CoverURL,
		GameID:         req.GameID,
		GameName:       req.GameName,
		ServerID:       req.ServerID,
		ServerName:     req.ServerName,
		Status:         "live",
		RoomURL:        playURL,
		PushURL:        pushURL,
		StreamName:     streamName,
		StartedAt:      time.Now().UTC(),
	}

	if err := a.Orm.Create(&room).Error; err != nil {
		a.Error(90002, err, "创建房间失败")
		return
	}

	data := gin.H{
		"room": gin.H{
			"id":             fmt.Sprintf("live_%d", room.ID),
			"title":          room.Title,
			"streamerName":   room.StreamerName,
			"streamerAvatar": room.StreamerAvatar,
			"coverUrl":       room.CoverURL,
			"pushUrl":        room.PushURL,
			"roomUrl":        room.RoomURL,
			"gameId":         numericID(room.GameID),
			"gameName":       room.GameName,
			"serverId":       numericID(room.ServerID),
			"serverName":     room.ServerName,
			"status":         room.Status,
			"startedAt":      room.StartedAt,
		},
	}
	a.OK(data, "创建成功")
}

// GetMyRoom returns the user's current live room.
// @Summary Get user's current live room
// @Tags client-live
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/client/live/room [get]
func GetMyRoom(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		a.Error(401, nil, "未登录")
		return
	}

	var room models.Room
	err := a.Orm.Where("user_id = ? AND status = ?", userID, "live").First(&room).Error
	if err != nil {
		a.Error(90002, err, "未找到房间")
		return
	}

	data := gin.H{
		"room": gin.H{
			"id":             fmt.Sprintf("live_%d", room.ID),
			"title":          room.Title,
			"streamerName":   room.StreamerName,
			"streamerAvatar": room.StreamerAvatar,
			"coverUrl":       room.CoverURL,
			"pushUrl":        room.PushURL,
			"roomUrl":        room.RoomURL,
			"viewers":        room.Viewers,
			"gameId":         numericID(room.GameID),
			"gameName":       room.GameName,
			"serverId":       numericID(room.ServerID),
			"serverName":     room.ServerName,
			"status":         room.Status,
			"startedAt":      room.StartedAt,
		},
	}
	a.OK(data, "success")
}

// EndRoom ends the live room.
// @Summary End the live room
// @Tags client-live
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/client/live/room [delete]
func EndRoom(c *gin.Context) {
	a, ok := db(c)
	if !ok {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		a.Error(401, nil, "未登录")
		return
	}

	now := time.Now().UTC()
	err := a.Orm.Model(&models.Room{}).
		Where("user_id = ? AND status = ?", userID, "live").
		Updates(map[string]interface{}{
			"status":   "ended",
			"ended_at": now,
		}).Error
	if err != nil {
		a.Error(90002, err, "结束房间失败")
		return
	}

	a.OK(nil, "结束成功")
}

func optionalEntity(prefix string, id *uint, name string) interface{} {

	if id == nil {
		return nil
	}
	return gin.H{"id": rowID(prefix, *id), "name": name}
}
func numericID(id *uint) interface{} {
	if id == nil {
		return nil
	}
	return gin.H{"id": *id}
}

func optionalEntityWithStatus(prefix string, id *uint, name, status string) interface{} {
	entity := optionalEntity(prefix, id, name)
	if entity == nil {
		return nil
	}
	return gin.H{"id": rowID(prefix, *id), "name": name, "status": status}
}

func prefixedID(prefix string, id *uint) *string {
	if id == nil {
		return nil
	}
	value := fmt.Sprintf("%s_%d", prefix, *id)
	return &value
}

func getUserID(c *gin.Context) uint {
	data, exists := c.Get(jwtauth.JwtPayloadKey)
	if !exists {
		return 0
	}

	claims, ok := data.(jwtauth.MapClaims)
	if !ok {
		return 0
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return 0
	}
	id, err := strconv.ParseUint(sub, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}
