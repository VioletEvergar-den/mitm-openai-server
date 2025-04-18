package server

import (
	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openapi-server/pkg/storage"
)

// Request 表示接收到的请求数据
type Request struct {
	ID        string            `json:"id"`
	Timestamp string            `json:"timestamp"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Query     map[string]string `json:"query"`
	Body      interface{}       `json:"body,omitempty"`
	IPAddress string            `json:"ip_address"`
	Response  *ProxyResponse    `json:"response,omitempty"`
}

// ProxyResponse 表示从目标API获取的响应
type ProxyResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       interface{}       `json:"body,omitempty"`
}

// StandardResponse 表示API返回的标准响应
type StandardResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ServerConfig 存储服务器配置
type ServerConfig struct {
	Storage        storage.Storage
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
	storage storage.Storage
	config  ServerConfig
}

// OpenAPISpec 表示OpenAPI规范
type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    map[string]interface{} `json:"info"`
	Paths   map[string]interface{} `json:"paths"`
}
