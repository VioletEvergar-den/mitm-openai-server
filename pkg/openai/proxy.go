package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// proxyService 实现代理OpenAI服务
type proxyService struct {
	config       Config
	httpClient   *http.Client
	supportedAPI map[string]bool
}

// 初始化函数，注册代理服务创建器
func init() {
	ProxyServiceCreator = newProxyService
}

// newProxyService 创建代理OpenAI服务
func newProxyService(config Config) Service {
	// 创建HTTP客户端，设置较长的超时时间（AI请求可能需要几分钟）
	client := &http.Client{
		Timeout: 10 * time.Minute, // AI请求可能需要较长时间
	}

	// 初始化支持的API
	supportedAPI := map[string]bool{
		"POST /chat/completions":     true,
		"GET /chat/completions":      true,
		"POST /completions":          true,
		"GET /completions":           true,
		"POST /embeddings":           true,
		"GET /embeddings":            true,
		"POST /images/generations":   true,
		"GET /images/generations":    true,
		"POST /audio/transcriptions": true,
		"POST /moderations":          true,
		"GET /models":                true,
		"POST /models":               true,
	}

	return &proxyService{
		config:       config,
		httpClient:   client,
		supportedAPI: supportedAPI,
	}
}

// Name 返回服务名称
func (s *proxyService) Name() string {
	return "OpenAI API Proxy"
}

// ServeOpenAISpec 提供OpenAI API规范
func (s *proxyService) ServeOpenAISpec(c *gin.Context) {
	// 尝试从目标服务器获取规范
	if s.config.TargetURL != "" {
		specURL := s.buildURL("/openapi.json")
		req, err := http.NewRequest(http.MethodGet, specURL, nil)
		if err == nil {
			// 添加认证
			s.addAuthHeader(req)

			// 发送请求
			resp, err := s.httpClient.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()

				// 读取响应体
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					var specData map[string]interface{}
					if json.Unmarshal(body, &specData) == nil {
						c.JSON(http.StatusOK, specData)
						return
					}
				}
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}

	// 如果无法从目标服务器获取，返回基本规范
	c.JSON(http.StatusOK, map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "OpenAI API (Proxy)",
			"description": "OpenAI API Proxy Server",
			"version":     "1.0.0",
		},
		"paths": map[string]interface{}{
			"/v1/chat/completions": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Creates a completion for the chat message",
				},
			},
			"/v1/completions": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Creates a completion for the provided prompt",
				},
			},
			"/v1/embeddings": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Creates an embedding vector representing the input text",
				},
			},
			"/v1/models": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Lists the currently available models",
				},
			},
		},
	})
}

// HandleRequest 处理API请求并代理转发
func (s *proxyService) HandleRequest(method, path string, headers, queryParams map[string]string, body []byte) (int, map[string]string, interface{}, error) {
	fmt.Printf("[ProxyService] 开始处理请求: %s %s\n", method, path)

	// 检查目标URL是否配置
	if s.config.TargetURL == "" {
		fmt.Printf("[ProxyService] 错误: 未配置目标URL\n")
		return http.StatusBadRequest, map[string]string{"Content-Type": "application/json"}, ErrorResp{
			Message: "未配置目标OpenAI API URL",
			Type:    "configuration_error",
		}, nil
	}

	// 模拟延迟（如果配置了延迟）
	if s.config.ResponseDelayMs > 0 {
		time.Sleep(time.Duration(s.config.ResponseDelayMs) * time.Millisecond)
	}

	// 标准化路径，确保以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 检查API是否支持
	apiKey := strings.ToUpper(method) + " " + path
	if !s.isAPISupported(apiKey) {
		return http.StatusNotFound, map[string]string{"Content-Type": "application/json"}, ErrorResp{
			Message: "不支持的API路径: " + path,
			Type:    "unsupported_api",
		}, nil
	}

	// 应用模型ID映射
	body = s.applyModelMapping(body)

	// 构建目标URL
	targetURL := s.buildURL(path)

	// 添加查询参数
	if len(queryParams) > 0 {
		queryStrings := make([]string, 0, len(queryParams))
		for k, v := range queryParams {
			queryStrings = append(queryStrings, fmt.Sprintf("%s=%s", k, v))
		}
		targetURL += "?" + strings.Join(queryStrings, "&")
	}

	fmt.Printf("[ProxyService] 目标URL: %s\n", targetURL)

	// 创建请求
	req, err := http.NewRequest(method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		return http.StatusInternalServerError, map[string]string{"Content-Type": "application/json"}, ErrorResp{
			Message: "创建代理请求失败: " + err.Error(),
			Type:    "proxy_error",
		}, err
	}

	// 添加请求头
	for k, v := range headers {
		// 跳过某些特殊的头，这些头由HTTP客户端自动处理
		if strings.EqualFold(k, "Content-Length") || strings.EqualFold(k, "Host") {
			continue
		}
		req.Header.Set(k, v)
	}

	// 确保Content-Type
	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 添加认证
	s.addAuthHeader(req)

	// 发送请求
	fmt.Printf("[ProxyService] 发送请求到目标服务器...\n")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[ProxyService] 请求失败: %v\n", err)
		return http.StatusBadGateway, map[string]string{"Content-Type": "application/json"}, ErrorResp{
			Message: "代理请求失败: " + err.Error(),
			Type:    "proxy_error",
		}, err
	}
	defer resp.Body.Close()

	fmt.Printf("[ProxyService] 收到响应: HTTP %d\n", resp.StatusCode)

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusBadGateway, map[string]string{"Content-Type": "application/json"}, ErrorResp{
			Message: "读取响应失败: " + err.Error(),
			Type:    "proxy_error",
		}, err
	}

	// 提取响应头
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	// 解析响应体
	var parsedBody interface{}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// 如果无法解析JSON，则作为字符串返回
			parsedBody = string(respBody)
		}
	} else {
		// 非JSON响应，返回字符串
		parsedBody = string(respBody)
	}

	// 返回响应
	return resp.StatusCode, respHeaders, parsedBody, nil
}

// 检查API是否支持
func (s *proxyService) isAPISupported(apiKey string) bool {
	fmt.Printf("DEBUG: 检查API支持: %s\n", apiKey)

	// 精确匹配
	if s.supportedAPI[apiKey] {
		fmt.Printf("DEBUG: API支持(精确匹配): %s\n", apiKey)
		return true
	}

	// 前缀匹配（适用于带参数的路径）
	for key := range s.supportedAPI {
		parts := strings.Split(key, " ")
		if len(parts) == 2 {
			method, path := parts[0], parts[1]
			if strings.HasPrefix(apiKey, method+" "+path+"/") {
				fmt.Printf("DEBUG: API支持(前缀匹配): %s 匹配 %s\n", apiKey, key)
				return true
			}
		}
	}

	fmt.Printf("DEBUG: API不支持: %s\n", apiKey)
	return false
}

// 构建完整URL
func (s *proxyService) buildURL(path string) string {
	baseURL := s.config.TargetURL

	// 处理URL斜杠，确保URL拼接正确
	if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(path, "/") {
		baseURL += "/"
	} else if strings.HasSuffix(baseURL, "/") && strings.HasPrefix(path, "/") {
		baseURL = baseURL[:len(baseURL)-1]
	}

	return baseURL + path
}

// 添加认证头
func (s *proxyService) addAuthHeader(req *http.Request) {
	switch strings.ToLower(s.config.TargetAuthType) {
	case "basic":
		if s.config.TargetUsername != "" {
			req.SetBasicAuth(s.config.TargetUsername, s.config.TargetPassword)
		}
	case "token", "bearer":
		if s.config.TargetToken != "" {
			token := s.config.TargetToken
			if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
				token = "Bearer " + token
			}
			req.Header.Set("Authorization", token)
		}
	}
}

// applyModelMapping 应用模型ID映射
// 将请求体中的模型名称替换为配置的实际模型ID
func (s *proxyService) applyModelMapping(body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	var reqBody map[string]interface{}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return body
	}

	if model, ok := reqBody["model"].(string); ok {
		fmt.Printf("[ModelMapping] 当前配置的映射表: %v\n", s.config.ModelMapping)
		if s.config.ModelMapping != nil && len(s.config.ModelMapping) > 0 {
			// 先尝试精确匹配
			if actualModel, exists := s.config.ModelMapping[model]; exists && actualModel != "" {
				fmt.Printf("[ModelMapping] 映射模型(精确匹配): %s -> %s\n", model, actualModel)
				reqBody["model"] = actualModel
				newBody, err := json.Marshal(reqBody)
				if err != nil {
					return body
				}
				return newBody
			}

			// 尝试不区分大小写匹配
			modelLower := strings.ToLower(model)
			for key, actualModel := range s.config.ModelMapping {
				if strings.ToLower(key) == modelLower && actualModel != "" {
					fmt.Printf("[ModelMapping] 映射模型(忽略大小写): %s -> %s (配置key: %s)\n", model, actualModel, key)
					reqBody["model"] = actualModel
					newBody, err := json.Marshal(reqBody)
					if err != nil {
						return body
					}
					return newBody
				}
			}

			fmt.Printf("[ModelMapping] 未找到模型 '%s' 的映射，可用映射: %v\n", model, s.config.ModelMapping)
		} else {
			fmt.Printf("[ModelMapping] 映射表为空或未配置，使用原始模型名: %s\n", model)
		}
	}

	return body
}

// UpdateConfig 更新服务配置
func (s *proxyService) UpdateConfig(config Config) {
	s.config = config
	fmt.Printf("[ProxyService] 配置已更新, ModelMapping: %v\n", config.ModelMapping)
}

// StreamHandleRequest 处理流式API请求（SSE）
// 返回通道以实时发送数据
func (s *proxyService) StreamHandleRequest(method, path string, headers, queryParams map[string]string, body []byte) (<-chan []byte, <-chan error) {
	dataChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		fmt.Printf("[ProxyService] 开始流式请求: %s %s\n", method, path)

		// 检查目标URL是否配置
		if s.config.TargetURL == "" {
			errChan <- fmt.Errorf("未配置目标OpenAI API URL")
			return
		}

		// 标准化路径
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		// 应用模型ID映射
		body = s.applyModelMapping(body)

		// 构建目标URL
		targetURL := s.buildURL(path)

		// 添加查询参数
		if len(queryParams) > 0 {
			queryStrings := make([]string, 0, len(queryParams))
			for k, v := range queryParams {
				queryStrings = append(queryStrings, fmt.Sprintf("%s=%s", k, v))
			}
			targetURL += "?" + strings.Join(queryStrings, "&")
		}

		// 创建请求
		req, err := http.NewRequest(method, targetURL, bytes.NewBuffer(body))
		if err != nil {
			errChan <- fmt.Errorf("创建代理请求失败: %v", err)
			return
		}

		// 添加请求头
		for k, v := range headers {
			if strings.EqualFold(k, "Content-Length") || strings.EqualFold(k, "Host") {
				continue
			}
			req.Header.Set(k, v)
		}

		// 确保Content-Type
		if len(body) > 0 && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// 添加认证
		s.addAuthHeader(req)

		// 发送请求
		fmt.Printf("[ProxyService] 发送流式请求到: %s\n", targetURL)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("代理请求失败: %v", err)
			return
		}
		defer resp.Body.Close()

		// 读取响应
		buffer := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				dataChan <- buffer[:n]
			}
			if err != nil {
				break
			}
		}

		fmt.Printf("[ProxyService] 流式请求完成\n")
	}()

	return dataChan, errChan
}
