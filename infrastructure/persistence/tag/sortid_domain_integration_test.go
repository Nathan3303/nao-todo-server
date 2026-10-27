//go:build integration

// F-06 R1 领域服务 + 仓储全链（tags）：显式 sortId 原样落库（pull 读回），
// 未提供 / 显式 0 维持 max+1，二次推送以最后一次显式值为准。
package tag

import (
	"context"
	"testing"
	"time"

	"naotodoserver/domain/tag/repositories"
	tagService "naotodoserver/domain/tag/service"
	"naotodoserver/domain/tag/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

const sortIdDomainTagUser = int64(3014)

// assertTagSortId 通过 sync 拉取路径（ListSync）读回库中 sortId
func assertTagSortId(t *testing.T, repo repositories.TagRepository, id int64, want uint16) {
	t.Helper()
	list, err := repo.ListSync(context.Background(), sortIdDomainTagUser, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	for _, tag := range list {
		if tag.Id == id {
			if tag.SortId != want {
				t.Fatalf("pull 读回 sortId = %d, want %d", tag.SortId, want)
			}
			return
		}
	}
	t.Fatalf("ListSync 未返回标签 %d", id)
}

// TestTagDomainCreateSortIdIntegration F-06 R1 ⑤（①–④ 的 tag 版，真 MySQL）
func TestTagDomainCreateSortIdIntegration(t *testing.T) {
	cleanTags(t)
	// 本包 TestMain 未初始化雪花节点；Id=0 创建路径需要它
	models.InitSnowflake(1)

	repo := testRepo
	domain := tagService.NewTagDomain(repo, NewTagPreferenceRepo(testDB))
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 基线：显式 255
	if _, _, err := domain.Create(ctx, sortIdDomainTagUser, &valueobjects.CreateTag{
		Name: "基线", Color: "#fff", SortId: 255, SortIdSet: true,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("创建基线: %v", err)
	}

	// ① 显式 7 ⇒ 落库 7（既有缺陷：被 max+1 覆盖）
	explicit, _, err := domain.Create(ctx, sortIdDomainTagUser, &valueobjects.CreateTag{
		Name: "显式7", Color: "#fff", SortId: 7, SortIdSet: true,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("显式 7 创建: %v", err)
	}
	if explicit.SortId != 7 {
		t.Fatalf("显式 7 被改写为 %d", explicit.SortId)
	}
	assertTagSortId(t, repo, explicit.Id, 7)

	// ④ 二次推送同 id、不同显式值 ⇒ 以最后一次为准
	second, _, err := domain.Create(ctx, sortIdDomainTagUser, &valueobjects.CreateTag{
		Name: "显式10", Color: "#fff", Id: explicit.Id,
		SortId: 10, SortIdSet: true,
		CreatedAt: explicit.CreatedAt, UpdatedAt: now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatalf("二次显式 10: %v", err)
	}
	if second.SortId != 10 {
		t.Fatalf("二次显式值应为 10，实际 %d", second.SortId)
	}
	assertTagSortId(t, repo, explicit.Id, 10)

	// ② 未提供 ⇒ max+1
	maxBefore := repo.GetMaxSortId(ctx, sortIdDomainTagUser)
	absent, _, err := domain.Create(ctx, sortIdDomainTagUser, &valueobjects.CreateTag{
		Name: "未提供", Color: "#fff", CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("未提供 sortId 创建: %v", err)
	}
	if absent.SortId != maxBefore+1 {
		t.Fatalf("未提供时应为 max+1=%d，实际 %d", maxBefore+1, absent.SortId)
	}

	// ③ 显式 0（约定 = 未设置）⇒ max+1
	maxBefore = repo.GetMaxSortId(ctx, sortIdDomainTagUser)
	zero, _, err := domain.Create(ctx, sortIdDomainTagUser, &valueobjects.CreateTag{
		Name: "显式0", Color: "#fff", SortId: 0, SortIdSet: true,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("显式 0 创建: %v", err)
	}
	if zero.SortId != maxBefore+1 {
		t.Fatalf("显式 0 应按未设置处理（max+1=%d），实际 %d", maxBefore+1, zero.SortId)
	}
}
