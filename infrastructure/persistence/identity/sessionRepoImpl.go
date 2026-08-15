package identity

import (
	"context"
	"database/sql"
	"errors"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/ip2region"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	// 1. 获取上下文 ClientInfo
	clientInfo := iCtx.GetClientInfo(ctx)
	deviceId := clientInfo.DeviceId
	// 2. 构建会话模型（deviceId 为空时写入 NULL，不参与去重）
	createCond := SessionEntity2Model(sessionEntity)
	createCond.DeviceId = sql.NullString{String: deviceId, Valid: deviceId != ""}
	createCond.IP4 = clientInfo.IP4
	createCond.Region = clientInfo.IPRegion
	createCond.DeviceType = clientInfo.DeviceType
	createCond.ExpiredAt = time.Now().Add(time.Hour * 24 * 7)
	// 3. deviceId 非空：同设备重复登录覆盖旧会话（先取旧 token 用于失效缓存）
	var oldToken string
	if deviceId != "" {
		var existing models.UserSession
		err := sr.db.
			WithContext(ctx).
			Model(&models.UserSession{}).
			Where("user_id = ? AND device_id = ?", int64(sessionEntity.UserId), deviceId).
			First(&existing).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		oldToken = existing.Token
	}
	// 4. 写入会话；deviceId 非空时依赖 (user_id, device_id) 唯一索引做 upsert，避免并发产生孤儿行
	//    updated_at 一并刷新，与 CheckIn 换发（GORM 自动更新）保持行为一致
	tx := sr.db.WithContext(ctx)
	if deviceId != "" {
		tx = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"token", "expired_at", "ip4", "region", "device_type", "updated_at"}),
		})
	}
	if err := tx.Create(createCond).Error; err != nil {
		return err
	}
	// 5. 失效旧 token（同设备覆盖）与新 token 的会话缓存
	if oldToken != "" {
		sr.cache.Del(ctx, cache.SessionKey(
			int64(sessionEntity.UserId),
			oldToken,
		))
	}
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
	// 1. 创建删除模型（同时限定 user_id 与 token，避免误删其他用户会话）
	deleteCond := &models.UserSession{UserId: int64(userId), Token: token}
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

// DeleteById 根据会话 ID 删除指定会话
// @param ctx 上下文
// @param userId 用户 ID
// @param sessionId 会话 ID
// @return error 错误
func (sr *sessionRepoImpl) DeleteById(ctx context.Context, userId types.UserID, sessionId int64) error {
	// 1. 查询会话（限定 user_id + id，避免越权删除他人会话），拿 token 用于失效缓存
	session := &models.UserSession{}
	err := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND id = ?", int64(userId), sessionId).
		First(session).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domerr.ErrSessionNotFound
		}
		return err
	}
	// 2. 执行删除（仍限定 user_id + id）
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND id = ?", int64(userId), sessionId).
		Delete(&models.UserSession{})
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 失效该会话缓存
	sr.cache.Del(ctx, cache.SessionKey(int64(userId), session.Token))
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
	if err := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ?", int64(userId)).
		Pluck("token", &tokens).Error; err != nil {
		return err
	}
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

// DeleteByUserIdExceptToken 删除用户除指定 token 外的所有会话（退出其他全部设备）
// @param ctx 上下文
// @param userId 用户 ID
// @param keepToken 保留的会话令牌
// @return error 错误
func (sr *sessionRepoImpl) DeleteByUserIdExceptToken(ctx context.Context, userId types.UserID, keepToken string) error {
	// 1. 获取待删除会话的 token（用于失效缓存）
	var tokens []string
	if err := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND token <> ?", int64(userId), keepToken).
		Pluck("token", &tokens).Error; err != nil {
		return err
	}
	// 2. 执行删除
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND token <> ?", int64(userId), keepToken).
		Delete(&models.UserSession{})
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 失效所有被删会话缓存
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
	// 2. 创建查找模型（仅按用户 ID 与令牌匹配；Region/DeviceType 仅为记录字段，不参与匹配）
	findCond := &models.UserSession{UserId: int64(userId), Token: token}
	// 3. 执行查找
	sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(findCond).
		First(&session)
	// 4. 模型转换并返回
	return SessionModel2Entity(session)
}

// FindByUserId 根据用户 ID 查找现存会话（未过期）
// @param ctx 上下文
// @param userId 用户 ID
// @return 会话实体切片
// @return error 错误
func (sr *sessionRepoImpl) FindByUserId(
	ctx context.Context,
	userId types.UserID,
) ([]*entities.UserSession, error) {
	var sessions []*models.UserSession
	err := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND expired_at > ?", int64(userId), time.Now()).
		Order("created_at DESC").
		Find(&sessions).
		Error
	if err != nil {
		return nil, err
	}
	result := make([]*entities.UserSession, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, SessionModel2Entity(s))
	}
	return result, nil
}

// UpdateToken 更新会话令牌
// @param ctx 上下文
// @param sessionEntity 会话实体（含新令牌）
// @param oldToken 更新前的旧令牌
// @return error 错误
func (sr *sessionRepoImpl) UpdateToken(
	ctx context.Context,
	sessionEntity *entities.UserSession,
	oldToken string,
) error {
	// 1. 创建模型（CAS：仅当该用户仍持有 oldToken 时才更新，避免并发轮换互相覆盖）
	updateCond := &models.UserSession{Token: sessionEntity.Token}
	findCond := &models.UserSession{UserId: int64(sessionEntity.UserId), Token: oldToken}
	// 2. 执行更新
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(findCond).
		Updates(updateCond)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		// 会话已被其他请求轮换或删除，拒绝本次换发
		return errors.New("会话令牌已失效，请重新登录")
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
