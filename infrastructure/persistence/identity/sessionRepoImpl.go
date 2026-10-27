package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/ip2region"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"

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
	//    deleted_at 必须显式置 NULL：同 (user_id, device_id) 若是被登出/下线的软删行（墓碑），
	//    本次登录已验密码，应当接管并复活该行；否则 token 会写进「已删」行，
	//    客户端拿到的新 token 永远无法通过带软删作用域的鉴权/checkin 查询（10041 / 10022）
	tx := sr.db.WithContext(ctx)
	if deviceId != "" {
		doUpdates := clause.AssignmentColumns([]string{
			"token", "expired_at", "ip4", "region", "device_type", "updated_at",
		})
		doUpdates = append(doUpdates, clause.Assignment{
			Column: clause.Column{Name: "deleted_at"},
			Value:  gorm.Expr("NULL"),
		})
		tx = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
			DoUpdates: doUpdates,
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
	// 6. 写后回读：确认新 token 能被鉴权路径（deleted_at IS NULL 且未过期）命中，
	//    禁止「登录 200 却发出死 token」再次静默发生
	return sr.assertSessionVisible(ctx, sessionEntity.UserId, sessionEntity.Token)
}

// assertSessionVisible 写后回读校验：新 token 必须能被鉴权路径命中
// @param ctx 上下文
// @param userId 用户 ID
// @param token 期望可用的会话令牌（仅用于查询，不写日志）
// @return error 不可见时返回带原因分类的错误（软删行未复活 / 行未落库 / 会话已过期）
func (sr *sessionRepoImpl) assertSessionVisible(
	ctx context.Context,
	userId types.UserID,
	token string,
) error {
	// 1. 走与鉴权相同的读路径（GORM 软删作用域 + IsValid 过期判定）
	if session := sr.FindByUserIdAndToken(ctx, userId, token); session.IsValid() {
		return nil
	}
	// 2. 失败路径诊断：raw 查询绕过软删作用域，区分具体原因（便于线上一次定位）
	var row struct {
		ID        int64
		DeletedAt sql.NullTime
		ExpiredAt sql.NullTime
	}
	err := sr.db.WithContext(ctx).Raw(
		"SELECT id, deleted_at, expired_at FROM user_sessions "+
			"WHERE user_id = ? AND token = ? LIMIT 1",
		int64(userId),
		token,
	).Scan(&row).Error
	switch {
	case err != nil:
		return fmt.Errorf(
			"会话写入后不可见（回读查询失败）: user_id=%d token_len=%d: %w",
			int64(userId), len(token), err,
		)
	case row.ID == 0:
		return fmt.Errorf(
			"会话写入后不可见（行未落库）: user_id=%d token_len=%d",
			int64(userId), len(token),
		)
	case row.DeletedAt.Valid:
		return fmt.Errorf(
			"会话写入后不可见（软删行未复活）: session_id=%d user_id=%d deleted_at=%s",
			row.ID, int64(userId), row.DeletedAt.Time.Format(time.RFC3339),
		)
	default:
		return fmt.Errorf(
			"会话写入后不可见（会话已过期）: session_id=%d user_id=%d expired_at=%s",
			row.ID, int64(userId), row.ExpiredAt.Time.Format(time.RFC3339),
		)
	}
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
	// 3. 删除未命中任何行 ⇒ 该 (user_id, token) 已不存在（token 已被 checkin 轮换、
	//    会话已被删除或已登出），必须上抛而不是回「登出成功」：
	//    否则调用方无法区分「已登出」与「什么都没做」，客户端会误以为会话已撤销
	if tx.RowsAffected == 0 {
		return domerr.ErrSessionNotFound
	}
	// 4. 失效该会话缓存
	sr.cache.Del(ctx, cache.SessionKey(int64(userId), token))
	// 5. 返回结果
	return nil
}

// DeleteById 根据会话 ID 删除指定会话
// @param ctx 上下文
// @param userId 用户 ID
// @param sessionId 会话 ID
// @return error 错误
func (sr *sessionRepoImpl) DeleteById(
	ctx context.Context,
	userId types.UserID,
	sessionId int64,
) error {
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
// @param keepToken 保留的会话令牌（空串 = 不保留任何会话，全部撤销）
// @return 实际撤销的会话条数（供审计日志；不含 token）
// @return error 错误
func (sr *sessionRepoImpl) DeleteByUserIdExceptToken(
	ctx context.Context,
	userId types.UserID,
	keepToken string,
) (int64, error) {
	// 1. 获取待删除会话的 token（用于失效缓存）
	var tokens []string
	if err := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND token <> ?", int64(userId), keepToken).
		Pluck("token", &tokens).Error; err != nil {
		return 0, err
	}
	// 2. 执行删除
	tx := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where("user_id = ? AND token <> ?", int64(userId), keepToken).
		Delete(&models.UserSession{})
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 3. 失效所有被删会话缓存
	for _, token := range tokens {
		sr.cache.Del(ctx, cache.SessionKey(int64(userId), token))
	}
	// 4. 返回结果（RowsAffected = 实际软删条数）
	return tx.RowsAffected, nil
}

// DeleteExpiredBefore 硬删除过期时间早于 cutoff 的未登出会话
// @param ctx 上下文
// @param cutoff 过期时间上界（不含）
// @param limit 单批最多删除行数
// @return 实际删除行数
// @return error 错误
// @note 使用 raw DELETE + LIMIT 物理删除：绕过 GORM 软删语义，无 DDL、不触及 token 缓存
// @note deleted_at IS NULL 保证已登出会话只走登出宽限期，不被过期条件提前回收
func (sr *sessionRepoImpl) DeleteExpiredBefore(
	ctx context.Context,
	cutoff time.Time,
	limit int,
) (int64, error) {
	tx := sr.db.WithContext(ctx).Exec(
		"DELETE FROM user_sessions WHERE expired_at < ? AND deleted_at IS NULL "+
			"ORDER BY expired_at LIMIT ?",
		cutoff,
		limit,
	)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}

// DeleteRevokedBefore 硬删除登出时间早于 cutoff 的会话
// @param ctx 上下文
// @param cutoff 登出时间上界（不含）
// @param limit 单批最多删除行数
// @return 实际删除行数
// @return error 错误
// @note 使用 raw DELETE + LIMIT 物理删除：绕过 GORM 软删语义，无 DDL、不触及 token 缓存
func (sr *sessionRepoImpl) DeleteRevokedBefore(
	ctx context.Context,
	cutoff time.Time,
	limit int,
) (int64, error) {
	tx := sr.db.WithContext(ctx).Exec(
		"DELETE FROM user_sessions WHERE deleted_at IS NOT NULL AND deleted_at < ? "+
			"ORDER BY deleted_at LIMIT ?",
		cutoff,
		limit,
	)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
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
	// 查不到（软删/无此行）是正常结果；其余 DB 故障必须留痕，否则会被上层统一折叠成
	// 「用户凭证验证失败（10041）」，线上无法区分「会话不存在」与「数据库故障」
	if err := sr.db.
		WithContext(ctx).
		Model(&models.UserSession{}).
		Where(findCond).
		First(&session).
		Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logging.Warnf(
			"会话查询失败（按 user_id + token）: user_id=%d token_len=%d err=%v",
			int64(userId), len(token), err,
		)
	}
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
		return domerr.Publicf("会话令牌已失效，请重新登录")
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
