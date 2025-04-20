package utils

import (
	"net/http"
)

// ConvertToHTTPHeader 将字符串映射转换为http.Header
// 参数:
//   - headers: 键值对形式的头部映射
//
// 返回:
//   - http.Header: 转换后的HTTP头
func ConvertToHTTPHeader(headers map[string]string) http.Header {
	result := make(http.Header)
	for key, value := range headers {
		result.Add(key, value)
	}
	return result
}

// ConvertToStringArray 将字符串映射转换为字符串数组映射
// 参数:
//   - data: 键值对形式的映射
//
// 返回:
//   - map[string][]string: 转换后的字符串数组映射
func ConvertToStringArray(data map[string]string) map[string][]string {
	result := make(map[string][]string)
	for key, value := range data {
		result[key] = []string{value}
	}
	return result
}

// ConvertToStringMap 将字符串数组映射转换为字符串映射
// 参数:
//   - data: 键值对数组形式的映射
//
// 返回:
//   - map[string]string: 转换后的字符串映射
func ConvertToStringMap(data map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range data {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// ConvertHeaderToStringMap 将HTTP头部转换为字符串映射
// 参数:
//   - header: HTTP头部
//
// 返回:
//   - map[string]string: 转换后的字符串映射
func ConvertHeaderToStringMap(header http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}
