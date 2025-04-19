package server

import (
	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/openai"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
)

// Request 表示接收到的请求数据
// 这是一个面向服务器层的请求模型，与存储层的模型有所区别
// 主要用于API响应和内部处理
type Request struct {
	ID        string            `json:"id"`                 // 请求的唯一标识符
	Timestamp string            `json:"timestamp"`          // 请求的时间戳，格式为RFC3339
	Method    string            `json:"method"`             // HTTP请求方法：GET、POST、PUT等
	Path      string            `json:"path"`               // 请求路径，不包含查询参数
	Headers   map[string]string `json:"headers"`            // HTTP请求头，扁平化为单值映射
	Query     map[string]string `json:"query"`              // URL查询参数，扁平化为单值映射
	Body      interface{}       `json:"body,omitempty"`     // 请求体，可以是任意类型
	IPAddress string            `json:"ip_address"`         // 客户端IP地址
	Response  *ProxyResponse    `json:"response,omitempty"` // 响应数据，如果有的话
}

// ProxyResponse 表示从目标API获取的响应
// 包含状态码、响应头和响应体
type ProxyResponse struct {
	StatusCode int               `json:"status_code"`    // HTTP状态码
	Headers    map[string]string `json:"headers"`        // HTTP响应头，扁平化为单值映射
	Body       interface{}       `json:"body,omitempty"` // 响应体，可以是任意类型
}

// StandardResponse 表示API返回的标准响应
// 用于统一API响应格式，包含状态码、消息和数据
type StandardResponse struct {
	Code int         `json:"code"`           // 状态码：0表示成功，非0表示各种错误
	Msg  string      `json:"msg"`            // 响应消息
	Data interface{} `json:"data,omitempty"` // 响应数据，可选
}

// ServerConfig 存储服务器配置
// 包含所有服务器运行时所需的配置参数
type ServerConfig struct {
	Storage        storage.Storage // 存储接口的实现
	EnableAuth     bool            // 是否启用认证
	AuthType       string          // 认证类型：basic或token
	Username       string          // 基本认证的用户名
	Password       string          // 基本认证的密码
	Token          string          // 令牌认证的令牌
	EnableCORS     bool            // 是否启用CORS
	AllowOrigins   string          // CORS允许的源
	UIUsername     string          // 前端UI用户名
	UIPassword     string          // 前端UI密码
	GenerateUIAuth bool            // 是否生成随机UI认证凭证
	UIDir          string          // 前端UI目录

	// 中间人代理相关配置
	ProxyMode      bool   // 是否启用代理模式
	TargetURL      string // 目标OpenAI服务地址
	TargetAuthType string // 目标API认证类型：none, basic, token
	TargetUsername string // 目标API基本认证用户名
	TargetPassword string // 目标API基本认证密码
	TargetToken    string // 目标API令牌
}

// Server 表示API服务器
// 是服务器的核心结构，包含路由引擎、存储接口和配置
type Server struct {
	router        *gin.Engine     // Gin路由引擎
	storage       storage.Storage // 存储接口
	config        ServerConfig    // 服务器配置
	openaiService openai.Service  // OpenAI服务接口
}

// OpenAISpec 表示OpenAI规范
// 用于生成和提供OpenAI文档
type OpenAISpec struct {
	Version string                 `json:"version"` // API版本，如"1.0.0"
	Info    map[string]interface{} `json:"info"`    // API信息，包含标题、描述等
	Models  []string               `json:"models"`  // 支持的模型列表
}
