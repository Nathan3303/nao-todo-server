//go:build integration

// F-06 R1 领域服务 + 仓储全链：显式 sortId 必须原样落库（pull 读回），
// 未提供 / 显式 0 维持「组末 max+1」，二次推送以最后一次显式值为准。
package project

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"

	"naotodoserver/domain/project/repositories"
	projService "naotodoserver/domain/project/service"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/infrastructure/persistence/cache"
)

const sortIdDomainTestUser = int64(3013)

func cleanSortIdDomainUser(t *testing.T) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM project_preferences WHERE user_id = ?",
		"DELETE FROM projects WHERE user_id = ?",
	} {
		if err := testDB.Exec(stmt, sortIdDomainTestUser).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// TestProjectDomainCreateSortIdIntegration F-06 R1 ①–④（真 MySQL，走 domain.Create → repo → ListSync）
func TestProjectDomainCreateSortIdIntegration(t *testing.T) {
	cleanSortIdDomainUser(t)
	addr := os.Getenv("NAO_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	repo := NewProjectRepo(testDB, cache.NewCache(redis.NewClient(&redis.Options{Addr: addr})))
	domain := projService.NewProjectDomain(repo, NewProjectPreferenceRepo(testDB))
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 基线：显式 255
	if _, _, err := domain.Create(ctx, &valueobjects.CreateProject{
		UserId: sortIdDomainTestUser, Name: "基线", SortId: 255, SortIdSet: true,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("创建基线: %v", err)
	}

	// ① 显式 7 ⇒ 落库 7（既有缺陷：会被 max+1 = 256 覆盖）
	explicit, _, err := domain.Create(ctx, &valueobjects.CreateProject{
		UserId: sortIdDomainTestUser, Name: "显式7", SortId: 7, SortIdSet: true,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("显式 7 创建: %v", err)
	}
	if explicit.SortId != 7 {
		t.Fatalf("显式 7 被改写为 %d", explicit.SortId)
	}
	assertProjectSortId(t, repo, explicit.Id, 7)

	// ④ 二次推送同 id、不同显式值 ⇒ 以最后一次为准
	second, _, err := domain.Create(ctx, &valueobjects.CreateProject{
		UserId: sortIdDomainTestUser, Id: explicit.Id, Name: "显式10",
		SortId: 10, SortIdSet: true,
		CreatedAt: explicit.CreatedAt, UpdatedAt: now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatalf("二次显式 10: %v", err)
	}
	if second.SortId != 10 {
		t.Fatalf("二次显式值应为 10，实际 %d", second.SortId)
	}
	assertProjectSortId(t, repo, explicit.Id, 10)

	// ② 未提供 ⇒ 组末 max+1（既有语义不变）
	maxBefore := repo.GetMaxSortId(ctx, sortIdDomainTestUser)
	absent, _, err := domain.Create(ctx, &valueobjects.CreateProject{
		UserId: sortIdDomainTestUser, Name: "未提供", CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("未提供 sortId 创建: %v", err)
	}
	if absent.SortId != maxBefore+1 {
		t.Fatalf("未提供时应为 max+1=%d，实际 %d", maxBefore+1, absent.SortId)
	}

	// ③ 显式 0（约定 = 未设置）⇒ 组末 max+1
	maxBefore = repo.GetMaxSortId(ctx, sortIdDomainTestUser)
	zero, _, err := domain.Create(ctx, &valueobjects.CreateProject{
		UserId: sortIdDomainTestUser, Name: "显式0",
		SortId: 0, SortIdSet: true, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("显式 0 创建: %v", err)
	}
	if zero.SortId != maxBefore+1 {
		t.Fatalf("显式 0 应按未设置处理（max+1=%d），实际 %d", maxBefore+1, zero.SortId)
	}
}

// assertProjectSortId 通过 sync 拉取路径（ListSync）读回库中 sortId
func assertProjectSortId(t *testing.T, repo repositories.Project, id int64, want uint16) {
	t.Helper()
	list, err := repo.ListSync(context.Background(), sortIdDomainTestUser, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	for _, p := range list {
		if p.Id == id {
			if p.SortId != want {
				t.Fatalf("pull 读回 sortId = %d, want %d", p.SortId, want)
			}
			return
		}
	}
	t.Fatalf("ListSync 未返回项目 %d", id)
}
