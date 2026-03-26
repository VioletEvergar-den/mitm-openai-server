package openai

// Config 定义OpenAI服务配置
type Config struct {
	// 是否启用这个服务
	Enabled bool `json:"enabled"`

	// 模拟响应的延迟毫秒数
	ResponseDelayMs int `json:"response_delay_ms"`

	// API密钥验证设置
	APIKeyAuth bool `json:"api_key_auth"`

	// 用于验证的API密钥
	APIKey string `json:"api_key"`

	// 自定义响应内容
	CustomResponse string `json:"custom_response"`

	// 模型列表文件路径
	ModelsFile string `json:"models_file"`

	// 是否启用代理模式
	ProxyMode bool `json:"proxy_mode"`

	// 目标API URL
	TargetURL string `json:"target_url"`

	// 目标认证类型 (none, basic, token)
	TargetAuthType string `json:"target_auth_type"`

	// 目标用户名（Basic认证）
	TargetUsername string `json:"target_username"`

	// 目标密码（Basic认证）
	TargetPassword string `json:"target_password"`

	// 目标Token（Token认证）
	TargetToken string `json:"target_token"`

	// Model ID 映射 (自定义模型名 -> 实际模型ID)
	ModelMapping map[string]string `json:"model_mapping"`
}

// DefaultConfig 返回默认的OpenAI API服务配置
func DefaultConfig() Config {
	return Config{
		Enabled:         true,
		ResponseDelayMs: 100,
		APIKeyAuth:      true,
		APIKey:          "sk-mock-openai-key",
		ProxyMode:       false,
		TargetAuthType:  "none",
		ModelMapping:    make(map[string]string),
	}
}
