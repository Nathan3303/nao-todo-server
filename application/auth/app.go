package auth

import (
	"context"
	"naotodoserver/application/auth/dto"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/service"
	domaintypes "naotodoserver/domain/types"
)

// AuthApp 认证应用接口
// @description 提供认证相关的应用服务
type AuthApp interface {
	// 处理用户登录
	SignIn(ctx context.Context, signInInput *dto.SignInInput) (*dto.SignInOutput, error)

	// 处理用户注册
	SignUp(ctx context.Context, signUpInput *dto.SignUpInput) error

	// 处理用户检查登录状态
	CheckIn(ctx context.Context, checkInInput *dto.CheckInInput) (*dto.CheckInOutput, error)

	// 处理用户登出
	SignOut(ctx context.Context, signOutInput *dto.SignOutInput) error

	// 处理用户令牌验证
	Validate(ctx context.Context, token string) (domaintypes.UserID, error)

	// 获取用户现存会话列表
	ListSessions(ctx context.Context, userId int64, currentToken string) ([]*dto.SessionItem, error)

	// 下线指定会话
	LogoutSession(ctx context.Context, userId int64, sessionId int64) error

	// 退出其他全部设备（除当前 token 外）
	LogoutOtherSessions(ctx context.Context, userId int64, currentToken string) error

	// 处理用户限流
	RateLimit(ctx context.Context, clientIP string, limit int64) error
}

// authAppImpl 认证应用实现
type authAppImpl struct {
	identityDomain service.IdentityDomain
	userRepo       repositories.User
	sessionRepo    repositories.UserSession
}
