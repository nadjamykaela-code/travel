package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nadjamykaela-code/travel/apps/api/handlers"
	"github.com/nadjamykaela-code/travel/apps/api/middleware"
	"github.com/nadjamykaela-code/travel/apps/api/service"
)

type Dependencies struct {
	FilterService *service.FilterService
	AuthService   service.AuthService
}

func RegisterRoutes(router *gin.Engine, deps *Dependencies) {
	filterHandler := handlers.NewFilterHandler(deps.FilterService)
	authHandler := handlers.NewAuthHandler(deps.AuthService)
	authMW := middleware.NewAuthMiddleware(deps.AuthService)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/metrics", middleware.PrometheusHandler())

	api := router.Group("/api")
	api.Use(middleware.RateLimitMiddleware())
	api.Use(authMW.RequireAuth())
	{
		api.GET("/filters", filterHandler.GetFilters)
		api.POST("/filters", filterHandler.CreateFilter)
		api.PUT("/filters/:id", filterHandler.UpdateFilter)
		api.DELETE("/filters/:id", filterHandler.DeleteFilter)
		api.GET("/auth/verify", authHandler.VerifyToken)
	}

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Travel Bot API - Online")
	})
}
