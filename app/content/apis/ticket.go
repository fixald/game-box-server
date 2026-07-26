package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/content/models"
	"go-admin/common/apis"
	"strconv"
)

// CreateReport submits a client report for a live room.
// @Summary Submit client report
// @Tags client-content
// @Accept json
// @Produce json
// @Param request body reportInput true "Report request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/client/reports [post]
func CreateReport(c *gin.Context) {
	a, ok := base(c)
	if !ok {
		return
	}
	claims := jwt.ExtractClaims(c)
	userID, err := strconv.ParseUint(fmt.Sprint(claims["sub"]), 10, 32)
	if err != nil || userID == 0 {
		a.Error(10001, err, "用户身份无效")
		return
	}
	var in reportInput
	if err := c.ShouldBindJSON(&in); err != nil || in.TargetType == "" || in.TargetID == "" || in.Reason == "" {
		a.Error(90001, nil, "参数错误")
		return
	}
	if in.TargetType != "live_room" || len(in.TargetID) > 64 || len(in.Reason) > 64 || len(in.Detail) > 2000 {
		a.Error(90001, nil, "举报目标或内容无效")
		return
	}
	row := models.Report{UserID: uintPtr(uint(userID)), TargetType: in.TargetType, TargetID: in.TargetID, Reason: in.Reason, Detail: in.Detail, Status: "submitted"}
	if err := a.Orm.Create(&row).Error; err != nil {
		a.Error(90002, err, "提交举报失败")
		return
	}
	a.OK(gin.H{"reportId": fmt.Sprintf("report_%d", row.ID), "status": row.Status}, "ok")
}

type reportInput struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Reason     string `json:"reason"`
	Detail     string `json:"detail"`
}

func uintPtr(value uint) *uint { return &value }

func base(c *gin.Context) (*apis.Api, bool) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		a.Error(90002, a.Errors, "数据库连接失败")
		return a, false
	}
	return a, true
}
func Reports(c *gin.Context) {
	a, ok := base(c)
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
	q := a.Orm.Model(&models.Report{})
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}
	var n int64
	q.Count(&n)
	var rows []models.Report
	q.Order("id DESC").Offset((p - 1) * s).Limit(s).Find(&rows)
	a.PageOK(rows, int(n), p, s, "success")
}
func ProcessReport(c *gin.Context) {
	a, ok := base(c)
	if !ok {
		return
	}
	var in struct {
		Status      string `json:"status" binding:"required"`
		HandlerNote string `json:"handlerNote"`
	}
	if c.ShouldBindJSON(&in) != nil {
		a.Error(90001, nil, "参数错误")
		return
	}
	if in.Status != "processing" && in.Status != "resolved" && in.Status != "rejected" && in.Status != "closed" {
		a.Error(90001, nil, "状态错误")
		return
	}
	if a.Orm.Model(&models.Report{}).Where("id = ?", c.Param("id")).Updates(map[string]interface{}{"status": in.Status, "handler_note": in.HandlerNote}).RowsAffected == 0 {
		a.Error(90001, nil, "举报不存在")
		return
	}
	a.OK(nil, "success")
}
func Feedbacks(c *gin.Context) {
	a, ok := base(c)
	if !ok {
		return
	}
	var rows []models.Feedback
	q := a.Orm.Model(&models.Feedback{})
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}
	q.Order("id DESC").Find(&rows)
	a.OK(rows, "success")
}
func ProcessFeedback(c *gin.Context) {
	a, ok := base(c)
	if !ok {
		return
	}
	var in struct {
		Status string `json:"status" binding:"required"`
		Result string `json:"result"`
	}
	if c.ShouldBindJSON(&in) != nil {
		a.Error(90001, nil, "参数错误")
		return
	}
	if a.Orm.Model(&models.Feedback{}).Where("id = ?", c.Param("id")).Updates(map[string]interface{}{"status": in.Status, "result": in.Result}).RowsAffected == 0 {
		a.Error(90001, nil, "反馈不存在")
		return
	}
	a.OK(nil, "success")
}
