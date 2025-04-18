package utils

import (
	"testing"
	"time"
)

func TestFormatTimestamp(t *testing.T) {
	// 创建一个固定的时间用于测试
	testTime := time.Date(2023, 5, 15, 12, 30, 45, 0, time.UTC)
	expected := "2023-05-15T12:30:45Z" // RFC3339格式

	result := FormatTimestamp(testTime)
	if result != expected {
		t.Errorf("FormatTimestamp() = %s, 期望 %s", result, expected)
	}

	// 测试本地时间
	localTime := time.Date(2023, 5, 15, 12, 30, 45, 0, time.Local)
	localExpected := localTime.Format(time.RFC3339)

	localResult := FormatTimestamp(localTime)
	if localResult != localExpected {
		t.Errorf("FormatTimestamp(本地时间) = %s, 期望 %s", localResult, localExpected)
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		expected  time.Time
		hasError  bool
	}{
		{
			name:      "RFC3339格式",
			timestamp: "2023-05-15T12:30:45Z",
			expected:  time.Date(2023, 5, 15, 12, 30, 45, 0, time.UTC),
			hasError:  false,
		},
		{
			name:      "RFC3339Nano格式",
			timestamp: "2023-05-15T12:30:45.123456789Z",
			expected:  time.Date(2023, 5, 15, 12, 30, 45, 123456789, time.UTC),
			hasError:  false,
		},
		{
			name:      "ISO日期时间格式",
			timestamp: "2023-05-15T12:30:45",
			expected:  time.Date(2023, 5, 15, 12, 30, 45, 0, time.UTC),
			hasError:  false,
		},
		{
			name:      "常见日期时间格式",
			timestamp: "2023-05-15 12:30:45",
			expected:  time.Date(2023, 5, 15, 12, 30, 45, 0, time.UTC),
			hasError:  false,
		},
		{
			name:      "仅日期格式",
			timestamp: "2023-05-15",
			expected:  time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			hasError:  false,
		},
		{
			name:      "RFC1123格式",
			timestamp: "Mon, 15 May 2023 12:30:45 GMT",
			expected:  time.Date(2023, 5, 15, 12, 30, 45, 0, time.UTC),
			hasError:  false,
		},
		{
			name:      "紧凑日期格式",
			timestamp: "20230515",
			expected:  time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			hasError:  false,
		},
		{
			name:      "斜杠分隔日期",
			timestamp: "2023/05/15",
			expected:  time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			hasError:  false,
		},
		{
			name:      "Unix时间戳格式(秒)",
			timestamp: "1684154645",
			expected:  time.Unix(1684154645, 0),
			hasError:  false,
		},
		{
			name:      "Unix时间戳格式(毫秒)",
			timestamp: "1684154645123",
			expected:  time.Unix(1684154645, 123000000),
			hasError:  false,
		},
		{
			name:      "无效格式",
			timestamp: "invalid-timestamp",
			expected:  time.Time{},
			hasError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseTimestamp(tc.timestamp)

			// 检查错误状态
			if tc.hasError && err == nil {
				t.Errorf("ParseTimestamp(%s) 期望错误，但没有错误", tc.timestamp)
				return
			}
			if !tc.hasError && err != nil {
				t.Errorf("ParseTimestamp(%s) 期望成功，但有错误: %v", tc.timestamp, err)
				return
			}

			// 检查结果，除了Unix时间戳外
			if !tc.hasError {
				if tc.name == "Unix时间戳格式(秒)" || tc.name == "Unix时间戳格式(毫秒)" {
					// 对于Unix时间戳，只比较Unix时间
					expectedUnix := tc.expected.Unix()
					resultUnix := result.Unix()
					if expectedUnix != resultUnix {
						t.Errorf("ParseTimestamp(%s) = %v (Unix: %d), 期望 %v (Unix: %d)",
							tc.timestamp, result, resultUnix, tc.expected, expectedUnix)
					}

					// 对于毫秒时间戳，还需要检查纳秒部分
					if tc.name == "Unix时间戳格式(毫秒)" {
						expectedNano := tc.expected.Nanosecond()
						resultNano := result.Nanosecond()
						if expectedNano != resultNano {
							t.Errorf("ParseTimestamp(%s) 纳秒部分 = %d, 期望 %d",
								tc.timestamp, resultNano, expectedNano)
						}
					}
				} else {
					// 仅比较年月日时分秒，忽略纳秒和时区差异
					expected := tc.expected
					if result.Year() != expected.Year() ||
						result.Month() != expected.Month() ||
						result.Day() != expected.Day() ||
						result.Hour() != expected.Hour() ||
						result.Minute() != expected.Minute() ||
						result.Second() != expected.Second() {
						t.Errorf("ParseTimestamp(%s) = %v, 期望 %v", tc.timestamp, result, expected)
					}
				}
			}
		})
	}
}

func TestTimeToInterface(t *testing.T) {
	testTime := time.Date(2023, 5, 15, 12, 30, 45, 0, time.UTC)
	testTimePtr := &testTime
	nilTimePtr := (*time.Time)(nil)
	timeStr := "2023-05-15T12:30:45Z"

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "time.Time类型",
			input:    testTime,
			expected: FormatTimestamp(testTime),
		},
		{
			name:     "*time.Time非空指针",
			input:    testTimePtr,
			expected: FormatTimestamp(testTime),
		},
		{
			name:     "*time.Time空指针",
			input:    nilTimePtr,
			expected: nil,
		},
		{
			name:     "字符串类型",
			input:    timeStr,
			expected: timeStr,
		},
		{
			name:     "nil值",
			input:    nil,
			expected: nil,
		},
		{
			name:     "其他类型",
			input:    123,
			expected: "123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TimeToInterface(tc.input)

			// 对于nil，直接比较
			if tc.expected == nil {
				if result != nil {
					t.Errorf("TimeToInterface(%v) = %v, 期望 nil", tc.input, result)
				}
				return
			}

			// 对于非nil，比较字符串表示
			if result != tc.expected {
				t.Errorf("TimeToInterface(%v) = %v, 期望 %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestNow(t *testing.T) {
	// 获取当前时间作为参考
	before := time.Now()

	// 调用Now()函数
	result := Now()

	// 再次获取当前时间
	after := time.Now()

	// 确保Now()返回的时间在before和after之间
	if result.Before(before) || result.After(after) {
		t.Errorf("Now() = %v，不在预期的时间范围内 [%v, %v]", result, before, after)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "几秒钟",
			duration: 5 * time.Second,
			expected: "5.0秒",
		},
		{
			name:     "几分钟",
			duration: 3*time.Minute + 45*time.Second,
			expected: "3分45秒",
		},
		{
			name:     "几小时",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2小时30分",
		},
		{
			name:     "几天",
			duration: 3*24*time.Hour + 5*time.Hour,
			expected: "3天5小时",
		},
		{
			name:     "零时长",
			duration: 0,
			expected: "0.0秒",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatDuration(tc.duration)
			if result != tc.expected {
				t.Errorf("FormatDuration(%v) = %q, 期望 %q", tc.duration, result, tc.expected)
			}
		})
	}
}

func TestGetStartOfDay(t *testing.T) {
	// 测试任意时间点
	testTime := time.Date(2023, 5, 15, 12, 30, 45, 123456789, time.UTC)
	expected := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)

	result := GetStartOfDay(testTime)
	if !result.Equal(expected) {
		t.Errorf("GetStartOfDay(%v) = %v, 期望 %v", testTime, result, expected)
	}

	// 测试当天开始时间
	startOfDay := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
	expected = startOfDay

	result = GetStartOfDay(startOfDay)
	if !result.Equal(expected) {
		t.Errorf("GetStartOfDay(%v) = %v, 期望 %v", startOfDay, result, expected)
	}

	// 测试带时区的时间
	localTime := time.Date(2023, 5, 15, 12, 30, 45, 123456789, time.Local)
	expected = time.Date(2023, 5, 15, 0, 0, 0, 0, time.Local)

	result = GetStartOfDay(localTime)
	if !result.Equal(expected) {
		t.Errorf("GetStartOfDay(%v) = %v, 期望 %v", localTime, result, expected)
	}
}

func TestGetEndOfDay(t *testing.T) {
	// 测试任意时间点
	testTime := time.Date(2023, 5, 15, 12, 30, 45, 123456789, time.UTC)
	expected := time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.UTC)

	result := GetEndOfDay(testTime)
	if !result.Equal(expected) {
		t.Errorf("GetEndOfDay(%v) = %v, 期望 %v", testTime, result, expected)
	}

	// 测试当天结束时间
	endOfDay := time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.UTC)
	expected = endOfDay

	result = GetEndOfDay(endOfDay)
	if !result.Equal(expected) {
		t.Errorf("GetEndOfDay(%v) = %v, 期望 %v", endOfDay, result, expected)
	}

	// 测试带时区的时间
	localTime := time.Date(2023, 5, 15, 12, 30, 45, 123456789, time.Local)
	expected = time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.Local)

	result = GetEndOfDay(localTime)
	if !result.Equal(expected) {
		t.Errorf("GetEndOfDay(%v) = %v, 期望 %v", localTime, result, expected)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name            string
		durationStr     string
		defaultDuration time.Duration
		expected        time.Duration
	}{
		{
			name:            "有效的小时",
			durationStr:     "2h",
			defaultDuration: time.Hour,
			expected:        2 * time.Hour,
		},
		{
			name:            "有效的分钟",
			durationStr:     "30m",
			defaultDuration: time.Hour,
			expected:        30 * time.Minute,
		},
		{
			name:            "有效的混合",
			durationStr:     "1h30m",
			defaultDuration: time.Hour,
			expected:        90 * time.Minute,
		},
		{
			name:            "无效的格式",
			durationStr:     "invalid",
			defaultDuration: 5 * time.Minute,
			expected:        5 * time.Minute,
		},
		{
			name:            "空字符串",
			durationStr:     "",
			defaultDuration: 10 * time.Second,
			expected:        10 * time.Second,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseDuration(tc.durationStr, tc.defaultDuration)
			if result != tc.expected {
				t.Errorf("ParseDuration(%q, %v) = %v, 期望 %v",
					tc.durationStr, tc.defaultDuration, result, tc.expected)
			}
		})
	}
}
