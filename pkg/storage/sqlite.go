package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage 实现Storage接口，使用SQLite作为存储后端
type SQLiteStorage struct {
	db         *sql.DB
	dbPath     string
	dataFolder string
}

// NewSQLiteStorage 创建一个新的SQLite存储实例
func NewSQLiteStorage(dataFolder string) (*SQLiteStorage, error) {
	// 如果未提供路径，使用默认位置
	if dataFolder == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("获取用户主目录失败: %w", err)
		}
		dataFolder = filepath.Join(homeDir, ".fake-openapi-server", "sqlite")
	}

	// 确保数据文件夹存在
	if err := os.MkdirAll(dataFolder, 0755); err != nil {
		return nil, fmt.Errorf("创建数据文件夹失败: %w", err)
	}

	dbPath := filepath.Join(dataFolder, "requests.db")

	// 确保数据库文件所在目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开SQLite数据库失败: %w", err)
	}

	// 验证连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("无法连接到SQLite数据库: %w", err)
	}

	storage := &SQLiteStorage{
		db:         db,
		dbPath:     dbPath,
		dataFolder: dataFolder,
	}

	// 初始化数据库表
	if err := storage.initDatabase(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	return storage, nil
}

// initDatabase 初始化数据库表
func (s *SQLiteStorage) initDatabase() error {
	// 创建请求表
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS requests (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			method TEXT NOT NULL,
			path TEXT NOT NULL,
			headers TEXT NOT NULL,
			query TEXT NOT NULL,
			body TEXT,
			ip_address TEXT NOT NULL,
			response TEXT
		)
	`)
	return err
}

// SaveRequest 将请求保存到SQLite数据库
func (s *SQLiteStorage) SaveRequest(req *Request) error {
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

	// 获取客户端IP地址，兼容IPAddress和ClientIP字段
	clientIP := req.ClientIP
	if clientIP == "" && req.IPAddress != "" {
		clientIP = req.IPAddress
	}

	_, err = s.db.Exec(
		`INSERT OR REPLACE INTO requests 
		(id, timestamp, method, path, headers, query, body, ip_address, response) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ID,
		req.Timestamp,
		req.Method,
		req.Path,
		string(headersJSON),
		string(queryJSON),
		string(bodyJSON),
		clientIP,
		string(responseJSON),
	)

	if err != nil {
		return fmt.Errorf("保存请求到数据库失败: %w", err)
	}

	return nil
}

// GetRequestByID 根据ID获取请求
func (s *SQLiteStorage) GetRequestByID(id string) (*Request, error) {
	var (
		req          Request
		timestampStr string
		headersJSON  string
		queryJSON    string
		bodyJSON     sql.NullString
		responseJSON sql.NullString
		ipAddress    string
	)

	err := s.db.QueryRow(
		`SELECT id, timestamp, method, path, headers, query, body, ip_address, response 
		FROM requests WHERE id = ?`, id,
	).Scan(
		&req.ID,
		&timestampStr,
		&req.Method,
		&req.Path,
		&headersJSON,
		&queryJSON,
		&bodyJSON,
		&ipAddress,
		&responseJSON,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("未找到ID为 %s 的请求", id)
		}
		return nil, fmt.Errorf("查询请求失败: %w", err)
	}

	// 设置IP地址字段
	req.ClientIP = ipAddress
	req.IPAddress = ipAddress

	// 解析时间戳
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		req.Timestamp = timestampStr
	} else {
		req.Timestamp = timestamp
	}

	// 解析JSON字段
	if err := json.Unmarshal([]byte(headersJSON), &req.Headers); err != nil {
		return nil, fmt.Errorf("解析请求头JSON失败: %w", err)
	}

	if err := json.Unmarshal([]byte(queryJSON), &req.Query); err != nil {
		return nil, fmt.Errorf("解析查询参数JSON失败: %w", err)
	}

	if bodyJSON.Valid && bodyJSON.String != "" {
		if err := json.Unmarshal([]byte(bodyJSON.String), &req.Body); err != nil {
			return nil, fmt.Errorf("解析请求体JSON失败: %w", err)
		}
	}

	if responseJSON.Valid && responseJSON.String != "" {
		var resp ProxyResponse
		if err := json.Unmarshal([]byte(responseJSON.String), &resp); err != nil {
			return nil, fmt.Errorf("解析响应JSON失败: %w", err)
		}
		req.Response = &resp
	}

	return &req, nil
}

// GetAllRequests 获取所有请求
func (s *SQLiteStorage) GetAllRequests(limit int, offset int) ([]*Request, error) {
	if limit <= 0 {
		limit = 100 // 默认限制为100条记录
	}

	rows, err := s.db.Query(
		`SELECT id, timestamp, method, path, headers, query, body, ip_address, response 
		FROM requests ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("查询所有请求失败: %w", err)
	}
	defer rows.Close()

	var requests []*Request
	for rows.Next() {
		var (
			req          Request
			timestampStr string
			headersJSON  string
			queryJSON    string
			bodyJSON     sql.NullString
			responseJSON sql.NullString
			ipAddress    string
		)

		err := rows.Scan(
			&req.ID,
			&timestampStr,
			&req.Method,
			&req.Path,
			&headersJSON,
			&queryJSON,
			&bodyJSON,
			&ipAddress,
			&responseJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描请求行失败: %w", err)
		}

		// 设置IP地址字段
		req.ClientIP = ipAddress
		req.IPAddress = ipAddress

		// 解析时间戳
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			req.Timestamp = timestampStr
		} else {
			req.Timestamp = timestamp
		}

		// 解析JSON字段
		if err := json.Unmarshal([]byte(headersJSON), &req.Headers); err != nil {
			return nil, fmt.Errorf("解析请求头JSON失败: %w", err)
		}

		if err := json.Unmarshal([]byte(queryJSON), &req.Query); err != nil {
			return nil, fmt.Errorf("解析查询参数JSON失败: %w", err)
		}

		if bodyJSON.Valid && bodyJSON.String != "" {
			if err := json.Unmarshal([]byte(bodyJSON.String), &req.Body); err != nil {
				return nil, fmt.Errorf("解析请求体JSON失败: %w", err)
			}
		}

		if responseJSON.Valid && responseJSON.String != "" {
			var resp ProxyResponse
			if err := json.Unmarshal([]byte(responseJSON.String), &resp); err != nil {
				return nil, fmt.Errorf("解析响应JSON失败: %w", err)
			}
			req.Response = &resp
		}

		requests = append(requests, &req)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("迭代请求行时出错: %w", err)
	}

	return requests, nil
}

// DeleteRequest 删除指定ID的请求
func (s *SQLiteStorage) DeleteRequest(id string) error {
	result, err := s.db.Exec("DELETE FROM requests WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除请求失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取受影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("未找到ID为 %s 的请求", id)
	}

	return nil
}

// DeleteAllRequests 删除所有请求
func (s *SQLiteStorage) DeleteAllRequests() error {
	_, err := s.db.Exec("DELETE FROM requests")
	if err != nil {
		return fmt.Errorf("删除所有请求失败: %w", err)
	}
	return nil
}

// Close 关闭数据库连接
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// ExportRequests 导出所有请求到JSON文件
func (s *SQLiteStorage) ExportRequests() (string, error) {
	requests, err := s.GetAllRequests(1000000, 0) // 使用一个很大的限制来获取所有请求
	if err != nil {
		return "", fmt.Errorf("获取所有请求失败: %w", err)
	}

	exportPath := filepath.Join(s.dataFolder, fmt.Sprintf("requests_export_%s.json", time.Now().Format("20060102_150405")))
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
func (s *SQLiteStorage) GetStats() (map[string]interface{}, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM requests").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("获取请求数量失败: %w", err)
	}

	stats := map[string]interface{}{
		"total_requests": count,
		"database_path":  s.dbPath,
	}

	// 获取方法统计
	rows, err := s.db.Query("SELECT method, COUNT(*) as count FROM requests GROUP BY method")
	if err != nil {
		return nil, fmt.Errorf("获取方法统计失败: %w", err)
	}
	defer rows.Close()

	methodStats := make(map[string]int)
	for rows.Next() {
		var method string
		var methodCount int
		if err := rows.Scan(&method, &methodCount); err != nil {
			return nil, fmt.Errorf("扫描方法统计行失败: %w", err)
		}
		methodStats[method] = methodCount
	}
	stats["methods"] = methodStats

	// 获取最近的请求时间
	var latestTimestamp sql.NullString
	err = s.db.QueryRow("SELECT timestamp FROM requests ORDER BY timestamp DESC LIMIT 1").Scan(&latestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("获取最近请求时间失败: %w", err)
	}

	if latestTimestamp.Valid {
		stats["latest_request"] = latestTimestamp.String
	} else {
		stats["latest_request"] = nil
	}

	return stats, nil
}
