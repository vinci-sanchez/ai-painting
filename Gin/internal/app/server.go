package app

import (
	"log"

	"github.com/gin-gonic/gin"

	"comic-proxy/internal/config"
	"comic-proxy/internal/crawler"
	"comic-proxy/internal/handlers"
	"comic-proxy/internal/middleware"
	"comic-proxy/internal/storage"
)

// NewRouter 创建并配置 Gin 引擎。
func NewRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	userStore, err := storage.NewUserStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("初始化用户存储失败: %v", err)
	}

	handler := handlers.New(cfg, crawler.NewService(), userStore)
	handler.RegisterRoutes(r)

	return r
}
