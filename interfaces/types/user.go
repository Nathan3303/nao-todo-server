package types

type GetUserProfileReq struct{}

type GetUserProfileRes struct {
	Email      string `json:"email"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Role       string `json:"role"`
	CreateForm any    `json:"createForm"`
	State      string `json:"state"`
	Config     any    `json:"config"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
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
