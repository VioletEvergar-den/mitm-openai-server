package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		request  func() *http.Request
		expected string
	}{
		{
			name: "从X-Forwarded-For获取",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "http://example.com", nil)
				req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
				return req
			},
			expected: "192.168.1.1",
		},
		{
			name: "从X-Real-IP获取",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "http://example.com", nil)
				req.Header.Set("X-Real-IP", "192.168.1.2")
				return req
			},
			expected: "192.168.1.2",
		},
		{
			name: "从RemoteAddr获取",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "http://example.com", nil)
				req.RemoteAddr = "192.168.1.3:12345"
				return req
			},
			expected: "192.168.1.3",
		},
		{
			name: "无效RemoteAddr格式",
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "http://example.com", nil)
				req.RemoteAddr = "invalid-format"
				return req
			},
			expected: "invalid-format",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.request()
			result := GetClientIP(req)
			if result != tc.expected {
				t.Errorf("GetClientIP() = %v, 期望 %v", result, tc.expected)
			}
		})
	}
}

func TestHeadersToMap(t *testing.T) {
	headers := http.Header{
		"Content-Type":    []string{"application/json"},
		"Accept":          []string{"text/html", "application/xml"},
		"X-Request-ID":    []string{"123456"},
		"Accept-Encoding": []string{"gzip", "deflate"},
	}

	t.Run("扁平化处理", func(t *testing.T) {
		result := HeadersToMap(headers, true)
		expected := map[string]string{
			"Content-Type":    "application/json",
			"Accept":          "text/html, application/xml",
			"X-Request-ID":    "123456",
			"Accept-Encoding": "gzip, deflate",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("HeadersToMap(flatten=true) = %v, 期望 %v", result, expected)
		}
	})

	t.Run("不扁平化处理", func(t *testing.T) {
		result := HeadersToMap(headers, false)
		expected := map[string]string{
			"Content-Type":    "application/json",
			"Accept":          "text/html",
			"X-Request-ID":    "123456",
			"Accept-Encoding": "gzip",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("HeadersToMap(flatten=false) = %v, 期望 %v", result, expected)
		}
	})
}

func TestHeadersToMapArray(t *testing.T) {
	headers := http.Header{
		"Content-Type": []string{"application/json"},
		"Accept":       []string{"text/html", "application/xml"},
		"X-Request-ID": []string{"123456"},
	}

	result := HeadersToMapArray(headers)
	expected := map[string][]string{
		"Content-Type": {"application/json"},
		"Accept":       {"text/html", "application/xml"},
		"X-Request-ID": {"123456"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("HeadersToMapArray() = %v, 期望 %v", result, expected)
	}
}

func TestQueryToMap(t *testing.T) {
	queries := map[string][]string{
		"page":  {"1"},
		"limit": {"10"},
		"sort":  {"name", "age"},
		"tags":  {"tag1", "tag2", "tag3"},
	}

	t.Run("扁平化处理", func(t *testing.T) {
		result := QueryToMap(queries, true)
		expected := map[string]string{
			"page":  "1",
			"limit": "10",
			"sort":  "name, age",
			"tags":  "tag1, tag2, tag3",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("QueryToMap(flatten=true) = %v, 期望 %v", result, expected)
		}
	})

	t.Run("不扁平化处理", func(t *testing.T) {
		result := QueryToMap(queries, false)
		expected := map[string]string{
			"page":  "1",
			"limit": "10",
			"sort":  "name",
			"tags":  "tag1",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("QueryToMap(flatten=false) = %v, 期望 %v", result, expected)
		}
	})
}

func TestQueryToMapArray(t *testing.T) {
	queries := map[string][]string{
		"page":  {"1"},
		"limit": {"10"},
		"sort":  {"name", "age"},
		"tags":  {"tag1", "tag2", "tag3"},
	}

	result := QueryToMapArray(queries)
	expected := map[string][]string{
		"page":  {"1"},
		"limit": {"10"},
		"sort":  {"name", "age"},
		"tags":  {"tag1", "tag2", "tag3"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("QueryToMapArray() = %v, 期望 %v", result, expected)
	}
}

func TestIsJSONContentType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"APPLICATION/JSON", true},
		{"text/plain", false},
		{"text/html", false},
		{"application/xml", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.contentType, func(t *testing.T) {
			result := IsJSONContentType(tc.contentType)
			if result != tc.expected {
				t.Errorf("IsJSONContentType(%q) = %v, 期望 %v", tc.contentType, result, tc.expected)
			}
		})
	}
}

func TestParseJSONBody(t *testing.T) {
	tests := []struct {
		name     string
		body     []byte
		expected interface{}
		hasError bool
	}{
		{
			name:     "空JSON",
			body:     []byte{},
			expected: nil,
			hasError: false,
		},
		{
			name:     "有效JSON对象",
			body:     []byte(`{"name":"测试","value":123}`),
			expected: map[string]interface{}{"name": "测试", "value": float64(123)},
			hasError: false,
		},
		{
			name:     "有效JSON数组",
			body:     []byte(`[1,2,3]`),
			expected: []interface{}{float64(1), float64(2), float64(3)},
			hasError: false,
		},
		{
			name:     "无效JSON",
			body:     []byte(`{"name":"测试",}`),
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseJSONBody(tc.body)

			if tc.hasError && err == nil {
				t.Errorf("期望错误，但未获得错误")
				return
			}

			if !tc.hasError && err != nil {
				t.Errorf("不期望错误，但获得错误: %v", err)
				return
			}

			if !tc.hasError {
				// 转换为JSON字符串比较，避免类型差异导致的比较失败
				expectedJSON, _ := json.Marshal(tc.expected)
				resultJSON, _ := json.Marshal(result)

				if string(expectedJSON) != string(resultJSON) {
					t.Errorf("ParseJSONBody() = %v, 期望 %v", result, tc.expected)
				}
			}
		})
	}
}

func TestSendHTTPRequest(t *testing.T) {
	// 设置测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求头
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("请求头不匹配，期望 'test-value', 得到 %s", r.Header.Get("X-Test-Header"))
		}

		// 检查请求方法
		if r.Method != "POST" {
			t.Errorf("请求方法不匹配，期望 POST, 得到 %s", r.Method)
		}

		// 读取请求体
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("解析请求体失败: %v", err)
		}

		// 检查请求体内容
		if body["test"] != "value" {
			t.Errorf("请求体不匹配，期望 {'test':'value'}, 得到 %v", body)
		}

		// 返回响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// 执行测试
	headers := map[string]string{
		"X-Test-Header": "test-value",
		"Content-Type":  "application/json",
	}
	body := []byte(`{"test":"value"}`)

	resp, respBody, err := SendHTTPRequest("POST", server.URL, headers, body, 5)

	// 验证结果
	if err != nil {
		t.Errorf("SendHTTPRequest失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("状态码不匹配，期望 200, 得到 %d", resp.StatusCode)
	}

	if !strings.Contains(string(respBody), "success") {
		t.Errorf("响应体不匹配，期望包含 'success', 得到 %s", string(respBody))
	}
}

func TestSendProxyRequest(t *testing.T) {
	// 设置测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回响应
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer server.Close()

	// 执行测试
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	body := []byte(`{"test":"data"}`)

	proxyResp, err := SendProxyRequest(
		"GET",
		server.URL,
		"/api/test",
		headers,
		body,
		"token",
		"",
		"",
		"test-token",
	)

	// 验证结果
	if err != nil {
		t.Errorf("SendProxyRequest失败: %v", err)
		return
	}

	// 检查状态码
	if proxyResp["status_code"].(int) != http.StatusOK {
		t.Errorf("状态码不匹配，期望 200, 得到 %d", proxyResp["status_code"].(int))
	}

	// 检查头部
	respHeaders := proxyResp["headers"].(map[string]string)
	if respHeaders["X-Custom-Header"] != "custom-value" {
		t.Errorf("响应头不匹配，期望 'custom-value', 得到 %s", respHeaders["X-Custom-Header"])
	}

	// 检查响应体
	respBody := proxyResp["body"].(map[string]interface{})
	if respBody["result"] != "success" {
		t.Errorf("响应体不匹配，期望 {'result':'success'}, 得到 %v", respBody)
	}

	// 测试路径拼接的各种情况
	t.Run("URL无斜杠，路径有斜杠", func(t *testing.T) {
		_, err := SendProxyRequest("GET", server.URL, "/test", headers, body, "none", "", "", "")
		if err != nil {
			t.Errorf("路径拼接失败: %v", err)
		}
	})

	t.Run("URL有斜杠，路径无斜杠", func(t *testing.T) {
		_, err := SendProxyRequest("GET", server.URL+"/", "test", headers, body, "none", "", "", "")
		if err != nil {
			t.Errorf("路径拼接失败: %v", err)
		}
	})

	t.Run("URL有斜杠，路径有斜杠", func(t *testing.T) {
		_, err := SendProxyRequest("GET", server.URL+"/", "/test", headers, body, "none", "", "", "")
		if err != nil {
			t.Errorf("路径拼接失败: %v", err)
		}
	})

	// 测试基本认证
	t.Run("基本认证", func(t *testing.T) {
		authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				t.Error("未找到授权头")
			}

			// 返回响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer authServer.Close()

		_, err := SendProxyRequest("GET", authServer.URL, "/test", headers, body, "basic", "user", "pass", "")
		if err != nil {
			t.Errorf("基本认证失败: %v", err)
		}
	})
}
