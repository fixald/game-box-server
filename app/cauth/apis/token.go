package apis

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"github.com/golang-jwt/jwt/v5"
	"go-admin/app/cauth/models"
	"gorm.io/gorm"
)

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func issueAccess(userID uint, tokenVersion int64) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(accessTTL)
	claims := jwt.MapClaims{
		"sub":          fmt.Sprint(userID),
		"userType":     "player",
		"tokenType":    "access",
		"tokenVersion": tokenVersion,
		"exp":          exp.Unix(),
		"iat":          now.Unix(),
	}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.JwtConfig.Secret))
	return t, exp, err
}

func Refresh(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in refreshRequest
	if c.ShouldBindJSON(&in) != nil {
		write(c, codeTokenExpired, "Token 过期，需要刷新", nil)
		return
	}
	var old models.RefreshToken
	if err := a.Orm.Where("token_hash = ?", hash(in.RefreshToken)).First(&old).Error; err != nil {
		write(c, codeTokenExpired, "Token 过期，需要刷新", nil)
		return
	}
	now := time.Now().UTC()
	if old.RevokedAt != nil || now.After(old.ExpiresAt) {
		write(c, codeTokenExpired, "Token 过期，需要刷新", nil)
		return
	}
	var u models.User
	if err := a.Orm.First(&u, old.UserID).Error; err != nil || u.Status != "active" || isLoginBanned(a.Orm, u.ID) {
		write(c, codeAccountBanned, "账号被封禁", nil)
		return
	}
	newRaw := randomToken()
	if newRaw == "" {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	next := models.RefreshToken{UserID: u.ID, TokenHash: hash(newRaw), ExpiresAt: now.Add(refreshTTL)}
	access, _, err := issueAccess(u.ID, u.TokenVersion)
	if err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	err = a.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&next).Error; err != nil {
			return err
		}
		return tx.Model(&old).Updates(map[string]interface{}{"revoked_at": now, "replaced_by": next.TokenHash}).Error
	})
	if err != nil {
		write(c, codeServerError, "服务端异常", nil)
		return
	}
	write(c, codeOK, "ok", gin.H{
		"accessToken":  access,
		"refreshToken": newRaw,
		"expiresIn":    int(accessTTL.Seconds()),
	})
}

func Logout(c *gin.Context) {
	a, ok := ormAPI(c)
	if !ok {
		return
	}
	var in refreshRequest
	if c.ShouldBindJSON(&in) == nil && in.RefreshToken != "" {
		now := time.Now().UTC()
		a.Orm.Model(&models.RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL", hash(in.RefreshToken)).Update("revoked_at", now)
	}
	if raw := c.GetHeader("Authorization"); len(raw) > 7 && raw[:7] == "Bearer " {
		if parsed, err := jwt.Parse(raw[7:], func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JwtConfig.Secret), nil
		}); err == nil && parsed.Valid {
			if claims, ok := parsed.Claims.(jwt.MapClaims); ok {
				if sub, ok := claims["sub"].(string); ok {
					a.Orm.Model(&models.User{}).Where("id = ?", sub).UpdateColumn("token_version", gorm.Expr("token_version + 1"))
				}
			}
		}
	}
	write(c, codeOK, "ok", nil)
}
