package apis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-admin/app/cauth/models"
	"go-admin/common/apis"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	codeOK              = 0
	codeBadCredential   = 10001
	codeSMSInvalid      = 10002
	codeSMSTooFrequent  = 10003
	codeAccountBanned   = 10004
	codeTokenExpired    = 10005
	codeAgreementNeeded = 10007
	codeRiskChallenge   = 10008
	codeServerError     = 90001

	codeRegAccountFormat   = 11001
	codeRegAccountLength   = 11002
	codeRegAccountExists   = 11003
	codeRegPasswordRule    = 11004
	codeRegPasswordMismatch = 11005
	codeRegTooFrequent     = 11006
	codeRegNeedVerify      = 11007
	codeRegAgreementNeeded = 11008
	codeRegInviteInvalid   = 11009
	codeRegRiskBlocked     = 11010

	accessTTL          = 2 * time.Hour
	refreshTTL         = 30 * 24 * time.Hour
	smsCooldown         = 60 * time.Second
	smsTTL             = 5 * time.Minute
	registerCooldown   = 60 * time.Second
	registerHourlyMax  = 5
	registerTicketTTL  = 10 * time.Minute
	accountMinLen      = 4
	accountMaxLen      = 16
	passwordMinLen     = 8
	passwordMaxLen     = 32
)

var (
	phoneRE         = regexp.MustCompile(`^1[3-9][0-9]{9}$`)
	accountFormatRE = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	passwordLetter  = regexp.MustCompile(`[A-Za-z]`)
	passwordDigit   = regexp.MustCompile(`[0-9]`)
)

func hash(v string) string {
	s := sha256.Sum256([]byte(v))
	return hex.EncodeToString(s[:])
}

func requestID(c *gin.Context) string {
	if v := c.GetHeader("X-Request-ID"); v != "" {
		return v
	}
	return uuid.NewString()
}

func write(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": code, "message": message, "data": data, "requestId": requestID(c)})
}

func ormAPI(c *gin.Context) (*apis.Api, bool) {
	a := new(apis.Api).MakeContext(c).MakeOrm()
	if a.Errors != nil {
		write(c, codeServerError, "服务端异常", nil)
		return a, false
	}
	return a, true
}

func maskAccount(account string) string {
	account = strings.TrimSpace(account)
	if phoneRE.MatchString(account) {
		return account[:3] + "****" + account[7:]
	}
	if at := strings.IndexByte(account, '@'); at > 0 {
		name := account[:at]
		domain := account[at:]
		if len(name) <= 3 {
			return name[:1] + "***" + domain
		}
		return name[:3] + "***" + domain
	}
	n := len(account)
	if n >= 6 {
		return account[:3] + "***" + account[n-3:]
	}
	if n > 2 {
		return account[:1] + "***" + account[n-1:]
	}
	return "***"
}

func ensureAgreementWithCode(c *gin.Context, db *gorm.DB, version string, mismatchCode int, mismatchMsg string) bool {
	latest, err := latestAgreement(db)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return true
		}
		write(c, codeServerError, "服务端异常", nil)
		return false
	}
	if strings.TrimSpace(version) == "" || version != latest.Version {
		write(c, mismatchCode, mismatchMsg, gin.H{
			"version":    latest.Version,
			"title":      latest.Title,
			"contentUrl": latest.ContentURL,
			"summary":    latest.Summary,
		})
		return false
	}
	return true
}

func maskPhone(phone string) string {
	if phoneRE.MatchString(phone) {
		return phone[:3] + "****" + phone[7:]
	}
	return maskAccount(phone)
}

func latestAgreement(db *gorm.DB) (*models.Agreement, error) {
	var row models.Agreement
	err := db.Where("status = ?", "published").Order("published_at DESC, id DESC").First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func ensureAgreement(c *gin.Context, db *gorm.DB, version string) bool {
	return ensureAgreementWithCode(c, db, version, codeAgreementNeeded, "需要重新确认用户协议")
}

// createSession 签发 access/refresh，可选写入设备账号记忆；失败时已写响应并返回 ok=false。
func createSession(c *gin.Context, db *gorm.DB, u *models.User, deviceID string, remember bool) (access, refresh string, ok bool) {
	now := time.Now().UTC()
	if u.Status != "active" || isLoginBanned(db, u.ID) {
		write(c, codeAccountBanned, "账号被封禁", nil)
		return "", "", false
	}
	access, _, err := issueAccess(u.ID, u.TokenVersion)
	if err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return "", "", false
	}
	raw := randomToken()
	if raw == "" {
		write(c, codeServerError, "服务端异常", nil)
		return "", "", false
	}
	if err := db.Create(&models.RefreshToken{
		UserID:    u.ID,
		TokenHash: hash(raw),
		ExpiresAt: now.Add(refreshTTL),
	}).Error; err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return "", "", false
	}
	_ = db.Model(u).Updates(map[string]interface{}{"last_login_at": now}).Error

	masked := u.PhoneCiphertext
	if masked == "" && u.Account != "" {
		masked = maskAccount(u.Account)
	}
	if deviceID != "" && remember {
		var da models.DeviceAccount
		q := db.Where("device_id = ? AND user_id = ?", deviceID, u.ID)
		if q.First(&da).Error != nil {
			_ = db.Create(&models.DeviceAccount{
				DeviceID:      deviceID,
				UserID:        u.ID,
				AccountMasked: masked,
				LastLoginAt:   now,
			}).Error
		} else {
			_ = db.Model(&da).Updates(map[string]interface{}{
				"account_masked": masked,
				"last_login_at":  now,
			}).Error
		}
	}
	return access, raw, true
}

func isLoginBanned(db *gorm.DB, userID uint) bool {
	var n int64
	now := time.Now().UTC()
	db.Model(&models.UserBan{}).
		Where("user_id = ? AND status = ? AND ban_type IN ?", userID, "active", []string{"login", "all"}).
		Where("starts_at <= ?", now).
		Where("expires_at IS NULL OR expires_at > ?", now).
		Count(&n)
	return n > 0
}

func issueSession(c *gin.Context, db *gorm.DB, u *models.User, deviceID string, remember bool) {
	access, refresh, ok := createSession(c, db, u, deviceID, remember)
	if !ok {
		return
	}
	masked := u.PhoneCiphertext
	if masked == "" && u.Account != "" {
		masked = maskAccount(u.Account)
	}
	avatar := interface{}(u.AvatarURL)
	if u.AvatarURL == "" {
		avatar = nil
	}
	write(c, codeOK, "ok", gin.H{
		"accessToken":  access,
		"refreshToken": refresh,
		"expiresIn":    int(accessTTL.Seconds()),
		"user": gin.H{
			"id":            fmt.Sprintf("user_%d", u.ID),
			"nickname":      u.Nickname,
			"avatarUrl":     avatar,
			"accountMasked": masked,
			"vipLevel":      u.VipLevel,
		},
		"redirect": gin.H{"type": "home"},
	})
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(hashed, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}
