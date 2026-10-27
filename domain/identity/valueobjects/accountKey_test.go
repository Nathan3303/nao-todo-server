package valueobjects

import (
	"regexp"
	"strings"
	"testing"
)

var accountKeyPattern = regexp.MustCompile(`^[0-9a-f]{16}$`)

// TestAccountKeyNormalizesEmail 小写 + 去空白后同一邮箱得到同一标识（SEC-04 规范化口径）
func TestAccountKeyNormalizesEmail(t *testing.T) {
	base := AccountKey("victim@example.com")
	for _, variant := range []string{
		"victim@example.com",
		"VICTIM@example.com",
		"  Victim@Example.COM  ",
		"\tvictim@example.com\n",
	} {
		if got := AccountKey(variant); got != base {
			t.Fatalf("AccountKey(%q) = %q, want %q（应规范化到同一标识）", variant, got, base)
		}
	}
}

// TestAccountKeyIsHashedAndStable 标识为 16 位十六进制、不含邮箱明文、同输入稳定
func TestAccountKeyIsHashedAndStable(t *testing.T) {
	email := "victim@example.com"
	key := AccountKey(email)
	if !accountKeyPattern.MatchString(key) {
		t.Fatalf("AccountKey = %q, want 16 位十六进制", key)
	}
	if strings.Contains(strings.ToLower(key), "victim") || strings.Contains(key, email) {
		t.Fatalf("AccountKey = %q 不应包含邮箱明文", key)
	}
	if key != AccountKey(email) {
		t.Fatalf("AccountKey 不稳定: %q", key)
	}
	if key == AccountKey("other@example.com") {
		t.Fatal("不同邮箱不应得到同一标识")
	}
}
