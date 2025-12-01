package container

import (
	authApp "naotodoserver/application/auth"
	userApp "naotodoserver/application/user"
	authService "naotodoserver/domain/auth/service"
	userService "naotodoserver/domain/user/service"
	authRepo "naotodoserver/infrastructure/persistence/auth"
	"naotodoserver/infrastructure/persistence/dbs"
	userRepo "naotodoserver/infrastructure/persistence/user"
)

func LoadDomain() {
	authApp.RegistDomainImpl(authService.GetAuthDomainImpl(
		authRepo.NewJWTRepo(),
		authRepo.NewUserRepo(dbs.DB),
		authRepo.NewSessionRepo(dbs.DB),
	))
	userApp.RegistDomainImpl(userService.NewUserDomain(
		userRepo.NewUserRepo(dbs.DB),
	))
}
