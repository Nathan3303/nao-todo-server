package user

import (
	"context"
	"errors"
	"naotodoserver/domain/user/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
)

// 注册 Domain
func RegistDomainImpl(userDomain service.UserDomain) UserApp {
	once.Do(func() {
		App = &userAppImpl{userDomain: userDomain}
	})
	return App
}

func (u *userAppImpl) UpdateNickname(
	ctx context.Context,
	req types.UpdateUserNicknameReq,
) (*types.UpdateUserNicknameRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("参数无效")
	}
	// 2. 更新用户昵称
	err := u.userDomain.UpdateNickname(ctx, userId, req.Nickname)
	if err != nil {
		return nil, err
	}
	// 3. 返回结果
	return &types.UpdateUserNicknameRes{}, nil
}
