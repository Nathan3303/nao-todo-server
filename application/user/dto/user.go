// Package dto 定义用户应用层的入参与出参结构体
// @description 应用层自有 DTO，不含任何 HTTP 关注点（如 json / form tag）
package dto

// UpdateNicknameInput 更新用户昵称入参
type UpdateNicknameInput struct {
	Nickname string
}

// GetProfileOutput 获取用户个人信息出参
type GetProfileOutput struct {
	Email         string
	Nickname      string
	Avatar        string
	CreatedFrom   string
	Role          string
	State         uint8
	DeactivedAt   string
	LastRestoreAt string
	Config        any
	CreatedAt     string
	UpdatedAt     string
}

// UpdatePasswordInput 更新用户密码入参
type UpdatePasswordInput struct {
	OldPassword string
	NewPassword string
}

// UpdateAvatarInput 更新用户头像入参
type UpdateAvatarInput struct {
	AvatarURL string
}

// UpdateAvatarOutput 更新用户头像出参
type UpdateAvatarOutput struct {
	AvatarURL string
}

// DeleteUserInput 删除用户（注销账户）入参
type DeleteUserInput struct {
	Password string
	Token    string // 当前请求的会话令牌（注销成功后删除当次 Token）
}

// RestoreUserInput 激活用户入参
type RestoreUserInput struct {
	Password string
}

// GetConfigOutput 获取用户配置出参
type GetConfigOutput struct {
	Appearance string
}

// UpdateConfigInput 更新用户配置入参
type UpdateConfigInput struct {
	Appearance string
}
