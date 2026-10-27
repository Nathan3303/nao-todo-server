package service

import (
	"context"
	"testing"

	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

// fakeSortIdProjectRepo 只实现 sortId 相关方法；Upsert 返回 Created=false（避免偏好初始化）
type fakeSortIdProjectRepo struct {
	repositories.Project
	maxSortId uint16
	lastVO    *valueobjects.CreateProject
}

func (f *fakeSortIdProjectRepo) GetMaxSortId(_ context.Context, _ int64) uint16 {
	return f.maxSortId
}

func (f *fakeSortIdProjectRepo) Upsert(
	_ context.Context, _ int64, vo *valueobjects.CreateProject,
) (*entities.Project, domaintypes.UpsertResult, error) {
	f.lastVO = vo
	return &entities.Project{SortId: vo.SortId},
		domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite}, nil
}

// TestProjectDomainCreateSortIdTriState F-06 R1：显式 >0 原样落库；未提供 / 显式 0 ⇒ max+1
func TestProjectDomainCreateSortIdTriState(t *testing.T) {
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
			repo := &fakeSortIdProjectRepo{maxSortId: maxSortId}
			domain := NewProjectDomain(repo, nil)

			entity, _, err := domain.Create(context.Background(), &valueobjects.CreateProject{
				UserId: 1, Name: "清单", SortId: tc.sortId, SortIdSet: tc.set,
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
