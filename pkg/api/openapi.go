package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// OpenAPISpec 表示OpenAPI规范
type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    map[string]interface{} `json:"info"`
	Paths   map[string]interface{} `json:"paths"`
}

// DefaultOpenAPISpec 返回默认的OpenAPI规范
func DefaultOpenAPISpec() *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: map[string]interface{}{
			"title":       "Fake OpenAPI Server",
			"description": "一个用于记录请求的假OpenAPI服务器",
			"version":     "1.0.0",
		},
		Paths: map[string]interface{}{
			"/v1/echo": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Echo API",
					"description": "返回接收到的请求数据",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "成功响应",
						},
					},
				},
			},
			"/v1/users": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "获取用户列表",
					"description": "返回用户列表",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "成功获取用户列表",
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "创建用户",
					"description": "创建新用户",
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "成功创建用户",
						},
					},
				},
			},
			"/v1/users/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "获取用户",
					"description": "根据ID获取用户信息",
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"description": "用户ID",
							"required":    true,
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "成功获取用户",
						},
						"404": map[string]interface{}{
							"description": "用户不存在",
						},
					},
				},
			},
		},
	}
}

// LoadOpenAPISpec 从文件加载OpenAPI规范
func LoadOpenAPISpec(filePath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// SaveOpenAPISpec 保存OpenAPI规范到文件
func SaveOpenAPISpec(spec *OpenAPISpec, filePath string) error {
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// ServeOpenAPISpec 向客户端提供OpenAPI规范
func (s *Server) ServeOpenAPISpec(c *gin.Context) {
	c.JSON(http.StatusOK, DefaultOpenAPISpec())
}
