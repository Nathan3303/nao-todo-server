package infrastructure

import (
	"naotodoserver/application"
	authApp "naotodoserver/application/auth"
	pomodoroApp "naotodoserver/application/pomodoro"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
	identityService "naotodoserver/domain/identity/service"
	pomodoroService "naotodoserver/domain/pomodoro/service"
	projectService "naotodoserver/domain/project/service"
	tagService "naotodoserver/domain/tag/service"
	taskService "naotodoserver/domain/task/service"
	"naotodoserver/infrastructure/cron"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/sse"
	authPkg "naotodoserver/infrastructure/persistence/auth"
	"naotodoserver/infrastructure/persistence/dbs"
	identityRepo "naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/infrastructure/persistence/models"
	pomodoroRepo "naotodoserver/infrastructure/persistence/pomodoro"
	projectRepo "naotodoserver/infrastructure/persistence/project"
	tagRepo "naotodoserver/infrastructure/persistence/tag"
	taskRepo "naotodoserver/infrastructure/persistence/task"
)

func LoadLogger() {
	logging.InitLogger()
}

func LoadDBs() {
	dbs.InitMySQL()
	dbs.InitRedis()
	models.InitSnowflake(1)
}

func LoadDomains() {
	identityDomain := identityService.NewIdentityDomain(
		authPkg.NewJWTRepo(),
		identityRepo.NewUserRepo(dbs.DB),
		authPkg.NewSessionRepo(dbs.DB),
		authPkg.NewRateLimitRepo(dbs.RdsCli),
	)

	taskAppInst := taskApp.NewTaskApp(taskService.NewTaskDomain(
		taskRepo.NewTaskRepo(dbs.DB),
	))

	pomodoroAppInst := pomodoroApp.NewPomodoroApp(pomodoroService.NewPomodoroDomain(
		pomodoroRepo.NewPomodoroRepo(dbs.DB),
	))

	application.App = &application.Services{
		Auth:    authApp.NewAuthApp(identityDomain),
		User:    userApp.NewUserApp(identityDomain, taskAppInst),
		Task:    taskAppInst,
		Project: projectApp.NewProjectApp(projectService.NewProjectDomain(
			projectRepo.NewProjectRepo(dbs.DB),
			projectRepo.NewProjectPreferenceRepo(dbs.DB),
		)),
		Tag: tagApp.NewTagApp(tagService.NewTagDomain(
			tagRepo.NewTagRepo(dbs.DB),
			tagRepo.NewTagPreferenceRepo(dbs.DB),
		)),
		Pomodoro: pomodoroAppInst,
	}
}

func WireSSE() {
	sessionRepo := authPkg.NewSessionRepo(dbs.DB)
	sse.GetHub().SessionValidator = func(userId int64, token string) bool {
		return sessionRepo.IsSessionValid(nil, userId, token)
	}
}

func LoadCron() {
	cronService := cron.GetCronServiceImpl()

	_, err := cronService.AddJob("0 2 * * *", cron.NewDeleteDeactivedUserJob(15))
	if err != nil {
		panic("删除注销用户定时任务添加失败：" + err.Error())
	}

	_, err = cronService.AddJob("* * * * *", cron.NewReminderJob())
	if err != nil {
		panic("任务提醒扫描定时任务添加失败：" + err.Error())
	}

	cronService.Start()
}
