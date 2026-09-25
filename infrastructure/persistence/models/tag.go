package models

// Tag 标签
type Tag struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_tag_user_id"`

	// 标签名称
	Name string `gorm:"not null;size:64"`

	// 标签描述
	Description string `gorm:"null;size:512"`

	// 标签颜色
	Color string `gorm:"not null;size:16"`

	// 排序 ID
	// 用于在标签列表中显示标签的顺序
	// 通常采用大间距的整数，例如 100、200 等，更新排序时取被排序项的 SortId +- 1/2
	// 超出类型范围，则需要重新计算所有标签的 SortId
	SortId uint16

	// 标签偏好设置
	Preference *TagPreference `gorm:"foreignKey:TagId;constraint:OnDelete:CASCADE;"`
}

// TagPreference 标签偏好设置
type TagPreference struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_tag_pref_user"`

	// 标签 ID
	TagId int64 `gorm:"not null;index:idx_tag_pref_tag"`

	// 视图类型
	// 用于指定标签视图的类型，例如列表、卡片等
	ViewType string `gorm:"not null;size:16"`

	// 获取选项
	// 用于指定标签视图的获取选项，例如按创建时间、按完成时间等
	GetOptions string `gorm:"not null;size:512"`

	// 列选项
	// 用于指定标签视图的列选项，例如显示标签名称、显示标签描述等
	Columns string `gorm:"not null;size:512"`
}
