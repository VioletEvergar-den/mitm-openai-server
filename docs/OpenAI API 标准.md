# OpenAI API 标准

本文档描述了fake-openapi-server项目中OpenAI API的标准实现。

## 1. 架构概述

OpenAI API模块被抽象为以下几个关键组件：

1. **接口定义** - 定义了Service接口，所有OpenAI服务的实现都要遵循这个接口
2. **模拟实现** - 提供了一个本地的模拟实现，返回固定或随机数据，不需要连接真实的OpenAI服务
3. **代理实现** - 提供了一个代理实现，转发请求到真实的OpenAI服务，并将响应返回给客户端
4. **配置系统** - 提供了统一的配置结构，用于配置服务的行为

## 2. 服务接口

所有OpenAI API服务实现都遵循`Service`接口：

```go
type Service interface {
    // Name 返回服务名称
    Name() string

    // ServeOpenAISpec 提供OpenAI API规范
    ServeOpenAISpec(c *gin.Context)

    // HandleRequest 处理API请求并返回模拟响应
    // 参数:
    //   - method: HTTP方法(GET, POST等)
    //   - path: 请求路径
    //   - headers: 请求头
    //   - queryParams: 查询参数
    //   - body: 请求体(字节数组)
    // 返回值: 状态码，响应头，响应体，错误信息
    HandleRequest(method, path string, headers, queryParams map[string]string, body []byte) (int, map[string]string, interface{}, error)
}
```

## 3. 配置选项

OpenAI服务使用统一的`Config`结构体进行配置：

```go
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
}
```

## 4. 标准响应格式

OpenAI API的标准响应格式如下：

```go
// Response 定义标准API响应
type Response struct {
    // 通用字段
    ID      string   `json:"id,omitempty"`
    Object  string   `json:"object,omitempty"`
    Created int64    `json:"created,omitempty"`
    Model   string   `json:"model,omitempty"`
    Choices []Choice `json:"choices,omitempty"`
    Usage   *Usage   `json:"usage,omitempty"`

    // 错误字段
    Error *ErrorResp `json:"error,omitempty"`

    // 其他数据，用于自定义响应
    Data interface{} `json:"data,omitempty"`
}
```

## 5. 支持的API路径

当前实现支持以下OpenAI API路径：

| 方法 | 路径 | 描述 |
| --- | --- | --- |
| POST | /v1/chat/completions | 聊天补全API |
| POST | /v1/completions | 文本补全API（传统） |
| POST | /v1/embeddings | 嵌入API |
| POST | /v1/images/generations | 图片生成API |
| POST | /v1/audio/transcriptions | 音频转写API |
| POST | /v1/moderations | 内容审核API |
| GET | /v1/models | 获取可用模型列表 |

## 6. 使用方法

### 6.1 创建服务实例

```go
// 创建默认配置
config := openai.DefaultConfig()

// 自定义配置
config.ResponseDelayMs = 500 // 模拟500ms延迟
config.ProxyMode = true // 使用代理模式
config.TargetURL = "https://api.openai.com" // 设置目标URL
config.TargetAuthType = "token"
config.TargetToken = "YOUR_API_KEY"

// 创建服务
service := openai.NewService(config)
```

### 6.2 选择服务实现

```go
// 使用服务名选择服务实现
mockService := openai.GetServiceByName("mock", config)
proxyService := openai.GetServiceByName("proxy", config)
```

### 6.3 处理请求

```go
// 处理请求
statusCode, headers, respBody, err := service.HandleRequest(
    "POST",
    "/v1/chat/completions",
    map[string]string{"Content-Type": "application/json"},
    nil,
    []byte(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}`),
)

if err != nil {
    // 处理错误
    log.Printf("请求处理错误: %v", err)
    return
}

// 使用响应
fmt.Printf("状态码: %d\n", statusCode)
fmt.Printf("响应: %+v\n", respBody)
```

## 7. 请求和响应示例

### 7.1 聊天补全API

**请求:**
```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {"role": "system", "content": "你是一个有用的助手。"},
    {"role": "user", "content": "你好！"}
  ]
}
```

**响应:**
```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1679358384,
  "model": "gpt-3.5-turbo",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "你好！有什么我可以帮助你的吗？"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 12,
    "total_tokens": 37
  }
}
```

### 7.2 嵌入API

**请求:**
```json
{
  "model": "text-embedding-3-small",
  "input": "这是一段要嵌入的文本。"
}
```

**响应:**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [-0.006968,..,0.028131],
      "index": 0
    }
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

## 8. 错误处理

服务返回的错误格式如下：

```json
{
  "error": {
    "message": "错误消息",
    "type": "错误类型",
    "param": "错误参数",
    "code": "错误代码"
  }
}
```

主要错误类型包括：

- `invalid_request_error`: 请求无效
- `authentication_error`: 认证错误
- `permission_error`: 权限错误
- `rate_limit_error`: 速率限制错误
- `service_unavailable`: 服务不可用
- `api_error`: API错误

## 9. 扩展和定制

### 9.1 添加新的API支持

要添加新的API支持，请在mock.go文件的`registerHandlers`方法中注册新的处理函数：

```go
func (s *mockService) registerHandlers() {
    // 现有的处理函数...
    
    // 添加新的API处理函数
    s.handlerMapping["POST /v1/new/api/path"] = s.handleNewAPI
}

// 实现处理函数
func (s *mockService) handleNewAPI(method string, pathParams, queryParams map[string]string, body []byte) (int, interface{}) {
    // 处理逻辑...
}
```

### 9.2 自定义响应

可以通过配置自定义响应：

```go
config.CustomResponse = `{
  "choices": [
    {
      "message": {
        "content": "这是一个自定义的响应。"
      }
    }
  ]
}`
```

## 10. 测试

包含完整的单元测试，确保各种情况下的功能正常。运行测试：

```bash
cd pkg/openai
go test -v
``` 