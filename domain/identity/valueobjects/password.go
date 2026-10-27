package valueobjects

import (
	"strings"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/textutils"
)

// 口令策略常量（SEC-05）：注册与改密共用同一套口径
const (
	// PasswordMinRunes 口令最短长度（按字符计，1 个 CJK 字符算 1 位）
	PasswordMinRunes = 8
	// PasswordMaxRunes 口令最长长度（按字符计）
	PasswordMaxRunes = 64
	// passwordShortDigitsMax 纯数字口令的最短长度：短于该长度视为弱口令
	passwordShortDigitsMax = 12
	// passwordMaxBytes bcrypt 只处理前 72 字节：更长的口令直接拒绝，
	// 避免"字符合规但字节超限"在加密阶段失败并回显内部错误
	passwordMaxBytes = 72
)

// weakPasswords 常见弱口令小名单（小写比较）。
// 刻意保持 ≤20 条内联常量：不引入外部字典/依赖，只挡最常见的撞库口令。
var weakPasswords = map[string]struct{}{
	"password":   {},
	"passw0rd":   {},
	"12345678":   {},
	"123456789":  {},
	"1234567890": {},
	"123456789a": {},
	"qwerty":     {},
	"qwertyuiop": {},
	"abc12345":   {},
	"a1234567":   {},
	"11111111":   {},
	"00000000":   {},
	"admin123":   {},
	"admin888":   {},
	"letmein":    {},
	"iloveyou":   {},
	"woaini1314": {},
}

// ValidatePassword 校验明文口令是否满足口令策略（SEC-05），注册与改密共用。
// 规则（按字符计）：长度 8–64；拒绝全同字符、纯数字且不足 12 位、
// 常见弱口令名单命中、与邮箱本地部分（@ 之前，忽略大小写）相同。
// 已注册用户的旧口令不在登录路径校验，保证兼容（不强制改密）。
// @param password 明文口令
// @param email 用户邮箱（可为空：为空时跳过"邮箱本地部分"规则）
// @return 满足策略返回 nil，否则返回 domerr.ErrPasswordTooWeak（对外文案由接口层统一）
func ValidatePassword(password, email string) error {
	length := textutils.RuneLength(password)
	if length < PasswordMinRunes || length > PasswordMaxRunes {
		return domerr.ErrPasswordTooWeak
	}
	if len(password) > passwordMaxBytes {
		return domerr.ErrPasswordTooWeak
	}
	if isAllSameRune(password) {
		return domerr.ErrPasswordTooWeak
	}
	if isShortDigits(password, length) {
		return domerr.ErrPasswordTooWeak
	}
	if _, hit := weakPasswords[strings.ToLower(password)]; hit {
		return domerr.ErrPasswordTooWeak
	}
	if localPart := emailLocalPart(email); localPart != "" &&
		strings.EqualFold(password, localPart) {
		return domerr.ErrPasswordTooWeak
	}
	return nil
}

// isAllSameRune 判断口令是否由同一字符重复构成（如 aaaaaaaa / 11111111）
// @param password 明文口令
// @return 全同字符返回 true
func isAllSameRune(password string) bool {
	runes := []rune(password)
	for _, r := range runes[1:] {
		if r != runes[0] {
			return false
		}
	}
	return len(runes) > 0
}

// isShortDigits 判断口令是否为纯数字且长度不足 passwordShortDigitsMax
// @param password 明文口令
// @param length 口令字符数
// @return 命中返回 true
func isShortDigits(password string, length int) bool {
	if length >= passwordShortDigitsMax {
		return false
	}
	for _, r := range password {
		if r < '0' || r > '9' {
			return false
		}
	}
	return length > 0
}

// emailLocalPart 取邮箱 @ 之前的本地部分（去首尾空白）
// @param email 邮箱
// @return 本地部分；空邮箱返回空串
func emailLocalPart(email string) string {
	trimmed := strings.TrimSpace(email)
	if idx := strings.Index(trimmed, "@"); idx >= 0 {
		return trimmed[:idx]
	}
	return trimmed
}
