package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/llm-sec/fake-openapi-server/pkg/api"
)

func main() {
	// 命令行参数
	port := flag.Int("port", 8080, "服务器监听端口")
	dataDir := flag.String("data", "./data", "请求数据存储目录")
	flag.Parse()

	// 确保数据目录存在
	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Fatalf("无法获取数据目录的绝对路径: %v", err)
	}

	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	log.Printf("数据将存储在: %s", absDataDir)

	// 创建存储
	storage, err := api.NewFileStorage(absDataDir)
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	// 创建服务器
	server := api.NewServer(storage)

	// 启动服务器
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("启动服务器，监听 %s", addr)
	log.Printf("OpenAPI规范: http://localhost%s/openapi.json", addr)
	log.Printf("健康检查: http://localhost%s/health", addr)
	log.Printf("API请求记录: http://localhost%s/api/requests", addr)

	if err := server.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
