package entities

import (
	"naotodoserver/domain/types"
	"time"
)

// UserSession 用户会话实体
// 用于表示用户登录后的会话信息
// 包含用户 ID、会话令牌、过期时间、设备类型、设备 ID 等属性
type UserSession struct {
	types.EntityBase
	UserId     types.UserID
	Token      string
	ExpiredAt  time.Time
	DeviceType string
	DeviceId   string
	IP4        string
	Region     string
}

// IsIdValid 检查用户会话 ID 是否有效
func (s *UserSession) IsIdValid() bool {
	return s != nil && s.Id > 0
}

// IsExpired 检查用户会话是否过期
func (s *UserSession) IsExpired() bool {
	return s != nil && s.ExpiredAt.Before(time.Now())
}

// IsValid 检查用户会话是否有效
func (s *UserSession) IsValid() bool {
	return s.IsIdValid() && !s.IsExpired()
}
