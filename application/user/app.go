package user

import (
	"context"
	"naotodoserver/domain/user/service"
	"naotodoserver/interfaces/types"
	"sync"

	"github.com/gin-gonic/gin"
)

type UserApp interface {
	UpdateNickname(
		ctx context.Context,
		req types.UpdateUserNicknameReq,
	) (*types.UpdateUserNicknameRes, error)
	GetProfile(ctx context.Context) (*types.GetUserProfileRes, error)
	UpdatePassword(
		ctx context.Context,
		req types.UpdateUserPasswordReq,
	) (*types.UpdateUserPasswordRes, error)
	UpdateAvatar(
		ctx context.Context,
		req types.UpdateUserAvatarReq,
	) (*types.UpdateUserAvatarRes, error)
	UpdateAvatarByFile(ctx *gin.Context) (*types.UpdateUserAvatarRes, error)
}

type userAppImpl struct {
	userDomain service.UserDomain
}

var (
	App  *userAppImpl
	once sync.Once
)
