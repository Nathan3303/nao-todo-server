package types

type GetUserProfileRes struct {
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	CreatedFrom string `json:"createdFrom"`
	Role        string `json:"role"`
	State       uint8  `json:"state"`
	Config      any    `json:"config"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type UpdateUserNicknameReq struct {
	Nickname string `json:"nickname" form:"nickname" binding:"required"`
}

type UpdateUserPasswordReq struct {
	OldPassword string `json:"oldPassword" form:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" binding:"required"`
}

type UpdateUserAvatarReq struct {
	AvatarURL string `json:"avatarURL" form:"avatarURL"`
}

type UpdateUserAvatarRes struct {
	AvatarURL string `json:"avatarURL"`
}

type DeactiveUserReq struct {
	Password string `json:"password" form:"password" binding:"required"`
}

type ActiveUserReq struct {
	Password string `json:"password" form:"password" binding:"required"`
}

type GetUserConfigRes struct {
	Appearance string `json:"appearance"`
}

type UpdateUserConfigReq struct {
	Appearance string `json:"appearance" form:"appearance" binding:"required"`
}
