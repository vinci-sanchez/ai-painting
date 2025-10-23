package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

// 配置
var (
	APIKey  = ""
	BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
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

	// 新API：文本分段和关键词提取
	r.POST("/api/segment_keywords", func(c *gin.Context) {
		var requestBody struct {
			Text string `json:"text"`
		}
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法解析请求体"})
			return
		}

		// 自动创建 python 临时文件夹
		os.MkdirAll("python", os.ModePerm)

		// 写入临时文件
		inputFile := "python/temp_input.txt"
		if err := os.WriteFile(inputFile, []byte(requestBody.Text), 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法写入临时文件"})
			return
		}

		// 打印当前路径（可删）
		cwd, _ := os.Getwd()
		fmt.Println("当前工作目录：", cwd)

		// 执行 Python 脚本
		cmd := exec.Command("D:\\github\\ai_painting\\.venv\\Scripts\\python.exe", "../python/text_segment_keyword_extract.py", inputFile)

		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Python脚本执行失败: " + stderr.String()})
			return
		}

		// 读取输出
		result, err := os.ReadFile("../python/output.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取输出文件"})
			return
		}

		var jsonResult interface{}
		if err := json.Unmarshal(result, &jsonResult); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON解析失败"})
			return
		}

		c.JSON(http.StatusOK, jsonResult)
	})

	// 启动服务
	r.Run(":3000")
}
