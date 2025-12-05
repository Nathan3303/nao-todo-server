package container

import (
	authApp "naotodoserver/application/auth"
	projectApp "naotodoserver/application/project"
	userApp "naotodoserver/application/user"
	authService "naotodoserver/domain/auth/service"
	projectService "naotodoserver/domain/project/service"
	userService "naotodoserver/domain/user/service"
	authRepo "naotodoserver/infrastructure/persistence/auth"
	"naotodoserver/infrastructure/persistence/dbs"
	projectRepo "naotodoserver/infrastructure/persistence/project"
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
	projectApp.RegistDomainImpl(projectService.NewProjectDomain(
		projectRepo.NewProjectRepo(dbs.DB),
	))
}
