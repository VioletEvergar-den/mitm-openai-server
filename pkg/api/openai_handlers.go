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
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 注册所有需要的OpenAI API路径
	openaiGroup.POST("/chat/completions", h.HandleOpenAIRequest)
	openaiGroup.POST("/completions", h.HandleOpenAIRequest)
	openaiGroup.POST("/embeddings", h.HandleOpenAIRequest)
	openaiGroup.GET("/models", h.HandleOpenAIRequest)

	// 以下是更多可能需要的OpenAI API路径
	openaiGroup.GET("/models/:model", h.HandleOpenAIRequest)
	openaiGroup.POST("/images/generations", h.HandleOpenAIRequest)
	openaiGroup.POST("/audio/transcriptions", h.HandleOpenAIRequest)
	openaiGroup.POST("/audio/translations", h.HandleOpenAIRequest)
	openaiGroup.POST("/fine-tuning/jobs", h.HandleOpenAIRequest)
	openaiGroup.GET("/fine-tuning/jobs", h.HandleOpenAIRequest)
	openaiGroup.GET("/fine-tuning/jobs/:job_id", h.HandleOpenAIRequest)
}

// HandleOpenAIRequest 处理OpenAI API请求
// 根据配置，将请求代理到真实API或返回模拟响应
// 参数:
//   - c: Gin上下文
func (h *OpenAIHandler) HandleOpenAIRequest(c *gin.Context) {
	// 获取请求方法和路径
	method := c.Request.Method
	path := c.Param("path")

	// 读取请求体
	var bodyBytes []byte
	var err error
	if c.Request.Body != nil {
		bodyBytes, err = io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]interface{}{
					"message": "无法读取请求体: " + err.Error(),
					"type":    "server_error",
					"code":    "internal_server_error",
				},
			})
			return
		}
		// 恢复请求体，以便其他中间件可以再次读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 收集请求头
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 收集查询参数
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// 使用OpenAI服务处理请求
	statusCode, responseHeaders, responseBody, err := h.openaiService.HandleRequest(
		method, path, headers, queryParams, bodyBytes,
	)

	if err != nil {
		// 发生错误，返回错误响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"message": "处理请求失败: " + err.Error(),
				"type":    "server_error",
				"code":    "internal_server_error",
			},
		})
		return
	}

	// 设置响应头
	for key, value := range responseHeaders {
		c.Header(key, value)
	}

	// 返回响应
	c.JSON(statusCode, responseBody)
}
