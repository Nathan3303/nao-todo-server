package identity

import (
	"context"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/ip2region"
	"naotodoserver/infrastructure/persistence/models"
	"time"

	"gorm.io/gorm"
)

type sessionRepoImpl struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) repositories.Session {
	return &sessionRepoImpl{
		db: db,
	}
}

/**
 * Create Session
 */
func (sr *sessionRepoImpl) Create(ctx context.Context, sessionEntity *entities.Session) error {
	// 1. 查找现存记录
	currentSession := &models.Session{}
	findCond := &models.Session{UserId: sessionEntity.UserId}
	sr.db.WithContext(ctx).Model(&models.Session{}).Where(findCond).First(currentSession)
	// 2. 获取上下文 ClientInfo
	clientInfo := iCtx.GetClientInfo(ctx)
	// 2. 执行结果存在逻辑 - 更新记录
	if currentSession.ID != 0 {
		tx := sr.db.WithContext(ctx).Model(&models.Session{}).
			Where(findCond).
			UpdateColumns(&models.Session{
				Token:      sessionEntity.Token,
				ExpiredAt:  time.Now().Add(time.Hour * 24 * 7),
				IP4:        clientInfo.IP4,
				Region:     clientInfo.IPRegion,
				DeviceType: clientInfo.DeviceType,
			})
		if tx.Error != nil {
			return tx.Error
		}
		return nil
	}
	// 3. 执行结果不存在逻辑 - 创建记录
	createCond := SessionEntity2Model(sessionEntity)
	createCond.IP4 = clientInfo.IP4
	createCond.Region = clientInfo.IPRegion
	createCond.DeviceType = clientInfo.DeviceType
	createCond.ExpiredAt = time.Now().Add(time.Hour * 24 * 7)
	tx := sr.db.WithContext(ctx).Model(&models.Session{}).
		Create(createCond)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

/**
 * Delete Session
 */
func (sr *sessionRepoImpl) Delete(ctx context.Context, userId int64, token string) error {
	// 1. 创建删除模型
	deleteCond := &models.Session{Token: token}
	// 2. 执行删除
	tx := sr.db.WithContext(ctx).Model(&models.Session{}).Where(&deleteCond).Delete(
		&models.Session{},
	)
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 返回结果
	return nil
}

/**
 * Find Session by User ID and Token
 */
func (sr *sessionRepoImpl) FindByUserIdAndToken(
	ctx context.Context,
	userId int64,
	token string,
) *entities.Session {
	// 1. 创建结果模型
	session := &models.Session{}
	// 2. 创建查找模型（携带区域信息）
	findCond := &models.Session{UserId: userId, Token: token}
	// 3. 填充 区域信息和设备类型
	clientInfo := iCtx.GetClientInfo(ctx)
	findCond.Region = clientInfo.IPRegion
	findCond.DeviceType = clientInfo.DeviceType
	// 4. 执行查找
	sr.db.WithContext(ctx).Model(&models.Session{}).Where(findCond).First(&session)
	// 5. 模型转换并返回
	return SessionModel2Entity(session)
}

/**
 * Update Session Token
 */
func (sr *sessionRepoImpl) UpdateToken(ctx context.Context, sessionEntity *entities.Session) error {
	// 1. 创建模型
	updateCond := &models.Session{Token: sessionEntity.Token}
	findCond := &models.Session{UserId: sessionEntity.UserId}
	// 2. 执行更新
	tx := sr.db.WithContext(ctx).Model(&models.Session{}).Where(findCond).Updates(updateCond)
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 返回结果
	return nil
}

/**
 * Is Session Valid
 */
func (sr *sessionRepoImpl) IsSessionValid(ctx context.Context, userId int64, token string) bool {
	session := &models.Session{}
	tx := sr.db.WithContext(ctx).Model(&models.Session{}).
		Where("user_id = ? AND token = ? AND expired_at > ?", userId, token, time.Now()).
		First(session)
	return tx.Error == nil && session.ID != 0
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
