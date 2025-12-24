package crawler

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"

	"comic-proxy/internal/models"
)

// Service 提供小说爬取相关功能，并管理章节缓存。
type Service struct {
	cache      map[string]*models.NovelCache
	cacheMutex sync.Mutex
	userAgents []string
}

// Error 描述爬虫过程中的可预期错误。
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// NewService 创建带默认配置的 Service。
func NewService() *Service {
	return &Service{
		cache: make(map[string]*models.NovelCache),
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
		},
	}
}

// InferDirectoryURL 从章节 URL 推断目录页地址。
func (s *Service) InferDirectoryURL(novelURL string) string {
	parsed, err := url.Parse(novelURL)
	if err != nil {
		log.Printf("URL解析失败: %v", err)
		return novelURL
	}
	path := parsed.Path
	if strings.HasSuffix(path, ".html") {
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

// FetchChapters 根据目录页、起始索引与数量抓取章节内容。
func (s *Service) FetchChapters(directoryURL string, startIndex, limit int) ([]models.Chapter, int, int, error) {
	if limit <= 0 {
		limit = 1
	}

	cache, totalChapters, err := s.prepareCache(directoryURL)
	if err != nil {
		return nil, 0, 0, err
	}

	if startIndex >= totalChapters {
		return nil, 0, totalChapters, &Error{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("起始索引超出总章节数: %d", totalChapters),
		}
	}

	endIndex := startIndex + limit
	if endIndex > totalChapters {
		endIndex = totalChapters
	}

	log.Printf("开始爬取章节 %d-%d", startIndex, endIndex-1)
	chapters, crawlErr := s.crawlChapters(directoryURL, cache.Chapters[startIndex:endIndex])
	if crawlErr != nil {
		return nil, 0, totalChapters, crawlErr
	}

	if len(chapters) > 0 {
		if err := s.saveToFile(chapters, directoryURL); err != nil {
			log.Printf("保存文件失败: %v", err)
		}
	}

	return chapters, endIndex, totalChapters, nil
}

func (s *Service) prepareCache(directoryURL string) (*models.NovelCache, int, error) {
	s.cacheMutex.Lock()
	cache, exists := s.cache[directoryURL]
	if exists && time.Since(cache.LastCrawlTime).Hours() <= 24 {
		total := len(cache.Chapters)
		s.cacheMutex.Unlock()
		return cache, total, nil
	}
	s.cacheMutex.Unlock()

	log.Printf("缓存不存在或已过期，重新爬取章节列表: %s", directoryURL)
	chapters, err := s.crawlChapterList(directoryURL)
	if err != nil {
		return nil, 0, err
	}
	if len(chapters) == 0 {
		return nil, 0, fmt.Errorf("未找到任何章节链接，请检查URL或站点结构（建议试试 https://www.xsbiquge.com/）")
	}

	s.cacheMutex.Lock()
	cache = &models.NovelCache{Chapters: chapters, LastCrawlTime: time.Now()}
	s.cache[directoryURL] = cache
	total := len(cache.Chapters)
	s.cacheMutex.Unlock()
	log.Printf("缓存更新: 找到 %d 个章节链接", total)
	return cache, total, nil
}

func (s *Service) crawlChapterList(novelURL string) ([]string, error) {
	c := colly.NewCollector(
		colly.UserAgent(s.userAgents[rand.Intn(len(s.userAgents))]),
		colly.MaxDepth(1),
		colly.Async(true),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", novelURL)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
	})

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

	var chapters []string
	c.OnHTML(`div.listmain dl dt`, func(e *colly.HTMLElement) {
		log.Printf("找到 <dt>: %s", e.Text)
	})

	c.OnHTML(`div.listmain dl dt:nth-of-type(2) ~ dd a[href]`, func(e *colly.HTMLElement) {
		href := e.Attr("href")
		fullURL := s.resolveURL(baseURL, href)
		if strings.Contains(fullURL, novelID) && strings.HasSuffix(fullURL, ".html") {
			chapters = append(chapters, fullURL)
			//log.Printf("找到章节链接: %s (标题: %s)", fullURL, e.Text)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("爬取章节列表错误: %v, URL: %s, 状态码: %d", err, r.Request.URL, r.StatusCode)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("目录页响应: 状态码=%d, 长度=%d, 预览=%s...", r.StatusCode, len(r.Body), string(r.Body[:min(300, len(r.Body))]))
	})

	if err := c.Visit(novelURL); err != nil {
		return nil, fmt.Errorf("访问目录页失败: %v", err)
	}

	c.Wait()
	if len(chapters) == 0 {
		return nil, fmt.Errorf("未找到任何章节链接，可能站点结构变化或URL无效")
	}
	log.Printf("找到有效章节链接: %d 个", len(chapters))
	return chapters, nil
}

func (s *Service) crawlChapters(novelURL string, chapterURLs []string) ([]models.Chapter, error) {
	var chapters []models.Chapter
	c := colly.NewCollector(
		colly.UserAgent(s.userAgents[rand.Intn(len(s.userAgents))]),
		colly.Async(true),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", novelURL)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
	})

	contentSelectors := []string{
		"div#content",
		"div#htmlContent",
		"div.showtxt",
		"div.read-content",
		"div[id*='content']",
		"div.articlecontent",
	}
	var isChapterFound bool = false
	for _, selector := range contentSelectors {
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			if isChapterFound {
				return // 如果已找到，则立即退出回调，不执行后续代码
			}
			title := strings.TrimSpace(e.DOM.Closest("body").Find("h1").First().Text())
			if title == "" {
				title = strings.TrimSpace(e.DOM.Closest("body").Find("title").Text())
			}
			content := strings.TrimSpace(e.Text)
			content = strings.ReplaceAll(content, "\u00a0", " ")
			content = strings.ReplaceAll(content, "\r\n", "\n")
			content = strings.ReplaceAll(content, "\n\n", "\n")
			if len(content) > 100 {
				chapters = append(chapters, models.Chapter{Title: title, Content: content})
				log.Printf("爬取章节成功: %s (长度: %d, 选择器: %s)", title, len(content), selector)
				isChapterFound = true
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
			if err := c.Visit(chapURL); err == nil {
				break
			} else if retry == maxRetries-1 {
				log.Printf("章节 %d 访问失败: %s", i, chapURL)
			} else {
				log.Printf("章节 %d 重试 %d/%d: %s", i, retry+1, maxRetries, chapURL)
				time.Sleep(time.Duration(retry+1) * time.Second)
			}
		}
		time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
	}

	c.Wait()
	if len(chapters) == 0 {
		return nil, fmt.Errorf("未爬取到有效章节内容，可能选择器失效或站点限制")
	}
	return chapters, nil
}

func (s *Service) saveToFile(chapters []models.Chapter, novelURL string) error {
	parsedURL, _ := url.Parse(novelURL)
	filename := "novel_" + url.PathEscape(parsedURL.Path) + ".txt"
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, ch := range chapters {
		if _, err := file.WriteString(ch.Title + "\n\n" + ch.Content + "\n\n"); err != nil {
			return err
		}
	}
	log.Printf("保存章节到文件 %s (%d 个)", filename, len(chapters))
	return nil
}

func (s *Service) resolveURL(base, href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	baseParsed, _ := url.Parse(base)
	relParsed, _ := url.Parse(href)
	return baseParsed.ResolveReference(relParsed).String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
