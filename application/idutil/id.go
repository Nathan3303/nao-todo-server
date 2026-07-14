package idutil

import "strconv"

// ParseID 将字符串 ID 转换为 int64
func ParseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// FormatID 将 int64 或自定义类型 ID 转换为字符串
func FormatID[T ~int64](id T) string {
	return strconv.FormatInt(int64(id), 10)
}
