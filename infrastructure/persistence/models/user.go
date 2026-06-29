package models

import (
	"database/sql"
	"time"
)

// User 用户模型
type User struct {
	// 基础属性
	ModelBase

	// 用户账号
	// 唯一，不能重复
	// 例如：nathan
	Account string `gorm:"unique;not null;size:128;index:idx_user_account"`

	// 用户邮箱
	// 唯一，不能重复
	// 例如：nathan@example.com
	Email string `gorm:"unique;not null;size:128;index:idx_user_email"`

	// 用户密码
	// 密码长度为 64 个字符
	Password string `gorm:"not null;size:64"`

	// 用户昵称
	// 例如：Nathan
	Nickname string `gorm:"not null;size:64"`

	// 用户头像
	Avatar string `gorm:"null;size:256"`

	// 创建来源
	// 例如：unknown:unknown
	CreatedFrom string `gorm:"not null;size:64;default:'unknown:unknown'"`

	// 用户角色
	// 例如：user
	// 0: 普通用户
	// 1: 管理员
	Role uint8 `gorm:"default:0;type:tinyint(1)"`

	// 用户状态（未使用）
	State uint8 `gorm:"default:0;type:tinyint(1)"`

	// 用户禁用时间
	// 例如：2023-01-01 00:00:00
	// 当 DeactivedAt 不为 NULL 时被检查，判断用户是否需要更新 DeletedAt
	DeactivedAt sql.NullTime `gorm:"null;index:idx_user_deactived_at"`

	// 用户配置
	Config *UserConfig `gorm:"foreignKey:UserId;references:ID;constraint:OnDelete:CASCADE;"`
}

// UserConfig 用户配置模型
type UserConfig struct {
	// 基础属性
	ModelBase

	// 用户 ID
	// 用于关联 User 模型
	UserId int64 `gorm:"not null;index:idx_user_config_user"`

	// 用户外观
	// 例如：system、light 以及 dark
	Appearance string `gorm:"size:8;default:'system'"`
}

// UserSession 用户会话模型
type UserSession struct {
	// 基础属性
	ModelBase

	// 用户 ID
	// 用于关联用户表，获取用户会话信息
	UserId int64 `gorm:"not null;index:idx_user_session_user_id"`

	// 会话令牌
	// 用于验证用户会话，防止未授权访问
	// 通常采用 JWT 格式，包含用户 ID、过期时间等信息
	Token string `gorm:"not null;size:512;index:idx_user_session_token"`

	// 过期时间
	// 用于判断会话是否需要重新登录
	ExpiredAt time.Time `gorm:"not null;index:idx_user_session_expired"`

	// 设备类型
	// 用于记录用户会话的设备类型，例如 PC、移动端、Web 等
	// 目标用于多设备登录
	DeviceType string `gorm:"not null;size:32"`

	// IP4 地址
	IP4 string `gorm:"not null;size:64"`

	// 区域
	Region string `gorm:"not null;size:64"`
}
