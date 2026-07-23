package apis

import (
	"github.com/gin-gonic/gin"
	"go-admin/app/content/models"
	"go-admin/common/apis"
	"strconv"
)

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
