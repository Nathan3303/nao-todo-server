package types

type GetUserProfileReq struct{}

type GetUserProfileRes struct {
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Role        string `json:"role"`
	CreatedFrom string `json:"createdFrom"`
	State       int8   `json:"state"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Config      any    `json:"config"`
}

type UpdateUserNicknameReq struct {
	Nickname string `json:"nickname" form:"nickname" binding:"required"`
}

type UpdateUserNicknameRes struct{}

type UpdateUserPasswordReq struct {
	OldPassword string `json:"oldPassword" form:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" form:"newPassword" binding:"required"`
}

type UpdateUserPasswordRes struct{}

type UpdateUserAvatarReq struct {
	AvatarURL string `json:"avatarURL" form:"avatarURL"`
}

type UpdateUserAvatarRes struct {
	AvatarURL string `json:"avatarURL"`
}
