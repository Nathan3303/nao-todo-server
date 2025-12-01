package user

import (
	"context"
	"naotodoserver/domain/user/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type UserApp interface {
	UpdateNickname(
		ctx context.Context,
		req types.UpdateUserNicknameReq,
	) (*types.UpdateUserNicknameRes, error)
}

type userAppImpl struct {
	userDomain service.UserDomain
}

var (
	App  *userAppImpl
	once sync.Once
)
