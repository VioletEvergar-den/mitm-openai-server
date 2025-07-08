package storage

import (
	"time"
)

// User 用户表结构，包含基本信息和配置
type User struct {
	ID          int64     `json:"id" db:"id"`                       // 用户ID（主键）
	Username    string    `json:"username" db:"username"`           // 用户名（唯一）
	Password    string    `json:"password" db:"password"`           // 密码（应该加密存储）
	UserType    string    `json:"user_type" db:"user_type"`         // 用户类型：root, user
	IsActive    bool      `json:"is_active" db:"is_active"`         // 是否激活
	CreatedAt   time.Time `json:"created_at" db:"created_at"`       // 创建时间
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`       // 更新时间
	LastLoginAt time.Time `json:"last_login_at" db:"last_login_at"` // 最后登录时间

	// 代理配置
	ProxyEnabled   bool   `json:"proxy_enabled" db:"proxy_enabled"`       // 代理是否启用
	ProxyTargetURL string `json:"proxy_target_url" db:"proxy_target_url"` // 代理目标URL
	ProxyAuthType  string `json:"proxy_auth_type" db:"proxy_auth_type"`   // 代理认证类型
	ProxyUsername  string `json:"proxy_username" db:"proxy_username"`     // 代理用户名
	ProxyPassword  string `json:"proxy_password" db:"proxy_password"`     // 代理密码
	ProxyToken     string `json:"proxy_token" db:"proxy_token"`           // 代理Token

	// 其他配置
	StoragePath       string `json:"storage_path" db:"storage_path"`               // 存储路径（保留字段，但不再允许用户修改）
	MaxRequests       int    `json:"max_requests" db:"max_requests"`               // 最大请求数限制
	DataRetentionDays int    `json:"data_retention_days" db:"data_retention_days"` // 数据保留天数
}

// Request 表示一个API请求及其相关数据
type Request struct {
	ID        string                 `json:"id" db:"id"`                       // 请求的唯一标识符
	UserID    int64                  `json:"user_id" db:"user_id"`             // 所属用户ID（新增）
	Method    string                 `json:"method" db:"method"`               // HTTP方法
	Path      string                 `json:"path" db:"path"`                   // 请求路径
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`         // 请求时间
	Headers   map[string]string      `json:"headers" db:"headers"`             // 请求头
	Query     map[string]string      `json:"query" db:"query"`                 // 查询参数
	Body      interface{}            `json:"body" db:"body"`                   // 请求体
	ClientIP  string                 `json:"client_ip" db:"client_ip"`         // 客户端IP
	IPAddress string                 `json:"ip_address" db:"ip_address"`       // 兼容旧版的IP字段
	Response  *ProxyResponse         `json:"response,omitempty" db:"response"` // 响应
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata"` // 元数据
}

// ProxyResponse 表示代理请求的响应
type ProxyResponse struct {
	StatusCode int               `json:"status_code"`          // HTTP状态码
	Headers    map[string]string `json:"headers"`              // 响应头
	Body       interface{}       `json:"body"`                 // 响应体
	Latency    int64             `json:"latency_ms,omitempty"` // 响应延迟（毫秒）
}

// Stats 表示存储的统计信息
type Stats struct {
	TotalRequests int                    `json:"total_requests"`          // 请求总数
	ApiCounts     map[string]int         `json:"api_counts,omitempty"`    // 各API路径计数
	MethodCounts  map[string]int         `json:"method_counts,omitempty"` // 各方法计数
	StatusCounts  map[int]int            `json:"status_counts,omitempty"` // 各状态码计数
	ClientIPs     map[string]int         `json:"client_ips,omitempty"`    // 各客户端IP计数
	Extra         map[string]interface{} `json:"extra,omitempty"`         // 其他统计信息
}

// UserConfig 用户配置结构（用于API交互）
type UserConfig struct {
	ProxyEnabled      bool   `json:"proxy_enabled"`
	ProxyTargetURL    string `json:"proxy_target_url"`
	ProxyAuthType     string `json:"proxy_auth_type"`
	ProxyUsername     string `json:"proxy_username,omitempty"`
	ProxyPassword     string `json:"proxy_password,omitempty"`
	ProxyToken        string `json:"proxy_token,omitempty"`
	MaxRequests       int    `json:"max_requests"`
	DataRetentionDays int    `json:"data_retention_days"`
}

// UserProfile 用户资料结构（用于API交互，不包含敏感信息）
type UserProfile struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	UserType    string    `json:"user_type"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}
