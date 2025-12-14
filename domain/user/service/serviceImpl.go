package service

import (
	"context"
	"naotodoserver/domain/user/entities"
	"naotodoserver/domain/user/repositories"
)

func NewUserDomain(repo repositories.User) UserDomain {
	return &UserDomainImpl{
		repo: repo,
	}
}

/*
 * Create user
 * User 领域中的创建用户方法
 */
func (userDomain *UserDomainImpl) CreateUser(
	ctx context.Context,
	userEntity *entities.User,
) (*entities.User, error) {
	return userDomain.repo.Create(ctx, userEntity)
}

/*
 * Find user by id
 * User 领域中的根据 id 查询用户方法
 */
func (userDomain *UserDomainImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	return userDomain.repo.FindById(ctx, id)
}

/*
 * Update user avatar
 * 更新用户头像方法
 */
func (userDomain *UserDomainImpl) UpdateAvatar(ctx context.Context, userId int64, avatar string) error {
	return userDomain.repo.UpdateAvatar(ctx, userId, avatar)
}

/*
 * Update user nickname
 * 更新用户昵称方法
 */
func (userDomain *UserDomainImpl) UpdateNickname(ctx context.Context, userId int64, nickname string) error {
	return userDomain.repo.UpdateNickname(ctx, userId, nickname)
}

/*
 * Update user password
 * 更新用户密码方法
 */
func (userDomain *UserDomainImpl) UpdatePassword(
	ctx context.Context,
	userId int64,
	password string,
	newPassword string,
) error {
	return userDomain.repo.UpdatePassword(ctx, userId, password, newPassword)
}

/*
 * Delete user
 * 删除用户方法
 */
func (userDomain *UserDomainImpl) DeleteUser(ctx context.Context, userId int64) error {
	panic("unimplemented")
}

/*
 * Deactive user
 * 禁用用户方法
 */
func (userDomain *UserDomainImpl) Deactive(ctx context.Context, userId int64) error {
	return userDomain.repo.Deactive(ctx, userId)
}

/*
 * Active user
 * 激活用户方法
 */
func (userDomain *UserDomainImpl) Active(ctx context.Context, userId int64) error {
	return userDomain.repo.Active(ctx, userId)
}

/*
 * Password compare
 * 密码比较方法
 */
func (userDomain *UserDomainImpl) PasswordCompare(
	password, encryptedPassword []byte,
) bool {
	return userDomain.repo.PasswordCompare(password, encryptedPassword)
}
