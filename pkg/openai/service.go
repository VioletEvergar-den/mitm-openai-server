package openai

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
