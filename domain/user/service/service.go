package service

import (
	"context"
	"naotodoserver/domain/user/entities"
	"naotodoserver/domain/user/repositories"
)

type UserDomain interface {
	CreateUser(ctx context.Context, userEntity *entities.User) (*entities.User, error)
	FindById(ctx context.Context, id int64) (*entities.User, error)
	UpdateNickname(ctx context.Context, userId int64, nickname string) error
	UpdateAvatar(ctx context.Context, userId int64, avatar string) error
	UpdatePassword(ctx context.Context, userId int64, password string, newPassword string) error
	DeleteUser(ctx context.Context, userId int64) error
	Deactive(ctx context.Context, userId int64) error
	Active(ctx context.Context, userId int64) error
	PasswordCompare(password, encryptedPassword []byte) bool
	GetConfig(ctx context.Context, userId int64) (*entities.UserConfig, error)
	UpdateConfig(ctx context.Context, userId int64, appearance string) error
	DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) (int64, error)
}

type UserDomainImpl struct {
	repo repositories.User
}
