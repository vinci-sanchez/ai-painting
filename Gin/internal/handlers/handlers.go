package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"

	"comic-proxy/internal/config"
	"comic-proxy/internal/crawler"
	"comic-proxy/internal/storage"
)

// Handler 聚合所有 HTTP 处理逻辑
type Handler struct {
	cfg       *config.Config
	crawler   *crawler.Service
	userStore *storage.UserStore
	apiKey    string
	baseURL   string
	colabURL  string
}

// New 创建 Handler 实例
func New(cfg *config.Config, crawlerService *crawler.Service, userStore *storage.UserStore) *Handler {
	return &Handler{
		cfg:       cfg,
		crawler:   crawlerService,
		userStore: userStore,
		apiKey:    cfg.APIKey,
		baseURL:   cfg.BaseURL,
		colabURL:  cfg.ColabEndpoint,
	}
}

// RegisterRoutes 将所有路由挂载到 Gin 引擎
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	h.registerTextRoutes(r)
	h.registerColabRoutes(r)
	h.registerArkRoutes(r)
	h.registerCrawlerRoutes(r)
	h.registerHelloRoutes(r)
	h.registerUserRoutes(r)
}

// resolveAPIKey 优先使用请求体中的 apiKey，若为空则回退到服务端配置
func (h *Handler) resolveAPIKey(candidate string) string {
	if trimmed := strings.TrimSpace(candidate); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(h.apiKey)
}
