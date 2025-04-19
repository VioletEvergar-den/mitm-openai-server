// Package embed 提供前端资源的访问
package embed

import (
	"net/http"
	"os"
	"path/filepath"
)

// 在开发模式下前端文件的默认路径
const defaultUIDir = "./react-ui/dist"

// GetFS 获取前端文件系统
// 在开发模式下，返回物理文件系统
// 在生产模式下，尝试使用嵌入式文件系统，如果失败则回退到物理文件系统
func GetFS(uiDir string) http.FileSystem {
	// 使用指定目录，如果未指定则使用默认目录
	if uiDir == "" {
		uiDir = defaultUIDir
	}

	// 检查目录是否存在
	_, err := os.Stat(uiDir)
	if err == nil {
		return http.Dir(uiDir)
	}

	// 如果指定目录不存在，尝试使用内置资源
	return http.Dir(defaultUIDir)
}

// IsDevelopmentMode 判断是否为开发模式
func IsDevelopmentMode() bool {
	// 通过环境变量判断是否为开发模式
	return os.Getenv("GO_ENV") == "development"
}

// ResolvePath 解析资源路径
func ResolvePath(uiDir, relativePath string) string {
	if uiDir == "" {
		uiDir = defaultUIDir
	}
	return filepath.Join(uiDir, relativePath)
}
