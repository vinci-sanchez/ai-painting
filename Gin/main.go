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
    APIKey       = ""
    BaseURL      = "https://ark.cn-beijing.volces.com/api/v3"
    colabEndpoint = "https://articular-proportionable-alverta.ngrok-free.dev/"
)

// ImageDataRequest 定义请求结构
type ImageDataRequest struct {
    Prompt     string `json:"prompt"`
    Storyboard string `json:"storyboard"`
}

// ImageDataResponse 定义响应结构
type ImageDataResponse struct {
    ImageDataBase64 string `json:"image_base64"`
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

        // 解析响应（宽松假设，允许字符串或结构化内容）
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

        // 提取 prompt 和 storyboard
        var prompt, storyboard string
        content := textResp.Choices[0].Message.Content
        switch c := content.(type) {
        case string:
            log.Printf("收到字符串内容: %s", c)
            prompt = c
            storyboard = c
        case map[string]interface{}:
            log.Printf("收到结构化内容: %v", c)
            if p, ok := c["prompt"].(string); ok {
                prompt = p
            } else {
                log.Printf("提示词缺失，使用默认")
                prompt = "默认提示词"
            }
            if s, ok := c["storyboard"].(string); ok {
                storyboard = s
            } else {
                log.Printf("分镜缺失，使用默认")
                storyboard = prompt
            }
        default:
            log.Printf("未知内容格式: %v", c)
            //c.JSON(http.StatusInternalServerError, gin.H{"error": "外部 API 返回未知内容格式"})
            return
        }

        data := ImageDataRequest{
            Prompt:     prompt,
            Storyboard: storyboard,
        }

        // 保存到 output.json
        jsonData, err := json.Marshal(data)
        if err != nil {
            log.Printf("序列化 JSON 失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("序列化 JSON 失败: %v", err)})
            return
        }
        if err := os.WriteFile("output.json", jsonData, 0644); err != nil {
            log.Printf("写入 output.json 失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("写入 output.json 失败: %v", err)})
            return
        }

        log.Printf("成功保存 prompt: %s 和 storyboard: %s 到 output.json", data.Prompt, data.Storyboard)
        c.JSON(http.StatusOK, gin.H{
            "message": "Prompt 和 Storyboard 已保存到 output.json，等待确认",
            "data":    data,
        })
    })

    // 路由：生成图像（从 output.json 读取）
    r.POST("/generate_image", func(c *gin.Context) {
        log.Printf("开始处理 /generate_image 请求, 时间: %s", time.Now().Format(time.RFC3339))

        // 读取请求体（虽然前端可能发送空 body）
        var body bytes.Buffer
        if _, err := io.Copy(&body, c.Request.Body); err != nil {
            log.Printf("无法读取请求体: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
            return
        }
        log.Printf("接收到请求体: %s", body.String())

        // 读取 output.json
        jsonData, err := os.ReadFile("output.json")
        if err != nil {
            log.Printf("读取 output.json 失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取 output.json 失败: %v", err)})
            return
        }
        log.Printf("读取 output.json 内容: %s", string(jsonData))

        var data ImageDataRequest
        if err := json.Unmarshal(jsonData, &data); err != nil {
            log.Printf("解析 output.json 失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解析 output.json 失败: %v", err)})
            return
        }
        log.Printf("从 output.json 读取到数据 - 提示词: %s, 分镜: %s", data.Prompt, data.Storyboard)

        // 转发到 Colab
        colabData, err := json.Marshal(data)
        if err != nil {
            log.Printf("序列化 Colab 数据失败: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("序列化 Colab 数据失败: %v", err)})
            return
        }
        log.Printf("发送到 Colab 的数据: %s", string(colabData))

        resp, err := http.Post(colabEndpoint+"/generate", "application/json", bytes.NewBuffer(colabData))
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

        // 调用外部 API
        client := resty.New()
        resp, err := client.R().
            SetHeader("Authorization", "Bearer "+APIKey).
            SetHeader("Content-Type", "application/json").
            SetBody(body.Bytes()).
            Post(BaseURL + "/images/generations")

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

        c.Data(resp.StatusCode(), "application/json", resp.Body())
        log.Printf("成功返回图像生成响应")
    })

    // 启动服务
    log.Printf("启动服务器于 :3000")
    if err := r.Run(":3000"); err != nil {
        log.Fatal("服务器运行失败: ", err)
    }
}