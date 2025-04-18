package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/llm-sec/mitm-openai-server/pkg/testutils"
)

// TestV1APIHandlers 测试v1版本API处理程序
func TestV1APIHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建服务器
	server := NewServer(store)

	// 测试v1/users GET接口
	t.Run("获取用户列表", func(t *testing.T) {
		// 创建测试请求
		req := httptest.NewRequest("GET", "/v1/users", nil)
		w := httptest.NewRecorder()

		// 模拟请求
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		server.handleV1Users(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code, "应返回200状态码")

		var users []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &users)
		assert.NoError(t, err, "响应应为有效JSON")
		assert.NotEmpty(t, users, "用户列表不应为空")
	})

	// 测试v1/users/:id GET接口
	t.Run("获取单个用户", func(t *testing.T) {
		// 创建测试请求
		req := httptest.NewRequest("GET", "/v1/users/1", nil)
		w := httptest.NewRecorder()

		// 模拟请求
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		server.handleV1UserByID(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code, "应返回200状态码")

		var user map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &user)
		assert.NoError(t, err, "响应应为有效JSON")
		assert.Equal(t, "1", user["id"], "用户ID应匹配")
		assert.Contains(t, user, "name", "应包含name字段")
	})

	// 测试v1/echo POST接口
	t.Run("Echo接口", func(t *testing.T) {
		// 创建测试请求
		body := `{"message":"测试消息"}`
		req := httptest.NewRequest("POST", "/v1/echo", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// 模拟请求处理
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		server.handleV1Echo(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code, "应返回200状态码")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "响应应为有效JSON")
		assert.Equal(t, "测试消息", response["message"], "消息应匹配")
		assert.Contains(t, response, "timestamp", "应包含timestamp字段")
		assert.Contains(t, response, "request_id", "应包含request_id字段")
	})
}

// TestV2APIHandlers 测试v2版本API处理程序
func TestV2APIHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时存储
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()

	// 创建服务器
	server := NewServer(store)

	// 测试v2/users GET接口
	t.Run("获取用户列表", func(t *testing.T) {
		// 创建测试请求
		req := httptest.NewRequest("GET", "/v2/users", nil)
		w := httptest.NewRecorder()

		// 模拟请求
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		server.handleV2Users(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code, "应返回200状态码")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "响应应为有效JSON")
		assert.Contains(t, response, "data", "应包含data字段")
		assert.Contains(t, response, "total", "应包含total字段")

		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "data应为数组")
		assert.NotEmpty(t, data, "用户列表不应为空")
	})

	// 测试v2/users/:id GET接口
	t.Run("获取单个用户", func(t *testing.T) {
		// 创建测试请求
		req := httptest.NewRequest("GET", "/v2/users/1", nil)
		w := httptest.NewRecorder()

		// 模拟请求
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		server.handleV2UserByID(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code, "应返回200状态码")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "响应应为有效JSON")
		assert.Contains(t, response, "data", "应包含data字段")

		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok, "data应为对象")
		assert.Equal(t, "1", data["id"], "用户ID应匹配")
	})

	// 测试v2/echo POST接口
	t.Run("Echo接口", func(t *testing.T) {
		// 创建测试请求
		body := `{"message":"测试消息"}`
		req := httptest.NewRequest("POST", "/v2/echo", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// 模拟请求处理
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		server.handleV2Echo(c)

		// 验证结果
		assert.Equal(t, http.StatusOK, w.Code, "应返回200状态码")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "响应应为有效JSON")
		assert.Contains(t, response, "data", "应包含data字段")
		assert.Contains(t, response, "metadata", "应包含metadata字段")

		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok, "data应为对象")
		assert.Equal(t, "测试消息", data["message"], "消息应匹配")

		metadata, ok := response["metadata"].(map[string]interface{})
		assert.True(t, ok, "metadata应为对象")
		assert.Contains(t, metadata, "timestamp", "应包含timestamp字段")
		assert.Contains(t, metadata, "request_id", "应包含request_id字段")
		assert.Equal(t, "v2", metadata["version"], "版本应为v2")
	})
}
