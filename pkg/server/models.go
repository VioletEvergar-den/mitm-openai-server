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
