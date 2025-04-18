package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/llm-sec/mitm-openai-server/pkg/server"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
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
	enableCors         bool
	corsAllowedOrigins string

	// 前端UI相关标志
	uiUsername     string
	uiPassword     string
	generateUIAuth bool
	uiDir          string

	// 中间人代理相关标志
	proxyMode      bool
	targetURL      string
	targetAuthType string
	targetUsername string
	targetPassword string
	targetToken    string

	// 存储相关标志
	useSQL     bool
	sqlitePath string
)

// serverCmd 表示server子命令
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动API服务器",
	Long: `启动一个中间人OpenAPI服务器，用于接收、转发和记录请求。

例如:
  # 启动独立模式服务器（返回模拟数据）
  mitm-openai-server server --port 8080 --data ./data
  
  # 带认证的独立模式
  mitm-openai-server server --enable-auth --auth-type basic --auth-username admin --auth-password secret
  
  # 代理模式（转发到真实API）
  mitm-openai-server server --proxy-mode --target-url https://api.example.com
  
  # 带认证的代理模式
  mitm-openai-server server --proxy-mode --target-url https://api.example.com --target-auth-type basic --target-username user --target-password pass`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// API服务器特定的标志
	serverCmd.Flags().IntVar(&port, "port", 8080, "服务器监听端口")
	serverCmd.Flags().BoolVar(&enableAuth, "enable-auth", false, "启用API认证")
	serverCmd.Flags().StringVar(&authType, "auth-type", "basic", "认证类型 (basic 或 token)")
	serverCmd.Flags().StringVar(&authUsername, "auth-username", "admin", "API Basic认证用户名")
	serverCmd.Flags().StringVar(&authPassword, "auth-password", "", "API Basic认证密码")
	serverCmd.Flags().StringVar(&authToken, "auth-token", "", "API Token认证令牌")
	serverCmd.Flags().BoolVar(&enableCors, "enable-cors", true, "启用CORS")
	serverCmd.Flags().StringVar(&corsAllowedOrigins, "cors-allowed-origins", "*", "允许的源，用于CORS")

	// 前端UI相关标志
	serverCmd.Flags().BoolVar(&generateUIAuth, "generate-ui-auth", true, "生成随机UI认证凭证")
	serverCmd.Flags().StringVar(&uiUsername, "ui-username", "admin", "前端UI用户名")
	serverCmd.Flags().StringVar(&uiPassword, "ui-password", "", "前端UI密码 (留空则自动生成)")
	serverCmd.Flags().StringVar(&uiDir, "ui-dir", "./ui", "前端UI文件目录")

	// 中间人代理相关标志
	serverCmd.Flags().BoolVar(&proxyMode, "proxy-mode", false, "启用代理模式（转发到真实API）")
	serverCmd.Flags().StringVar(&targetURL, "target-url", "", "目标API服务器地址")
	serverCmd.Flags().StringVar(&targetAuthType, "target-auth-type", "none", "目标API认证类型 (none, basic, token)")
	serverCmd.Flags().StringVar(&targetUsername, "target-username", "", "目标API基本认证用户名")
	serverCmd.Flags().StringVar(&targetPassword, "target-password", "", "目标API基本认证密码")
	serverCmd.Flags().StringVar(&targetToken, "target-token", "", "目标API令牌")

	// 存储相关标志
	serverCmd.Flags().BoolVar(&useSQL, "use-sql", false, "启用SQLite存储")
	serverCmd.Flags().StringVar(&sqlitePath, "sqlite-path", "", "SQLite数据库路径")
}

func runServer() {
	// 确保数据目录存在
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatalf("无法获取数据目录的绝对路径: %v", err)
	}

	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 确保UI目录存在
	absUIDir, err := filepath.Abs(uiDir)
	if err != nil {
		log.Fatalf("无法获取UI目录的绝对路径: %v", err)
	}

	if err := ensureUIDirectory(absUIDir); err != nil {
		log.Fatalf("UI目录准备失败: %v", err)
	}

	if verbose {
		log.Printf("数据将存储在: %s", absDataDir)
		log.Printf("UI文件目录: %s", absUIDir)
	}

	// 创建存储
	var storageImpl storage.Storage
	var storageErr error

	if useSQL {
		// 使用SQLite存储
		dbPath := sqlitePath
		if dbPath == "" {
			dbPath = filepath.Join(absDataDir, "requests.db")
		}

		log.Printf("使用SQLite存储，数据库路径: %s", dbPath)
		storageImpl, storageErr = storage.NewSQLiteStorage(dbPath)
		if storageErr != nil {
			log.Fatalf("初始化SQLite存储失败: %v", storageErr)
		}
	} else {
		// 使用文件存储
		log.Printf("使用文件存储，数据目录: %s", absDataDir)
		storageImpl, storageErr = storage.NewFileSystemStorage(absDataDir)
		if storageErr != nil {
			log.Fatalf("初始化文件存储失败: %v", storageErr)
		}
	}

	// 创建服务器配置
	config := server.ServerConfig{
		Storage:        storageImpl,
		EnableAuth:     enableAuth,
		AuthType:       authType,
		Username:       authUsername,
		Password:       authPassword,
		Token:          authToken,
		EnableCORS:     enableCors,
		AllowOrigins:   corsAllowedOrigins,
		UIUsername:     uiUsername,
		UIPassword:     uiPassword,
		GenerateUIAuth: generateUIAuth,
		UIDir:          absUIDir,

		// 中间人代理相关配置
		ProxyMode:      proxyMode,
		TargetURL:      targetURL,
		TargetAuthType: targetAuthType,
		TargetUsername: targetUsername,
		TargetPassword: targetPassword,
		TargetToken:    targetToken,
	}

	// 创建并启动服务器
	apiServer := server.NewServerWithConfig(config)

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

	// 代理模式信息
	if proxyMode {
		if targetURL == "" {
			log.Printf("警告: 代理模式已启用，但未指定目标URL，将使用模拟响应")
		} else {
			log.Printf("代理模式已启用，目标: %s", targetURL)

			if targetAuthType != "none" {
				log.Printf("目标API认证类型: %s", targetAuthType)
			} else {
				log.Printf("目标API认证已禁用")
			}
		}
	} else {
		log.Printf("独立模式，使用模拟响应")
	}

	if err := apiServer.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 确保UI目录存在
// 如果目录不存在则创建，并检查子目录
// 参数:
//   - uiDir: UI目录路径
//
// 返回:
//   - error: 如果发生错误，返回错误信息
func ensureUIDirectory(uiDir string) error {
	// 检查主目录是否存在
	stat, err := os.Stat(uiDir)
	if os.IsNotExist(err) {
		log.Printf("创建UI目录: %s", uiDir)
		if err := os.MkdirAll(uiDir, 0755); err != nil {
			return fmt.Errorf("创建UI目录失败: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("检查UI目录失败: %v", err)
	} else if !stat.IsDir() {
		return fmt.Errorf("UI路径已存在但不是目录: %s", uiDir)
	}

	// 确保CSS子目录存在
	cssDir := filepath.Join(uiDir, "css")
	if _, err := os.Stat(cssDir); os.IsNotExist(err) {
		log.Printf("创建CSS目录: %s", cssDir)
		if err := os.MkdirAll(cssDir, 0755); err != nil {
			return fmt.Errorf("创建CSS目录失败: %v", err)
		}
	}

	// 确保JS子目录存在
	jsDir := filepath.Join(uiDir, "js")
	if _, err := os.Stat(jsDir); os.IsNotExist(err) {
		log.Printf("创建JS目录: %s", jsDir)
		if err := os.MkdirAll(jsDir, 0755); err != nil {
			return fmt.Errorf("创建JS目录失败: %v", err)
		}
	}

	return nil
}
