package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	//"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
)

// 配置日志输出到文件和终端
func init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件: ", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// 配置
var (
	APIKey        = ""
	BaseURL       = "https://ark.cn-beijing.volces.com/api/v3"
	colabEndpoint = "https://articular-proportionable-alverta.ngrok-free.dev/"
	novelCache    = make(map[string]*NovelCache)
	cacheMutex    sync.Mutex
	userAgents    = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
	}
)

// TextGenerationResponse 定义文本生成请求结构
type TextGenerationResponse struct {
	Scene     string `json:"scene"`
	Prompt    string `json:"prompt"`
	Character string `json:"character,omitempty"`
}

// ImageDataResponse 定义图像生成响应结构
type ImageDataResponse struct {
	ImageDataBase64 string `json:"image_base64"`
}

// ImageDataResponse1 定义火山方舟 API 图像响应结构
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

// CrawlRequest 定义爬虫请求结构
type CrawlRequest struct {
	NovelURL   string `json:"novel_url"`   // 可为目录页或章节页
	StartIndex int    `json:"start_index"` // 起始章节索引（默认0）
	Limit      int    `json:"limit"`       // 每次爬取章节数
}

// CrawlResponse 定义爬虫响应结构
type CrawlResponse struct {
	Chapters      []Chapter `json:"chapters"`
	NextIndex     int       `json:"next_index"`
	TotalChapters int       `json:"total_chapters"`
	Error         string    `json:"error,omitempty"`
}

// Chapter 定义单章节结构
type Chapter struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// NovelCache 缓存章节链接
type NovelCache struct {
	Chapters      []string
	LastCrawlTime time.Time
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

		var body bytes.Buffer
		if _, err := io.Copy(&body, c.Request.Body); err != nil {
			log.Printf("无法读取请求体: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
			return
		}
		log.Printf("接收到请求体: %s", body.String())

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
			log.Printf("外部 API 响应无效: 无 choices 数据")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "外部 API 响应无效: 无 choices 数据"})
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
	})

	// 代理：图像生成（Colab）
	r.POST("/generate_image", func(c *gin.Context) {
		log.Printf("开始处理 /generate_image 请求, 时间: %s", time.Now().Format(time.RFC3339))

		var body bytes.Buffer
		if _, err := io.Copy(&body, c.Request.Body); err != nil {
			log.Printf("无法读取请求体: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
			return
		}
		log.Printf("接收到请求体: %s", body.String())

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

	// 代理：图像生成（火山方舟）
	r.POST("/api/image", func(c *gin.Context) {
		log.Printf("开始处理 /api/image 请求, 时间: %s", time.Now().Format(time.RFC3339))

		var body bytes.Buffer
		if _, err := io.Copy(&body, c.Request.Body); err != nil {
			log.Printf("无法读取请求体: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无法读取请求体: %v", err)})
			return
		}
		log.Printf("接收到请求体: %s", body.String())

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

		arkData := struct {
			Model     string `json:"model"`
			Prompt    string `json:"prompt"`
			Role      string `json:"role,omitempty"`
			Storyboard string `json:"storyboard,omitempty"`
		}{
			Model:     "ep-20251026170710-8pgpm",
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

		log.Printf("火山方舟 API 状态码: %d", resp.StatusCode())
		log.Printf("火山方舟 API 响应体: %s", string(resp.Body()))
		if resp.StatusCode() != http.StatusOK {
			log.Printf("火山方舟 API 返回非 200 状态码: %d", resp.StatusCode())
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("火山方舟 API 返回非 200 状态码: %d", resp.StatusCode())})
			return
		}

		var arkResponse ImageDataResponse1
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
	})

	// 新增：小说爬虫路由
	r.POST("/api/crawl", func(c *gin.Context) {
		log.Printf("开始处理 /api/crawl 请求, 时间: %s", time.Now().Format(time.RFC3339))

		var req CrawlRequest
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

		// 推断目录页URL
		directoryURL := inferDirectoryURL(req.NovelURL)
		log.Printf("推断目录页URL: %s (原URL: %s)", directoryURL, req.NovelURL)

		if !strings.HasPrefix(directoryURL, "http") {
			log.Printf("无效的URL: %s", directoryURL)
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入有效的URL（目录页或章节页）"})
			return
		}

		// 强制从第一章开始
		if req.StartIndex != 0 {
			log.Printf("强制从第一章开始，忽略 start_index=%d", req.StartIndex)
			req.StartIndex = 0
		}
		// 强制最多爬取一章
		req.Limit = 1
		log.Printf("爬虫参数: 目录=%s, StartIndex=%d, Limit=%d", directoryURL, req.StartIndex, req.Limit)

		// 检查缓存
		cacheMutex.Lock()
		cache, exists := novelCache[directoryURL]
		if !exists || time.Since(cache.LastCrawlTime).Hours() > 24 {
			log.Printf("缓存不存在或已过期，重新爬取章节列表: %s", directoryURL)
			chapters, err := crawlChapterList(directoryURL)
			if err != nil {
				cacheMutex.Unlock()
				log.Printf("爬取章节列表失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("爬取章节列表失败: %v", err)})
				return
			}
			if len(chapters) == 0 {
				cacheMutex.Unlock()
				log.Printf("未找到任何章节链接: %s", directoryURL)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到任何章节链接，请检查URL或站点结构（建议试试 https://www.xsbiquge.com/）"})
				return
			}
			cache = &NovelCache{Chapters: chapters, LastCrawlTime: time.Now()}
			novelCache[directoryURL] = cache
			log.Printf("缓存更新: 找到 %d 个章节链接", len(chapters))
		}
		cacheMutex.Unlock()

		if req.StartIndex >= len(cache.Chapters) {
			log.Printf("起始索引超出范围: %d >= %d", req.StartIndex, len(cache.Chapters))
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("起始索引超出总章节数: %d", len(cache.Chapters))})
			return
		}

		endIndex := req.StartIndex + req.Limit
		if endIndex > len(cache.Chapters) {
			endIndex = len(cache.Chapters)
		}
		log.Printf("开始爬取章节 %d-%d", req.StartIndex, endIndex-1)
		chapters, err := crawlChapters(directoryURL, cache.Chapters[req.StartIndex:endIndex])
		if err != nil {
			log.Printf("爬取章节内容失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("爬取章节内容失败: %v", err)})
			return
		}

		// 保存到文件
		if len(chapters) > 0 {
			if err := saveToFile(chapters, directoryURL); err != nil {
				log.Printf("保存文件失败: %v", err)
			}
		}

		resp := CrawlResponse{
			Chapters:      chapters,
			NextIndex:     endIndex,
			TotalChapters: len(cache.Chapters),
		}
		log.Printf("爬取成功: 返回 %d 章, 下次索引 %d, 总章节 %d", len(chapters), endIndex, len(cache.Chapters))
		c.JSON(http.StatusOK, resp)
	})

	// 启动服务
	log.Printf("启动服务器于 :3000")
	if err := r.Run(":3000"); err != nil {
		log.Fatal("服务器运行失败: ", err)
	}
}

// inferDirectoryURL 从章节URL推断目录页
func inferDirectoryURL(novelURL string) string {
	parsed, err := url.Parse(novelURL)
	if err != nil {
		log.Printf("URL解析失败: %v", err)
		return novelURL
	}
	path := parsed.Path
	if strings.HasSuffix(path, ".html") {
		// 从章节页（如 /3_3037/22745259.html）推断目录页（如 /3_3037/）
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) >= 2 {
			path = "/" + parts[0] + "/" + parts[1] + "/"
		}
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	parsed.Path = path
	return parsed.String()
}

// crawlChapterList 爬取小说章节列表
func crawlChapterList(novelURL string) ([]string, error) {
	c := colly.NewCollector(
		colly.UserAgent(userAgents[rand.Intn(len(userAgents))]),
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// 设置请求头
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", novelURL)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
	})

	var chapters []string
	parsedURL, err := url.Parse(novelURL)
	if err != nil {
		return nil, fmt.Errorf("URL解析失败: %v", err)
	}
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host + "/"
	novelID := parsedURL.Path
	if strings.HasSuffix(novelID, "/") {
		novelID = novelID[:len(novelID)-1]
	}
	log.Printf("BaseURL: %s, NovelID: %s", baseURL, novelID)

	// 调试：记录所有 <dt> 的内容
	c.OnHTML(`div.listmain dl dt`, func(e *colly.HTMLElement) {
		log.Printf("找到 <dt>: %s", e.Text)
	})

	// 精确选择第二个 <dt> 后的 <dd> 中的 <a>
	c.OnHTML(`div.listmain dl dt:nth-of-type(2) ~ dd a[href]`, func(e *colly.HTMLElement) {
		href := e.Attr("href")
		fullURL := resolveURL(baseURL, href)
		// 过滤无关链接，确保包含小说ID（如 /3_3037/）且以.html结尾
		if strings.Contains(fullURL, novelID) && strings.HasSuffix(fullURL, ".html") {
			chapters = append(chapters, fullURL)
			log.Printf("找到章节链接: %s (标题: %s)", fullURL, e.Text)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("爬取章节列表错误: %v, URL: %s, 状态码: %d", err, r.Request.URL, r.StatusCode)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("目录页响应: 状态码=%d, 长度=%d, 预览=%s...", r.StatusCode, len(r.Body), string(r.Body[:min(300, len(r.Body))]))
	})

	err = c.Visit(novelURL)
	if err != nil {
		return nil, fmt.Errorf("访问目录页失败: %v", err)
	}

	c.Wait()
	if len(chapters) == 0 {
		return nil, fmt.Errorf("未找到任何章节链接，可能站点结构变化或URL无效")
	}
	log.Printf("找到有效章节链接: %d 个", len(chapters))
	return chapters, nil
}

// crawlChapters 爬取指定章节内容
func crawlChapters(novelURL string, chapterURLs []string) ([]Chapter, error) {
	var chapters []Chapter
	c := colly.NewCollector(
		colly.UserAgent(userAgents[rand.Intn(len(userAgents))]),
		colly.Async(true),
	)

	// 设置请求头
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", novelURL)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
	})

	// 尝试多种正文选择器
	contentSelectors := []string{
		"div#content",
		"div#htmlContent",
		"div.showtxt",
		"div.read-content",
		"div[id*='content']",
		"div.articlecontent",
	}

	for _, selector := range contentSelectors {
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			// 获取标题
			title := strings.TrimSpace(e.DOM.Closest("body").Find("h1").First().Text())
			if title == "" {
				title = strings.TrimSpace(e.DOM.Closest("body").Find("title").Text())
			}
			// 清理正文
			content := strings.TrimSpace(e.Text)
			content = strings.ReplaceAll(content, "\u00a0", " ")
			content = strings.ReplaceAll(content, "\r\n", "\n")
			content = strings.ReplaceAll(content, "\n\n", "\n")
			if len(content) > 100 { // 过滤无效内容
				chapters = append(chapters, Chapter{Title: title, Content: content})
				log.Printf("爬取章节成功: %s (长度: %d, 选择器: %s)", title, len(content), selector)
			}
		})
	}

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("爬取章节错误: %v, URL: %s, 状态码: %d", err, r.Request.URL, r.StatusCode)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("章节响应: URL=%s, 状态码=%d, 长度=%d", r.Request.URL, r.StatusCode, len(r.Body))
	})

	for i, chapURL := range chapterURLs {
		maxRetries := 3
		for retry := 0; retry < maxRetries; retry++ {
			err := c.Visit(chapURL)
			if err == nil {
				break
			}
			log.Printf("章节 %d 重试 %d/%d: %s (错误: %v)", i, retry+1, maxRetries, chapURL, err)
			time.Sleep(time.Duration(retry+1) * time.Second)
		}
		time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
	}

	c.Wait()
	if len(chapters) == 0 {
		return nil, fmt.Errorf("未爬取到有效章节内容，可能选择器失效或站点限制")
	}
	return chapters, nil
}

// resolveURL 安全解析相对URL
func resolveURL(base, href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	baseParsed, _ := url.Parse(base)
	relParsed, _ := url.Parse(href)
	return baseParsed.ResolveReference(relParsed).String()
}

// saveToFile 保存章节到文件
func saveToFile(chapters []Chapter, novelURL string) error {
	parsedURL, _ := url.Parse(novelURL)
	filename := "novel_" + url.PathEscape(parsedURL.Path) + ".txt"
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("无法保存文件: %v", err)
		return err
	}
	defer file.Close()
	for _, ch := range chapters {
		file.WriteString(ch.Title + "\n\n" + ch.Content + "\n\n")
	}
	log.Printf("保存章节到文件: %s (%d 章)", filename, len(chapters))
	return nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}