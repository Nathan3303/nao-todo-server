package infrastructure

import (
	authApp "naotodoserver/application/auth"
	commentApp "naotodoserver/application/comment"
	eventApp "naotodoserver/application/event"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
	authService "naotodoserver/domain/auth/service"
	commentService "naotodoserver/domain/comment/service"
	eventService "naotodoserver/domain/event/service"
	projectService "naotodoserver/domain/project/service"
	tagService "naotodoserver/domain/tag/service"
	taskService "naotodoserver/domain/task/service"
	userService "naotodoserver/domain/user/service"
	authRepo "naotodoserver/infrastructure/persistence/auth"
	commentRepo "naotodoserver/infrastructure/persistence/comment"
	"naotodoserver/infrastructure/persistence/dbs"
	eventRepo "naotodoserver/infrastructure/persistence/event"
	"naotodoserver/infrastructure/persistence/models"
	projectRepo "naotodoserver/infrastructure/persistence/project"
	tagRepo "naotodoserver/infrastructure/persistence/tag"
	taskRepo "naotodoserver/infrastructure/persistence/task"
	userRepo "naotodoserver/infrastructure/persistence/user"
)

func Init() {
	dbs.InitMySQL()
	dbs.InitRedis()
	models.InitSnowflake(1)
	loadDomain()
}

func loadDomain() {
	authApp.RegistDomainImpl(authService.GetAuthDomainImpl(
		authRepo.NewJWTRepo(),
		authRepo.NewUserRepo(dbs.DB, dbs.RdsCli),
		authRepo.NewSessionRepo(dbs.DB),
		authRepo.NewRateLimitRepo(dbs.RdsCli),
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
	eventApp.RegistDomainImpl(eventService.NewEventDomain(
		eventRepo.NewEventRepo(dbs.DB),
	))
	commentApp.RegistDomainImpl(commentService.NewCommentDomain(
		commentRepo.NewCommentRepo(dbs.DB),
	))
}
