package logging

import (
	"crypto/sha256"
	"encoding/hex"
)

// keyTagLength 日志键标识长度（SHA-256 十六进制前 8 位）
const keyTagLength = 8

// KeyTag 生成日志可用的键标识：SHA-256 前 8 位十六进制。
// 限流 / 节流键可能含 userId、客户端 IP 或账号键，日志中只保留不可逆摘要，
// 不落明文 key、userId、token、邮箱（SEC-10 / #88 C5/C6）。
// @param key 原始计数键
// @return 8 位十六进制摘要
func KeyTag(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])[:keyTagLength]
}
