package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// EnsureDirectory 确保目录存在，如果不存在则创建
//
// 参数:
//   - path: 要确保存在的目录路径
//
// 返回:
//   - error: 创建目录过程中的错误，如果成功则为nil
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// EnsureDir 确保目录存在的别名函数，与EnsureDirectory功能相同
//
// 参数:
//   - dir: 目录路径
//
// 返回:
//   - error: 如果创建过程中出错则返回错误，否则为nil
func EnsureDir(dir string) error {
	return EnsureDirectory(dir)
}

// BuildFilePath 构建文件完整路径
//
// 参数:
//   - baseDir: 基础目录
//   - id: 文件ID或名称
//   - ext: 文件扩展名（带或不带点号）
//
// 返回:
//   - string: 完整的文件路径
func BuildFilePath(baseDir, id string, ext string) string {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return filepath.Join(baseDir, fmt.Sprintf("%s%s", id, ext))
}

// SanitizeFilename 净化文件名，移除不安全的字符
//
// 参数:
//   - name: 原始文件名
//
// 返回:
//   - string: 处理后的安全文件名
func SanitizeFilename(name string) string {
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

// IsFileExists 检查文件是否存在
//
// 参数:
//   - path: 文件路径
//
// 返回:
//   - bool: 如果文件存在则为true，否则为false
//   - error: 检查过程中的错误，如果成功则为nil
func IsFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		// 只有当路径存在，且不是目录时返回true
		return !info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// FileExists 检查文件是否存在（简化版本）
//
// 参数:
//   - filePath: 文件路径
//
// 返回:
//   - bool: 如果文件存在则为true，否则为false
func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists 检查目录是否存在
//
// 参数:
//   - dirPath: 目录路径
//
// 返回:
//   - bool: 如果目录存在则为true，否则为false
func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize 获取文件大小
//
// 参数:
//   - path: 文件路径
//
// 返回:
//   - int64: 文件大小（字节数）
//   - error: 获取过程中的错误，如果成功则为nil
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("获取文件信息失败: %w", err)
	}
	return info.Size(), nil
}

// CreateTempFile 创建临时文件
//
// 参数:
//   - dir: 临时文件所在目录，为空则使用系统临时目录
//   - pattern: 临时文件名模式
//
// 返回:
//   - string: 临时文件路径
//   - error: 创建过程中的错误，如果成功则为nil
func CreateTempFile(dir, pattern string) (string, error) {
	if dir == "" {
		dir = os.TempDir()
	} else if err := EnsureDirectory(dir); err != nil {
		return "", fmt.Errorf("创建临时文件目录失败: %w", err)
	}

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}

	path := file.Name()
	if err := file.Close(); err != nil {
		return path, fmt.Errorf("关闭临时文件失败: %w", err)
	}

	return path, nil
}

// CopyFile 复制文件
//
// 参数:
//   - src: 源文件路径
//   - dst: 目标文件路径
//
// 返回:
//   - error: 如果复制过程中出错则返回错误，否则为nil
func CopyFile(src, dst string) error {
	// 确保目标目录存在
	err := EnsureDirectory(filepath.Dir(dst))
	if err != nil {
		return err
	}

	// 打开源文件
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// 创建目标文件
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 复制内容
	_, err = io.Copy(destination, source)
	return err
}

// ListFiles 列出目录中的所有文件（不包括子目录）
//
// 参数:
//   - dir: 目录路径
//   - pattern: 文件名匹配模式，如"*.json"
//
// 返回:
//   - []string: 匹配的文件路径列表
//   - error: 如果列出过程中出错则返回错误，否则为nil
func ListFiles(dir, pattern string) ([]string, error) {
	var files []string

	// 获取目录中的所有条目
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// 遍历并筛选文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		path := filepath.Join(dir, name)

		// 如果提供了模式，使用模式进行匹配
		if pattern != "" {
			matched, err := filepath.Match(pattern, name)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}

		files = append(files, path)
	}

	return files, nil
}

// GetFileExtension 获取文件扩展名（不带点）
//
// 参数:
//   - path: 文件路径
//
// 返回:
//   - string: 文件扩展名（不带点）
func GetFileExtension(path string) string {
	base := filepath.Base(path)
	// 对于格式如 .gitignore 的隐藏文件，返回空扩展名
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
		return ""
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return ext[1:] // 去掉前导点
}

// ReadFileToString 读取文件内容为字符串
//
// 参数:
//   - filePath: 文件路径
//
// 返回:
//   - string: 文件内容
//   - error: 如果读取过程中出错则返回错误，否则为nil
func ReadFileToString(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteStringToFile 将字符串写入文件
//
// 参数:
//   - filePath: 文件路径
//   - content: 要写入的内容
//
// 返回:
//   - error: 如果写入过程中出错则返回错误，否则为nil
func WriteStringToFile(filePath, content string) error {
	// 确保目录存在
	err := EnsureDirectory(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

// GetFileName 从路径中获取文件名（不包含扩展名）
//
// 参数:
//   - path: 文件路径
//
// 返回:
//   - string: 文件名（不含扩展名）
func GetFileName(path string) string {
	base := filepath.Base(path)
	// 特殊处理以点开头的隐藏文件
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
		return base
	}
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// JoinPaths 连接多个路径段
//
// 参数:
//   - paths: 要连接的路径段
//
// 返回:
//   - string: 连接后的路径
func JoinPaths(paths ...string) string {
	return filepath.Join(paths...)
}

// IsAbsolutePath 检查是否为绝对路径
//
// 参数:
//   - path: 要检查的路径
//
// 返回:
//   - bool: 如果是绝对路径则为true，否则为false
func IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}
