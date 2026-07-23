package apis

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go-admin/app/cauth/models"

	"github.com/gin-gonic/gin"
)

type passwordResetRequest struct {
	Account  string `json:"account" binding:"required"`
	DeviceID string `json:"deviceId"`
}

type passwordResetConfirm struct {
	Account     string `json:"account" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
	DeviceID    string `json:"deviceId"`
}

func PasswordResetRequest(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in passwordResetRequest
	if c.ShouldBindJSON(&in) != nil {
		write(c, codeServerError, "参数错误", nil)
		return
	}
	account := strings.TrimSpace(in.Account)
	if account == "" {
		write(c, codeServerError, "参数错误", nil)
		return
	}
	ah := hash(account)
	var last models.PasswordReset
	if a.Orm.Where("account_hash = ?", ah).Order("created_at DESC").First(&last).Error == nil && time.Since(last.CreatedAt) < smsCooldown {
		write(c, codeSMSTooFrequent, "验证码发送过于频繁", nil)
		return
	}

	// 无论账号是否存在都返回成功，避免账号枚举；有账号时写入重置码。
	var u models.User
	exists := a.Orm.Where("account_hash = ? OR phone_hash = ? OR account = ?", ah, ah, account).First(&u).Error == nil
	if exists {
		code := fmt.Sprintf("%06d", rand.Intn(1000000))
		row := models.PasswordReset{
			AccountHash: ah,
			CodeHash:    hash(code),
			ExpiresAt:   time.Now().UTC().Add(smsTTL),
		}
		if err := a.Orm.Create(&row).Error; err != nil {
			write(c, codeServerError, "服务端异常", nil)
			return
		}
		_ = code // 真实环境通过短信/邮件下发
		//打印日志
		fmt.Println("password reset code:", code)
	}
	write(c, codeOK, "ok", gin.H{
		"accountMasked": maskAccount(account),
		"expireIn":      int(smsTTL.Seconds()),
	})
}

func PasswordResetConfirm(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in passwordResetConfirm
	if c.ShouldBindJSON(&in) != nil {
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	account := strings.TrimSpace(in.Account)
	if !validatePassword(in.NewPassword) {
		write(c, codeRegPasswordRule, "密码需为 8-32 个字符，并包含至少一个英文字母和一个数字", nil)
		return
	}
	ah := hash(account)
	var row models.PasswordReset
	if err := a.Orm.Where("account_hash = ? AND used_at IS NULL", ah).Order("created_at DESC").First(&row).Error; err != nil {
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	now := time.Now().UTC()
	if now.After(row.ExpiresAt) {
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	if hash(in.Code) != row.CodeHash {
		row.FailedAttempts++
		a.Orm.Save(&row)
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}

	var u models.User
	if a.Orm.Where("account_hash = ? OR phone_hash = ? OR account = ?", ah, ah, account).First(&u).Error != nil {
		write(c, codeBadCredential, "账号或密码错误", nil)
		return
	}
	if u.Status != "active" || isLoginBanned(a.Orm, u.ID) {
		write(c, codeAccountBanned, "账号被封禁", nil)
		return
	}
	pw, err := hashPassword(in.NewPassword)
	if err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	if err := a.Orm.Model(&row).Update("used_at", now).Error; err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	if err := a.Orm.Model(&u).Updates(map[string]interface{}{
		"password_hash": pw,
		"token_version": u.TokenVersion + 1,
		"account":       account,
		"account_hash":  ah,
	}).Error; err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	a.Orm.Model(&models.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", u.ID).Update("revoked_at", now)
	u.PasswordHash = pw
	u.TokenVersion++
	issueSession(c, a.Orm, &u, in.DeviceID, false)
}
