package storage

// Request 表示一个HTTP请求及其响应的记录
type Request struct {
	ID        string                 `json:"id"`                   // 请求的唯一标识
	Method    string                 `json:"method"`               // HTTP方法（GET、POST等）
	Path      string                 `json:"path"`                 // 请求路径
	Timestamp interface{}            `json:"timestamp"`            // 请求的时间戳，可以是time.Time或string
	Headers   interface{}            `json:"headers"`              // 请求头，可以是map[string][]string或map[string]string
	Query     interface{}            `json:"query"`                // 查询参数，可以是map[string][]string或map[string]string
	Body      interface{}            `json:"body,omitempty"`       // 请求体，可以是string或interface{}
	Response  *ProxyResponse         `json:"response,omitempty"`   // 响应信息
	Metadata  map[string]interface{} `json:"metadata,omitempty"`   // 元数据，用于存储附加信息
	ClientIP  string                 `json:"client_ip,omitempty"`  // 客户端IP
	IPAddress string                 `json:"ip_address,omitempty"` // 兼容旧版的IP地址字段
}

// ProxyResponse 表示HTTP响应
type ProxyResponse struct {
	StatusCode int         `json:"status_code"`       // HTTP状态码
	Headers    interface{} `json:"headers"`           // 响应头，可以是map[string][]string或map[string]string
	Body       interface{} `json:"body,omitempty"`    // 响应体，可以是string或interface{}
	Latency    int64       `json:"latency,omitempty"` // 响应延迟（毫秒）
}

// StorageConfig 存储配置信息
type StorageConfig struct {
	Type      string `json:"type"`      // 存储类型：filesystem或sqlite
	Directory string `json:"directory"` // 文件系统存储的目录路径
	DBPath    string `json:"db_path"`   // SQLite数据库文件路径
}

// StandardResponse 表示API返回的标准响应
type StandardResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
