package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage 接口定义了存储请求数据的方法
type Storage interface {
	SaveRequest(req *Request) error
	GetAllRequests() ([]*Request, error)
	GetRequestByID(id string) (*Request, error)
	DeleteRequest(id string) error
	DeleteAllRequests() error
	GetDatabaseStats() (map[string]interface{}, error)
	ExportRequestsAsJSONL() (string, error)
}

// FileStorage 实现了将请求存储到文件中
type FileStorage struct {
	dataDir string
	mutex   sync.RWMutex
}

// NewFileStorage 创建一个新的文件存储实例
func NewFileStorage(dataDir string) (*FileStorage, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
	}, nil
}

// SaveRequest 将请求保存到文件中
func (fs *FileStorage) SaveRequest(req *Request) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 如果没有ID，生成一个
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// 如果没有时间戳，添加当前时间
	if req.Timestamp == "" {
		req.Timestamp = time.Now().Format(time.RFC3339)
	}

	// 将请求转换为JSON
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	// 保存到文件
	filename := filepath.Join(fs.dataDir, fmt.Sprintf("%s.json", req.ID))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入请求数据失败: %w", err)
	}

	return nil
}

// GetAllRequests 获取所有保存的请求
func (fs *FileStorage) GetAllRequests() ([]*Request, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	var requests []*Request

	files, err := os.ReadDir(fs.dataDir)
	if err != nil {
		return nil, fmt.Errorf("读取数据目录失败: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(fs.dataDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue // 跳过无法读取的文件
			}

			var req Request
			if err := json.Unmarshal(data, &req); err != nil {
				continue // 跳过无法解析的文件
			}

			requests = append(requests, &req)
		}
	}

	return requests, nil
}

// GetRequestByID 根据ID获取请求
func (fs *FileStorage) GetRequestByID(id string) (*Request, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filePath := filepath.Join(fs.dataDir, fmt.Sprintf("%s.json", id))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取请求数据失败: %w", err)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("解析请求数据失败: %w", err)
	}

	return &req, nil
}

// DeleteRequest 根据ID删除请求
func (fs *FileStorage) DeleteRequest(id string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filePath := filepath.Join(fs.dataDir, fmt.Sprintf("%s.json", id))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("请求ID不存在: %s", id)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除请求失败: %w", err)
	}

	return nil
}

// DeleteAllRequests 删除所有请求
func (fs *FileStorage) DeleteAllRequests() error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	files, err := os.ReadDir(fs.dataDir)
	if err != nil {
		return fmt.Errorf("读取数据目录失败: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(fs.dataDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("删除文件失败: %w", err)
			}
		}
	}

	return nil
}

// GetDatabaseStats 获取存储统计信息
func (fs *FileStorage) GetDatabaseStats() (map[string]interface{}, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	stats := map[string]interface{}{}

	files, err := os.ReadDir(fs.dataDir)
	if err != nil {
		return nil, fmt.Errorf("读取数据目录失败: %w", err)
	}

	count := 0
	var totalSize int64
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			count++
			fileInfo, err := file.Info()
			if err == nil {
				totalSize += fileInfo.Size()
			}
		}
	}

	stats["count"] = count
	stats["size"] = totalSize
	stats["path"] = fs.dataDir

	return stats, nil
}

// ExportRequestsAsJSONL 导出所有请求为JSONL格式
func (fs *FileStorage) ExportRequestsAsJSONL() (string, error) {
	// 获取所有请求
	requests, err := fs.GetAllRequests()
	if err != nil {
		return "", fmt.Errorf("获取请求列表失败: %w", err)
	}

	// 创建导出目录
	exportDir := filepath.Join(fs.dataDir, "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("创建导出目录失败: %w", err)
	}

	// 创建导出文件
	exportPath := filepath.Join(exportDir, fmt.Sprintf("requests_export_%s.jsonl", time.Now().Format("20060102_150405")))
	file, err := os.Create(exportPath)
	if err != nil {
		return "", fmt.Errorf("创建导出文件失败: %w", err)
	}
	defer file.Close()

	// 写入JSONL格式
	encoder := json.NewEncoder(file)
	for _, req := range requests {
		if err := encoder.Encode(req); err != nil {
			return "", fmt.Errorf("写入请求数据失败: %w", err)
		}
	}

	return exportPath, nil
}
