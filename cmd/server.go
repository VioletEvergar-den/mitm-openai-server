package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/server"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
	"github.com/llm-sec/mitm-openai-server/pkg/utils"
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
  
  # API认证默认开启，使用basic认证方式和--auth-username/--auth-password参数提供的凭据
  # 而UI认证使用login.json中存储的凭据（默认用户名root，初次运行自动生成密码）
  # 如需禁用API认证，请使用 --enable-auth=false
  
  # 独立模式（默认带认证）
  mitm-openai-server server --auth-username admin --auth-password secret
  
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
	serverCmd.Flags().BoolVar(&enableAuth, "enable-auth", true, "启用API认证")
	serverCmd.Flags().StringVar(&authType, "auth-type", "basic", "认证类型 (basic 或 token)")
	serverCmd.Flags().StringVar(&authUsername, "auth-username", "admin", "API Basic认证用户名")
	serverCmd.Flags().StringVar(&authPassword, "auth-password", "", "API Basic认证密码")
	serverCmd.Flags().StringVar(&authToken, "auth-token", "", "API Token认证令牌")
	serverCmd.Flags().BoolVar(&enableCors, "enable-cors", true, "启用CORS")
	serverCmd.Flags().StringVar(&corsAllowedOrigins, "cors-allowed-origins", "*", "允许的源，用于CORS")

	// 前端UI相关标志
	serverCmd.Flags().BoolVar(&generateUIAuth, "generate-ui-auth", true, "生成随机UI认证凭证")
	serverCmd.Flags().StringVar(&uiUsername, "ui-username", "root", "前端UI用户名")
	serverCmd.Flags().StringVar(&uiPassword, "ui-password", "", "前端UI密码 (留空则自动生成)")
	serverCmd.Flags().StringVar(&uiDir, "ui-dir", "./react-ui/dist", "前端UI文件目录")

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
	// 设置Gin为发布模式，减少控制台输出
	gin.SetMode(gin.ReleaseMode)

	// 确定最终使用的数据目录
	finalDataDir := dataDir
	if finalDataDir == "" {
		// 如果没有指定路径，使用当前目录下的data文件夹
		currentDir, err := os.Getwd()
		if err == nil {
			finalDataDir = filepath.Join(currentDir, "data")
		} else {
			finalDataDir = "./data"
		}
	}

	// 确保数据目录存在
	absDataDir, err := filepath.Abs(finalDataDir)
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

	// 强制使用SQLite存储（用户隔离需要数据库支持）
	if verbose {
		log.Printf("使用SQLite数据库存储，数据库路径: %s/mitm_server.db", absDataDir)
	}

	storageImpl, storageErr := storage.NewSQLiteStorage(absDataDir)
	if storageErr != nil {
		log.Fatalf("初始化SQLite存储失败: %v", storageErr)
	}

	// 从 login.json 加载 UI 凭据（如果存在）
	loginFile := filepath.Join(absDataDir, "login.json")
	if _, err := os.Stat(loginFile); err == nil {
		loginData, err := os.ReadFile(loginFile)
		if err == nil {
			var loginCreds struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if json.Unmarshal(loginData, &loginCreds) == nil {
				if loginCreds.Username != "" {
					uiUsername = loginCreds.Username
				}
				if loginCreds.Password != "" {
					uiPassword = loginCreds.Password
				}
				if verbose {
					log.Printf("从 login.json 加载凭据: 用户名=%s", uiUsername)
				}
			}
		}
	}

	// 如果密码仍为空，生成随机密码
	if uiPassword == "" {
		uiPassword = generateRandomPassword(12)
		if verbose {
			log.Printf("生成随机密码: %s", uiPassword)
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

	// 打印服务器信息(带颜色)
	addr := fmt.Sprintf(":%d", port)

	// 定义颜色代码
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorWhite  = "\033[37m"
		colorBold   = "\033[1m"
	)

	// 获取UI配置
	serverConfig := apiServer.GetConfig()

	// 获取本机IP地址
	privateIPs := utils.GetPrivateIPs()
	publicIPs := utils.GetPublicIPs()

	// 精简登录信息显示 - 数据库版本
	fmt.Println()
	fmt.Printf("%s┌─────────────────────────────────────────────────┐%s\n", colorGreen+colorBold, colorReset)
	fmt.Printf("%s│%s%s            MITM OpenAI Server 已启动            %s%s│%s\n", colorGreen+colorBold, colorReset, colorWhite+colorBold, colorReset, colorGreen+colorBold, colorReset)
	fmt.Printf("%s│%s%s              (多用户数据库版本)                %s%s│%s\n", colorGreen+colorBold, colorReset, colorCyan, colorReset, colorGreen+colorBold, colorReset)
	fmt.Printf("%s└─────────────────────────────────────────────────┘%s\n", colorGreen+colorBold, colorReset)
	fmt.Println()
	fmt.Printf("%s%s登录地址:%s\n", colorBold, colorWhite, colorReset)
	fmt.Printf("  %s本地访问:%s   http://localhost:%d/ui/login\n", colorBlue, colorReset, port)

	if len(privateIPs) > 0 {
		fmt.Printf("  %s内网访问:%s   ", colorGreen, colorReset)
		for i, ip := range privateIPs {
			if i > 0 {
				fmt.Printf("              ")
			}
			fmt.Printf("http://%s:%d/ui/login\n", ip, port)
		}
	}

	if len(publicIPs) > 0 {
		fmt.Printf("  %s公网访问:%s   ", colorYellow, colorReset)
		for i, ip := range publicIPs {
			if i > 0 {
				fmt.Printf("              ")
			}
			fmt.Printf("http://%s:%d/ui/login\n", ip, port)
		}
	}

	fmt.Println()
	fmt.Printf("%s用户名:%s   %s%s%s\n", colorBold, colorReset, colorYellow+colorBold, serverConfig.UIUsername, colorReset)
	fmt.Printf("%s密码:%s     %s%s%s\n", colorBold, colorReset, colorYellow+colorBold, serverConfig.UIPassword, colorReset)
	fmt.Printf("%s数据库:%s   %s%s/mitm_server.db%s\n", colorBold, colorReset, colorCyan, absDataDir, colorReset)
	fmt.Printf("%s存储模式:%s %s多用户隔离 (SQLite)%s\n", colorBold, colorReset, colorGreen+colorBold, colorReset)
	fmt.Println()
	fmt.Printf("%s请使用上述凭据登录系统，监控和分析OpenAI API请求。%s\n", colorWhite, colorReset)
	fmt.Printf("%s每个用户的数据完全隔离，支持多用户同时使用。%s\n", colorGreen, colorReset)
	fmt.Println()

	// 启动服务器
	if err := apiServer.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// 确保UI目录存在
// 如果目录不存在则创建
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

	return nil
}

// generateRandomPassword 生成指定长度的随机密码
func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "defaultpassword"
	}
	return hex.EncodeToString(bytes)[:length]
}
