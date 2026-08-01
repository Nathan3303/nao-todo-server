package identity

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepoImpl struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewUserRepo(db *gorm.DB, c *cache.Cache) repositories.User {
	return &UserRepoImpl{db: db, cache: c}
}

// CreateByVO 通过值对象创建用户（注册流程）
func (r *UserRepoImpl) CreateByVO(
	ctx context.Context,
	vo *valueobjects.CreateUser,
) (*entities.User, error) {
	clientInfo := iCtx.GetClientInfo(ctx)
	userModel := &models.User{
		Account:     vo.Email,
		Email:       vo.Email,
		Password:    vo.EncryptedPassword,
		Nickname:    vo.Nickname,
		CreatedFrom: clientInfo.IPRegion + " " + clientInfo.DeviceType,
	}
	if err := r.db.WithContext(ctx).Create(userModel).Error; err != nil {
		return nil, err
	}
	// 创建默认配置
	r.db.WithContext(ctx).Create(
		&models.UserConfig{UserId: userModel.ID, Appearance: "auto"},
	)
	return UserModel2Entity(userModel), nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepoImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	user := &models.User{}
	r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where(&models.User{Email: email}).
		First(user)
	if user.ID == 0 {
		return nil, errors.New("用户不存在")
	}
	return UserModel2Entity(user), nil
}

// FindById 根据ID查找用户
func (r *UserRepoImpl) FindById(ctx context.Context, id types.UserID) (*entities.User, error) {
	// 先查缓存
	cached := &entities.User{}
	if r.cache.Get(ctx, cache.UserProfileKey(int64(id)), cached) {
		return cached, nil
	}
	user := &models.User{}
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where(&models.User{ModelBase: models.ModelBase{ID: int64(id)}}).
		First(user).Error; err != nil {
		return nil, err
	}
	result := UserModel2Entity(user)
	// 写入缓存
	r.cache.Set(ctx, cache.UserProfileKey(int64(id)), result, time.Minute*30)
	return result, nil
}

// UpdateAvatar 更新头像
func (r *UserRepoImpl) UpdateAvatar(
	ctx context.Context,
	userId types.UserID,
	avatarUrl string,
) error {
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", int64(userId)).
		Update("avatar", avatarUrl).Error; err != nil {
		return err
	}
	// 失效资料缓存
	r.cache.Del(ctx, cache.UserProfileKey(int64(userId)))
	return nil
}

// UpdateNickname 更新昵称
func (r *UserRepoImpl) UpdateNickname(
	ctx context.Context,
	userId types.UserID,
	nickname string,
) error {
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", int64(userId)).
		Update("nickname", nickname).Error; err != nil {
		return err
	}
	// 失效资料缓存
	r.cache.Del(ctx, cache.UserProfileKey(int64(userId)))
	return nil
}

// UpdatePassword 更新密码
func (r *UserRepoImpl) UpdatePassword(
	ctx context.Context,
	userId types.UserID,
	oldPassword string,
	newPassword string,
) error {
	user, err := r.FindById(ctx, userId)
	if err != nil {
		return err
	}
	if !r.PasswordCompare([]byte(oldPassword), []byte(user.Password)) {
		return errors.New("旧密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", int64(userId)).
		Update("password", string(hashed)).
		Error; err != nil {
		return err
	}
	// 失效资料缓存
	r.cache.Del(ctx, cache.UserProfileKey(int64(userId)))
	return nil
}

// PasswordCompare 密码比对
func (r *UserRepoImpl) PasswordCompare(password, encryptedPassword []byte) bool {
	return bcrypt.CompareHashAndPassword(encryptedPassword, password) == nil
}

// Deactive 注销用户
func (r *UserRepoImpl) Deactive(ctx context.Context, userId types.UserID) error {
	now := time.Now()
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", int64(userId)).
		Updates(map[string]interface{}{
			"deactived_at":          &sql.NullTime{Time: now, Valid: true},
			"last_cancel_restore_at": &sql.NullTime{Time: now, Valid: true},
		}).
		Error; err != nil {
		return err
	}
	// 失效资料与配置缓存
	r.cache.Del(ctx, cache.UserProfileKey(int64(userId)), cache.UserConfigKey(int64(userId)))
	return nil
}

// Active 激活用户
func (r *UserRepoImpl) Active(ctx context.Context, userId types.UserID) error {
	now := time.Now()
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", int64(userId)).
		Updates(map[string]interface{}{
			"deactived_at":          &sql.NullTime{Time: time.Time{}, Valid: false},
			"last_cancel_restore_at": &sql.NullTime{Time: now, Valid: true},
		}).
		Error; err != nil {
		return err
	}
	// 失效资料与配置缓存
	r.cache.Del(ctx, cache.UserProfileKey(int64(userId)), cache.UserConfigKey(int64(userId)))
	return nil
}

// GetConfig 获取用户配置
func (r *UserRepoImpl) GetConfig(
	ctx context.Context,
	userId types.UserID,
) (*entities.UserConfig, error) {
	// 先查缓存
	cached := &entities.UserConfig{}
	if r.cache.Get(ctx, cache.UserConfigKey(int64(userId)), cached) {
		return cached, nil
	}
	config := &models.UserConfig{}
	err := r.db.
		WithContext(ctx).
		Model(&models.UserConfig{}).
		Where("user_id = ?", int64(userId)).
		First(config).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultCfg := &models.UserConfig{UserId: int64(userId), Appearance: "auto"}
			r.db.WithContext(ctx).Create(defaultCfg)
			result := UserConfigModel2Entity(defaultCfg)
			// 写入缓存
			r.cache.Set(ctx, cache.UserConfigKey(int64(userId)), result, time.Minute*30)
			return result, nil
		}
		return nil, err
	}
	result := UserConfigModel2Entity(config)
	// 写入缓存
	r.cache.Set(ctx, cache.UserConfigKey(int64(userId)), result, time.Minute*30)
	return result, nil
}

// UpdateConfig 更新用户配置
func (r *UserRepoImpl) UpdateConfig(
	ctx context.Context,
	userId types.UserID,
	appearance string,
) error {
	config := &models.UserConfig{}
	err := r.db.
		WithContext(ctx).
		Model(&models.UserConfig{}).
		Where("user_id = ?", int64(userId)).
		First(config).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.db.
				WithContext(ctx).
				Create(&models.UserConfig{UserId: int64(userId), Appearance: appearance}).
				Error; err != nil {
				return err
			}
			// 失效配置缓存
			r.cache.Del(ctx, cache.UserConfigKey(int64(userId)))
			return nil
		}
		return err
	}
	if err := r.db.
		WithContext(ctx).
		Model(config).
		Update("appearance", appearance).
		Error; err != nil {
		return err
	}
	// 失效配置缓存
	r.cache.Del(ctx, cache.UserConfigKey(int64(userId)))
	return nil
}

// Delete 删除用户
func (r *UserRepoImpl) Delete(ctx context.Context, userId types.UserID) error {
	r.db.
		WithContext(ctx).
		Model(&models.UserConfig{}).
		Where("user_id = ?", int64(userId)).
		Delete(&models.UserConfig{})
	if err := r.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", int64(userId)).
		Delete(&models.User{}).
		Error; err != nil {
		return err
	}
	// 失效资料与配置缓存
	r.cache.Del(ctx, cache.UserProfileKey(int64(userId)), cache.UserConfigKey(int64(userId)))
	return nil
}

// DeleteDeactivatedUsers 删除已注销用户
// 用户本体与其 11 张关联表的删除在同一事务内完成，任一步失败即整体回滚，避免产生孤儿数据。
// 缓存失效在事务提交成功之后执行，回滚时不会误清空缓存。
// @param ctx 上下文
// @param dayOffset 注销天数偏移，注销时间早于当前时间减去该天数的用户将被删除
// @return int64 删除的用户行数
// @return error 错误
func (r *UserRepoImpl) DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -1*int(dayOffset))
	var userIds []int64
	var rowsAffected int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&models.User{}).
			Where("deactived_at < ?", cutoff).
			Pluck("id", &userIds).
			Error; err != nil {
			return err
		}
		// 无待删除用户时提前返回，避免空 IN 子句
		if len(userIds) == 0 {
			return nil
		}
		if err := tx.
			Model(&models.UserSession{}).
			Where("user_id IN ?", userIds).
			Delete(&models.UserSession{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.TaskComment{}).
			Where("user_id IN ?", userIds).
			Delete(&models.TaskComment{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.TaskCheckItem{}).
			Where("user_id IN ?", userIds).
			Delete(&models.TaskCheckItem{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.Task{}).
			Where("user_id IN ?", userIds).
			Delete(&models.Task{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.ProjectPreference{}).
			Where("user_id IN ?", userIds).
			Delete(&models.ProjectPreference{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.Project{}).
			Where("user_id IN ?", userIds).
			Delete(&models.Project{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.TagPreference{}).
			Where("user_id IN ?", userIds).
			Delete(&models.TagPreference{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.Tag{}).
			Where("user_id IN ?", userIds).
			Delete(&models.Tag{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.PomodoroRecord{}).
			Where("user_id IN ?", userIds).
			Delete(&models.PomodoroRecord{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.Pomodoro{}).
			Where("user_id IN ?", userIds).
			Delete(&models.Pomodoro{}).
			Error; err != nil {
			return err
		}
		if err := tx.
			Model(&models.UserConfig{}).
			Where("user_id IN ?", userIds).
			Delete(&models.UserConfig{}).
			Error; err != nil {
			return err
		}
		deleteTx := tx.
			Model(&models.User{}).
			Where("id IN ?", userIds).
			Delete(&models.User{})
		if deleteTx.Error != nil {
			return deleteTx.Error
		}
		rowsAffected = deleteTx.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	// 事务提交成功后失效资料与配置缓存
	for _, userId := range userIds {
		r.cache.Del(ctx, cache.UserProfileKey(userId), cache.UserConfigKey(userId))
	}
	return rowsAffected, nil
}
