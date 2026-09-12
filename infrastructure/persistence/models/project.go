package models

import (
	"database/sql"
)

// Project 项目模型
type Project struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_project_user_id"`

	// 项目名称
	Name string `gorm:"not null;size:128"`

	// 项目描述
	Description string `gorm:"null;size:512"`

	// 归档时间
	// 项目归档后，项目下的所有待办任务也应该被归档
	ArchivedAt sql.NullTime `gorm:"null;index:idx_project_archived_at"`

	// 项目停用时间
	// 项目停用后，项目下的所有待办任务也应该被停用
	DeactivedAt sql.NullTime `gorm:"null;index:idx_project_deactived_at"`

	// 项目排序 ID
	// 用于在项目列表中排序
	// 通常采用大间距的整数，例如 100、200 等，更新排序时取被排序项的 SortId +- 1/2
	// 超出类型范围，则需要重新计算所有项目排序的 SortId
	SortId uint16 `gorm:"default:0"`

	// 项目任务数量（含子任务/含归档/含放弃；不含已删除；服务端 owned 反规范化列，事件联动维护）
	TaskCount uint `gorm:"default:0"`

	// 项目偏好设置
	// 用于存储用户对项目的偏好设置，例如视图类型、获取选项、列等
	Preference *ProjectPreference `gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE;"`
}

// ProjectPreference 项目偏好设置模型
// 用于存储用户对项目的偏好设置，例如视图类型、获取选项、列等
type ProjectPreference struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_proj_pref_user"`

	// 项目 ID
	// 用于关联项目表，获取项目偏好设置
	ProjectId int64 `gorm:"not null;index:idx_proj_pref_project"`

	// 视图类型
	// 用于指定项目视图的类型，例如列表、卡片等
	ViewType string `gorm:"not null;size:16"`

	// 获取选项
	// 用于指定项目视图的获取选项，例如按创建时间、按完成时间等
	GetOptions string `gorm:"not null;size:512"`

	// 列选项
	// 用于指定项目视图的列选项，例如显示项目名称、显示项目描述等
	Columns string `gorm:"not null;size:512"`
}
