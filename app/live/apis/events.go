package apis

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/live/models"
	commonapis "go-admin/common/apis"
)

type clientEventInput struct {
	EventType    string `json:"eventType" binding:"required"`
	ResourceType string `json:"resourceType" binding:"required"`
	ResourceID   string `json:"resourceId" binding:"required"`
	Source       string `json:"source"`
	SessionID    string `json:"sessionId"`
}

// CreateEvent records client live exposure, click, enter and follow events.
// @Summary Record client event
// @Tags client-events
// @Accept json
// @Produce json
// @Param request body clientEventInput true "Event request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/client/events [post]
func CreateEvent(c *gin.Context) {
	a := new(commonapis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接获取失败")
		return
	}
	var in clientEventInput
	if err := c.ShouldBindJSON(&in); err != nil {
		a.Error(90001, err, "事件参数无效")
		return
	}
	in.EventType = strings.TrimSpace(in.EventType)
	in.ResourceType = strings.TrimSpace(in.ResourceType)
	in.ResourceID = strings.TrimSpace(in.ResourceID)
	validTypes := map[string]bool{"live_exposure": true, "live_click": true, "live_enter": true, "live_follow": true}
	if !validTypes[in.EventType] || (in.ResourceType != "live_room" && in.ResourceType != "streamer") || in.ResourceID == "" || len(in.Source) > 64 || len(in.SessionID) > 128 {
		a.Error(90001, nil, "事件参数无效")
		return
	}
	claims := jwt.ExtractClaims(c)
	var userID uint
	if _, err := fmt.Sscan(fmt.Sprint(claims["sub"]), &userID); err != nil || userID == 0 {
		a.Error(10001, nil, "用户身份无效")
		return
	}
	row := models.ClientEvent{UserID: userID, EventType: in.EventType, ResourceType: in.ResourceType, ResourceID: in.ResourceID, Source: in.Source, SessionID: in.SessionID, OccurredAt: time.Now().UTC()}
	if err := a.Orm.Create(&row).Error; err != nil {
		a.Error(90002, err, "事件记录失败")
		return
	}
	a.OK(nil, "ok")
}
