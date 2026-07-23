package apis

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"go-admin/app/cauth/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type checkAccountRequest struct {
	Account  string `json:"account" binding:"required"`
	DeviceID string `json:"deviceId"`
}

type registerRequest struct {
	Account              string `json:"account" binding:"required"`
	Password             string `json:"password" binding:"required"`
	PasswordConfirmation string `json:"passwordConfirmation" binding:"required"`
	DeviceID             string `json:"deviceId"`
	InviteCode           string `json:"inviteCode"`
	AgreementVersion     string `json:"agreementVersion"`
	ConsentID            string `json:"consentId"`
	VerifyToken          string `json:"verifyToken"`
}

type registerVerifyRequest struct {
	DeviceID    string `json:"deviceId" binding:"required"`
	CaptchaID   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
	Phone       string `json:"phone"`
	Code        string `json:"code"`
}

func validateAccountFormat(account string) (code int, msg string) {
	account = strings.TrimSpace(account)
	n := utf8.RuneCountInString(account)
	if n < accountMinLen || n > accountMaxLen {
		return codeRegAccountLength, "账号长度不符合要求"
	}
	if phoneRE.MatchString(account) {
		return 0, ""
	}
	if strings.Contains(account, "@") {
		if at := strings.IndexByte(account, '@'); at < 1 || at == len(account)-1 {
			return codeRegAccountFormat, "账号格式错误"
		}
		return 0, ""
	}
	if !accountFormatRE.MatchString(account) {
		return codeRegAccountFormat, "账号格式错误"
	}
	return 0, ""
}

func validatePassword(password string) bool {
	n := utf8.RuneCountInString(password)
	if n < passwordMinLen || n > passwordMaxLen {
		return false
	}
	return passwordLetter.MatchString(password) && passwordDigit.MatchString(password)
}

func accountExists(db *gorm.DB, account string) bool {
	ah := hash(account)
	var n int64
	db.Model(&models.User{}).Where("account_hash = ? OR account = ? OR phone_hash = ?", ah, account, ah).Count(&n)
	return n > 0
}

func CheckAccount(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in checkAccountRequest
	if c.ShouldBindJSON(&in) != nil {
		write(c, codeRegAccountFormat, "账号格式错误", nil)
		return
	}
	account := strings.TrimSpace(in.Account)
	if code, msg := validateAccountFormat(account); code != 0 {
		write(c, code, msg, gin.H{"account": account, "available": false})
		return
	}
	if accountExists(a.Orm, account) {
		write(c, codeRegAccountExists, "账号已存在", gin.H{"account": account, "available": false})
		return
	}
	write(c, codeOK, "账号可用", gin.H{"account": account, "available": true})
}

func RegisterVerify(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in registerVerifyRequest
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.DeviceID) == "" {
		write(c, codeRegNeedVerify, "需要图形验证码/短信验证", nil)
		return
	}

	passed := false
	// 短信验证优先
	if phoneRE.MatchString(in.Phone) && strings.TrimSpace(in.Code) != "" {
		ph := hash(in.Phone)
		var row models.SMSCode
		now := time.Now().UTC()
		if a.Orm.Where("phone_hash = ? AND used_at IS NULL", ph).Order("created_at DESC").First(&row).Error == nil &&
			!now.After(row.ExpiresAt) && hash(in.Code) == row.CodeHash {
			_ = a.Orm.Model(&row).Update("used_at", now).Error
			passed = true
		}
	}
	// 开发态：图形验证码占位（非空且长度>=4 即通过）；生产应接入真实 captcha。
	if !passed && strings.TrimSpace(in.CaptchaID) != "" && len(strings.TrimSpace(in.CaptchaCode)) >= 4 {
		passed = true
	}
	if !passed {
		write(c, codeRegNeedVerify, "需要图形验证码/短信验证", nil)
		return
	}

	raw := randomToken()
	if raw == "" {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	now := time.Now().UTC()
	ticket := models.RegisterTicket{
		DeviceID:  in.DeviceID,
		TokenHash: hash(raw),
		ExpiresAt: now.Add(registerTicketTTL),
	}
	if err := a.Orm.Create(&ticket).Error; err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	write(c, codeOK, "ok", gin.H{
		"verifyToken": raw,
		"expiresIn":   int(registerTicketTTL.Seconds()),
	})
}

func consumeRegisterTicket(db *gorm.DB, deviceID, token string) bool {
	if strings.TrimSpace(token) == "" {
		return false
	}
	var row models.RegisterTicket
	now := time.Now().UTC()
	if err := db.Where("token_hash = ? AND used_at IS NULL", hash(token)).First(&row).Error; err != nil {
		return false
	}
	if deviceID != "" && row.DeviceID != deviceID {
		return false
	}
	if now.After(row.ExpiresAt) {
		return false
	}
	return db.Model(&row).Update("used_at", now).Error == nil
}

func registerRateLimited(db *gorm.DB, deviceID, clientIP string) (tooFrequent, needVerify, blocked bool) {
	now := time.Now().UTC()
	q := db.Model(&models.RegisterEvent{})
	if deviceID != "" {
		q = q.Where("device_id = ?", deviceID)
	} else if clientIP != "" {
		q = q.Where("client_ip = ?", clientIP)
	} else {
		return false, false, false
	}
	var recent int64
	q.Where("created_at >= ?", now.Add(-registerCooldown)).Count(&recent)
	if recent > 0 {
		return true, false, false
	}
	q2 := db.Model(&models.RegisterEvent{})
	if deviceID != "" {
		q2 = q2.Where("device_id = ?", deviceID)
	} else {
		q2 = q2.Where("client_ip = ?", clientIP)
	}
	q2.Where("created_at >= ?", now.Add(-time.Hour)).Count(&recent)
	if recent >= registerHourlyMax {
		return false, false, true
	}
	if recent >= 2 {
		return false, true, false
	}
	return false, false, false
}

func validateInvite(db *gorm.DB, code string) (*models.InviteCode, bool) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, true
	}
	var row models.InviteCode
	now := time.Now().UTC()
	if err := db.Where("code = ? AND status = ?", code, "active").First(&row).Error; err != nil {
		return nil, false
	}
	if row.ExpiresAt != nil && now.After(*row.ExpiresAt) {
		return nil, false
	}
	if row.MaxUses > 0 && row.UsedCount >= row.MaxUses {
		return nil, false
	}
	return &row, true
}

func Register(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in registerRequest
	if c.ShouldBindJSON(&in) != nil {
		write(c, codeRegAccountFormat, "账号格式错误", nil)
		return
	}
	account := strings.TrimSpace(in.Account)
	if code, msg := validateAccountFormat(account); code != 0 {
		write(c, code, msg, nil)
		return
	}
	if !validatePassword(in.Password) {
		write(c, codeRegPasswordRule, "密码不符合规则", nil)
		return
	}
	if in.Password != in.PasswordConfirmation {
		write(c, codeRegPasswordMismatch, "两次密码不一致", nil)
		return
	}
	// // 检查用户协议
	// if !ensureAgreementWithCode(c, a.Orm, in.AgreementVersion, codeRegAgreementNeeded, "需要重新确认用户协议") {
	// 	return
	// }
	// _ = in.ConsentID
	// 检查邀请码
	invite, inviteOK := validateInvite(a.Orm, in.InviteCode)
	if !inviteOK {
		write(c, codeRegInviteInvalid, "邀请码无效", nil)
		return
	}

	tooFrequent, needVerify, blocked := registerRateLimited(a.Orm, in.DeviceID, c.ClientIP())
	if blocked {
		write(c, codeRegRiskBlocked, "账号注册被风控拦截", nil)
		return
	}
	if tooFrequent {
		write(c, codeRegTooFrequent, "注册过于频繁", nil)
		return
	}
	if needVerify && !consumeRegisterTicket(a.Orm, in.DeviceID, in.VerifyToken) {
		write(c, codeRegNeedVerify, "需要图形验证码/短信验证", nil)
		return
	}
	// 主动带了 verifyToken 也消费掉（幂等失败则忽略，未强制时不拦截）
	if !needVerify && in.VerifyToken != "" {
		_ = consumeRegisterTicket(a.Orm, in.DeviceID, in.VerifyToken)
	}

	if accountExists(a.Orm, account) {
		write(c, codeRegAccountExists, "账号已存在", nil)
		return
	}

	pw, err := hashPassword(in.Password)
	if err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	ah := hash(account)
	// 账号注册无手机号：用 acct: 前缀保证 phone_hash 唯一且不与真实手机冲突
	phoneHash := hash("acct:" + account)
	nickname := "玩家"
	if n := utf8.RuneCountInString(account); n >= 3 {
		runes := []rune(account)
		if len(runes) > 3 {
			nickname = "玩家" + string(runes[len(runes)-3:])
		} else {
			nickname = "玩家" + account
		}
	}
	masked := maskAccount(account)
	u := models.User{
		PhoneHash:        phoneHash,
		PhoneCiphertext:  masked,
		Account:          account,
		AccountHash:      ah,
		PasswordHash:     pw,
		Nickname:         nickname,
		Status:           "active",
		AgreementVersion: strings.TrimSpace(in.AgreementVersion),
	}

	err = a.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		if invite != nil {
			if err := tx.Model(invite).UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
				return err
			}
		}
		evDevice := in.DeviceID
		if evDevice == "" {
			evDevice = "unknown"
		}
		return tx.Create(&models.RegisterEvent{
			DeviceID: evDevice,
			Account:  account,
			ClientIP: c.ClientIP(),
		}).Error
	})
	if err != nil {
		if accountExists(a.Orm, account) {
			write(c, codeRegAccountExists, "账号已存在", nil)
			return
		}
		write(c, codeServerError, "服务端异常", nil)
		return
	}

	access, refresh, ok := createSession(c, a.Orm, &u, in.DeviceID, true)
	if !ok {
		return
	}
	write(c, codeOK, "注册成功", gin.H{
		"user": gin.H{
			"id":            fmt.Sprintf("user_%d", u.ID),
			"account":       account,
			"accountMasked": masked,
			"nickname":      u.Nickname,
			"avatarUrl":     nil,
		},
		"autoLogin":    true,
		"accessToken":  access,
		"refreshToken": refresh,
		"expiresIn":    int(accessTTL.Seconds()),
		"redirect":     gin.H{"type": "home"},
	})
}
