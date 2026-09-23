package entities

import "naotodoserver/domain/types"

// UserConfig 用户配置实体
// 用于表示用户在系统中的配置信息
// 包含用户 ID、外观等属性
type UserConfig struct {
	types.EntityBase
	UserId     types.UserID
	Appearance string
	// Preferences 偏好快照（客户端拥有的 JSON blob，服务端不解析）
	// 空串表示尚未设置
	Preferences string
}
