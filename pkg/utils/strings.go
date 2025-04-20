package utils

import (
	"encoding/base64"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

// TruncateString 截断字符串到指定长度
//
// 参数:
//   - s: 原始字符串
//   - maxLen: 最大长度
//   - suffix: 截断后添加的后缀，如"..."
//
// 返回:
//   - string: 截断后的字符串
func TruncateString(s string, maxLen int, suffix string) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	// 确保有足够空间放后缀
	suffixRunes := []rune(suffix)
	truncLen := maxLen - len(suffixRunes)
	if truncLen < 0 {
		truncLen = 0
	}

	return string(runes[:truncLen]) + suffix
}

// GenerateRandomString 生成指定长度的随机字符串
//
// 参数:
//   - length: 字符串长度
//
// 返回:
//   - string: 随机字符串
//   - error: 生成过程中的错误，如果成功则为nil
func GenerateRandomString(length int) (string, error) {
	// 字符集
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 创建一个基于当前时间的随机源
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成随机字符串
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[r.Intn(len(charset))]
	}

	return string(result), nil
}

// GenerateRandomPassword 生成随机密码
//
// 参数:
//   - length: 密码长度
//
// 返回:
//   - string: 随机密码
func GenerateRandomPassword(length int) string {
	if length < 8 {
		length = 8 // 确保最小长度为8
	}

	// 简化字符集，只使用字母和数字，避免特殊字符可能带来的问题
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 随机种子
	seed := time.Now().UnixNano()
	rand.Seed(seed)

	// 生成密码
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		password[i] = charset[rand.Intn(len(charset))]
	}

	return string(password)
}

// Base64Encode 将字节数组编码为Base64字符串
//
// 参数:
//   - data: 要编码的数据
//
// 返回:
//   - string: Base64编码后的字符串
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode 解码Base64字符串
//
// 参数:
//   - s: Base64编码的字符串
//
// 返回:
//   - []byte: 解码后的数据
//   - error: 解码过程中的错误，如果成功则为nil
func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// SplitAndTrim 分割字符串并修剪空白
//
// 参数:
//   - s: 要分割的字符串
//   - sep: 分隔符
//
// 返回:
//   - []string: 分割并修剪后的字符串数组
func SplitAndTrim(s string, sep string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// IsAlphanumeric 检查字符串是否仅包含字母和数字
//
// 参数:
//   - s: 要检查的字符串
//
// 返回:
//   - bool: 如果仅包含字母和数字则为true，否则为false
func IsAlphanumeric(s string) bool {
	if s == "" {
		return true // 空字符串视为字母数字字符串
	}

	for _, r := range s {
		// 只允许ASCII字母和数字
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

// IsNumeric 检查字符串是否仅包含数字
//
// 参数:
//   - s: 要检查的字符串
//
// 返回:
//   - bool: 如果仅包含数字则为true，否则为false
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}
