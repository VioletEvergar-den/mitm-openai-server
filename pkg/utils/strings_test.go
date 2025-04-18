package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		suffix   string
		expected string
	}{
		{
			name:     "不需要截断",
			s:        "短字符串",
			maxLen:   10,
			suffix:   "...",
			expected: "短字符串",
		},
		{
			name:     "需要截断",
			s:        "这是一个很长的字符串需要被截断",
			maxLen:   10,
			suffix:   "...",
			expected: "这是一个很长的...",
		},
		{
			name:     "截断长度等于字符串长度",
			s:        "测试字符串",
			maxLen:   5,
			suffix:   "...",
			expected: "测试字符串",
		},
		{
			name:     "截断长度小于后缀长度",
			s:        "测试字符串",
			maxLen:   2,
			suffix:   "......",
			expected: "......",
		},
		{
			name:     "空字符串",
			s:        "",
			maxLen:   5,
			suffix:   "...",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TruncateString(tc.s, tc.maxLen, tc.suffix)
			if result != tc.expected {
				t.Errorf("TruncateString(%q, %d, %q) = %q, 期望 %q", tc.s, tc.maxLen, tc.suffix, result, tc.expected)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	lengths := []int{0, 5, 10, 20}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("长度为%d", length), func(t *testing.T) {
			result, err := GenerateRandomString(length)
			if err != nil {
				t.Fatalf("GenerateRandomString(%d)错误: %v", length, err)
			}

			if len(result) != length {
				t.Errorf("GenerateRandomString(%d)生成的字符串长度为 %d, 期望 %d", length, len(result), length)
			}

			// 检查随机性，两次调用应该生成不同的字符串（除非长度为0）
			if length > 0 {
				result2, err := GenerateRandomString(length)
				if err != nil {
					t.Fatalf("第二次GenerateRandomString(%d)错误: %v", length, err)
				}

				if result == result2 {
					t.Errorf("两次调用GenerateRandomString(%d)生成了相同的字符串: %s", length, result)
				}
			}
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	lengths := []int{8, 12, 16}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("长度为%d", length), func(t *testing.T) {
			password, err := GenerateRandomPassword(length)
			if err != nil {
				t.Fatalf("GenerateRandomPassword(%d)错误: %v", length, err)
			}

			if len(password) != length {
				t.Errorf("GenerateRandomPassword(%d)生成的密码长度为 %d, 期望 %d", length, len(password), length)
			}

			// 检查密码中包含的字符类型，但不强制要求每种类型都出现
			// 只是检查字符是否来自预期的字符集
			for _, c := range password {
				if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+=-", c) {
					t.Errorf("密码中包含意外字符: %c", c)
				}
			}

			// 检查随机性，两次调用应该生成不同的密码
			password2, err := GenerateRandomPassword(length)
			if err != nil {
				t.Fatalf("第二次GenerateRandomPassword(%d)错误: %v", length, err)
			}

			if password == password2 {
				t.Errorf("两次调用GenerateRandomPassword(%d)生成了相同的密码: %s", length, password)
			}
		})
	}
}

func TestBase64EncodeDecode(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "空字节数组",
			input: []byte{},
		},
		{
			name:  "ASCII字符",
			input: []byte("Hello, World!"),
		},
		{
			name:  "UTF-8字符",
			input: []byte("你好，世界！"),
		},
		{
			name:  "二进制数据",
			input: []byte{0x00, 0xFF, 0x80, 0x3C, 0x7E},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 编码测试
			encoded := Base64Encode(tc.input)

			// 解码测试
			decoded, err := Base64Decode(encoded)
			if err != nil {
				t.Fatalf("Base64Decode(%q)失败: %v", encoded, err)
			}

			// 验证解码后的数据与原始数据一致
			if len(decoded) != len(tc.input) {
				t.Errorf("解码后的数据长度 %d, 期望 %d", len(decoded), len(tc.input))
			}

			for i := 0; i < len(tc.input); i++ {
				if decoded[i] != tc.input[i] {
					t.Errorf("解码后的数据与原始数据不一致，在位置 %d: 得到 %d, 期望 %d", i, decoded[i], tc.input[i])
					break
				}
			}
		})
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected []string
	}{
		{
			name:     "空字符串",
			s:        "",
			sep:      ",",
			expected: []string{},
		},
		{
			name:     "只有分隔符",
			s:        ",,,",
			sep:      ",",
			expected: []string{},
		},
		{
			name:     "带空白的字符串",
			s:        "  a  ,  b  ,  c  ",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "不同的分隔符",
			s:        "a|b|c",
			sep:      "|",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "包含空元素",
			s:        "a,,b,c",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "多字分隔符",
			s:        "a::b::c",
			sep:      "::",
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitAndTrim(tc.s, tc.sep)

			if len(result) != len(tc.expected) {
				t.Errorf("SplitAndTrim(%q, %q) 结果长度 = %d, 期望 %d", tc.s, tc.sep, len(result), len(tc.expected))
				return
			}

			for i, v := range result {
				if v != tc.expected[i] {
					t.Errorf("SplitAndTrim(%q, %q) 结果[%d] = %q, 期望 %q", tc.s, tc.sep, i, v, tc.expected[i])
				}
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		s        string
		expected bool
	}{
		{"abc123", true},
		{"ABC123", true},
		{"123", true},
		{"abc", true},
		{"abc123!", false},
		{"abc 123", false},
		{"中文abc123", false},
		{"", true},
	}

	for _, tc := range tests {
		t.Run(tc.s, func(t *testing.T) {
			result := IsAlphanumeric(tc.s)
			if result != tc.expected {
				t.Errorf("IsAlphanumeric(%q) = %v, 期望 %v", tc.s, result, tc.expected)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		s        string
		expected bool
	}{
		{"123", true},
		{"123.45", false},
		{"abc123", false},
		{"123abc", false},
		{"123 456", false},
		{"-123", false},
		{"+123", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.s, func(t *testing.T) {
			result := IsNumeric(tc.s)
			if result != tc.expected {
				t.Errorf("IsNumeric(%q) = %v, 期望 %v", tc.s, result, tc.expected)
			}
		})
	}
}
