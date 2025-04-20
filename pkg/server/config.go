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

	// 是否为首次启动（保存随机生成的密码后设置为false）
	FirstRun bool `json:"first_run,omitempty"`

	// 其他用户设置可以在此添加
}

// ConfigManager 管理用户配置的读取和保存
type ConfigManager struct {
	configDir   string
	configFile  string
	loginFile   string
	defaultUser string
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
		configDir:   configDir,
		configFile:  filepath.Join(configDir, "config.json"),
		loginFile:   "login.json", // 当前目录下的登录文件
		defaultUser: "root",       // 默认用户名
	}, nil
}

// LoadConfig 从文件加载用户配置
func (cm *ConfigManager) LoadConfig() (UserConfig, error) {
	var config UserConfig

	// 首先检查配置目录中的配置文件
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		// 如果配置文件不存在，检查login.json是否存在
		if cm.loginFileExists() {
			// 从login.json中加载凭据
			return cm.loadFromLoginFile()
		}
		// 如果都不存在，创建默认配置但不立即保存
		config.FirstRun = false // 避免无限循环的关键
		config.UIUsername = cm.defaultUser
		config.UIPassword = "" // 不设置密码，让调用方决定如何处理
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

	// 如果配置中没有设置用户名，使用默认用户名
	if config.UIUsername == "" {
		config.UIUsername = cm.defaultUser
	}

	return config, nil
}

// loginFileExists 检查login.json文件是否存在
func (cm *ConfigManager) loginFileExists() bool {
	_, err := os.Stat(cm.loginFile)
	return !os.IsNotExist(err)
}

// loadFromLoginFile 从login.json加载凭据
func (cm *ConfigManager) loadFromLoginFile() (UserConfig, error) {
	var config UserConfig
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// 读取login.json
	data, err := os.ReadFile(cm.loginFile)
	if err != nil {
		return config, fmt.Errorf("读取login.json失败: %v", err)
	}

	// 解析JSON
	if err := json.Unmarshal(data, &loginData); err != nil {
		return config, fmt.Errorf("解析login.json失败: %v", err)
	}

	// 设置配置
	config.UIUsername = loginData.Username
	config.UIPassword = loginData.Password
	config.FirstRun = false // 标记为非首次运行

	// 打印安全的凭据信息 (不显示密码)
	fmt.Printf("从login.json加载的凭据: 用户名=%s, 密码=•••••••• (长度: %d)\n",
		config.UIUsername, len(config.UIPassword))

	// 不要在这里调用SaveConfig以避免循环
	// 如果需要将登录凭据同步到配置文件，应该由调用方处理

	return config, nil
}

// ApplyConfig 将用户配置应用到服务器配置
func (cm *ConfigManager) ApplyConfig(config UserConfig, server *Server) {
	fmt.Printf("\n=================== 配置应用过程 ===================\n")
	fmt.Printf("配置中的UI用户名: %s\n", config.UIUsername)
	fmt.Printf("配置中的UI密码: •••••••• (长度: %d)\n", len(config.UIPassword))
	fmt.Printf("服务器当前UI密码: •••••••• (长度: %d)\n", len(server.config.UIPassword))

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
		fmt.Printf("应用了UI密码: •••••••• (长度: %d)\n", len(server.config.UIPassword))
	}

	// 应用存储路径设置
	if config.StoragePath != "" {
		// 存储路径保存在配置中，实际应用在启动时处理
		server.storagePath = config.StoragePath
	}

	fmt.Printf("应用后的服务器UI密码: •••••••• (长度: %d)\n", len(server.config.UIPassword))
	fmt.Printf("=================== 配置应用结束 ===================\n\n")
}

// SaveConfig 将用户配置保存到文件
func (cm *ConfigManager) SaveConfig(config UserConfig) error {
	// 创建配置目录（如果不存在）
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %v", err)
	}

	// 如果是首次运行且有密码，同时创建login.json
	if config.FirstRun && config.UIPassword != "" {
		// 创建login.json (不依赖updateLoginFile方法以避免循环)
		loginConfig := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: config.UIUsername,
			Password: config.UIPassword,
		}

		// 序列化为JSON
		data, err := json.MarshalIndent(loginConfig, "", "  ")
		if err != nil {
			fmt.Printf("警告: 无法序列化登录配置: %v\n", err)
		} else {
			// 写入文件
			if err := os.WriteFile(cm.loginFile, data, 0644); err != nil {
				fmt.Printf("警告: 无法写入登录文件: %v\n", err)
			} else {
				fmt.Printf("成功创建登录文件，用户名=%s\n", config.UIUsername)
			}
		}

		// 保存后标记为非首次启动，防止下次再次生成
		config.FirstRun = false
	}

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
	config.FirstRun = true // 标记为首次运行，以便保存到login.json

	// 保存配置
	return cm.SaveConfig(config)
}

// SavePassword 直接设置密码并更新配置与login.json
func (cm *ConfigManager) SavePassword(password string) error {
	// 加载当前配置时使用一个全新的配置对象，避免循环加载
	config := UserConfig{
		UIUsername: cm.defaultUser,
		UIPassword: password,
		FirstRun:   true, // 标记为需要创建login.json
	}

	// 尝试加载现有配置，但如果失败则使用默认配置
	existingConfig, err := cm.LoadConfig()
	if err == nil && existingConfig.UIUsername != "" {
		// 只更新密码，保留其他配置
		existingConfig.UIPassword = password
		existingConfig.FirstRun = true
		return cm.SaveConfig(existingConfig)
	}

	// 如果加载失败或用户名为空，使用默认配置
	return cm.SaveConfig(config)
}

// updateLoginFile 更新login.json文件
// 完全独立的方法，不依赖于其他配置方法，仅创建基本的login.json结构
func (cm *ConfigManager) updateLoginFile() error {
	// 检查登录文件是否存在
	if _, err := os.Stat(cm.loginFile); os.IsNotExist(err) {
		// 文件不存在，创建目录
		dir := filepath.Dir(cm.loginFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("无法创建登录文件目录: %v", err)
		}
	}

	// 创建基础登录配置，保持简单
	loginConfig := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: cm.defaultUser,
		Password: "", // 这里不设置密码，由调用方设置
	}

	// 将配置写入文件
	data, err := json.MarshalIndent(loginConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化登录配置失败: %v", err)
	}

	// 将数据写入文件
	if err := os.WriteFile(cm.loginFile, data, 0644); err != nil {
		return fmt.Errorf("写入登录文件失败: %v", err)
	}

	fmt.Printf("已创建初始登录文件模板: 用户名=%s (无密码)\n", loginConfig.Username)

	return nil
}
