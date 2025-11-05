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

	"comic-proxy/internal/models"
)

func (h *Handler) registerColabRoutes(r *gin.Engine) {
	r.POST("/generate_image", h.handleColabImage)
}

func (h *Handler) handleColabImage(c *gin.Context) {
	log.Printf("开始处理 /generate_image 请求, 时间: %s", time.Now().Format(time.RFC3339))

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

	colabData := struct {
		Role       string `json:"role"`
		Prompt     string `json:"prompt"`
		Storyboard string `json:"storyboard"`
	}{
		Role:       requestData.Role,
		Prompt:     requestData.Prompt,
		Storyboard: requestData.Storyboard,
	}
	colabDataBytes, err := json.Marshal(colabData)
	if err != nil {
		log.Printf("序列化 Colab 数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("序列化 Colab 数据失败: %v", err)})
		return
	}
	log.Printf("发送到 Colab 的数据: %s", string(colabDataBytes))

	resp, err := http.Post(h.colabURL+"/generate", "application/json", bytes.NewBuffer(colabDataBytes))
	if err != nil {
		log.Printf("转发数据到 Colab 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("转发数据到 Colab 失败: %v", err)})
		return
	}
	defer resp.Body.Close()

	log.Printf("Colab API 状态码: %d", resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取 Colab 响应体失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取 Colab 响应体失败: %v", err)})
		return
	}
	log.Printf("Colab API 响应体: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		log.Printf("从 Colab 收到非 OK 状态码: %d", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("从 Colab 收到非 OK 状态码: %d", resp.StatusCode)})
		return
	}

	var colabResponse models.ImageDataResponse
	if err := json.Unmarshal(bodyBytes, &colabResponse); err != nil {
		log.Printf("解析 Colab 响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析 Colab 响应失败: %v", err)})
		return
	}

	log.Printf("成功从 Colab 接收到图片数据, Base64 长度: %d", len(colabResponse.ImageDataBase64))
	c.JSON(http.StatusOK, colabResponse)
}
