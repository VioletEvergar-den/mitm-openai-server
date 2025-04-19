package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UserConfig 存储用户的配置信息
type UserConfig struct {
	// 代理设置
	ProxyMode      bool   `json:"proxy_mode"`
	TargetURL      string `json:"target_url"`
	TargetAuthType string `json:"target_auth_type"`
	TargetUsername string `json:"target_username,omitempty"`
	TargetPassword string `json:"target_password,omitempty"`
	TargetToken    string `json:"target_token,omitempty"`

	// 数据存储设置
	StoragePath string `json:"storage_path,omitempty"`

	// UI认证设置
	UIUsername string `json:"ui_username,omitempty"`
	UIPassword string `json:"ui_password,omitempty"`

	// 其他用户设置可以在此添加
}

// ConfigManager 管理用户配置的读取和保存
type ConfigManager struct {
	configDir  string
	configFile string
}

// NewConfigManager 创建一个新的配置管理器
func NewConfigManager() (*ConfigManager, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户主目录: %v", err)
	}

	// 创建配置目录
	configDir := filepath.Join(homeDir, ".mitm-openai-server")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建配置目录: %v", err)
	}

	return &ConfigManager{
		configDir:  configDir,
		configFile: filepath.Join(configDir, "config.json"),
	}, nil
}

// LoadConfig 从文件加载用户配置
func (cm *ConfigManager) LoadConfig() (UserConfig, error) {
	var config UserConfig

	// 检查配置文件是否存在
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		// 如果不存在，返回默认配置
		return config, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// ApplyConfig 将用户配置应用到服务器配置
func (cm *ConfigManager) ApplyConfig(config UserConfig, server *Server) {
	fmt.Printf("\n=================== 配置应用过程 ===================\n")
	fmt.Printf("配置中的UI用户名: %s\n", config.UIUsername)
	fmt.Printf("配置中的UI密码: %s (长度: %d)\n", config.UIPassword, len(config.UIPassword))
	fmt.Printf("服务器当前UI密码: %s (长度: %d)\n", server.config.UIPassword, len(server.config.UIPassword))

	// 应用代理设置
	server.config.ProxyMode = config.ProxyMode
	server.config.TargetURL = config.TargetURL
	server.config.TargetAuthType = config.TargetAuthType

	// 只有在有值时才更新这些字段
	if config.TargetUsername != "" {
		server.config.TargetUsername = config.TargetUsername
	}
	if config.TargetPassword != "" {
		server.config.TargetPassword = config.TargetPassword
	}
	if config.TargetToken != "" {
		server.config.TargetToken = config.TargetToken
	}

	// 应用UI认证设置
	if config.UIUsername != "" {
		server.config.UIUsername = strings.TrimSpace(config.UIUsername)
		fmt.Printf("应用了UI用户名: %s\n", server.config.UIUsername)
	}

	// 确保密码被正确处理和设置
	if config.UIPassword != "" {
		// 去除可能的前后空格
		server.config.UIPassword = strings.TrimSpace(config.UIPassword)
		// 如果有保存的密码，禁用随机密码生成
		server.config.GenerateUIAuth = false
		fmt.Printf("应用了UI密码: %s (长度: %d)\n", server.config.UIPassword, len(server.config.UIPassword))
	}

	// 应用存储路径设置
	if config.StoragePath != "" {
		// 存储路径保存在配置中，实际应用在启动时处理
		server.storagePath = config.StoragePath
	}

	fmt.Printf("应用后的服务器UI密码: %s (长度: %d)\n", server.config.UIPassword, len(server.config.UIPassword))
	fmt.Printf("=================== 配置应用结束 ===================\n\n")
}

// SaveConfig 将用户配置保存到文件
func (cm *ConfigManager) SaveConfig(config UserConfig) error {
	// 将配置转换为JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(cm.configFile, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// SaveUICredentials 保存UI认证凭据到配置文件
func (cm *ConfigManager) SaveUICredentials(username, password string) error {
	// 加载当前配置
	config, err := cm.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 更新认证凭据
	config.UIUsername = username
	config.UIPassword = password

	// 保存配置
	return cm.SaveConfig(config)
}
