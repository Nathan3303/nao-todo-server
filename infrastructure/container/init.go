package container

import (
	authApp "naotodoserver/application/auth"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
	authService "naotodoserver/domain/auth/service"
	projectService "naotodoserver/domain/project/service"
	tagService "naotodoserver/domain/tag/service"
	taskService "naotodoserver/domain/task/service"
	userService "naotodoserver/domain/user/service"
	authRepo "naotodoserver/infrastructure/persistence/auth"
	"naotodoserver/infrastructure/persistence/dbs"
	projectRepo "naotodoserver/infrastructure/persistence/project"
	tagRepo "naotodoserver/infrastructure/persistence/tag"
	taskRepo "naotodoserver/infrastructure/persistence/task"
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
	tagApp.RegistDomainImpl(tagService.NewTagDomain(
		tagRepo.NewTagRepo(dbs.DB),
	))
	taskApp.RegistDomainImpl(taskService.NewTaskDomain(
		taskRepo.NewTaskRepo(dbs.DB),
	))
}
