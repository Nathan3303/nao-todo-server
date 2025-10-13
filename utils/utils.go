package utils

import (
	"strings"
	"unicode"
)

// ToCamelCase 将大驼峰转换为小驼峰
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}
	// 将第一个字符转为小写
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// ToSnakeCase 将驼峰命名（大小驼峰均可）转换为下划线命名
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	n := len(s)

	for i, r := range s {
		// 如果是大写字母
		if unicode.IsUpper(r) {
			// 如果不是第一个字符，并且前一个字符不是大写（处理如 HTTP -> http），则加下划线
			if i > 0 &&
				(i < n-1 && !unicode.IsUpper(rune(s[i-1])) || // 前一个不是大写，如 XMLHttp -> xml_http
					i > 1 && !unicode.IsUpper(rune(s[i-2]))) { // 连续大写后接大写，如 HTTP -> http_
				result.WriteRune('_')
			}
			// 写入小写字符
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
