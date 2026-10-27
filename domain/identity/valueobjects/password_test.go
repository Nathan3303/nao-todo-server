package valueobjects

import (
	"errors"
	"strings"
	"testing"

	domerr "naotodoserver/domain/errors"
)

// TestValidatePasswordAcceptsCompliant 合规口令：长度达标、非全同、非短纯数字、不在名名单、非邮箱本地部分
func TestValidatePasswordAcceptsCompliant(t *testing.T) {
	cases := map[string]string{
		"八字非纯数字":   "abcd1234",
		"长口令":      "correct-horse-battery-9",
		"十二位纯数字":   "123456789012",
		"含符号":      "p@ssw0rd!go",
		"等于邮箱其他段":  "victim-local-different",
		"六十四位内长口令": strings.Repeat("ab12", 16),
	}
	for name, password := range cases {
		t.Run(name, func(t *testing.T) {
			if err := ValidatePassword(password, "victim@example.com"); err != nil {
				t.Fatalf("合规口令 %q 被拒绝: %v", password, err)
			}
		})
	}
}

// TestValidatePasswordRejectsWeak 口令策略拒绝矩阵（SEC-05）：
// 过短 / 过长 / 全同 / 纯数字短 / 名单命中 / 等于邮箱本地部分 / 超 bcrypt 字节上限
func TestValidatePasswordRejectsWeak(t *testing.T) {
	cases := map[string]struct {
		password string
		email    string
	}{
		"过短":            {strings.Repeat("a", PasswordMinRunes-1), ""},
		"过长":            {"ab12" + strings.Repeat("c", PasswordMaxRunes), ""},
		"全同字符":          {"aaaaaaaa", ""},
		"全同数字":          {"11111111", ""},
		"纯数字短":          {"12345678", ""},
		"纯数字十一位":        {"12345678901", ""},
		"名单命中":          {"password", ""},
		"名单命中大写":        {"PASSWORD", ""},
		"名单命中数字串":       {"12345678", ""},
		"名单命中 admin":    {"admin123", ""},
		"等于邮箱本地部分":      {"victim", "victim@example.com"},
		"等于邮箱本地部分大写":    {"Victim", "VICTIM@Example.com"},
		"超 bcrypt 字节上限": {strings.Repeat("密", 25), ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidatePassword(tc.password, tc.email)
			if !errors.Is(err, domerr.ErrPasswordTooWeak) {
				t.Fatalf("口令 %q 应被拒绝（want ErrPasswordTooWeak），实际 err=%v", tc.password, err)
			}
		})
	}
}

// TestValidatePasswordEmptyEmailSkipsLocalPartRule 空邮箱时跳过"等于邮箱本地部分"规则，其余规则照常
func TestValidatePasswordEmptyEmailSkipsLocalPartRule(t *testing.T) {
	if err := ValidatePassword("victim-strong-1", ""); err != nil {
		t.Fatalf("空邮箱不应因本地部分规则被拒: %v", err)
	}
	if err := ValidatePassword("aaaaaaaa", ""); !errors.Is(err, domerr.ErrPasswordTooWeak) {
		t.Fatal("空邮箱时其余弱口令规则仍须生效")
	}
}

// TestCreateUserValidateAppliesPasswordPolicy 注册路径（CreateUser.Validate）复用同一口令策略
func TestCreateUserValidateAppliesPasswordPolicy(t *testing.T) {
	vo := &CreateUser{
		Account:  "victim@example.com",
		Password: "12345678",
		Email:    "victim@example.com",
	}
	if err := vo.Validate(); !errors.Is(err, domerr.ErrPasswordTooWeak) {
		t.Fatalf("注册弱口令应返回 ErrPasswordTooWeak，实际 %v", err)
	}

	vo.Password = "victim"
	if err := vo.Validate(); !errors.Is(err, domerr.ErrPasswordTooWeak) {
		t.Fatalf("注册口令等于邮箱本地部分应被拒绝，实际 %v", err)
	}

	vo.Password = "correct-horse-9"
	if err := vo.Validate(); err != nil {
		t.Fatalf("合规口令注册不应报错: %v", err)
	}
}
