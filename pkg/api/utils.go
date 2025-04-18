package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ensureDirectory 确保目录存在
func ensureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// formatTimestamp 格式化时间戳
func formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// buildFilePath 构建文件路径
func buildFilePath(baseDir, id string, ext string) string {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return filepath.Join(baseDir, fmt.Sprintf("%s%s", id, ext))
}

// getIPAddress 获取客户端IP地址
func getIPAddress(r *http.Request) string {
	// 尝试从X-Forwarded-For头获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP头获取
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// 直接从请求中获取
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// sanitizeFilename 净化文件名
func sanitizeFilename(name string) string {
	// 移除不安全的字符
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)

	// 确保文件名不为空
	if name == "" {
		name = "unnamed"
	}

	return name
}

// sendProxyRequest 发送请求到目标API并返回响应
func sendProxyRequest(reqMethod, targetURL, path string, headers map[string]string, body []byte, config ServerConfig) (*ProxyResponse, error) {
	// 构建完整URL
	fullURL := targetURL
	if !strings.HasSuffix(targetURL, "/") && !strings.HasPrefix(path, "/") {
		fullURL += "/"
	} else if strings.HasSuffix(targetURL, "/") && strings.HasPrefix(path, "/") {
		fullURL = fullURL[:len(fullURL)-1]
	}
	fullURL += path

	// 创建请求
	req, err := http.NewRequest(reqMethod, fullURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建代理请求失败: %w", err)
	}

	// 添加请求头
	for k, v := range headers {
		// 跳过一些特定的头，这些头由HTTP客户端自动处理
		if strings.ToLower(k) == "content-length" || strings.ToLower(k) == "host" {
			continue
		}
		req.Header.Set(k, v)
	}

	// 添加目标服务认证
	switch config.TargetAuthType {
	case "basic":
		req.SetBasicAuth(config.TargetUsername, config.TargetPassword)
	case "token":
		token := config.TargetToken
		if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = "Bearer " + token
		}
		req.Header.Set("Authorization", token)
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送代理请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取代理响应失败: %w", err)
	}

	// 将响应体解析为JSON（如果是JSON格式）
	var respData interface{}
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(respBody, &respData); err != nil {
			// 如果不是有效的JSON，就使用原始响应
			respData = string(respBody)
		}
	} else {
		// 非JSON响应直接使用字符串
		respData = string(respBody)
	}

	// 构建响应对象
	proxyResp := &ProxyResponse{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       respData,
	}

	// 复制响应头
	for k, v := range resp.Header {
		if len(v) > 0 {
			proxyResp.Headers[k] = v[0]
		}
	}

	return proxyResp, nil
}
