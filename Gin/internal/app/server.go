package app

import (
	"github.com/gin-gonic/gin"

	"comic-proxy/internal/config"
	"comic-proxy/internal/crawler"
	"comic-proxy/internal/handlers"
	"comic-proxy/internal/middleware"
)

// NewRouter 创建并配置 Gin 引擎。
func NewRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	handler := handlers.New(cfg, crawler.NewService())
	handler.RegisterRoutes(r)

	return r
}
