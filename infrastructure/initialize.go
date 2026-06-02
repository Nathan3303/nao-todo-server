package infrastructure

import (
	"naotodoserver/application"
	authApp "naotodoserver/application/auth"
	commentApp "naotodoserver/application/comment"
	checkitemApp "naotodoserver/application/checkitem"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
	identityService "naotodoserver/domain/identity/service"
	commentService "naotodoserver/domain/comment/service"
	checkitemService "naotodoserver/domain/checkitem/service"
	projectService "naotodoserver/domain/project/service"
	tagService "naotodoserver/domain/tag/service"
	taskService "naotodoserver/domain/task/service"
	"naotodoserver/infrastructure/cron"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/sse"
	authPkg "naotodoserver/infrastructure/persistence/auth"
	commentRepo "naotodoserver/infrastructure/persistence/comment"
	"naotodoserver/infrastructure/persistence/dbs"
	checkitemRepo "naotodoserver/infrastructure/persistence/checkitem"
	identityRepo "naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/infrastructure/persistence/models"
	projectRepo "naotodoserver/infrastructure/persistence/project"
	tagRepo "naotodoserver/infrastructure/persistence/tag"
	taskRepo "naotodoserver/infrastructure/persistence/task"
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

	identityDomain := identityService.NewIdentityDomain(
		authPkg.NewJWTRepo(),
		identityRepo.NewUserRepo(dbs.DB),
		authPkg.NewSessionRepo(dbs.DB),
		authPkg.NewRateLimitRepo(dbs.RdsCli),
	)

	application.App = &application.Services{
		Auth:    authApp.NewAuthApp(identityDomain),
		User:    userApp.NewUserApp(identityDomain, commentAppInst),
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
		CheckItem: checkitemApp.NewCheckItemApp(checkitemService.NewCheckItemDomain(
			checkitemRepo.NewCheckItemRepo(dbs.DB),
		)),
		Comment: commentAppInst,
	}
}

// WireSSE 装配 SSE Hub 的 SessionValidator
func WireSSE() {
	sessionRepo := authPkg.NewSessionRepo(dbs.DB)
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
