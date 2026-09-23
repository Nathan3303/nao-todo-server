package valueobjects

// UpdateUserConfig 更新用户配置值对象
// Appearance / Preferences 为 nil 表示不修改对应字段（部分更新语义）
// Preferences 为空串表示清除偏好快照（写 NULL）
type UpdateUserConfig struct {
	Appearance  *string
	Preferences *string
}

// NewUpdateUserConfig 创建更新用户配置值对象
func NewUpdateUserConfig(appearance, preferences *string) UpdateUserConfig {
	return UpdateUserConfig{
		Appearance:  appearance,
		Preferences: preferences,
	}
}
