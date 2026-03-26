package openai

import (
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
)

// NewService 根据配置创建适当的OpenAI服务
// 根据ProxyMode配置，返回代理服务或模拟服务
func NewService(config Config) Service {
	if config.ProxyMode && config.TargetURL != "" {
		// 使用代理模式
		return ProxyServiceCreator(config)
	}

	// 使用模拟模式
	return MockServiceCreator(config)
}

// GetServiceByName 根据服务名称获取相应的服务实现
func GetServiceByName(name string, config Config) Service {
	switch name {
	case "mock", "模拟":
		return MockServiceCreator(config)
	case "proxy", "代理":
		return ProxyServiceCreator(config)
	default:
		// 默认返回模拟服务
		return MockServiceCreator(config)
	}
}

// 全局处理器实例，用于直接保存请求
var globalHandler *Handler

// InitGlobalHandler 初始化全局处理器
// 这个函数应该在服务启动时被调用一次
// 返回创建的Handler实例，供调用方使用
func InitGlobalHandler(storage storage.Storage, service Service) *Handler {
	globalHandler = NewHandler(storage, service)
	return globalHandler
}

// UpdateGlobalHandlerConfig 更新全局处理器的服务配置
// 当代理模式配置改变时调用此函数切换服务实例
func UpdateGlobalHandlerConfig(config Config) {
	if globalHandler != nil {
		globalHandler.UpdateServiceConfig(config)
	}
}

// GetGlobalHandler 获取全局处理器实例
func GetGlobalHandler() *Handler {
	return globalHandler
}

// SaveRequest 全局函数，用于保存请求记录
// 这个函数可以从包外部调用，而不需要直接访问处理器实例
// 参数:
//   - userID: 用户ID
//   - request: 要保存的请求记录
//
// 返回:
//   - error: 如果保存失败，返回错误信息
func SaveRequest(userID int64, request *storage.Request) error {
	if globalHandler == nil {
		return nil // 如果全局处理器未初始化，静默忽略
	}
	return globalHandler.SaveRequest(userID, request)
}
