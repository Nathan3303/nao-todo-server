package identity

import (
	"context"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
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

/**
 * Create Session
 */
func (sr *sessionRepoImpl) Create(ctx context.Context, sessionEntity *entities.UserSession) error {
	// 1. 查找现存记录
	currentSession := &models.UserSession{}
	findCond := &models.UserSession{UserId: sessionEntity.UserId}
	sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(findCond).
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
			sr.cache.Del(ctx, cache.SessionKey(sessionEntity.UserId, currentSession.Token))
		}
		sr.cache.Del(ctx, cache.SessionKey(sessionEntity.UserId, sessionEntity.Token))
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
	sr.cache.Del(ctx, cache.SessionKey(sessionEntity.UserId, sessionEntity.Token))
	return nil
}

/**
 * Delete Session
 */
func (sr *sessionRepoImpl) Delete(ctx context.Context, userId int64, token string) error {
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
	sr.cache.Del(ctx, cache.SessionKey(userId, token))
	// 4. 返回结果
	return nil
}

/**
 * Find Session by User ID and Token
 */
func (sr *sessionRepoImpl) FindByUserIdAndToken(
	ctx context.Context,
	userId int64,
	token string,
) *entities.UserSession {
	// 1. 创建结果模型
	session := &models.UserSession{}
	// 2. 创建查找模型（携带区域信息）
	findCond := &models.UserSession{UserId: userId, Token: token}
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

/**
 * Update Session Token
 */
func (sr *sessionRepoImpl) UpdateToken(
	ctx context.Context,
	sessionEntity *entities.UserSession,
) error {
	// 1. 创建模型
	updateCond := &models.UserSession{Token: sessionEntity.Token}
	findCond := &models.UserSession{UserId: sessionEntity.UserId}
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
	sr.cache.Del(ctx, cache.SessionKey(sessionEntity.UserId, sessionEntity.Token))
	// 4. 返回结果
	return nil
}

/**
 * Is Session Valid
 */
func (sr *sessionRepoImpl) IsSessionValid(ctx context.Context, userId int64, token string) bool {
	// 1. 优先读取缓存，命中直接返回
	key := cache.SessionKey(userId, token)
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
			userId,
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

/**
 * Ip to Region
 */
func (sr *sessionRepoImpl) Ip2Region(ip string) (string, error) {
	// 1. 获取 IP 地址区域信息
	region, err := ip2region.GetIp2RegionImpl().ParseIp(ip)
	if err != nil {
		return "", err
	}
	// 2. 返回结果
	return region, nil
}
