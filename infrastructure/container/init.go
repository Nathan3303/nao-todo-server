package container

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/domain/auth/service"
	authRepo "naotodoserver/infrastructure/persistence/auth"
	"naotodoserver/infrastructure/persistence/dbs"
)

func LoadDomain() {
	authApp.RegistDomainImpl(service.GetAuthDomainImpl(
		authRepo.NewJWTRepo(),
		authRepo.NewUserRepo(dbs.DB),
		authRepo.NewSessionRepo(dbs.DB),
	))
}
