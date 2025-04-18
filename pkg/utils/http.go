package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetClientIP 获取客户端IP地址
//
// 参数:
//   - r: HTTP请求对象
//
// 返回:
//   - string: 客户端IP地址
func GetClientIP(r *http.Request) string {
	// 首先尝试从X-Forwarded-For头获取
	// X-Forwarded-For格式: client, proxy1, proxy2, ...
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0]) // 返回最左边的IP（客户端）
		}
	}

	// 然后尝试从X-Real-IP头获取（通常由反向代理设置）
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// 最后从请求的RemoteAddr获取
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// 如果解析失败（可能格式不是IP:port），直接返回原始值
		return r.RemoteAddr
	}
	return ip
}

// HeadersToMap 将HTTP头转换为map
//
// 参数:
//   - headers: HTTP头
//   - flatten: 是否将多值头压缩为单一字符串（用逗号分隔）
//
// 返回:
//   - map[string]string: 头部map
func HeadersToMap(headers http.Header, flatten bool) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) == 0 {
			continue // 跳过空值
		}

		if flatten && len(v) > 0 {
			result[k] = strings.Join(v, ", ")
		} else if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// HeadersToMapArray 将HTTP头转换为map，保留多值
//
// 参数:
//   - headers: HTTP头
//
// 返回:
//   - map[string][]string: 头部map，保留所有值
func HeadersToMapArray(headers http.Header) map[string][]string {
	result := make(map[string][]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v
		}
	}
	return result
}

// QueryToMap 将URL查询参数转换为map
//
// 参数:
//   - values: URL查询参数
//   - flatten: 是否将多值参数压缩为单一字符串
//
// 返回:
//   - map[string]string: 查询参数map
func QueryToMap(values map[string][]string, flatten bool) map[string]string {
	result := make(map[string]string)
	for k, v := range values {
		if len(v) == 0 {
			continue // 跳过空值
		}

		if flatten && len(v) > 0 {
			result[k] = strings.Join(v, ", ")
		} else if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// QueryToMapArray 将URL查询参数转换为map，保留多值
//
// 参数:
//   - values: URL查询参数
//
// 返回:
//   - map[string][]string: 查询参数map，保留所有值
func QueryToMapArray(values map[string][]string) map[string][]string {
	result := make(map[string][]string)
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v
		}
	}
	return result
}

// SendHTTPRequest 发送HTTP请求并返回响应
//
// 参数:
//   - method: HTTP方法（GET、POST等）
//   - url: 请求URL
//   - headers: 请求头
//   - body: 请求体
//   - timeout: 超时时间（秒）
//
// 返回:
//   - *http.Response: HTTP响应
//   - []byte: 响应体
//   - error: 请求过程中的错误，如果成功则为nil
func SendHTTPRequest(method, url string, headers map[string]string, body []byte, timeout int) (*http.Response, []byte, error) {
	// 创建请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 添加请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 如果没有设置Content-Type且存在请求体，设置默认Content-Type
	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 创建带超时的客户端
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return resp, respBody, nil
}

// ParseJSONBody 解析JSON请求体
//
// 参数:
//   - body: 请求体字节数组
//
// 返回:
//   - interface{}: 解析后的JSON对象
//   - error: 解析过程中的错误，如果成功则为nil
func ParseJSONBody(body []byte) (interface{}, error) {
	if len(body) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return result, nil
}

// IsJSONContentType 检查Content-Type是否为JSON
//
// 参数:
//   - contentType: Content-Type头的值
//
// 返回:
//   - bool: 如果是JSON类型则为true，否则为false
func IsJSONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// SendProxyRequest 发送请求到目标API并返回响应
//
// 参数:
//   - reqMethod: HTTP方法（GET、POST等）
//   - targetURL: 目标URL
//   - path: API路径
//   - headers: 请求头
//   - body: 请求体
//   - authType: 认证类型（none、basic、token）
//   - username: 基本认证用户名
//   - password: 基本认证密码
//   - token: 令牌认证的令牌
//
// 返回:
//   - map[string]interface{}: 包含状态码、头部和响应体的响应信息
//   - error: 请求过程中的错误，如果成功则为nil
func SendProxyRequest(reqMethod, targetURL, path string, headers map[string]string, body []byte,
	authType, username, password, token string) (map[string]interface{}, error) {
	// 构建完整URL，处理斜杠问题
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

	// 添加请求头（排除某些特定头部）
	for k, v := range headers {
		// 跳过一些特定的头，这些头由HTTP客户端自动处理
		if strings.EqualFold(k, "Content-Length") || strings.EqualFold(k, "Host") {
			continue
		}
		req.Header.Set(k, v)
	}

	// 添加目标服务认证
	switch strings.ToLower(authType) {
	case "basic":
		req.SetBasicAuth(username, password)
	case "token", "bearer":
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

	// 处理响应头
	respHeaders := HeadersToMap(resp.Header, true)

	// 解析响应体（如果是JSON）
	var parsedBody interface{}
	contentType := resp.Header.Get("Content-Type")
	if IsJSONContentType(contentType) {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// 如果解析失败，使用原始响应体
			parsedBody = string(respBody)
		}
	} else {
		// 非JSON响应，返回字符串
		parsedBody = string(respBody)
	}

	// 构建响应结果
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     respHeaders,
		"body":        parsedBody,
	}

	return result, nil
}
