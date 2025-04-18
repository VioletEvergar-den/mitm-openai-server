package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemStorage(t *testing.T) {
	// 创建临时测试目录
	testDir, err := os.MkdirTemp("", "filesystem-storage-test")
	if err != nil {
		t.Fatalf("创建临时测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir) // 测试完成后清理

	// 创建FileSystemStorage实例
	storage, err := NewFileSystemStorage(testDir)
	if err != nil {
		t.Fatalf("创建FileSystemStorage失败: %v", err)
	}

	// 测试保存请求
	t.Run("SaveRequest", func(t *testing.T) {
		req := &Request{
			ID:        "test-id-1",
			Method:    "GET",
			Path:      "/api/test",
			Timestamp: time.Now(),
			Headers: map[string]string{
				"User-Agent": "Test-Agent",
			},
			ClientIP: "127.0.0.1",
		}

		if err := storage.SaveRequest(req); err != nil {
			t.Fatalf("保存请求失败: %v", err)
		}

		// 验证文件是否存在
		filePath := filepath.Join(testDir, "test-id-1.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatalf("请求文件不存在: %s", filePath)
		}
	})

	// 测试获取单个请求
	t.Run("GetRequestByID", func(t *testing.T) {
		req, err := storage.GetRequestByID("test-id-1")
		if err != nil {
			t.Fatalf("获取请求失败: %v", err)
		}

		if req.ID != "test-id-1" || req.Method != "GET" || req.Path != "/api/test" {
			t.Fatalf("获取的请求数据不匹配: %+v", req)
		}
	})

	// 测试获取不存在的请求
	t.Run("GetRequestByID_NotFound", func(t *testing.T) {
		_, err := storage.GetRequestByID("non-existent-id")
		if err == nil {
			t.Fatalf("应该返回错误，但返回了nil")
		}
	})

	// 保存更多请求用于测试分页
	t.Run("SaveMultipleRequests", func(t *testing.T) {
		// 保存10个请求
		for i := 2; i <= 10; i++ {
			req := &Request{
				ID:        "test-id-" + string([]byte{byte('0') + byte(i)}),
				Method:    "POST",
				Path:      "/api/test" + string([]byte{byte('0') + byte(i)}),
				Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			}
			if err := storage.SaveRequest(req); err != nil {
				t.Fatalf("保存请求失败: %v", err)
			}
		}
	})

	// 测试获取所有请求
	t.Run("GetAllRequests", func(t *testing.T) {
		// 不分页，获取所有
		requests, err := storage.GetAllRequests(0, 0)
		if err != nil {
			t.Fatalf("获取所有请求失败: %v", err)
		}

		if len(requests) != 10 {
			t.Fatalf("期望获取10条请求，实际获取 %d 条", len(requests))
		}

		// 测试分页
		pageRequests, err := storage.GetAllRequests(5, 0)
		if err != nil {
			t.Fatalf("获取分页请求失败: %v", err)
		}

		if len(pageRequests) != 5 {
			t.Fatalf("期望获取5条请求，实际获取 %d 条", len(pageRequests))
		}

		// 测试偏移
		offsetRequests, err := storage.GetAllRequests(5, 5)
		if err != nil {
			t.Fatalf("获取偏移请求失败: %v", err)
		}

		if len(offsetRequests) != 5 {
			t.Fatalf("期望获取5条请求，实际获取 %d 条", len(offsetRequests))
		}

		// 超出范围的偏移
		beyondRequests, err := storage.GetAllRequests(5, 20)
		if err != nil {
			t.Fatalf("获取超出范围请求失败: %v", err)
		}

		if len(beyondRequests) != 0 {
			t.Fatalf("期望获取0条请求，实际获取 %d 条", len(beyondRequests))
		}
	})

	// 测试删除单个请求
	t.Run("DeleteRequest", func(t *testing.T) {
		if err := storage.DeleteRequest("test-id-1"); err != nil {
			t.Fatalf("删除请求失败: %v", err)
		}

		// 验证文件已被删除
		filePath := filepath.Join(testDir, "test-id-1.json")
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Fatalf("请求文件应该已被删除: %s", filePath)
		}

		// 尝试再次获取已删除的请求
		_, err := storage.GetRequestByID("test-id-1")
		if err == nil {
			t.Fatalf("应该返回错误，但返回了nil")
		}
	})

	// 测试删除不存在的请求
	t.Run("DeleteRequest_NotFound", func(t *testing.T) {
		err := storage.DeleteRequest("non-existent-id")
		if err == nil {
			t.Fatalf("应该返回错误，但返回了nil")
		}
	})

	// 测试导出请求
	t.Run("ExportRequests", func(t *testing.T) {
		exportPath, err := storage.ExportRequests()
		if err != nil {
			t.Fatalf("导出请求失败: %v", err)
		}

		// 验证导出文件是否存在
		if _, err := os.Stat(exportPath); os.IsNotExist(err) {
			t.Fatalf("导出文件不存在: %s", exportPath)
		}

		// 清理导出文件
		os.Remove(exportPath)
	})

	// 测试获取统计信息
	t.Run("GetStats", func(t *testing.T) {
		stats, err := storage.GetStats()
		if err != nil {
			t.Fatalf("获取统计信息失败: %v", err)
		}

		totalRequests, ok := stats["total_requests"].(int)
		if !ok {
			t.Fatalf("统计信息中没有total_requests字段或类型错误")
		}

		// 我们已删除一个请求，所以现在应该有9个
		if totalRequests != 9 {
			t.Fatalf("期望有9个请求，但统计结果为 %d", totalRequests)
		}

		methods, ok := stats["methods"].(map[string]int)
		if !ok {
			t.Fatalf("统计信息中没有methods字段或类型错误")
		}

		// 验证方法统计
		if methods["POST"] != 9 {
			t.Fatalf("期望有9个POST请求，但统计结果为 %d", methods["POST"])
		}
	})

	// 测试删除所有请求
	t.Run("DeleteAllRequests", func(t *testing.T) {
		if err := storage.DeleteAllRequests(); err != nil {
			t.Fatalf("删除所有请求失败: %v", err)
		}

		// 验证是否所有文件都被删除
		files, err := os.ReadDir(testDir)
		if err != nil {
			t.Fatalf("读取测试目录失败: %v", err)
		}

		jsonCount := 0
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".json" {
				jsonCount++
			}
		}

		if jsonCount > 0 {
			t.Fatalf("所有JSON文件应该已被删除，但仍有 %d 个", jsonCount)
		}

		// 确认获取所有请求返回空结果
		requests, err := storage.GetAllRequests(100, 0)
		if err != nil {
			t.Fatalf("获取所有请求失败: %v", err)
		}

		if len(requests) != 0 {
			t.Fatalf("期望获取0条请求，实际获取 %d 条", len(requests))
		}
	})

	// 测试关闭存储
	t.Run("Close", func(t *testing.T) {
		if err := storage.Close(); err != nil {
			t.Fatalf("关闭存储失败: %v", err)
		}
	})
}

func TestFileSystemStorageReopen(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "filesystem-storage-reopen-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	createStorage := func() (Storage, error) {
		return NewFileSystemStorage(tempDir)
	}

	t.Run("FileSystemStorageReopen", func(t *testing.T) {
		testStorageReopen(t, createStorage, func() {
			os.RemoveAll(tempDir)
		})
	})
}

func TestFileSystemStorageEdgeCases(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "filesystem-storage-edge-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("DefaultLocation", func(t *testing.T) {
		// 测试不提供路径时的默认位置
		storage, err := NewFileSystemStorage("")
		require.NoError(t, err)
		defer storage.Close()

		// 验证可以正常工作
		req := createTestRequest("default-location-test")
		err = storage.SaveRequest(req)
		require.NoError(t, err)

		// 获取存储统计信息，确认工作正常
		stats, err := storage.GetStats()
		require.NoError(t, err)
		require.NotNil(t, stats)
	})

	t.Run("AccessDenied", func(t *testing.T) {
		// 跳过Windows测试，因为权限模型不同
		if os.Getenv("GOOS") == "windows" {
			t.Skip("跳过Windows测试")
		}

		// 创建一个只读目录
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0500) // 只有读/执行权限
		require.NoError(t, err)

		// 使用只读目录测试写入操作
		storage, err := NewFileSystemStorage(readOnlyDir)
		if err != nil {
			// 如果创建失败，跳过后续测试
			t.Skip("无法创建只读目录存储")
		}
		defer storage.Close()

		// 尝试写入
		req := createTestRequest("readonly-test")
		err = storage.SaveRequest(req)
		assert.Error(t, err, "应该无法写入只读目录")
	})

	t.Run("FileCorruption", func(t *testing.T) {
		storage, err := NewFileSystemStorage(tempDir)
		require.NoError(t, err)
		defer storage.Close()

		// 保存正常请求
		req := createTestRequest("corruption-test")
		err = storage.SaveRequest(req)
		require.NoError(t, err)

		// 使文件损坏
		filePath := filepath.Join(tempDir, "corruption-test.json")
		err = os.WriteFile(filePath, []byte("invalid json"), 0644)
		require.NoError(t, err)

		// 尝试读取损坏的文件
		_, err = storage.GetRequestByID("corruption-test")
		assert.Error(t, err, "读取损坏的文件应该返回错误")
		assert.Contains(t, err.Error(), "解析请求数据失败", "应该是解析错误")

		// 能够删除损坏的文件
		err = storage.DeleteRequest("corruption-test")
		assert.NoError(t, err, "应该能删除损坏的文件")
	})

	t.Run("Timestamps", func(t *testing.T) {
		storage, err := NewFileSystemStorage(tempDir)
		require.NoError(t, err)
		defer storage.Close()

		// 清理存储
		err = storage.DeleteAllRequests()
		require.NoError(t, err)

		// 测试不同类型的时间戳
		testCases := []struct {
			name      string
			timestamp interface{}
		}{
			{"TimeObject", time.Now()},
			{"ISO8601String", time.Now().Format(time.RFC3339)},
			{"UnixTimestamp", time.Now().Unix()},
			{"NilTimestamp", nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				id := fmt.Sprintf("timestamp-test-%s", tc.name)
				req := createTestRequest(id)
				req.Timestamp = tc.timestamp
				err := storage.SaveRequest(req)
				assert.NoError(t, err)

				// 读取并验证
				retrievedReq, err := storage.GetRequestByID(id)
				assert.NoError(t, err)
				assert.NotNil(t, retrievedReq.Timestamp, "时间戳不应为nil")
			})
		}
	})
}
