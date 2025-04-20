package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
)

// UIServer 表示UI API服务器
type UIServer struct {
	storage       storage.Storage
	config        ServerConfig
	openaiService OpenAIServiceInterface
	// 添加令牌验证存储
	tokenStore sync.Map
}

// NewUIServer 创建一个新的UI服务器实例
func NewUIServer(storage storage.Storage, config ServerConfig, openaiService OpenAIServiceInterface) *UIServer {
	return &UIServer{
		storage:       storage,
		config:        config,
		openaiService: openaiService,
	}
}

// SetupUIRoutes 设置UI相关的路由
func (s *UIServer) SetupUIRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc, apiMiddleware gin.HandlerFunc) {
	// 确保UI路径有效
	uiPath := s.config.UIDir
	if uiPath == "" {
		uiPath = "./ui" // 默认UI目录
	}

	// UI认证中间件
	uiAuthMiddleware := func(c *gin.Context) {
		// 检查Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, StandardResponse{
				Code: 10002,
				Msg:  "认证失败",
			})
			c.Abort()
			return
		}

		// 支持"Bearer token"格式和直接的令牌格式
		token := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = authHeader[7:] // 去掉"Bearer "前缀
		}

		// 验证令牌
		_, found := s.tokenStore.Load(token)
		if !found {
			c.JSON(http.StatusUnauthorized, StandardResponse{
				Code: 10002,
				Msg:  "认证失败",
			})
			c.Abort()
			return
		}

		c.Next()
	}

	// UI API路由组 - 需要UI认证
	uiAPI := router.Group("/ui/api")

	// UI 登录接口 - 登录接口不需要认证
	uiAPI.POST("/login", s.HandleUILogin)

	// 添加UI认证中间件
	uiAPI.Use(uiAuthMiddleware)

	// 请求相关接口
	uiAPI.GET("/requests", s.GetRequests)
	uiAPI.GET("/requests/:id", s.GetRequestByID)
	uiAPI.DELETE("/requests/:id", s.DeleteRequest)
	uiAPI.DELETE("/requests", s.DeleteAllRequests)

	// 文件导出
	uiAPI.GET("/export", s.ExportRequests)

	// 存储统计
	uiAPI.GET("/storage-stats", s.GetStorageStats)

	// 服务器信息
	uiAPI.GET("/server-info", s.GetServerInfo)

	// 代理配置
	uiAPI.GET("/proxy-config", s.GetProxyConfig)
	uiAPI.POST("/proxy-config", s.SaveProxyConfig)

	// 添加API Token获取接口
	uiAPI.GET("/token", s.GetAPIToken)

	// 添加聊天测试接口
	uiAPI.POST("/chat", s.HandleUIChat)
}

// HandleUILogin 处理UI登录请求
func (s *UIServer) HandleUILogin(c *gin.Context) {
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

	// 去除用户名和密码中可能的前后空格
	loginReq.Username = strings.TrimSpace(loginReq.Username)
	loginReq.Password = strings.TrimSpace(loginReq.Password)
	configUsername := strings.TrimSpace(s.config.UIUsername)
	configPassword := strings.TrimSpace(s.config.UIPassword)

	// 添加更详细的调试信息
	fmt.Printf("\n=================== 登录请求详细信息 ===================\n")
	fmt.Printf("用户名: %s\n", loginReq.Username)
	fmt.Printf("收到的密码: %s (长度:%d)\n", loginReq.Password, len(loginReq.Password))
	fmt.Printf("配置中的密码: %s (长度:%d)\n", configPassword, len(configPassword))

	if len(loginReq.Password) > 0 && len(configPassword) > 0 {
		fmt.Printf("\n密码字符比较:\n")
		if loginReq.Password != configPassword {
			// 详细比较密码差异
			if len(loginReq.Password) > len(configPassword) {
				fmt.Printf("收到的密码多余字符:\n")
				for i := 0; i < len(loginReq.Password); i++ {
					if i < len(configPassword) {
						if loginReq.Password[i] != configPassword[i] {
							fmt.Printf("位置 %d: '%c'(ASCII:%d) vs '%c'(ASCII:%d)\n",
								i, loginReq.Password[i], loginReq.Password[i],
								configPassword[i], configPassword[i])
						}
					} else {
						fmt.Printf("位置 %d: '%c'(ASCII:%d)\n",
							i, loginReq.Password[i], loginReq.Password[i])
					}
				}
			} else if len(configPassword) > len(loginReq.Password) {
				fmt.Printf("配置密码多余字符:\n")
				for i := 0; i < len(configPassword); i++ {
					if i < len(loginReq.Password) {
						if loginReq.Password[i] != configPassword[i] {
							fmt.Printf("位置 %d: '%c'(ASCII:%d) vs '%c'(ASCII:%d)\n",
								i, loginReq.Password[i], loginReq.Password[i],
								configPassword[i], configPassword[i])
						}
					} else {
						fmt.Printf("位置 %d: '%c'(ASCII:%d)\n",
							i, configPassword[i], configPassword[i])
					}
				}
			} else {
				fmt.Printf("密码长度相同但内容不同:\n")
				for i := 0; i < len(loginReq.Password); i++ {
					if loginReq.Password[i] != configPassword[i] {
						fmt.Printf("位置 %d: '%c'(ASCII:%d) vs '%c'(ASCII:%d)\n",
							i, loginReq.Password[i], loginReq.Password[i],
							configPassword[i], configPassword[i])
					}
				}
			}
		} else {
			fmt.Printf("密码完全匹配\n")
		}
	}

	fmt.Printf("=================== 结束调试信息 ===================\n\n")

	// 验证用户名和密码，只使用配置中的凭据
	if loginReq.Username == configUsername && loginReq.Password == configPassword {
		// 登录成功，生成令牌
		token := fmt.Sprintf("%s_%d", uuid.New().String(), time.Now().Unix())
		fmt.Println("登录成功!")

		// 将令牌存储到内存中
		s.tokenStore.Store(token, loginReq.Username)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "登录成功",
			"token":   token,
		})
		return
	}

	// 登录失败
	fmt.Println("登录失败: 用户名或密码不匹配")

	c.JSON(http.StatusUnauthorized, gin.H{
		"status":  "error",
		"message": "用户名或密码错误",
	})
}

// ValidateToken 验证令牌是否有效
// 此方法供authMiddleware调用
func (s *UIServer) ValidateToken(token string) bool {
	_, found := s.tokenStore.Load(token)
	return found
}

// GetServerInfo 返回服务器信息
func (s *UIServer) GetServerInfo(c *gin.Context) {
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

// GetProxyConfig 获取代理配置
func (s *UIServer) GetProxyConfig(c *gin.Context) {
	// 从配置中获取存储路径
	storagePath := getStoragePathConfig()

	// 构建代理配置响应
	config := map[string]interface{}{
		"enabled":     s.config.ProxyMode,
		"targetURL":   s.config.TargetURL,
		"authType":    s.config.TargetAuthType,
		"storagePath": storagePath,
	}

	// 根据认证类型添加相应字段
	switch s.config.TargetAuthType {
	case "basic":
		config["username"] = s.config.TargetUsername
		// 密码使用占位符
		if s.config.TargetPassword != "" {
			config["password"] = "••••••••"
		} else {
			config["password"] = ""
		}
	case "token":
		// 令牌使用占位符
		if s.config.TargetToken != "" {
			config["token"] = "••••••••"
		} else {
			config["token"] = ""
		}
	}

	// 返回配置
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "代理配置获取成功",
		Data: config,
	})
}

// SaveProxyConfig 保存代理配置
func (s *UIServer) SaveProxyConfig(c *gin.Context) {
	var config struct {
		Enabled     bool   `json:"enabled"`
		TargetURL   string `json:"targetURL"`
		AuthType    string `json:"authType"`
		Username    string `json:"username,omitempty"`
		Password    string `json:"password,omitempty"`
		Token       string `json:"token,omitempty"`
		StoragePath string `json:"storagePath,omitempty"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10001,
			Msg:  "无效的请求格式",
		})
		return
	}

	// 保存存储路径配置
	if config.StoragePath != "" {
		if err := saveStoragePathConfig(config.StoragePath); err != nil {
			c.JSON(http.StatusInternalServerError, StandardResponse{
				Code: 10003,
				Msg:  "保存存储路径失败: " + err.Error(),
			})
			return
		}
	}

	// 返回成功响应
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "代理配置已更新",
	})
}

// HandleUIChat 处理前端UI的聊天请求
func (s *UIServer) HandleUIChat(c *gin.Context) {
	// 从请求中获取消息数据
	var chatRequest struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
	}

	if err := c.ShouldBindJSON(&chatRequest); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10020,
			Msg:  "无效的聊天请求: " + err.Error(),
		})
		return
	}

	// 使用OpenAI服务处理聊天请求
	requestBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10021,
			Msg:  "读取请求失败: " + err.Error(),
		})
		return
	}

	// 设置请求头
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 使用OpenAI服务处理请求
	statusCode, responseHeaders, responseBody, err := s.openaiService.HandleRequest(
		"POST", "/chat/completions", headers, map[string]string{}, requestBody,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10022,
			Msg:  "聊天请求处理失败: " + err.Error(),
		})
		return
	}

	// 设置响应头
	for key, value := range responseHeaders {
		c.Header(key, value)
	}

	// 返回OpenAI的原始响应
	c.JSON(statusCode, responseBody)
}

// GetAPIToken 返回用于OpenAI API请求的token
func (s *UIServer) GetAPIToken(c *gin.Context) {
	// 根据不同的认证模式返回适当的token
	var tokenValue string

	if s.config.ProxyMode && s.config.TargetAuthType == "token" && s.config.TargetToken != "" {
		// 如果是代理模式且使用token认证，返回目标API的token
		tokenValue = s.config.TargetToken
	} else {
		// 否则返回默认token
		tokenValue = "mt-mitm-server-token"
	}

	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "获取API Token成功",
		Data: map[string]string{
			"token": tokenValue,
		},
	})
}

// 简单地从固定位置读取存储路径配置
func getStoragePathConfig() string {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// 配置文件路径
	configDir := filepath.Join(homeDir, ".mitm-openai-server")
	configFile := filepath.Join(configDir, "storage_path.txt")

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return ""
	}

	return string(data)
}

// 保存存储路径配置到固定位置
func saveStoragePathConfig(path string) error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户主目录: %v", err)
	}

	// 配置文件路径
	configDir := filepath.Join(homeDir, ".mitm-openai-server")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %v", err)
	}

	configFile := filepath.Join(configDir, "storage_path.txt")

	// 保存配置
	if err := os.WriteFile(configFile, []byte(path), 0644); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	return nil
}

// SetConfig 设置或更新服务器配置
// 用于在服务器启动后更新UI配置
func (s *UIServer) SetConfig(config ServerConfig) {
	if config.UIUsername != "" {
		s.config.UIUsername = config.UIUsername
	}
	if config.UIPassword != "" {
		s.config.UIPassword = config.UIPassword
		fmt.Printf("UI服务器密码已更新: %s (长度: %d)\n", s.config.UIPassword, len(s.config.UIPassword))
	}
}
