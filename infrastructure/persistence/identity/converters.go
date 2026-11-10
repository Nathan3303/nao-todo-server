package identity

import (
	"database/sql"
	"errors"

	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/auth"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// UserEntity2Model 用户实体转换为数据库模型
func UserEntity2Model(e *entities.User) *models.User {
	var m models.User
	m.ID = e.Id
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	m.DeletedAt = gorm.DeletedAt(e.DeletedAt.ToSqlNullTime())
	m.Account = e.Account
	m.Email = e.Email
	m.Password = e.Password
	m.Nickname = e.Nickname
	m.Avatar = e.Avatar
	m.CreatedFrom = e.CreatedFrom
	m.Role = uint8(e.Role)
	m.State = uint8(e.State)
	m.DeactivedAt = e.DeactivedAt.ToSqlNullTime()
	m.LastCancelRestoreAt = e.LastCancelRestoreAt.ToSqlNullTime()
	return &m
}

// UserModel2Entity 数据库模型转换为用户实体
func UserModel2Entity(m *models.User) *entities.User {
	var e entities.User
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.Account = m.Account
	e.Email = m.Email
	e.Password = m.Password
	e.Nickname = m.Nickname
	e.Avatar = m.Avatar
	e.CreatedFrom = m.CreatedFrom
	e.Role = entities.UserRole(m.Role)
	e.State = entities.UserState(m.State)
	e.DeactivedAt = types.NewNullableTimeByTime(m.DeactivedAt.Time)
	e.LastCancelRestoreAt = types.NewNullableTimeByTime(m.LastCancelRestoreAt.Time)
	return &e
}

// UserConfigEntity2Model 用户配置实体转换为数据库模型
func UserConfigEntity2Model(e *entities.UserConfig) *models.UserConfig {
	var m models.UserConfig
	m.ID = e.Id
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	m.DeletedAt = gorm.DeletedAt(e.DeletedAt.ToSqlNullTime())
	m.UserId = int64(e.UserId)
	m.Appearance = e.Appearance
	if e.Preferences != "" {
		preferences := e.Preferences
		m.Preferences = &preferences
	}
	return &m
}

// UserConfigModel2Entity 数据库模型转换为用户配置实体
func UserConfigModel2Entity(m *models.UserConfig) *entities.UserConfig {
	var e entities.UserConfig
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = types.UserID(m.UserId)
	e.Appearance = m.Appearance
	if m.Preferences != nil {
		e.Preferences = *m.Preferences
	}
	return &e
}

// SessionModel2Entity 数据库模型转换为会话实体
func SessionModel2Entity(m *models.UserSession) *entities.UserSession {
	var e entities.UserSession
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = types.UserID(m.UserId)
	e.Token = m.Token
	e.ExpiredAt = m.ExpiredAt
	e.DeviceType = m.DeviceType
	e.DeviceId = m.DeviceId.String
	e.IP4 = m.IP4
	e.Region = m.Region
	return &e
}

// SessionEntity2Model 会话实体转换为数据库模型
func SessionEntity2Model(e *entities.UserSession) *models.UserSession {
	var m models.UserSession
	m.ID = e.Id
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	m.DeletedAt = gorm.DeletedAt(e.DeletedAt.ToSqlNullTime())
	m.UserId = int64(e.UserId)
	m.Token = e.Token
	m.ExpiredAt = e.ExpiredAt
	m.DeviceType = e.DeviceType
	m.DeviceId = sql.NullString{String: e.DeviceId, Valid: e.DeviceId != ""}
	m.IP4 = e.IP4
	m.Region = e.Region
	return &m
}

// Claims2JWTClaimsVO 转换 JWT 断言为 JWT 令牌对象
// @param c JWT 断言
// @return JWT 令牌对象（出错时为 nil）
// @return error 断言为空 / 缺少有效签发时间（iat）时返回显式错误
// @note DEF-157：修复前直接解引用 `c.IssuedAt.Time`，遇到缺失 `iat` 的断言会 panic
// ⇒ gin Recovery 后返回 500 空体。手工构造的合法签名 token 即可触发，故必须显式报错。
func Claims2JWTClaimsVO(c *auth.Claims) (*valueobjects.JWTClaims, error) {
	if c == nil {
		return nil, errors.New("JWT 断言为空")
	}
	// iat 缺失（非 nil 判定）或显式 null（解析成零值时间）都视为非法断言
	if c.IssuedAt == nil || c.IssuedAt.IsZero() {
		return nil, errors.New("JWT 断言缺少有效的签发时间（iat）")
	}
	var vo valueobjects.JWTClaims
	vo.UserId = c.Id
	vo.Email = c.Payload
	vo.IssuedAt = c.IssuedAt.Time
	vo.Issuer = c.Issuer
	return &vo, nil
}
