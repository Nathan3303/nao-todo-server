// Package dto 定义认证应用层的入参与出参结构体
// @description 应用层自有 DTO，不含任何 HTTP 关注点（如 json / form tag）
package dto

// SignInInput 登录入参
type SignInInput struct {
	Email    string
	Password string
}

// SignInOutput 登录出参
type SignInOutput struct {
	Token           string
	PendingDeletion bool
	DeletedAt       string
}

// SignUpInput 注册入参
type SignUpInput struct {
	Email    string
	Password string
	Nickname string
}

// SignOutInput 登出入参
type SignOutInput struct {
	Token      string
	DeviceType string
}

// CheckInInput 检入入参
type CheckInInput struct {
	Token      string
	DeviceType string
}

// CheckInOutput 检入出参
type CheckInOutput struct {
	Token           string
	PendingDeletion bool
	DeletedAt       string
}

// SessionItem 会话列表项
type SessionItem struct {
	Id         string
	DeviceId   string
	DeviceType string
	IP4        string
	Region     string
	CreatedAt  string
	UpdatedAt  string
	Current    bool
}
