package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/domain/types"
)

// fakeCopyRepo 仅实现 Copy 路径会用到的方法；其余方法由内嵌 nil 接口占位
// （若意外被调用会 panic，便于暴露越界依赖）。
type fakeCopyRepo struct {
	repositories.Task
	source        *entities.Task
	upsertCalled  bool
	maxSortCalled bool
	getByIDCalled bool
}

func (f *fakeCopyRepo) GetById(
	ctx context.Context,
	userId int64,
	taskId int64,
	includeDeleted bool,
) (*entities.Task, error) {
	f.getByIDCalled = true
	return f.source, nil
}

func (f *fakeCopyRepo) GetMaxSortId(ctx context.Context, userId, parentTaskId int64) (uint16, error) {
	f.maxSortCalled = true
	return 255, nil
}

func (f *fakeCopyRepo) Upsert(
	ctx context.Context,
	userId int64,
	createTaskValueObject *valueobjects.CreateTask,
) (*entities.Task, bool, error) {
	f.upsertCalled = true
	return nil, false, errors.New("校验失败时不应触发写入")
}

// TestCopy_ValidationErrorNotSwallowed 回归：Copy 的 VO 校验失败必须返回真实错误且不写入
// （原缺陷：`if vo.Validate() != nil { return nil, err }` 中 err 为 nil ⇒ 返回 (nil, nil)，
// 应用层随即解引用 nil 实体 panic）
func TestCopy_ValidationErrorNotSwallowed(t *testing.T) {
	// 源任务名称已达 256 上限 ⇒ 复制品名称（+3 rune“的复制”）必然超限触发校验失败
	source := &entities.Task{
		EntityBase: types.EntityBase{Id: 4242},
		Name:       strings.Repeat("任", 256),
	}
	repo := &fakeCopyRepo{source: source}
	domain := NewTaskDomain(repo, nil)

	entity, err := domain.Copy(context.Background(), 1, 4242)

	if err == nil {
		t.Fatal("Copy 校验失败必须返回非 nil error（不得吞错）")
	}
	if !strings.Contains(err.Error(), "任务名称最多256个字符") {
		t.Fatalf("Copy 返回了非预期错误：%v", err)
	}
	if entity != nil {
		t.Fatalf("校验失败时不应返回实体：%+v", entity)
	}
	if !repo.getByIDCalled {
		t.Fatal("未读取源任务（前置路径未执行）")
	}
	if repo.upsertCalled {
		t.Fatal("校验失败时不得写入任何数据（Upsert 被调用）")
	}
	if repo.maxSortCalled {
		t.Fatal("校验失败时应早于排序值查询返回（GetMaxSortId 被调用）")
	}
}
