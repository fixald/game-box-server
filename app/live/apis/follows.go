package apis

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/live/models"
	commonapis "go-admin/common/apis"
	"gorm.io/gorm"
)

func currentUserID(c *gin.Context) (uint, bool) {
	claims := jwt.ExtractClaims(c)
	raw := fmt.Sprint(claims["sub"])
	if raw == "<nil>" {
		raw = fmt.Sprint(claims["identity"])
	}
	id, err := strconv.ParseUint(raw, 10, 32)
	return uint(id), err == nil && id > 0
}

// FollowStreamer follows or unfollows a streamer. The route method determines the action.
// @Summary Follow or unfollow a live streamer
// @Tags client-live
// @Param id path string true "Streamer ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/client/live/streamers/{id}/follow [post]
// @Router /api/v1/client/live/streamers/{id}/follow [delete]
func FollowStreamer(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		a.Error(10001, nil, "用户身份无效")
		return
	}
	streamerID := c.Param("id")
	var streamer models.Streamer
	if err := a.Orm.First(&streamer, "id = ?", streamerID).Error; err != nil {
		a.Error(90001, err, "主播不存在")
		return
	}

	following := c.Request.Method == "POST"
	err := a.Orm.Transaction(func(tx *gorm.DB) error {
		var relation models.StreamerFollow
		findErr := tx.Where("user_id = ? AND streamer_id = ?", userID, streamerID).First(&relation).Error
		if following {
			if findErr == gorm.ErrRecordNotFound {
				if err := tx.Create(&models.StreamerFollow{UserID: userID, StreamerID: streamerID}).Error; err != nil {
					return err
				}
				return tx.Model(&models.Streamer{}).Where("id = ?", streamerID).UpdateColumn("fans", gorm.Expr("fans + 1")).Error
			}
			return findErr
		}
		if findErr == gorm.ErrRecordNotFound {
			return nil
		}
		if findErr != nil {
			return findErr
		}
		if err := tx.Delete(&relation).Error; err != nil {
			return err
		}
		return tx.Model(&models.Streamer{}).Where("id = ? AND fans > 0", streamerID).UpdateColumn("fans", gorm.Expr("fans - 1")).Error
	})
	if err != nil {
		a.Error(90002, err, "关注操作失败")
		return
	}
	a.Orm.Select("fans").First(&streamer, "id = ?", streamerID)
	a.OK(gin.H{"following": following, "fans": streamer.Fans}, "ok")
}

// MyFollowing returns the authenticated user's followed streamers.
// @Summary List followed live streamers
// @Tags client-live
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/client/users/me/live/following [get]
func MyFollowing(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		a.Error(10001, nil, "用户身份无效")
		return
	}
	var rows []models.Streamer
	if err := a.Orm.Joins("JOIN gb_live_streamer_follows f ON f.streamer_id = gb_live_streamers.id").Where("f.user_id = ?", userID).Order("f.created_at DESC").Find(&rows).Error; err != nil {
		a.Error(90002, err, "查询失败")
		return
	}
	for i := range rows {
		rows[i].Following = true
	}
	a.OK(gin.H{"list": rows}, "ok")
}
