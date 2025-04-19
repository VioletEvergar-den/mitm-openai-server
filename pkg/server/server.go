package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/llm-sec/mitm-openai-server/pkg/embed"
	"github.com/llm-sec/mitm-openai-server/pkg/openai"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
	"github.com/llm-sec/mitm-openai-server/pkg/utils"
)

// NewServer 创建一个新的服务器实例
// 使用默认配置创建服务器，只需提供存储接口
// 参数:
//   - storage: 用于存储请求数据的存储接口
//
// 返回:
//   - *Server: 服务器实例
func NewServer(storage storage.Storage) *Server {
	return NewServerWithConfig(ServerConfig{
		Storage: storage,
	})
}

// NewServerWithConfig 使用配置创建一个新的服务器实例
// 允许通过ServerConfig结构体提供详细配置
// 参数:
//   - config: 服务器配置
//
// 返回:
//   - *Server: 配置好的服务器实例
func NewServerWithConfig(config ServerConfig) *Server {
	// 如果未设置存储，则返回错误
	if config.Storage == nil {
		log.Fatal("必须提供存储实例")
	}

	// 如果生成UI认证且没有设置密码，则生成随机密码
	if config.GenerateUIAuth && config.UIPassword == "" {
		config.UIPassword = generateRandomPassword(12)
	}

	// 创建默认路由
	router := gin.Default()

	// 创建服务器实例
	server := &Server{
		router:  router,
		storage: config.Storage,
		config:  config,
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

	// 设置路由
	server.setupRoutes()
	server.setupUIRoutes()
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

	// OpenAPI规范路由
	s.router.GET("/openai.json", s.ServeOpenAISpec)

	// API路由组，有认证中间件保护
	apiGroup := s.router.Group("/api")
	apiGroup.Use(s.authMiddleware())
	apiGroup.Use(s.apiMiddleware())

	// 请求管理路由
	apiGroup.GET("/requests", s.getRequests)
	apiGroup.GET("/requests/:id", s.getRequestByID)
	apiGroup.DELETE("/requests/:id", s.deleteRequest)
	apiGroup.DELETE("/requests", s.deleteAllRequests)
	apiGroup.GET("/export", s.exportRequests)
	apiGroup.GET("/stats", s.getStorageStats)

	// v1 API组
	v1 := apiGroup.Group("/v1")
	{
		v1.GET("/users", s.handleV1Users)
		v1.POST("/users", s.handleV1CreateUser)
		v1.GET("/users/:id", s.handleV1UserByID)
		v1.POST("/echo", s.handleV1Echo)
	}

	// v2 API组
	v2 := apiGroup.Group("/v2")
	{
		v2.GET("/users", s.handleV2Users)
		v2.POST("/users", s.handleV2CreateUser)
		v2.GET("/users/:id", s.handleV2UserByID)
		v2.POST("/echo", s.handleV2Echo)
	}
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

	// UI API路由组 - 需要认证
	uiAPI := s.router.Group("/ui/api")
	uiAPI.Use(s.authMiddleware())

	// UI 登录接口 - 登录接口不需要认证
	s.router.POST("/ui/api/login", s.handleUILogin)

	// 请求相关接口
	uiAPI.GET("/requests", s.getRequests)
	uiAPI.GET("/requests/:id", s.getRequestByID)
	uiAPI.DELETE("/requests/:id", s.deleteRequest)
	uiAPI.DELETE("/requests", s.deleteAllRequests)

	// 文件导出
	uiAPI.GET("/export", s.exportRequests)

	// 存储统计
	uiAPI.GET("/storage-stats", s.getStorageStats)

	// 服务器信息
	uiAPI.GET("/server-info", s.getServerInfo)

	// 代理配置
	uiAPI.GET("/proxy-config", s.getProxyConfig)
	uiAPI.POST("/proxy-config", s.saveProxyConfig)

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
	// 如果没有初始化openaiService，跳过设置OpenAI路由
	if s.openaiService == nil {
		return
	}

	// OpenAI API路由组
	openaiGroup := s.router.Group("/v1")

	// 应用API中间件记录请求
	openaiGroup.Use(s.apiMiddleware())

	// 处理所有请求
	openaiGroup.Any("/*path", s.handleOpenAIRequest)
}

// handleOpenAIRequest 处理OpenAI API请求
// 根据配置，将请求代理到真实API或返回模拟响应
// 参数:
//   - c: Gin上下文
func (s *Server) handleOpenAIRequest(c *gin.Context) {
	// 获取请求方法和路径
	method := c.Request.Method
	path := c.Param("path")

	// 读取请求体
	var bodyBytes []byte
	var err error
	if c.Request.Body != nil {
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]interface{}{
					"message": "无法读取请求体: " + err.Error(),
					"type":    "server_error",
					"code":    "internal_server_error",
				},
			})
			return
		}
		// 恢复请求体，以便其他中间件可以再次读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 收集请求头
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 收集查询参数
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// 使用OpenAI服务处理请求
	statusCode, responseHeaders, responseBody, err := s.openaiService.HandleRequest(
		method, path, headers, queryParams, bodyBytes,
	)

	if err != nil {
		// 发生错误，返回错误响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"message": "处理请求失败: " + err.Error(),
				"type":    "server_error",
				"code":    "internal_server_error",
			},
		})
		return
	}

	// 设置响应头
	for key, value := range responseHeaders {
		c.Header(key, value)
	}

	// 返回响应
	c.JSON(statusCode, responseBody)
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
				c.JSON(http.StatusInternalServerError, StandardResponse{
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

// saveRequest 保存请求到存储
// 将服务器请求模型转换为存储请求模型并保存
// 参数:
//   - req: 要保存的请求
//
// 返回:
//   - error: 如果保存失败，返回错误
func (s *Server) saveRequest(req *Request) error {
	// 转换服务器请求模型为存储请求模型
	storageReq := &storage.Request{
		ID:        req.ID,
		Method:    req.Method,
		Path:      req.Path,
		Timestamp: req.Timestamp,
		ClientIP:  req.IPAddress,
	}

	// 转换Headers（从map[string]string到map[string][]string）
	headers := make(map[string][]string)
	for k, v := range req.Headers {
		headers[k] = []string{v}
	}
	storageReq.Headers = headers

	// 转换Query（从map[string]string到map[string][]string）
	query := make(map[string][]string)
	for k, v := range req.Query {
		query[k] = []string{v}
	}
	storageReq.Query = query

	// 设置请求体
	storageReq.Body = req.Body

	// 如果有响应，也转换响应
	if req.Response != nil {
		// 转换响应头
		respHeaders := make(map[string][]string)
		for k, v := range req.Response.Headers {
			respHeaders[k] = []string{v}
		}

		storageReq.Response = &storage.ProxyResponse{
			StatusCode: req.Response.StatusCode,
			Headers:    respHeaders,
			Body:       req.Response.Body,
		}
	}

	// 保存到存储
	return s.storage.SaveRequest(storageReq)
}

// convertStorageToServerRequest 将存储请求模型转换为服务器请求模型
// 参数:
//   - req: 存储请求模型
//
// 返回:
//   - *Request: 服务器请求模型
func convertStorageToServerRequest(req *storage.Request) *Request {
	// 创建服务器请求模型
	serverReq := &Request{
		ID:     req.ID,
		Method: req.Method,
		Path:   req.Path,
		Body:   req.Body,
	}

	// 处理时间戳
	switch ts := req.Timestamp.(type) {
	case string:
		serverReq.Timestamp = ts
	case time.Time:
		serverReq.Timestamp = ts.Format(time.RFC3339)
	default:
		// 如果无法识别时间戳类型，使用当前时间
		serverReq.Timestamp = time.Now().Format(time.RFC3339)
	}

	// 处理IP地址（兼容两个字段）
	if req.ClientIP != "" {
		serverReq.IPAddress = req.ClientIP
	} else {
		serverReq.IPAddress = req.IPAddress
	}

	// 转换Headers
	serverReq.Headers = make(map[string]string)
	switch headers := req.Headers.(type) {
	case map[string][]string:
		for k, v := range headers {
			if len(v) > 0 {
				serverReq.Headers[k] = v[0]
			}
		}
	case map[string]string:
		serverReq.Headers = headers
	}

	// 转换Query
	serverReq.Query = make(map[string]string)
	switch query := req.Query.(type) {
	case map[string][]string:
		for k, v := range query {
			if len(v) > 0 {
				serverReq.Query[k] = v[0]
			}
		}
	case map[string]string:
		serverReq.Query = query
	}

	// 转换Response（如果存在）
	if req.Response != nil {
		respHeaders := make(map[string]string)

		// 转换响应头
		switch headers := req.Response.Headers.(type) {
		case map[string][]string:
			for k, v := range headers {
				if len(v) > 0 {
					respHeaders[k] = v[0]
				}
			}
		case map[string]string:
			respHeaders = headers
		}

		serverReq.Response = &ProxyResponse{
			StatusCode: req.Response.StatusCode,
			Headers:    respHeaders,
			Body:       req.Response.Body,
		}
	}

	return serverReq
}

// getPaginationParams 从请求中获取分页参数
func getPaginationParams(c *gin.Context) (page, size int) {
	page = 1  // 默认页码为1
	size = 20 // 默认每页20条

	// 解析查询参数
	pageParam := c.Query("page")
	sizeParam := c.Query("size")

	// 转换页码参数
	if pageParam != "" {
		if parsedPage, err := strconv.Atoi(pageParam); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	// 转换每页大小参数
	if sizeParam != "" {
		if parsedSize, err := strconv.Atoi(sizeParam); err == nil && parsedSize > 0 {
			size = parsedSize
		}
	}

	return page, size
}

// getRequests 获取请求列表
// 参数:
//   - c: Gin上下文
func (s *Server) getRequests(c *gin.Context) {
	// 处理分页参数
	page, size := getPaginationParams(c)
	limit := size
	offset := (page - 1) * size

	// 从存储中获取请求列表
	requests, err := s.storage.GetAllRequests(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10012,
			Msg:  "获取请求列表失败: " + err.Error(),
		})
		return
	}

	// 获取总请求数
	// 尝试使用s.storage.GetRequestsCount()，如果不存在则返回请求列表的长度
	var total int64
	total = int64(len(requests))

	// 将存储请求转换为服务器请求
	serverRequests := make([]*Request, len(requests))
	for i, req := range requests {
		serverRequests[i] = convertStorageToServerRequest(req)
	}

	// 返回请求列表和分页信息
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "获取请求列表成功",
		Data: map[string]interface{}{
			"requests": serverRequests,
			"total":    total,
			"page":     page,
			"size":     size,
		},
	})
}

// getRequestByID 获取特定ID的请求
// 处理GET /api/requests/:id请求
// 参数:
//   - c: Gin上下文
func (s *Server) getRequestByID(c *gin.Context) {
	// 获取请求ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10003,
			Msg:  "请求ID不能为空",
		})
		return
	}

	// 从存储中获取请求
	req, err := s.storage.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Code: 10004,
			Msg:  "请求不存在: " + err.Error(),
		})
		return
	}

	// 转换为服务器请求模型
	serverReq := convertStorageToServerRequest(req)

	// 返回请求详情
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "请求获取成功",
		Data: serverReq,
	})
}

// deleteRequest 删除特定ID的请求
// 处理DELETE /api/requests/:id请求
// 参数:
//   - c: Gin上下文
func (s *Server) deleteRequest(c *gin.Context) {
	// 获取请求ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10006,
			Msg:  "请求ID不能为空",
		})
		return
	}

	// 从存储中删除请求
	err := s.storage.DeleteRequest(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "不存在") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, StandardResponse{
			Code: 10007,
			Msg:  "删除请求失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "请求已成功删除",
	})
}

// deleteAllRequests 删除所有请求
// 处理DELETE /api/requests请求
// 参数:
//   - c: Gin上下文
func (s *Server) deleteAllRequests(c *gin.Context) {
	// 从存储中删除所有请求
	err := s.storage.DeleteAllRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10008,
			Msg:  "删除所有请求失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "所有请求已成功删除",
	})
}

// exportRequests 导出请求为JSONL格式
// 处理GET /api/export请求
// 参数:
//   - c: Gin上下文
func (s *Server) exportRequests(c *gin.Context) {
	// 从存储中导出请求
	filePath, err := s.storage.ExportRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10009,
			Msg:  "导出请求失败: " + err.Error(),
		})
		return
	}

	// 返回导出文件路径
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "请求已成功导出",
		Data: map[string]string{
			"file_path": filePath,
		},
	})
}

// getStorageStats 获取存储统计信息
// 处理GET /api/stats请求
// 参数:
//   - c: Gin上下文
func (s *Server) getStorageStats(c *gin.Context) {
	// 获取存储统计信息
	stats, err := s.storage.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10010,
			Msg:  "获取存储统计信息失败: " + err.Error(),
		})
		return
	}

	// 返回统计信息
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "获取存储统计信息成功",
		Data: stats,
	})
}

// handleV1Users 处理V1版本用户列表请求
// 处理GET /v1/users请求
// 参数:
//   - c: Gin上下文
func (s *Server) handleV1Users(c *gin.Context) {
	// 创建模拟用户数据
	users := []map[string]interface{}{
		{
			"id":       "1",
			"name":     "张三",
			"email":    "zhangsan@example.com",
			"created":  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"active":   true,
			"role":     "admin",
			"location": "北京",
		},
		{
			"id":       "2",
			"name":     "李四",
			"email":    "lisi@example.com",
			"created":  time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			"active":   true,
			"role":     "user",
			"location": "上海",
		},
		{
			"id":       "3",
			"name":     "王五",
			"email":    "wangwu@example.com",
			"created":  time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
			"active":   false,
			"role":     "user",
			"location": "广州",
		},
	}

	// 返回用户列表
	c.JSON(http.StatusOK, users)
}

// handleV1CreateUser 处理V1版本创建用户请求
// 处理POST /v1/users请求
// 参数:
//   - c: Gin上下文
func (s *Server) handleV1CreateUser(c *gin.Context) {
	// 解析请求体
	var user map[string]interface{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "无效的请求体",
			"message": err.Error(),
		})
		return
	}

	// 添加创建时间和ID
	user["id"] = uuid.New().String()
	user["created"] = time.Now().Format(time.RFC3339)

	// 返回创建的用户
	c.JSON(http.StatusCreated, user)
}

// handleV1UserByID 处理V1版本获取单个用户请求
// 处理GET /v1/users/:id请求
// 参数:
//   - c: Gin上下文
func (s *Server) handleV1UserByID(c *gin.Context) {
	// 获取用户ID
	id := c.Param("id")

	// 根据ID返回不同的用户数据
	switch id {
	case "1":
		c.JSON(http.StatusOK, map[string]interface{}{
			"id":       "1",
			"name":     "张三",
			"email":    "zhangsan@example.com",
			"created":  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"active":   true,
			"role":     "admin",
			"location": "北京",
			"profile": map[string]interface{}{
				"avatar":      "https://example.com/avatars/1.jpg",
				"phone":       "13800138001",
				"description": "系统管理员",
			},
		})
	case "2":
		c.JSON(http.StatusOK, map[string]interface{}{
			"id":       "2",
			"name":     "李四",
			"email":    "lisi@example.com",
			"created":  time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			"active":   true,
			"role":     "user",
			"location": "上海",
			"profile": map[string]interface{}{
				"avatar":      "https://example.com/avatars/2.jpg",
				"phone":       "13800138002",
				"description": "普通用户",
			},
		})
	case "3":
		c.JSON(http.StatusOK, map[string]interface{}{
			"id":       "3",
			"name":     "王五",
			"email":    "wangwu@example.com",
			"created":  time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
			"active":   false,
			"role":     "user",
			"location": "广州",
			"profile": map[string]interface{}{
				"avatar":      "https://example.com/avatars/3.jpg",
				"phone":       "13800138003",
				"description": "已禁用用户",
			},
		})
	default:
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "用户不存在",
			"message": fmt.Sprintf("ID为%s的用户不存在", id),
		})
	}
}

// handleV1Echo 处理回显请求
// 这是一个简单的测试端点，将请求数据回显给调用方
// 参数:
//   - c: Gin上下文
func (s *Server) handleV1Echo(c *gin.Context) {
	// 从请求中获取数据
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10011,
			Msg:  "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 回显数据
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "回显成功",
		Data: data,
	})
}

// handleV2Users 处理V2版本用户列表请求
// 处理GET /v2/users请求
// 参数:
//   - c: Gin上下文
func (s *Server) handleV2Users(c *gin.Context) {
	// 创建模拟用户数据
	users := []map[string]interface{}{
		{
			"id":      "1",
			"name":    "张三",
			"email":   "zhangsan@example.com",
			"created": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"active":  true,
			"role":    "admin",
		},
		{
			"id":      "2",
			"name":    "李四",
			"email":   "lisi@example.com",
			"created": time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			"active":  true,
			"role":    "user",
		},
		{
			"id":      "3",
			"name":    "王五",
			"email":   "wangwu@example.com",
			"created": time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
			"active":  false,
			"role":    "user",
		},
	}

	// 返回标准格式的响应
	c.JSON(http.StatusOK, map[string]interface{}{
		"data":     users,
		"total":    len(users),
		"page":     1,
		"per_page": len(users),
		"metadata": map[string]interface{}{
			"version":   "v2",
			"timestamp": time.Now().Format(time.RFC3339),
			"api_name":  "users",
		},
	})
}

// handleV2CreateUser 处理V2版本创建用户请求
// 处理POST /v2/users请求
// 参数:
//   - c: Gin上下文
func (s *Server) handleV2CreateUser(c *gin.Context) {
	// 解析请求体
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "BAD_REQUEST",
				"message": "无效的请求体: " + err.Error(),
			},
			"metadata": map[string]interface{}{
				"version":   "v2",
				"timestamp": time.Now().Format(time.RFC3339),
				"api_name":  "users",
			},
		})
		return
	}

	// 获取用户数据
	var userData map[string]interface{}
	if data, ok := requestData["data"].(map[string]interface{}); ok {
		userData = data
	} else {
		userData = requestData // 兼容直接传递用户数据的情况
	}

	// 添加ID和创建时间
	userData["id"] = uuid.New().String()
	userData["created"] = time.Now().Format(time.RFC3339)

	// 返回标准格式的响应
	c.JSON(http.StatusCreated, map[string]interface{}{
		"data": userData,
		"metadata": map[string]interface{}{
			"version":   "v2",
			"timestamp": time.Now().Format(time.RFC3339),
			"api_name":  "users",
			"action":    "create",
		},
	})
}

// handleV2UserByID 处理V2版本获取单个用户请求
// 处理GET /v2/users/:id请求
// 参数:
//   - c: Gin上下文
func (s *Server) handleV2UserByID(c *gin.Context) {
	// 获取用户ID
	id := c.Param("id")

	// 根据ID查找对应的用户数据
	var userData map[string]interface{}
	var exists bool

	switch id {
	case "1":
		userData = map[string]interface{}{
			"id":      "1",
			"name":    "张三",
			"email":   "zhangsan@example.com",
			"created": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"active":  true,
			"role":    "admin",
			"profile": map[string]interface{}{
				"avatar": "https://example.com/avatars/1.jpg",
				"phone":  "13800138001",
			},
		}
		exists = true
	case "2":
		userData = map[string]interface{}{
			"id":      "2",
			"name":    "李四",
			"email":   "lisi@example.com",
			"created": time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			"active":  true,
			"role":    "user",
			"profile": map[string]interface{}{
				"avatar": "https://example.com/avatars/2.jpg",
				"phone":  "13800138002",
			},
		}
		exists = true
	case "3":
		userData = map[string]interface{}{
			"id":      "3",
			"name":    "王五",
			"email":   "wangwu@example.com",
			"created": time.Now().Add(-72 * time.Hour).Format(time.RFC3339),
			"active":  false,
			"role":    "user",
			"profile": map[string]interface{}{
				"avatar": "https://example.com/avatars/3.jpg",
				"phone":  "13800138003",
			},
		}
		exists = true
	default:
		exists = false
	}

	if exists {
		// 返回标准格式的响应
		c.JSON(http.StatusOK, map[string]interface{}{
			"data": userData,
			"metadata": map[string]interface{}{
				"version":   "v2",
				"timestamp": time.Now().Format(time.RFC3339),
				"api_name":  "users",
				"action":    "get",
			},
		})
	} else {
		// 返回标准格式的错误响应
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "USER_NOT_FOUND",
				"message": fmt.Sprintf("ID为%s的用户不存在", id),
			},
			"metadata": map[string]interface{}{
				"version":   "v2",
				"timestamp": time.Now().Format(time.RFC3339),
				"api_name":  "users",
				"action":    "get",
			},
		})
	}
}

// handleV2Echo 处理V2版本Echo请求
// 处理POST /v2/echo请求，返回请求信息的标准格式
// 参数:
//   - c: Gin上下文
func (s *Server) handleV2Echo(c *gin.Context) {
	// 解析请求体
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		requestBody = map[string]interface{}{} // 使用空对象
	}

	// 提取message字段
	var message interface{}
	if msg, exists := requestBody["message"]; exists {
		message = msg
	} else {
		message = "No message provided"
	}

	// 构建数据部分
	data := map[string]interface{}{
		"message":    message,
		"request_id": uuid.New().String(),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.Query(),
		"body":       requestBody,
	}

	// 构建元数据部分
	metadata := map[string]interface{}{
		"version":    "v2",
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": uuid.New().String(),
		"api_name":   "echo",
	}

	// 返回标准格式的响应
	c.JSON(http.StatusOK, map[string]interface{}{
		"data":     data,
		"metadata": metadata,
	})
}

// generateRandomPassword 生成随机密码
// 生成指定长度的随机密码，包含字母、数字和特殊字符
// 参数:
//   - length: 密码长度
//
// 返回:
//   - string: 生成的随机密码
func generateRandomPassword(length int) string {
	if length < 8 {
		length = 8 // 确保最小长度为8
	}

	// 字符集
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"

	// 随机种子
	rand.Seed(time.Now().UnixNano())

	// 生成密码
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		password[i] = chars[rand.Intn(len(chars))]
	}

	return string(password)
}

// GetConfig 返回服务器的配置
// 获取当前服务器使用的配置信息
// 返回:
//   - ServerConfig: 服务器配置
func (s *Server) GetConfig() ServerConfig {
	return s.config
}

// handleUILogin 处理UI登录请求
// 验证前端登录表单提交的用户名和密码
// 参数:
//   - c: Gin上下文
func (s *Server) handleUILogin(c *gin.Context) {
	// 定义请求体结构
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// 解析请求体
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "无效的请求格式",
		})
		return
	}

	// 验证用户名和密码
	if loginReq.Username == s.config.UIUsername && loginReq.Password == s.config.UIPassword {
		// 登录成功，生成一个简单的会话令牌
		// 实际应用中应使用更安全的会话管理
		token := fmt.Sprintf("%s_%d", uuid.New().String(), time.Now().Unix())

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "登录成功",
			"token":   token,
		})
		return
	}

	// 登录失败
	c.JSON(http.StatusUnauthorized, gin.H{
		"status":  "error",
		"message": "用户名或密码错误",
	})
}

// getServerInfo 返回服务器信息
// 返回服务器的基本配置信息，包括版本、API路径和认证设置
// 参数:
//   - c: Gin上下文
func (s *Server) getServerInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":     "1.0.0",
		"apiBasePath": "/",
		"openApiPath": "/openai.json",
		"auth": gin.H{
			"enabled": s.config.EnableAuth,
			"type":    s.config.AuthType,
		},
		"proxy": gin.H{
			"enabled":   s.config.ProxyMode,
			"targetURL": s.config.TargetURL,
		},
	})
}

// getProxyConfig 获取代理配置
// 参数:
//   - c: Gin上下文
func (s *Server) getProxyConfig(c *gin.Context) {
	// 构建代理配置响应
	config := map[string]interface{}{
		"enabled":       s.config.ProxyMode,
		"target_url":    s.config.TargetURL,
		"auth_type":     s.config.TargetAuthType,
		"target_auth":   s.config.TargetAuthType != "none",
		"target_user":   s.config.TargetUsername,
		"target_passwd": "******", // 密码隐藏
		"target_token":  "******", // 令牌隐藏
	}

	// 返回配置
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "代理配置获取成功",
		Data: config,
	})
}

// saveProxyConfig 保存代理配置
// 更新服务器的代理配置
// 参数:
//   - c: Gin上下文
func (s *Server) saveProxyConfig(c *gin.Context) {
	var config struct {
		Enabled      bool   `json:"enabled"`
		TargetURL    string `json:"targetURL"`
		AuthType     string `json:"authType"`
		TargetUser   string `json:"targetUser"`
		TargetPasswd string `json:"targetPasswd"`
		TargetToken  string `json:"targetToken"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "无效的请求格式",
		})
		return
	}

	// 更新配置
	s.config.ProxyMode = config.Enabled
	s.config.TargetURL = config.TargetURL
	s.config.TargetAuthType = config.AuthType

	// 只有在提供了新值时才更新凭据
	if config.TargetUser != "" {
		s.config.TargetUsername = config.TargetUser
	}
	if config.TargetPasswd != "" && config.TargetPasswd != strings.Repeat("*", len(s.config.TargetPassword)) {
		s.config.TargetPassword = config.TargetPasswd
	}
	if config.TargetToken != "" && config.TargetToken != strings.Repeat("*", len(s.config.TargetToken)) {
		s.config.TargetToken = config.TargetToken
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "代理配置已更新",
	})
}

// prepareReplayRequest 准备用于重播的HTTP请求
func prepareReplayRequest(storedReq *storage.Request) (*http.Request, error) {
	// 创建URL
	// 从storedReq.Path构建URL，storedReq.URL不可用
	reqURL := storedReq.Path
	if !strings.HasPrefix(reqURL, "http") {
		reqURL = "https://" + reqURL
	}

	// 从Body创建请求体，Body可能是接口类型，需要进行类型转换
	var bodyReader io.Reader
	if storedReq.Body != nil {
		switch body := storedReq.Body.(type) {
		case string:
			bodyReader = strings.NewReader(body)
		case []byte:
			bodyReader = bytes.NewReader(body)
		default:
			// 如果是其他类型，尝试将其转换为JSON字符串
			bodyBytes, err := json.Marshal(storedReq.Body)
			if err != nil {
				return nil, fmt.Errorf("无法序列化请求体: %v", err)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}
	}

	// 创建请求
	req, err := http.NewRequest(storedReq.Method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// 复制原始请求的头部，需要先进行类型断言
	switch headers := storedReq.Headers.(type) {
	case map[string][]string:
		for key, values := range headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	case map[string]string:
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}

// replayRequest 重播请求
// 参数:
//   - c: Gin上下文
func (s *Server) replayRequest(c *gin.Context) {
	// 提取请求ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10013,
			Msg:  "请求ID不能为空",
		})
		return
	}

	// 从存储中获取请求
	storedReq, err := s.storage.GetRequestByID(id)
	if err != nil {
		if err.Error() == "request not found" {
			c.JSON(http.StatusNotFound, StandardResponse{
				Code: 10014,
				Msg:  "请求不存在: " + id,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10015,
			Msg:  "获取请求失败: " + err.Error(),
		})
		return
	}

	// 准备重播请求
	replayReq, err := prepareReplayRequest(storedReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10016,
			Msg:  "准备重播请求失败: " + err.Error(),
		})
		return
	}

	// 发送重播请求
	resp, err := httpClient.Do(replayReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10017,
			Msg:  "发送重播请求失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10018,
			Msg:  "读取响应失败: " + err.Error(),
		})
		return
	}

	// 返回重播结果
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "重播请求成功",
		Data: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     resp.Header,
			"body":        string(respBody),
		},
	})
}

// 添加必要的变量
var (
	// httpClient 是用于发送HTTP请求的客户端
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)
