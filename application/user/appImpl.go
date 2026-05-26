package user

import (
	"context"
	"errors"
	"fmt"
	"naotodoserver/conf"
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
	err = u.userDomain.UpdateAvatar(ctx, userId, avatarURL)
	if err != nil {
		// 更新失败时删除已上传的文件
		os.Remove(savePath)
		return nil, err
	}
	// 9. 返回结果
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

// GetConfig 获取用户配置
func (u *userAppImpl) GetConfig(ctx context.Context) (*types.GetUserConfigRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	config, err := u.userDomain.GetConfig(ctx, userId)
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
	return u.userDomain.UpdateConfig(ctx, userId, req.Appearance)
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
