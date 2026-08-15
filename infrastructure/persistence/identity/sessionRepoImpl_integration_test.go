//go:build integration

// 集成测试：需要真实 MySQL 与 Redis（docker 一次性容器）
//
//	启动方式：
//	docker run -d --name nao-todo-test-mysql \
//	  -e MYSQL_ROOT_PASSWORD=dev_password -e MYSQL_DATABASE=nao_todo_test \
//	  -p 3307:3306 mysql:8.4.9
//	docker run -d --name nao-todo-test-redis -p 6379:6379 redis:7
//
//	运行：go test -tags integration ./infrastructure/persistence/identity/...
//	可用 NAO_TEST_MYSQL_DSN / NAO_TEST_REDIS_ADDR 覆盖连接串
package identity

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB    *gorm.DB
	testCache *cache.Cache
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("NAO_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:dev_password@tcp(127.0.0.1:3307)/nao_todo_test?parseTime=true&loc=UTC&charset=utf8mb4"
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

	if err := db.AutoMigrate(&models.UserSession{}); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}
	// 初始化雪花节点（生产在 main.go 由 models.InitSnowflake 调用），
	// 否则 UserSession.BeforeCreate 生成 ID 时会空指针 panic
	models.InitSnowflake(1)
	testDB = db

	redisAddr := os.Getenv("NAO_TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	rds := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rds.Ping(context.Background()).Err(); err != nil {
		// 会话缓存断言不在本测试范围内，Redis 不可用仅记录警告，不阻断 DB 行为测试
		fmt.Printf("警告：Redis 不可用（会话缓存操作将静默降级）: %v\n", err)
	}
	testCache = cache.NewCache(rds)

	os.Exit(m.Run())
}

func newTestRepo() repositories.UserSession {
	return NewSessionRepo(testDB, testCache)
}

func cleanSessions(t *testing.T) {
	t.Helper()
	if err := testDB.Exec("DELETE FROM user_sessions").Error; err != nil {
		t.Fatalf("清理失败: %v", err)
	}
}

func ctxWithClient(deviceId string) context.Context {
	return iCtx.SetClientInfo(context.Background(), iCtx.ClientInfo{
		DeviceId:   deviceId,
		IP4:        "127.0.0.1",
		IPRegion:   "测试",
		DeviceType: "Chrome",
	})
}

// TestCreate_MultiSessionAndDeviceDedup 同设备去重、不同设备与空设备多会话
func TestCreate_MultiSessionAndDeviceDedup(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(1001)

	// 同设备两次 → 去重为一条，token 覆盖
	if err := repo.Create(ctxWithClient("dev-A"), &entities.UserSession{UserId: userID, Token: "t1"}); err != nil {
		t.Fatalf("dev-A 首次 Create: %v", err)
	}
	if err := repo.Create(ctxWithClient("dev-A"), &entities.UserSession{UserId: userID, Token: "t2"}); err != nil {
		t.Fatalf("dev-A 重复 Create: %v", err)
	}
	// 不同设备 → 新增
	if err := repo.Create(ctxWithClient("dev-B"), &entities.UserSession{UserId: userID, Token: "t3"}); err != nil {
		t.Fatalf("dev-B Create: %v", err)
	}
	// 无 deviceId（空）→ 每次新增
	if err := repo.Create(ctxWithClient(""), &entities.UserSession{UserId: userID, Token: "t4"}); err != nil {
		t.Fatalf("空设备 Create t4: %v", err)
	}
	if err := repo.Create(ctxWithClient(""), &entities.UserSession{UserId: userID, Token: "t5"}); err != nil {
		t.Fatalf("空设备 Create t5: %v", err)
	}

	var rows []models.UserSession
	testDB.Where("user_id = ?", int64(userID)).Find(&rows)
	// 期望 4 条：dev-A(1) + dev-B(1) + NULL(2)
	if len(rows) != 4 {
		t.Fatalf("期望 4 条会话, got %d", len(rows))
	}

	var devA models.UserSession
	testDB.Where("user_id = ? AND device_id = ?", int64(userID), "dev-A").First(&devA)
	if devA.Token != "t2" {
		t.Fatalf("dev-A token 应为 t2, got %q", devA.Token)
	}
}

// TestFindByUserId_OnlyUnexpired 列表仅返回未过期会话
func TestFindByUserId_OnlyUnexpired(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(1002)

	for _, d := range []struct {
		device string
		token  string
	}{
		{"d1", "a"},
		{"d2", "b"},
		{"d3", "c"},
	} {
		if err := repo.Create(ctxWithClient(d.device), &entities.UserSession{UserId: userID, Token: d.token}); err != nil {
			t.Fatalf("Create %s: %v", d.device, err)
		}
	}
	// 将 d2 置为过期
	testDB.Model(&models.UserSession{}).
		Where("user_id = ? AND device_id = ?", int64(userID), "d2").
		Update("expired_at", time.Now().Add(-time.Minute))

	sessions, err := repo.FindByUserId(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindByUserId: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("期望 2 条未过期会话, got %d", len(sessions))
	}
	for _, s := range sessions {
		if s.DeviceId == "d2" {
			t.Fatal("过期会话不应被返回")
		}
	}
}

// TestDeleteById_Scoped 按 ID 删除仅限本人，不越权
func TestDeleteById_Scoped(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	u1 := types.UserID(2001)
	u2 := types.UserID(2002)

	if err := repo.Create(ctxWithClient("d1"), &entities.UserSession{UserId: u1, Token: "t1"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctxWithClient("d2"), &entities.UserSession{UserId: u2, Token: "t2"}); err != nil {
		t.Fatal(err)
	}

	var s1, s2 models.UserSession
	testDB.Where("user_id = ?", int64(u1)).First(&s1)
	testDB.Where("user_id = ?", int64(u2)).First(&s2)

	// 删除自己的会话 → 成功
	if err := repo.DeleteById(context.Background(), u1, s1.ID); err != nil {
		t.Fatalf("删除本人会话: %v", err)
	}
	// 删除他人会话 → ErrSessionNotFound 且不生效
	if err := repo.DeleteById(context.Background(), u1, s2.ID); !errors.Is(err, domerr.ErrSessionNotFound) {
		t.Fatalf("删除他人会话应返回 ErrSessionNotFound, got %v", err)
	}
	var cnt int64
	testDB.Model(&models.UserSession{}).Where("id = ?", s2.ID).Count(&cnt)
	if cnt != 1 {
		t.Fatal("他人会话不应被删除")
	}
	// 删除不存在 → ErrSessionNotFound
	if err := repo.DeleteById(context.Background(), u1, 999999); !errors.Is(err, domerr.ErrSessionNotFound) {
		t.Fatalf("删除不存在会话应返回 ErrSessionNotFound, got %v", err)
	}
}

// TestDeleteByUserIdExceptToken 批量退出仅保留当前 token
func TestDeleteByUserIdExceptToken(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(3001)

	for _, d := range []struct {
		device string
		token  string
	}{
		{"d1", "t1"},
		{"d2", "t2"},
		{"d3", "t3"},
	} {
		if err := repo.Create(ctxWithClient(d.device), &entities.UserSession{UserId: userID, Token: d.token}); err != nil {
			t.Fatalf("Create %s: %v", d.device, err)
		}
	}

	if err := repo.DeleteByUserIdExceptToken(context.Background(), userID, "t2"); err != nil {
		t.Fatalf("DeleteByUserIdExceptToken: %v", err)
	}

	var tokens []string
	testDB.Model(&models.UserSession{}).Where("user_id = ?", int64(userID)).Pluck("token", &tokens)
	if len(tokens) != 1 || tokens[0] != "t2" {
		t.Fatalf("仅应保留 t2, got %v", tokens)
	}
}
