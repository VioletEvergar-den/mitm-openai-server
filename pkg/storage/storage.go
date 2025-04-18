package storage

// Storage 定义了存储请求和响应的接口
type Storage interface {
	// SaveRequest 保存一个HTTP请求及其响应到存储中
	SaveRequest(req *Request) error

	// GetRequestByID 根据ID获取指定的请求
	GetRequestByID(id string) (*Request, error)

	// GetAllRequests 获取所有请求，支持分页
	// limit: 每页记录数量，0表示不限制
	// offset: 跳过的记录数量，用于分页
	GetAllRequests(limit, offset int) ([]*Request, error)

	// DeleteRequest 从存储中删除指定ID的请求
	DeleteRequest(id string) error

	// DeleteAllRequests 从存储中删除所有请求
	DeleteAllRequests() error

	// ExportRequests 导出请求记录到文件，返回文件路径
	ExportRequests() (string, error)

	// GetStats 获取存储的统计信息
	GetStats() (map[string]interface{}, error)

	// Close 关闭存储，释放资源
	Close() error
}
