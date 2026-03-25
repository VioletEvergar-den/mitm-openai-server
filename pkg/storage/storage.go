package storage

// Storage 定义了存储请求和响应的接口
// 该接口提供了对HTTP请求和响应数据的持久化存储、检索和管理功能
// 现在支持用户隔离，每个用户只能访问自己的数据
type Storage interface {
	// ==================== 用户管理 ====================

	// CreateUser 创建新用户
	CreateUser(user *User) error

	// GetUserByUsername 根据用户名获取用户信息
	GetUserByUsername(username string) (*User, error)

	// GetUserByID 根据用户ID获取用户信息
	GetUserByID(userID int64) (*User, error)

	// UpdateUser 更新用户信息
	UpdateUser(user *User) error

	// UpdateUserConfig 更新用户配置
	UpdateUserConfig(userID int64, config *UserConfig) error

	// UpdateUserLastLogin 更新用户最后登录时间
	UpdateUserLastLogin(userID int64) error

	// ValidateUserCredentials 验证用户凭据
	ValidateUserCredentials(username, password string) (*User, error)

	// ListUsers 获取用户列表（仅管理员可用）
	ListUsers(limit, offset int) ([]*User, error)

	// ==================== 请求管理（用户隔离） ====================

	// SaveRequest 保存HTTP请求及其响应到存储中
	// 现在需要指定用户ID，确保数据归属正确
	//
	// 参数:
	//   - userID: 请求所属的用户ID
	//   - req: 包含请求和响应数据的Request对象
	//
	// 返回:
	//   - error: 保存过程中的错误，成功时返回nil
	SaveRequest(userID int64, req *Request) error

	// GetRequestByID 根据ID获取特定用户的请求
	// 只能获取属于指定用户的请求，确保数据隔离
	//
	// 参数:
	//   - userID: 用户ID，用于权限验证
	//   - id: 请求的唯一标识符
	//
	// 返回:
	//   - *Request: 请求对象，如果找到且属于该用户
	//   - error: 检索过程中的错误，或者当请求不存在/无权限时返回错误
	GetRequestByID(userID int64, id string) (*Request, error)

	// GetRequestByIDOnly 仅根据ID获取请求（不验证用户归属）
	// 用于root用户或管理员查看任意请求
	//
	// 参数:
	//   - id: 请求的唯一标识符
	//
	// 返回:
	//   - *Request: 请求对象，如果找到
	//   - error: 检索过程中的错误，或者当请求不存在时返回错误
	GetRequestByIDOnly(id string) (*Request, error)

	// GetUserRequests 获取指定用户的所有请求，支持分页
	//
	// 参数:
	//   - userID: 用户ID，只返回该用户的请求
	//   - limit: 每页返回的最大记录数
	//   - offset: 分页的起始位置(跳过的记录数)
	//
	// 返回:
	//   - []*Request: 请求对象的切片，如果没有记录则返回空切片
	//   - error: 检索过程中的错误，成功时返回nil
	GetUserRequests(userID int64, limit, offset int) ([]*Request, error)

	// DeleteRequest 删除指定用户的特定请求
	// 只能删除属于指定用户的请求
	//
	// 参数:
	//   - userID: 用户ID，用于权限验证
	//   - id: 要删除的请求的唯一标识符
	//
	// 返回:
	//   - error: 删除过程中的错误，成功时返回nil
	DeleteRequest(userID int64, id string) error

	// DeleteAllUserRequests 删除指定用户的所有请求
	//
	// 参数:
	//   - userID: 用户ID
	//
	// 返回:
	//   - error: 删除过程中的错误，成功时返回nil
	DeleteAllUserRequests(userID int64) error

	// GetUserStats 获取指定用户的统计信息
	//
	// 参数:
	//   - userID: 用户ID
	//
	// 返回:
	//   - map[string]interface{}: 统计信息的键值对
	//   - error: 获取过程中的错误，成功时返回nil
	GetUserStats(userID int64) (map[string]interface{}, error)

	// ExportUserRequests 导出指定用户的请求数据
	//
	// 参数:
	//   - userID: 用户ID
	//
	// 返回:
	//   - string: 导出文件的路径
	//   - error: 导出过程中的错误，成功时返回nil
	ExportUserRequests(userID int64) (string, error)

	// ==================== 管理员专用方法 ====================

	// GetAllRequests 获取所有请求（仅管理员可用）
	//
	// 参数:
	//   - limit: 每页返回的最大记录数
	//   - offset: 分页的起始位置(跳过的记录数)
	//
	// 返回:
	//   - []*Request: 请求对象的切片
	//   - error: 检索过程中的错误，成功时返回nil
	GetAllRequests(limit, offset int) ([]*Request, error)

	// DeleteAllRequests 删除所有请求（仅管理员可用）
	//
	// 返回:
	//   - error: 删除过程中的错误，成功时返回nil
	DeleteAllRequests() error

	// GetStats 获取全局统计信息（仅管理员可用）
	//
	// 返回:
	//   - map[string]interface{}: 统计信息的键值对
	//   - error: 获取过程中的错误，成功时返回nil
	GetStats() (map[string]interface{}, error)

	// ExportRequests 导出所有请求数据（仅管理员可用）
	//
	// 返回:
	//   - string: 导出文件的路径
	//   - error: 导出过程中的错误，成功时返回nil
	ExportRequests() (string, error)

	// ==================== 数据库管理 ====================

	// Close 关闭存储连接
	Close() error

	// InitDatabase 初始化数据库表结构
	InitDatabase() error
}

// 为了向后兼容，保留一些旧方法的别名
// 这些方法将在后续版本中移除

// Deprecated: 使用 GetUserRequests 替代
// GetRequestsByUser 是 GetUserRequests 的别名
func GetRequestsByUser(s Storage, userID int64, limit, offset int) ([]*Request, error) {
	return s.GetUserRequests(userID, limit, offset)
}
