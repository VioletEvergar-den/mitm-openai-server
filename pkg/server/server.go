package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/llm-sec/mitm-openai-server/pkg/storage"
	"github.com/llm-sec/mitm-openai-server/pkg/utils"
)

// NewServer 创建一个新的服务器实例
func NewServer(storage storage.Storage) *Server {
	return NewServerWithConfig(ServerConfig{
		Storage: storage,
	})
}

// NewServerWithConfig 使用配置创建一个新的服务器实例
func NewServerWithConfig(config ServerConfig) *Server {
	// 如果启用生成UI认证凭证
	if config.GenerateUIAuth {
		if config.UIUsername == "" {
			config.UIUsername = "admin"
		}
		if config.UIPassword == "" {
			config.UIPassword = generateRandomPassword(12)
		}

		// 在控制台打印UI认证凭证
		fmt.Printf("\n==================================================\n")
		fmt.Printf("前端UI访问凭证:\n")
		fmt.Printf("用户名: %s\n", config.UIUsername)
		fmt.Printf("密码: %s\n", config.UIPassword)
		fmt.Printf("==================================================\n\n")
	}

	s := &Server{
		router:  gin.Default(),
		storage: config.Storage,
		config:  config,
	}
	s.setupRoutes()
	return s
}

// Run 启动服务器
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// 认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果认证被禁用，则跳过
		if !s.config.EnableAuth {
			c.Next()
			return
		}

		var authorized bool
		switch s.config.AuthType {
		case "basic":
			authorized = s.validateBasicAuth(c)
		case "token":
			authorized = s.validateTokenAuth(c)
		default:
			// 不支持的认证类型，拒绝访问
			c.JSON(http.StatusUnauthorized, StandardResponse{
				Status:  "error",
				Message: "认证配置错误",
			})
			c.Abort()
			return
		}

		if !authorized {
			c.JSON(http.StatusUnauthorized, StandardResponse{
				Status:  "error",
				Message: "认证失败",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 验证基本认证
func (s *Server) validateBasicAuth(c *gin.Context) bool {
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		return false
	}

	return username == s.config.Username && password == s.config.Password
}

// 验证令牌认证
func (s *Server) validateTokenAuth(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	// 支持"Bearer token"格式和直接的令牌格式
	token := authHeader
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = authHeader[7:]
	}

	return token == s.config.Token
}

// CORS中间件
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.config.EnableCORS {
			c.Writer.Header().Set("Access-Control-Allow-Origin", s.config.AllowOrigins)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

			// 处理预检请求
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}
		c.Next()
	}
}

// setupRoutes 设置所有路由
func (s *Server) setupRoutes() {
	// 添加CORS中间件
	s.router.Use(s.corsMiddleware())

	// 健康检查 - 不需要认证
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 前端静态文件 - 不需要API认证，但有自己的认证
	s.setupUIRoutes()

	// OpenAPI规范 - 不需要认证
	s.router.GET("/openapi.json", s.ServeOpenAPISpec)
	s.router.GET("/v1/openapi.json", s.ServeOpenAPISpec)
	s.router.GET("/v2/openapi.json", s.ServeOpenAPISpec)

	// API请求 - 需要认证
	apiGroup := s.router.Group("/")
	apiGroup.Use(s.authMiddleware()) // 应用认证中间件
	apiGroup.Use(s.apiMiddleware())  // 应用API中间件记录请求

	// 获取所有记录的请求
	apiGroup.GET("/api/requests", s.getAllRequests)

	// 获取单个请求
	apiGroup.GET("/api/requests/:id", s.getRequestByID)

	// 删除单个请求
	apiGroup.DELETE("/api/requests/:id", s.deleteRequest)

	// 删除所有请求
	apiGroup.DELETE("/api/requests", s.deleteAllRequests)

	// 导出请求为JSONL
	apiGroup.GET("/api/export", s.exportRequests)

	// 获取存储统计信息
	apiGroup.GET("/api/stats", s.getStorageStats)

	// v1 API组
	v1 := apiGroup.Group("/v1")
	{
		// 在这里添加你的v1 API路由
		v1.GET("/users", s.handleV1Users)
		v1.POST("/users", s.handleV1CreateUser)
		v1.GET("/users/:id", s.handleV1UserByID)
		v1.POST("/echo", s.handleV1Echo)
	}

	// v2 API组
	v2 := apiGroup.Group("/v2")
	{
		// 在这里添加你的v2 API路由
		v2.GET("/users", s.handleV2Users)
		v2.POST("/users", s.handleV2CreateUser)
		v2.GET("/users/:id", s.handleV2UserByID)
		v2.POST("/echo", s.handleV2Echo)
	}
}

// apiMiddleware API中间件，记录API调用
func (s *Server) apiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过OpenAPI规范和内部API请求
		path := c.Request.URL.Path
		if strings.HasSuffix(path, "/openapi.json") ||
			strings.HasPrefix(path, "/api/") ||
			path == "/health" {
			c.Next()
			return
		}

		// 创建请求记录
		req := &Request{
			ID:        uuid.New().String(),
			Timestamp: time.Now().Format(time.RFC3339),
			Method:    c.Request.Method,
			Path:      path,
			Headers:   make(map[string]string),
			Query:     make(map[string]string),
			IPAddress: c.ClientIP(),
		}

		// 获取请求头
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				req.Headers[k] = v[0]
			}
		}

		// 获取查询参数
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				req.Query[k] = v[0]
			}
		}

		// 获取请求体
		var bodyBytes []byte
		var body interface{}

		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，因为读取后Body会被消耗
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 尝试解析JSON
			if len(bodyBytes) > 0 {
				if err := json.Unmarshal(bodyBytes, &body); err == nil {
					req.Body = body
				}
			}
		}

		// 设置请求ID和时间戳
		c.Set("request_id", req.ID)
		c.Set("request_timestamp", req.Timestamp)

		// 在代理模式下，转发请求到目标API
		if s.config.ProxyMode && s.config.TargetURL != "" {
			// 调用代理函数
			proxyRespMap, err := utils.SendProxyRequest(
				c.Request.Method,
				s.config.TargetURL,
				path,
				req.Headers,
				bodyBytes,
				s.config.TargetAuthType,
				s.config.TargetUsername,
				s.config.TargetPassword,
				s.config.TargetToken,
			)

			if err != nil {
				// 代理请求失败，返回错误信息
				c.JSON(http.StatusBadGateway, StandardResponse{
					Status:  "error",
					Message: "代理请求失败: " + err.Error(),
				})

				// 记录失败的请求
				req.Response = &ProxyResponse{
					StatusCode: http.StatusBadGateway,
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body: map[string]interface{}{
						"status":  "error",
						"message": "代理请求失败: " + err.Error(),
					},
				}

				if err := s.saveRequest(req); err != nil {
					fmt.Printf("保存请求记录失败: %v\n", err)
				}

				c.Abort()
				return
			}

			// 转换并记录代理响应
			proxyResp := &ProxyResponse{
				StatusCode: proxyRespMap["status_code"].(int),
				Headers:    proxyRespMap["headers"].(map[string]string),
				Body:       proxyRespMap["body"],
			}
			req.Response = proxyResp

			// 保存请求记录
			if err := s.saveRequest(req); err != nil {
				fmt.Printf("保存请求记录失败: %v\n", err)
			}

			// 设置响应状态码
			c.Status(proxyResp.StatusCode)

			// 转发响应头
			for k, v := range proxyResp.Headers {
				// 避免设置一些特定的头，这些头由Gin框架处理
				if strings.ToLower(k) != "content-length" {
					c.Header(k, v)
				}
			}

			// 返回代理响应体
			if proxyResp.Body != nil {
				switch body := proxyResp.Body.(type) {
				case string:
					c.String(proxyResp.StatusCode, body)
				default:
					c.JSON(proxyResp.StatusCode, body)
				}
			}

			c.Abort() // 终止后续处理
			return
		}

		// 在非代理模式下，保存请求继续处理
		if err := s.saveRequest(req); err != nil {
			c.JSON(http.StatusInternalServerError, StandardResponse{
				Status:  "error",
				Message: "保存请求记录失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		c.Next() // 继续处理请求
	}
}

// 将请求记录保存到存储中
func (s *Server) saveRequest(req *Request) error {
	// 转换Request为storage.Request
	storageReq := &storage.Request{
		ID:        req.ID,
		Timestamp: req.Timestamp,
		Method:    req.Method,
		Path:      req.Path,
		Headers:   req.Headers,
		Query:     req.Query,
		Body:      req.Body,
		IPAddress: req.IPAddress,
	}

	if req.Response != nil {
		storageReq.Response = &storage.ProxyResponse{
			StatusCode: req.Response.StatusCode,
			Headers:    req.Response.Headers,
			Body:       req.Response.Body,
		}
	}

	return s.storage.SaveRequest(storageReq)
}

// 从storage.Request转换为server.Request
func convertStorageToServerRequest(req *storage.Request) *Request {
	serverReq := &Request{
		ID:        req.ID,
		Timestamp: req.Timestamp.(string),
		Method:    req.Method,
		Path:      req.Path,
		Headers:   req.Headers.(map[string]string),
		Query:     req.Query.(map[string]string),
		Body:      req.Body,
		IPAddress: req.IPAddress,
	}

	if req.Response != nil {
		serverReq.Response = &ProxyResponse{
			StatusCode: req.Response.StatusCode,
			Headers:    req.Response.Headers.(map[string]string),
			Body:       req.Response.Body,
		}
	}

	return serverReq
}

// getAllRequests 处理获取所有请求的API
func (s *Server) getAllRequests(c *gin.Context) {
	// 获取分页参数
	limit := 100 // 默认每页100条
	offset := 0  // 默认从0开始

	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	if limitParam != "" {
		fmt.Sscanf(limitParam, "%d", &limit)
	}
	if offsetParam != "" {
		fmt.Sscanf(offsetParam, "%d", &offset)
	}

	reqs, err := s.storage.GetAllRequests(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: "获取请求列表失败: " + err.Error(),
		})
		return
	}

	// 转换为server.Request列表
	serverReqs := make([]*Request, len(reqs))
	for i, req := range reqs {
		serverReqs[i] = convertStorageToServerRequest(req)
	}

	c.JSON(http.StatusOK, StandardResponse{
		Status: "success",
		Data:   serverReqs,
	})
}

// getRequestByID 处理根据ID获取请求的API
func (s *Server) getRequestByID(c *gin.Context) {
	id := c.Param("id")
	req, err := s.storage.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Status:  "error",
			Message: "请求不存在: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Status: "success",
		Data:   convertStorageToServerRequest(req),
	})
}

// deleteRequest 删除单个请求
func (s *Server) deleteRequest(c *gin.Context) {
	id := c.Param("id")
	err := s.storage.DeleteRequest(id)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Status:  "error",
			Message: "删除请求失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "请求已删除",
	})
}

// deleteAllRequests 删除所有请求
func (s *Server) deleteAllRequests(c *gin.Context) {
	err := s.storage.DeleteAllRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: "删除所有请求失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "所有请求已删除",
	})
}

// exportRequests 导出请求为JSONL
func (s *Server) exportRequests(c *gin.Context) {
	filePath, err := s.storage.ExportRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: "导出请求失败: " + err.Error(),
		})
		return
	}

	// 设置文件下载头
	c.Header("Content-Disposition", "attachment; filename=requests.jsonl")
	c.Header("Content-Type", "application/jsonl")
	c.File(filePath)
}

// getStorageStats 获取存储统计信息
func (s *Server) getStorageStats(c *gin.Context) {
	stats, err := s.storage.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: "获取存储统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Status: "success",
		Data:   stats,
	})
}

// handleV1Users 处理v1版本获取用户列表的API
func (s *Server) handleV1Users(c *gin.Context) {
	users := []map[string]interface{}{
		{"id": "1", "name": "用户1", "email": "user1@example.com"},
		{"id": "2", "name": "用户2", "email": "user2@example.com"},
		{"id": "3", "name": "用户3", "email": "user3@example.com"},
	}
	c.JSON(http.StatusOK, users)
}

// handleV1CreateUser 处理v1版本创建用户的API
func (s *Server) handleV1CreateUser(c *gin.Context) {
	var userData map[string]interface{}
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户数据"})
		return
	}

	// 添加ID
	userData["id"] = uuid.New().String()
	userData["created_at"] = time.Now().Format(time.RFC3339)

	c.JSON(http.StatusCreated, userData)
}

// handleV1UserByID 处理v1版本根据ID获取用户的API
func (s *Server) handleV1UserByID(c *gin.Context) {
	id := c.Param("id")
	user := map[string]interface{}{
		"id":         id,
		"name":       "用户" + id,
		"email":      "user" + id + "@example.com",
		"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, user)
}

// handleV1Echo 处理v1版本回显API
func (s *Server) handleV1Echo(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 添加时间戳和ID
	data["timestamp"] = time.Now().Format(time.RFC3339)
	data["request_id"] = uuid.New().String()

	c.JSON(http.StatusOK, data)
}

// handleV2Users 处理v2版本获取用户列表的API
func (s *Server) handleV2Users(c *gin.Context) {
	users := map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": "1", "name": "用户1", "email": "user1@example.com", "role": "admin"},
			{"id": "2", "name": "用户2", "email": "user2@example.com", "role": "user"},
			{"id": "3", "name": "用户3", "email": "user3@example.com", "role": "user"},
		},
		"total": 3,
	}
	c.JSON(http.StatusOK, users)
}

// handleV2CreateUser 处理v2版本创建用户的API
func (s *Server) handleV2CreateUser(c *gin.Context) {
	var userData map[string]interface{}
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":  "无效的用户数据",
			"detail": err.Error(),
		})
		return
	}

	// 添加ID
	userData["id"] = uuid.New().String()
	userData["created_at"] = time.Now().Format(time.RFC3339)

	result := map[string]interface{}{
		"data":    userData,
		"message": "用户创建成功",
	}
	c.JSON(http.StatusCreated, result)
}

// handleV2UserByID 处理v2版本根据ID获取用户的API
func (s *Server) handleV2UserByID(c *gin.Context) {
	id := c.Param("id")
	user := map[string]interface{}{
		"data": map[string]interface{}{
			"id":         id,
			"name":       "用户" + id,
			"email":      "user" + id + "@example.com",
			"role":       "user",
			"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		},
	}
	c.JSON(http.StatusOK, user)
}

// handleV2Echo 处理v2版本回显API
func (s *Server) handleV2Echo(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":  "无效的请求数据",
			"detail": err.Error(),
		})
		return
	}

	// 添加元数据
	metadata := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": uuid.New().String(),
		"version":    "v2",
	}

	result := map[string]interface{}{
		"data":     data,
		"metadata": metadata,
	}
	c.JSON(http.StatusOK, result)
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
