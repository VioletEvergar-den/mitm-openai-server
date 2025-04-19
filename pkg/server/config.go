package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	// 应用存储路径设置
	if config.StoragePath != "" {
		// 存储路径保存在配置中，实际应用在启动时处理
		server.storagePath = config.StoragePath
	}
}
