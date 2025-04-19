package api

import (
	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
)

// ServerConfig 表示服务器配置
type ServerConfig struct {
	// 存储
	Storage storage.Storage

	// 认证
	EnableAuth bool
	AuthType   string
	Username   string
	Password   string
	Token      string

	// UI
	UIDir          string
	UIUsername     string
	UIPassword     string
	GenerateUIAuth bool

	// CORS
	EnableCORS   bool
	AllowOrigins string

	// 代理模式
	ProxyMode      bool
	TargetURL      string
	TargetAuthType string
	TargetUsername string
	TargetPassword string
	TargetToken    string
}

// OpenAIServiceInterface 定义与OpenAI服务交互的接口
type OpenAIServiceInterface interface {
	// ServeOpenAISpec 提供OpenAI API规范
	ServeOpenAISpec(c *gin.Context)

	// HandleRequest 处理OpenAI请求
	HandleRequest(method, path string, headers, queryParams map[string]string, body []byte) (int, map[string]string, interface{}, error)
}

// UIServerInterface 定义与UI服务交互的接口
type UIServerInterface interface {
	// SetupUIRoutes 配置UI相关路由
	SetupUIRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc, apiMiddleware gin.HandlerFunc)

	// 请求管理
	GetRequests(c *gin.Context)
	GetRequestByID(c *gin.Context)
	DeleteRequest(c *gin.Context)
	DeleteAllRequests(c *gin.Context)
	ExportRequests(c *gin.Context)
	GetStorageStats(c *gin.Context)

	// UI接口
	HandleUILogin(c *gin.Context)
	GetServerInfo(c *gin.Context)
	GetProxyConfig(c *gin.Context)
	SaveProxyConfig(c *gin.Context)
	HandleUIChat(c *gin.Context)
}

// ConfigManagerInterface 定义配置管理接口
type ConfigManagerInterface interface {
	LoadConfig() (interface{}, error)
	SaveConfig(config interface{}) error
	ApplyConfig(config interface{}, server interface{}) error
}

// StandardResponse 标准API响应格式
type StandardResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// 共享的请求和响应结构体
type Request struct {
	ID        string                 `json:"id"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Timestamp string                 `json:"timestamp"`
	IPAddress string                 `json:"ip_address"`
	Headers   map[string]string      `json:"headers"`
	Query     map[string]string      `json:"query"`
	Body      interface{}            `json:"body"`
	Response  *ProxyResponse         `json:"response,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ProxyResponse 代理响应
type ProxyResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       interface{}       `json:"body"`
	Latency    int64             `json:"latency_ms,omitempty"`
}

// 提供给server包的辅助函数

// UIServerFactory 创建UI服务器
func UIServerFactory(storage storage.Storage, config ServerConfig, openaiService OpenAIServiceInterface) UIServerInterface {
	return NewUIServer(storage, config, openaiService)
}
