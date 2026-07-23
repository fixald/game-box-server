package apis

import (
	"github.com/gin-gonic/gin"
	"go-admin/app/cauth/models"
)

func AgreementsLatest(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	row, err := latestAgreement(a.Orm)
	if err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	write(c, codeOK, "ok", gin.H{
		"version":     row.Version,
		"title":       row.Title,
		"contentUrl":  row.ContentURL,
		"summary":     row.Summary,
		"publishedAt": row.PublishedAt,
	})
}

func Accounts(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	deviceID := c.Query("deviceId")
	if deviceID == "" {
		write(c, codeOK, "ok", gin.H{"list": []interface{}{}})
		return
	}
	var rows []models.DeviceAccount
	if err := a.Orm.Where("device_id = ?", deviceID).Order("last_login_at DESC").Limit(20).Find(&rows).Error; err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		list = append(list, gin.H{
			"userId":        r.UserID,
			"accountMasked": r.AccountMasked,
			"lastLoginAt":   r.LastLoginAt,
		})
	}
	write(c, codeOK, "ok", gin.H{"list": list})
}
