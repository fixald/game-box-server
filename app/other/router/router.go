package router

import (
	bannerapis "go-admin/app/banners/apis"
	cauth "go-admin/app/cauth/apis"
	contentapis "go-admin/app/content/apis"
	gameapis "go-admin/app/games/apis"
	liveapis "go-admin/app/live/apis"
	searchapis "go-admin/app/search/apis"
	serverapis "go-admin/app/servers/apis"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func registerAuth(r *gin.RouterGroup) {
	r.POST("/auth/login", cauth.Login)
	r.POST("/auth/register", cauth.Register)
	r.POST("/auth/register/check-account", cauth.CheckAccount)
	r.POST("/auth/register/verify", cauth.RegisterVerify)
	r.POST("/auth/sms/send", cauth.SendSMS)
	r.POST("/auth/sms/login", cauth.SMSLogin)
	r.POST("/auth/refresh", cauth.Refresh)
	r.POST("/auth/logout", cauth.Logout)
	r.POST("/auth/password/reset/request", cauth.PasswordResetRequest)
	r.POST("/auth/password/reset/confirm", cauth.PasswordResetConfirm)
	r.GET("/auth/agreements/latest", cauth.AgreementsLatest)
	r.GET("/auth/accounts", cauth.Accounts)
}

var (
	routerNoCheckRole = []func(*gin.RouterGroup){func(v1 *gin.RouterGroup) {
		v1.GET("/banners", bannerapis.Active)
		v1.GET("/games", func(c *gin.Context) { c.Set("publishedOnly", true); gameapis.List(c) })
		v1.GET("/games/:id", func(c *gin.Context) { c.Set("publishedOnly", true); gameapis.Detail(c) })
	}}
	routerCheckRole = []func(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware){registerAccountRouter}
)

func registerAccountRouter(r *gin.RouterGroup, auth *jwt.GinJWTMiddleware) {
	{
		g := r.Group("").Use(auth.MiddlewareFunc())
		g.GET("/users/me", cauth.Account)
		g.POST("/reports", contentapis.CreateReport)
		g.GET("/live/streamers/:id/rooms", liveapis.StreamerRooms)
		g.GET("/live/streamers/:id", liveapis.StreamerDetail)
		g.POST("/live/streamers/:id/follow", liveapis.FollowStreamer)
		g.DELETE("/live/streamers/:id/follow", liveapis.FollowStreamer)
		g.GET("/users/me/live/following", liveapis.MyFollowing)
		g.GET("/live/streamers", liveapis.Streamers)
		g.PATCH("/users/me", cauth.UpdateAccount)
		g.GET("/users/me/stats", cauth.AccountStats)
		g.GET("/tasksinfo", cauth.Tasks)
		g.GET("/tasks", cauth.TaskList)
		g.GET("/checkin-rewards", cauth.CheckinRewards)
		g.POST("/checkin", cauth.Checkin)
		g.POST("/checkin-rewards/claim", cauth.ClaimCheckinReward)
		g.POST("/tasks/claim", cauth.ClaimTask)
		g.GET("/vip/levels", cauth.VIPLevels)
		g.GET("/users/me/games/:kind", cauth.AccountList)
		g.GET("/users/me/servers/recent", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "kind", Value: "recent-server"})
			cauth.AccountList(c)
		})
		g.POST("/users/me/games/:gameId/favorite", cauth.Favorite)
		g.DELETE("/users/me/games/:gameId/favorite", cauth.Favorite)
		g.GET("/users/me/downloads", cauth.AccountList)
		g.GET("/messages", cauth.Messages)
		g.POST("/messages/:id/read", cauth.MessageRead)
		g.POST("/messages/read-all", cauth.MessageReadAll)
		g.GET("/users/me/devices", cauth.Devices)
		g.DELETE("/users/me/devices/:deviceId", cauth.DeviceDelete)
		g.GET("/users/me/login-records", cauth.LoginRecords)
		g.POST("/users/me/password/change", cauth.ChangePassword)
		g.GET("/users/me/security", cauth.Security)
		g.GET("/users/me/settings", cauth.Settings)
		g.PATCH("/users/me/settings", cauth.UpdateSettings)
		g.GET("/users/me/rewards", cauth.Rewards)
		g.POST("/users/me/phone/bind/send", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "kind", Value: "phone"})
			cauth.BindSend(c)
		})
		g.POST("/users/me/phone/bind/confirm", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "kind", Value: "phone"})
			cauth.BindConfirm(c)
		})
		g.POST("/users/me/email/bind/send", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "kind", Value: "email"})
			cauth.BindSend(c)
		})
		g.POST("/users/me/email/bind/confirm", func(c *gin.Context) {
			c.Params = append(c.Params, gin.Param{Key: "kind", Value: "email"})
			cauth.BindConfirm(c)
		})
	}
}

// initRouter 路由示例
func initRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) *gin.Engine {
	noCheckRoleRouter(r)
	checkRoleRouter(r, authMiddleware)
	return r
}

func noCheckRoleRouter(r *gin.Engine) {
	g := r.Group("/api/v1")
	for _, f := range routerNoCheckRole {
		f(g)
	}
	client := r.Group("/api/v1/client")
	registerAuth(client)
	client.GET("/banners", bannerapis.Active)
	client.GET("/games", func(c *gin.Context) { c.Set("publishedOnly", true); gameapis.List(c) })
	client.GET("/games/:id", func(c *gin.Context) { c.Set("publishedOnly", true); gameapis.Detail(c) })
	client.GET("/game-servers", serverapis.RecommendedList)
	client.GET("/live/categories", liveapis.Categories)
	client.GET("/live/rooms/:id", liveapis.Detail)
	client.GET("/search", searchapis.Search)
	client.GET("/search/suggestions", searchapis.Suggestions)
	client.GET("/search/hot", searchapis.Hot)
}

func checkRoleRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	v1 := r.Group("/api/v1")
	for _, f := range routerCheckRole {
		f(v1, authMiddleware)
	}
	client := r.Group("/api/v1/client")
	for _, f := range routerCheckRole {
		f(client, authMiddleware)
	}
	authenticated := client.Group("/search").Use(authMiddleware.MiddlewareFunc())
	authenticated.GET("/history", searchapis.History)
	authenticated.POST("/history", searchapis.AddHistory)
	authenticated.DELETE("/history", searchapis.ClearHistory)
	authenticated.POST("/events", searchapis.Event)
	client.POST("/events", liveapis.CreateEvent)
}
