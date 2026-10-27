//go:build integration

// SEC-04 + SEC-10 / T570 端到端回归：真实 MySQL + Redis，装配与生产一致的
// 路由（IP 限流 → 登录控制器 → 应用 → 领域 → 仓储）。
//
//	docker run -d --name nao-todo-test-mysql -e MYSQL_ROOT_PASSWORD=dev_password \
//	  -e MYSQL_DATABASE=nao_todo_test -p 3307:3306 mysql:8.4.9
//	docker run -d --name nao-todo-test-redis -p 6379:6379 redis:7
//
// 运行：go test -tags integration -p 1 ./application/auth/...
// 可用 NAO_TEST_MYSQL_DSN / NAO_TEST_REDIS_ADDR / NAO_TEST_REDIS_PASSWORD / NAO_TEST_REDIS_DB 覆盖
package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	authApp "naotodoserver/application/auth"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/service"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/routers"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// integrationRedisDB 集成测试专用逻辑库
	integrationRedisDB = 15
	// signInQuota authRouter 中 signin 的 IP 配额（回归 10051 口径）
	signInQuota = 8
)

var (
	itDB  *gorm.DB
	itRds *redis.Client
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("NAO_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:dev_password@tcp(127.0.0.1:3307)/nao_todo_test" +
			"?parseTime=true&loc=UTC&charset=utf8mb4"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("连接测试数据库失败（需先启动测试 MySQL）: %v\n", err)
		os.Exit(1)
	}
	if err := db.AutoMigrate(&models.User{}, &models.UserSession{}); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}
	models.InitSnowflake(1)
	itDB = db

	redisAddr := os.Getenv("NAO_TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	redisDB := integrationRedisDB
	if raw := os.Getenv("NAO_TEST_REDIS_DB"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			fmt.Printf("NAO_TEST_REDIS_DB 非法: %v\n", err)
			os.Exit(1)
		}
		redisDB = parsed
	}
	itRds = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("NAO_TEST_REDIS_PASSWORD"),
		DB:       redisDB,
	})
	if err := itRds.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("连接测试 Redis 失败（本用例需真实 Redis 断言降级与节流键）: %v\n", err)
		os.Exit(1)
	}
	// 专用逻辑库：每次运行前清空，避免上一次运行的限流/节流键残留
	if err := itRds.FlushDB(context.Background()).Err(); err != nil {
		fmt.Printf("清空测试 Redis 库失败: %v\n", err)
		os.Exit(1)
	}

	// JWT 签名密钥（登录成功路径走真实签发）
	conf.Conf = &conf.Config{Server: &conf.Server{JwtSecret: strings.Repeat("k", 32)}}
	os.Exit(m.Run())
}

// uniqueEmail 生成测试专用邮箱（users 表 email/account 唯一）
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.test", prefix, time.Now().UnixNano())
}

// seedUser 直插用户（bcrypt cost 4：测试足够且快）
func seedUser(t *testing.T, email, password string) int64 {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	user := &models.User{
		Account:  email,
		Email:    email,
		Password: string(hashed),
		Nickname: "t570",
	}
	if err := itDB.Create(user).Error; err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}
	t.Cleanup(func() {
		_ = itDB.Exec("DELETE FROM user_sessions WHERE user_id = ?", user.ID).Error
		_ = itDB.Exec("DELETE FROM users WHERE id = ?", user.ID).Error
	})
	return user.ID
}

// newAuthEngine 生产同构装配：真实仓储 → 领域 → 应用 → authRouter（含 IP 限流中间件）
func newAuthEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	userRepo := identity.NewUserRepo(itDB, cache.NewCache(itRds))
	sessionRepo := identity.NewSessionRepo(itDB, cache.NewCache(itRds))
	localFallback := identity.NewLocalFallback()
	domain := service.NewIdentityDomain(
		identity.NewJWTRepo(),
		sessionRepo,
		identity.NewRateLimitRepo(itRds, localFallback),
		identity.NewLoginThrottleRepo(itRds, localFallback),
	)
	auth := authApp.NewAuthApp(domain, userRepo, sessionRepo)
	engine := gin.New()
	routers.UseAuthRouter(engine.Group("/api"), controllers.NewAuthController(auth), auth)
	return engine
}

// signInRes 登录响应的稳定字段
type signInRes struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// postSignIn 以指定客户端 IP 发一次登录请求，返回状态码与原始响应体
func postSignIn(t *testing.T, engine *gin.Engine, ip, email, password string) (int, []byte) {
	t.Helper()
	payload := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signin", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = ip
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

// parseSignInRes 解析响应体
func parseSignInRes(t *testing.T, body []byte) signInRes {
	t.Helper()
	res := signInRes{}
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, body)
	}
	return res
}

// captureLogs 捕获应用层结构化日志（JSON）
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := logging.Logger
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	logging.Logger = logger
	t.Cleanup(func() { logging.Logger = previous })
	return buf
}

// logEntries 解析 JSON 日志行
func logEntries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	entries := make([]map[string]any, 0, 8)
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entry := map[string]any{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("日志不是结构化 JSON: %v（line=%s）", err, line)
		}
		entries = append(entries, entry)
	}
	return entries
}

// findLogEntry 取指定 event 的第一条日志
func findLogEntry(entries []map[string]any, event string) map[string]any {
	for _, entry := range entries {
		if entry["event"] == event {
			return entry
		}
	}
	return nil
}

// TestSignInThrottleEndToEnd ③④⑤⑥ 连续 5 次失败 ⇒ 第 6 次暂停；
// 暂停响应与密码错误逐字节一致；日志只留哈希；到期可登录；成功后计数清零。
func TestSignInThrottleEndToEnd(t *testing.T) {
	logs := captureLogs(t)
	engine := newAuthEngine()
	ctx := context.Background()
	email := uniqueEmail("throttle")
	seedUser(t, email, "right-pass")
	clientIP := "203.0.113.60:40001"

	// ① 第 1 次失败作为基线响应
	status, baseline := postSignIn(t, engine, clientIP, email, "wrong-pass")
	if status != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", status)
	}
	if res := parseSignInRes(t, baseline); res.Code != 10012 {
		t.Fatalf("密码错误 code = %d, want 10012", res.Code)
	}

	// ② 再失败 4 次（累计 5 次）：响应仍逐字节一致
	for i := 2; i <= 5; i++ {
		_, body := postSignIn(t, engine, clientIP, email, "wrong-pass")
		if !bytes.Equal(body, baseline) {
			t.Fatalf("第 %d 次失败响应与基线不一致:\n want=%s\n got=%s", i, baseline, body)
		}
	}

	// ③ 第 6 次：口令正确也必须按密码错误返回（暂停生效）；只截取本次请求产生的日志
	logMark := logs.Len()
	status, throttled := postSignIn(t, engine, clientIP, email, "right-pass")
	throttleLogs := logs.String()[logMark:]
	if status != http.StatusOK {
		t.Fatalf("暂停响应状态码 = %d, want 200", status)
	}
	if !bytes.Equal(throttled, baseline) {
		t.Fatalf("暂停响应必须与密码错误逐字节一致（防账号枚举）:\n want=%s\n got=%s",
			baseline, throttled)
	}
	if bytes.Contains(throttled, []byte("jwt")) {
		t.Fatalf("暂停响应不得签发凭证: %s", throttled)
	}

	// ④ 计数与暂停键落 Redis，暂停 TTL 为 5 分钟（到期自动解除）
	accountKey := valueobjects.AccountKey(email)
	failCount, err := itRds.Get(ctx, "login_fail:"+accountKey).Int64()
	if err != nil {
		t.Fatalf("读取失败计数失败: %v", err)
	}
	if failCount != 5 {
		t.Fatalf("失败计数 = %d, want 5", failCount)
	}
	ttl, err := itRds.TTL(ctx, "login_block:"+accountKey).Result()
	if err != nil {
		t.Fatalf("读取暂停 TTL 失败: %v", err)
	}
	if ttl <= 0 || ttl > 5*time.Minute {
		t.Fatalf("暂停 TTL = %v, want (0, 5m]", ttl)
	}

	// ⑤ 内部日志：暂停请求只记 login.throttled + 哈希前缀，不含邮箱明文
	throttleBuf := &bytes.Buffer{}
	throttleBuf.WriteString(throttleLogs)
	entry := findLogEntry(logEntries(t, throttleBuf), "login.throttled")
	if entry == nil {
		t.Fatalf("暂停必须记录 login.throttled（实际日志: %s）", throttleLogs)
	}
	if entry["key_tag"] != accountKey {
		t.Fatalf("key_tag = %v, want %q", entry["key_tag"], accountKey)
	}
	if strings.Contains(strings.ToLower(throttleLogs), strings.ToLower(email)) {
		t.Fatalf("暂停日志不得包含邮箱明文: %s", throttleLogs)
	}

	// ⑥ 暂停到期（此处删键等价于 TTL 到期）⇒ 正确口令可登录，且失败计数被清零
	if err := itRds.Del(ctx, "login_block:"+accountKey).Err(); err != nil {
		t.Fatalf("解除暂停失败: %v", err)
	}
	_, okBody := postSignIn(t, engine, clientIP, email, "right-pass")
	if res := parseSignInRes(t, okBody); res.Code != 10010 {
		t.Fatalf("到期后登录 code = %d, want 10010（body=%s）", res.Code, okBody)
	}
	if n, _ := itRds.Exists(ctx, "login_fail:"+accountKey).Result(); n != 0 {
		t.Fatalf("登录成功后失败计数残留 = %d, want 0", n)
	}
}

// TestSignInSuccessClearsCounterBeforeThreshold ④ 未达阈值即成功登录 ⇒ 计数清零
func TestSignInSuccessClearsCounterBeforeThreshold(t *testing.T) {
	captureLogs(t)
	engine := newAuthEngine()
	ctx := context.Background()
	email := uniqueEmail("clear")
	seedUser(t, email, "right-pass")
	clientIP := "203.0.113.61:40002"

	for i := 0; i < 3; i++ {
		_, body := postSignIn(t, engine, clientIP, email, "wrong-pass")
		if res := parseSignInRes(t, body); res.Code != 10012 {
			t.Fatalf("第 %d 次失败 code = %d, want 10012", i+1, res.Code)
		}
	}
	accountKey := valueobjects.AccountKey(email)
	if n, _ := itRds.Get(ctx, "login_fail:"+accountKey).Int64(); n != 3 {
		t.Fatalf("失败计数 = %d, want 3", n)
	}

	_, body := postSignIn(t, engine, clientIP, email, "right-pass")
	if res := parseSignInRes(t, body); res.Code != 10010 {
		t.Fatalf("成功登录 code = %d, want 10010（body=%s）", res.Code, body)
	}
	if n, _ := itRds.Exists(ctx, "login_fail:"+accountKey).Result(); n != 0 {
		t.Fatalf("成功后失败计数残留 = %d, want 0", n)
	}
}

// TestSignInIPRateLimitUnchanged ⑥ 回归：IP 维度限流仍按原配额返回 10051
func TestSignInIPRateLimitUnchanged(t *testing.T) {
	captureLogs(t)
	engine := newAuthEngine()
	clientIP := "203.0.113.62:40003"

	// 用不存在的邮箱，避免账号维度节流干扰（不存在的账号不计数）
	for i := 0; i < signInQuota; i++ {
		_, body := postSignIn(t, engine, clientIP, uniqueEmail("ratelimit"), "whatever")
		if res := parseSignInRes(t, body); res.Code != 10012 {
			t.Fatalf("第 %d 次请求 code = %d, want 10012（配额内）", i+1, res.Code)
		}
	}
	_, body := postSignIn(t, engine, clientIP, uniqueEmail("ratelimit"), "whatever")
	res := parseSignInRes(t, body)
	if res.Code != 10051 {
		t.Fatalf("超出 IP 配额 code = %d, want 10051（body=%s）", res.Code, body)
	}
}
