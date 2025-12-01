package auth

import (
	"context"
	"naotodoserver/domain/auth/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type AuthApp interface {
	SignIn(ctx context.Context, signInReq *types.SignInReq) (*types.SignInRes, error)
	SignUp(ctx context.Context, signUpReq *types.SignUpReq) (*types.SignUpRes, error)
	CheckIn(ctx context.Context, checkInReq *types.CheckInReq) (*types.CheckInRes, error)
	SignOut(ctx context.Context, signOutReq *types.SignOutReq) (*types.SignOutRes, error)
	Validate(ctx context.Context, token string) (int64, error)
}

type authAppImpl struct {
	authDomain service.AuthDomain
}

var (
	App  *authAppImpl
	once sync.Once
)
