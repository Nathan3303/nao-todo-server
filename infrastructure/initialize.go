package infrastructure

import (
	"naotodoserver/application"
	authApp "naotodoserver/application/auth"
	commentApp "naotodoserver/application/comment"
	eventApp "naotodoserver/application/event"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
	authService "naotodoserver/domain/auth/service"
	commentService "naotodoserver/domain/comment/service"
	eventService "naotodoserver/domain/checkitem/service"
	projectService "naotodoserver/domain/project/service"
	tagService "naotodoserver/domain/tag/service"
	taskService "naotodoserver/domain/task/service"
	userService "naotodoserver/domain/user/service"
	"naotodoserver/infrastructure/cron"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/sse"
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

// LoadLogger 初始化日志系统
func LoadLogger() {
	logging.InitLogger()
}

func LoadDBs() {
	dbs.InitMySQL()
	dbs.InitRedis()
	models.InitSnowflake(1)
}

func LoadDomains() {
	commentAppInst := commentApp.NewCommentApp(commentService.NewCommentDomain(
		commentRepo.NewCommentRepo(dbs.DB),
	))

	application.App = &application.Services{
		Auth: authApp.NewAuthApp(authService.GetAuthDomainImpl(
			authRepo.NewJWTRepo(),
			authRepo.NewUserRepo(dbs.DB, dbs.RdsCli),
			authRepo.NewSessionRepo(dbs.DB),
			authRepo.NewRateLimitRepo(dbs.RdsCli),
		)),
		User: userApp.NewUserApp(userService.NewUserDomain(
			userRepo.NewUserRepo(dbs.DB),
		), commentAppInst),
		Project: projectApp.NewProjectApp(projectService.NewProjectDomain(
			projectRepo.NewProjectRepo(dbs.DB),
			projectRepo.NewProjectPreferenceRepo(dbs.DB),
		)),
		Tag: tagApp.NewTagApp(tagService.NewTagDomain(
			tagRepo.NewTagRepo(dbs.DB),
			tagRepo.NewTagPreferenceRepo(dbs.DB),
		)),
		Task: taskApp.NewTaskApp(taskService.NewTaskDomain(
			taskRepo.NewTaskRepo(dbs.DB),
		)),
		Event: eventApp.NewEventApp(eventService.NewEventDomain(
			eventRepo.NewEventRepo(dbs.DB),
		)),
		Comment: commentAppInst,
	}
}

// WireSSE 装配 SSE Hub 的 SessionValidator
func WireSSE() {
	sessionRepo := authRepo.NewSessionRepo(dbs.DB)
	sse.GetHub().SessionValidator = func(userId int64, token string) bool {
		return sessionRepo.IsSessionValid(nil, userId, token)
	}
}

func LoadCron() {
	cronService := cron.GetCronServiceImpl()

	// 删除注销用户定时任务
	_, err := cronService.AddJob("0 2 * * *", cron.NewDeleteDeactivedUserJob(15))
	if err != nil {
		panic("删除注销用户定时任务添加失败：" + err.Error())
	}

	// 任务提醒扫描定时任务（每分钟）
	_, err = cronService.AddJob("* * * * *", cron.NewReminderJob())
	if err != nil {
		panic("任务提醒扫描定时任务添加失败：" + err.Error())
	}

	cronService.Start()
}
