//go:build integration

// SEC-12 / T570 集成回归（真实 MySQL）：跨租户资源 ID 的"不存在 / 他人 / 已删除"
// 三态响应必须逐字节相同，且不含 ORM/SQL/表名/主键等内部细节。
//
// 运行：go test -tags integration -p 1 -count=1 -run TestProjectNotFoundThreeStates -v ./interfaces/controllers/
package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	projectApp "naotodoserver/application/project"
	projectdto "naotodoserver/application/project/dto"
	domerr "naotodoserver/domain/errors"
	projectService "naotodoserver/domain/project/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	projectPersistence "naotodoserver/infrastructure/persistence/project"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// redactionLeaks 响应体不得包含的内部细节标记（SEC-12）
var redactionLeaks = []string{
	"sql", "SELECT", "record not found", "Duplicate", "PRIMARY",
	"projects", "user_sessions", "tasks", "tags",
}

// newRedactionEngine 真实仓储/应用 + 真实控制器；userId 注入为攻击者 B
func newRedactionEngine(t *testing.T, userId int64) *gin.Engine {
	t.Helper()
	repo := projectPersistence.NewProjectRepo(nullClearDB, cache.NewCache(noopRedisClient()))
	domain := projectService.NewProjectDomain(repo, nil)
	app := projectApp.NewProjectApp(domain, dbs.NewTxManager(nullClearDB), repo, nil, nil, nil)
	ctrl := NewProjectController(app)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		c := iCtx.SetRequestId(ctx.Request.Context(), "it-redaction")
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(c, userId))
		ctx.Next()
	})
	engine.GET("/api/projects/:projectId", ctrl.GetProject)
	engine.PUT("/api/projects/:projectId", ctrl.UpdateProject)
	return engine
}

// seedRedactionProject 直插清单行（deletedAt 非空 ⇒ 软删除）
func seedRedactionProject(t *testing.T, userId int64, name string, deleted bool) int64 {
	t.Helper()
	base := time.Now().UTC().Truncate(time.Second)
	p := &models.Project{
		ModelBase: models.ModelBase{ID: time.Now().UnixNano(), CreatedAt: base, UpdatedAt: base},
		UserId:    userId,
		Name:      name,
		SortId:    1000,
	}
	if deleted {
		p.DeletedAt = gorm.DeletedAt{Time: base, Valid: true}
	}
	if err := nullClearDB.Create(p).Error; err != nil {
		t.Fatalf("插入清单失败: %v", err)
	}
	t.Cleanup(func() { _ = nullClearDB.Unscoped().Delete(&models.Project{}, p.ID).Error })
	return p.ID
}

// requestRedaction 发一次请求并返回响应体
func requestRedaction(t *testing.T, engine *gin.Engine, method, path string, body string) []byte {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Body.Bytes()
}

// assertNoRedactionLeak 断言响应不含内部细节
func assertNoRedactionLeak(t *testing.T, body []byte) {
	t.Helper()
	text := string(body)
	for _, leak := range redactionLeaks {
		if strings.Contains(text, leak) {
			t.Fatalf("响应泄漏内部细节 %q: %s", leak, text)
		}
	}
}

// decodeRedactionRes 解析统一响应
func decodeRedactionRes(t *testing.T, body []byte) types.ResponseData {
	t.Helper()
	res := types.ResponseData{}
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, body)
	}
	return res
}

// TestProjectNotFoundThreeStatesGetIntegration 读取：他人 / 不存在 / 已删除 ⇒ 响应逐字节相同
func TestProjectNotFoundThreeStatesGetIntegration(t *testing.T) {
	const userA, userB int64 = 910001, 910002
	otherTenant := seedRedactionProject(t, userA, "A 的清单", false)
	deleted := seedRedactionProject(t, userB, "B 已删清单", true)

	engine := newRedactionEngine(t, userB)
	gotOther := requestRedaction(t, engine, http.MethodGet, projectPath(otherTenant), "")
	gotMissing := requestRedaction(t, engine, http.MethodGet, projectPath(-1), "")
	gotDeleted := requestRedaction(t, engine, http.MethodGet, projectPath(deleted), "")

	if !bytes.Equal(gotOther, gotMissing) || !bytes.Equal(gotOther, gotDeleted) {
		t.Fatalf("三态响应必须逐字节相同:\n 他人=%s\n 不存在=%s\n 已删除=%s",
			gotOther, gotMissing, gotDeleted)
	}
	assertNoRedactionLeak(t, gotOther)
	res := decodeRedactionRes(t, gotOther)
	if res.Code != 20002 || res.Error != "记录不存在" {
		t.Fatalf("code/error = %d/%q, want 20002/记录不存在", res.Code, res.Error)
	}
}

// TestProjectUpdateThreeStatesIdenticalIntegration 更新：他人 / 不存在 / 已删除 ⇒ 响应逐字节相同。
// 注意（既有行为，非本单改动）：仓储 Update 未校验 RowsAffected，三态当前都是「成功但零行更新」，
// 因此三态本身已无差异化信息；本用例锁定该一致性，避免后续改动引入差异响应（SEC-12）。
func TestProjectUpdateThreeStatesIdenticalIntegration(t *testing.T) {
	const userA, userB int64 = 910011, 910012
	otherTenant := seedRedactionProject(t, userA, "A 的清单2", false)
	deleted := seedRedactionProject(t, userB, "B 已删清单2", true)

	engine := newRedactionEngine(t, userB)
	body := `{"name":"被改名","description":"x"}`
	gotOther := requestRedaction(t, engine, http.MethodPut, projectPath(otherTenant), body)
	gotMissing := requestRedaction(t, engine, http.MethodPut, projectPath(-2), body)
	gotDeleted := requestRedaction(t, engine, http.MethodPut, projectPath(deleted), body)

	// 更新响应会回显请求里的清单 ID（客户端协议的一部分，与本单无关）：
	// 归一化掉该项后，三态必须逐字节相同
	normalize := func(body []byte) []byte {
		return dataFieldPattern.ReplaceAll(body, []byte(`"data":""`))
	}
	if !bytes.Equal(normalize(gotOther), normalize(gotMissing)) ||
		!bytes.Equal(normalize(gotOther), normalize(gotDeleted)) {
		t.Fatalf("三态响应（忽略回显 ID）必须逐字节相同:\n 他人=%s\n 不存在=%s\n 已删除=%s",
			gotOther, gotMissing, gotDeleted)
	}
	assertNoRedactionLeak(t, gotOther)
	if res := decodeRedactionRes(t, gotOther); res.Code != 20020 {
		t.Fatalf("code = %d, want 20020（既有静默 no-op 语义，本单不改）", res.Code)
	}
}

// dataFieldPattern 匹配响应里的 data 字段值（更新/删除类接口会回显请求 ID）
var dataFieldPattern = regexp.MustCompile(`"data":"[^"]*"`)

// TestCrossTenantDuplicateIdNoRawMysqlIntegration 跨租户同 ID 创建（sync 等价路径）
// ⇒ 领域哨兵（ID 冲突），不再出现 MySQL 1062 原文/主键值（SEC-12）
func TestCrossTenantDuplicateIdNoRawMysqlIntegration(t *testing.T) {
	const userA, userB int64 = 910021, 910022
	owned := seedRedactionProject(t, userA, "A 的清单3", false)

	repo := projectPersistence.NewProjectRepo(nullClearDB, cache.NewCache(noopRedisClient()))
	domain := projectService.NewProjectDomain(repo, nil)
	app := projectApp.NewProjectApp(domain, dbs.NewTxManager(nullClearDB), repo, nil, nil, nil)

	idStr := strconv.FormatInt(owned, 10)
	_, _, err := app.Create(t.Context(), userB, &projectdto.CreateProjectReq{
		Name: "B 抢同一个 ID",
		Id:   &idStr,
	})
	if err == nil {
		t.Fatal("跨租户同 ID 创建必须失败（ID 冲突）")
	}
	text := err.Error()
	for _, leak := range []string{"Duplicate", "PRIMARY", "sql", "Error 1062", "projects"} {
		if strings.Contains(text, leak) {
			t.Fatalf("创建冲突错误泄漏 MySQL 原文 %q: %s", leak, text)
		}
	}
	if !errors.Is(err, domerr.ErrIDConflict) {
		t.Fatalf("冲突错误应为 ErrIDConflict，实际: %v", err)
	}
}

// noopRedisClient 返回一个"连不上也不影响断言"的 Redis 客户端：
// 项目仓储仅在写路径做缓存失效，错误本身会被静默忽略（测试无需真实 Redis）
func noopRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379",
		DialTimeout: 20 * time.Millisecond,
		MaxRetries:  -1,
	})
}

// projectPath 构造清单读取/更新路径
func projectPath(id int64) string {
	return "/api/projects/" + strconv.FormatInt(id, 10)
}
