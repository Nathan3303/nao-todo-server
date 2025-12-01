package types

type UserSignInReq struct {
	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type UserSignInRes struct {
	Token string `json:"jwt"`
}

type UserSignUpReq struct {
	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
	Nickname string `json:"nickname" form:"nickname"`
}

type UserSignUpRes struct{}

type UserSignOutReq struct {
	Token      string `json:"jwt" form:"jwt" binding:"required"`
	DeviceType string `json:"deviceType" form:"deviceType"`
}

type UserSignOutRes struct{}

type UserCheckInReq struct {
	Token      string `json:"jwt" form:"jwt" binding:"required"`
	DeviceType string `json:"deviceType" form:"deviceType"`
}

type UserCheckInRes struct {
	Token string `json:"jwt"`
}

type UpdateProfileReq struct {
	Nickname string `json:"nickname" form:"nickname"`
}

type UpdateProfileRes struct{}
