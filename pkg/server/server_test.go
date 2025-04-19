package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/testutils"
	"github.com/llm-sec/mitm-openai-server/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// 测试初始化
func init() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
}

// TestNewServer 测试服务器创建
func TestNewServer(t *testing.T) {
	// 创建测试存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建服务器
	server := NewServer(store)

	// 断言服务器不为空且已正确初始化
	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.storage)
}

// TestNewServerWithConfig 测试使用配置创建服务器
func TestNewServerWithConfig(t *testing.T) {
	// 创建测试存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建自定义配置
	config := ServerConfig{
		Storage:    store,
		EnableAuth: true,
		AuthType:   "basic",
		Username:   "testuser",
		Password:   "testpass",
		EnableCORS: true,
	}

	// 创建服务器
	server := NewServerWithConfig(config)

	// 断言服务器不为空且已正确初始化
	assert.NotNil(t, server)
	assert.Equal(t, true, server.config.EnableAuth)
	assert.Equal(t, "basic", server.config.AuthType)
	assert.Equal(t, "testuser", server.config.Username)
	assert.Equal(t, "testpass", server.config.Password)
	assert.Equal(t, true, server.config.EnableCORS)
}

// 测试认证中间件 - 基本认证
func TestAuthMiddleware_BasicAuth(t *testing.T) {
	// 创建测试存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建带认证的服务器
	config := ServerConfig{
		Storage:    store,
		EnableAuth: true,
		AuthType:   "basic",
		Username:   "testuser",
		Password:   "testpass",
	}
	server := NewServerWithConfig(config)

	// 创建一个测试路由
	r := gin.New()
	r.Use(server.authMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "认证成功")
	})

	// 测试无认证头的请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 测试有效认证的请求
	req = httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("testuser", "testpass")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "认证成功", w.Body.String())

	// 测试无效认证的请求
	req = httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("wronguser", "wrongpass")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// 测试认证中间件 - 令牌认证
func TestAuthMiddleware_TokenAuth(t *testing.T) {
	// 创建测试存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建带认证的服务器
	config := ServerConfig{
		Storage:    store,
		EnableAuth: true,
		AuthType:   "token",
		Token:      "test-token-123",
	}
	server := NewServerWithConfig(config)

	// 创建一个测试路由
	r := gin.New()
	r.Use(server.authMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "认证成功")
	})

	// 测试无认证头的请求
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 测试有效令牌的请求 (直接令牌)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "test-token-123")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "认证成功", w.Body.String())

	// 测试有效令牌的请求 (Bearer格式)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "认证成功", w.Body.String())

	// 测试无效令牌的请求
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "wrong-token")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// 测试CORS中间件
func TestCorsMiddleware(t *testing.T) {
	// 创建测试存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建带CORS的服务器
	config := ServerConfig{
		Storage:      store,
		EnableCORS:   true,
		AllowOrigins: "*",
	}
	server := NewServerWithConfig(config)

	// 创建一个测试路由
	r := gin.New()
	r.Use(server.corsMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "CORS测试")
	})

	// 测试普通请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS, PATCH", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// 测试预检请求
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// 创建测试Gin上下文
func createTestGinContext(method string, path string, body *bytes.Buffer) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()

	var requestBody io.Reader
	if body != nil {
		requestBody = body
	} else {
		requestBody = bytes.NewBuffer([]byte{})
	}

	req := httptest.NewRequest(method, path, requestBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

// 创建带路径参数的测试请求
func createTestRequestWithParams(method string, path string, params gin.Params, body *bytes.Buffer) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()

	var requestBody io.Reader
	if body != nil {
		requestBody = body
	} else {
		requestBody = bytes.NewBuffer([]byte{})
	}

	req := httptest.NewRequest(method, path, requestBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = params
	return c, w
}

// TestGenerateRandomPassword 测试生成随机密码函数
func TestGenerateRandomPassword(t *testing.T) {
	// 测试默认长度
	password := utils.GenerateRandomPassword(12)
	assert.Len(t, password, 12, "密码长度应为12")

	// 测试最小长度限制
	shortPassword := utils.GenerateRandomPassword(4)
	assert.GreaterOrEqual(t, len(shortPassword), 8, "密码长度应至少为8")

	// 测试生成的多个密码应不同
	passwords := make(map[string]bool)
	for i := 0; i < 10; i++ {
		pwd := utils.GenerateRandomPassword(12)
		assert.False(t, passwords[pwd], "生成的密码应该不重复")
		passwords[pwd] = true
	}
}
