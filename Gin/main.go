package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

// 配置日志输出到文件和终端
func init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件: ", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.LstdFlags | log.Lshortfile) // 添加文件名和行号
}

// 配置
var (
	APIKey        = ""
	BaseURL       = "https://ark.cn-beijing.volces.com/api/v3"
	colabEndpoint = "https://articular-proportionable-alverta.ngrok-free.dev/"
)

// ImageDataRequest 定义请求结构
type TextGenerationResponse struct {
	Scene     string `json:"scene"`
	Prompt    string `json:"prompt"`
	Character string `json:"character,omitempty"` // 人物可选
}

// ImageDataResponse 定义响应结构
type ImageDataResponse struct {
	ImageDataBase64 string `json:"image_base64"`
}

// ImageDataResponse1 定义火山方舟 API 的响应结构
type ImageDataResponse1 struct {
    Model   string `json:"model"`
    Created int64  `json:"created"`
    Data    []struct {
        URL  string `json:"url"`
        Size string `json:"size"`
    } `json:"data"`
    Usage struct {
        GeneratedImages int `json:"generated_images"`
        OutputTokens    int `json:"output_tokens"`
        TotalTokens     int `json:"total_tokens"`
    } `json:"usage"`
}
// CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func main() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Printf("警告：未找到 .env 文件，使用默认配置: %v", err)
	}
	APIKey = os.Getenv("API_KEY")
	if APIKey == "" {
		log.Fatal("API_KEY 未设置，请在 .env 文件中配置")
	}
	log.Printf("APIKey 已加载: %s", APIKey)
	log.Printf("BaseURL: %s, ColabEndpoint: %s", BaseURL, colabEndpoint)

	// 初始化 Gin
	r := gin.Default()
	r.Use(CORSMiddleware())

	// 代理：文本生成
	// 代理：文本生成
// 代理：文本生成
r.POST("/api/text", func(c *gin.Context) {
    log.Printf("开始处理 /api/text 请求, 时间: %s", time.Now().Format(time.RFC3339))

    // 读取请求体
    var body bytes.Buffer
    if _, err := io.Copy(&body, c.Request.Body); err != nil {
        log.Printf("无法读取请求体: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
        return
    }
    log.Printf("接收到请求体: %s", body.String())

    // 调用外部 API
    client := resty.New()
    resp, err := client.R().
        SetHeader("Authorization", "Bearer "+APIKey).
        SetHeader("Content-Type", "application/json").
        SetBody(body.Bytes()).
        Post(BaseURL + "/chat/completions")

    if err != nil {
        log.Printf("调用外部 API 失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("调用外部 API 失败: %v", err)})
        return
    }

    // 检查状态码
    log.Printf("外部 API 状态码: %d", resp.StatusCode())
    log.Printf("外部 API 响应体: %s", string(resp.Body()))
    if resp.StatusCode() != http.StatusOK {
        log.Printf("外部 API 返回非 200 状态码: %d", resp.StatusCode())
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("外部 API 返回非 200 状态码: %d", resp.StatusCode())})
        return
    }

    // 解析外部 API 响应
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

    // 检查是否收到有效数据
    if len(textResp.Choices) == 0 {
        log.Printf("外部 API 响应无效: 无 choices 数据")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "外部 API 响应无效: 无 choices 数据"})
        return
    }

    // 返回 message 格式
    response := struct {
        Message interface{} `json:"message"`
    }{
        Message: textResp.Choices[0].Message.Content,
    }

    log.Printf("返回响应: %v", response)
    c.JSON(http.StatusOK, gin.H{
        "data": response,
    })
})


	// 路由：生成图像
// 路由：生成图像
r.POST("/generate_image", func(c *gin.Context) {
    log.Printf("开始处理 /generate_image 请求, 时间: %s", time.Now().Format(time.RFC3339))

    // 读取请求体
    var body bytes.Buffer
    if _, err := io.Copy(&body, c.Request.Body); err != nil {
        log.Printf("无法读取请求体: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
        return
    }
    log.Printf("接收到请求体: %s", body.String())

    // 解析请求体
    var requestData struct {
        Role      string `json:"role"`
        Prompt    string `json:"prompt"`
        Storyboard string `json:"storyboard"`
    }
    if err := json.Unmarshal(body.Bytes(), &requestData); err != nil {
        log.Printf("解析请求体失败: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("解析请求体失败: %v", err)})
        return
    }

    // 构造 Colab 请求数据
    colabData := struct {
        Role      string `json:"role"`
        Prompt    string `json:"prompt"`
        Storyboard string `json:"storyboard"`
    }{
        Role:      requestData.Role,
        Prompt:    requestData.Prompt,
        Storyboard: requestData.Storyboard,
    }
    colabDataBytes, err := json.Marshal(colabData)
    if err != nil {
        log.Printf("序列化 Colab 数据失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("序列化 Colab 数据失败: %v", err)})
        return
    }
    log.Printf("发送到 Colab 的数据: %s", string(colabDataBytes))

    resp, err := http.Post(colabEndpoint+"/generate", "application/json", bytes.NewBuffer(colabDataBytes))
    if err != nil {
        log.Printf("转发数据到 Colab 失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("转发数据到 Colab 失败: %v", err)})
        return
    }
    defer resp.Body.Close()

    // 检查 Colab 响应
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

    var colabResponse ImageDataResponse
    if err := json.Unmarshal(bodyBytes, &colabResponse); err != nil {
        log.Printf("解析 Colab 响应失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析 Colab 响应失败: %v", err)})
        return
    }

    log.Printf("成功从 Colab 接收到图片数据, Base64 长度: %d", len(colabResponse.ImageDataBase64))
    c.JSON(http.StatusOK, colabResponse)
})

// 代理：图像生成
// 路由：生成图像
// 路由：生成图像
r.POST("/api/image", func(c *gin.Context) {
    log.Printf("开始处理 /api/image 请求, 时间: %s", time.Now().Format(time.RFC3339))

    // 读取请求体
    var body bytes.Buffer
    if _, err := io.Copy(&body, c.Request.Body); err != nil {
        log.Printf("无法读取请求体: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
        return
    }
    log.Printf("接收到请求体: %s", body.String())

    // 解析请求体
    var requestData struct {
        Role      string `json:"role"`
        Prompt    string `json:"prompt"`
        Storyboard string `json:"storyboard"`
    }
    if err := json.Unmarshal(body.Bytes(), &requestData); err != nil {
        log.Printf("解析请求体失败: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("解析请求体失败: %v", err)})
        return
    }

    // 构造火山方舟请求数据
    arkData := struct {
        Model     string `json:"model"`
        Prompt    string `json:"prompt"`
        Role      string `json:"role,omitempty"`
        Storyboard string `json:"storyboard,omitempty"`
    }{
        Model:     "ep-20251021153509-xh86n",
        Prompt:    requestData.Prompt,
        Role:      requestData.Role,
        Storyboard: requestData.Storyboard,
    }
    arkDataBytes, err := json.Marshal(arkData)
    if err != nil {
        log.Printf("序列化火山方舟数据失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("序列化火山方舟数据失败: %v", err)})
        return
    }
    log.Printf("发送到火山方舟的数据: %s", string(arkDataBytes))

    // 调用火山方舟 API
    client := resty.New()
    resp, err := client.R().
        SetHeader("Authorization", "Bearer "+APIKey).
        SetHeader("Content-Type", "application/json").
        SetBody(arkDataBytes).
        Post(BaseURL + "/images/generations")

    if err != nil {
        log.Printf("调用火山方舟 API 失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("调用火山方舟 API 失败: %v", err)})
        return
    }

    // 检查状态码
    log.Printf("火山方舟 API 状态码: %d", resp.StatusCode())
    log.Printf("火山方舟 API 响应体: %s", string(resp.Body()))
    if resp.StatusCode() != http.StatusOK {
        log.Printf("火山方舟 API 返回非 200 状态码: %d", resp.StatusCode())
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("火山方舟 API 返回非 200 状态码: %d", resp.StatusCode())})
        return
    }

    // 解析火山方舟响应
    var arkResponse ImageDataResponse1
    if err := json.Unmarshal(resp.Body(), &arkResponse); err != nil {
        log.Printf("解析火山方舟响应失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析火山方舟响应失败: %v", err)})
        return
    }

    // 检查是否收到有效数据
    if len(arkResponse.Data) == 0 {
        log.Printf("火山方舟 API 响应无效: 无数据")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "火山方舟 API 响应无效: 无数据"})
        return
    }

    log.Printf("成功从火山方舟接收到图像数据, URL: %s", arkResponse.Data[0].URL)
    c.JSON(http.StatusOK, arkResponse)
})
	// 启动服务
	log.Printf("启动服务器于 :3000")
	if err := r.Run(":3000"); err != nil {
		log.Fatal("服务器运行失败: ", err)
	}
}
