package utils

import (
	"fmt"
	"strconv"
	"time"
)

// FormatTimestamp 格式化时间戳为RFC3339格式
//
// 参数:
//   - t: 要格式化的时间
//
// 返回:
//   - string: 格式化后的时间字符串
func FormatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTimestamp 解析时间戳字符串为time.Time
//
// 参数:
//   - s: 要解析的时间字符串
//
// 返回:
//   - time.Time: 解析后的时间对象
//   - error: 解析过程中的错误，如果成功则为nil
func ParseTimestamp(s string) (time.Time, error) {
	// 常用时间格式列表，按照常见程度排序
	formats := []string{
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,      // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05", // 不带时区的ISO格式
		"2006-01-02 15:04:05", // 常见的日期时间格式
		"2006-01-02",          // 仅日期
		time.RFC1123,          // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC1123Z,         // Mon, 02 Jan 2006 15:04:05 -0700
		time.RFC822,           // 02 Jan 06 15:04 MST
		time.RFC822Z,          // 02 Jan 06 15:04 -0700
		"20060102",            // 紧凑日期格式（如YYYYMMDD）
		"2006/01/02",          // 斜杠分隔的日期
		"2006/01/02 15:04:05", // 斜杠分隔的日期时间
	}

	// 尝试使用上述格式解析
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// 尝试解析Unix时间戳（秒）
	if t, err := parseUnixTimestamp(s); err == nil {
		return t, nil
	}

	// 尝试解析Unix毫秒时间戳
	if t, err := parseUnixMillisTimestamp(s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("无法解析时间戳: %s", s)
}

// parseUnixTimestamp 尝试将字符串解析为Unix时间戳（秒）
//
// 参数:
//   - s: 要解析的Unix时间戳字符串
//
// 返回:
//   - time.Time: 解析后的时间
//   - error: 解析错误，如果成功则为nil
func parseUnixTimestamp(s string) (time.Time, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	// 验证时间戳的合理性（避免过去或将来太远的时间）
	// Unix时间戳应该在1970年至2100年之间
	if sec < 0 || sec > 4102444800 { // 4102444800是2100-01-01 00:00:00的Unix时间戳
		return time.Time{}, fmt.Errorf("Unix时间戳超出合理范围: %d", sec)
	}

	return time.Unix(sec, 0), nil
}

// parseUnixMillisTimestamp 尝试将字符串解析为Unix毫秒时间戳
//
// 参数:
//   - s: 要解析的Unix毫秒时间戳字符串
//
// 返回:
//   - time.Time: 解析后的时间
//   - error: 解析错误，如果成功则为nil
func parseUnixMillisTimestamp(s string) (time.Time, error) {
	msec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	// 检查是否可能是毫秒时间戳（长度通常为13位，且值较大）
	// Unix毫秒时间戳通常大于1000000000000（2001年9月）
	if msec < 1000000000000 || msec > 4102444800000 { // 4102444800000是2100年的毫秒时间戳
		return time.Time{}, fmt.Errorf("Unix毫秒时间戳超出合理范围: %d", msec)
	}

	return time.Unix(msec/1000, (msec%1000)*1000000), nil
}

// TimeToInterface 将时间转换为适合JSON序列化的接口类型
//
// 参数:
//   - t: 时间对象，可以是time.Time或nil
//
// 返回:
//   - interface{}: 适合JSON序列化的时间表示
func TimeToInterface(t interface{}) interface{} {
	switch v := t.(type) {
	case time.Time:
		return FormatTimestamp(v)
	case *time.Time:
		if v == nil {
			return nil
		}
		return FormatTimestamp(*v)
	case string:
		return v
	case nil:
		return nil
	default:
		return fmt.Sprintf("%v", t)
	}
}

// Now 获取当前时间，封装time.Now()
//
// 返回:
//   - time.Time: 当前时间
func Now() time.Time {
	return time.Now()
}

// FormatDuration 将持续时间格式化为可读字符串
//
// 参数:
//   - d: 持续时间
//
// 返回:
//   - string: 格式化后的持续时间字符串
func FormatDuration(d time.Duration) string {
	// 根据持续时间的长短选择不同的显示方式
	if d < time.Minute {
		return fmt.Sprintf("%.1f秒", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%d小时%d分", hours, minutes)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%d天%d小时", days, hours)
	}
}

// GetStartOfDay 获取指定时间的当天开始时间（00:00:00）
//
// 参数:
//   - t: 指定时间
//
// 返回:
//   - time.Time: 当天开始时间
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay 获取指定时间的当天结束时间（23:59:59.999999999）
//
// 参数:
//   - t: 指定时间
//
// 返回:
//   - time.Time: 当天结束时间
func GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// ParseDuration 解析表示持续时间的字符串
//
// 参数:
//   - s: 表示持续时间的字符串（例如"3h"、"5m30s"）
//   - defaultDuration: 解析失败时的默认值
//
// 返回:
//   - time.Duration: 解析后的持续时间
func ParseDuration(s string, defaultDuration time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultDuration
	}
	return d
}
