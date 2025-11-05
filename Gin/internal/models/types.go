package models

import "time"

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
	StartIndex int    `json:"start_index"` // 起始章节索引（默认 0）
	Limit      int    `json:"limit"`       // 每次爬取章节数
}

// CrawlResponse 定义爬虫响应结构
type CrawlResponse struct {
	Chapters      []Chapter `json:"chapters"`
	NextIndex     int       `json:"next_index"`
	TotalChapters int       `json:"total_chapters"`
	Error         string    `json:"error,omitempty"`
}

// Chapter 定义单章节结果
type Chapter struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// NovelCache 缓存章节链接
type NovelCache struct {
	Chapters      []string
	LastCrawlTime time.Time
}
