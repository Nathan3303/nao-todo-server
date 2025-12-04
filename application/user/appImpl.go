package user

import (
	"context"
	"errors"
	"fmt"
	"naotodoserver/domain/user/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
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

/*
 * Update User Avatar
 */
func (u *userAppImpl) UpdateAvatar(
	ctx context.Context,
	req types.UpdateUserAvatarReq,
) (*types.UpdateUserAvatarRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("参数无效 - 用户 ID 不存在")
	}
	// 2. 更新用户头像
	err := u.userDomain.UpdateAvatar(ctx, userId, req.AvatarURL)
	if err != nil {
		return nil, err
	}
	// 3. 返回结果
	return &types.UpdateUserAvatarRes{AvatarURL: req.AvatarURL}, nil
}

/*
 * Update User Avatar by File
 */
func (u *userAppImpl) UpdateAvatarByFile(ctx *gin.Context) (*types.UpdateUserAvatarRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("参数无效 - 用户 ID 不存在")
	}
	// 2. 获取文件
	file, err := ctx.FormFile("avatar")
	if err != nil {
		return nil, errors.New("文件上传失败 - " + err.Error())
	}
	// 3. 验证文件大小和类型
	const maxSize = 1024 * 1024 * 2 // 2MB
	if file.Size > maxSize {
		return nil, errors.New("文件大小不能超过 2MB")
	}
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		return nil, errors.New("不支持的文件类型")
	}
	// 4. 生成唯一文件名
	uniqueFilename := fmt.Sprintf("/uploads/avatars/%d%s", userId, ext)
	// 5. 确保目录存在
	os.MkdirAll("/uploads/avatars", os.ModePerm)
	// 6. 保存文件
	if err = ctx.SaveUploadedFile(file, uniqueFilename); err != nil {
		return nil, err
	}
	// 7. 更新用户头像
	err = u.userDomain.UpdateAvatar(ctx, userId, uniqueFilename)
	if err != nil {
		return nil, err
	}
	// 8. 返回结果
	return &types.UpdateUserAvatarRes{AvatarURL: uniqueFilename}, nil
}
