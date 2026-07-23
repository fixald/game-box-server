package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	contentapi "go-admin/app/content/apis"
	gameapi "go-admin/app/games/apis"
	serverapi "go-admin/app/servers/apis"
	"go-admin/common/middleware"
)

func registerBusinessRouter(r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	g := r.Group("/api/v1")
	g.Use(auth.MiddlewareFunc(), middleware.AuthCheckRole())
	s := g.Group("/servers")
	games := g.Group("/admin/games")
	games.GET("", gameapi.AdminList)
	games.GET("/:id", gameapi.AdminDetail)
	games.POST("", gameapi.Create)
	games.PUT("/:id", gameapi.Update)
	games.PUT("/:id/status", gameapi.UpdateStatus)
	games.DELETE("/:id", gameapi.Delete)
	s.GET("", serverapi.List)
	s.GET("/:id", serverapi.Get)
	s.POST("", serverapi.Create)
	s.PUT("/:id", serverapi.Update)
	s.POST("/import", serverapi.Import)
	s.POST("/batch-maintain", serverapi.BatchMaintain)
	g.GET("/reports", contentapi.Reports)
	g.PUT("/reports/:id", contentapi.ProcessReport)
	g.GET("/feedbacks", contentapi.Feedbacks)
	g.PUT("/feedbacks/:id", contentapi.ProcessFeedback)
}
