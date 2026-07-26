package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/cauth/models"
	gameModels "go-admin/app/games/models"
	serverModels "go-admin/app/servers/models"
	commonapis "go-admin/common/apis"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	accountUserMissing    = 12001
	accountProfileInvalid = 12002
	accountOldPassword    = 12003
	accountCodeInvalid    = 12004
	accountDeviceMissing  = 12005
)

func accountUser(c *gin.Context, a *commonapis.Api) (*models.User, bool) {
	claims := jwt.ExtractClaims(c)
	raw := fmt.Sprint(claims["sub"])
	if raw == "<nil>" {
		raw = fmt.Sprint(claims["identity"])
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		write(c, accountUserMissing, "用户资料不存在", nil)
		return nil, false
	}
	var u models.User
	if a.Orm.First(&u, uint(id)).Error != nil {
		write(c, accountUserMissing, "用户资料不存在", nil)
		return nil, false
	}
	if u.Status != "active" {
		write(c, 12006, "账号冻结/封禁", nil)
		return nil, false
	}
	return &u, true
}

// Account returns the authenticated client's profile.
// @Summary Get current client user
// @Tags client-users
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/client/users/me [get]
func Account(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	write(c, 0, "ok", gin.H{"user": clientUserProfile(a.Orm, u)})
}

func clientUserProfile(db *gorm.DB, u *models.User) gin.H {
	accountMasked := maskAccount(u.Account)
	emailMasked := ""
	if strings.Contains(u.Account, "@") {
		emailMasked = accountMasked
	}
	return gin.H{
		"id": fmt.Sprintf("user_%d", u.ID), "account": u.Account, "accountMasked": accountMasked,
		"nickname": u.Nickname, "avatarUrl": u.AvatarURL, "status": "normal", "registeredAt": u.CreatedAt,
		"taskUnreadCount": taskUnreadCount(db, u.ID), "messageUnreadCount": messageUnreadCount(db, u.ID),
		"vip":      gin.H{"level": u.VipLevel, "expiresAt": nil, "growthValue": 0, "growthTarget": 0},
		"security": gin.H{"phoneMasked": maskPhone(u.PhoneCiphertext), "emailMasked": emailMasked, "realNameStatus": u.RealNameStatus},
	}
}

// taskUnreadCount is the number of tasks currently ready to claim. A claimed
// task is no longer unread, while in-progress tasks are not actionable yet.
func taskUnreadCount(db *gorm.DB, userID uint) int64 {
	var count int64
	var tasks []models.Task
	db.Where("status = ?", "active").Find(&tasks)
	for _, d := range tasks {
		if d.Code != "task-1" && d.Code != "task-3" {
			continue
		}
		var claimed int64
		db.Model(&models.TaskClaim{}).Where("user_id = ? AND task_id = ?", userID, d.Code).Count(&claimed)
		var logins int64
		if d.Code == "task-1" {
			db.Model(&models.LoginRecord{}).Where("user_id=? AND success=? AND login_at>=?", userID, true, time.Now().UTC().Truncate(24*time.Hour)).Count(&logins)
		} else {
			db.Model(&models.RecentGame{}).Where("user_id=? AND visited_at>=?", userID, time.Now().UTC().Truncate(24*time.Hour)).Count(&logins)
		}
		if claimed == 0 && logins > 0 {
			count++
		}
	}
	return count
}

func messageUnreadCount(db *gorm.DB, userID uint) int64 {
	var count int64
	db.Model(&models.Message{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&count)
	return count
}
func UpdateAccount(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var in struct {
		Nickname  *string `json:"nickname"`
		AvatarURL *string `json:"avatarUrl"`
	}
	if c.ShouldBindJSON(&in) != nil {
		write(c, accountProfileInvalid, "昵称或简介不符合规则", nil)
		return
	}
	updates := map[string]interface{}{}
	if in.Nickname != nil {
		n := strings.TrimSpace(*in.Nickname)
		if len([]rune(n)) < 1 || len([]rune(n)) > 32 {
			write(c, accountProfileInvalid, "昵称或简介不符合规则", nil)
			return
		}
		updates["nickname"] = n
	}
	if in.AvatarURL != nil {
		updates["avatar_url"] = strings.TrimSpace(*in.AvatarURL)
	}
	if len(updates) > 0 {
		a.Orm.Model(u).Updates(updates)
	}
	a.Orm.First(u, u.ID)
	write(c, 0, "ok", u)
}

// AccountStats returns counters displayed in the client account overview.
// @Summary Get current client account statistics
// @Tags client-users
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/client/users/me/stats [get]
func AccountStats(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}

	var favoriteGameCount, downloadCount int64
	a.Orm.Model(&models.FavoriteGame{}).Where("user_id = ?", u.ID).Count(&favoriteGameCount)
	a.Orm.Model(&models.DownloadRecord{}).Where("user_id = ?", u.ID).Count(&downloadCount)

	continuousCheckinDays := consecutiveCheckinDays(a.Orm, u.ID, time.Now().UTC())
	write(c, 0, "ok", gin.H{"stats": gin.H{
		"points":                u.Points,
		"favoriteGameCount":     favoriteGameCount,
		"continuousCheckinDays": continuousCheckinDays,
		"downloadCount":         downloadCount,
	}})
}

func consecutiveCheckinDays(db *gorm.DB, userID uint, now time.Time) int {
	var records []models.CheckinRecord
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -366)
	if db.Where("user_id = ? AND checkin_at >= ?", userID, start).Find(&records).Error != nil {
		return 0
	}
	days := make(map[string]struct{}, len(records))
	for _, record := range records {
		days[record.CheckinAt.UTC().Format("2006-01-02")] = struct{}{}
	}
	count := 0
	for day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC); ; day = day.AddDate(0, 0, -1) {
		if _, ok := days[day.Format("2006-01-02")]; !ok {
			break
		}
		count++
	}
	return count
}
func Favorite(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	gid, _ := strconv.ParseUint(c.Param("gameId"), 10, 64)
	if gid == 0 {
		write(c, 20001, "游戏不存在", nil)
		return
	}
	var g gameModels.Game
	if a.Orm.First(&g, uint(gid)).Error != nil {
		write(c, 20001, "游戏不存在", nil)
		return
	}
	row := models.FavoriteGame{UserID: u.ID, GameID: uint(gid)}
	if c.Request.Method == http.MethodPost {
		a.Orm.Where("user_id = ? AND game_id = ?", u.ID, gid).FirstOrCreate(&row)
	} else {
		a.Orm.Where("user_id = ? AND game_id = ?", u.ID, gid).Delete(&models.FavoriteGame{})
	}
	write(c, 0, "ok", nil)
}
func AccountList(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	kind := c.Param("kind")
	if kind == "favorites" {
		page, pageSize := accountPage(c)
		var total int64
		a.Orm.Model(&models.FavoriteGame{}).Where("user_id = ?", u.ID).Count(&total)
		var rows []models.FavoriteGame
		a.Orm.Where("user_id = ?", u.ID).Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows)
		list := make([]gin.H, 0, len(rows))
		for _, row := range rows {
			var game gameModels.Game
			if a.Orm.First(&game, row.GameID).Error != nil {
				continue
			}
			lastPlayedAt, serverName := latestGamePlay(a.Orm, u.ID, row.GameID)
			list = append(list, gin.H{"gameId": fmt.Sprintf("game_%d", game.ID), "gameName": game.Name, "serverName": serverName, "lastPlayedAt": lastPlayedAt, "installed": hasDownload(a.Orm, u.ID, row.GameID), "favorite": true})
		}
		write(c, 0, "ok", gin.H{"list": list, "page": page, "pageSize": pageSize, "total": total})
		return
	}
	if kind == "recent" {
		var rows []models.RecentGame
		a.Orm.Where("user_id = ?", u.ID).Order("visited_at DESC").Find(&rows)
		games := make([]gin.H, 0, len(rows))
		for _, row := range rows {
			var game gameModels.Game
			if a.Orm.First(&game, row.GameID).Error != nil {
				continue
			}
			serverName := ""
			if row.ServerID != 0 {
				var server serverModels.Server
				if a.Orm.First(&server, row.ServerID).Error == nil {
					serverName = server.Name
				}
			}
			games = append(games, gin.H{"id": fmt.Sprintf("game_%d", game.ID), "name": game.Name, "serverName": serverName, "lastPlayedAt": row.VisitedAt, "iconUrl": game.IconURL, "coverUrl": ""})
		}
		write(c, 0, "ok", gin.H{"games": games})
		return
	}
	if kind == "recent-server" {
		var rows []models.RecentServer
		a.Orm.Where("user_id = ?", u.ID).Order("visited_at DESC").Find(&rows)
		write(c, 0, "ok", rows)
		return
	}
	var rows []models.DownloadRecord
	page, pageSize := accountPage(c)
	var total int64
	a.Orm.Model(&models.DownloadRecord{}).Where("user_id = ?", u.ID).Count(&total)
	a.Orm.Where("user_id = ?", u.ID).Order("downloaded_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows)
	list := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var game gameModels.Game
		if a.Orm.First(&game, row.GameID).Error != nil {
			continue
		}
		status := row.Status
		if status == "" {
			status = "completed"
		}
		progress := row.Progress
		if progress == 0 && status == "completed" {
			progress = 100
		}
		list = append(list, gin.H{
			"id": fmt.Sprintf("download_%d", row.ID), "gameId": fmt.Sprintf("game_%d", game.ID),
			"gameName": game.Name, "version": row.Version, "size": row.Size,
			"status": status, "progress": progress, "downloadedAt": row.DownloadedAt,
		})
	}
	write(c, 0, "ok", gin.H{"list": list, "page": page, "pageSize": pageSize, "total": total})
}

func accountPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func latestGamePlay(db *gorm.DB, userID, gameID uint) (*time.Time, string) {
	var row models.RecentGame
	if db.Where("user_id = ? AND game_id = ?", userID, gameID).Order("visited_at DESC").First(&row).Error != nil {
		return nil, ""
	}
	var server serverModels.Server
	if row.ServerID != 0 {
		db.First(&server, row.ServerID)
	}
	return &row.VisitedAt, server.Name
}

func hasDownload(db *gorm.DB, userID, gameID uint) bool {
	var n int64
	db.Model(&models.DownloadRecord{}).Where("user_id = ? AND game_id = ?", userID, gameID).Count(&n)
	return n > 0
}
func ChangePassword(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var in struct {
		OldPassword          string `json:"oldPassword"`
		NewPassword          string `json:"newPassword"`
		PasswordConfirmation string `json:"passwordConfirmation"`
		LogoutOtherDevices   bool   `json:"logoutOtherDevices"`
	}
	if c.ShouldBindJSON(&in) != nil || in.NewPassword != in.PasswordConfirmation || len(in.NewPassword) < 8 {
		write(c, accountProfileInvalid, "密码参数不符合规则", nil)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.OldPassword)) != nil {
		write(c, accountOldPassword, "旧密码错误", nil)
		return
	}
	h, _ := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	now := time.Now().UTC()
	updates := map[string]interface{}{"password_hash": string(h), "password_updated_at": now}
	if in.LogoutOtherDevices {
		updates["token_version"] = gorm.Expr("token_version + 1")
	}
	a.Orm.Model(u).Updates(updates)
	write(c, 0, "ok", nil)
}
func Messages(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var rows []models.Message
	a.Orm.Where("user_id = ?", u.ID).Order("created_at DESC").Find(&rows)
	list := make([]gin.H, 0, len(rows))
	var unreadCount int64
	a.Orm.Model(&models.Message{}).Where("user_id = ? AND read_at IS NULL", u.ID).Count(&unreadCount)
	for _, row := range rows {
		list = append(list, gin.H{"id": fmt.Sprintf("message_%d", row.ID), "type": row.Type, "title": row.Title, "content": row.Content, "read": row.ReadAt != nil, "createdAt": row.CreatedAt})
	}
	write(c, 0, "ok", gin.H{"list": list, "unreadCount": unreadCount})
}
func MessageRead(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	id := messageID(c.Param("id"))
	a.Orm.Model(&models.Message{}).Where("id = ? AND user_id = ?", id, u.ID).Update("read_at", time.Now().UTC())
	write(c, 0, "ok", nil)
}

func messageID(raw string) uint {
	raw = strings.TrimPrefix(raw, "message_")
	id, _ := strconv.ParseUint(raw, 10, 64)
	return uint(id)
}

func MessageReadAll(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	a.Orm.Model(&models.Message{}).Where("user_id = ? AND read_at IS NULL", u.ID).Update("read_at", time.Now().UTC())
	write(c, 0, "ok", nil)
}
func Devices(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var rows []models.DeviceAccount
	a.Orm.Where("user_id = ?", u.ID).Order("last_login_at DESC").Find(&rows)
	write(c, 0, "ok", rows)
}
func DeviceDelete(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	q := a.Orm.Where("user_id = ? AND device_id = ?", u.ID, c.Param("deviceId"))
	var d models.DeviceAccount
	if q.First(&d).Error != nil {
		write(c, accountDeviceMissing, "设备会话不存在", nil)
		return
	}
	a.Orm.Delete(&d)
	a.Orm.Where("user_id = ?", u.ID).Delete(&models.RefreshToken{})
	write(c, 0, "ok", nil)
}
func LoginRecords(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var rows []models.LoginRecord
	a.Orm.Where("user_id = ?", u.ID).Order("login_at DESC").Find(&rows)
	write(c, 0, "ok", rows)
}
func Security(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	emailMasked := ""
	if u.Email != "" {
		emailMasked = maskAccount(u.Email)
	}
	write(c, 0, "ok", gin.H{
		"phone":    gin.H{"bound": u.PhoneCiphertext != "", "masked": maskPhone(u.PhoneCiphertext)},
		"email":    gin.H{"bound": u.Email != "", "masked": emailMasked},
		"realName": gin.H{"status": u.RealNameStatus},
		"password": gin.H{"updatedAt": u.PasswordUpdatedAt},
	})
}

func Settings(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	settings := loadSettings(a.Orm, u.ID)
	write(c, 0, "ok", settingsResponse(settings))
}

func UpdateSettings(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	settings := loadSettings(a.Orm, u.ID)
	var in struct {
		Privacy *struct {
			ShowOnlineStatus *bool `json:"showOnlineStatus"`
			AllowMessages    *bool `json:"allowMessages"`
		} `json:"privacy"`
		Notification *struct {
			System   *bool `json:"system"`
			Activity *bool `json:"activity"`
			Live     *bool `json:"live"`
		} `json:"notification"`
	}
	if c.ShouldBindJSON(&in) != nil {
		write(c, accountProfileInvalid, "设置参数错误", nil)
		return
	}
	if in.Privacy != nil {
		if in.Privacy.ShowOnlineStatus != nil {
			settings.ShowOnlineStatus = *in.Privacy.ShowOnlineStatus
		}
		if in.Privacy.AllowMessages != nil {
			settings.AllowMessages = *in.Privacy.AllowMessages
		}
	}
	if in.Notification != nil {
		if in.Notification.System != nil {
			settings.NotifySystem = *in.Notification.System
		}
		if in.Notification.Activity != nil {
			settings.NotifyActivity = *in.Notification.Activity
		}
		if in.Notification.Live != nil {
			settings.NotifyLive = *in.Notification.Live
		}
	}
	a.Orm.Save(settings)
	write(c, 0, "ok", settingsResponse(settings))
}

func loadSettings(db *gorm.DB, userID uint) *models.UserSettings {
	var s models.UserSettings
	if db.Where("user_id = ?", userID).First(&s).Error != nil {
		s.UserID = userID
		db.Create(&s)
	}
	return &s
}
func settingsResponse(s *models.UserSettings) gin.H {
	return gin.H{"privacy": gin.H{"showOnlineStatus": s.ShowOnlineStatus, "allowMessages": s.AllowMessages}, "notification": gin.H{"system": s.NotifySystem, "activity": s.NotifyActivity, "live": s.NotifyLive}}
}

func Rewards(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	page, pageSize := accountPage(c)
	var total int64
	a.Orm.Model(&models.RewardRecord{}).Where("user_id = ?", u.ID).Count(&total)
	var rows []models.RewardRecord
	a.Orm.Where("user_id = ?", u.ID).Order("claimed_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows)
	list := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		list = append(list, gin.H{"id": fmt.Sprintf("reward_%d", row.ID), "name": row.Name, "code": row.Code, "status": row.Status, "claimedAt": row.ClaimedAt})
	}
	write(c, 0, "ok", gin.H{"list": list, "page": page, "pageSize": pageSize, "total": total})
}

func BindSend(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	kind := c.Param("kind")
	var in struct {
		Target string `json:"target"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Target) == "" {
		write(c, accountProfileInvalid, "绑定参数错误", nil)
		return
	}
	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	row := models.BindCode{UserID: u.ID, Kind: kind, Target: strings.TrimSpace(in.Target), CodeHash: hash(code), ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	if a.Orm.Create(&row).Error != nil {
		write(c, codeServerError, "验证码发送失败", nil)
		return
	}
	write(c, 0, "ok", gin.H{"expireIn": 600})
}

func BindConfirm(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	kind := c.Param("kind")
	var in struct {
		Target string `json:"target"`
		Code   string `json:"code"`
	}
	if c.ShouldBindJSON(&in) != nil {
		write(c, accountCodeInvalid, "验证码错误或过期", nil)
		return
	}
	var row models.BindCode
	if a.Orm.Where("user_id = ? AND kind = ? AND target = ? AND used_at IS NULL", u.ID, kind, strings.TrimSpace(in.Target)).Order("id DESC").First(&row).Error != nil || time.Now().UTC().After(row.ExpiresAt) || hash(in.Code) != row.CodeHash {
		write(c, accountCodeInvalid, "验证码错误或过期", nil)
		return
	}
	now := time.Now().UTC()
	a.Orm.Model(&row).Update("used_at", now)
	updates := map[string]interface{}{}
	if kind == "phone" {
		updates["phone_ciphertext"] = row.Target
		updates["phone_hash"] = hash(row.Target)
	} else {
		updates["email"] = row.Target
	}
	a.Orm.Model(u).Updates(updates)
	write(c, 0, "ok", nil)
}
