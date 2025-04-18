package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 测试mock服务创建
func TestNewMockService(t *testing.T) {
	// 创建基础配置
	config := Config{
		Enabled:         true,
		ResponseDelayMs: 0, // 测试中不延迟
		APIKeyAuth:      true,
		APIKey:          "test-key",
	}

	// 创建服务
	service := newMockService(config)

	// 验证服务实例
	assert.NotNil(t, service, "服务实例不应为nil")
	assert.Equal(t, "MockOpenAIService", service.Name(), "服务名称应为MockOpenAIService")
}

// 测试模拟服务的API处理
func TestMockService_HandleRequest(t *testing.T) {
	// 创建mock服务
	service := newMockService(Config{})

	// 测试用例
	testCases := []struct {
		name            string
		method          string
		path            string
		body            []byte
		expectedStatus  int
		validateSuccess func(*testing.T, interface{})
	}{
		{
			name:           "聊天补全API",
			method:         "POST",
			path:           "/v1/chat/completions",
			body:           []byte(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"测试消息"}]}`),
			expectedStatus: http.StatusOK,
			validateSuccess: func(t *testing.T, respBody interface{}) {
				resp, ok := respBody.(Response)
				assert.True(t, ok, "响应应为Response类型")
				if ok {
					assert.NotEmpty(t, resp.ID, "ID不应为空")
					assert.Equal(t, "chat.completion", resp.Object, "对象类型应为chat.completion")
				}
			},
		},
		{
			name:           "文本补全API",
			method:         "POST",
			path:           "/v1/completions",
			body:           []byte(`{"model":"gpt-3.5-turbo","prompt":"测试提示"}`),
			expectedStatus: http.StatusOK,
			validateSuccess: func(t *testing.T, respBody interface{}) {
				resp, ok := respBody.(Response)
				assert.True(t, ok, "响应应为Response类型")
				if ok {
					assert.NotEmpty(t, resp.ID, "ID不应为空")
					assert.Equal(t, "text_completion", resp.Object, "对象类型应为text_completion")
				}
			},
		},
		{
			name:           "嵌入API",
			method:         "POST",
			path:           "/v1/embeddings",
			body:           []byte(`{"model":"text-embedding-3-small","input":"测试文本"}`),
			expectedStatus: http.StatusOK,
			validateSuccess: func(t *testing.T, respBody interface{}) {
				resp, ok := respBody.(map[string]interface{})
				assert.True(t, ok, "响应应为map类型")
				if ok {
					assert.Equal(t, "list", resp["object"], "对象类型应为list")
					assert.NotNil(t, resp["data"], "data字段不应为nil")
				}
			},
		},
		{
			name:           "图片生成API",
			method:         "POST",
			path:           "/v1/images/generations",
			body:           []byte(`{"model":"dall-e-3","prompt":"测试图片生成"}`),
			expectedStatus: http.StatusOK,
			validateSuccess: func(t *testing.T, respBody interface{}) {
				resp, ok := respBody.(map[string]interface{})
				assert.True(t, ok, "响应应为map类型")
				if ok {
					assert.NotNil(t, resp["created"], "created字段不应为nil")
					assert.NotNil(t, resp["data"], "data字段不应为nil")
				}
			},
		},
		{
			name:           "获取模型列表",
			method:         "GET",
			path:           "/v1/models",
			expectedStatus: http.StatusOK,
			validateSuccess: func(t *testing.T, respBody interface{}) {
				// 将响应转换为JSON并反序列化，确保正确处理类型
				jsonData, err := json.Marshal(respBody)
				assert.NoError(t, err, "应能将响应序列化为JSON")

				var parsedResp map[string]interface{}
				err = json.Unmarshal(jsonData, &parsedResp)
				assert.NoError(t, err, "应能将JSON解析回map")

				// 验证对象类型
				assert.Equal(t, "list", parsedResp["object"], "对象类型应为list")

				// 验证data字段是否为数组且非空
				dataArr, ok := parsedResp["data"].([]interface{})
				assert.True(t, ok, "data字段应为数组类型")
				assert.NotEmpty(t, dataArr, "模型列表不应为空")

				// 验证第一个模型的字段
				if len(dataArr) > 0 {
					model, ok := dataArr[0].(map[string]interface{})
					assert.True(t, ok, "模型数据应为map类型")
					if ok {
						assert.NotEmpty(t, model["id"], "模型id不应为空")
						assert.Equal(t, "model", model["object"], "模型object应为model")
					}
				}
			},
		},
		{
			name:           "不存在的API",
			method:         "GET",
			path:           "/v1/unknown",
			expectedStatus: http.StatusNotFound,
			validateSuccess: func(t *testing.T, respBody interface{}) {
				resp, ok := respBody.(Response)
				assert.True(t, ok, "响应应为Response类型")
				if ok {
					assert.NotNil(t, resp.Error, "应包含错误信息")
				}
			},
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, _, respBody, err := service.HandleRequest(tc.method, tc.path, nil, nil, tc.body)
			assert.NoError(t, err, "处理请求不应返回错误")
			assert.Equal(t, tc.expectedStatus, status, "HTTP状态码应匹配预期")
			if status == tc.expectedStatus && tc.validateSuccess != nil {
				tc.validateSuccess(t, respBody)
			}
		})
	}
}

// 专门测试模型列表API
func TestMockService_HandleListModels(t *testing.T) {
	// 创建服务
	service := &mockService{
		responseDelay:   0,
		handlerMapping:  make(map[string]mockHandler),
		supportedModels: getDefaultModels(),
	}
	service.registerHandlers()

	// 调用handleListModels
	statusCode, respBody := service.handleListModels("GET", nil, nil, nil)

	// 验证结果
	assert.Equal(t, http.StatusOK, statusCode, "状态码应为200")
	assert.NotNil(t, respBody, "响应体不应为nil")

	// 转换为map并验证结构
	respMap, ok := respBody.(map[string]interface{})
	assert.True(t, ok, "响应应为map")

	// 输出实际结构，帮助调试
	fmt.Printf("模型列表响应: %+v\n", respMap)

	// 验证基本字段
	assert.Equal(t, "list", respMap["object"], "对象类型应为list")
	assert.Contains(t, respMap, "data", "响应应包含data字段")

	// 验证data字段非空
	assert.NotNil(t, respMap["data"], "data字段不应为nil")

	// 直接将响应转换为JSON并检查
	respJson, err := json.Marshal(respBody)
	assert.NoError(t, err, "应能将响应序列化为JSON")

	var jsonMap map[string]interface{}
	err = json.Unmarshal(respJson, &jsonMap)
	assert.NoError(t, err, "应能反序列化JSON")

	// 通过JSON转换后检查data字段
	dataArr, ok := jsonMap["data"].([]interface{})
	assert.True(t, ok, "JSON中data字段应为数组")
	assert.NotEmpty(t, dataArr, "模型列表不应为空")

	// 检查第一个模型的结构
	if len(dataArr) > 0 {
		firstModel, ok := dataArr[0].(map[string]interface{})
		assert.True(t, ok, "模型应为map类型")
		assert.Contains(t, firstModel, "id", "模型应包含id字段")
		assert.Contains(t, firstModel, "object", "模型应包含object字段")
		assert.Contains(t, firstModel, "created", "模型应包含created字段")
		assert.Contains(t, firstModel, "owned_by", "模型应包含owned_by字段")
	}
}

// 测试ServeOpenAISpec方法
func TestMockService_ServeOpenAISpec(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个响应记录器
	w := createTestResponseRecorder()

	// 创建一个上下文
	c, _ := gin.CreateTestContext(w)

	// 创建服务
	config := Config{
		Enabled:         true,
		ResponseDelayMs: 0,
	}
	service := newMockService(config)

	// 调用方法
	service.ServeOpenAISpec(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code, "状态码应为200")

	// 解析响应体
	var respMap map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &respMap)
	assert.NoError(t, err, "应该能够解析JSON响应")

	// 验证关键字段
	assert.Contains(t, respMap, "title", "响应应包含title字段")
	assert.Contains(t, respMap, "description", "响应应包含description字段")
	assert.Contains(t, respMap, "models", "响应应包含models字段")
}

// 测试处理聊天补全请求
func TestMockService_HandleChatCompletions(t *testing.T) {
	// 创建mock服务
	service := &mockService{
		responseDelay:   0,
		handlerMapping:  make(map[string]mockHandler),
		supportedModels: getDefaultModels(),
	}

	// 注册处理函数
	service.registerHandlers()

	// 有效请求
	validBody := []byte(`{
		"model": "gpt-3.5-turbo",
		"messages": [
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Hello!"}
		]
	}`)

	// 无效请求（缺少model）
	invalidBody1 := []byte(`{
		"messages": [
			{"role": "user", "content": "Hello!"}
		]
	}`)

	// 无效请求（缺少messages）
	invalidBody2 := []byte(`{
		"model": "gpt-3.5-turbo"
	}`)

	tests := []struct {
		name         string
		body         []byte
		expectStatus int
	}{
		{
			name:         "有效请求",
			body:         validBody,
			expectStatus: http.StatusOK,
		},
		{
			name:         "缺少model",
			body:         invalidBody1,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "缺少messages",
			body:         invalidBody2,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "无效JSON",
			body:         []byte(`{"model": "gpt-3.5-turbo", "messages": [}`),
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用处理函数
			status, respBody := service.handleChatCompletions("POST", nil, nil, tt.body)

			// 验证结果
			assert.Equal(t, tt.expectStatus, status, "状态码应匹配")

			// 如果是成功响应，验证响应字段
			if tt.expectStatus == http.StatusOK {
				resp, ok := respBody.(Response)
				assert.True(t, ok, "响应应为Response类型")
				if ok {
					assert.NotEmpty(t, resp.ID, "ID不应为空")
					assert.Equal(t, "chat.completion", resp.Object, "Object应为chat.completion")
					assert.NotEmpty(t, resp.Choices, "Choices不应为空")
					assert.Equal(t, "assistant", resp.Choices[0].Message.Role, "角色应为assistant")
					assert.NotEmpty(t, resp.Choices[0].Message.Content, "内容不应为空")
				}
			} else {
				// 如果是错误响应，验证错误字段
				resp, ok := respBody.(Response)
				assert.True(t, ok, "响应应为Response类型")
				if ok {
					assert.NotNil(t, resp.Error, "Error字段不应为nil")
					assert.NotEmpty(t, resp.Error.Message, "错误消息不应为空")
					assert.Equal(t, "invalid_request_error", resp.Error.Type, "错误类型应为invalid_request_error")
				}
			}
		})
	}
}

// TestMockService_MockData 测试模拟服务生成的数据格式是否符合OpenAI规范
func TestMockService_MockData(t *testing.T) {
	// 创建mock服务
	service := &mockService{
		responseDelay:   0,
		handlerMapping:  make(map[string]mockHandler),
		supportedModels: getDefaultModels(),
	}
	service.registerHandlers()

	// 定义Model类型用于测试
	type Model struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	t.Run("聊天补全数据格式", func(t *testing.T) {
		body := []byte(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"测试消息"}]}`)
		status, respData := service.handleChatCompletions("POST", nil, nil, body)

		assert.Equal(t, http.StatusOK, status)
		resp, ok := respData.(Response)
		assert.True(t, ok, "响应应为Response类型")

		// 验证字段格式
		assert.NotEmpty(t, resp.ID, "ID不应为空")
		assert.Regexp(t, `^chatcmpl-[a-zA-Z0-9]+$`, resp.ID, "ID格式应匹配chatcmpl-*")
		assert.Equal(t, "chat.completion", resp.Object)
		assert.NotZero(t, resp.Created, "created时间戳不应为0")
		assert.NotEmpty(t, resp.Model, "model不应为空")
		assert.NotEmpty(t, resp.Choices, "choices不应为空")
		assert.Equal(t, "stop", resp.Choices[0].FinishReason, "finish_reason应为stop")
		assert.Equal(t, "assistant", resp.Choices[0].Message.Role, "role应为assistant")
		assert.NotEmpty(t, resp.Choices[0].Message.Content, "content不应为空")
		assert.NotNil(t, resp.Usage, "usage不应为nil")
		assert.GreaterOrEqual(t, resp.Usage.TotalTokens, resp.Usage.PromptTokens+resp.Usage.CompletionTokens,
			"total_tokens应大于等于prompt_tokens+completion_tokens")
	})

	t.Run("文本嵌入数据格式", func(t *testing.T) {
		body := []byte(`{"model":"text-embedding-3-small","input":"测试文本"}`)
		status, respData := service.handleEmbeddings("POST", nil, nil, body)

		assert.Equal(t, http.StatusOK, status)
		respMap, ok := respData.(map[string]interface{})
		assert.True(t, ok, "响应应为map类型")

		// 验证字段格式
		assert.Equal(t, "list", respMap["object"])
		assert.NotNil(t, respMap["data"], "data不应为空")

		// 将数据转换为JSON再解析回来，确保类型转换正确
		jsonData, err := json.Marshal(respData)
		assert.NoError(t, err, "应能将响应序列化为JSON")

		var parsedData map[string]interface{}
		err = json.Unmarshal(jsonData, &parsedData)
		assert.NoError(t, err, "应能将JSON解析回map")

		// 验证嵌入向量
		dataArray, ok := parsedData["data"].([]interface{})
		assert.True(t, ok, "data应为数组类型")
		if !ok {
			return
		}

		assert.NotEmpty(t, dataArray, "嵌入向量数组不应为空")
		if len(dataArray) > 0 {
			embedData, ok := dataArray[0].(map[string]interface{})
			assert.True(t, ok, "嵌入数据应为map类型")
			if !ok {
				return
			}

			assert.Equal(t, "embedding", embedData["object"])
			embedding, ok := embedData["embedding"].([]interface{})
			assert.True(t, ok, "embedding应为数组")
			assert.NotEmpty(t, embedding, "embedding数组不应为空")
		}

		// 验证usage
		usage, ok := parsedData["usage"].(map[string]interface{})
		assert.True(t, ok, "usage应为map")
		if ok {
			assert.NotNil(t, usage["prompt_tokens"], "prompt_tokens不应为nil")
			assert.NotNil(t, usage["total_tokens"], "total_tokens不应为nil")
		}
	})

	t.Run("图片生成数据格式", func(t *testing.T) {
		body := []byte(`{"model":"dall-e-3","prompt":"测试图片生成"}`)
		status, respData := service.handleImagesGenerations("POST", nil, nil, body)

		assert.Equal(t, http.StatusOK, status)
		respMap, ok := respData.(map[string]interface{})
		assert.True(t, ok, "响应应为map类型")

		// 验证字段格式
		assert.NotNil(t, respMap["created"], "created不应为nil")
		assert.NotNil(t, respMap["data"], "data不应为空")

		// 将数据转换为JSON再解析回来，确保类型转换正确
		jsonData, err := json.Marshal(respData)
		assert.NoError(t, err, "应能将响应序列化为JSON")

		var parsedData map[string]interface{}
		err = json.Unmarshal(jsonData, &parsedData)
		assert.NoError(t, err, "应能将JSON解析回map")

		// 验证图片数据
		dataArray, ok := parsedData["data"].([]interface{})
		assert.True(t, ok, "data应为数组类型")
		if !ok {
			return
		}

		assert.NotEmpty(t, dataArray, "data应包含元素")
		if len(dataArray) > 0 {
			// 检查第一个元素的结构
			firstItem, ok := dataArray[0].(map[string]interface{})
			assert.True(t, ok, "图片数据应为map类型")
			if !ok {
				return
			}

			assert.NotEmpty(t, firstItem["url"], "url不应为空")
			url, urlOk := firstItem["url"].(string)
			assert.True(t, urlOk, "url应为字符串")
			assert.Contains(t, url, "https://", "url应包含https://")
		}
	})

	t.Run("模型列表数据格式", func(t *testing.T) {
		status, respData := service.handleListModels("GET", nil, nil, nil)

		assert.Equal(t, http.StatusOK, status)
		respMap, ok := respData.(map[string]interface{})
		assert.True(t, ok, "响应应为map类型")

		// 验证字段格式
		assert.Equal(t, "list", respMap["object"])
		assert.NotNil(t, respMap["data"], "data不应为空")

		// 将数据转换为JSON再解析回来，确保类型转换正确
		jsonData, err := json.Marshal(respData)
		assert.NoError(t, err, "应能将响应序列化为JSON")

		var parsedData map[string]interface{}
		err = json.Unmarshal(jsonData, &parsedData)
		assert.NoError(t, err, "应能将JSON解析回map")

		// 验证模型数据
		dataArray, ok := parsedData["data"].([]interface{})
		assert.True(t, ok, "data应为数组类型")
		if !ok {
			return
		}

		assert.NotEmpty(t, dataArray, "data应包含元素")
		if len(dataArray) == 0 {
			return
		}

		// 检查是否包含预期的模型
		foundModels := make(map[string]bool)
		for _, item := range dataArray {
			modelMap, ok := item.(map[string]interface{})
			assert.True(t, ok, "模型数据应为map类型")
			if !ok {
				continue
			}

			if modelMap["id"] != nil {
				modelID, idOk := modelMap["id"].(string)
				if idOk {
					foundModels[modelID] = true

					// 验证必要字段
					assert.Equal(t, "model", modelMap["object"], "object应为model")
					assert.NotNil(t, modelMap["created"], "created不应为nil")
					assert.NotNil(t, modelMap["owned_by"], "owned_by不应为nil")
				}
			}
		}

		// 至少应该找到这两个模型
		assert.True(t, foundModels["gpt-3.5-turbo"], "应包含gpt-3.5-turbo模型")
		assert.True(t, foundModels["dall-e-3"], "应包含dall-e-3模型")
	})
}

// createTestResponseRecorder 创建测试响应记录器
type testResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *testResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func createTestResponseRecorder() *testResponseRecorder {
	return &testResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closeChannel:     make(chan bool, 1),
	}
}
