package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	cauth "go-admin/app/cauth/apis"
	"go-admin/common/middleware"
)

func registerPlayerUserRouter(r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	g := r.Group("/api/v1/users").Use(auth.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	g.GET("", cauth.UserList)
	g.GET("/:id", cauth.UserGet)
	g.GET("/:id/bans", cauth.BanList)
	g.POST("/:id/ban", cauth.Ban)
	b := r.Group("/api/v1/user-bans").Use(auth.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	b.POST("/:id/unban", cauth.Unban)
}
