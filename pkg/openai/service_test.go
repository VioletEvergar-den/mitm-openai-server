package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试根据配置创建服务
func TestNewService(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedType   string
		expectedResult string
	}{
		{
			name: "使用模拟模式",
			config: Config{
				Enabled:   true,
				ProxyMode: false,
			},
			expectedType:   "*openai.mockService",
			expectedResult: "MockOpenAIService",
		},
		{
			name: "使用代理模式，有目标URL",
			config: Config{
				Enabled:   true,
				ProxyMode: true,
				TargetURL: "https://api.openai.com",
			},
			expectedType:   "*openai.proxyService",
			expectedResult: "OpenAI API Proxy",
		},
		{
			name: "使用代理模式，无目标URL（回退到模拟模式）",
			config: Config{
				Enabled:   true,
				ProxyMode: true,
				TargetURL: "",
			},
			expectedType:   "*openai.mockService",
			expectedResult: "MockOpenAIService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建服务
			service := NewService(tt.config)

			// 验证服务名称
			assert.Equal(t, tt.expectedResult, service.Name(), "服务名称应匹配")
		})
	}
}

// 测试根据名称获取服务
func TestGetServiceByName(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		expectedResult string
	}{
		{
			name:           "使用mock名称",
			serviceName:    "mock",
			expectedResult: "MockOpenAIService",
		},
		{
			name:           "使用中文mock名称",
			serviceName:    "模拟",
			expectedResult: "MockOpenAIService",
		},
		{
			name:           "使用proxy名称",
			serviceName:    "proxy",
			expectedResult: "OpenAI API Proxy",
		},
		{
			name:           "使用中文proxy名称",
			serviceName:    "代理",
			expectedResult: "OpenAI API Proxy",
		},
		{
			name:           "使用未知名称（回退到模拟模式）",
			serviceName:    "unknown",
			expectedResult: "MockOpenAIService",
		},
	}

	// 创建基本配置
	config := Config{
		Enabled:   true,
		ProxyMode: false,
		TargetURL: "https://api.openai.com", // 代理模式需要
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 获取服务
			service := GetServiceByName(tt.serviceName, config)

			// 验证服务名称
			assert.Equal(t, tt.expectedResult, service.Name(), "服务名称应匹配")
		})
	}
}
