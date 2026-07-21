package types

// GetUserProfileRes 获取用户个人信息响应
type GetUserProfileRes struct {
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	CreatedFrom string `json:"createdFrom"`
	Role        string `json:"role"`
	State       uint8  `json:"state"`
	DeactivedAt string `json:"deactivedAt"`
	Config      any    `json:"config"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
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
}

// UpdateUserConfigReq 更新用户配置请求
type UpdateUserConfigReq struct {
	Appearance string `json:"appearance" form:"appearance" binding:"required"`
}
