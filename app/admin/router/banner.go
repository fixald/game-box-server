package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	bannerrouter "go-admin/app/banners/router"
)

func registerBannerRouter(r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	// 后台 Banner 挂到 /admin，避免与 C 端 GET /api/v1/banners 冲突
	bannerrouter.Register(r.Group("/api/v1/admin"), auth)
}
