package user

import (
	"context"
	"naotodoserver/domain/user/service"
	"naotodoserver/interfaces/types"
	"sync"

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

	// DeactiveUser 禁用用户
	DeactiveUser(ctx context.Context, req *types.DeactiveUserReq) error

	// ActiveUser 激活用户
	ActiveUser(ctx context.Context, req *types.ActiveUserReq) error

	// GetConfig 获取用户配置
	GetConfig(ctx context.Context) (*types.GetUserConfigRes, error)

	// UpdateConfig 更新用户配置
	UpdateConfig(ctx context.Context, req types.UpdateUserConfigReq) error
}

// userAppImpl 用户应用实现
type userAppImpl struct {
	userDomain service.UserDomain
}

// App 用户应用实例
var (
	App  *userAppImpl
	once sync.Once
)
