package apis

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-admin/app/cauth/models"
)

type phoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type smsLoginRequest struct {
	Phone            string `json:"phone" binding:"required"`
	Code             string `json:"code" binding:"required"`
	DeviceID         string `json:"deviceId"`
	AgreementVersion string `json:"agreementVersion"`
	ConsentID        string `json:"consentId"`
}

func SendSMS(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in phoneRequest
	if c.ShouldBindJSON(&in) != nil || !phoneRE.MatchString(in.Phone) {
		write(c, codeServerError, "手机号格式错误", nil)
		return
	}
	ph := hash(in.Phone)
	var last models.SMSCode
	if a.Orm.Where("phone_hash = ?", ph).Order("created_at DESC").First(&last).Error == nil && time.Since(last.CreatedAt) < smsCooldown {
		write(c, codeSMSTooFrequent, "验证码发送过于频繁", nil)
		return
	}
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	row := models.SMSCode{PhoneHash: ph, CodeHash: hash(code), ExpiresAt: time.Now().UTC().Add(smsTTL)}
	if err := a.Orm.Create(&row).Error; err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	// 真实环境由短信 provider 发送；验证码不进入响应。
	_ = code
	write(c, codeOK, "ok", gin.H{"maskedPhone": maskPhone(in.Phone), "expireIn": int(smsTTL.Seconds())})
}

func SMSLogin(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in smsLoginRequest
	if c.ShouldBindJSON(&in) != nil || !phoneRE.MatchString(in.Phone) {
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	if !ensureAgreement(c, a.Orm, in.AgreementVersion) {
		return
	}
	_ = in.ConsentID

	ph := hash(in.Phone)
	var row models.SMSCode
	if err := a.Orm.Where("phone_hash = ? AND used_at IS NULL", ph).Order("created_at DESC").First(&row).Error; err != nil {
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	now := time.Now().UTC()
	if now.After(row.ExpiresAt) {
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	if row.LockedUntil != nil && now.Before(*row.LockedUntil) {
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	if hash(in.Code) != row.CodeHash {
		row.FailedAttempts++
		if row.FailedAttempts >= 5 {
			t := now.Add(15 * time.Minute)
			row.LockedUntil = &t
		}
		a.Orm.Save(&row)
		write(c, codeSMSInvalid, "验证码错误或过期", nil)
		return
	}
	if err := a.Orm.Model(&row).Updates(map[string]interface{}{"used_at": now}).Error; err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}

	var u models.User
	if a.Orm.Where("phone_hash = ?", ph).First(&u).Error != nil {
		masked := maskPhone(in.Phone)
		u = models.User{
			PhoneHash:        ph,
			PhoneCiphertext:  masked,
			Account:          in.Phone,
			AccountHash:      ph,
			Nickname:         "玩家" + in.Phone[7:],
			Status:           "active",
			AgreementVersion: strings.TrimSpace(in.AgreementVersion),
		}
		if err := a.Orm.Create(&u).Error; err != nil {
			write(c, codeServerError, "服务端异常", nil)
			return
		}
	} else if in.AgreementVersion != "" {
		_ = a.Orm.Model(&u).Update("agreement_version", in.AgreementVersion).Error
		u.AgreementVersion = in.AgreementVersion
	}
	issueSession(c, a.Orm, &u, in.DeviceID, true)
}
