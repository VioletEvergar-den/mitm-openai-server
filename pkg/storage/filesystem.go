package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileSystemStorage 实现Storage接口，使用文件系统作为存储后端
type FileSystemStorage struct {
	dataFolder string
}

// NewFileSystemStorage 创建一个新的文件系统存储实例
func NewFileSystemStorage(dataFolder string) (*FileSystemStorage, error) {
	// 如果未提供路径，使用默认位置
	if dataFolder == "" {
		// 获取当前目录作为默认存储位置
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("获取当前工作目录失败: %w", err)
		}
		dataFolder = filepath.Join(currentDir, "data")
	}

	// 确保数据文件夹存在
	if err := os.MkdirAll(dataFolder, 0755); err != nil {
		return nil, fmt.Errorf("创建数据文件夹失败: %w", err)
	}

	return &FileSystemStorage{
		dataFolder: dataFolder,
	}, nil
}

// getRequestPath 获取请求文件的完整路径
func (fs *FileSystemStorage) getRequestPath(id string) string {
	return filepath.Join(fs.dataFolder, fmt.Sprintf("%s.json", id))
}

// SaveRequest 将请求保存到文件系统
func (fs *FileSystemStorage) SaveRequest(req *Request) error {
	filePath := fs.getRequestPath(req.ID)

	// 检查ID是否为空
	if req.ID == "" {
		return fmt.Errorf("请求ID不能为空")
	}

	// 确保ClientIP字段有值（兼容旧的IPAddress字段）
	if req.ClientIP == "" && req.IPAddress != "" {
		req.ClientIP = req.IPAddress
	} else if req.IPAddress == "" && req.ClientIP != "" {
		req.IPAddress = req.ClientIP
	}

	// 确保Timestamp字段有值
	if req.Timestamp == nil {
		req.Timestamp = time.Now()
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建请求文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(req); err != nil {
		return fmt.Errorf("编码请求数据失败: %w", err)
	}

	return nil
}

// GetRequestByID 根据ID获取请求
func (fs *FileSystemStorage) GetRequestByID(id string) (*Request, error) {
	filePath := fs.getRequestPath(id)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("未找到ID为 %s 的请求", id)
		}
		return nil, fmt.Errorf("读取请求文件失败: %w", err)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("解析请求数据失败: %w", err)
	}

	return &req, nil
}

// GetAllRequests 获取所有请求，支持分页
func (fs *FileSystemStorage) GetAllRequests(limit int, offset int) ([]*Request, error) {
	if limit <= 0 {
		limit = 100 // 默认限制为100条记录
	}

	// 读取数据目录中的所有JSON文件
	files, err := os.ReadDir(fs.dataFolder)
	if err != nil {
		return nil, fmt.Errorf("读取数据目录失败: %w", err)
	}

	var requestFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			requestFiles = append(requestFiles, file.Name())
		}
	}

	// 按照文件修改时间排序
	sort.Slice(requestFiles, func(i, j int) bool {
		iPath := filepath.Join(fs.dataFolder, requestFiles[i])
		jPath := filepath.Join(fs.dataFolder, requestFiles[j])

		iInfo, err := os.Stat(iPath)
		if err != nil {
			return false
		}

		jInfo, err := os.Stat(jPath)
		if err != nil {
			return true
		}

		return iInfo.ModTime().After(jInfo.ModTime())
	})

	// 应用分页
	end := offset + limit
	if end > len(requestFiles) {
		end = len(requestFiles)
	}

	if offset >= len(requestFiles) {
		return []*Request{}, nil
	}

	pagedFiles := requestFiles[offset:end]

	// 读取并解析每个请求文件
	var requests []*Request
	for _, fileName := range pagedFiles {
		filePath := filepath.Join(fs.dataFolder, fileName)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("读取请求文件失败: %w", err)
		}

		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, fmt.Errorf("解析请求数据失败: %w", err)
		}

		requests = append(requests, &req)
	}

	return requests, nil
}

// DeleteRequest 删除指定ID的请求
func (fs *FileSystemStorage) DeleteRequest(id string) error {
	filePath := fs.getRequestPath(id)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("未找到ID为 %s 的请求", id)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除请求文件失败: %w", err)
	}

	return nil
}

// DeleteAllRequests 删除所有请求
func (fs *FileSystemStorage) DeleteAllRequests() error {
	files, err := os.ReadDir(fs.dataFolder)
	if err != nil {
		return fmt.Errorf("读取数据目录失败: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(fs.dataFolder, file.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("删除请求文件失败: %w", err)
			}
		}
	}

	return nil
}

// Close 关闭存储，对于文件系统实现，不需要特殊操作
func (fs *FileSystemStorage) Close() error {
	return nil
}

// ExportRequests 导出所有请求到JSON文件
func (fs *FileSystemStorage) ExportRequests() (string, error) {
	requests, err := fs.GetAllRequests(1000000, 0) // 使用一个很大的限制来获取所有请求
	if err != nil {
		return "", fmt.Errorf("获取所有请求失败: %w", err)
	}

	exportPath := filepath.Join(fs.dataFolder, fmt.Sprintf("requests_export_%s.json", time.Now().Format("20060102_150405")))
	file, err := os.Create(exportPath)
	if err != nil {
		return "", fmt.Errorf("创建导出文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(requests); err != nil {
		return "", fmt.Errorf("编码请求数据失败: %w", err)
	}

	return exportPath, nil
}

// GetStats 获取存储统计信息
func (fs *FileSystemStorage) GetStats() (map[string]interface{}, error) {
	files, err := os.ReadDir(fs.dataFolder)
	if err != nil {
		return nil, fmt.Errorf("读取数据目录失败: %w", err)
	}

	// 统计请求文件数量
	var requestFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") && !strings.HasPrefix(file.Name(), "requests_export_") {
			requestFiles = append(requestFiles, file.Name())
		}
	}

	totalRequests := len(requestFiles)

	// 统计各HTTP方法数量
	methodStats := make(map[string]int)
	var latestTime time.Time

	for _, fileName := range requestFiles {
		filePath := filepath.Join(fs.dataFolder, fileName)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // 跳过无法读取的文件
		}

		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			continue // 跳过无法解析的文件
		}

		// 统计HTTP方法
		methodStats[req.Method]++

		// 检查是否是最新请求
		fileInfo, err := os.Stat(filePath)
		if err == nil {
			if fileInfo.ModTime().After(latestTime) {
				latestTime = fileInfo.ModTime()
			}
		}
	}

	stats := map[string]interface{}{
		"total_requests": totalRequests,
		"storage_path":   fs.dataFolder,
		"methods":        methodStats,
	}

	if !latestTime.IsZero() {
		stats["latest_request"] = latestTime.Format(time.RFC3339)
	} else {
		stats["latest_request"] = nil
	}

	return stats, nil
}
