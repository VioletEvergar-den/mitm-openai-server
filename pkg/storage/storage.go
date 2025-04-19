package storage

// Storage 定义了存储请求和响应的接口
// 该接口提供了对HTTP请求和响应数据的持久化存储、检索和管理功能
// 可以有多种实现，如基于文件系统的存储、SQLite数据库存储等
type Storage interface {
	// SaveRequest 将HTTP请求及其响应保存到存储中
	//
	// 参数:
	//   - req: 包含请求和响应数据的Request对象
	//     格式示例:
	//     {
	//       "id": "req-123e4567-e89b-12d3-a456-426614174000",
	//       "method": "POST",
	//       "path": "/v1/chat/completions",
	//       "timestamp": "2023-04-01T12:34:56Z",
	//       "headers": {"Content-Type": "application/json"},
	//       "body": {"model": "gpt-3.5-turbo", "messages": [...]}
	//     }
	//
	// 返回:
	//   - error: 保存过程中的错误，成功时返回nil
	SaveRequest(req *Request) error

	// GetRequestByID 根据唯一ID检索特定请求
	//
	// 参数:
	//   - id: 请求的唯一标识符
	//     例如: "req-123e4567-e89b-12d3-a456-426614174000"
	//
	// 返回:
	//   - *Request: 请求对象，如果找到的话
	//   - error: 检索过程中的错误，或者当请求不存在时返回"not found"类型的错误
	GetRequestByID(id string) (*Request, error)

	// GetAllRequests 获取所有请求，支持分页
	//
	// 参数:
	//   - limit: 每页返回的最大记录数
	//     例如: 10表示最多返回10条记录
	//     如果为0，则使用默认值(通常是100)
	//   - offset: 分页的起始位置(跳过的记录数)
	//     例如: 0表示从第一条记录开始，10表示从第11条记录开始
	//
	// 返回:
	//   - []*Request: 请求对象的切片，如果没有记录则返回空切片
	//   - error: 检索过程中的错误，成功时返回nil
	GetAllRequests(limit, offset int) ([]*Request, error)

	// DeleteRequest 从存储中删除指定ID的请求
	//
	// 参数:
	//   - id: 要删除的请求的唯一标识符
	//     例如: "req-123e4567-e89b-12d3-a456-426614174000"
	//
	// 返回:
	//   - error: 删除过程中的错误，如果请求不存在则返回"not found"类型的错误
	//     成功删除时返回nil
	DeleteRequest(id string) error

	// DeleteAllRequests 删除存储中的所有请求
	//
	// 此操作通常不可撤销，应谨慎使用
	//
	// 返回:
	//   - error: 删除过程中的错误，成功时返回nil
	DeleteAllRequests() error

	// ExportRequests 将所有请求导出到文件
	//
	// 通常导出为JSON格式，包含所有存储的请求记录
	//
	// 返回:
	//   - string: 导出文件的路径
	//     例如: "/data/exports/requests_20230401_123456.json"
	//   - error: 导出过程中的错误，成功时返回nil
	ExportRequests() (string, error)

	// GetStats 获取存储系统的统计信息
	//
	// 返回包含各种统计数据的映射，如请求总数、各种HTTP方法的数量等
	//
	// 返回:
	//   - map[string]interface{}: 包含统计信息的映射
	//     例如:
	//     {
	//       "total_requests": 1250,
	//       "methods": {"GET": 450, "POST": 800},
	//       "latest_request": "2023-04-01T12:34:56Z"
	//     }
	//   - error: 获取过程中的错误，成功时返回nil
	GetStats() (map[string]interface{}, error)

	// Close 关闭存储并释放相关资源
	//
	// 应在程序退出前调用，以确保数据正确保存和资源释放
	// 对于文件系统存储可能只是刷新缓冲区，对于数据库存储则关闭连接
	//
	// 返回:
	//   - error: 关闭过程中的错误，成功时返回nil
	Close() error
}
