package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"comic-proxy/internal/crawler"
	"comic-proxy/internal/models"
)

func (h *Handler) registerCrawlerRoutes(r *gin.Engine) {
	r.POST("/api/crawl", h.handleCrawl)
}

func (h *Handler) handleCrawl(c *gin.Context) {
	log.Printf("开始处理 /api/crawl 请求, 时间: %s", time.Now().Format(time.RFC3339))

	var req models.CrawlRequest
	var body bytes.Buffer
	if _, err := io.Copy(&body, c.Request.Body); err != nil {
		log.Printf("无法读取请求体: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
		return
	}
	if err := json.Unmarshal(body.Bytes(), &req); err != nil {
		log.Printf("解析请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("解析请求体失败: %v", err)})
		return
	}

	directoryURL := h.crawler.InferDirectoryURL(req.NovelURL)
	log.Printf("推断目录页URL: %s (原URL: %s)", directoryURL, req.NovelURL)

	if !strings.HasPrefix(directoryURL, "http") {
		log.Printf("无效的URL: %s", directoryURL)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入有效的URL（目录页或章节页）"})
		return
	}

	// if req.StartIndex != 0 {
	// 	log.Printf("强制从第一章开始，忽略 start_index=%d", req.StartIndex)
	// 	req.StartIndex = 0
	// }
	// req.Limit = 1
	log.Printf("爬虫参数: 目录=%s, StartIndex=%d, Limit=%d", directoryURL, req.StartIndex, req.Limit)

	chapters, nextIndex, totalChapters, err := h.crawler.FetchChapters(directoryURL, req.StartIndex, req.Limit)
	if err != nil {
		log.Printf("爬取章节内容失败: %v", err)
		if crawlerErr, ok := err.(*crawler.Error); ok {
			c.JSON(crawlerErr.Status, gin.H{"error": crawlerErr.Message})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("爬取章节内容失败: %v", err)})
		}
		return
	}

	resp := models.CrawlResponse{
		Chapters:      chapters,
		NextIndex:     nextIndex,
		TotalChapters: totalChapters,
	}
	log.Printf("爬取成功: 返回 %d 个章节, 下次索引 %d, 总章节 %d", len(chapters), nextIndex, totalChapters)
	c.JSON(http.StatusOK, resp)
}
