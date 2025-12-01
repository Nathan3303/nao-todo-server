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

// CreateUser implements UserDomain.
func (userDomain *UserDomainImpl) CreateUser(
	ctx context.Context,
	userEntity *entities.User,
) (*entities.User, error) {
	return userDomain.repo.Create(ctx, userEntity)
}

// FindById implements UserDomain.
func (userDomain *UserDomainImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	panic("unimplemented")
}

// UpdateAvatar implements UserDomain.
func (userDomain *UserDomainImpl) UpdateAvatar(ctx context.Context, userId int64, avatar string) error {
	panic("unimplemented")
}

// UpdateNickname implements UserDomain.
func (userDomain *UserDomainImpl) UpdateNickname(ctx context.Context, userId int64, nickname string) error {
	panic("unimplemented")
}

// UpdatePassword implements UserDomain.
func (userDomain *UserDomainImpl) UpdatePassword(
	ctx context.Context,
	userId int64,
	password string,
	newPassword string,
) error {
	return userDomain.repo.UpdatePassword(ctx, userId, password, newPassword)
}

// DeleteUser implements UserDomain.
func (userDomain *UserDomainImpl) DeleteUser(ctx context.Context, userId int64) error {
	panic("unimplemented")
}
