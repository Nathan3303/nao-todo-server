package types

type SignInReq struct {
	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type SignInRes struct {
	Token string `json:"jwt"`
}

type SignUpReq struct {
	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
	Nickname string `json:"nickname" form:"nickname"`
}

type SignOutReq struct {
	Token      string `json:"jwt" form:"jwt" binding:"required"`
	DeviceType string `json:"deviceType" form:"deviceType"`
}

type SignOutRes struct{}

type CheckInReq struct {
	Token      string `json:"jwt" form:"jwt" binding:"required"`
	DeviceType string `json:"deviceType" form:"deviceType"`
}

type CheckInRes struct {
	Token string `json:"jwt"`
}
