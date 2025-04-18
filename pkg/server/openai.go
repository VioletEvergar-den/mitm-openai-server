package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// DefaultOpenAISpec 返回默认的OpenAI规范
// 当没有自定义规范时使用此函数生成基本的规范
func (s *Server) DefaultOpenAISpec() OpenAISpec {
	// 创建基本的OpenAI规范
	spec := OpenAISpec{
		Version: "1.0.0",
		Info: map[string]interface{}{
			"title":       "MITM OpenAI API",
			"description": "中间人OpenAI API服务，用于记录和分析OpenAI API请求",
			"version":     "1.0.0",
		},
		Models: []string{
			"gpt-3.5-turbo",
			"gpt-4",
			"gpt-4-turbo",
			"text-embedding-ada-002",
		},
	}

	return spec
}

// LoadOpenAISpecFromFile 从文件加载OpenAI规范
// 参数:
// - filePath: 规范文件的路径
// 返回:
// - 加载的规范
// - 错误信息（如果有）
func (s *Server) LoadOpenAISpecFromFile(filePath string) (OpenAISpec, error) {
	var spec OpenAISpec

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return spec, fmt.Errorf("规范文件不存在: %s", filePath)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return spec, fmt.Errorf("无法读取规范文件: %v", err)
	}

	// 解析JSON
	err = json.Unmarshal(data, &spec)
	if err != nil {
		return spec, fmt.Errorf("无法解析规范JSON: %v", err)
	}

	return spec, nil
}

// SaveOpenAISpecToFile 将OpenAI规范保存到文件
// 参数:
// - spec: 要保存的规范
// - filePath: 保存文件的路径
// 返回:
// - 错误信息（如果有）
func (s *Server) SaveOpenAISpecToFile(spec OpenAISpec, filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建目录: %v", err)
	}

	// 将规范转换为JSON
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("无法将规范转换为JSON: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("无法写入规范文件: %v", err)
	}

	return nil
}

// ServeOpenAISpec 通过HTTP请求提供OpenAI规范
// 这是一个HTTP处理函数，用于在API文档请求时返回OpenAI规范
func (s *Server) ServeOpenAISpec(c *gin.Context) {
	// 如果有OpenAI服务，则委托给它处理规范请求
	if s.openaiService != nil {
		s.openaiService.ServeOpenAISpec(c)
		return
	}

	// 否则使用默认规范
	spec := s.DefaultOpenAISpec()
	c.JSON(http.StatusOK, spec)
}
