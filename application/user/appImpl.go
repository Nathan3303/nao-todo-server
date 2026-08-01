package user

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	taskApp "naotodoserver/application/task"
	"naotodoserver/application/user/dto"
	"naotodoserver/conf"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/repositories"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
)

// NewUserApp 创建用户应用层实例
func NewUserApp(
	userRepo repositories.User,
	sessionRepo repositories.UserSession,
	taskApp taskApp.TaskCommentApp,
	avatarStorage AvatarStorage,
) UserApp {
	return &userAppImpl{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		taskApp:       taskApp,
		avatarStorage: avatarStorage,
	}
}

// UpdateNickname 更新用户昵称
// @param ctx 上下文
// @param req 更新用户昵称请求
// @return error 错误
func (u *userAppImpl) UpdateNickname(
	ctx context.Context,
	req dto.UpdateNicknameInput,
) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	// 2. 更新用户昵称
	err := u.userRepo.UpdateNickname(ctx, domaintypes.UserID(userId), req.Nickname)
	if err != nil {
		return err
	}
	// 3. 同步评论中的用户昵称
	_ = u.taskApp.SyncTaskCommentUserProfile(ctx, userId, req.Nickname, "")
	return nil
}

// GetProfile 获取用户个人信息
// @param ctx 上下文
// @return *dto.GetProfileOutput 用户个人信息
// @return error 错误
func (u *userAppImpl) GetProfile(ctx context.Context) (*dto.GetProfileOutput, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 2. 获取用户详情
	userEntity, err := u.userRepo.FindById(ctx, domaintypes.UserID(userId))
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
	req dto.UpdatePasswordInput,
) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	// 2. 更新用户密码
	err := u.userRepo.UpdatePassword(
		ctx,
		domaintypes.UserID(userId),
		req.OldPassword, req.NewPassword,
	)
	if err != nil {
		return fmt.Errorf("user.UpdatePassword: %w", err)
	}
	return nil
}

// UpdateAvatar 更新用户头像
// @param ctx 上下文
// @param req 更新用户头像请求
// @return *dto.UpdateAvatarOutput 更新用户头像响应
// @return error 错误
func (u *userAppImpl) UpdateAvatar(
	ctx context.Context,
	req dto.UpdateAvatarInput,
) (*dto.UpdateAvatarOutput, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 2. 更新用户头像
	err := u.userRepo.UpdateAvatar(ctx, domaintypes.UserID(userId), req.AvatarURL)
	if err != nil {
		return nil, err
	}
	// 3. 同步评论中的用户头像
	_ = u.taskApp.SyncTaskCommentUserProfile(ctx, userId, "", req.AvatarURL)
	// 4. 返回结果
	return &dto.UpdateAvatarOutput{
		AvatarURL: conf.Conf.Uploads.AvatarURL(req.AvatarURL),
	}, nil
}

// UpdateAvatarByFile 更新用户头像（通过文件上传）
// @param ctx 上下文
// @param file 文件内容
// @param filename 文件名
// @param size 文件大小（字节）
// @return *dto.UpdateAvatarOutput 更新用户头像响应
// @return error 错误
func (u *userAppImpl) UpdateAvatarByFile(
	ctx context.Context,
	file io.Reader,
	filename string,
	size int64,
) (*dto.UpdateAvatarOutput, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 2. 验证文件大小和类型
	maxSize := conf.Conf.Uploads.MaxFileSize
	if maxSize <= 0 {
		maxSize = 1024 * 1024 * 2 // 默认 2MB
	}
	if size > maxSize {
		return nil, fmt.Errorf("文件大小不能超过 %dMB", maxSize/1024/1024)
	}
	ext := filepath.Ext(filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		return nil, errors.New("不支持的文件类型，仅支持 JPG、JPEG、PNG 格式")
	}
	// 3. 生成唯一文件名并保存文件
	uniqueFilename := fmt.Sprintf("%d%s", userId, ext)
	avatarURL, err := u.avatarStorage.Save(ctx, uniqueFilename, file)
	if err != nil {
		return nil, err
	}
	// 4. 更新用户头像
	err = u.userRepo.UpdateAvatar(ctx, domaintypes.UserID(userId), avatarURL)
	if err != nil {
		// 更新失败时删除已上传的文件
		_ = u.avatarStorage.Delete(ctx, uniqueFilename)
		return nil, err
	}
	// 5. 同步评论中的用户头像
	_ = u.taskApp.SyncTaskCommentUserProfile(ctx, userId, "", avatarURL)
	// 6. 返回结果
	return &dto.UpdateAvatarOutput{
		AvatarURL: conf.Conf.Uploads.AvatarURL(avatarURL),
	}, nil
}

// DeleteUser 删除用户（注销用户）
// @param ctx 上下文
// @param req 删除用户请求
// @return error 错误
func (u *userAppImpl) DeleteUser(ctx context.Context, req dto.DeleteUserInput) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	// 2. 查询用户是否存在
	user, err := u.userRepo.FindById(ctx, domaintypes.UserID(userId))
	if err != nil {
		return err
	}
	// 3. 密码比对
	isPasswordValid := u.userRepo.PasswordCompare(
		[]byte(req.Password),
		[]byte(user.Password),
	)
	if !isPasswordValid {
		return domerr.ErrPasswordMismatch
	}
	// 4. 检查用户状态
	if user.IsDeactived() {
		return domerr.ErrUserDeactivated
	}
	// 5. 检查冷却期
	if user.IsInCooldown() {
		return domerr.ErrUserInCooldown
	}
	// 6. 更新用户状态
	if err := u.userRepo.Deactive(ctx, domaintypes.UserID(userId)); err != nil {
		return fmt.Errorf("user.Deactive: %w", err)
	}
	return nil
}

// DeleteDeactivatedUsers 删除已注销用户（供定时任务调用）
// @param ctx 上下文
// @param dayOffset 注销时间偏移天数
// @return error 错误
func (u *userAppImpl) DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) error {
	_, err := u.userRepo.DeleteDeactivatedUsers(ctx, dayOffset)
	if err != nil {
		return fmt.Errorf("user.DeleteDeactivatedUsers: %w", err)
	}
	return nil
}

// RestoreUser 激活用户（取消注销）
// @param ctx 上下文
// @param req 激活用户请求
// @return error 错误
func (u *userAppImpl) RestoreUser(ctx context.Context, req dto.RestoreUserInput) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	// 2. 查询用户是否存在
	user, err := u.userRepo.FindById(ctx, domaintypes.UserID(userId))
	if err != nil {
		return err
	}
	// 3. 密码比对
	isPasswordValid := u.userRepo.PasswordCompare(
		[]byte(req.Password),
		[]byte(user.Password),
	)
	if !isPasswordValid {
		return domerr.ErrPasswordMismatch
	}
	// 4. 检查用户状态
	if !user.IsDeactived() {
		return errors.New("用户未处于注销状态")
	}
	// 5. 更新用户状态
	if err := u.userRepo.Active(ctx, domaintypes.UserID(userId)); err != nil {
		return fmt.Errorf("user.Active: %w", err)
	}
	return nil
}

// GetConfig 获取用户配置
// @param ctx 上下文
// @return *dto.GetConfigOutput 获取用户配置响应
// @return error 错误
func (u *userAppImpl) GetConfig(ctx context.Context) (*dto.GetConfigOutput, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	config, err := u.userRepo.GetConfig(ctx, domaintypes.UserID(userId))
	if err != nil {
		return nil, fmt.Errorf("user.GetConfig: %w", err)
	}
	return ConfigEntity2Res(config), nil
}

// UpdateConfig 更新用户配置
// @param ctx 上下文
// @param req 更新用户配置请求
// @return error 错误
func (u *userAppImpl) UpdateConfig(ctx context.Context, req dto.UpdateConfigInput) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	err := u.userRepo.UpdateConfig(
		ctx,
		domaintypes.UserID(userId),
		req.Appearance,
	)
	if err != nil {
		return fmt.Errorf("user.UpdateConfig: %w", err)
	}
	return nil
}
