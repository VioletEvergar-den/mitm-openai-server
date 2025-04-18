package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/llm-sec/mitm-openapi-server/pkg/utils"
)

// ServerConfig 存储服务器配置
type ServerConfig struct {
	Storage        Storage
	EnableAuth     bool
	AuthType       string
	Username       string
	Password       string
	Token          string
	EnableCORS     bool
	AllowOrigins   string
	UIUsername     string // 前端UI用户名
	UIPassword     string // 前端UI密码
	GenerateUIAuth bool   // 是否生成随机UI认证凭证

	// 中间人代理相关配置
	ProxyMode      bool   // 是否启用代理模式
	TargetURL      string // 目标OpenAPI服务地址
	TargetAuthType string // 目标API认证类型：none, basic, token
	TargetUsername string // 目标API基本认证用户名
	TargetPassword string // 目标API基本认证密码
	TargetToken    string // 目标API令牌
}

// Server 表示API服务器
type Server struct {
	router  *gin.Engine
	storage Storage
	config  ServerConfig
}

// NewServer 创建一个新的服务器实例
func NewServer(storage Storage) *Server {
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

// Run 启动服务器
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
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

				if err := s.storage.SaveRequest(req); err != nil {
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
			if err := s.storage.SaveRequest(req); err != nil {
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
		if err := s.storage.SaveRequest(req); err != nil {
			c.JSON(http.StatusInternalServerError, StandardResponse{
				Status:  "error",
				Message: "保存请求记录失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 继续处理请求（返回模拟响应）
		c.Next()
	}
}

// getAllRequests 处理获取所有请求的API
func (s *Server) getAllRequests(c *gin.Context) {
	requests, err := s.storage.GetAllRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: "获取请求记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "请求记录获取成功",
		Data:    requests,
	})
}

// getRequestByID 处理根据ID获取请求的API
func (s *Server) getRequestByID(c *gin.Context) {
	id := c.Param("id")
	request, err := s.storage.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Status:  "error",
			Message: "请求记录不存在",
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "请求记录获取成功",
		Data:    request,
	})
}

// handleV1Users 处理获取V1用户列表请求
func (s *Server) handleV1Users(c *gin.Context) {
	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "用户列表获取成功",
		Data: []map[string]interface{}{
			{"id": "1", "name": "用户1"},
			{"id": "2", "name": "用户2"},
			{"id": "3", "name": "用户3"},
		},
	})
}

// handleV1CreateUser 处理创建V1用户请求
func (s *Server) handleV1CreateUser(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	timestamp, _ := c.Get("request_timestamp")

	c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "用户创建成功",
		Data: gin.H{
			"request_id": requestID,
			"timestamp":  timestamp,
			"user_id":    "new-user-id",
		},
	})
}

// handleV1UserByID 处理获取V1用户详情请求
func (s *Server) handleV1UserByID(c *gin.Context) {
	id := c.Param("id")

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "用户获取成功",
		Data: gin.H{
			"id":       id,
			"name":     "测试用户",
			"email":    "test@example.com",
			"is_admin": false,
		},
	})
}

// handleV1Echo 处理V1回显请求
func (s *Server) handleV1Echo(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	timestamp, _ := c.Get("request_timestamp")

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "请求已记录",
		Data: gin.H{
			"request_id": requestID,
			"timestamp":  timestamp,
		},
	})
}

// V2版本API处理程序
func (s *Server) handleV2Users(c *gin.Context) {
	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "V2用户列表获取成功",
		Data: []map[string]interface{}{
			{"id": "1", "name": "V2用户1", "role": "admin"},
			{"id": "2", "name": "V2用户2", "role": "user"},
			{"id": "3", "name": "V2用户3", "role": "user"},
		},
	})
}

func (s *Server) handleV2CreateUser(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	timestamp, _ := c.Get("request_timestamp")

	c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "V2用户创建成功",
		Data: gin.H{
			"request_id": requestID,
			"timestamp":  timestamp,
			"user_id":    "v2-new-user-id",
			"version":    "v2",
		},
	})
}

func (s *Server) handleV2UserByID(c *gin.Context) {
	id := c.Param("id")

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "V2用户获取成功",
		Data: gin.H{
			"id":            id,
			"name":          "V2测试用户",
			"email":         "v2test@example.com",
			"is_admin":      false,
			"account_level": "premium",
			"version":       "v2",
		},
	})
}

func (s *Server) handleV2Echo(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	timestamp, _ := c.Get("request_timestamp")

	c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "V2请求已记录",
		Data: gin.H{
			"request_id": requestID,
			"timestamp":  timestamp,
			"version":    "v2",
		},
	})
}

// 生成随机密码的字符集
const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

// 生成随机密码
func generateRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		password[i] = passwordChars[rand.Intn(len(passwordChars))]
	}
	return string(password)
}

// 设置前端UI路由
func (s *Server) setupUIRoutes() {
	// UI API认证中间件
	uiAuth := func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok || username != s.config.UIUsername || password != s.config.UIPassword {
			c.Header("WWW-Authenticate", "Basic realm=fake-openapi-server")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}

	// 首页重定向到UI
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/ui/")
	})

	// 前端API路由
	ui := s.router.Group("/ui/api")
	ui.Use(uiAuth)
	{
		// 获取所有请求
		ui.GET("/requests", func(c *gin.Context) {
			requests, err := s.storage.GetAllRequests()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, requests)
		})

		// 获取单个请求详情
		ui.GET("/requests/:id", func(c *gin.Context) {
			id := c.Param("id")
			request, err := s.storage.GetRequestByID(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "请求不存在"})
				return
			}
			c.JSON(http.StatusOK, request)
		})

		// 删除单个请求
		ui.DELETE("/requests/:id", func(c *gin.Context) {
			id := c.Param("id")
			if err := s.storage.DeleteRequest(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// 获取服务器信息
		ui.GET("/server-info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version":     "1.0.0",
				"apiBasePath": "/",
				"openApiPath": "/openapi.json",
				"auth": gin.H{
					"enabled": s.config.EnableAuth,
					"type":    s.config.AuthType,
				},
				"proxy": gin.H{
					"enabled":   s.config.ProxyMode,
					"targetURL": s.config.TargetURL,
				},
			})
		})

		// 获取数据库统计信息
		ui.GET("/storage-stats", func(c *gin.Context) {
			stats, err := s.storage.GetDatabaseStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, stats)
		})

		// 清空所有请求
		ui.DELETE("/requests", func(c *gin.Context) {
			if err := s.storage.DeleteAllRequests(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// 导出请求为JSONL
		ui.GET("/export", func(c *gin.Context) {
			exportPath, err := s.storage.ExportRequestsAsJSONL()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success":      true,
				"path":         exportPath,
				"download_url": "/ui/api/download?path=" + url.QueryEscape(exportPath),
			})
		})

		// 下载导出的文件
		ui.GET("/download", func(c *gin.Context) {
			path := c.Query("path")
			if path == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件路径参数"})
				return
			}

			// 安全检查，确保路径以exports目录开头
			if !strings.Contains(path, "exports") {
				c.JSON(http.StatusForbidden, gin.H{"error": "无效的文件路径"})
				return
			}

			// 提供文件下载
			filename := filepath.Base(path)
			c.Header("Content-Disposition", "attachment; filename="+filename)
			c.Header("Content-Type", "application/octet-stream")
			c.File(path)
		})

		// 获取代理配置
		ui.GET("/proxy-config", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"enabled":   s.config.ProxyMode,
				"targetURL": s.config.TargetURL,
				"authType":  s.config.TargetAuthType,
				"username":  s.config.TargetUsername,
				"password":  s.config.TargetPassword != "", // 只返回是否设置了密码，不返回实际密码
				"tokenSet":  s.config.TargetToken != "",    // 只返回是否设置了令牌，不返回实际令牌
			})
		})

		// 更新代理配置
		ui.POST("/proxy-config", func(c *gin.Context) {
			var config struct {
				Enabled    bool   `json:"enabled"`
				TargetURL  string `json:"targetURL"`
				AuthType   string `json:"authType"`
				Username   string `json:"username"`
				Password   string `json:"password"`
				Token      string `json:"token"`
				UpdateAuth bool   `json:"updateAuth"`
			}

			if err := c.ShouldBindJSON(&config); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置数据"})
				return
			}

			// 更新配置
			s.config.ProxyMode = config.Enabled

			if config.TargetURL != "" {
				s.config.TargetURL = config.TargetURL
			}

			if config.AuthType != "" {
				s.config.TargetAuthType = config.AuthType
			}

			// 如果需要更新认证信息
			if config.UpdateAuth {
				s.config.TargetUsername = config.Username

				if config.Password != "" {
					s.config.TargetPassword = config.Password
				}

				if config.Token != "" {
					s.config.TargetToken = config.Token
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "代理配置已更新",
			})
		})
	}

	// 前端静态文件 - 注意：必须在特定路由之后注册，避免通配符路由冲突
	s.router.Static("/ui", "./ui")
}
