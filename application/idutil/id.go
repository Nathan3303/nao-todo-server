package idutil

import "strconv"

// ParseID 将字符串 ID 转换为 int64
func ParseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// FormatID 将 int64 ID 转换为字符串
func FormatID(id int64) string {
	return strconv.FormatInt(id, 10)
}
