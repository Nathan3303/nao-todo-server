package service

import (
	"context"
	"errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
)

func NewIdentityDomain(
	jwt repositories.JWT,
	userRepo repositories.User,
	userSessionRepo repositories.UserSession,
	rateLimitRepo repositories.RateLimit,
) IdentityDomain {
	return &identityDomainImpl{
		jwtRepo:         jwt,
		userRepo:        userRepo,
		userSessionRepo: userSessionRepo,
		rateLimitRepo:   rateLimitRepo,
	}
}

// --- 认证相关 --

// CreateUser 创建用户
func (d *identityDomainImpl) CreateUser(
	ctx context.Context,
	vo *valueobjects.CreateUser,
) (*entities.User, error) {
	return d.userRepo.CreateByVO(ctx, vo)
}

// CreateSession 创建用户会话
func (d *identityDomainImpl) CreateSession(ctx context.Context, userId int64, token string) error {
	return d.userSessionRepo.Create(
		ctx,
		&entities.UserSession{UserId: userId, Token: token},
	)
}

// FindSessionByUserIdAndToken 根据用户ID和会话令牌查找用户会话
func (d *identityDomainImpl) FindSessionByUserIdAndToken(
	ctx context.Context,
	userId int64,
	token string,
) (*entities.UserSession, error) {
	session := d.userSessionRepo.FindByUserIdAndToken(ctx, userId, token)
	if !session.IsValid() {
		return nil, errors.New("会话不存在")
	}
	return session, nil
}

// UpdateSessionToken 更新用户会话令牌
func (d *identityDomainImpl) UpdateSessionToken(
	ctx context.Context,
	sessionEntity *entities.UserSession,
) error {
	return d.userSessionRepo.UpdateToken(ctx, sessionEntity)
}

// DeleteSession 删除用户会话
func (d *identityDomainImpl) DeleteSession(
	ctx context.Context,
	sessionEntity *entities.UserSession,
) error {
	return d.userSessionRepo.Delete(ctx, sessionEntity.UserId, sessionEntity.Token)
}

// GenerateJWT 生成JWT令牌
func (d *identityDomainImpl) GenerateJWT(
	ctx context.Context,
	userEntity *entities.User,
) (string, error) {
	if userEntity.Id <= 0 {
		return "", errors.New("userEntity Id 无效")
	}
	jwtClaims := &valueobjects.JWTClaims{
		UserId: userEntity.Id,
		Email:  userEntity.Email,
	}
	return d.jwtRepo.Generate(ctx, jwtClaims)
}

// ParseJWT 解析JWT令牌
func (d *identityDomainImpl) ParseJWT(ctx context.Context, token string) (int64, error) {
	jwtClaims, err := d.jwtRepo.Parse(ctx, token)
	if err != nil {
		return 0, err
	}
	return jwtClaims.UserId, nil
}

// CheckRateLimit 检查用户请求次数是否超过限流阈值
func (d *identityDomainImpl) CheckRateLimit(ctx context.Context, key string, limit int8) error {
	if d.rateLimitRepo.Get(ctx, key) >= limit {
		return errors.New("用户请求次数超过限流阈值")
	}
	d.rateLimitRepo.Incr(ctx, key)
	return nil
}

// --- 用户管理 --

// FindByEmail 根据用户邮箱查找用户
func (d *identityDomainImpl) FindByEmail(
	ctx context.Context,
	email string,
) (*entities.User, error) {
	return d.userRepo.FindByEmail(ctx, email)
}

// FindById 根据用户ID查找用户
func (d *identityDomainImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	return d.userRepo.FindById(ctx, id)
}

// UpdateNickname 更新用户昵称
func (d *identityDomainImpl) UpdateNickname(
	ctx context.Context,
	userId int64,
	nickname string,
) error {
	return d.userRepo.UpdateNickname(ctx, userId, nickname)
}

// UpdateAvatar 更新用户头像
func (d *identityDomainImpl) UpdateAvatar(
	ctx context.Context,
	userId int64,
	avatar string,
) error {
	return d.userRepo.UpdateAvatar(ctx, userId, avatar)
}

// UpdatePassword 更新用户密码
func (d *identityDomainImpl) UpdatePassword(
	ctx context.Context,
	userId int64,
	password, newPassword string,
) error {
	return d.userRepo.UpdatePassword(ctx, userId, password, newPassword)
}

// Deactive 禁用用户
func (d *identityDomainImpl) Deactive(ctx context.Context, userId int64) error {
	return d.userRepo.Deactive(ctx, userId)
}

// Active 激活用户
func (d *identityDomainImpl) Active(ctx context.Context, userId int64) error {
	return d.userRepo.Active(ctx, userId)
}

// PasswordCompare 对比用户密码是否匹配
func (d *identityDomainImpl) PasswordCompare(password, encryptedPassword []byte) bool {
	return d.userRepo.PasswordCompare(password, encryptedPassword)
}

// GetConfig 获取用户配置
func (d *identityDomainImpl) GetConfig(
	ctx context.Context,
	userId int64,
) (*entities.UserConfig, error) {
	return d.userRepo.GetConfig(ctx, userId)
}

// UpdateConfig 更新用户配置
func (d *identityDomainImpl) UpdateConfig(
	ctx context.Context,
	userId int64,
	appearance string,
) error {
	return d.userRepo.UpdateConfig(ctx, userId, appearance)
}

// DeleteDeactivatedUsers 删除已注销用户
func (d *identityDomainImpl) DeleteDeactivatedUsers(
	ctx context.Context,
	dayOffset int8,
) (int64, error) {
	return d.userRepo.DeleteDeactivatedUsers(ctx, dayOffset)
}
