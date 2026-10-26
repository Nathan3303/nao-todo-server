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
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
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

	if err := db.AutoMigrate(
		&models.UserSession{},
		&models.UserConfig{},
		&models.ProjectPreference{},
	); err != nil {
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

// seedSessionRow 直插一行会话（可指定过期时间，可选置为已登出），用于清理谓词断言
func seedSessionRow(
	t *testing.T,
	userId int64,
	token string,
	expiredAt time.Time,
	deletedAt *time.Time,
) {
	t.Helper()
	row := &models.UserSession{
		UserId:     userId,
		Token:      token,
		ExpiredAt:  expiredAt,
		DeviceType: "test",
		IP4:        "127.0.0.1",
		Region:     "测试",
	}
	if err := testDB.Create(row).Error; err != nil {
		t.Fatalf("种子 %s 写入失败: %v", token, err)
	}
	if deletedAt != nil {
		err := testDB.Model(&models.UserSession{}).
			Where("id = ?", row.ID).
			Update("deleted_at", *deletedAt).
			Error
		if err != nil {
			t.Fatalf("种子 %s 置登出失败: %v", token, err)
		}
	}
}

// sessionRowExists 物理存在性（Unscoped ⇒ 含软删行），用于确认「真的从表里删除」
func sessionRowExists(t *testing.T, token string) bool {
	t.Helper()
	var count int64
	if err := testDB.
		Unscoped().
		Model(&models.UserSession{}).
		Where("token = ?", token).
		Count(&count).
		Error; err != nil {
		t.Fatalf("查询 %s 失败: %v", token, err)
	}
	return count > 0
}

// TestDeleteExpiredBefore_OnlyBeyondGrace 过期清理：仅越过 7d 宽限且未登出的行被物理删除
func TestDeleteExpiredBefore_OnlyBeyondGrace(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	now := time.Now()
	revokedAt := now.Add(-2 * time.Hour)
	seedSessionRow(t, 4001, "expired-8d", now.AddDate(0, 0, -8), nil)
	seedSessionRow(t, 4002, "expired-6d", now.AddDate(0, 0, -6), nil)
	seedSessionRow(t, 4003, "active", now.AddDate(0, 0, 1), nil)
	// 过期 8d 但 2h 前才登出：登出宽限期未到 ⇒ 过期谓词不得回收
	seedSessionRow(t, 4004, "expired-8d-revoked-2h", now.AddDate(0, 0, -8), &revokedAt)

	removed, err := repo.DeleteExpiredBefore(context.Background(), now.AddDate(0, 0, -7), 500)
	if err != nil {
		t.Fatalf("DeleteExpiredBefore: %v", err)
	}
	if removed != 1 {
		t.Fatalf("删除行数 = %d, want 1", removed)
	}
	if sessionRowExists(t, "expired-8d") {
		t.Fatal("过期 8d 的行应被物理删除")
	}
	for _, token := range []string{"expired-6d", "active", "expired-8d-revoked-2h"} {
		if !sessionRowExists(t, token) {
			t.Fatalf("%s 不应被删除", token)
		}
	}
}

// TestDeleteRevokedBefore_OnlyBeyondGrace 登出清理：仅越过 24h 宽限的行被物理删除
func TestDeleteRevokedBefore_OnlyBeyondGrace(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	now := time.Now()
	revoked2d := now.Add(-48 * time.Hour)
	revoked2h := now.Add(-2 * time.Hour)
	seedSessionRow(t, 4101, "revoked-2d", revoked2d, &revoked2d)
	seedSessionRow(t, 4102, "revoked-2h", revoked2h, &revoked2h)
	seedSessionRow(t, 4103, "active", now.AddDate(0, 0, 1), nil)

	removed, err := repo.DeleteRevokedBefore(context.Background(), now.Add(-24*time.Hour), 500)
	if err != nil {
		t.Fatalf("DeleteRevokedBefore: %v", err)
	}
	if removed != 1 {
		t.Fatalf("删除行数 = %d, want 1", removed)
	}
	if sessionRowExists(t, "revoked-2d") {
		t.Fatal("登出 2d 的行应被物理删除")
	}
	for _, token := range []string{"revoked-2h", "active"} {
		if !sessionRowExists(t, token) {
			t.Fatalf("%s 不应被删除", token)
		}
	}
}

// TestDeleteExpiredBefore_Limit 批量上限：raw DELETE ... LIMIT 每批最多删 limit 行
func TestDeleteExpiredBefore_Limit(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	now := time.Now()
	cutoff := now.AddDate(0, 0, -7)
	var tokens []string
	for i := 0; i < 5; i++ {
		token := fmt.Sprintf("batch-expired-%d", i)
		tokens = append(tokens, token)
		seedSessionRow(t, int64(4200+i), token, now.AddDate(0, 0, -8), nil)
	}

	for _, want := range []int64{2, 2, 1} {
		removed, err := repo.DeleteExpiredBefore(context.Background(), cutoff, 2)
		if err != nil {
			t.Fatalf("DeleteExpiredBefore: %v", err)
		}
		if removed != want {
			t.Fatalf("单批删除行数 = %d, want %d", removed, want)
		}
	}
	for _, token := range tokens {
		if sessionRowExists(t, token) {
			t.Fatalf("%s 应已被删除", token)
		}
	}
}

// sessionRawRow 会话行原始状态（raw 查询绕过 GORM 软删作用域）
type sessionRawRow struct {
	ID        int64
	Token     string
	DeviceId  sql.NullString
	ExpiredAt time.Time
	DeletedAt sql.NullTime
}

// readSessionRow 按 user_id + token 读一条会话行（含已软删行），不存在即失败
func readSessionRow(t *testing.T, userId types.UserID, token string) sessionRawRow {
	t.Helper()
	var row sessionRawRow
	err := testDB.Raw(
		"SELECT id, token, device_id, expired_at, deleted_at FROM user_sessions "+
			"WHERE user_id = ? AND token = ? LIMIT 1",
		int64(userId),
		token,
	).Scan(&row).Error
	if err != nil {
		t.Fatalf("读取会话行 %s 失败: %v", token, err)
	}
	if row.ID == 0 {
		t.Fatalf("会话行不存在: user_id=%d token=%s", userId, token)
	}
	return row
}

// countSessionRows 统计该用户全部会话行数（含已软删行）
func countSessionRows(t *testing.T, userId types.UserID) int64 {
	t.Helper()
	var count int64
	if err := testDB.
		Unscoped().
		Model(&models.UserSession{}).
		Where("user_id = ?", int64(userId)).
		Count(&count).
		Error; err != nil {
		t.Fatalf("统计会话行数失败: %v", err)
	}
	return count
}

// TestCreate_RevivesSoftDeletedSameDevice 同 (user_id, device_id) 存在墓碑行时，signin 必须复活该行
// 回归 nao-todo-server#58：upsert 未回置 deleted_at ⇒ 新 token 写进「已删」行，
// 鉴权查询（deleted_at IS NULL）永远命中不到 ⇒ 客户端拿到的 token 一律 10041、checkin 一律 10022
func TestCreate_RevivesSoftDeletedSameDevice(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(5001)
	deviceCtx := ctxWithClient("dev-tombstone")

	// 1. 首次登录 → 活行
	if err := repo.Create(deviceCtx, &entities.UserSession{UserId: userID, Token: "token-old"}); err != nil {
		t.Fatalf("首次 Create: %v", err)
	}
	oldRow := readSessionRow(t, userID, "token-old")
	// 2. 登出（软删，制造墓碑）
	if err := repo.Delete(context.Background(), userID, "token-old"); err != nil {
		t.Fatalf("Delete(登出): %v", err)
	}
	tomb := readSessionRow(t, userID, "token-old")
	if !tomb.DeletedAt.Valid {
		t.Fatalf("前置不成立：登出后 deleted_at 应非 NULL")
	}
	t.Logf("复活前：id=%d token=%s deleted_at=%s",
		tomb.ID, tomb.Token, tomb.DeletedAt.Time.Format(time.RFC3339))

	// 3. 同设备重新登录 → upsert 命中墓碑行并复活
	if err := repo.Create(deviceCtx, &entities.UserSession{UserId: userID, Token: "token-new"}); err != nil {
		t.Fatalf("重登 Create 应成功（墓碑应被复活），got err: %v", err)
	}
	revived := readSessionRow(t, userID, "token-new")
	if revived.DeletedAt.Valid {
		t.Fatalf("墓碑未复活：deleted_at=%s", revived.DeletedAt.Time.Format(time.RFC3339))
	}
	if revived.ID != oldRow.ID {
		t.Fatalf("应复用原会话行（id 不变），got id=%d want %d", revived.ID, oldRow.ID)
	}
	if got := countSessionRows(t, userID); got != 1 {
		t.Fatalf("复活不应新增行，got %d 行 want 1", got)
	}
	// 4. 鉴权路径（软删作用域 + 过期判定）必须命中新 token
	session := repo.FindByUserIdAndToken(context.Background(), userID, "token-new")
	if !session.IsValid() {
		t.Fatalf("复活后新 token 仍不可被鉴权查询命中（session=%+v）", session)
	}
	t.Logf("复活后：id=%d token=%s deleted_at=<NULL>，鉴权查询命中 session_id=%d",
		revived.ID, revived.Token, session.Id)
}

// TestCreate_InsertBranchUnaffected 无墓碑行时行为不变：新增一行且立即可鉴权
func TestCreate_InsertBranchUnaffected(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(5002)
	deviceCtx := ctxWithClient("dev-fresh")

	if err := repo.Create(deviceCtx, &entities.UserSession{UserId: userID, Token: "token-fresh"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	row := readSessionRow(t, userID, "token-fresh")
	if row.DeletedAt.Valid {
		t.Fatalf("新建设备行不应是软删行")
	}
	if !row.DeviceId.Valid || row.DeviceId.String != "dev-fresh" {
		t.Fatalf("device_id 应落库为 dev-fresh，got %v", row.DeviceId)
	}
	if got := countSessionRows(t, userID); got != 1 {
		t.Fatalf("应有 1 行，got %d", got)
	}
	if !repo.FindByUserIdAndToken(context.Background(), userID, "token-fresh").IsValid() {
		t.Fatal("新 token 应可被鉴权查询命中")
	}
}

// TestCreate_OtherDeviceTombstoneUntouched 只复活本次登录的设备行，其他设备的墓碑不受影响
func TestCreate_OtherDeviceTombstoneUntouched(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(5003)
	otherCtx := ctxWithClient("dev-other")
	selfCtx := ctxWithClient("dev-self")

	// dev-other 先登录再登出 → 墓碑
	if err := repo.Create(otherCtx, &entities.UserSession{UserId: userID, Token: "token-other"}); err != nil {
		t.Fatalf("Create(dev-other): %v", err)
	}
	if err := repo.Delete(context.Background(), userID, "token-other"); err != nil {
		t.Fatalf("Delete(dev-other): %v", err)
	}
	// dev-self 登录
	if err := repo.Create(selfCtx, &entities.UserSession{UserId: userID, Token: "token-self"}); err != nil {
		t.Fatalf("Create(dev-self): %v", err)
	}
	// dev-self 再次登录（同设备 upsert 走更新分支）
	if err := repo.Create(selfCtx, &entities.UserSession{UserId: userID, Token: "token-self-2"}); err != nil {
		t.Fatalf("Create(dev-self 二次): %v", err)
	}

	if !readSessionRow(t, userID, "token-other").DeletedAt.Valid {
		t.Fatal("其他设备的墓碑行不应被复活")
	}
	if readSessionRow(t, userID, "token-self-2").DeletedAt.Valid {
		t.Fatal("本次登录的设备行不应为软删行")
	}
	if got := countSessionRows(t, userID); got != 2 {
		t.Fatalf("应有 2 行（1 活 + 1 墓碑），got %d", got)
	}
}

// TestCreate_GuardReportsInvisibleSession 写后回读校验：墓碑未复活 / 行未落库两种失败可区分
func TestCreate_GuardReportsInvisibleSession(t *testing.T) {
	cleanSessions(t)
	repo := NewSessionRepo(testDB, testCache).(*sessionRepoImpl)
	const userID = types.UserID(5004)
	deviceCtx := ctxWithClient("dev-guard")

	// 1. 行未落库 ⇒ 错误信息应指出「行未落库」
	err := repo.assertSessionVisible(context.Background(), userID, "token-never-written")
	if err == nil || !strings.Contains(err.Error(), "行未落库") {
		t.Fatalf("未落库应返回「行未落库」错误，got %v", err)
	}

	// 2. 存在墓碑行且未复活 ⇒ 错误信息应指出「软删行未复活」（并带 session_id/deleted_at）
	if err := repo.Create(deviceCtx, &entities.UserSession{UserId: userID, Token: "token-ghost"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Delete(context.Background(), userID, "token-ghost"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	err = repo.assertSessionVisible(context.Background(), userID, "token-ghost")
	if err == nil || !strings.Contains(err.Error(), "软删行未复活") {
		t.Fatalf("墓碑行应返回「软删行未复活」错误，got %v", err)
	}
	t.Logf("墓碑未复活时的报错样本: %v", err)

	// 3. 同设备重登复活后 ⇒ 校验通过
	if err := repo.Create(deviceCtx, &entities.UserSession{UserId: userID, Token: "token-alive"}); err != nil {
		t.Fatalf("重登 Create: %v", err)
	}
	if err := repo.assertSessionVisible(context.Background(), userID, "token-alive"); err != nil {
		t.Fatalf("复活后校验应通过，got %v", err)
	}
}

// TestDelete_FreshTokenRemovesSession 新鲜 token 登出 ⇒ 真删；重复登出必须报「会话不存在」而不是再次假成功
// 回归 nao-todo-server#60：Delete 只判 tx.Error、不判 RowsAffected ⇒ 删不到也回 10030「登出成功」
func TestDelete_FreshTokenRemovesSession(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(6001)

	if err := repo.Create(ctxWithClient("dev-signout"), &entities.UserSession{UserId: userID, Token: "token-fresh-so"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// ① 新鲜 token 登出 ⇒ 成功
	if err := repo.Delete(context.Background(), userID, "token-fresh-so"); err != nil {
		t.Fatalf("Delete(新鲜 token) 应成功, got %v", err)
	}
	// ③ 正常路径不变：此后鉴权查不到该 token，行是软删行
	if repo.FindByUserIdAndToken(context.Background(), userID, "token-fresh-so").IsValid() {
		t.Fatal("登出后会话不应再被鉴权查询命中")
	}
	if row := readSessionRow(t, userID, "token-fresh-so"); !row.DeletedAt.Valid {
		t.Fatal("登出应软删该行（deleted_at 非 NULL）")
	}
	// ② 重复登出（该 token 已不存在）⇒ 必须 ErrSessionNotFound，不得静默成功
	if err := repo.Delete(context.Background(), userID, "token-fresh-so"); !errors.Is(err, domerr.ErrSessionNotFound) {
		t.Fatalf("重复登出应返回 ErrSessionNotFound, got %v", err)
	}
}

// TestDelete_StaleTokenReportsNotFoundAndDeletesNothing token 已被 checkin 轮换 ⇒ 登出必须失败且不误删
// 现场语义（qa-T556 A/B 对照）：旧 token 登出若「假成功」，用户以为已退出、实际会话仍然有效
func TestDelete_StaleTokenReportsNotFoundAndDeletesNothing(t *testing.T) {
	cleanSessions(t)
	repo := newTestRepo()
	const userID = types.UserID(6002)

	if err := repo.Create(ctxWithClient("dev-rotate"), &entities.UserSession{UserId: userID, Token: "token-old-so"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// 模拟 checkin 原地轮换：token-old-so → token-new-so
	err := repo.UpdateToken(
		context.Background(),
		&entities.UserSession{UserId: userID, Token: "token-new-so"},
		"token-old-so",
	)
	if err != nil {
		t.Fatalf("UpdateToken(轮换): %v", err)
	}
	rowsBefore := countSessionRows(t, userID)

	// 用轮换前的旧 token 登出 ⇒ 必须报错（不得假成功）
	err = repo.Delete(context.Background(), userID, "token-old-so")
	if !errors.Is(err, domerr.ErrSessionNotFound) {
		t.Fatalf("旧 token 登出应返回 ErrSessionNotFound（不得假成功）, got %v", err)
	}
	// 且不误删任何行，轮换后的新 token 仍然可用
	if rowsAfter := countSessionRows(t, userID); rowsAfter != rowsBefore {
		t.Fatalf("登出失败不得误删行: rows %d → %d", rowsBefore, rowsAfter)
	}
	if !repo.FindByUserIdAndToken(context.Background(), userID, "token-new-so").IsValid() {
		t.Fatal("轮换后的新 token 不应受影响")
	}
	t.Logf("旧 token 登出被正确拒绝：err=%v，行数仍为 %d，新 token 仍可鉴权", err, rowsBefore)
}
