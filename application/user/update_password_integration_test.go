//go:build integration

// SEC-03 / T564 集成回归：改密成功后撤销其它设备会话（保留当前设备）。
//
// 需要真实 MySQL（可选 Redis 用于缓存降级）：
//
//	docker run -d --name nao-todo-test-mysql \
//	  -e MYSQL_ROOT_PASSWORD=dev_password -e MYSQL_DATABASE=nao_todo_test \
//	  -p 3307:3306 mysql:8.4.9
//	docker run -d --name nao-todo-test-redis -p 6379:6379 redis:7
//
// 运行：go test -tags integration ./application/user/...
// 可用 NAO_TEST_MYSQL_DSN / NAO_TEST_REDIS_ADDR 覆盖连接串
package user_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	authApp "naotodoserver/application/auth"
	userApp "naotodoserver/application/user"
	"naotodoserver/application/user/dto"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/service"
	"naotodoserver/domain/identity/valueobjects"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	itDB          *gorm.DB
	itUserRepo    repositories.User
	itSessionRepo repositories.UserSession
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
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(5)
	if err := db.AutoMigrate(&models.User{}, &models.UserSession{}); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}
	// 生产在 main.go 初始化雪花节点；会话/用户创建前必须初始化，否则 ID 生成 panic
	models.InitSnowflake(1)
	itDB = db

	redisAddr := os.Getenv("NAO_TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	rds := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rds.Ping(context.Background()).Err(); err != nil {
		// 会话缓存断言不在本测试范围：Redis 不可用时缓存操作静默降级
		fmt.Printf("警告：Redis 不可用（缓存降级）: %v\n", err)
	}

	// JWT 签名密钥（改密撤销断言走真实 Validate 路径）
	conf.Conf = &conf.Config{Server: &conf.Server{JwtSecret: strings.Repeat("k", 32)}}

	itUserRepo = identity.NewUserRepo(db, cache.NewCache(rds))
	itSessionRepo = identity.NewSessionRepo(db, cache.NewCache(rds))
	os.Exit(m.Run())
}

// uniqueEmail 生成测试专用邮箱（users 表 email/account 唯一）
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.test", prefix, time.Now().UnixNano())
}

// seedUser 直插用户（bcrypt cost 4 足够测试且快）
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
		Nickname: "t564",
	}
	if err := itDB.Create(user).Error; err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}
	return user.ID
}

// seedTwoSessions 为同一用户创建两个设备的会话，返回两个 JWT
func seedTwoSessions(t *testing.T, userId int64) (string, string) {
	t.Helper()
	jwtRepo := identity.NewJWTRepo()
	ctx := context.Background()
	tokens := make([]string, 0, 2)
	for i := 0; i < 2; i++ {
		token, err := jwtRepo.Generate(ctx, &valueobjects.JWTClaims{
			UserId: userId,
			Email:  uniqueEmail("jwt"),
		})
		if err != nil {
			t.Fatalf("生成 JWT 失败: %v", err)
		}
		if err := itSessionRepo.Create(ctx, &entities.UserSession{
			UserId: domaintypes.UserID(userId),
			Token:  token,
		}); err != nil {
			t.Fatalf("创建会话失败: %v", err)
		}
		tokens = append(tokens, token)
	}
	return tokens[0], tokens[1]
}

// countSessions 统计用户现存（未软删/未过期）会话行数
func countSessions(t *testing.T, userId int64) int64 {
	t.Helper()
	var count int64
	if err := itDB.Model(&models.UserSession{}).
		Where("user_id = ?", userId).
		Count(&count).Error; err != nil {
		t.Fatalf("统计会话失败: %v", err)
	}
	return count
}

// newPasswordEngine 挂载「鉴权中间件 + 获取用户详情」，用于断言 10041 / 10060
func newPasswordEngine(app userApp.UserApp, auth authApp.AuthApp) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewUserController(app)
	engine.GET("/api/user/profile", middlewares.JWTValidator(auth), ctrl.GetUserProfile)
	return engine
}

// profileCode 请求 /api/user/profile，返回响应体 code（10041 = 鉴权失败，10060 = 成功）
func profileCode(t *testing.T, engine *gin.Engine, token string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	var res struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	return res.Code
}

// TestUpdatePasswordRevokesOtherSessionsIntegration 改密成功：
// 其它设备 token 立即 10041，当前设备 token 仍 10060；库中仅剩当前会话，密码已更新。
func TestUpdatePasswordRevokesOtherSessionsIntegration(t *testing.T) {
	ctx := context.Background()
	userId := seedUser(t, uniqueEmail("revoke"), "old-pass-1")
	otherToken, currentToken := seedTwoSessions(t, userId)

	identityDomain := service.NewIdentityDomain(
		identity.NewJWTRepo(), itSessionRepo, nil, nil,
	)
	auth := authApp.NewAuthApp(identityDomain, itUserRepo, itSessionRepo)
	app := userApp.NewUserApp(itUserRepo, itSessionRepo, nil, nil)
	engine := newPasswordEngine(app, auth)

	if code := profileCode(t, engine, otherToken); code != 10060 {
		t.Fatalf("改密前其它设备 code = %d, want 10060", code)
	}

	err := app.UpdatePassword(ctx, userId, dto.UpdatePasswordInput{
		OldPassword: "old-pass-1",
		NewPassword: "new-pass-1",
		Token:       currentToken,
	})
	if err != nil {
		t.Fatalf("改密失败: %v", err)
	}

	if code := profileCode(t, engine, otherToken); code != 10041 {
		t.Fatalf("改密后其它设备 code = %d, want 10041（会话必须立即失效）", code)
	}
	if code := profileCode(t, engine, currentToken); code != 10060 {
		t.Fatalf("改密后当前设备 code = %d, want 10060（保留当前会话）", code)
	}
	if count := countSessions(t, userId); count != 1 {
		t.Fatalf("剩余会话 = %d, want 1", count)
	}

	user, err := itUserRepo.FindById(ctx, domaintypes.UserID(userId))
	if err != nil {
		t.Fatalf("回读用户失败: %v", err)
	}
	if !itUserRepo.PasswordCompare([]byte("new-pass-1"), []byte(user.Password)) {
		t.Fatal("新密码未生效")
	}
}

// TestUpdatePasswordFailureKeepsSessionsIntegration 旧密码错 ⇒ 不改密、不撤销任何会话
func TestUpdatePasswordFailureKeepsSessionsIntegration(t *testing.T) {
	ctx := context.Background()
	userId := seedUser(t, uniqueEmail("keep"), "old-pass-1")
	otherToken, currentToken := seedTwoSessions(t, userId)

	identityDomain := service.NewIdentityDomain(
		identity.NewJWTRepo(), itSessionRepo, nil, nil,
	)
	auth := authApp.NewAuthApp(identityDomain, itUserRepo, itSessionRepo)
	app := userApp.NewUserApp(itUserRepo, itSessionRepo, nil, nil)
	engine := newPasswordEngine(app, auth)

	err := app.UpdatePassword(ctx, userId, dto.UpdatePasswordInput{
		OldPassword: "wrong-pass",
		NewPassword: "new-pass-1",
		Token:       currentToken,
	})
	if err == nil {
		t.Fatal("旧密码错误必须返回失败")
	}
	if code := profileCode(t, engine, otherToken); code != 10060 {
		t.Fatalf("改密失败后其它设备 code = %d, want 10060（不得误杀）", code)
	}
	if code := profileCode(t, engine, currentToken); code != 10060 {
		t.Fatalf("改密失败后当前设备 code = %d, want 10060", code)
	}
	if count := countSessions(t, userId); count != 2 {
		t.Fatalf("失败后会话数 = %d, want 2", count)
	}
}
