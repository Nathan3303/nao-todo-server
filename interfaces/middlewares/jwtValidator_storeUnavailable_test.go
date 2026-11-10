package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authApp "naotodoserver/application/auth"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/service"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/infrastructure/logging"
	identityRepo "naotodoserver/infrastructure/persistence/identity"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// deadDSN 指向必然拒绝连接的地址（127.0.0.1:1），用于构造「数据库不可用」。
const deadDSN = "root:invalid@tcp(127.0.0.1:1)/none?charset=utf8mb4&parseTime=True"

// newDeadDB 构造「打开成功、查询必然失败」的 gorm 连接。
// 两处开关缺一不可：DisableAutomaticPing 跳过 Open 后的探活；SkipInitializeWithVersion
// 跳过驱动初始化时的 SELECT VERSION()，否则 gorm.Open 就失败，构造不出「可用连接 + 故障查询」。
func newDeadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       deadDSN,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("构造死 DSN 连接失败: %v", err)
	}
	return db
}

// captureLogs 接管全局日志，返回可断言的输出缓冲。
// 生产环境由 LoadLogger 初始化；测试里必须显式接管，否则仓储层的 Warn/Error 会解引用空 Logger。
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := logging.Logger
	instance := logrus.New()
	instance.SetOutput(buf)
	instance.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	logging.Logger = instance
	t.Cleanup(func() { logging.Logger = previous })
	return buf
}

// TestJWTValidator_StoreUnavailable_Returns503 是 DEF-143 的核心回归：
// 数据库/连接类故障不得折叠成凭证失效（200 + 10041），必须回 503 + 50300。
// 链路为真实实现：真实 gorm 连接 → 真实会话仓储 → 真实领域服务 → 真实应用层 → 真实鉴权中间件。
func TestJWTValidator_StoreUnavailable_Returns503(t *testing.T) {
	logs := captureLogs(t)

	// 1. 准备真实 JWT 服务（依赖 conf 中的密钥）
	previousConf := conf.Conf
	conf.Conf = &conf.Config{Server: &conf.Server{
		JwtSecret: "def143-store-unavailable-secret",
	}}
	t.Cleanup(func() { conf.Conf = previousConf })

	jwtRepo := identityRepo.NewJWTRepo()
	token, err := jwtRepo.Generate(context.Background(), &valueobjects.JWTClaims{
		UserId:    42,
		ExpOffset: int64(time.Hour),
	})
	if err != nil {
		t.Fatalf("生成合法 JWT 失败: %v", err)
	}

	// 2. 组装真实链路（会话仓储指向死 DSN）
	identityDomain := service.NewIdentityDomain(
		jwtRepo,
		identityRepo.NewSessionRepo(newDeadDB(t), nil),
		nil,
		nil,
	)
	auth := authApp.NewAuthApp(identityDomain, nil, nil)

	// 3. 发请求
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/api/tasks", JWTValidator(auth), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	// 4. 断言：503 + 50300（不得为凭证失效）
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v（原文: %s）", err, rec.Body.String())
	}
	if rec.Code != http.StatusServiceUnavailable || body.Code != 50300 {
		t.Fatalf(
			"数据库不可用时鉴权应回 503 + 50300，实际 HTTP=%d code=%d body=%s",
			rec.Code, body.Code, rec.Body.String(),
		)
	}
	if body.Message != "服务暂时不可用，请稍后重试" {
		t.Fatalf("响应文案应面向用户且不暴露内部细节，实际: %q", body.Message)
	}

	// 5. 可观测性：必须有一条 error 级结构化事件，且带 err 原文（AC-5）
	out := logs.String()
	if !strings.Contains(out, "auth.session_lookup_failed") {
		t.Fatalf("缺少 auth.session_lookup_failed 事件，实际日志: %s", out)
	}
	if !strings.Contains(out, "level\":\"error") {
		t.Fatalf("鉴权依赖失败必须以 error 级别记录，实际日志: %s", out)
	}
	if !strings.Contains(out, "connection refused") {
		t.Fatalf("日志必须保留 err 原文（含连接失败原因），实际日志: %s", out)
	}
	// 6. 无凭证泄漏（AC-6）：不得出现 JWT 明文
	if strings.Contains(out, token) {
		t.Fatalf("日志不得包含 JWT 明文")
	}
}
