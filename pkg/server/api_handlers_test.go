package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
	"github.com/llm-sec/mitm-openai-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAllRequests 测试获取所有请求的API处理程序
func TestGetAllRequests(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 添加一些测试请求数据
	t.Run("EmptyRequestList", func(t *testing.T) {
		// 确保请求列表为空
		err := store.DeleteAllRequests()
		require.NoError(t, err)

		// 创建测试上下文
		c, w := createTestGinContext("GET", "/api/requests", nil)

		// 调用处理函数
		server.getAllRequests(c)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "success", response.Status, "状态应为success")
		assert.Contains(t, response.Message, "找到 0 个请求", "消息应包含找到0个请求")

		// 检查数据
		var dataList []*Request
		dataBytes, _ := json.Marshal(response.Data)
		err = json.Unmarshal(dataBytes, &dataList)
		require.NoError(t, err, "应能解析为请求列表")
		assert.Empty(t, dataList, "请求列表应为空")
	})

	// 测试存在多个请求时的情况
	t.Run("WithRequests", func(t *testing.T) {
		// 确保请求列表为空
		err := store.DeleteAllRequests()
		require.NoError(t, err)

		// 添加一些测试请求
		for i := 1; i <= 5; i++ {
			req := testutils.CreateTestRequest("")
			err = store.SaveRequest(req)
			require.NoError(t, err)
		}

		// 创建测试上下文
		c, w := createTestGinContext("GET", "/api/requests?limit=3&offset=0", nil)

		// 调用处理函数
		server.getAllRequests(c)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "success", response.Status, "状态应为success")

		// 检查数据
		var dataList []*Request
		dataBytes, _ := json.Marshal(response.Data)
		err = json.Unmarshal(dataBytes, &dataList)
		require.NoError(t, err, "应能解析为请求列表")
		assert.Len(t, dataList, 3, "应返回3个请求")
	})

	// 测试分页
	t.Run("Pagination", func(t *testing.T) {
		// 确保请求列表为空
		err := store.DeleteAllRequests()
		require.NoError(t, err)

		// 添加一些测试请求
		for i := 1; i <= 10; i++ {
			req := testutils.CreateTestRequest("")
			err = store.SaveRequest(req)
			require.NoError(t, err)
		}

		// 测试第一页
		c1, w1 := createTestGinContext("GET", "/api/requests?limit=4&offset=0", nil)
		server.getAllRequests(c1)
		assert.Equal(t, http.StatusOK, w1.Code)

		var response1 StandardResponse
		err = json.Unmarshal(w1.Body.Bytes(), &response1)
		require.NoError(t, err)

		var page1 []*Request
		dataBytes, _ := json.Marshal(response1.Data)
		err = json.Unmarshal(dataBytes, &page1)
		require.NoError(t, err)
		assert.Len(t, page1, 4, "第一页应返回4个请求")

		// 测试第二页
		c2, w2 := createTestGinContext("GET", "/api/requests?limit=4&offset=4", nil)
		server.getAllRequests(c2)
		assert.Equal(t, http.StatusOK, w2.Code)

		var response2 StandardResponse
		err = json.Unmarshal(w2.Body.Bytes(), &response2)
		require.NoError(t, err)

		var page2 []*Request
		dataBytes, _ = json.Marshal(response2.Data)
		err = json.Unmarshal(dataBytes, &page2)
		require.NoError(t, err)
		assert.Len(t, page2, 4, "第二页应返回4个请求")

		// 确保两页没有重复项
		for _, r1 := range page1 {
			for _, r2 := range page2 {
				assert.NotEqual(t, r1.ID, r2.ID, "两页请求ID不应重复")
			}
		}
	})

	// 测试无效的分页参数
	t.Run("InvalidPaginationParams", func(t *testing.T) {
		// 创建测试上下文（无效的limit和offset）
		c, w := createTestGinContext("GET", "/api/requests?limit=invalid&offset=invalid", nil)

		// 调用处理函数
		server.getAllRequests(c)

		// 断言 - 应该使用默认值而不是失败
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetRequestByID 测试根据ID获取请求的API处理程序
func TestGetRequestByID(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 添加一个测试请求
	testReq := testutils.CreateTestRequest("test-get-id")
	err := store.SaveRequest(testReq)
	require.NoError(t, err)

	// 测试获取存在的请求
	t.Run("ExistingRequest", func(t *testing.T) {
		// 创建测试上下文
		c, w := createTestRequestWithParams("GET", "/api/requests/:id", gin.Params{
			{Key: "id", Value: "test-get-id"},
		}, nil)

		// 调用处理函数
		server.getRequestByID(c)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "success", response.Status, "状态应为success")
		assert.Equal(t, "请求获取成功", response.Message, "消息应为请求获取成功")

		// 检查数据
		var req Request
		dataBytes, _ := json.Marshal(response.Data)
		err = json.Unmarshal(dataBytes, &req)
		require.NoError(t, err, "应能解析为请求对象")
		assert.Equal(t, "test-get-id", req.ID, "请求ID应匹配")
	})

	// 测试获取不存在的请求
	t.Run("NonExistingRequest", func(t *testing.T) {
		// 创建测试上下文
		c, w := createTestRequestWithParams("GET", "/api/requests/:id", gin.Params{
			{Key: "id", Value: "non-existing-id"},
		}, nil)

		// 调用处理函数
		server.getRequestByID(c)

		// 断言
		assert.Equal(t, http.StatusNotFound, w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "error", response.Status, "状态应为error")
		assert.Contains(t, response.Message, "请求不存在", "消息应包含请求不存在")
	})

	// 测试空ID参数
	t.Run("EmptyIDParam", func(t *testing.T) {
		// 创建测试上下文
		c, w := createTestRequestWithParams("GET", "/api/requests/:id", gin.Params{
			{Key: "id", Value: ""},
		}, nil)

		// 调用处理函数
		server.getRequestByID(c)

		// 断言
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "error", response.Status, "状态应为error")
		assert.Contains(t, response.Message, "请求ID不能为空", "消息应包含ID不能为空")
	})
}

// TestDeleteRequest 测试删除请求的API处理程序
func TestDeleteRequest(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 添加一个测试请求
	testReq := testutils.CreateTestRequest("test-delete-id")
	err := store.SaveRequest(testReq)
	require.NoError(t, err)

	// 测试删除存在的请求
	t.Run("ExistingRequest", func(t *testing.T) {
		// 创建测试上下文
		c, w := createTestRequestWithParams("DELETE", "/api/requests/:id", gin.Params{
			{Key: "id", Value: "test-delete-id"},
		}, nil)

		// 调用处理函数
		server.deleteRequest(c)

		// 断言
		assert.Equal(t, http.StatusOK, w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "success", response.Status, "状态应为success")
		assert.Contains(t, response.Message, "请求已成功删除", "消息应包含请求已成功删除")

		// 验证请求确实被删除
		_, err = store.GetRequestByID("test-delete-id")
		assert.Error(t, err, "应返回错误因为请求已被删除")
	})

	// 测试删除不存在的请求
	t.Run("NonExistingRequest", func(t *testing.T) {
		// 创建测试上下文
		c, w := createTestRequestWithParams("DELETE", "/api/requests/:id", gin.Params{
			{Key: "id", Value: "non-existing-id"},
		}, nil)

		// 调用处理函数
		server.deleteRequest(c)

		// 断言 - 删除不存在的请求应该返回错误，可能是404或500
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError,
			"应返回404或500状态码，实际返回: %d", w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "error", response.Status, "状态应为error")
		assert.Contains(t, response.Message, "删除请求失败", "消息应包含删除请求失败")
	})

	// 测试空ID参数
	t.Run("EmptyIDParam", func(t *testing.T) {
		// 创建测试上下文
		c, w := createTestRequestWithParams("DELETE", "/api/requests/:id", gin.Params{
			{Key: "id", Value: ""},
		}, nil)

		// 调用处理函数
		server.deleteRequest(c)

		// 断言
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// 解析响应内容
		var response StandardResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "应能解析JSON响应")

		// 检查状态
		assert.Equal(t, "error", response.Status, "状态应为error")
		assert.Contains(t, response.Message, "请求ID不能为空", "消息应包含ID不能为空")
	})
}

// TestDeleteAllRequests 测试删除所有请求的API处理程序
func TestDeleteAllRequests(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 添加一些测试请求
	for i := 1; i <= 3; i++ {
		req := testutils.CreateTestRequest("")
		err := store.SaveRequest(req)
		require.NoError(t, err)
	}

	// 验证初始状态
	initialReqs, err := store.GetAllRequests(10, 0)
	require.NoError(t, err)
	assert.Len(t, initialReqs, 3, "应有3个请求")

	// 创建测试上下文
	c, w := createTestGinContext("DELETE", "/api/requests", nil)

	// 调用处理函数
	server.deleteAllRequests(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应内容
	var response StandardResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查状态
	assert.Equal(t, "success", response.Status, "状态应为success")
	assert.Contains(t, response.Message, "所有请求已成功删除", "消息应包含所有请求已成功删除")

	// 验证请求确实被删除
	afterReqs, err := store.GetAllRequests(10, 0)
	require.NoError(t, err)
	assert.Empty(t, afterReqs, "删除后应该没有请求")
}

// TestExportRequests 测试导出请求的API处理程序
func TestExportRequests(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 添加一些测试请求
	for i := 1; i <= 3; i++ {
		req := testutils.CreateTestRequest("")
		err := store.SaveRequest(req)
		require.NoError(t, err)
	}

	// 创建测试上下文
	c, w := createTestGinContext("GET", "/api/export", nil)

	// 调用处理函数
	server.exportRequests(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应内容
	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查状态
	assert.Equal(t, "success", response.Status, "状态应为success")
	assert.Contains(t, response.Message, "请求已成功导出", "消息应包含请求已成功导出")

	// 检查数据
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok, "数据应为对象")
	assert.Contains(t, data, "file_path", "数据应包含文件路径")
	assert.NotEmpty(t, data["file_path"], "文件路径不应为空")
}

// TestGetStorageStats 测试获取存储统计信息的API处理程序
func TestGetStorageStats(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试存储和服务器
	store, cleanup := testutils.CreateTestStorage(t)
	defer cleanup()
	server := NewServer(store)

	// 添加一些测试请求
	for i := 1; i <= 3; i++ {
		req := testutils.CreateTestRequest("")
		err := store.SaveRequest(req)
		require.NoError(t, err)
	}

	// 创建测试上下文
	c, w := createTestGinContext("GET", "/api/stats", nil)

	// 调用处理函数
	server.getStorageStats(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应内容
	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "应能解析JSON响应")

	// 检查状态
	assert.Equal(t, "success", response.Status, "状态应为success")
	assert.Contains(t, response.Message, "存储统计获取成功", "消息应包含存储统计获取成功")

	// 检查数据
	stats, ok := response.Data.(map[string]interface{})
	require.True(t, ok, "数据应为对象")
	assert.Contains(t, stats, "total_requests", "应包含请求总数")
	assert.Equal(t, float64(3), stats["total_requests"], "请求总数应为3")
}

// TestConvertStorageToServerRequest 测试存储请求模型到服务器请求模型的转换
func TestConvertStorageToServerRequest(t *testing.T) {
	// 测试字符串时间戳
	t.Run("StringTimestamp", func(t *testing.T) {
		storageReq := &storage.Request{
			ID:        "test-id",
			Method:    "GET",
			Path:      "/test",
			Timestamp: "2023-01-01T00:00:00Z",
			ClientIP:  "127.0.0.1",
		}

		serverReq := convertStorageToServerRequest(storageReq)
		assert.Equal(t, "test-id", serverReq.ID)
		assert.Equal(t, "GET", serverReq.Method)
		assert.Equal(t, "/test", serverReq.Path)
		assert.Equal(t, "2023-01-01T00:00:00Z", serverReq.Timestamp)
		assert.Equal(t, "127.0.0.1", serverReq.IPAddress)
	})

	// 测试时间对象时间戳
	t.Run("TimeTimestamp", func(t *testing.T) {
		ts := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		storageReq := &storage.Request{
			ID:        "test-id",
			Method:    "GET",
			Path:      "/test",
			Timestamp: ts,
			ClientIP:  "127.0.0.1",
		}

		serverReq := convertStorageToServerRequest(storageReq)
		assert.Equal(t, "test-id", serverReq.ID)
		assert.Equal(t, ts.Format(time.RFC3339), serverReq.Timestamp)
	})

	// 测试响应转换
	t.Run("WithResponse", func(t *testing.T) {
		storageReq := &storage.Request{
			ID:        "test-id",
			Timestamp: "2023-01-01T00:00:00Z",
			Response: &storage.ProxyResponse{
				StatusCode: 200,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: map[string]interface{}{
					"result": "success",
				},
			},
		}

		serverReq := convertStorageToServerRequest(storageReq)
		assert.NotNil(t, serverReq.Response)
		assert.Equal(t, 200, serverReq.Response.StatusCode)
		assert.Equal(t, "application/json", serverReq.Response.Headers["Content-Type"])
		body, ok := serverReq.Response.Body.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "success", body["result"])
	})

	// 测试Headers和Query转换
	t.Run("HeadersAndQuery", func(t *testing.T) {
		storageReq := &storage.Request{
			ID:        "test-id",
			Timestamp: "2023-01-01T00:00:00Z",
			Headers: map[string][]string{
				"Accept":     {"application/json"},
				"User-Agent": {"Test"},
			},
			Query: map[string][]string{
				"param1": {"value1"},
				"param2": {"value2", "value3"},
			},
		}

		serverReq := convertStorageToServerRequest(storageReq)
		assert.Equal(t, "application/json", serverReq.Headers["Accept"])
		assert.Equal(t, "Test", serverReq.Headers["User-Agent"])
		assert.Equal(t, "value1", serverReq.Query["param1"])
		assert.Equal(t, "value2", serverReq.Query["param2"]) // 只保留第一个值
	})
}
