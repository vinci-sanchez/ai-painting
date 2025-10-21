package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/go-resty/resty/v2"
)

// 配置
var (
	APIKey   = ""
	BaseURL  = "https://ark.cn-beijing.volces.com/api/v3"
)

func main() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		fmt.Println("警告：未找到 .env 文件，使用默认配置")
	}
	APIKey = os.Getenv("API_KEY")
	if APIKey == "" {
		panic("API_KEY 未设置，请在 .env 文件中配置")
	}

	// 初始化 Gin
	r := gin.Default()

	// 设置 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 代理：文本生成
	r.POST("/api/text", func(c *gin.Context) {
		client := resty.New()
		var body bytes.Buffer
		if _, err := io.Copy(&body, c.Request.Body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
			return
		}

		resp, err := client.R().
			SetHeader("Authorization", "Bearer "+APIKey).
			SetHeader("Content-Type", "application/json").
			SetBody(body.Bytes()).
			Post(BaseURL + "/chat/completions")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(resp.StatusCode(), "application/json", resp.Body())
	})

	// 代理：图像生成
	r.POST("/api/image", func(c *gin.Context) {
		client := resty.New()
		var body bytes.Buffer
		if _, err := io.Copy(&body, c.Request.Body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
			return
		}

		resp, err := client.R().
			SetHeader("Authorization", "Bearer "+APIKey).
			SetHeader("Content-Type", "application/json").
			SetBody(body.Bytes()).
			Post(BaseURL + "/images/generations")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(resp.StatusCode(), "application/json", resp.Body())
	})

	// 启动服务
	r.Run(":3000") // 监听 http://localhost:3000
}