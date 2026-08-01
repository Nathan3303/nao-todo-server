package entities

import (
	"time"

	"naotodoserver/domain/types"
)

// Project 任务清单实体
// 用于表示用户在系统中的任务清单记录
// 包含用户 ID、名称、描述、归档时间、创建时间、更新时间、删除时间、停用时间、排序 ID等属性
type Project struct {
	types.EntityBase
	UserId      types.UserID
	Name        string
	Description string
	ArchivedAt  types.NullableTime
	DeactivedAt types.NullableTime
	SortId      uint16
}

// Archive 归档任务清单（幂等重设归档时间）
func (project *Project) Archive() {
	project.ArchivedAt = types.NewNullableTimeByTime(time.Now())
}

// Unarchive 取消归档任务清单
func (project *Project) Unarchive() {
	project.ArchivedAt = types.NewNullableTimeNull()
}

// Delete 删除任务清单（软删除，设置停用时间）
func (project *Project) Delete() {
	project.DeactivedAt = types.NewNullableTimeByTime(time.Now())
}

// Restore 恢复任务清单（清空停用时间）
func (project *Project) Restore() {
	project.DeactivedAt = types.NewNullableTimeNull()
}
