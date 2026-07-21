package identity

import (
	"context"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/ip2region"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"
	"time"

	"gorm.io/gorm"
)

type sessionRepoImpl struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewSessionRepo(db *gorm.DB, c *cache.Cache) repositories.UserSession {
	return &sessionRepoImpl{
		db:    db,
		cache: c,
	}
}

// Create 创建会话
// @param ctx 上下文
// @param sessionEntity 会话实体
// @return error 错误
func (sr *sessionRepoImpl) Create(ctx context.Context, sessionEntity *entities.UserSession) error {
	// 1. 查找现存记录
	currentSession := &models.UserSession{}
	findCond := &models.UserSession{UserId: int64(sessionEntity.UserId)}
	sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ?", findCond.UserId).
		First(currentSession)
	// 2. 获取上下文 ClientInfo
	clientInfo := iCtx.GetClientInfo(ctx)
	// 2. 执行结果存在逻辑 - 更新记录
	if currentSession.ID != 0 {
		tx := sr.db.WithContext(ctx).Model(&models.UserSession{}).
			Where(findCond).
			UpdateColumns(&models.UserSession{
				Token:      sessionEntity.Token,
				ExpiredAt:  time.Now().Add(time.Hour * 24 * 7),
				IP4:        clientInfo.IP4,
				Region:     clientInfo.IPRegion,
				DeviceType: clientInfo.DeviceType,
			})
		if tx.Error != nil {
			return tx.Error
		}
		// 3. 失效旧 token 与新 token 的会话缓存
		if currentSession.Token != "" {
			sr.cache.Del(ctx, cache.SessionKey(
				int64(sessionEntity.UserId),
				currentSession.Token,
			))
		}
		sr.cache.Del(ctx, cache.SessionKey(
			int64(sessionEntity.UserId),
			sessionEntity.Token,
		))
		return nil
	}
	// 3. 执行结果不存在逻辑 - 创建记录
	createCond := SessionEntity2Model(sessionEntity)
	createCond.IP4 = clientInfo.IP4
	createCond.Region = clientInfo.IPRegion
	createCond.DeviceType = clientInfo.DeviceType
	createCond.ExpiredAt = time.Now().Add(time.Hour * 24 * 7)
	tx := sr.db.WithContext(ctx).Model(&models.UserSession{}).
		Create(createCond)
	if tx.Error != nil {
		return tx.Error
	}
	// 4. 失效新 token 的会话缓存
	sr.cache.Del(ctx, cache.SessionKey(
		int64(sessionEntity.UserId),
		sessionEntity.Token,
	))
	return nil
}

// Delete 删除会话
// @param ctx 上下文
// @param userId 用户 ID
// @param token 会话令牌
// @return error 错误
func (sr *sessionRepoImpl) Delete(ctx context.Context, userId types.UserID, token string) error {
	// 1. 创建删除模型
	deleteCond := &models.UserSession{Token: token}
	// 2. 执行删除
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(&deleteCond).
		Delete(&models.UserSession{})
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 失效该会话缓存
	sr.cache.Del(ctx, cache.SessionKey(int64(userId), token))
	// 4. 返回结果
	return nil
}

// DeleteByUserId 删除用户所有会话
// @param ctx 上下文
// @param userId 用户 ID
// @return error 错误
func (sr *sessionRepoImpl) DeleteByUserId(ctx context.Context, userId types.UserID) error {
	// 1. 获取用户所有会话的 token
	var tokens []string
	sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ?", int64(userId)).
		Pluck("token", &tokens)
	// 2. 执行删除
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ?", int64(userId)).
		Delete(&models.UserSession{})
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 失效所有会话缓存
	for _, token := range tokens {
		sr.cache.Del(ctx, cache.SessionKey(int64(userId), token))
	}
	// 4. 返回结果
	return nil
}

// FindByUserIdAndToken 根据用户 ID和会话令牌查找会话
// @param ctx 上下文
// @param userId 用户 ID
// @param token 会话令牌
// @return 会话实体
// @note 会话不存在时返回 nil
func (sr *sessionRepoImpl) FindByUserIdAndToken(
	ctx context.Context,
	userId types.UserID,
	token string,
) *entities.UserSession {
	// 1. 创建结果模型
	session := &models.UserSession{}
	// 2. 创建查找模型（携带区域信息）
	findCond := &models.UserSession{UserId: int64(userId), Token: token}
	// 3. 填充 区域信息和设备类型
	clientInfo := iCtx.GetClientInfo(ctx)
	findCond.Region = clientInfo.IPRegion
	findCond.DeviceType = clientInfo.DeviceType
	// 4. 执行查找
	sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(findCond).
		First(&session)
	// 5. 模型转换并返回
	return SessionModel2Entity(session)
}

// UpdateToken 更新会话令牌
// @param ctx 上下文
// @param sessionEntity 会话实体
// @return error 错误
func (sr *sessionRepoImpl) UpdateToken(
	ctx context.Context,
	sessionEntity *entities.UserSession,
) error {
	// 1. 创建模型
	updateCond := &models.UserSession{Token: sessionEntity.Token}
	findCond := &models.UserSession{UserId: int64(sessionEntity.UserId)}
	// 2. 执行更新
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(findCond).
		Updates(updateCond)
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 失效该会话缓存
	sr.cache.Del(ctx, cache.SessionKey(
		int64(sessionEntity.UserId),
		sessionEntity.Token,
	))
	// 4. 返回结果
	return nil
}

// IsSessionValid 验证会话是否有效
// @param ctx 上下文
// @param userId 用户 ID
// @param token 会话令牌
// @return 会话是否有效
func (sr *sessionRepoImpl) IsSessionValid(
	ctx context.Context,
	userId types.UserID,
	token string,
) bool {
	// 1. 优先读取缓存，命中直接返回
	key := cache.SessionKey(int64(userId), token)
	var cached bool
	if sr.cache.Get(ctx, key, &cached) {
		return cached
	}
	// 2. 未命中则执行 MySQL 查询
	session := &models.UserSession{}
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(
			"user_id = ? AND token = ? AND expired_at > ?",
			int64(userId),
			token,
			time.Now(),
		).
		First(session)
	valid := tx.Error == nil && session.ID != 0
	// 3. 仅缓存有效会话，TTL 5 分钟
	if valid {
		sr.cache.Set(ctx, key, valid, time.Minute*5)
	}
	return valid
}

// Ip2Region IP 地址转换为区域
// @param ip IP 地址
// @return 区域
// @note 如果 IP 地址无效，返回空字符串
func (sr *sessionRepoImpl) Ip2Region(ip string) (string, error) {
	// 1. 获取 IP 地址区域信息
	region, err := ip2region.GetIp2RegionImpl().ParseIp(ip)
	if err != nil {
		return "", err
	}
	// 2. 返回结果
	return region, nil
}
