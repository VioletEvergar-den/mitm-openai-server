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
