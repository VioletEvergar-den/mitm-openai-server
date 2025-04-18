package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/llm-sec/fake-openapi-server/pkg/api"
	"github.com/spf13/cobra"
)

var (
	// API服务器特定的标志
	port               int
	enableAuth         bool
	authUsername       string
	authPassword       string
	authToken          string
	authType           string
	cors               bool
	corsAllowedOrigins string

	// 前端UI相关标志
	uiUsername     string
	uiPassword     string
	generateUIAuth bool
)

// apiCmd 表示api子命令
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "启动API服务器",
	Long: `启动一个假的OpenAPI服务器，用于接收请求并记录。

例如:
  fake-openapi-server api --port 8080 --data ./data
  fake-openapi-server api --auth-type basic --auth-username admin --auth-password secret
  fake-openapi-server api --auth-type token --auth-token my-secret-token
  fake-openapi-server api --generate-ui-auth --ui-username admin`,
	Run: func(cmd *cobra.Command, args []string) {
		runAPIServer()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)

	// API服务器特定的标志
	apiCmd.Flags().IntVar(&port, "port", 8080, "服务器监听端口")
	apiCmd.Flags().BoolVar(&enableAuth, "enable-auth", false, "启用认证")
	apiCmd.Flags().StringVar(&authType, "auth-type", "basic", "认证类型 (basic, token)")
	apiCmd.Flags().StringVar(&authUsername, "auth-username", "", "基本认证用户名")
	apiCmd.Flags().StringVar(&authPassword, "auth-password", "", "基本认证密码")
	apiCmd.Flags().StringVar(&authToken, "auth-token", "", "令牌认证的令牌")
	apiCmd.Flags().BoolVar(&cors, "cors", false, "启用CORS支持")
	apiCmd.Flags().StringVar(&corsAllowedOrigins, "cors-allowed-origins", "*", "CORS允许的源 (逗号分隔)")

	// 前端UI相关标志
	apiCmd.Flags().BoolVar(&generateUIAuth, "generate-ui-auth", true, "生成随机UI认证凭证")
	apiCmd.Flags().StringVar(&uiUsername, "ui-username", "admin", "前端UI用户名")
	apiCmd.Flags().StringVar(&uiPassword, "ui-password", "", "前端UI密码 (留空则自动生成)")
}

func runAPIServer() {
	// 确保数据目录存在
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatalf("无法获取数据目录的绝对路径: %v", err)
	}

	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 确保UI目录存在
	ensureUIDirectory()

	if verbose {
		log.Printf("数据将存储在: %s", absDataDir)
	}

	// 创建存储
	storage, err := api.NewFileStorage(absDataDir)
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	// 创建服务器配置
	config := api.ServerConfig{
		Storage:        storage,
		EnableAuth:     enableAuth,
		AuthType:       authType,
		Username:       authUsername,
		Password:       authPassword,
		Token:          authToken,
		EnableCORS:     cors,
		AllowOrigins:   corsAllowedOrigins,
		UIUsername:     uiUsername,
		UIPassword:     uiPassword,
		GenerateUIAuth: generateUIAuth,
	}

	// 创建并启动服务器
	server := api.NewServerWithConfig(config)

	// 打印服务器信息
	addr := fmt.Sprintf(":%d", port)
	log.Printf("启动服务器，监听 %s", addr)
	log.Printf("前端界面: http://localhost%s/ui/", addr)
	log.Printf("OpenAPI规范: http://localhost%s/openapi.json", addr)
	log.Printf("健康检查: http://localhost%s/health", addr)
	log.Printf("API请求记录: http://localhost%s/api/requests", addr)

	// 认证配置信息
	if enableAuth {
		switch authType {
		case "basic":
			log.Printf("已启用基本认证 (用户名: %s)", authUsername)
		case "token":
			log.Printf("已启用令牌认证")
		default:
			log.Printf("已启用认证，类型: %s", authType)
		}
	} else {
		log.Printf("认证已禁用")
	}

	if err := server.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 确保UI目录存在
func ensureUIDirectory() {
	uiDir := "./ui"
	if _, err := os.Stat(uiDir); os.IsNotExist(err) {
		log.Printf("创建UI目录: %s", uiDir)
		if err := os.MkdirAll(uiDir, 0755); err != nil {
			log.Fatalf("创建UI目录失败: %v", err)
		}
	}
}
