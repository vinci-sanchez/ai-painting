package main

import (
	"log"

	"comic-proxy/internal/app"
	"comic-proxy/internal/config"
	"comic-proxy/internal/logging"
)

func main() {
	logFile, err := logging.Init()
	if err != nil {
		log.Fatal("无法打开日志文件: ", err)
	}
	defer logFile.Close()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("APIKey 已加载: %s", cfg.APIKey)
	log.Printf("BaseURL: %s, ColabEndpoint: %s", cfg.BaseURL, cfg.ColabEndpoint)

	router := app.NewRouter(cfg)

	log.Printf("启动服务器于 :3000")
	if err := router.Run(":3000"); err != nil {
		log.Fatal("服务器运行失败: ", err)
	}
}
