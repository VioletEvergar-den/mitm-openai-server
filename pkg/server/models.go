package server

import (
	"github.com/llm-sec/mitm-openai-server/pkg/api"
)

// Re-export api types
type (
	StandardResponse = api.StandardResponse
	Request          = api.Request
	ProxyResponse    = api.ProxyResponse
	ServerConfig     = api.ServerConfig
)

// OpenAISpec 表示OpenAI规范
// 用于生成和提供OpenAI文档
type OpenAISpec struct {
	Version string                 `json:"version"` // API版本，如"1.0.0"
	Info    map[string]interface{} `json:"info"`    // API信息，包含标题、描述等
	Models  []string               `json:"models"`  // 支持的模型列表
}
