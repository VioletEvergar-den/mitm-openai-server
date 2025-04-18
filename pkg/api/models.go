package api

// Request 表示接收到的请求数据
type Request struct {
	ID        string            `json:"id"`
	Timestamp string            `json:"timestamp"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Query     map[string]string `json:"query"`
	Body      interface{}       `json:"body,omitempty"`
	IPAddress string            `json:"ip_address"`
}

// StandardResponse 表示API返回的标准响应
type StandardResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
