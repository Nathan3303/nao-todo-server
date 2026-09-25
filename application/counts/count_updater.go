// Package counts 领域统计属性联动的计数订阅者
// 订阅 CountEvent 并转换为仓储计数方法调用；所有仓储方法经 DBFrom 加入发布者外层事务
// （同事务强一致，ADR §4.2）。跨仓访问先例：projectApp 注入 taskRepo 用于级联操作。
package counts

import (
	"context"

	projectRepositories "naotodoserver/domain/project/repositories"
	taskRepositories "naotodoserver/domain/task/repositories"
	"naotodoserver/domain/types"
)

// CountUpdater 计数事件订阅者：事件 → 仓储计数方法
type CountUpdater struct {
	taskRepo    taskRepositories.Task
	projectRepo projectRepositories.Project
}

// NewCountUpdater 创建计数订阅者
func NewCountUpdater(
	taskRepo taskRepositories.Task,
	projectRepo projectRepositories.Project,
) *CountUpdater {
	return &CountUpdater{taskRepo: taskRepo, projectRepo: projectRepo}
}

// HandleCountEvent 处理计数事件（订阅者入口；错误 ⇒ 外层事务回滚）
// 事件映射见 domain/types/count_events.go 各常量注释。
func (u *CountUpdater) HandleCountEvent(ctx context.Context, event types.CountEvent) error {
	switch event.Type {
	case types.CountEventCheckItemChanged:
		return u.taskRepo.AdjustCheckItemCount(ctx, event.UserId, event.TaskId, int(event.Delta))
	case types.CountEventCommentChanged:
		return u.taskRepo.AdjustCommentCount(ctx, event.UserId, event.TaskId, int(event.Delta))
	case types.CountEventSubTaskChanged:
		return u.taskRepo.AdjustSubTaskCount(ctx, event.UserId, event.TaskId, int(event.Delta))
	case types.CountEventTaskCountChanged:
		return u.projectRepo.AdjustTaskCount(ctx, event.UserId, event.ProjectId, int(event.Delta))
	case types.CountEventTaskMoved:
		// E5：旧项目 -1、新项目 +1（两行均 bump updated_at，B2）
		if err := u.projectRepo.AdjustTaskCount(ctx, event.UserId, event.OldProjectId, -1); err != nil {
			return err
		}
		return u.projectRepo.AdjustTaskCount(ctx, event.UserId, event.ProjectId, 1)
	case types.CountEventTaskParentChanged:
		// E6：旧父 -1、新父 +1（A→0 仅 -1；0 无父跳过）
		if event.OldParentTaskId > 0 {
			if err := u.taskRepo.AdjustSubTaskCount(ctx, event.UserId, event.OldParentTaskId, -1); err != nil {
				return err
			}
		}
		if event.ParentTaskId > 0 {
			return u.taskRepo.AdjustSubTaskCount(ctx, event.UserId, event.ParentTaskId, 1)
		}
		return nil
	case types.CountEventTaskCountRecounted:
		// E7：批量重算写最终值（级联删/恢复场景，不逐事件）
		return u.projectRepo.RecountTaskCount(ctx, event.UserId, event.ProjectId)
	}
	return nil
}
