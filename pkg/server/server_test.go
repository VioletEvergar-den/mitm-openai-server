package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/llm-sec/mitm-openapi-server/pkg/testutils"
)

// TestNewServer 测试服务器创建
func TestNewServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建服务器
	server := NewServer(store)
	assert.NotNil(t, server, "服务器不应为nil")
	assert.NotNil(t, server.router, "路由器不应为nil")
	assert.Equal(t, store, server.storage, "存储应匹配")
}

// TestNewServerWithConfig 测试使用配置创建服务器
func TestNewServerWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 测试配置
	config := ServerConfig{
		Storage:        store,
		EnableAuth:     true,
		AuthType:       "basic",
		Username:       "testuser",
		Password:       "testpass",
		EnableCORS:     true,
		AllowOrigins:   "*",
		GenerateUIAuth: true,
		UIUsername:     "admin",
	}

	// 创建服务器
	server := NewServerWithConfig(config)
	assert.NotNil(t, server, "服务器不应为nil")
	assert.Equal(t, config.Username, server.config.Username, "用户名应匹配")
	assert.Equal(t, config.Password, server.config.Password, "密码应匹配")
	assert.Equal(t, config.EnableCORS, server.config.EnableCORS, "CORS设置应匹配")
	assert.Equal(t, "admin", server.config.UIUsername, "UI用户名应匹配")
	assert.NotEmpty(t, server.config.UIPassword, "UI密码应自动生成")
}

// TestAuthMiddleware 测试认证中间件
func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 定义测试用例
	tests := []struct {
		name       string
		config     ServerConfig
		setupAuth  func(*http.Request)
		expectCode int
	}{
		{
			name: "禁用认证",
			config: ServerConfig{
				Storage:    store,
				EnableAuth: false,
			},
			setupAuth:  nil,
			expectCode: http.StatusOK,
		},
		{
			name: "基本认证-有效",
			config: ServerConfig{
				Storage:    store,
				EnableAuth: true,
				AuthType:   "basic",
				Username:   "testuser",
				Password:   "testpass",
			},
			setupAuth: func(req *http.Request) {
				req.SetBasicAuth("testuser", "testpass")
			},
			expectCode: http.StatusOK,
		},
		{
			name: "基本认证-无效",
			config: ServerConfig{
				Storage:    store,
				EnableAuth: true,
				AuthType:   "basic",
				Username:   "testuser",
				Password:   "testpass",
			},
			setupAuth: func(req *http.Request) {
				req.SetBasicAuth("wronguser", "wrongpass")
			},
			expectCode: http.StatusUnauthorized,
		},
		{
			name: "令牌认证-有效",
			config: ServerConfig{
				Storage:    store,
				EnableAuth: true,
				AuthType:   "token",
				Token:      "test-token",
			},
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer test-token")
			},
			expectCode: http.StatusOK,
		},
		{
			name: "令牌认证-无效",
			config: ServerConfig{
				Storage:    store,
				EnableAuth: true,
				AuthType:   "token",
				Token:      "test-token",
			},
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer wrong-token")
			},
			expectCode: http.StatusUnauthorized,
		},
		{
			name: "不支持的认证类型",
			config: ServerConfig{
				Storage:    store,
				EnableAuth: true,
				AuthType:   "unsupported",
			},
			setupAuth:  nil,
			expectCode: http.StatusUnauthorized,
		},
	}

	// 执行测试
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 创建带配置的服务器
			server := NewServerWithConfig(tc.config)

			// 创建测试路由
			router := gin.New()
			router.Use(server.authMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			// 创建测试请求
			req := httptest.NewRequest("GET", "/test", nil)
			if tc.setupAuth != nil {
				tc.setupAuth(req)
			}

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tc.expectCode, w.Code, "状态码应匹配")
		})
	}
}

// TestCorsMiddleware 测试CORS中间件
func TestCorsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 定义测试用例
	tests := []struct {
		name          string
		enableCORS    bool
		allowOrigins  string
		origin        string
		method        string
		expectHeaders bool
	}{
		{
			name:          "启用CORS-简单请求",
			enableCORS:    true,
			allowOrigins:  "*",
			origin:        "http://example.com",
			method:        "GET",
			expectHeaders: true,
		},
		{
			name:          "启用CORS-预检请求",
			enableCORS:    true,
			allowOrigins:  "*",
			origin:        "http://example.com",
			method:        "OPTIONS",
			expectHeaders: true,
		},
		{
			name:          "禁用CORS",
			enableCORS:    false,
			allowOrigins:  "*",
			origin:        "http://example.com",
			method:        "GET",
			expectHeaders: false,
		},
	}

	// 执行测试
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 创建带配置的服务器
			server := NewServerWithConfig(ServerConfig{
				Storage:      store,
				EnableCORS:   tc.enableCORS,
				AllowOrigins: tc.allowOrigins,
			})

			// 创建测试路由
			router := gin.New()
			router.Use(server.corsMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			// 创建测试请求
			req := httptest.NewRequest(tc.method, "/test", nil)
			req.Header.Set("Origin", tc.origin)

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证结果
			if tc.expectHeaders {
				if tc.method == "OPTIONS" {
					assert.Equal(t, http.StatusNoContent, w.Code, "预检请求应返回204")
				}
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"), "应有CORS头")
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"), "应有CORS头")
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), "不应有CORS头")
			}
		})
	}
}

// TestHealthEndpoint 测试健康检查端点
func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建服务器
	server := NewServer(store)

	// 创建测试请求
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// 执行请求
	server.router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code, "健康检查应返回200")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "响应应为有效JSON")
	assert.Equal(t, "ok", response["status"], "响应状态应为ok")
}

// TestOpenAPISpec 测试OpenAPI规范端点
func TestOpenAPISpec(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建服务器
	server := NewServer(store)

	// 测试所有OpenAPI规范端点
	endpoints := []string{"/openapi.json", "/v1/openapi.json", "/v2/openapi.json"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			// 执行请求
			server.router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, http.StatusOK, w.Code, "OpenAPI规范请求应返回200")

			var spec OpenAPISpec
			err := json.Unmarshal(w.Body.Bytes(), &spec)
			assert.NoError(t, err, "响应应为有效JSON")
			assert.Equal(t, "3.0.0", spec.OpenAPI, "OpenAPI版本应为3.0.0")
			assert.NotNil(t, spec.Info, "应包含Info对象")
			assert.NotNil(t, spec.Paths, "应包含Paths对象")
		})
	}
}
