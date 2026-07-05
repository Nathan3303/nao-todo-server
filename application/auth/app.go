package auth

import (
	"context"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/service"
	"naotodoserver/interfaces/types"
)

// AuthApp 认证应用接口
// @description 提供认证相关的应用服务
type AuthApp interface {
	// 处理用户登录
	SignIn(ctx context.Context, signInReq *types.SignInReq) (*types.SignInRes, error)

	// 处理用户注册
	SignUp(ctx context.Context, signUpReq *types.SignUpReq) error

	// 处理用户检查登录状态
	CheckIn(ctx context.Context, checkInReq *types.CheckInReq) (*types.CheckInRes, error)

	// 处理用户登出
	SignOut(ctx context.Context, signOutReq *types.SignOutReq) error

	// 处理用户令牌验证
	Validate(ctx context.Context, token string) (int64, error)

	// 处理用户限流
	RateLimit(ctx context.Context, clientIP string, limit int8) error
}

// authAppImpl 认证应用实现
type authAppImpl struct {
	identityDomain service.IdentityDomain
	userRepo       repositories.User
	sessionRepo    repositories.UserSession
}
