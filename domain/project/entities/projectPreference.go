package entities

import (
	"naotodoserver/domain/types"
)

// ProjectPreference 任务清单偏好实体
// 用于表示用户在系统中的任务清单偏好记录
// 包含用户 ID、任务清单 ID、视图类型、获取选项、列等属性
type ProjectPreference struct {
	types.EntityBase
	UserId     types.UserID
	ProjectId  types.ProjectID
	ViewType   string
	GetOptions string
	Columns    string
}
