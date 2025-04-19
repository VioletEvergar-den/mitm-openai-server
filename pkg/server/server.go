package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/llm-sec/mitm-openai-server/pkg/api"
	"github.com/llm-sec/mitm-openai-server/pkg/embed"
	"github.com/llm-sec/mitm-openai-server/pkg/openai"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
	"github.com/llm-sec/mitm-openai-server/pkg/utils"
)

// Server 表示API服务器
// 是服务器的核心结构，包含路由引擎、存储接口和配置
type Server struct {
	router        *gin.Engine           // Gin路由引擎
	storage       storage.Storage       // 存储接口
	config        api.ServerConfig      // 服务器配置
	openaiService openai.Service        // OpenAI服务接口
	configManager *ConfigManager        // 配置管理器
	uiServer      api.UIServerInterface // UI服务器接口
	openaiHandler interface{}           // OpenAI API处理器，使用openai.Handler
	storagePath   string                // 用户配置的数据存储路径
}

// NewServer 创建一个新的服务器实例
// 使用默认配置创建服务器，只需提供存储接口
// 参数:
//   - storage: 用于存储请求数据的存储接口
//
// 返回:
//   - *Server: 服务器实例
func NewServer(storage storage.Storage) *Server {
	return NewServerWithConfig(api.ServerConfig{
		Storage: storage,
	})
}

// NewServerWithConfig 使用给定的配置创建服务器
// 使用指定的配置创建和初始化Server实例
// 参数:
//   - config: 服务器配置
//
// 返回:
//   - *Server: 初始化的服务器实例
func NewServerWithConfig(config api.ServerConfig) *Server {
	// 如果未设置存储，则返回错误
	if config.Storage == nil {
		log.Fatal("必须提供存储实例")
	}

	// 创建配置管理器
	configManager, err := NewConfigManager()
	if err != nil {
		fmt.Printf("警告: 无法创建配置管理器: %v\n", err)
	}

	// 创建默认路由
	router := gin.Default()

	// 创建服务器实例
	server := &Server{
		router:        router,
		storage:       config.Storage,
		config:        config,
		configManager: configManager,
	}

	// 如果配置管理器创建成功，尝试加载配置
	if configManager != nil {
		userConfig, err := configManager.LoadConfig()
		if err != nil {
			fmt.Printf("警告: 无法加载用户配置: %v\n", err)
		} else {
			// 应用用户配置
			configManager.ApplyConfig(userConfig, server)
		}
	}

	// 如果启用了生成UI认证并且没有设置密码，检查是否有保存的密码，否则生成一个新的随机密码
	if config.GenerateUIAuth && config.UIPassword == "" && server.config.UIPassword == "" {
		// 如果密码为空，生成一个新的随机密码
		newPassword := utils.GenerateRandomPassword(12)
		server.config.UIPassword = newPassword

		// 尝试保存新生成的密码到配置文件
		if configManager != nil {
			username := server.config.UIUsername // 通常是"root"
			if err := configManager.SaveUICredentials(username, newPassword); err != nil {
				fmt.Printf("警告: 无法保存UI凭据到配置文件: %v\n", err)
			} else {
				fmt.Printf("信息: UI凭据已保存到配置文件\n")
			}
		}
	} else if server.config.UIPassword != "" {
		// 使用已有密码，确保GenerateUIAuth设置为false，避免日志显示错误信息
		server.config.GenerateUIAuth = false
		fmt.Printf("信息: 使用配置文件中的UI密码\n")
	}

	// 初始化OpenAI服务
	if config.ProxyMode {
		// 创建OpenAI服务配置
		openaiConfig := openai.Config{
			Enabled:         true,
			ResponseDelayMs: 0,     // 在代理模式下不添加额外延迟
			APIKeyAuth:      false, // 不在服务器端验证API密钥
			ProxyMode:       true,
			TargetURL:       config.TargetURL,
			TargetAuthType:  config.TargetAuthType,
			TargetUsername:  config.TargetUsername,
			TargetPassword:  config.TargetPassword,
			TargetToken:     config.TargetToken,
		}
		server.openaiService = openai.NewService(openaiConfig)
	} else if !config.ProxyMode && config.EnableAuth {
		// 创建带有API密钥验证的模拟服务
		openaiConfig := openai.Config{
			Enabled:         true,
			ResponseDelayMs: 100, // 添加一些响应延迟以模拟真实API
			APIKeyAuth:      true,
			APIKey:          config.Token, // 使用同一个token作为API密钥
			ProxyMode:       false,
		}
		server.openaiService = openai.NewService(openaiConfig)
	} else {
		// 创建不需要验证的模拟服务
		openaiConfig := openai.DefaultConfig()
		openaiConfig.APIKeyAuth = false
		server.openaiService = openai.NewService(openaiConfig)
	}

	// 初始化OpenAI全局处理器，用于外部保存请求
	if server.openaiService != nil {
		openai.InitGlobalHandler(config.Storage, server.openaiService)
	}

	// 创建UI服务器
	server.uiServer = api.UIServerFactory(config.Storage, config, server.openaiService)

	// 创建OpenAI处理器 - 直接使用openai包中的Handler
	openaiHandler := openai.NewHandler(config.Storage, server.openaiService)
	server.openaiHandler = openaiHandler

	// 设置路由
	server.setupRoutes()
	server.setupOpenAIRoutes() // 设置OpenAI相关路由

	return server
}

// Run 启动服务器
// 在指定地址上启动HTTP服务器
// 参数:
//   - addr: 服务器监听地址，格式为"host:port"
//
// 返回:
//   - error: 如果服务器启动失败，返回错误
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// authMiddleware 认证中间件
// 提供API认证功能，支持basic和token两种认证方式
// 返回:
//   - gin.HandlerFunc: Gin中间件函数
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果认证被禁用，则跳过
		if !s.config.EnableAuth {
			c.Next()
			return
		}

		// 根据配置的认证类型进行认证
		var authorized bool
		switch s.config.AuthType {
		case "basic": // 基本认证
			authorized = s.validateBasicAuth(c)
		case "token": // 令牌认证
			authorized = s.validateTokenAuth(c)
		default:
			// 不支持的认证类型，拒绝访问
			c.JSON(http.StatusUnauthorized, StandardResponse{
				Code: 10001,
				Msg:  "认证配置错误",
			})
			c.Abort()
			return
		}

		// 认证失败，返回401状态码
		if !authorized {
			c.JSON(http.StatusUnauthorized, StandardResponse{
				Code: 10002,
				Msg:  "认证失败",
			})
			c.Abort()
			return
		}

		// 认证成功，继续处理请求
		c.Next()
	}
}

// validateBasicAuth 验证基本认证
// 检查HTTP Basic认证的用户名和密码是否匹配配置
// 参数:
//   - c: Gin上下文
//
// 返回:
//   - bool: 认证是否成功
func (s *Server) validateBasicAuth(c *gin.Context) bool {
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		return false // 没有提供认证信息
	}

	return username == s.config.Username && password == s.config.Password
}

// validateTokenAuth 验证令牌认证
// 检查Authorization头中的令牌是否匹配配置
// 参数:
//   - c: Gin上下文
//
// 返回:
//   - bool: 认证是否成功
func (s *Server) validateTokenAuth(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false // 没有提供认证头
	}

	// 支持"Bearer token"格式和直接的令牌格式
	token := authHeader
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = authHeader[7:] // 去掉"Bearer "前缀
	}

	return token == s.config.Token
}

// corsMiddleware CORS中间件
// 处理跨域资源共享，允许来自不同域的请求
// 返回:
//   - gin.HandlerFunc: Gin中间件函数
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.config.EnableCORS {
			// 设置CORS相关的HTTP头
			c.Writer.Header().Set("Access-Control-Allow-Origin", s.config.AllowOrigins)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

			// 处理预检请求
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent) // 返回204状态码
				return
			}
		}
		c.Next()
	}
}

// setupRoutes 配置服务器的所有HTTP路由和中间件
// 设置HTTP路由处理程序，包括API路由、认证中间件和CORS中间件
func (s *Server) setupRoutes() {
	// 设置全局中间件
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(s.corsMiddleware())

	// 健康检查路由
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API路由组，有认证中间件保护
	apiGroup := s.router.Group("/api")
	apiGroup.Use(s.authMiddleware())
	apiGroup.Use(s.apiMiddleware())

	// 设置UI相关路由
	s.setupUIRoutes()
}

// setupUIRoutes 设置UI相关的路由
// 处理静态文件和用户登录
func (s *Server) setupUIRoutes() {
	// 确保UI路径有效
	uiPath := s.config.UIDir
	if uiPath == "" {
		uiPath = "./ui" // 默认UI目录
	}

	// 获取文件系统
	fileSystem := embed.GetFS(uiPath)

	// 设置UI路由
	s.uiServer.SetupUIRoutes(s.router, s.authMiddleware(), s.apiMiddleware())

	// 单独处理基本UI路由，确保index.html总是被加载
	s.router.GET("/ui", func(c *gin.Context) {
		c.File(embed.ResolvePath(uiPath, "index.html"))
	})
	s.router.GET("/ui/", func(c *gin.Context) {
		c.File(embed.ResolvePath(uiPath, "index.html"))
	})

	// 根路径直接提供内容，不重定向
	s.router.GET("/", func(c *gin.Context) {
		c.File(embed.ResolvePath(uiPath, "index.html"))
	})

	// 处理静态文件，注意顺序
	s.router.StaticFS("/ui/static", fileSystem)
	s.router.StaticFS("/ui/css", fileSystem)
	s.router.StaticFS("/ui/js", fileSystem)
	s.router.StaticFS("/ui/assets", fileSystem)

	// 根路径下的静态文件
	s.router.StaticFS("/static", fileSystem)
	s.router.StaticFS("/css", fileSystem)
	s.router.StaticFS("/js", fileSystem)
	s.router.StaticFS("/assets", fileSystem)

	// 任何未处理的UI路由都重定向到index.html，实现SPA路由
	s.router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/ui/") {
			c.File(embed.ResolvePath(uiPath, "index.html"))
		} else if !strings.HasPrefix(c.Request.URL.Path, "/api/") &&
			!strings.HasPrefix(c.Request.URL.Path, "/v1/") {
			// 对于非API路径的请求，也返回前端应用
			c.File(embed.ResolvePath(uiPath, "index.html"))
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		}
	})
}

// setupOpenAIRoutes 设置OpenAI API代理路由
// 配置所有与OpenAI API相关的路由
func (s *Server) setupOpenAIRoutes() {
	// 如果没有初始化openaiHandler，跳过设置OpenAI路由
	if s.openaiHandler == nil {
		fmt.Println("警告: openaiHandler未初始化，跳过OpenAI路由设置")
		return
	}

	// 使用OpenAI处理器设置路由 - 直接使用Handler的SetupRoutes方法
	if handler, ok := s.openaiHandler.(*openai.Handler); ok {
		handler.SetupRoutes(s.router, s.apiMiddleware())
	} else {
		fmt.Println("警告: openaiHandler类型错误，无法设置OpenAI路由")
	}
}

// apiMiddleware API中间件
// 记录所有API请求，包括请求和响应数据
// 返回:
//   - gin.HandlerFunc: Gin中间件函数
func (s *Server) apiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求开始时间
		startTime := time.Now()

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 捕获请求
		requestID := uuid.New().String()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := make(map[string][]string)

		// 复制URL查询参数
		for k, v := range c.Request.URL.Query() {
			query[k] = v
		}

		// 复制请求头
		headers := make(map[string][]string)
		for k, v := range c.Request.Header {
			headers[k] = v
		}

		// 读取请求体
		var bodyBytes []byte
		var err error

		if c.Request.Body != nil {
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				// 如果读取失败，记录错误并继续
				c.JSON(http.StatusInternalServerError, api.StandardResponse{
					Code: 10005,
					Msg:  "无法读取请求体: " + err.Error(),
				})
				c.Abort()
				return
			}

			// 恢复请求体，以便后续处理程序可以再次读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 解析请求体为JSON（如果可能）
		var bodyData interface{}
		if len(bodyBytes) > 0 {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// 尝试解析JSON
				if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
					// 如果解析失败，使用原始字符串
					bodyData = string(bodyBytes)
				}
			} else {
				// 非JSON内容，使用原始字符串
				bodyData = string(bodyBytes)
			}
		}

		// 创建一个响应记录器来捕获响应
		responseRecorder := utils.NewResponseRecorder(c.Writer)
		c.Writer = responseRecorder

		// 处理请求，这将调用下一个中间件或路由处理程序
		c.Next()

		// 计算请求处理时间
		latency := time.Since(startTime).Milliseconds()

		// 捕获响应数据
		statusCode := responseRecorder.Status()
		responseHeaders := responseRecorder.Header()
		responseBodyBytes := responseRecorder.Body.Bytes()

		// 解析响应体为JSON（如果可能）
		var responseBody interface{}
		if len(responseBodyBytes) > 0 {
			contentType := responseHeaders.Get("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// 尝试解析JSON
				if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
					// 如果解析失败，使用原始字符串
					responseBody = string(responseBodyBytes)
				}
			} else {
				// 非JSON内容，使用原始字符串
				responseBody = string(responseBodyBytes)
			}
		}

		// 创建响应对象
		response := &storage.ProxyResponse{
			StatusCode: statusCode,
			Headers:    responseHeaders,
			Body:       responseBody,
			Latency:    latency,
		}

		// 创建请求记录
		request := &storage.Request{
			ID:        requestID,
			Method:    method,
			Path:      path,
			Timestamp: time.Now(),
			Headers:   headers,
			Query:     query,
			Body:      bodyData,
			ClientIP:  clientIP,
			Response:  response,
			Metadata: map[string]interface{}{
				"latency_ms": latency,
			},
		}

		// 保存请求到存储
		if err := s.storage.SaveRequest(request); err != nil {
			// 记录错误，但不中断请求处理
			fmt.Printf("无法保存请求: %v\n", err)
		}
	}
}

// GetConfig 返回服务器的配置
// 获取当前服务器使用的配置信息
// 返回:
//   - ServerConfig: 服务器配置
func (s *Server) GetConfig() api.ServerConfig {
	return s.config
}
