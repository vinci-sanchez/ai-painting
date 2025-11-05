package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultBaseURL       = "https://ark.cn-beijing.volces.com/api/v3"
	defaultColabEndpoint = "https://articular-proportionable-alverta.ngrok-free.dev/"
)

// Config 保存应用运行所需的配置信息。
type Config struct {
	APIKey        string
	BaseURL       string
	ColabEndpoint string
}

// Load 从环境变量加载配置，同时保留与旧程序一致的默认行为。
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("警告：未找到 .env 文件，使用默认配置: %v", err)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY 未设置，请在 .env 文件中配置")
	}

	colabEndpoint := os.Getenv("COLAB_ENDPOINT")
	if colabEndpoint == "" {
		colabEndpoint = defaultColabEndpoint
	}

	return &Config{
		APIKey:        apiKey,
		BaseURL:       defaultBaseURL,
		ColabEndpoint: colabEndpoint,
	}, nil
}
