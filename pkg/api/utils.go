package api

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ensureDirectory 确保目录存在
func ensureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// formatTimestamp 格式化时间戳
func formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// buildFilePath 构建文件路径
func buildFilePath(baseDir, id string, ext string) string {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return filepath.Join(baseDir, fmt.Sprintf("%s%s", id, ext))
}

// getIPAddress 获取客户端IP地址
func getIPAddress(r *http.Request) string {
	// 尝试从X-Forwarded-For头获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP头获取
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// 直接从请求中获取
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// sanitizeFilename 净化文件名
func sanitizeFilename(name string) string {
	// 移除不安全的字符
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)

	// 确保文件名不为空
	if name == "" {
		name = "unnamed"
	}

	return name
}
