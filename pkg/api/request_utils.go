package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/storage"
)

// SaveRequest 保存请求到存储
// 将API请求模型转换为存储请求模型并保存
// 参数:
//   - req: 要保存的请求
//   - storage: 存储接口
//
// 返回:
//   - error: 如果保存失败，返回错误
func SaveRequest(req *Request, store storage.Storage) error {
	// 转换API请求模型为存储请求模型
	storageReq := &storage.Request{
		ID:        req.ID,
		Method:    req.Method,
		Path:      req.Path,
		Timestamp: req.Timestamp,
		ClientIP:  req.IPAddress,
	}

	// 转换Headers（从map[string]string到map[string][]string）
	headers := make(map[string][]string)
	for k, v := range req.Headers {
		headers[k] = []string{v}
	}
	storageReq.Headers = headers

	// 转换Query（从map[string]string到map[string][]string）
	query := make(map[string][]string)
	for k, v := range req.Query {
		query[k] = []string{v}
	}
	storageReq.Query = query

	// 设置请求体
	storageReq.Body = req.Body

	// 如果有响应，也转换响应
	if req.Response != nil {
		// 转换响应头
		respHeaders := make(map[string][]string)
		for k, v := range req.Response.Headers {
			respHeaders[k] = []string{v}
		}

		storageReq.Response = &storage.ProxyResponse{
			StatusCode: req.Response.StatusCode,
			Headers:    respHeaders,
			Body:       req.Response.Body,
		}
	}

	// 保存到存储
	return store.SaveRequest(storageReq)
}

// ConvertStorageToAPIRequest 将存储请求模型转换为API请求模型
// 参数:
//   - req: 存储请求模型
//
// 返回:
//   - *Request: API请求模型
func ConvertStorageToAPIRequest(req *storage.Request) *Request {
	// 创建API请求模型
	apiReq := &Request{
		ID:     req.ID,
		Method: req.Method,
		Path:   req.Path,
		Body:   req.Body,
	}

	// 处理时间戳
	switch ts := req.Timestamp.(type) {
	case string:
		apiReq.Timestamp = ts
	case time.Time:
		apiReq.Timestamp = ts.Format(time.RFC3339)
	default:
		// 如果无法识别时间戳类型，使用当前时间
		apiReq.Timestamp = time.Now().Format(time.RFC3339)
	}

	// 处理IP地址（兼容两个字段）
	if req.ClientIP != "" {
		apiReq.IPAddress = req.ClientIP
	} else {
		apiReq.IPAddress = req.IPAddress
	}

	// 转换Headers
	apiReq.Headers = make(map[string]string)
	switch headers := req.Headers.(type) {
	case map[string][]string:
		for k, v := range headers {
			if len(v) > 0 {
				apiReq.Headers[k] = v[0]
			}
		}
	case map[string]string:
		apiReq.Headers = headers
	}

	// 转换Query
	apiReq.Query = make(map[string]string)
	switch query := req.Query.(type) {
	case map[string][]string:
		for k, v := range query {
			if len(v) > 0 {
				apiReq.Query[k] = v[0]
			}
		}
	case map[string]string:
		apiReq.Query = query
	}

	// 转换Response（如果存在）
	if req.Response != nil {
		respHeaders := make(map[string]string)

		// 转换响应头
		switch headers := req.Response.Headers.(type) {
		case map[string][]string:
			for k, v := range headers {
				if len(v) > 0 {
					respHeaders[k] = v[0]
				}
			}
		case map[string]string:
			respHeaders = headers
		}

		apiReq.Response = &ProxyResponse{
			StatusCode: req.Response.StatusCode,
			Headers:    respHeaders,
			Body:       req.Response.Body,
		}
	}

	return apiReq
}

// GetPaginationParams 从请求中获取分页参数
// 参数:
//   - c: Gin上下文
//
// 返回:
//   - page: 页码
//   - size: 每页数量
func GetPaginationParams(c *gin.Context) (page, size int) {
	page = 1  // 默认页码为1
	size = 20 // 默认每页20条

	// 解析查询参数
	pageParam := c.Query("page")
	sizeParam := c.Query("size")

	// 转换页码参数
	if pageParam != "" {
		if parsedPage, err := strconv.Atoi(pageParam); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	// 转换每页大小参数
	if sizeParam != "" {
		if parsedSize, err := strconv.Atoi(sizeParam); err == nil && parsedSize > 0 {
			size = parsedSize
		}
	}

	return page, size
}
