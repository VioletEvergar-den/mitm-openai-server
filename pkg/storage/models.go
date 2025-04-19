package storage

// Request 表示一个HTTP请求及其响应的记录
type Request struct {
	// ID 请求的唯一标识符
	// 例如: "req-123e4567-e89b-12d3-a456-426614174000"
	ID string `json:"id"`

	// Method HTTP请求方法
	// 例如: "GET", "POST", "PUT", "DELETE"等
	Method string `json:"method"`

	// Path 请求的路径部分，不包含查询参数
	// 例如: "/v1/chat/completions"
	Path string `json:"path"`

	// Timestamp 请求的时间戳
	// 可以是多种类型:
	// - time.Time: Go原生时间对象
	// - string: ISO8601格式的时间字符串，如"2023-04-01T12:34:56Z"
	// - int64: Unix时间戳(秒)
	Timestamp interface{} `json:"timestamp"`

	// Headers 请求头信息
	// 可以是两种格式:
	// - map[string][]string: 一个键对应多个值，如{"Content-Type": ["application/json"]}
	// - map[string]string: 一个键对应一个值(扁平化)，如{"Content-Type": "application/json"}
	Headers interface{} `json:"headers"`

	// Query 查询参数
	// 可以是两种格式:
	// - map[string][]string: 一个键对应多个值，如{"filter": ["active", "pending"]}
	// - map[string]string: 一个键对应一个值(扁平化)，如{"filter": "active"}
	Query interface{} `json:"query"`

	// Body 请求体内容
	// 可以是多种类型:
	// - 对于JSON请求: map[string]interface{}或结构化对象
	// - 对于文本请求: string
	// - 对于二进制请求: Base64编码的字符串
	// 例如: {"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Hello"}]}
	Body interface{} `json:"body,omitempty"`

	// Response 请求的响应信息
	// 如果没有响应或请求失败，则为nil
	Response *ProxyResponse `json:"response,omitempty"`

	// Metadata 与请求相关的额外元数据
	// 可用于存储非标准字段或系统特定的信息
	// 例如: {"source": "browser", "latency_ms": 150, "user_id": "u123"}
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// ClientIP 发起请求的客户端IP地址
	// 例如: "192.168.1.1", "10.0.0.2", "2001:db8::1"
	ClientIP string `json:"client_ip,omitempty"`

	// IPAddress 客户端IP地址的旧字段名
	// 为了向后兼容而保留，新代码应使用ClientIP
	// 例如: "192.168.1.1"
	IPAddress string `json:"ip_address,omitempty"`
}

// ProxyResponse 表示HTTP响应信息
type ProxyResponse struct {
	// StatusCode HTTP响应状态码
	// 例如: 200(成功), 404(未找到), 500(服务器错误)
	StatusCode int `json:"status_code"`

	// Headers 响应头信息
	// 可以是两种格式:
	// - map[string][]string: 一个键对应多个值，如{"Content-Type": ["application/json"]}
	// - map[string]string: 一个键对应一个值，如{"Content-Type": "application/json"}
	Headers interface{} `json:"headers"`

	// Body 响应体内容
	// 可以是多种类型:
	// - 对于JSON响应: map[string]interface{}或结构化对象
	// - 对于文本响应: string
	// - 对于二进制响应: Base64编码的字符串
	// 例如: {"id": "chatcmpl-123", "object": "chat.completion", "choices": [...]}
	Body interface{} `json:"body,omitempty"`

	// Latency 响应延迟时间(毫秒)
	// 从接收到请求到生成响应的时间
	// 例如: 120表示120毫秒的延迟
	Latency int64 `json:"latency,omitempty"`
}

// StorageConfig 存储系统的配置信息
type StorageConfig struct {
	// Type 存储后端类型
	// 可选值:
	// - "filesystem": 基于文件系统的存储
	// - "sqlite": 基于SQLite数据库的存储
	Type string `json:"type"`

	// Directory 文件系统存储时的目录路径
	// 当Type为"filesystem"时使用
	// 例如: "/data/requests", "./storage"
	Directory string `json:"directory"`

	// DBPath SQLite数据库文件的路径
	// 当Type为"sqlite"时使用
	// 例如: "/data/storage.db", "./requests.db"
	DBPath string `json:"db_path"`
}

// StandardResponse 表示API返回的标准格式响应
type StandardResponse struct {
	// Status 响应状态
	// 常见值: "success", "error", "pending"
	Status string `json:"status"`

	// Message 响应消息
	// 成功或错误的描述信息
	// 例如: "操作成功", "参数无效"
	Message string `json:"message"`

	// Data 响应的主体数据
	// 成功时返回的数据，可以是任何类型
	// 如列表、对象、数值等
	Data interface{} `json:"data,omitempty"`
}
