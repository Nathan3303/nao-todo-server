package user

import (
	"context"
	"errors"
	"fmt"

	taskApp "naotodoserver/application/task"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/repositories"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// NewUserApp 创建用户应用层实例
func NewUserApp(userRepo repositories.User, taskApp taskApp.TaskApp) UserApp {
	return &userAppImpl{
		userRepo: userRepo,
		taskApp:  taskApp,
	}
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
	if err := u.userRepo.UpdateNickname(ctx, userId, req.Nickname); err != nil {
		return err
	}
	// 3. 同步评论中的用户昵称
	_ = u.taskApp.SyncTaskCommentUserProfile(ctx, userId, req.Nickname, "")
	return nil
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
	userEntity, err := u.userRepo.FindById(ctx, userId)
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
	return u.userRepo.UpdatePassword(
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
	if err := u.userRepo.UpdateAvatar(ctx, userId, req.AvatarURL); err != nil {
		return nil, err
	}
	// 3. 同步评论中的用户头像
	_ = u.taskApp.SyncTaskCommentUserProfile(ctx, userId, "", req.AvatarURL)
	// 4. 返回结果
	return &types.UpdateUserAvatarRes{AvatarURL: req.AvatarURL}, nil
}

// UpdateAvatarByFile 更新用户头像（通过文件上传）
// @param ctx 原始上下文
// @param iCtx 上下文
// @return *types.UpdateUserAvatarRes 更新用户头像响应
// @return error 错误
func (u *userAppImpl) UpdateAvatarByFile(
	ctxRaw *gin.Context,
	ctx context.Context,
) (*types.UpdateUserAvatarRes, error) {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 获取文件
	file, err := ctxRaw.FormFile("avatar")
	if err != nil {
		return nil, errors.New("文件上传失败 - " + err.Error())
	}
	// 3. 验证文件大小和类型
	maxSize := conf.Conf.Uploads.MaxFileSize
	if maxSize <= 0 {
		maxSize = 1024 * 1024 * 2 // 默认 2MB
	}
	if file.Size > maxSize {
		return nil, fmt.Errorf("文件大小不能超过 %dMB", maxSize/1024/1024)
	}
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		return nil, errors.New("不支持的文件类型，仅支持 JPG、JPEG、PNG 格式")
	}
	// 4. 确保上传目录存在
	uploadDir := filepath.Join(conf.Conf.Uploads.UploadDir, conf.Conf.Uploads.AvatarDir)
	if err = os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, errors.New("创建上传目录失败 - " + err.Error())
	}
	// 5. 生成唯一文件名
	uniqueFilename := fmt.Sprintf("%d%s", userId, ext)
	savePath := filepath.Join(uploadDir, uniqueFilename)
	// 6. 保存文件
	if err = ctxRaw.SaveUploadedFile(file, savePath); err != nil {
		return nil, errors.New("文件保存失败 - " + err.Error())
	}
	// 7. 构造可访问的 URL
	staticPath := conf.Conf.Uploads.StaticPath
	if staticPath == "" {
		staticPath = "/static/uploads"
	}
	avatarURL := fmt.Sprintf("%s/%s/%s", staticPath, conf.Conf.Uploads.AvatarDir, uniqueFilename)
	// 8. 更新用户头像
	if err = u.userRepo.UpdateAvatar(ctx, userId, avatarURL); err != nil {
		// 更新失败时删除已上传的文件
		os.Remove(savePath)
		return nil, err
	}
	// 9. 同步评论中的用户头像
	_ = u.taskApp.SyncTaskCommentUserProfile(ctx, userId, "", avatarURL)
	// 10. 返回结果
	return &types.UpdateUserAvatarRes{AvatarURL: avatarURL}, nil
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
	user, err := u.userRepo.FindById(ctx, userId)
	if err != nil {
		return err
	}
	// 3. 密码比对
	isPasswordValid := u.userRepo.PasswordCompare(
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
	return u.userRepo.Deactive(ctx, userId)
}

// GetConfig 获取用户配置
func (u *userAppImpl) GetConfig(ctx context.Context) (*types.GetUserConfigRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	config, err := u.userRepo.GetConfig(ctx, userId)
	if err != nil {
		return nil, err
	}
	return ConfigEntity2Res(config), nil
}

// UpdateConfig 更新用户配置
func (u *userAppImpl) UpdateConfig(ctx context.Context, req types.UpdateUserConfigReq) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	return u.userRepo.UpdateConfig(ctx, userId, req.Appearance)
}

// DeleteDeactivatedUsers 删除已注销用户（供定时任务调用）
func (u *userAppImpl) DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) error {
	_, err := u.userRepo.DeleteDeactivatedUsers(ctx, dayOffset)
	return err
}

// ActiveUser 激活用户
func (u *userAppImpl) ActiveUser(ctx context.Context, req *types.ActiveUserReq) error {
	// 1. 获取 User ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 查询用户是否存在
	user, err := u.userRepo.FindById(ctx, userId)
	if err != nil {
		return err
	}
	// 3. 密码比对
	isPasswordValid := u.userRepo.PasswordCompare(
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
	return u.userRepo.Active(ctx, userId)
}
