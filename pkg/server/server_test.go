package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestV1Users 测试V1版本用户列表API
func TestV1Users(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 创建测试上下文
	c, w := createTestGinContext("GET", "/v1/users", nil)

	// 调用处理函数
	server.handleV1Users(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应内容
	var users []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &users)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查用户列表
	assert.Len(t, users, 3, "应返回3个用户")
	assert.Equal(t, "1", users[0]["id"], "第一个用户ID应为1")
	assert.Equal(t, "张三", users[0]["name"], "第一个用户名应为张三")
	assert.Equal(t, "admin", users[0]["role"], "第一个用户角色应为admin")
}

// TestV1CreateUser 测试V1版本创建用户API
func TestV1CreateUser(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 创建用户数据
	userData := map[string]interface{}{
		"name":     "测试用户",
		"email":    "test@example.com",
		"active":   true,
		"role":     "user",
		"location": "测试城市",
	}
	userJson, _ := json.Marshal(userData)

	// 创建测试上下文
	c, w := createTestGinContext("POST", "/v1/users", bytes.NewBuffer(userJson))

	// 调用处理函数
	server.handleV1CreateUser(c)

	// 断言
	assert.Equal(t, http.StatusCreated, w.Code)

	// 解析响应内容
	var createdUser map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createdUser)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查创建的用户
	assert.NotEmpty(t, createdUser["id"], "用户ID不应为空")
	assert.Equal(t, "测试用户", createdUser["name"], "用户名应为测试用户")
	assert.Equal(t, "test@example.com", createdUser["email"], "邮箱应为test@example.com")
	assert.Equal(t, true, createdUser["active"], "用户应为激活状态")
	assert.Equal(t, "user", createdUser["role"], "用户角色应为user")
	assert.Equal(t, "测试城市", createdUser["location"], "位置应为测试城市")
	assert.NotEmpty(t, createdUser["created"], "创建时间不应为空")
}

// TestV1UserByID 测试V1版本获取单个用户API
func TestV1UserByID(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 测试获取存在的用户
	t.Run("ExistingUser", func(t *testing.T) {
		c, w := createTestRequestWithParams("GET", "/v1/users/:id", gin.Params{
			{Key: "id", Value: "1"},
		}, nil)

		// 调用处理函数
		server.handleV1UserByID(c)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应内容
		var user map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &user)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查用户数据
		assert.Equal(t, "1", user["id"], "用户ID应为1")
		assert.Equal(t, "张三", user["name"], "用户名应为张三")
		assert.Equal(t, "admin", user["role"], "用户角色应为admin")

		// 检查profile字段
		profile, ok := user["profile"].(map[string]interface{})
		require.True(t, ok, "应有profile字段")
		assert.Equal(t, "https://example.com/avatars/1.jpg", profile["avatar"], "头像URL应正确")
		assert.Equal(t, "13800138001", profile["phone"], "电话号码应正确")
	})

	// 测试获取不存在的用户
	t.Run("NonExistingUser", func(t *testing.T) {
		c, w := createTestRequestWithParams("GET", "/v1/users/:id", gin.Params{
			{Key: "id", Value: "999"},
		}, nil)

		// 调用处理函数
		server.handleV1UserByID(c)

		// 断言
		assert.Equal(t, http.StatusNotFound, w.Code)

		// 解析响应内容
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查错误消息
		assert.Equal(t, "用户不存在", response["error"], "应返回用户不存在错误")
		assert.Contains(t, response["message"].(string), "999", "错误消息应包含用户ID")
	})
}

// TestV1Echo 测试V1版本Echo API
func TestV1Echo(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 创建Echo数据
	echoData := map[string]interface{}{
		"message": "测试消息",
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}
	echoJson, _ := json.Marshal(echoData)

	// 创建测试上下文
	c, w := createTestGinContext("POST", "/v1/echo", bytes.NewBuffer(echoJson))

	// 调用处理函数
	server.handleV1Echo(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查响应内容
	assert.NotEmpty(t, response["request_id"], "请求ID不应为空")
	assert.NotEmpty(t, response["timestamp"], "时间戳不应为空")
	assert.Equal(t, "POST", response["method"], "方法应为POST")
	assert.Equal(t, "/v1/echo", response["path"], "路径应为/v1/echo")
	assert.Equal(t, "测试消息", response["message"], "应回显message字段")

	// 检查请求体
	body, ok := response["body"].(map[string]interface{})
	require.True(t, ok, "响应应包含请求体")
	assert.Equal(t, "测试消息", body["message"], "请求体应包含message字段")
}

// TestV2Users 测试V2版本用户列表API
func TestV2Users(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 创建测试上下文
	c, w := createTestGinContext("GET", "/v2/users", nil)

	// 调用处理函数
	server.handleV2Users(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查数据部分
	data, ok := response["data"].([]interface{})
	require.True(t, ok, "应包含数据数组")
	assert.Len(t, data, 3, "应返回3个用户")

	// 检查第一个用户
	user1, ok := data[0].(map[string]interface{})
	require.True(t, ok, "用户应为对象")
	assert.Equal(t, "1", user1["id"], "第一个用户ID应为1")
	assert.Equal(t, "张三", user1["name"], "第一个用户名应为张三")

	// 检查元数据
	metadata, ok := response["metadata"].(map[string]interface{})
	require.True(t, ok, "应包含元数据")
	assert.Equal(t, "v2", metadata["version"], "版本应为v2")
	assert.Equal(t, "users", metadata["api_name"], "API名称应为users")
}

// TestV2CreateUser 测试V2版本创建用户API
func TestV2CreateUser(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 创建用户数据
	userData := map[string]interface{}{
		"data": map[string]interface{}{
			"name":   "测试用户",
			"email":  "test@example.com",
			"active": true,
			"role":   "user",
		},
	}
	userJson, _ := json.Marshal(userData)

	// 创建测试上下文
	c, w := createTestGinContext("POST", "/v2/users", bytes.NewBuffer(userJson))

	// 调用处理函数
	server.handleV2CreateUser(c)

	// 断言
	assert.Equal(t, http.StatusCreated, w.Code)

	// 解析响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查数据部分
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "应包含数据对象")
	assert.NotEmpty(t, data["id"], "用户ID不应为空")
	assert.Equal(t, "测试用户", data["name"], "用户名应为测试用户")

	// 检查元数据
	metadata, ok := response["metadata"].(map[string]interface{})
	require.True(t, ok, "应包含元数据")
	assert.Equal(t, "v2", metadata["version"], "版本应为v2")
	assert.Equal(t, "create", metadata["action"], "动作应为create")
}

// TestV2UserByID 测试V2版本获取单个用户API
func TestV2UserByID(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 测试获取存在的用户
	t.Run("ExistingUser", func(t *testing.T) {
		c, w := createTestRequestWithParams("GET", "/v2/users/:id", gin.Params{
			{Key: "id", Value: "1"},
		}, nil)

		// 调用处理函数
		server.handleV2UserByID(c)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应内容
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查数据部分
		data, ok := response["data"].(map[string]interface{})
		require.True(t, ok, "应包含数据对象")
		assert.Equal(t, "1", data["id"], "用户ID应为1")
		assert.Equal(t, "张三", data["name"], "用户名应为张三")

		// 检查profile字段
		profile, ok := data["profile"].(map[string]interface{})
		require.True(t, ok, "应有profile字段")
		assert.Equal(t, "https://example.com/avatars/1.jpg", profile["avatar"], "头像URL应正确")

		// 检查元数据
		metadata, ok := response["metadata"].(map[string]interface{})
		require.True(t, ok, "应包含元数据")
		assert.Equal(t, "v2", metadata["version"], "版本应为v2")
		assert.Equal(t, "get", metadata["action"], "动作应为get")
	})

	// 测试获取不存在的用户
	t.Run("NonExistingUser", func(t *testing.T) {
		c, w := createTestRequestWithParams("GET", "/v2/users/:id", gin.Params{
			{Key: "id", Value: "999"},
		}, nil)

		// 调用处理函数
		server.handleV2UserByID(c)

		// 断言
		assert.Equal(t, http.StatusNotFound, w.Code)

		// 解析响应内容
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查错误部分
		error, ok := response["error"].(map[string]interface{})
		require.True(t, ok, "应包含错误对象")
		assert.Equal(t, "USER_NOT_FOUND", error["code"], "错误代码应为USER_NOT_FOUND")
		assert.Contains(t, error["message"].(string), "999", "错误消息应包含用户ID")
	})
}

// TestV2Echo 测试V2版本Echo API
func TestV2Echo(t *testing.T) {
	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 创建Echo数据
	echoData := map[string]interface{}{
		"message": "测试消息",
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}
	echoJson, _ := json.Marshal(echoData)

	// 创建测试上下文
	c, w := createTestGinContext("POST", "/v2/echo", bytes.NewBuffer(echoJson))

	// 调用处理函数
	server.handleV2Echo(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应内容
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查数据部分
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "应包含数据对象")
	assert.Equal(t, "测试消息", data["message"], "应回显message字段")
	assert.NotEmpty(t, data["request_id"], "请求ID不应为空")

	// 检查元数据
	metadata, ok := response["metadata"].(map[string]interface{})
	require.True(t, ok, "应包含元数据")
	assert.Equal(t, "v2", metadata["version"], "版本应为v2")
	assert.Equal(t, "echo", metadata["api_name"], "API名称应为echo")
}

// TestGenerateRandomPassword 测试生成随机密码函数
func TestGenerateRandomPassword(t *testing.T) {
	// 测试默认长度
	password := generateRandomPassword(12)
	assert.Len(t, password, 12, "密码长度应为12")

	// 测试最小长度限制
	shortPassword := generateRandomPassword(4)
	assert.GreaterOrEqual(t, len(shortPassword), 8, "密码长度应至少为8")

	// 测试生成的多个密码应不同
	passwords := make(map[string]bool)
	for i := 0; i < 10; i++ {
		pwd := generateRandomPassword(12)
		assert.False(t, passwords[pwd], "生成的密码应该不重复")
		passwords[pwd] = true
	}
}

// 更多测试函数...
