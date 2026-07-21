package user

import (
	"context"

	taskApp "naotodoserver/application/task"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// UserApp 用户应用接口
type UserApp interface {
	// UpdateNickname 更新用户昵称
	UpdateNickname(
		ctx context.Context,
		req types.UpdateUserNicknameReq,
	) error

	// GetProfile 获取用户个人信息
	GetProfile(ctx context.Context) (*types.GetUserProfileRes, error)

	// UpdatePassword 更新用户密码
	UpdatePassword(
		ctx context.Context,
		req types.UpdateUserPasswordReq,
	) error

	// UpdateAvatar 更新用户头像
	UpdateAvatar(
		ctx context.Context,
		req types.UpdateUserAvatarReq,
	) (*types.UpdateUserAvatarRes, error)

	// UpdateAvatarByFile 更新用户头像（通过文件上传）
	UpdateAvatarByFile(
		ctxRaw *gin.Context,
		ctx context.Context,
	) (*types.UpdateUserAvatarRes, error)

	// DeleteUser 删除用户（注销账户）
	DeleteUser(ctx context.Context, req *types.DeleteUserReq) error

	// RestoreUser 激活用户
	RestoreUser(ctx context.Context, req *types.RestoreUserReq) error

	// GetConfig 获取用户配置
	GetConfig(ctx context.Context) (*types.GetUserConfigRes, error)

	// UpdateConfig 更新用户配置
	UpdateConfig(ctx context.Context, req types.UpdateUserConfigReq) error

	// DeleteDeactivatedUsers 删除已注销用户（供定时任务调用）
	DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) error
}

// userAppImpl 用户应用实现
type userAppImpl struct {
	userRepo    repositories.User
	sessionRepo repositories.UserSession
	taskApp     taskApp.TaskCommentApp
}
