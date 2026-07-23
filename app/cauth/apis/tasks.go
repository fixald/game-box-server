package apis

import (
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

type taskDef struct {
	ID, Category, Title, Description, Icon string
	Target, Points                         int
	ActionLabel, ActionRoute               string
}

var taskDefs = []taskDef{
	{"task-1", "daily", "每日登录盒子", "登录客户端即可完成", "◷", 1, 20, "", ""}, {"task-2", "daily", "观看直播 10 分钟", "在直播频道观看任意直播", "▶", 10, 50, "去直播", "#/live"}, {"task-3", "game", "启动一次游戏", "启动任意已安装的传奇游戏", "⚔", 1, 0, "", ""}, {"task-4", "newbie", "完善个人资料", "上传头像并设置昵称", "◇", 1, 0, "去完善", "#/settings"}, {"task-5", "social", "关注 3 位主播", "关注你喜欢的传奇主播", "♡", 3, 30, "去关注", "#/live"}, {"task-6", "newbie", "完成首次下载", "下载并校验一款游戏", "↓", 1, 0, "", ""}, {"task-7", "daily", "浏览新服推荐", "查看今日推荐新区和开服信息", "◈", 1, 10, "去新服", "#/"}, {"task-8", "daily", "完成一次签到", "在任务中心完成每日签到", "✓", 1, 20, "", ""}, {"task-9", "game", "查看游戏详情", "浏览任意一款游戏的详情页面", "◉", 1, 15, "去游戏", "#/games"}, {"task-10", "game", "进入推荐区服", "从新服推荐中选择一个区服", "⚑", 1, 0, "", ""}, {"task-11", "social", "查看一条资讯", "阅读平台最新游戏资讯", "▤", 1, 10, "去资讯", "#/news"}, {"task-12", "newbie", "完成首次区服选择", "选择喜欢的游戏区服", "◇", 1, 0, "", ""},
}

func taskUser(c *gin.Context) (*models.User, bool) {
	a, ok := ormAPI(c)
	if !ok {
		return nil, false
	}
	u, ok := accountUser(c, a)
	return u, ok
}
func taskJSON(d taskDef, status string, progress int) gin.H {
	rewards := []gin.H{{"type": "points", "name": "积分", "amount": d.Points, "icon": "✦"}}
	return gin.H{"id": d.ID, "category": d.Category, "title": d.Title, "description": d.Description, "icon": d.Icon, "progress": progress, "target": d.Target, "status": status, "rewards": rewards, "actionLabel": d.ActionLabel, "actionRoute": d.ActionRoute}
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
	list := make([]gin.H, 0)
	completed := 0
	claimable := 0
	for _, d := range taskDefs {
		if category != "all" && d.Category != category {
			continue
		}
		var claim models.TaskClaim
		claimed := a.Orm.Where("user_id=? AND task_id=?", u.ID, d.ID).First(&claim).Error == nil
		status := "in_progress"
		progress := 0
		if d.ID == "task-1" || d.ID == "task-3" {
			progress = 1
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
	list := make([]gin.H, 0, len(taskDefs))
	for _, d := range taskDefs {
		var claim models.TaskClaim
		claimed := a.Orm.Where("user_id=? AND task_id=?", u.ID, d.ID).First(&claim).Error == nil
		status, progress := "in_progress", 0
		if d.ID == "task-1" || d.ID == "task-3" {
			status, progress = "claimable", 1
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
	var rec models.CheckinRecord
	if a.Orm.Where("user_id=? AND checkin_at=?", u.ID, t).First(&rec).Error == nil {
		taskWrite(c, 400, "今日已经签到", gin.H{"data": nil})
		return
	}
	if err := a.Orm.Transaction(func(tx *gorm.DB) error { return tx.Create(&models.CheckinRecord{UserID: u.ID, CheckinAt: t}).Error }); err != nil {
		taskWrite(c, 90001, "签到失败", gin.H{"data": nil})
		return
	}
	a.Orm.Model(u).UpdateColumn("points", gorm.Expr("points + ?", 20))
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
	var d *taskDef
	for i := range taskDefs {
		if taskDefs[i].ID == in.TaskID {
			d = &taskDefs[i]
		}
	}
	if d == nil {
		taskWrite(c, 400, "任务不存在", gin.H{"data": nil})
		return
	}
	var row models.TaskClaim
	if a.Orm.Where("user_id=? AND task_id=?", u.ID, in.TaskID).First(&row).Error == nil {
		taskWrite(c, 400, "任务奖励已领取", gin.H{"data": nil})
		return
	}
	if d.Points == 0 {
		d.Points = 20
	}
	a.Orm.Create(&models.TaskClaim{UserID: u.ID, TaskID: in.TaskID, Points: d.Points, ClaimedAt: time.Now().UTC()})
	a.Orm.Model(u).UpdateColumn("points", gorm.Expr("points + ?", d.Points))
	taskWrite(c, 200, "success", gin.H{"data": gin.H{"taskId": in.TaskID, "points": d.Points}})
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
