package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
)

// OpenAIHandler 处理OpenAI API请求的处理器
type OpenAIHandler struct {
	storage       storage.Storage
	openaiService OpenAIServiceInterface
}

// NewOpenAIHandler 创建一个新的OpenAI处理器
func NewOpenAIHandler(storage storage.Storage, openaiService OpenAIServiceInterface) *OpenAIHandler {
	return &OpenAIHandler{
		storage:       storage,
		openaiService: openaiService,
	}
}

// SetupOpenAIRoutes 设置OpenAI API代理路由
// 配置所有与OpenAI API相关的路由
func (h *OpenAIHandler) SetupOpenAIRoutes(router *gin.Engine, apiMiddleware gin.HandlerFunc) {
	// 如果没有初始化openaiService，跳过设置OpenAI路由
	if h.openaiService == nil {
		fmt.Println("警告: openaiService未初始化，跳过OpenAI路由设置")
		return
	}

	// 添加OpenAI API规范路由 - 这个路由只添加一次
	router.GET("/openai.json", h.openaiService.ServeOpenAISpec)

	// OpenAI API路由组
	openaiGroup := router.Group("/v1")

	// 应用API中间件记录请求
	openaiGroup.Use(apiMiddleware)

	// 使用更宽松的CORS设置，确保跨域请求能正常工作
	openaiGroup.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// 如果没有Origin头，允许所有源
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// 有Origin头，设置为请求的源
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24小时

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 注册所有需要的OpenAI API路径
	openaiGroup.POST("/chat/completions", h.HandleRequest)
	openaiGroup.POST("/completions", h.HandleRequest)
	openaiGroup.POST("/embeddings", h.HandleRequest)
	openaiGroup.GET("/models", h.HandleRequest)

	// 以下是更多可能需要的OpenAI API路径
	openaiGroup.GET("/models/:model", h.HandleRequest)
	openaiGroup.POST("/images/generations", h.HandleRequest)
	openaiGroup.POST("/audio/transcriptions", h.HandleRequest)
	openaiGroup.POST("/audio/translations", h.HandleRequest)
	openaiGroup.POST("/fine-tuning/jobs", h.HandleRequest)
	openaiGroup.GET("/fine-tuning/jobs", h.HandleRequest)
	openaiGroup.GET("/fine-tuning/jobs/:job_id", h.HandleRequest)
}

// HandleRequest 处理OpenAI API请求
// 根据配置，将请求代理到真实API或返回模拟响应
// 参数:
//   - c: Gin上下文
func (h *OpenAIHandler) HandleRequest(c *gin.Context) {
	// 获取请求方法和路径
	method := c.Request.Method
	path := c.Request.URL.Path
	if len(path) > 3 && path[:3] == "/v1" {
		path = path[3:] // 移除前缀"/v1"
	}

	// 读取请求体
	var body []byte
	if c.Request.Body != nil {
		body, _ = io.ReadAll(c.Request.Body)
		// 重置请求体，以便后续处理
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// 获取请求头和查询参数
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	queryParams := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	// 调用OpenAI服务处理请求
	statusCode, respHeaders, respBody, err := h.openaiService.HandleRequest(method, path, headers, queryParams, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("处理请求失败: %v", err),
				"type":    "internal_server_error",
			},
		})
		return
	}

	// 设置响应头
	for k, v := range respHeaders {
		c.Header(k, v)
	}

	// 返回响应
	c.JSON(statusCode, respBody)
}
