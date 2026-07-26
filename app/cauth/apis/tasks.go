package apis

import (
	"encoding/json"
	"go-admin/app/cauth/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func taskWrite(c *gin.Context, code int, msg string, body gin.H) {
	response := gin.H{"code": code, "msg": msg, "requestId": requestID(c)}
	for key, value := range body {
		response[key] = value
	}
	c.JSON(http.StatusOK, response)
}

func taskUser(c *gin.Context) (*models.User, bool) {
	a, ok := ormAPI(c)
	if !ok {
		return nil, false
	}
	u, ok := accountUser(c, a)
	return u, ok
}
func taskPoints(d models.Task) int {
	var rewards []struct {
		Amount int `json:"amount"`
	}
	if json.Unmarshal([]byte(d.Rewards), &rewards) == nil && len(rewards) > 0 {
		return rewards[0].Amount
	}
	return d.Points
}
func taskProgress(db *gorm.DB, u *models.User, d models.Task, now time.Time) int {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var n int64
	switch d.Code {
	case "task-1":
		db.Model(&models.LoginRecord{}).Where("user_id=? AND success=? AND login_at>=?", u.ID, true, start).Count(&n)
	case "task-3":
		db.Model(&models.RecentGame{}).Where("user_id=? AND visited_at>=?", u.ID, start).Count(&n)
	case "task-6":
		db.Model(&models.DownloadRecord{}).Where("user_id=? AND status=?", u.ID, "completed").Count(&n)
	case "task-8":
		db.Model(&models.CheckinRecord{}).Where("user_id=? AND checkin_at=?", u.ID, start).Count(&n)
	case "task-4":
		if u.Nickname != "" && u.AvatarURL != "" {
			n = 1
		}
	default:
		return 0
	}
	if n > int64(d.Target) {
		return d.Target
	}
	return int(n)
}
func taskJSON(d models.Task, status string, progress int) gin.H {
	var rewards interface{}
	if d.Rewards != "" {
		_ = json.Unmarshal([]byte(d.Rewards), &rewards)
	}
	return gin.H{"id": d.Code, "category": d.Category, "title": d.Title, "description": d.Description, "icon": d.Icon, "progress": progress, "target": d.Target, "status": status, "rewards": rewards, "actionLabel": d.ActionLabel, "actionRoute": d.ActionRoute}
}
func Tasks(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	category := c.Query("category")
	if category == "" {
		category = "all"
	}
	var defs []models.Task
	a.Orm.Where("status = ?", "active").Order("sort ASC, id ASC").Find(&defs)
	list := make([]gin.H, 0, len(defs))
	completed := 0
	claimable := 0
	for _, d := range defs {
		if category != "all" && d.Category != category {
			continue
		}
		var claim models.TaskClaim
		claimed := a.Orm.Where("user_id=? AND task_id=?", u.ID, d.Code).First(&claim).Error == nil
		status := "in_progress"
		progress := taskProgress(a.Orm, u, d, time.Now().UTC())
		if progress >= d.Target {
			status = "claimable"
		}
		if claimed {
			progress = d.Target
			status = "claimed"
		}
		if status == "claimed" {
			completed++
		}
		if status == "claimable" {
			claimable++
		}
		list = append(list, taskJSON(d, status, progress))
	}
	date := c.Query("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	dayStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		dayStart, _ = time.Parse("2006-01-02", time.Now().UTC().Format("2006-01-02"))
	}
	var day models.CheckinRecord
	// 有签到记录 ⇒ 今天已签到
	checkedToday := a.Orm.Where("user_id = ? AND checkin_at = ?", u.ID, dayStart).First(&day).Error == nil
	taskWrite(c, 200, "success", gin.H{"data": gin.H{"summary": gin.H{"points": u.Points, "continuousCheckinDays": consecutiveCheckinDays(a.Orm, u.ID, time.Now().UTC()), "totalCompleted": completed, "claimableCount": claimable, "checkin": gin.H{"checkedToday": checkedToday}}, "tasks": list}})
}
func TaskList(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var defs []models.Task
	a.Orm.Where("status = ?", "active").Order("sort ASC, id ASC").Find(&defs)
	list := make([]gin.H, 0, len(defs))
	for _, d := range defs {
		var claim models.TaskClaim
		claimed := a.Orm.Where("user_id=? AND task_id=?", u.ID, d.Code).First(&claim).Error == nil
		status, progress := "in_progress", 0
		progress = taskProgress(a.Orm, u, d, time.Now().UTC())
		if progress >= d.Target {
			status = "claimable"
		}
		if claimed {
			status, progress = "claimed", d.Target
		}
		list = append(list, taskJSON(d, status, progress))
	}
	taskWrite(c, 200, "success", gin.H{"data": list})
}
func Checkin(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var in struct {
		Date string `json:"date"`
	}
	if c.ShouldBindJSON(&in) != nil || in.Date == "" {
		taskWrite(c, 400, "日期无效", gin.H{"data": nil})
		return
	}
	t, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		taskWrite(c, 400, "日期无效", gin.H{"data": nil})
		return
	}
	if in.Date != time.Now().UTC().Format("2006-01-02") {
		taskWrite(c, 400, "只能签到当天", gin.H{"data": nil})
		return
	}
	var rec models.CheckinRecord
	if a.Orm.Where("user_id=? AND checkin_at=?", u.ID, t).First(&rec).Error == nil {
		taskWrite(c, 400, "今日已经签到", gin.H{"data": nil})
		return
	}
	if err := a.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.CheckinRecord{UserID: u.ID, CheckinAt: t}).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).Where("id = ?", u.ID).UpdateColumn("points", gorm.Expr("points + ?", 20)).Error
	}); err != nil {
		taskWrite(c, 90001, "签到失败", gin.H{"data": nil})
		return
	}
	taskWrite(c, 200, "success", gin.H{"data": gin.H{"date": in.Date, "points": 20}})
}
func ClaimTask(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var in struct {
		TaskID string `json:"taskId"`
	}
	if c.ShouldBindJSON(&in) != nil {
		taskWrite(c, 400, "任务无效", gin.H{"data": nil})
		return
	}
	var d models.Task
	if a.Orm.Where("code = ? AND status = ?", in.TaskID, "active").First(&d).Error != nil {
		taskWrite(c, 400, "任务不存在", gin.H{"data": nil})
		return
	}
	var row models.TaskClaim
	if a.Orm.Where("user_id=? AND task_id=?", u.ID, in.TaskID).First(&row).Error == nil {
		taskWrite(c, 400, "任务奖励已领取", gin.H{"data": nil})
		return
	}
	points := taskPoints(d)
	if taskProgress(a.Orm, u, d, time.Now().UTC()) < d.Target {
		taskWrite(c, 400, "任务尚未完成", gin.H{"data": nil})
		return
	}
	err := a.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.TaskClaim{UserID: u.ID, TaskID: in.TaskID, Points: points, ClaimedAt: time.Now().UTC()}).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).Where("id = ?", u.ID).UpdateColumn("points", gorm.Expr("points + ?", points)).Error
	})
	if err != nil {
		taskWrite(c, 90001, "领取失败", gin.H{"data": nil})
		return
	}
	taskWrite(c, 200, "success", gin.H{"data": gin.H{"taskId": in.TaskID, "points": points}})
}
func CheckinRewards(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var rows []models.CheckinReward
	if err := a.Orm.Where("status = ?", "active").Order("level ASC").Find(&rows).Error; err != nil {
		taskWrite(c, 500, "奖励加载失败", gin.H{"rewards": []gin.H{}})
		return
	}
	days := consecutiveCheckinDays(a.Orm, u.ID, time.Now().UTC())
	rewards := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		var claim models.CheckinRewardClaim
		claimed := a.Orm.Where("user_id = ? AND level = ?", u.ID, row.Level).First(&claim).Error == nil
		status := "locked"
		if days >= row.Level {
			status = "claimable"
		}
		if claimed {
			status = "claimed"
		}
		rewards = append(rewards, gin.H{"level": row.Level, "name": row.Name, "reward": row.Reward, "icon": row.Icon, "status": status})
	}
	taskWrite(c, 200, "success", gin.H{"rewards": rewards})
}
func ClaimCheckinReward(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	u, ok := accountUser(c, a)
	if !ok {
		return
	}
	var in struct {
		Level int `json:"level"`
	}
	if c.ShouldBindJSON(&in) != nil || in.Level <= 0 {
		taskWrite(c, 400, "奖励等级无效", gin.H{})
		return
	}
	var reward models.CheckinReward
	if a.Orm.Where("level = ? AND status = ?", in.Level, "active").First(&reward).Error != nil {
		taskWrite(c, 404, "奖励不存在", gin.H{})
		return
	}
	if consecutiveCheckinDays(a.Orm, u.ID, time.Now().UTC()) < reward.Level {
		taskWrite(c, 400, "尚未达到领取条件", gin.H{})
		return
	}
	var existing models.CheckinRewardClaim
	if a.Orm.Where("user_id = ? AND level = ?", u.ID, in.Level).First(&existing).Error == nil {
		taskWrite(c, 400, "奖励已领取", gin.H{})
		return
	}
	if err := a.Orm.Create(&models.CheckinRewardClaim{UserID: u.ID, Level: in.Level, ClaimedAt: time.Now().UTC()}).Error; err != nil {
		taskWrite(c, 500, "领取失败", gin.H{})
		return
	}
	taskWrite(c, 200, "success", gin.H{})
}
