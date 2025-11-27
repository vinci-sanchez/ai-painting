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
	"github.com/go-resty/resty/v2"
)

func (h *Handler) registerTextRoutes(r *gin.Engine) {
	r.POST("/api/text", h.handleTextGeneration)
}

func (h *Handler) handleTextGeneration(c *gin.Context) {
	log.Printf("开始处理 /api/text 请求, 时间: %s", time.Now().Format(time.RFC3339))

	var body bytes.Buffer
	if _, err := io.Copy(&body, c.Request.Body); err != nil {
		log.Printf("无法读取请求体: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
		return
	}
	log.Printf("接收到请求体: %s", body.String())

	var payload map[string]interface{}
	if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
		log.Printf("解析请求体失�? %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("解析请求体失�? %v", err)})
		return
	}

	var providedKey string
	var providedBaseURL string
	if rawKey, ok := payload["apiKey"].(string); ok {
		providedKey = rawKey
		delete(payload, "apiKey")
	}
	if rawBaseURL, ok := payload["baseUrl"].(string); ok {
		providedBaseURL = rawBaseURL
		delete(payload, "baseUrl")
	}

	cleanBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("重编码请求体失�? %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("重编码请求体失�? %v", err)})
		return
	}

	apiKey := h.resolveAPIKey(providedKey)
	if apiKey == "" {
		log.Printf("缺少可用的 API_KEY")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "缺少可用的 API_KEY"})
		return
	}

	targetBaseURL := h.baseURL
	if trimmed := strings.TrimSpace(providedBaseURL); trimmed != "" {
		targetBaseURL = trimmed
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(cleanBody).
		Post(targetBaseURL + "/chat/completions")

	if err != nil {
		log.Printf("调用外部 API 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("调用外部 API 失败: %v", err)})
		return
	}

	log.Printf("外部 API 状态码: %d", resp.StatusCode())
	log.Printf("外部 API 响应体: %s", string(resp.Body()))
	if resp.StatusCode() != http.StatusOK {
		log.Printf("外部 API 返回非 200 状态码: %d", resp.StatusCode())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("外部 API 返回非 200 状态码: %d", resp.StatusCode())})
		return
	}

	var textResp struct {
		Choices []struct {
			Message struct {
				Content interface{} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(resp.Body(), &textResp); err != nil {
		log.Printf("解析外部 API 响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析外部 API 响应失败: %v", err)})
		return
	}

	if len(textResp.Choices) == 0 {
		log.Printf("外部 API 响应无效: 缺少 choices 数据")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "外部 API 响应无效: 缺少 choices 数据"})
		return
	}

	response := struct {
		Message interface{} `json:"message"`
	}{
		Message: textResp.Choices[0].Message.Content,
	}

	log.Printf("返回响应: %v", response)
	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}
