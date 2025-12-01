package container

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/domain/auth/service"
	"naotodoserver/infrastructure/auth"
	authRepo "naotodoserver/infrastructure/persistence/auth"
	"naotodoserver/infrastructure/persistence/dbs"
)

func LoadDomain() {
	jwtService := auth.GetJWTService()
	authUserRepo := authRepo.NewUserRepo(dbs.DB)
	authSessionRepo := authRepo.NewSessionRepo(dbs.DB)
	authDomainImpl := service.GetAuthDomainImpl(jwtService, &authUserRepo, &authSessionRepo)
	authApp.RegistDomain(&authDomainImpl)
}
