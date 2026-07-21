package types

// SignInReq 登录请求结构体
type SignInReq struct {
	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// SignInRes 登录响应结构体
type SignInRes struct {
	Token            string `json:"jwt"`
	PendingDeletion  bool   `json:"pendingDeletion"`
	DeletionDeadline string `json:"deletionDeadline,omitempty"`
}

// SignUpReq 注册请求结构体
type SignUpReq struct {
	Email    string `json:"email" form:"email" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
	Nickname string `json:"nickname" form:"nickname"`
}

// SignOutReq 登出请求结构体
type SignOutReq struct {
	Token      string `json:"jwt" form:"jwt" binding:"required"`
	DeviceType string `json:"deviceType" form:"deviceType"`
}

// SignOutRes 登出响应结构体
type SignOutRes struct{}

// CheckInReq 检入请求结构体
type CheckInReq struct {
	Token      string `json:"jwt" form:"jwt" binding:"required"`
	DeviceType string `json:"deviceType" form:"deviceType"`
}

// CheckInRes 检入响应结构体
type CheckInRes struct {
	Token string `json:"jwt"`
}
