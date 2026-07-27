package router

import (
	liveapi "go-admin/app/live/apis"
	"go-admin/common/middleware"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func RegisterLiveRouter(r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	g := r.Group("/api/v1/client/live")

	g.GET("/rooms", liveapi.List)

	g.Use(auth.MiddlewareFunc(), middleware.AuthCheckRole())
	g.POST("/room", liveapi.Create)
	g.GET("/room", liveapi.GetMyRoom)
	g.DELETE("/room", liveapi.EndRoom)
}
