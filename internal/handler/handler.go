package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	_ "testAlvtoShp/docs"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) InitRoutes(ctx context.Context) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	log := logger.LoggerFromContext(ctx)

	router := gin.New()
	router.Use(gin.Recovery(), logger.LoggerMiddleware(log))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := router.Group("/api")
	{
		api.POST("/auth", h.GetAuthToken)
		api.Use(h.CheckAuth).GET("/info", h.GetUserInfo)
		api.Use(h.CheckAuth).GET("/buy/:item", h.BuyItem)
		api.Use(h.CheckAuth).POST("/sendCoin", h.SendCoin)

	}
	return router
}
