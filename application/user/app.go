package user

import (
	"context"
	"io"

	taskApp "naotodoserver/application/task"
	"naotodoserver/application/user/dto"
	"naotodoserver/domain/identity/repositories"
)

// UserApp 用户应用接口
type UserApp interface {
	// UpdateNickname 更新用户昵称
	UpdateNickname(
		ctx context.Context,
		req dto.UpdateNicknameInput,
	) error

	// GetProfile 获取用户个人信息
	GetProfile(ctx context.Context) (*dto.GetProfileOutput, error)

	// UpdatePassword 更新用户密码
	UpdatePassword(
		ctx context.Context,
		req dto.UpdatePasswordInput,
	) error

	// UpdateAvatar 更新用户头像
	UpdateAvatar(
		ctx context.Context,
		req dto.UpdateAvatarInput,
	) (*dto.UpdateAvatarOutput, error)

	// UpdateAvatarByFile 更新用户头像（通过文件上传）
	UpdateAvatarByFile(
		ctx context.Context,
		file io.Reader,
		filename string,
		size int64,
	) (*dto.UpdateAvatarOutput, error)

	// DeleteUser 删除用户（注销账户）
	DeleteUser(ctx context.Context, req dto.DeleteUserInput) error

	// RestoreUser 激活用户
	RestoreUser(ctx context.Context, req dto.RestoreUserInput) error

	// GetConfig 获取用户配置
	GetConfig(ctx context.Context) (*dto.GetConfigOutput, error)

	// UpdateConfig 更新用户配置
	UpdateConfig(ctx context.Context, req dto.UpdateConfigInput) error

	// DeleteDeactivatedUsers 删除已注销用户（供定时任务调用）
	DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) error
}

// userAppImpl 用户应用实现
type userAppImpl struct {
	userRepo      repositories.User
	sessionRepo   repositories.UserSession
	taskApp       taskApp.TaskCommentApp
	avatarStorage AvatarStorage
}
