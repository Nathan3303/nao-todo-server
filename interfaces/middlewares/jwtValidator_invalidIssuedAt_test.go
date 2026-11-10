package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authApp "naotodoserver/application/auth"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/service"
	authinfra "naotodoserver/infrastructure/auth"
	identityRepo "naotodoserver/infrastructure/persistence/identity"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// signTestToken 用给定密钥签发 HS256 令牌（用于构造签名正确、断言残缺/非法的手工 token）
func signTestToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("签发测试 token 失败: %v", err)
	}
	return signed
}

// validClaims 构造「除 iat 外都合法」的断言。
func validClaims() jwt.MapClaims {
	now := time.Now()
	return jwt.MapClaims{
		"Id":      42,
		"Payload": "def157@test.local",
		"sub":     "Token",
		"iat":     now.Unix(),
		"exp":     now.Add(time.Hour).Unix(),
	}
}

// TestJWTValidator_InvalidIssuedAt_CredentialFailure 覆盖 DEF-157（红基线）：
// 断言缺少/非法 iat 时，鉴权必须走既有失败路径回 200 + 10041，
// **不得**因解引用 nil 指针 panic（表现为 HTTP 500 空体 / 连接被 gin Recovery 兜住）。
// 链表为真实实现：真实 JWT 服务 → 真实仓储适配 → 真实领域服务 → 真实应用层 → 真实中间件。
func TestJWTValidator_InvalidIssuedAt_CredentialFailure(t *testing.T) {
	previousConf := conf.Conf
	conf.Conf = &conf.Config{Server: &conf.Server{JwtSecret: "def157-claims-secret"}}
	t.Cleanup(func() { conf.Conf = previousConf })

	// 真实 JWT 仓库为单例：用它自己持有的密钥签名，避免受其它用例初始化顺序影响
	jwtRepo := identityRepo.NewJWTRepo()
	secret := string(authinfra.GetJWTService().Secret)
	identityDomain := service.NewIdentityDomain(jwtRepo, nil, nil, nil)
	app := authApp.NewAuthApp(identityDomain, nil, nil)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/api/tasks", JWTValidator(app), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	missingIat := validClaims()
	delete(missingIat, "iat")
	nullIat := validClaims()
	nullIat["iat"] = nil
	nonNumericIat := validClaims()
	nonNumericIat["iat"] = "2026-01-01T00:00:00Z"

	cases := []struct {
		name  string
		token string
	}{
		{name: "缺少 iat", token: signTestToken(t, secret, missingIat)},
		{name: "iat 为 null", token: signTestToken(t, secret, nullIat)},
		{name: "iat 为非数字", token: signTestToken(t, secret, nonNumericIat)},
		{name: "非法签名（真失效回归）", token: signTestToken(t, secret+"-forged", validClaims())},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
			req.Header.Set("Authorization", "Bearer "+c.token)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			var body struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("响应体不是合法 JSON（疑似 panic 走了 500 空体）: %v 原文=%q", err, rec.Body.String())
			}
			if rec.Code != http.StatusOK || body.Code != 10041 {
				t.Fatalf("应回 200 + 10041（凭证类），实际 HTTP=%d code=%d 原文=%q",
					rec.Code, body.Code, rec.Body.String())
			}
		})
	}
}

// TestJWTRepoParse_InvalidIssuedAt_ReturnsError 直击解析层（红基线）：
// Parse 必须返回错误而不是 panic，并且不得暴露 token 明文。
func TestJWTRepoParse_InvalidIssuedAt_ReturnsError(t *testing.T) {
	previousConf := conf.Conf
	conf.Conf = &conf.Config{Server: &conf.Server{JwtSecret: "def157-claims-secret"}}
	t.Cleanup(func() { conf.Conf = previousConf })

	jwtRepo := identityRepo.NewJWTRepo()
	secret := string(authinfra.GetJWTService().Secret)
	missingIat := validClaims()
	delete(missingIat, "iat")
	token := signTestToken(t, secret, missingIat)

	claims, err := jwtRepo.Parse(context.Background(), token)
	if err == nil {
		t.Fatalf("缺少 iat 时 Parse 必须返回错误，实际 claims=%+v", claims)
	}
	if claims != nil {
		t.Fatalf("出错时不得返回半成品断言，实际 %+v", claims)
	}
}
