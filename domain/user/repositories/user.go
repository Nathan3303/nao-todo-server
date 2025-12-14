package repositories

import (
	"context"
	"naotodoserver/domain/user/entities"
)

type User interface {
	Create(ctx context.Context, userEntity *entities.User) (*entities.User, error)
	FindById(ctx context.Context, id int64) (*entities.User, error)
	UpdateNickname(ctx context.Context, userId int64, nickname string) error
	UpdateAvatar(ctx context.Context, userId int64, avatarUrl string) error
	UpdatePassword(ctx context.Context, userId int64, password string, newPassword string) error
	PasswordCompare(password, encryptedPassword []byte) bool
	Deactive(ctx context.Context, userId int64) error
	Active(ctx context.Context, userId int64) error
	Delete(ctx context.Context, whereEntity *entities.User) error
}
