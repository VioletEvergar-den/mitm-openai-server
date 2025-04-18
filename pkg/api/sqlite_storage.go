package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage 实现了将请求存储到SQLite数据库中
type SQLiteStorage struct {
	dbPath string
	db     *sql.DB
	mutex  sync.RWMutex
}

// NewSQLiteStorage 创建一个新的SQLite存储实例
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// 确保数据库目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 打开或创建数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开SQLite数据库失败: %w", err)
	}

	// 创建表
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("创建数据库表失败: %w", err)
	}

	return &SQLiteStorage{
		dbPath: dbPath,
		db:     db,
	}, nil
}

// 创建必要的表
func createTables(db *sql.DB) error {
	// 请求表
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS requests (
        id TEXT PRIMARY KEY,
        timestamp TEXT NOT NULL,
        method TEXT NOT NULL,
        path TEXT NOT NULL,
        headers TEXT NOT NULL,
        query TEXT NOT NULL,
        body TEXT,
        ip_address TEXT NOT NULL,
        response TEXT
    )`)

	return err
}

// Close 关闭数据库连接
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SaveRequest 将请求保存到数据库中
func (s *SQLiteStorage) SaveRequest(req *Request) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 如果没有ID，生成一个
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// 如果没有时间戳，添加当前时间
	if req.Timestamp == "" {
		req.Timestamp = time.Now().Format(time.RFC3339)
	}

	// 序列化JSON字段
	headersJSON, err := json.Marshal(req.Headers)
	if err != nil {
		return fmt.Errorf("序列化请求头失败: %w", err)
	}

	queryJSON, err := json.Marshal(req.Query)
	if err != nil {
		return fmt.Errorf("序列化查询参数失败: %w", err)
	}

	var bodyJSON []byte
	if req.Body != nil {
		bodyJSON, err = json.Marshal(req.Body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %w", err)
		}
	}

	var responseJSON []byte
	if req.Response != nil {
		responseJSON, err = json.Marshal(req.Response)
		if err != nil {
			return fmt.Errorf("序列化响应失败: %w", err)
		}
	}

	// 检查是否已存在该ID的记录
	var exists int
	err = s.db.QueryRow("SELECT COUNT(*) FROM requests WHERE id = ?", req.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("检查请求ID是否存在失败: %w", err)
	}

	if exists > 0 {
		// 更新现有记录
		_, err = s.db.Exec(
			"UPDATE requests SET timestamp = ?, method = ?, path = ?, headers = ?, query = ?, body = ?, ip_address = ?, response = ? WHERE id = ?",
			req.Timestamp, req.Method, req.Path, string(headersJSON), string(queryJSON),
			stringOrNil(bodyJSON), req.IPAddress, stringOrNil(responseJSON), req.ID,
		)
	} else {
		// 插入新记录
		_, err = s.db.Exec(
			"INSERT INTO requests (id, timestamp, method, path, headers, query, body, ip_address, response) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			req.ID, req.Timestamp, req.Method, req.Path, string(headersJSON), string(queryJSON),
			stringOrNil(bodyJSON), req.IPAddress, stringOrNil(responseJSON),
		)
	}

	if err != nil {
		return fmt.Errorf("保存请求到数据库失败: %w", err)
	}

	return nil
}

// GetAllRequests 获取所有保存的请求
func (s *SQLiteStorage) GetAllRequests() ([]*Request, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rows, err := s.db.Query("SELECT id, timestamp, method, path, headers, query, body, ip_address, response FROM requests ORDER BY timestamp DESC")
	if err != nil {
		return nil, fmt.Errorf("查询请求失败: %w", err)
	}
	defer rows.Close()

	var requests []*Request
	for rows.Next() {
		var (
			id, timestamp, method, path, ipAddress     string
			headersJSON, queryJSON, bodyJSON, respJSON sql.NullString
		)

		if err := rows.Scan(&id, &timestamp, &method, &path, &headersJSON, &queryJSON, &bodyJSON, &ipAddress, &respJSON); err != nil {
			return nil, fmt.Errorf("读取请求数据失败: %w", err)
		}

		req := &Request{
			ID:        id,
			Timestamp: timestamp,
			Method:    method,
			Path:      path,
			IPAddress: ipAddress,
			Headers:   make(map[string]string),
			Query:     make(map[string]string),
		}

		// 解析JSON字段
		if headersJSON.Valid && headersJSON.String != "" {
			if err := json.Unmarshal([]byte(headersJSON.String), &req.Headers); err != nil {
				continue // 跳过解析错误
			}
		}

		if queryJSON.Valid && queryJSON.String != "" {
			if err := json.Unmarshal([]byte(queryJSON.String), &req.Query); err != nil {
				continue // 跳过解析错误
			}
		}

		if bodyJSON.Valid && bodyJSON.String != "" {
			var body interface{}
			if err := json.Unmarshal([]byte(bodyJSON.String), &body); err == nil {
				req.Body = body
			}
		}

		if respJSON.Valid && respJSON.String != "" {
			var resp ProxyResponse
			if err := json.Unmarshal([]byte(respJSON.String), &resp); err == nil {
				req.Response = &resp
			}
		}

		requests = append(requests, req)
	}

	return requests, nil
}

// GetRequestByID 根据ID获取请求
func (s *SQLiteStorage) GetRequestByID(id string) (*Request, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var (
		timestamp, method, path, ipAddress         string
		headersJSON, queryJSON, bodyJSON, respJSON sql.NullString
	)

	err := s.db.QueryRow(
		"SELECT timestamp, method, path, headers, query, body, ip_address, response FROM requests WHERE id = ?",
		id,
	).Scan(&timestamp, &method, &path, &headersJSON, &queryJSON, &bodyJSON, &ipAddress, &respJSON)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("请求ID不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("获取请求失败: %w", err)
	}

	req := &Request{
		ID:        id,
		Timestamp: timestamp,
		Method:    method,
		Path:      path,
		IPAddress: ipAddress,
		Headers:   make(map[string]string),
		Query:     make(map[string]string),
	}

	// 解析JSON字段
	if headersJSON.Valid && headersJSON.String != "" {
		if err := json.Unmarshal([]byte(headersJSON.String), &req.Headers); err != nil {
			return nil, fmt.Errorf("解析请求头失败: %w", err)
		}
	}

	if queryJSON.Valid && queryJSON.String != "" {
		if err := json.Unmarshal([]byte(queryJSON.String), &req.Query); err != nil {
			return nil, fmt.Errorf("解析查询参数失败: %w", err)
		}
	}

	if bodyJSON.Valid && bodyJSON.String != "" {
		var body interface{}
		if err := json.Unmarshal([]byte(bodyJSON.String), &body); err == nil {
			req.Body = body
		}
	}

	if respJSON.Valid && respJSON.String != "" {
		var resp ProxyResponse
		if err := json.Unmarshal([]byte(respJSON.String), &resp); err == nil {
			req.Response = &resp
		}
	}

	return req, nil
}

// DeleteRequest 根据ID删除请求
func (s *SQLiteStorage) DeleteRequest(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.db.Exec("DELETE FROM requests WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除请求失败: %w", err)
	}

	return nil
}

// DeleteAllRequests 删除所有请求
func (s *SQLiteStorage) DeleteAllRequests() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.db.Exec("DELETE FROM requests")
	if err != nil {
		return fmt.Errorf("删除所有请求失败: %w", err)
	}

	return nil
}

// GetDatabaseStats 获取数据库统计信息
func (s *SQLiteStorage) GetDatabaseStats() (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := map[string]interface{}{}

	// 获取记录总数
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM requests").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("获取请求数量失败: %w", err)
	}
	stats["count"] = count

	// 获取数据库文件大小
	fileInfo, err := os.Stat(s.dbPath)
	if err != nil {
		return nil, fmt.Errorf("获取数据库文件信息失败: %w", err)
	}
	stats["size"] = fileInfo.Size()
	stats["path"] = s.dbPath

	return stats, nil
}

// ExportRequestsAsJSONL 导出所有请求为JSONL格式
func (s *SQLiteStorage) ExportRequestsAsJSONL() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 获取所有请求
	requests, err := s.GetAllRequests()
	if err != nil {
		return "", fmt.Errorf("获取请求列表失败: %w", err)
	}

	// 创建临时导出文件
	exportDir := filepath.Join(filepath.Dir(s.dbPath), "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("创建导出目录失败: %w", err)
	}

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

// stringOrNil 返回字符串或nil
func stringOrNil(data []byte) interface{} {
	if len(data) == 0 {
		return nil
	}
	return string(data)
}
