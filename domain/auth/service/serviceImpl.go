package service

import (
	"context"
	"errors"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
	"naotodoserver/domain/auth/valueobjects"
)

// 获取认证域实现
func GetAuthDomainImpl(
	jwt repositories.JWT,
	userRepo repositories.User,
	sessionRepo repositories.Session,
	rateLimitRepo repositories.RateLimit,
) AuthDomain {
	return &authDomainImpl{
		jwtRepo:       jwt,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		rateLimitRepo: rateLimitRepo,
	}
}

// 创建用户
// @param ctx 上下文
// @param createUserValueObject 用户值对象
// @return 用户实体
// @return error 错误
func (a *authDomainImpl) CreateUser(
	ctx context.Context,
	createUserValueObject *valueobjects.CreateUser,
) (*entities.User, error) {
	return a.userRepo.Create(ctx, createUserValueObject)
}

// 创建会话
// @param ctx 上下文
// @param userId 用户ID
// @param token 会话令牌
// @return error 错误
func (a *authDomainImpl) CreateSession(ctx context.Context, userId int64, token string) error {
	return a.sessionRepo.Create(ctx, &entities.Session{UserId: userId, Token: token})
}

// 根据邮箱查找用户
// @param ctx 上下文
// @param email 邮箱
// @return 用户实体
// @return error 错误
func (a *authDomainImpl) FindUserByEmail(
	ctx context.Context,
	email string,
) (*entities.User, error) {
	return a.userRepo.FindByEmail(ctx, email)
}

// 根据 Id 查找用户
// @param ctx 上下文
// @param id 用户ID
// @return 用户实体
// @return error 错误
func (a *authDomainImpl) FindUserById(
	ctx context.Context,
	id int64,
) (*entities.User, error) {
	return a.userRepo.FindById(ctx, id)
}

// 根据用户 Id 和 Token 查找会话
// @param ctx 上下文
// @param userId 用户ID
// @param token 会话令牌
// @return 会话实体
// @return error 错误
func (a *authDomainImpl) FindSessionByUserIdAndToken(
	ctx context.Context,
	userId int64,
	token string,
) (*entities.Session, error) {
	session := a.sessionRepo.FindByUserIdAndToken(ctx, userId, token)
	if !session.IsValid() {
		return nil, errors.New("会话不存在")
	}
	return session, nil
}

// 更新会话中 Token 字段值 (新的 JWT)
// @param ctx 上下文
// @param sessionEntity 会话实体
// @return error 错误
func (a *authDomainImpl) UpdateSessionToken(
	ctx context.Context,
	sessionEntity *entities.Session,
) error {
	return a.sessionRepo.UpdateToken(ctx, sessionEntity)
}

// 根据用户 Id 和 Token 删除会话
// @param ctx 上下文
// @param sessionEntity 会话实体
// @return error 错误
func (a *authDomainImpl) DeleteSession(ctx context.Context, sessionEntity *entities.Session) error {
	return a.sessionRepo.Delete(ctx, sessionEntity.UserId, sessionEntity.Token)
}

// 生成 JWT
// @param ctx 上下文
// @param userEntity 用户实体
// @return JWT 字符串
// @return error 错误
func (a *authDomainImpl) GenerateJWT(
	ctx context.Context,
	userEntity *entities.User,
) (string, error) {
	// 1. 判断 userEntity Id 是否合法
	if userEntity.Id <= 0 {
		return "", errors.New("userEntity Id 无效")
	}
	// 2. 创建 JWTClaims
	jwtClaims := &valueobjects.JWTClaims{
		UserId: userEntity.Id,
		Email:  userEntity.Email,
	}
	// 3. 创建 JWT
	token, err := a.jwtRepo.Generate(ctx, jwtClaims)
	if err != nil {
		return "", err
	}
	// 4. 返回结果
	return token, nil
}

// 解析 JWT
// @param ctx 上下文
// @param token JWT 字符串
// @return 用户ID
// @return error 错误
func (a *authDomainImpl) ParseJWT(ctx context.Context, token string) (int64, error) {
	jwtClaims, err := a.jwtRepo.Parse(ctx, token)
	if err != nil {
		return 0, err
	}
	return jwtClaims.UserId, nil
}

// 密码比较
// @param ctx 上下文
// @param password 明文密码
// @param encryptedPassword 加密后的密码
// @return bool
func (a *authDomainImpl) PasswordCompare(
	ctx context.Context,
	password []byte,
	encryptedPassword []byte,
) bool {
	return a.userRepo.PasswordCompare(password, encryptedPassword)
}

// 检查用户请求次数是否超出阈值 - 避免暴力注册
// @param ctx 上下文
// @param key 限流键
// @param limit 限流阈值
// @return error 错误
func (a *authDomainImpl) CheckRateLimit(ctx context.Context, key string, limit int8) error {
	// 1. 通过 Key 获取用户限流次数
	rateLimit := a.rateLimitRepo.Get(ctx, key)
	// 2. 判断用户请求次数是否超过限流阈值（1分钟内允许1次）
	if rateLimit >= limit {
		return errors.New("用户请求次数超过限流阈值")
	}
	// 3. 增加限流数
	a.rateLimitRepo.Incr(ctx, key)
	return nil
}
