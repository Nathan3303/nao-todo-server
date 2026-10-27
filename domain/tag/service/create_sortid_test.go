package service

import (
	"context"
	"testing"

	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

// fakeSortIdTagRepo 只实现 sortId 相关方法；Upsert 返回 Created=false（避免偏好初始化）
type fakeSortIdTagRepo struct {
	repositories.TagRepository
	maxSortId uint16
	lastVO    *valueobjects.CreateTag
}

func (f *fakeSortIdTagRepo) GetMaxSortId(_ context.Context, _ int64) uint16 {
	return f.maxSortId
}

func (f *fakeSortIdTagRepo) Upsert(
	_ context.Context, _ int64, vo *valueobjects.CreateTag,
) (*entities.Tag, domaintypes.UpsertResult, error) {
	f.lastVO = vo
	return &entities.Tag{SortId: vo.SortId},
		domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite}, nil
}

// TestTagDomainCreateSortIdTriState F-06 R1：显式 >0 原样落库；未提供 / 显式 0 ⇒ max+1
func TestTagDomainCreateSortIdTriState(t *testing.T) {
	const maxSortId uint16 = 255
	cases := []struct {
		name   string
		set    bool
		sortId uint16
		want   uint16
	}{
		{"显式值 7", true, 7, 7},
		{"未提供", false, 0, maxSortId + 1},
		{"显式 0 = 未设置", true, 0, maxSortId + 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeSortIdTagRepo{maxSortId: maxSortId}
			domain := NewTagDomain(repo, nil)

			entity, _, err := domain.Create(context.Background(), 1, &valueobjects.CreateTag{
				Name: "标签", Color: "#fff", SortId: tc.sortId, SortIdSet: tc.set,
			})
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if repo.lastVO.SortId != tc.want {
				t.Fatalf("传入仓储的 SortId = %d, want %d", repo.lastVO.SortId, tc.want)
			}
			if entity.SortId != tc.want {
				t.Fatalf("返回实体 SortId = %d, want %d", entity.SortId, tc.want)
			}
		})
	}
}
