package repositories

import (
	"context"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/valueobjects"
)

// User 用户仓储接口
type User interface {
	// Create 创建用户（注册用，传入值对象）
	CreateByVO(ctx context.Context, createUser *valueobjects.CreateUser) (*entities.User, error)

	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindById 根据ID查找用户
	FindById(ctx context.Context, id int64) (*entities.User, error)

	// UpdateNickname 更新昵称
	UpdateNickname(ctx context.Context, userId int64, nickname string) error

	// UpdateAvatar 更新头像
	UpdateAvatar(ctx context.Context, userId int64, avatarUrl string) error

	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, userId int64, oldPassword, newPassword string) error

	// PasswordCompare 密码比对
	PasswordCompare(password, encryptedPassword []byte) bool

	// Deactive 注销用户
	Deactive(ctx context.Context, userId int64) error

	// Active 激活用户
	Active(ctx context.Context, userId int64) error

	// Delete 删除用户
	Delete(ctx context.Context, userId int64) error

	// DeleteDeactivatedUsers 删除已注销用户
	DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) (int64, error)

	// GetConfig 获取用户配置
	GetConfig(ctx context.Context, userId int64) (*entities.UserConfig, error)

	// UpdateConfig 更新用户配置
	UpdateConfig(ctx context.Context, userId int64, appearance string) error
}
