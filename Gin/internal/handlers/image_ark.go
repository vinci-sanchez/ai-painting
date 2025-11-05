package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"

	"comic-proxy/internal/models"
)

func (h *Handler) registerArkRoutes(r *gin.Engine) {
	r.POST("/api/image", h.handleArkImage)
}

func (h *Handler) handleArkImage(c *gin.Context) {
	log.Printf("开始处理 /api/image 请求, 时间: %s", time.Now().Format(time.RFC3339))

	var body bytes.Buffer
	if _, err := io.Copy(&body, c.Request.Body); err != nil {
		log.Printf("无法读取请求体: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
		return
	}
	log.Printf("接收到请求体: %s", body.String())

	var requestData struct {
		Role       string `json:"role"`
		Prompt     string `json:"prompt"`
		Storyboard string `json:"storyboard"`
	}
	if err := json.Unmarshal(body.Bytes(), &requestData); err != nil {
		log.Printf("解析请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("解析请求体失败: %v", err)})
		return
	}

	arkData := struct {
		Model      string `json:"model"`
		Prompt     string `json:"prompt"`
		Role       string `json:"role,omitempty"`
		Storyboard string `json:"storyboard,omitempty"`
	}{
		Model:      "ep-20251026170710-8pgpm",
		Prompt:     requestData.Prompt,
		Role:       requestData.Role,
		Storyboard: requestData.Storyboard,
	}
	arkDataBytes, err := json.Marshal(arkData)
	if err != nil {
		log.Printf("序列化火山方舟数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("序列化火山方舟数据失败: %v", err)})
		return
	}
	log.Printf("发送到火山方舟的数据: %s", string(arkDataBytes))

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+h.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(arkDataBytes).
		Post(h.baseURL + "/images/generations")

	if err != nil {
		log.Printf("调用火山方舟 API 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("调用火山方舟 API 失败: %v", err)})
		return
	}

	log.Printf("火山方舟 API 状态码: %d", resp.StatusCode())
	log.Printf("火山方舟 API 响应体: %s", string(resp.Body()))
	if resp.StatusCode() != http.StatusOK {
		log.Printf("火山方舟 API 返回非 200 状态码: %d", resp.StatusCode())
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("火山方舟 API 返回非 200 状态码: %d", resp.StatusCode())})
		return
	}

	var arkResponse models.ImageDataResponse1
	if err := json.Unmarshal(resp.Body(), &arkResponse); err != nil {
		log.Printf("解析火山方舟响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析火山方舟响应失败: %v", err)})
		return
	}

	if len(arkResponse.Data) == 0 {
		log.Printf("火山方舟 API 响应无效: 无数据")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "火山方舟 API 响应无效: 无数据"})
		return
	}

	log.Printf("成功从火山方舟接收到图像数据, URL: %s", arkResponse.Data[0].URL)
	c.JSON(http.StatusOK, arkResponse)
}
