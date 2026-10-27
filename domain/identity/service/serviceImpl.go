package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/domain/types"
)

// LoginFailureThreshold 窗口内允许的连续登录失败次数，达到即暂停登录（SEC-04）
const LoginFailureThreshold int64 = 5

const (
	// LoginFailureWindow 登录失败计数窗口
	LoginFailureWindow = 5 * time.Minute
	// LoginBlockDuration 达到阈值后的暂停登录时长（到期自动解除）
	LoginBlockDuration = 5 * time.Minute
)

func NewIdentityDomain(
	jwt repositories.JWT,
	userSessionRepo repositories.UserSession,
	rateLimitRepo repositories.RateLimit,
	loginThrottle repositories.LoginThrottle,
) IdentityDomain {
	return &identityDomainImpl{
		jwtRepo:         jwt,
		userSessionRepo: userSessionRepo,
		rateLimitRepo:   rateLimitRepo,
		loginThrottle:   loginThrottle,
	}
}

// CreateSession 创建用户会话
func (d *identityDomainImpl) CreateSession(
	ctx context.Context,
	userId types.UserID,
	token string,
) error {
	return d.userSessionRepo.Create(
		ctx,
		&entities.UserSession{UserId: userId, Token: token},
	)
}

// FindSessionByUserIdAndToken 根据用户ID和会话令牌查找用户会话
func (d *identityDomainImpl) FindSessionByUserIdAndToken(
	ctx context.Context,
	userId types.UserID,
	token string,
) (*entities.UserSession, error) {
	session := d.userSessionRepo.FindByUserIdAndToken(ctx, userId, token)
	if !session.IsValid() {
		return nil, domerr.ErrSessionNotFound
	}
	return session, nil
}

// DeleteSession 删除用户会话
func (d *identityDomainImpl) DeleteSession(
	ctx context.Context,
	sessionEntity *entities.UserSession,
) error {
	return d.userSessionRepo.Delete(ctx, sessionEntity.UserId, sessionEntity.Token)
}

// ListSessions 获取用户现存会话列表
func (d *identityDomainImpl) ListSessions(
	ctx context.Context,
	userId types.UserID,
) ([]*entities.UserSession, error) {
	return d.userSessionRepo.FindByUserId(ctx, userId)
}

// DeleteSessionById 根据会话 ID 删除指定会话
func (d *identityDomainImpl) DeleteSessionById(
	ctx context.Context,
	userId types.UserID,
	sessionId int64,
) error {
	return d.userSessionRepo.DeleteById(ctx, userId, sessionId)
}

// DeleteOtherSessions 删除用户除指定 token 外的所有会话
func (d *identityDomainImpl) DeleteOtherSessions(
	ctx context.Context,
	userId types.UserID,
	keepToken string,
) error {
	// 撤销条数仅供审计日志，领域层不承载可观测性
	_, err := d.userSessionRepo.DeleteByUserIdExceptToken(ctx, userId, keepToken)
	return err
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
func (d *identityDomainImpl) ParseJWT(ctx context.Context, token string) (types.UserID, error) {
	jwtClaims, err := d.jwtRepo.Parse(ctx, token)
	if err != nil {
		return 0, err
	}
	return types.UserID(jwtClaims.UserId), nil
}

// IsJWTExpired 检查 JWT 令牌是否已过期
func (d *identityDomainImpl) IsJWTExpired(ctx context.Context, token string) bool {
	return d.jwtRepo.IsExpired(ctx, token)
}

// CheckRateLimit 检查用户请求次数是否超过限流阈值（原子检查 + 计数）
func (d *identityDomainImpl) CheckRateLimit(ctx context.Context, key string, limit int64) error {
	allowed, err := d.rateLimitRepo.Allow(ctx, key, limit)
	if err != nil {
		return fmt.Errorf("identity.CheckRateLimit: %w", err)
	}
	if !allowed {
		return errors.New("用户请求次数超过限流阈值")
	}
	return nil
}

// IsLoginBlocked 检查账号是否因连续登录失败被暂停（SEC-04）
// 只做一次计数查询：命中时上层可在查库与 bcrypt 之前快速返回
func (d *identityDomainImpl) IsLoginBlocked(ctx context.Context, email string) bool {
	return d.loginThrottle.IsBlocked(ctx, valueobjects.AccountKey(email))
}

// RecordLoginFailure 记录一次口令失败；窗口内累计达到阈值时暂停该账号登录
func (d *identityDomainImpl) RecordLoginFailure(ctx context.Context, email string) {
	accountKey := valueobjects.AccountKey(email)
	count := d.loginThrottle.RecordFailure(ctx, accountKey, LoginFailureWindow)
	if count < LoginFailureThreshold {
		return
	}
	d.loginThrottle.Block(ctx, accountKey, LoginBlockDuration)
}

// ClearLoginFailures 登录成功后清除账号的失败计数与暂停状态
func (d *identityDomainImpl) ClearLoginFailures(ctx context.Context, email string) {
	d.loginThrottle.Reset(ctx, valueobjects.AccountKey(email))
}
