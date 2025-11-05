package handlers

import (
	"github.com/gin-gonic/gin"

	"comic-proxy/internal/config"
	"comic-proxy/internal/crawler"
)

// Handler 聚合所有 HTTP 处理逻辑。
type Handler struct {
	cfg      *config.Config
	crawler  *crawler.Service
	apiKey   string
	baseURL  string
	colabURL string
}

// New 创建 Handler 实例。
func New(cfg *config.Config, crawlerService *crawler.Service) *Handler {
	return &Handler{
		cfg:      cfg,
		crawler:  crawlerService,
		apiKey:   cfg.APIKey,
		baseURL:  cfg.BaseURL,
		colabURL: cfg.ColabEndpoint,
	}
}

// RegisterRoutes 将所有路由挂载到 Gin 引擎。
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	h.registerTextRoutes(r)
	h.registerColabRoutes(r)
	h.registerArkRoutes(r)
	h.registerCrawlerRoutes(r)
}
