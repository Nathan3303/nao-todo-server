package types

import "encoding/json"

// GetUserProfileRes 获取用户个人信息响应
type GetUserProfileRes struct {
	Email                string `json:"email"`
	Nickname             string `json:"nickname"`
	Avatar               string `json:"avatar"`
	CreatedFrom          string `json:"createdFrom"`
	Role                 string `json:"role"`
	State                uint8  `json:"state"`
	DeactivedAt          string `json:"deactivedAt"`
	LastRestoreAt        string `json:"lastRestoreAt"`
	Config               any    `json:"config"`
	CreatedAt            string `json:"createdAt"`
	UpdatedAt            string `json:"updatedAt"`
}

// UpdateUserNicknameReq 更新用户昵称请求
type UpdateUserNicknameReq struct {
	Nickname string `json:"nickname" form:"nickname" binding:"required"`
}

// UpdateUserPasswordReq 更新用户密码请求
type UpdateUserPasswordReq struct {
	OldPassword string `json:"oldPassword" form:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" binding:"required"`
}

// UpdateUserAvatarReq 更新用户头像请求
type UpdateUserAvatarReq struct {
	AvatarURL string `json:"avatarURL" form:"avatarURL"`
}

// UpdateUserAvatarRes 更新用户头像响应
type UpdateUserAvatarRes struct {
	AvatarURL string `json:"avatarURL"`
}

// DeleteUserReq 删除用户请求
type DeleteUserReq struct {
	Password string `json:"password" form:"password" binding:"required"`
}

// RestoreUserReq 激活用户请求
type RestoreUserReq struct {
	Password string `json:"password" form:"password" binding:"required"`
}

// GetUserConfigRes 获取用户配置响应
type GetUserConfigRes struct {
	Appearance string `json:"appearance"`
	// Preferences 偏好快照（服务端哑存储，原样回传；未设置时为空对象）
	Preferences json.RawMessage `json:"preferences"`
	// UpdatedAt 服务端权威更新时间（LWW 判据）
	UpdatedAt string `json:"updatedAt"`
}

// UpdateUserConfigReq 更新用户配置请求
// Appearance / Preferences 均可选：未携带表示不修改对应字段
// Preferences 为全量快照（客户端推送时装配），仅走 JSON 绑定
type UpdateUserConfigReq struct {
	Appearance  *string         `json:"appearance" form:"appearance"`
	Preferences json.RawMessage `json:"preferences" form:"-"`
}
