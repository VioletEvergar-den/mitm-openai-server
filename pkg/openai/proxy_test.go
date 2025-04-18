package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 测试代理服务创建
func TestNewProxyService(t *testing.T) {
	// 创建基本配置
	config := Config{
		Enabled:         true,
		ProxyMode:       true,
		TargetURL:       "https://api.openai.com",
		TargetAuthType:  "bearer",
		TargetToken:     "test-token",
		ResponseDelayMs: 0,
	}

	// 创建服务
	service := newProxyService(config)

	// 验证服务实例
	assert.NotNil(t, service, "服务实例不应为nil")
	assert.Equal(t, "OpenAI API Proxy", service.Name(), "服务名称应为OpenAI API Proxy")
}

// 测试代理服务的ServeOpenAISpec方法
func TestProxyService_ServeOpenAISpec(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个响应记录器
	w := httptest.NewRecorder()

	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)

	// 创建服务（无目标URL的配置）
	config := Config{
		Enabled:   true,
		ProxyMode: true,
		TargetURL: "", // 无目标URL，将使用默认规范
	}
	service := newProxyService(config)

	// 调用方法
	service.ServeOpenAISpec(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code, "状态码应为200")

	// 解析响应体
	var respMap map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &respMap)
	assert.NoError(t, err, "应该能够解析JSON响应")

	// 验证关键字段
	assert.Contains(t, respMap, "openapi", "响应应包含openapi字段")
	assert.Contains(t, respMap, "info", "响应应包含info字段")
	assert.Contains(t, respMap, "paths", "响应应包含paths字段")
}

// 测试代理服务在缺少目标URL时的处理
func TestProxyService_HandleRequest_NoTargetURL(t *testing.T) {
	// 创建无目标URL的配置
	config := Config{
		Enabled:   true,
		ProxyMode: true,
		TargetURL: "", // 无目标URL
	}
	service := newProxyService(config)

	// 执行请求
	status, headers, respBody, err := service.HandleRequest(
		"POST",
		"/v1/chat/completions",
		map[string]string{"Content-Type": "application/json"},
		nil,
		[]byte(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}`),
	)

	// 验证结果
	assert.NoError(t, err, "不应该有错误")
	assert.Equal(t, http.StatusBadRequest, status, "状态码应为400")
	assert.Equal(t, "application/json", headers["Content-Type"], "内容类型应为JSON")
	assert.NotNil(t, respBody, "响应体不应为空")

	// 验证错误响应
	errResp, ok := respBody.(ErrorResp)
	assert.True(t, ok, "响应应为ErrorResp类型")
	assert.Contains(t, errResp.Message, "未配置目标OpenAI API URL", "错误消息应提及未配置URL")
	assert.Equal(t, "configuration_error", errResp.Type, "错误类型应为configuration_error")
}

// 测试API是否支持的方法
func TestProxyService_IsAPISupported(t *testing.T) {
	// 创建服务
	config := Config{
		Enabled:   true,
		ProxyMode: true,
		TargetURL: "https://api.openai.com",
	}
	service := newProxyService(config).(*proxyService)

	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "支持的API - 精确匹配",
			apiKey:   "POST /v1/chat/completions",
			expected: true,
		},
		{
			name:     "支持的API - GET模型列表",
			apiKey:   "GET /v1/models",
			expected: true,
		},
		{
			name:     "不支持的API - 不存在的路径",
			apiKey:   "GET /v1/unknown",
			expected: false,
		},
		{
			name:     "不支持的API - 不支持的方法",
			apiKey:   "PUT /v1/chat/completions",
			expected: false,
		},
		{
			name:     "支持的API - 带参数路径",
			apiKey:   "GET /v1/models/123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isAPISupported(tt.apiKey)
			assert.Equal(t, tt.expected, result, "API支持检查结果应匹配预期")
		})
	}
}

// 测试构建URL方法
func TestProxyService_BuildURL(t *testing.T) {
	tests := []struct {
		name      string
		targetURL string
		path      string
		expected  string
	}{
		{
			name:      "基础URL，无斜杠",
			targetURL: "https://api.openai.com",
			path:      "/v1/chat/completions",
			expected:  "https://api.openai.com/v1/chat/completions",
		},
		{
			name:      "基础URL，有斜杠",
			targetURL: "https://api.openai.com/",
			path:      "/v1/chat/completions",
			expected:  "https://api.openai.com/v1/chat/completions",
		},
		{
			name:      "基础URL，有斜杠，路径无斜杠",
			targetURL: "https://api.openai.com/",
			path:      "v1/chat/completions",
			expected:  "https://api.openai.com/v1/chat/completions",
		},
		{
			name:      "基础URL，无斜杠，路径无斜杠",
			targetURL: "https://api.openai.com",
			path:      "v1/chat/completions",
			expected:  "https://api.openai.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建服务
			service := &proxyService{
				config: Config{
					TargetURL: tt.targetURL,
				},
			}

			result := service.buildURL(tt.path)
			assert.Equal(t, tt.expected, result, "构建的URL应匹配预期")
		})
	}
}

// 测试添加认证头方法
func TestProxyService_AddAuthHeader(t *testing.T) {
	tests := []struct {
		name           string
		targetAuthType string
		targetUsername string
		targetPassword string
		targetToken    string
		checkHeader    func(*testing.T, http.Header)
	}{
		{
			name:           "基本认证",
			targetAuthType: "basic",
			targetUsername: "user",
			targetPassword: "pass",
			checkHeader: func(t *testing.T, h http.Header) {
				auth := h.Get("Authorization")
				assert.NotEmpty(t, auth, "Authorization头应存在")
				assert.Contains(t, auth, "Basic ", "应为基本认证")
			},
		},
		{
			name:           "令牌认证 - 带Bearer前缀",
			targetAuthType: "token",
			targetToken:    "Bearer test-token",
			checkHeader: func(t *testing.T, h http.Header) {
				auth := h.Get("Authorization")
				assert.Equal(t, "Bearer test-token", auth, "Authorization头应包含完整令牌")
			},
		},
		{
			name:           "令牌认证 - 不带Bearer前缀",
			targetAuthType: "token",
			targetToken:    "test-token",
			checkHeader: func(t *testing.T, h http.Header) {
				auth := h.Get("Authorization")
				assert.Equal(t, "Bearer test-token", auth, "Authorization头应添加Bearer前缀")
			},
		},
		{
			name:           "无认证",
			targetAuthType: "none",
			checkHeader: func(t *testing.T, h http.Header) {
				auth := h.Get("Authorization")
				assert.Empty(t, auth, "不应有Authorization头")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建服务
			service := &proxyService{
				config: Config{
					TargetAuthType: tt.targetAuthType,
					TargetUsername: tt.targetUsername,
					TargetPassword: tt.targetPassword,
					TargetToken:    tt.targetToken,
				},
			}

			// 创建请求
			req, _ := http.NewRequest("GET", "http://example.com", nil)

			// 添加认证
			service.addAuthHeader(req)

			// 验证头
			tt.checkHeader(t, req.Header)
		})
	}
}
