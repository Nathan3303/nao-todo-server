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

/*
 * Update Nickname
 */
func (u *userAppImpl) UpdateNickname(
	ctx context.Context,
	req types.UpdateUserNicknameReq,
) (*types.UpdateUserNicknameRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("参数无效 - 用户 ID 不存在")
	}
	// 2. 更新用户昵称
	err := u.userDomain.UpdateNickname(ctx, userId, req.Nickname)
	if err != nil {
		return nil, err
	}
	// 3. 返回结果
	return &types.UpdateUserNicknameRes{}, nil
}

/*
 * Get User Profile
 */
func (u *userAppImpl) GetProfile(ctx context.Context) (*types.GetUserProfileRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("参数无效 - 用户 ID 不存在")
	}
	// 2. 获取用户详情
	userEntity, err := u.userDomain.FindById(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 3. 返回结果
	return &types.GetUserProfileRes{
		Email:      userEntity.Email,
		Nickname:   userEntity.Nickname,
		Avatar:     userEntity.Avatar,
		Role:       userEntity.Role,
		CreateForm: userEntity.CreateForm,
		State:      userEntity.State,
		Config:     userEntity.Config,
		CreatedAt:  userEntity.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  userEntity.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

/*
 * Update User Password
 */
func (u *userAppImpl) UpdatePassword(
	ctx context.Context,
	req types.UpdateUserPasswordReq,
) (*types.UpdateUserPasswordRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("参数无效 - 用户 ID 不存在")
	}
	// 2. 更新用户密码
	err := u.userDomain.UpdatePassword(ctx, userId, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}
	// 3. 返回结果
	return &types.UpdateUserPasswordRes{}, nil
}
