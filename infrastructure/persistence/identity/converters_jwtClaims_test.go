package identity

import (
	"testing"
	"time"

	"naotodoserver/infrastructure/auth"

	"github.com/golang-jwt/jwt/v4"
)

// TestClaims2JWTClaimsVO_RejectsMissingIssuedAt 覆盖 DEF-157：
// 断言缺失/非法签发时间时必须返回显式错误，**不得解引用 nil 指针 panic**。
// 修复前的实际表现见红基线（panic: invalid memory address or nil pointer dereference）。
func TestClaims2JWTClaimsVO_RejectsMissingIssuedAt(t *testing.T) {
	cases := []struct {
		name   string
		claims *auth.Claims
	}{
		{name: "claims 为 nil", claims: nil},
		{name: "iat 缺失（IssuedAt 为 nil）", claims: &auth.Claims{Id: 42, Payload: "p"}},
		{
			name: "iat 为 null（解析成零值时间）",
			claims: &auth.Claims{
				Id:               42,
				Payload:          "p",
				RegisteredClaims: jwt.RegisteredClaims{IssuedAt: &jwt.NumericDate{}},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vo, err := Claims2JWTClaimsVO(c.claims)
			if err == nil {
				t.Fatalf("缺少/非法签发时间必须返回错误，实际 vo=%+v", vo)
			}
			if vo != nil {
				t.Fatalf("出错时不得返回半成品令牌对象，实际 vo=%+v", vo)
			}
		})
	}
}

// TestClaims2JWTClaimsVO_ValidClaims 正向：字段齐全时正常转换。
func TestClaims2JWTClaimsVO_ValidClaims(t *testing.T) {
	claims := &auth.Claims{
		Id:      42,
		Payload: "someone@test.local",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Subject:  "Token",
		},
	}
	vo, err := Claims2JWTClaimsVO(claims)
	if err != nil {
		t.Fatalf("合法断言不应返回错误: %v", err)
	}
	if vo == nil || vo.UserId != 42 || vo.Email != "someone@test.local" {
		t.Fatalf("转换结果不正确: %+v", vo)
	}
	if vo.IssuedAt.Unix() != claims.IssuedAt.Unix() {
		t.Fatalf("签发时间应透传: got %v want %v", vo.IssuedAt, claims.IssuedAt.Time)
	}
}
