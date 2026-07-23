package apis

import (
	"github.com/gin-gonic/gin"
	"go-admin/app/cauth/models"
	"go-admin/common/apis"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type banInput struct {
	BanType   string     `json:"banType" binding:"required"`
	Reason    string     `json:"reason" binding:"required"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

func adminDB(c *gin.Context) (*apis.Api, bool) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return a, false
	}
	return a, true
}
func UserList(c *gin.Context) {
	a, ok := adminDB(c)
	if !ok {
		return
	}
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
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
	q := a.Orm.Model(&models.User{})
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}
	if v := c.Query("keyword"); v != "" {
		q = q.Where("phone_hash = ? OR nickname LIKE ?", hash(v), "%"+v+"%")
	}
	var n int64
	q.Count(&n)
	var rows []models.User
	if q.Order("id DESC").Offset((p-1)*s).Limit(s).Find(&rows).Error != nil {
		a.Error(90002, nil, "查询失败")
		return
	}
	a.PageOK(rows, int(n), p, s, "success")
}
func UserGet(c *gin.Context) {
	a, ok := adminDB(c)
	if !ok {
		return
	}
	var u models.User
	if a.Orm.First(&u, c.Param("id")).Error != nil {
		a.Error(10007, nil, "用户不存在")
		return
	}
	a.OK(u, "success")
}
func BanList(c *gin.Context) {
	a, ok := adminDB(c)
	if !ok {
		return
	}
	var rows []models.UserBan
	if a.Orm.Where("user_id = ?", c.Param("id")).Order("id DESC").Find(&rows).Error != nil {
		a.Error(90002, nil, "查询失败")
		return
	}
	a.OK(rows, "success")
}
func Ban(c *gin.Context) {
	a, ok := adminDB(c)
	if !ok {
		return
	}
	var in banInput
	if c.ShouldBindJSON(&in) != nil || in.BanType != "mute" && in.BanType != "login" && in.BanType != "game" && in.BanType != "all" {
		a.Error(90001, nil, "封禁参数错误")
		return
	}
	var u models.User
	if a.Orm.First(&u, c.Param("id")).Error != nil {
		a.Error(10007, nil, "用户不存在")
		return
	}
	now := time.Now().UTC()
	var existing models.UserBan
	if a.Orm.Where("user_id = ? AND ban_type = ? AND status = ?", u.ID, in.BanType, "active").First(&existing).Error == nil {
		a.OK(existing, "success")
		return
	}
	row := models.UserBan{UserID: u.ID, BanType: in.BanType, Reason: in.Reason, Source: "admin", StartsAt: now, ExpiresAt: in.ExpiresAt, Status: "active"}
	err := a.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{"status": "banned", "token_version": u.TokenVersion + 1}
		if err := tx.Model(&u).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Model(&models.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", u.ID).Update("revoked_at", now).Error
	})
	if err != nil {
		a.Error(90002, err, "封禁失败")
		return
	}
	a.OK(row, "success")
}
func Unban(c *gin.Context) {
	a, ok := adminDB(c)
	if !ok {
		return
	}
	var row models.UserBan
	if a.Orm.Where("id = ? AND status = ?", c.Param("id"), "active").First(&row).Error != nil {
		a.Error(10007, nil, "封禁记录不存在")
		return
	}
	now := time.Now().UTC()
	if err := a.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&row).Updates(map[string]interface{}{"status": "lifted", "lifted_at": now}).Error; err != nil {
			return err
		}
		var count int64
		tx.Model(&models.UserBan{}).Where("user_id = ? AND status = ? AND id <> ?", row.UserID, "active", row.ID).Count(&count)
		if count == 0 {
			return tx.Model(&models.User{}).Where("id = ?", row.UserID).Update("status", "active").Error
		}
		return nil
	}); err != nil {
		a.Error(90002, err, "解封失败")
		return
	}
	a.OK(nil, "success")
}
