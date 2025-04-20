package storage

import (
	"time"
)

// Request 表示一个API请求及其相关数据
type Request struct {
	ID        string                 `json:"id"`                 // 请求的唯一标识符
	Method    string                 `json:"method"`             // HTTP方法
	Path      string                 `json:"path"`               // 请求路径
	Timestamp time.Time              `json:"timestamp"`          // 请求时间
	Headers   map[string]string      `json:"headers"`            // 请求头
	Query     map[string]string      `json:"query"`              // 查询参数
	Body      interface{}            `json:"body"`               // 请求体
	ClientIP  string                 `json:"client_ip"`          // 客户端IP
	IPAddress string                 `json:"ip_address"`         // 兼容旧版的IP字段
	Response  *ProxyResponse         `json:"response,omitempty"` // 响应
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // 元数据
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
