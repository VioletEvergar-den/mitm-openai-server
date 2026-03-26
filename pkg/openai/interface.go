package openai

import (
	"github.com/gin-gonic/gin"
)

// Service 定义OpenAI API服务接口
// 所有OpenAI服务的实现都应该遵循这个接口
type Service interface {
	// Name 返回服务名称
	Name() string

	// ServeOpenAISpec 提供OpenAI API规范
	ServeOpenAISpec(c *gin.Context)

	// HandleRequest 处理API请求并返回模拟响应
	// method: HTTP方法(GET, POST等)
	// path: 请求路径
	// headers: 请求头
	// queryParams: 查询参数
	// body: 请求体(字节数组)
	// 返回值: 状态码，响应头，响应体，错误信息
	HandleRequest(method, path string, headers, queryParams map[string]string, body []byte) (int, map[string]string, interface{}, error)

	// UpdateConfig 更新服务配置
	UpdateConfig(config Config)
}

// Response 定义标准API响应
type Response struct {
	// 通用字段
	ID      string   `json:"id,omitempty"`
	Object  string   `json:"object,omitempty"`
	Created int64    `json:"created,omitempty"`
	Model   string   `json:"model,omitempty"`
	Choices []Choice `json:"choices,omitempty"`
	Usage   *Usage   `json:"usage,omitempty"`

	// 错误字段
	Error *ErrorResp `json:"error,omitempty"`

	// 其他数据，用于自定义响应
	Data interface{} `json:"data,omitempty"`
}

// Choice 定义补全选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// Message 定义消息内容
type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content"`
}

// Usage 定义Token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResp 错误响应
type ErrorResp struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// 服务创建工厂函数类型定义
type ServiceCreator func(config Config) Service

// MockServiceCreator 创建Mock服务的函数
var MockServiceCreator ServiceCreator

// ProxyServiceCreator 创建代理服务的函数
var ProxyServiceCreator ServiceCreator
