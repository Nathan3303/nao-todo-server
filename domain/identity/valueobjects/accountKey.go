package valueobjects

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// accountKeyLength 账号标识取 SHA-256 十六进制前 16 位（64 bit）：
// 碰撞概率可忽略，且不可逆，可同时作为计数键与日志标识（不落邮箱明文）。
const accountKeyLength = 16

// AccountKey 将邮箱规范化为账号维度节流标识（SEC-04）。
// 去首尾空白 + 转小写后取 SHA-256 前 16 位十六进制，同一账号的不同大小写写法落到同一键。
// @param email 用户邮箱（任意大小写）
// @return 账号标识（16 位十六进制），不含邮箱明文
func AccountKey(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])[:accountKeyLength]
}
