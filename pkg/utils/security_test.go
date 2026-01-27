package utils

import (
	"testing"
)

func TestIsURLSafe(t *testing.T) {
	tests := []struct {
		name      string
		targetURL string
		serverAddr string
		expectErr bool
	}{
		{
			name:      "安全的外部URL",
			targetURL: "https://api.openai.com",
			serverAddr: "localhost:8081",
			expectErr: false,
		},
		{
			name:      "本地回环地址IPv4",
			targetURL: "http://127.0.0.1:8080",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "本地回环地址localhost",
			targetURL: "http://localhost:8080",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "本地回环地址IPv6",
			targetURL: "http://[::1]:8080",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "私有网络地址10.x",
			targetURL: "http://10.0.0.1:8080",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "私有网络地址172.x",
			targetURL: "http://172.16.0.1:8080",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "私有网络地址192.x",
			targetURL: "http://192.168.1.1:8080",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "云服务元数据地址",
			targetURL: "http://169.254.169.254",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "与服务器地址相同",
			targetURL: "http://localhost:8081",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
		{
			name:      "无效URL格式",
			targetURL: "not-a-valid-url",
			serverAddr: "localhost:8081",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsURLSafe(tt.targetURL, tt.serverAddr)
			if tt.expectErr && err == nil {
				t.Errorf("期望错误但没有错误")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("期望没有错误但得到了错误: %v", err)
			}
		})
	}
}

func TestIsLoopbackAddress(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.1:8080", true},
		{"localhost", true},
		{"localhost:8080", true},
		{"::1", true},
		{"::1:8080", true},
		{"192.168.1.1", false},
		{"192.168.1.1:8080", false},
		{"api.openai.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isLoopbackAddress(tt.host)
			if result != tt.expected {
				t.Errorf("isLoopbackAddress(%s) = %v, 期望 %v", tt.host, result, tt.expected)
			}
		})
	}
}

func TestIsPrivateNetwork(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"10.0.0.1", true},
		{"10.0.0.1:8080", true},
		{"172.16.0.1", true},
		{"172.16.0.1:8080", true},
		{"192.168.1.1", true},
		{"192.168.1.1:8080", true},
		{"127.0.0.1", true},
		{"127.0.0.1:8080", true},
		{"169.254.1.1", true},
		{"169.254.1.1:8080", true},
		{"8.8.8.8", false},
		{"8.8.8.8:53", false},
		{"api.openai.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isPrivateNetwork(tt.host)
			if result != tt.expected {
				t.Errorf("isPrivateNetwork(%s) = %v, 期望 %v", tt.host, result, tt.expected)
			}
		})
	}
}

func TestIsCloudMetadataAddress(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"169.254.169.254", true},
		{"169.254.169.254:80", true},
		{"169.254.0.1", true},
		{"169.254.0.1:80", true},
		{"metadata", true},
		{"metadata:80", true},
		{"metadata.google.internal", true},
		{"metadata.google.internal:80", true},
		{"100.100.100.200", true},
		{"100.100.100.200:80", true},
		{"127.0.0.1", false},
		{"127.0.0.1:8080", false},
		{"api.openai.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isCloudMetadataAddress(tt.host)
			if result != tt.expected {
				t.Errorf("isCloudMetadataAddress(%s) = %v, 期望 %v", tt.host, result, tt.expected)
			}
		})
	}
}