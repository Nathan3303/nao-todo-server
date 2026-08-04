package auth

import (
	"context"
	"errors"
	"fmt"
	"naotodoserver/application/auth/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/service"
	domaintypes "naotodoserver/domain/types"
	"time"
)

// NewAuthApp 创建认证应用层实例
func NewAuthApp(
	identityDomain service.IdentityDomain,
	userRepo repositories.User,
	sessionRepo repositories.UserSession,
) AuthApp {
	return &authAppImpl{
		identityDomain: identityDomain,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
	}
}

// 处理用户登录
// @description 通过 signInInput 中的邮箱和密码验证用户身份，验证成功后生成 JWT 令牌并创建用户会话，最后返回 JWT 令牌
// @param ctx 上下文
// @param signInInput 登录入参
// @return 登录出参
// @return error 错误
func (as *authAppImpl) SignIn(
	ctx context.Context,
	signInInput *dto.SignInInput,
) (*dto.SignInOutput, error) {
	// 1. 通过 Email 查找用户记录
	userEntity, err := as.userRepo.FindByEmail(ctx, signInInput.Email)
	if err != nil {
		return nil, fmt.Errorf("auth.SignIn.FindByEmail: %w", err)
	}
	// 2. 比对密码
	isMatched := as.userRepo.PasswordCompare(
		[]byte(signInInput.Password),
		[]byte(userEntity.Password),
	)
	if !isMatched {
		return nil, domerr.ErrPasswordMismatch
	}
	// 2. 创建 JWT 令牌
	jwtString, err := as.identityDomain.GenerateJWT(ctx, userEntity)
	if err != nil {
		return nil, fmt.Errorf("auth.SignIn.GenerateJWT: %w", err)
	}
	// 3. 创建 Session
	err = as.identityDomain.CreateSession(ctx, domaintypes.UserID(userEntity.Id), jwtString)
	if err != nil {
		return nil, fmt.Errorf("auth.SignIn.CreateSession: %w", err)
	}
	// 4. 检查是否出于待注销状态，并返回注销时间
	var pendingDeletion = false
	var deletedAt = ""
	if userEntity.IsDeactived() {
		pendingDeletion = true
		// deactiveAt 标记，但归属于 deletedAt
		deletedAt = userEntity.DeactivedAt.ToString(time.RFC3339)
	}
	// 5. 登录成功 返回 JWT
	return &dto.SignInOutput{
		Token:           jwtString,
		PendingDeletion: pendingDeletion,
		DeletedAt:       deletedAt,
	}, nil
}

/*
 * SignUp 处理用户注册
 * 通过 signUpInput 中的用户信息，执行密码加密后用户信息落库
 */
func (as *authAppImpl) SignUp(
	ctx context.Context,
	signUpInput *dto.SignUpInput,
) error {
	// 通过 Email 查找用户记录
	_, err := as.userRepo.FindByEmail(ctx, signUpInput.Email)
	switch {
	case err == nil:
		// 邮箱已存在，返回明确的领域错误（不包装 nil，避免错误信息损坏）
		return domerr.ErrEmailExists
	case !errors.Is(err, domerr.ErrUserNotFound):
		// DB 故障等真实错误直接透传，避免被误判为"邮箱可用"而继续走创建分支
		return fmt.Errorf("auth.SignUp.FindByEmail: %w", err)
	}
	// 转换为 CreateUserValueObject
	createUserValueObject, err := SignUpInputToCreateUserValueObject(signUpInput)
	if err != nil {
		return err
	}
	// 创建
	_, err = as.userRepo.CreateByVO(ctx, createUserValueObject)
	if err != nil {
		// 并发注册下唯一索引冲突，识别为邮箱已存在
		if errors.Is(err, domerr.ErrEmailExists) {
			return domerr.ErrEmailExists
		}
		return fmt.Errorf("auth.SignUp.Create: %w", err)
	}
	// 注册成功
	return nil
}

/*
 * CheckIn 处理用户检入
 * 通过 checkInInput 中的令牌信息，验证用户会话，并生成新的 JWT 并更新用户会话，最后返回新的 JWT
 * 用于在 JWT 或会话期限内的免密登录
 */
func (as *authAppImpl) CheckIn(
	ctx context.Context,
	checkInInput *dto.CheckInInput,
) (*dto.CheckInOutput, error) {
	// 1. 解析 JWT 令牌
	userId, err := as.identityDomain.ParseJWT(ctx, checkInInput.Token)
	if err != nil {
		return nil, fmt.Errorf("auth.CheckIn.ParseJWT: %w", err)
	}
	// 2. 通过 JWT 令牌和用户 ID 查找会话
	sessionEntity, err := as.identityDomain.FindSessionByUserIdAndToken(
		ctx,
		userId,
		checkInInput.Token,
	)
	// 会话不存在：
	if err != nil {
		return nil, fmt.Errorf("auth.CheckIn.FindSession: %w", err)
	}
	if !sessionEntity.IsValid() {
		return nil, domerr.ErrTokenExpired
	}
	// 会话存在：
	// 3. 查找最新的用户信息
	userEntity, err := as.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("auth.CheckIn.FindById: %w", err)
	}
	// 4. 构建新的 JWT 令牌
	newJWT, err := as.identityDomain.GenerateJWT(ctx, userEntity)
	if err != nil {
		return nil, fmt.Errorf("auth.CheckIn.GenerateJWT: %w", err)
	}
	// 5. 更新会话中的 Token 字段（CAS：仅当该用户仍持有旧 Token 时更新，避免并发轮换互相覆盖）
	sessionEntity.Token = newJWT
	err = as.sessionRepo.UpdateToken(ctx, sessionEntity, checkInInput.Token)
	if err != nil {
		return nil, fmt.Errorf("auth.CheckIn.UpdateToken: %w", err)
	}
	// 6. 检查是否出于待注销状态，并返回注销时间
	var pendingDeletion = false
	var deletedAt = ""
	if userEntity.IsDeactived() {
		pendingDeletion = true
		// deactiveAt 标记，但归属于 deletedAt
		deletedAt = userEntity.DeactivedAt.ToString(time.RFC3339)
	}
	// 7. 返回
	return &dto.CheckInOutput{
		Token:           newJWT,
		PendingDeletion: pendingDeletion,
		DeletedAt:       deletedAt,
	}, nil
}

/*
 * SignOut 处理用户登出
 * 通过 signOutInput 中的令牌信息，验证用户会话，并删除用户会话，最后返回成功结果
 */
func (as *authAppImpl) SignOut(
	ctx context.Context,
	signOutInput *dto.SignOutInput,
) error {
	// 1. 解析 JWT 令牌
	userId, err := as.identityDomain.ParseJWT(ctx, signOutInput.Token)
	if err != nil {
		return fmt.Errorf("auth.SignOut.ParseJWT: %w", err)
	}
	// 2. 通过 JWT 令牌和用户 ID 删除会话
	err = as.identityDomain.DeleteSession(ctx, &entities.UserSession{
		UserId: userId,
		Token:  signOutInput.Token,
	})
	if err != nil {
		return fmt.Errorf("auth.SignOut.DeleteSession: %w", err)
	}
	// 3. 返回结果
	return nil
}

/*
 * Validate 处理用户令牌验证
 * 通过 Header 中的令牌信息，验证用户会话，并返回验证结果
 */
func (as *authAppImpl) Validate(
	ctx context.Context,
	token string,
) (domaintypes.UserID, error) {
	// 1. 解析 JWT（仅验签，不校验过期）
	userId, err := as.identityDomain.ParseJWT(ctx, token)
	if err != nil {
		return 0, fmt.Errorf("auth.Validate.ParseJWT: %w", err)
	}
	// 2. 检查 JWT 是否过期（会话有效期长于 JWT 有效期，须单独校验）
	if as.identityDomain.IsJWTExpired(ctx, token) {
		return 0, domerr.ErrTokenExpired
	}
	// 3. 检查会话
	session, err := as.identityDomain.FindSessionByUserIdAndToken(ctx, userId, token)
	if err != nil {
		return 0, fmt.Errorf("auth.Validate.FindSession: %w", err)
	}
	if !session.IsValid() {
		return 0, domerr.ErrTokenExpired
	}
	// 4. 返回结果
	return userId, nil
}

/*
 * RateLimit 处理用户限流
 * 通过 clientIP 检查用户请求次数是否超过限流阈值
 */
func (as *authAppImpl) RateLimit(
	ctx context.Context,
	clientIP string,
	limit int64,
) error {
	// 1. 构造限流 key
	key := "rate_limit:" + clientIP
	// 2. 检查是否超过限流阈值
	err := as.identityDomain.CheckRateLimit(ctx, key, limit)
	// 3. 返回结果
	if err != nil {
		return fmt.Errorf("auth.RateLimit: %w", err)
	}
	return nil
}
