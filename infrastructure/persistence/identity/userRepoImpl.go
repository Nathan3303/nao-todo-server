package identity

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/persistence/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepoImpl struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) repositories.User {
	return &UserRepoImpl{db: db}
}

// CreateByVO 通过值对象创建用户（注册流程）
func (r *UserRepoImpl) CreateByVO(ctx context.Context, vo *valueobjects.CreateUser) (*entities.User, error) {
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
	r.db.WithContext(ctx).Create(&models.UserConfig{UserId: userModel.ID, Appearance: "auto"})
	return UserModel2Entity(userModel), nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepoImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	user := &models.User{}
	r.db.WithContext(ctx).Model(&models.User{}).Where(&models.User{Email: email}).First(user)
	if user.ID == 0 {
		return nil, errors.New("用户不存在")
	}
	return UserModel2Entity(user), nil
}

// FindById 根据ID查找用户
func (r *UserRepoImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	user := &models.User{}
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where(&models.User{ModelBase: models.ModelBase{ID: id}}).First(user).Error; err != nil {
		return nil, err
	}
	return UserModel2Entity(user), nil
}

// UpdateAvatar 更新头像
func (r *UserRepoImpl) UpdateAvatar(ctx context.Context, userId int64, avatarUrl string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Update("avatar", avatarUrl).Error
}

// UpdateNickname 更新昵称
func (r *UserRepoImpl) UpdateNickname(ctx context.Context, userId int64, nickname string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Update("nickname", nickname).Error
}

// UpdatePassword 更新密码
func (r *UserRepoImpl) UpdatePassword(ctx context.Context, userId int64, oldPassword, newPassword string) error {
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
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Update("password", string(hashed)).Error
}

// PasswordCompare 密码比对
func (r *UserRepoImpl) PasswordCompare(password, encryptedPassword []byte) bool {
	return bcrypt.CompareHashAndPassword(encryptedPassword, password) == nil
}

// Deactive 注销用户
func (r *UserRepoImpl) Deactive(ctx context.Context, userId int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).
		Update("deactived_at", &sql.NullTime{Time: time.Now(), Valid: true}).Error
}

// Active 激活用户
func (r *UserRepoImpl) Active(ctx context.Context, userId int64) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).
		Update("deactived_at", &sql.NullTime{Time: time.Time{}, Valid: false}).Error
}

// GetConfig 获取用户配置
func (r *UserRepoImpl) GetConfig(ctx context.Context, userId int64) (*entities.UserConfig, error) {
	config := &models.UserConfig{}
	if err := r.db.WithContext(ctx).Model(&models.UserConfig{}).Where("user_id = ?", userId).First(config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultCfg := &models.UserConfig{UserId: userId, Appearance: "auto"}
			r.db.WithContext(ctx).Create(defaultCfg)
			return UserConfigModel2Entity(defaultCfg), nil
		}
		return nil, err
	}
	return UserConfigModel2Entity(config), nil
}

// UpdateConfig 更新用户配置
func (r *UserRepoImpl) UpdateConfig(ctx context.Context, userId int64, appearance string) error {
	config := &models.UserConfig{}
	if err := r.db.WithContext(ctx).Model(&models.UserConfig{}).Where("user_id = ?", userId).First(config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(&models.UserConfig{UserId: userId, Appearance: appearance}).Error
		}
		return err
	}
	return r.db.WithContext(ctx).Model(config).Update("appearance", appearance).Error
}

// Delete 删除用户
func (r *UserRepoImpl) Delete(ctx context.Context, userId int64) error {
	r.db.WithContext(ctx).Model(&models.UserConfig{}).Where("user_id = ?", userId).Delete(&models.UserConfig{})
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userId).Delete(&models.User{}).Error
}

// DeleteDeactivatedUsers 删除已注销用户
func (r *UserRepoImpl) DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -1*int(dayOffset))
	var userIds []int64
	r.db.WithContext(ctx).Model(&models.User{}).Where("deactived_at < ?", cutoff).Pluck("id", &userIds)
	if len(userIds) > 0 {
		r.db.WithContext(ctx).Model(&models.UserConfig{}).Where("user_id IN ?", userIds).Delete(&models.UserConfig{})
	}
	tx := r.db.WithContext(ctx).Model(&models.User{}).Where("id IN ?", userIds).Delete(&models.User{})
	return tx.RowsAffected, tx.Error
}
