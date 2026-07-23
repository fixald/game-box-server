package apis

import (
	"strings"

	"go-admin/app/cauth/models"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Account          string `json:"account" binding:"required"`
	Password         string `json:"password" binding:"required"`
	RememberAccount  bool   `json:"rememberAccount"`
	DeviceID         string `json:"deviceId"`
	AgreementVersion string `json:"agreementVersion"`
	ConsentID        string `json:"consentId"`
}

func Login(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in loginRequest
	if c.ShouldBindJSON(&in) != nil {
		write(c, codeBadCredential, "账号或密码错误", nil)
		return
	}
	account := strings.TrimSpace(in.Account)
	if account == "" || strings.TrimSpace(in.Password) == "" {
		write(c, codeBadCredential, "账号或密码错误", nil)
		return
	}
	// if !ensureAgreement(c, a.Orm, in.AgreementVersion) {
	// 	return
	// }
	// _ = in.ConsentID

	var u models.User
	ah := hash(account)
	err := a.Orm.Where("account_hash = ? OR phone_hash = ? OR account = ?", ah, ah, account).First(&u).Error
	if err != nil || u.PasswordHash == "" || !checkPassword(u.PasswordHash, in.Password) {
		write(c, codeBadCredential, "账号或密码错误", nil)
		return
	}
	if in.AgreementVersion != "" {
		_ = a.Orm.Model(&u).Update("agreement_version", in.AgreementVersion).Error
		u.AgreementVersion = in.AgreementVersion
	}
	issueSession(c, a.Orm, &u, in.DeviceID, in.RememberAccount)
}
