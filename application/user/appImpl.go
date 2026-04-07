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
		App = &userAppImpl{
			userDomain: userDomain,
		}
	})
	return App
}

// UpdateNickname 更新用户昵称
// @param ctx 上下文
// @param req 更新用户昵称请求
// @return error 错误
func (u *userAppImpl) UpdateNickname(
	ctx context.Context,
	req types.UpdateUserNicknameReq,
) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 更新用户昵称
	return u.userDomain.UpdateNickname(ctx, userId, req.Nickname)
}

// GetProfile 获取用户个人信息
// @param ctx 上下文
// @return *types.GetUserProfileRes 用户个人信息
// @return error 错误
func (u *userAppImpl) GetProfile(ctx context.Context) (*types.GetUserProfileRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 获取用户详情
	userEntity, err := u.userDomain.FindById(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 3. 返回结果
	return UserEntity2Res(userEntity), nil
}

// UpdatePassword 更新用户密码
// @param ctx 上下文
// @param req 更新用户密码请求
// @return error 错误
func (u *userAppImpl) UpdatePassword(
	ctx context.Context,
	req types.UpdateUserPasswordReq,
) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 更新用户密码
	return u.userDomain.UpdatePassword(
		ctx, userId,
		req.OldPassword,
		req.NewPassword,
	)
}

// UpdateAvatar 更新用户头像
// @param ctx 上下文
// @param req 更新用户头像请求
// @return *types.UpdateUserAvatarRes 更新用户头像响应
// @return error 错误
func (u *userAppImpl) UpdateAvatar(
	ctx context.Context,
	req types.UpdateUserAvatarReq,
) (*types.UpdateUserAvatarRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 更新用户头像
	err := u.userDomain.UpdateAvatar(ctx, userId, req.AvatarURL)
	if err != nil {
		return nil, err
	}
	// 3. 返回结果
	return &types.UpdateUserAvatarRes{AvatarURL: req.AvatarURL}, nil
}

// UpdateAvatarByFile 更新用户头像（通过文件上传）
// @param ctx 上下文
// @return *types.UpdateUserAvatarRes 更新用户头像响应
// @return error 错误
func (u *userAppImpl) UpdateAvatarByFile(ctx *gin.Context) (*types.UpdateUserAvatarRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
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

// DeactiveUser 禁用用户
// @param ctx 上下文
// @param req 禁用用户请求
// @return error 错误
func (u *userAppImpl) DeactiveUser(ctx context.Context, req *types.DeactiveUserReq) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 查询用户是否存在
	user, err := u.userDomain.FindById(ctx, userId)
	if err != nil {
		return err
	}
	// 3. 密码比对
	isPasswordValid := u.userDomain.PasswordCompare(
		[]byte(req.Password),
		[]byte(user.Password),
	)
	if !isPasswordValid {
		return errors.New("密码错误")
	}
	// 4. 检查用户状态
	if user.IsDeactived() {
		return errors.New("用户已注销")
	}
	// 5. 更新用户状态
	return u.userDomain.Deactive(ctx, userId)
}

// ActiveUser 激活用户
// @param ctx 上下文
// @param req 激活用户请求
// @param ctx 上下文
// @param req 激活用户请求
// @return error 错误
func (u *userAppImpl) ActiveUser(ctx context.Context, req *types.ActiveUserReq) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 查询用户是否存在
	user, err := u.userDomain.FindById(ctx, userId)
	if err != nil {
		return err
	}
	// 3. 密码比对
	isPasswordValid := u.userDomain.PasswordCompare(
		[]byte(req.Password),
		[]byte(user.Password),
	)
	if !isPasswordValid {
		return errors.New("密码错误")
	}
	// 4. 检查用户状态
	if !user.IsDeactived() {
		return errors.New("用户未注销")
	}
	// 5. 更新用户状态
	return u.userDomain.Active(ctx, userId)
}
