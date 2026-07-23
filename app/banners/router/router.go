package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	"go-admin/app/banners/apis"
	"go-admin/common/middleware"
)

func Register(v1 *gin.RouterGroup, auth *jwtauth.GinJWTMiddleware) {
	r := v1.Group("/banners").Use(auth.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	r.GET("", apis.List)
	r.GET("/:id", apis.Get)
	r.POST("", apis.Create)
	r.PUT("/:id", apis.Update)
	r.DELETE("/:id", apis.Delete)
	r.POST("/:id/publish", apis.Publish)
	r.POST("/:id/recall", apis.Recall)
}
