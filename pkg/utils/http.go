package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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

// ResponseRecorder 实现了http.ResponseWriter接口，并记录响应数据
// 用于捕获API响应以便存储和分析
type ResponseRecorder struct {
	http.ResponseWriter
	Body   *bytes.Buffer // 响应体的缓冲区
	status int           // HTTP状态码
}

// NewResponseRecorder 创建一个新的ResponseRecorder
// 参数:
//   - w: 原始的ResponseWriter
//
// 返回:
//   - *ResponseRecorder: 包装后的ResponseWriter
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		Body:           &bytes.Buffer{},
		status:         http.StatusOK, // 默认状态码为200
	}
}

// Write 实现http.ResponseWriter接口的Write方法
// 将数据写入原始ResponseWriter的同时，也写入缓冲区
// 参数:
//   - b: 要写入的字节切片
//
// 返回:
//   - int: 写入的字节数
//   - error: 如果写入失败，返回错误
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	// 同时写入到原始ResponseWriter和缓冲区
	r.Body.Write(b)
	return r.ResponseWriter.Write(b)
}

// WriteHeader 实现http.ResponseWriter接口的WriteHeader方法
// 记录状态码并调用原始ResponseWriter的WriteHeader方法
// 参数:
//   - statusCode: HTTP状态码
func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Status 返回记录的HTTP状态码
// 返回:
//   - int: HTTP状态码
func (r *ResponseRecorder) Status() int {
	return r.status
}

// Header 返回HTTP响应头
// 返回:
//   - http.Header: 响应头
func (r *ResponseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

// Flush 实现http.Flusher接口
func (r *ResponseRecorder) Flush() {
	// 如果原始ResponseWriter实现了Flusher接口，则调用它的Flush方法
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现http.Hijacker接口
func (r *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// 如果原始ResponseWriter实现了Hijacker接口，则调用它的Hijack方法
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("underlying ResponseWriter does not implement http.Hijacker")
}

// CloseNotify 实现http.CloseNotifier接口
func (r *ResponseRecorder) CloseNotify() <-chan bool {
	// 如果原始ResponseWriter实现了CloseNotifier接口，则调用它的CloseNotify方法
	if closer, ok := r.ResponseWriter.(http.CloseNotifier); ok {
		return closer.CloseNotify()
	}
	return make(chan bool, 1)
}

// Push 实现http.Pusher接口
// 用于支持HTTP/2的服务器推送功能
// 参数:
//   - target: 推送的目标路径
//   - opts: 推送选项
//
// 返回:
//   - error: 如果底层ResponseWriter不支持Push，返回错误
func (r *ResponseRecorder) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return fmt.Errorf("底层ResponseWriter不支持Push方法")
}

// Pusher 实现gin.ResponseWriter的Pusher方法
// 返回底层ResponseWriter的http.Pusher
// 返回:
//   - http.Pusher: 底层ResponseWriter的Pusher，如果支持的话
func (r *ResponseRecorder) Pusher() http.Pusher {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

// ReadFrom 实现io.ReaderFrom接口
func (r *ResponseRecorder) ReadFrom(src io.Reader) (n int64, err error) {
	// 如果原始ResponseWriter实现了ReaderFrom接口，则调用它的ReadFrom方法
	if readFrom, ok := r.ResponseWriter.(io.ReaderFrom); ok {
		return readFrom.ReadFrom(src)
	}

	// 否则手动实现
	buf := make([]byte, 32*1024) // 32KB缓冲区
	var total int64

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			// 写入缓冲区
			r.Body.Write(buf[:nr])

			// 写入原始ResponseWriter
			nw, ew := r.ResponseWriter.Write(buf[:nr])
			if ew != nil {
				err = ew
				break
			}
			if nw != nr {
				err = io.ErrShortWrite
				break
			}
			total += int64(nw)
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}

	return total, err
}

// Size 返回已写入响应体的字节数
// 返回:
//   - int: 已写入的字节数
func (r *ResponseRecorder) Size() int {
	return r.Body.Len()
}

// WriteString 将字符串写入响应体
// 参数:
//   - s: 要写入的字符串
//
// 返回:
//   - int: 写入的字节数
//   - error: 如果写入失败，返回错误
func (r *ResponseRecorder) WriteString(s string) (int, error) {
	r.Body.WriteString(s)
	return r.ResponseWriter.Write([]byte(s))
}

// Written 返回响应体是否已经写入
// 返回:
//   - bool: 如果响应体已写入，返回true
func (r *ResponseRecorder) Written() bool {
	return r.Body.Len() > 0
}

// WriteHeaderNow 强制写入HTTP头（状态码+头部）
func (r *ResponseRecorder) WriteHeaderNow() {
	if !r.Written() {
		r.WriteHeader(r.status)
	}
}
