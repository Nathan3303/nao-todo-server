package types

type UpdateUserNicknameReq struct {
	Nickname string `json:"nickname" form:"nickname" binding:"required"`
}

type UpdateUserNicknameRes struct{}
