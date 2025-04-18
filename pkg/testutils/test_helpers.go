package testutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/llm-sec/mitm-openai-server/pkg/storage"
)

// CreateTestStorage 创建一个用于测试的存储实例
func CreateTestStorage(t *testing.T) (storage.Storage, func()) {
	tempDir, err := os.MkdirTemp("", "server_test")
	assert.NoError(t, err, "应能创建临时目录")

	// 创建SQLite存储
	dbPath := filepath.Join(tempDir, "test.db")
	store, err := storage.NewSQLiteStorage(dbPath)
	assert.NoError(t, err, "应能创建SQLite存储")

	// 清理函数
	cleanup := func() {
		store.Close()
		os.RemoveAll(tempDir)
	}

	return store, cleanup
}

// CreateTestGinContext 创建一个用于测试的Gin上下文
func CreateTestGinContext(method string, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

// AssertStatusCode 断言状态码
func AssertStatusCode(t *testing.T, expected int, actual int, message string) {
	assert.Equal(t, expected, actual, message)
}

// AssertJSONResponse 断言JSON响应
func AssertJSONResponse(t *testing.T, responseBody []byte, assertFunc func(map[string]interface{})) {
	var response map[string]interface{}
	err := json.Unmarshal(responseBody, &response)
	assert.NoError(t, err, "响应应为有效JSON")
	assertFunc(response)
}

// AssertJSONArrayResponse 断言JSON数组响应
func AssertJSONArrayResponse(t *testing.T, responseBody []byte, assertFunc func([]interface{})) {
	var response []interface{}
	err := json.Unmarshal(responseBody, &response)
	assert.NoError(t, err, "响应应为有效JSON数组")
	assertFunc(response)
}

// CreateTestRequestWithParams 创建带路径参数的测试请求
func CreateTestRequestWithParams(method string, path string, params gin.Params, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = params
	return c, w
}

// 创建一个测试请求
func CreateTestRequest(id string) *storage.Request {
	if id == "" {
		id = fmt.Sprintf("test-%d", time.Now().UnixNano())
	}
	return &storage.Request{
		ID:        id,
		Method:    "GET",
		Path:      "/api/test",
		Timestamp: time.Now().Format(time.RFC3339),
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Query: map[string][]string{
			"param": {"value"},
		},
		Body:     `{"test": "data"}`,
		ClientIP: "127.0.0.1",
		Response: &storage.ProxyResponse{
			StatusCode: 200,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: `{"result": "success"}`,
		},
	}
}

// testStorageImplementation 测试Storage接口的实现
func testStorageImplementation(t *testing.T, storage storage.Storage) {
	// 清理存储
	err := storage.DeleteAllRequests()
	assert.NoError(t, err, "应能清理所有请求")

	// 保存请求
	req1 := CreateTestRequest("test-id-1")
	err = storage.SaveRequest(req1)
	assert.NoError(t, err, "应能保存请求")

	// 获取请求
	retrieved, err := storage.GetRequestByID("test-id-1")
	assert.NoError(t, err, "应能获取请求")
	assert.Equal(t, req1.ID, retrieved.ID, "ID应匹配")
	assert.Equal(t, req1.Path, retrieved.Path, "Path应匹配")

	// 获取所有请求
	allReqs, err := storage.GetAllRequests(10, 0)
	assert.NoError(t, err, "应能获取所有请求")
	assert.Len(t, allReqs, 1, "应有1个请求")

	// 删除请求
	err = storage.DeleteRequest("test-id-1")
	assert.NoError(t, err, "应能删除请求")

	// 验证已删除
	_, err = storage.GetRequestByID("test-id-1")
	assert.Error(t, err, "获取已删除的请求应返回错误")

	// 批量保存
	for i := 1; i <= 5; i++ {
		req := CreateTestRequest(fmt.Sprintf("test-id-%d", i))
		err = storage.SaveRequest(req)
		assert.NoError(t, err, "应能保存请求")
	}

	// 获取所有请求（带分页）
	pagedReqs, err := storage.GetAllRequests(3, 0)
	assert.NoError(t, err, "应能获取分页请求")
	assert.Len(t, pagedReqs, 3, "应有3个请求")

	// 获取下一页
	pagedReqs2, err := storage.GetAllRequests(3, 3)
	assert.NoError(t, err, "应能获取第二页请求")
	assert.Len(t, pagedReqs2, 2, "应有2个请求")

	// 清理所有请求
	err = storage.DeleteAllRequests()
	assert.NoError(t, err, "应能清理所有请求")

	// 验证全部已删除
	emptyReqs, err := storage.GetAllRequests(10, 0)
	assert.NoError(t, err, "应能获取空列表")
	assert.Empty(t, emptyReqs, "请求列表应为空")
}

// testStorageReopen 测试重新打开存储后能否恢复数据
func testStorageReopen(t *testing.T, createStorage func() (storage.Storage, error), cleanupFn func()) {
	var storage storage.Storage
	var err error

	// 创建存储并保存数据
	storage, err = createStorage()
	require.NoError(t, err, "应能创建存储")

	req1 := CreateTestRequest("reopen-id-1")
	err = storage.SaveRequest(req1)
	assert.NoError(t, err, "应能保存请求")

	// 关闭存储
	err = storage.Close()
	assert.NoError(t, err, "应能关闭存储")

	// 重新打开存储
	storage, err = createStorage()
	require.NoError(t, err, "应能重新打开存储")
	defer cleanupFn()

	// 验证数据是否保留
	retrieved, err := storage.GetRequestByID("reopen-id-1")
	assert.NoError(t, err, "应能获取之前保存的请求")
	assert.Equal(t, req1.ID, retrieved.ID, "ID应匹配")
}
