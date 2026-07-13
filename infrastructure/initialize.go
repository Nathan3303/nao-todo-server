package infrastructure

import (
	"context"
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
	"naotodoserver/infrastructure/persistence/dbs"
	identityRepo "naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/infrastructure/persistence/models"
	pomodoroRepo "naotodoserver/infrastructure/persistence/pomodoro"
	projectRepo "naotodoserver/infrastructure/persistence/project"
	tagRepo "naotodoserver/infrastructure/persistence/tag"
	taskRepo "naotodoserver/infrastructure/persistence/task"
	"naotodoserver/infrastructure/sse"
)

// LoadLogger 初始化日志记录器
func LoadLogger() {
	logging.InitLogger()
}

// LoadDBs 初始化数据库连接
func LoadDBs() {
	dbs.InitMySQL()
	dbs.InitRedis()
	models.InitSnowflake(1)
}

// LoadDomains 初始化领域模型
func LoadDomains() {
	// 初始化身份领域模型
	userRepoInst := identityRepo.NewUserRepo(dbs.DB)
	sessionRepoInst := identityRepo.NewSessionRepo(dbs.DB)
	identityDomain := identityService.NewIdentityDomain(
		identityRepo.NewJWTRepo(),
		sessionRepoInst,
		identityRepo.NewRateLimitRepo(dbs.RdsCli),
	)
	// 初始化任务领域模型
	taskRepoInst := taskRepo.NewTaskRepo(dbs.DB)
	taskDomain := taskService.NewTaskDomain(taskRepoInst)
	taskAppInst := taskApp.NewTaskApp(taskDomain, taskRepoInst)
	// 初始化番茄领域模型
	pomodoroRecordRepoInst := pomodoroRepo.NewPomodoroRecordRepo(dbs.DB)
	pomodoroRepoInst := pomodoroRepo.NewPomodoroRepo(dbs.DB)
	pomodoroDomain := pomodoroService.NewPomodoroDomain(pomodoroRecordRepoInst, pomodoroRepoInst)
	pomodoroAppInst := pomodoroApp.NewPomodoroApp(
		pomodoroDomain,
		pomodoroRecordRepoInst,
		pomodoroRepoInst,
	)
	// 初始化项目领域模型
	projectRepoInst := projectRepo.NewProjectRepo(dbs.DB)
	projectPreferenceRepoInst := projectRepo.NewProjectPreferenceRepo(dbs.DB)
	projectDomain := projectService.NewProjectDomain(projectRepoInst, projectPreferenceRepoInst)
	projectAppInst := projectApp.NewProjectApp(
		projectDomain,
		projectRepoInst,
		projectPreferenceRepoInst,
	)
	// 初始化标签领域模型
	tagRepoInst := tagRepo.NewTagRepo(dbs.DB)
	tagPreferenceRepoInst := tagRepo.NewTagPreferenceRepo(dbs.DB)
	tagDomain := tagService.NewTagDomain(tagRepoInst, tagPreferenceRepoInst)
	tagAppInst := tagApp.NewTagApp(tagDomain, tagRepoInst, tagPreferenceRepoInst)
	// 初始化项目领域模型
	application.App = &application.Services{
		Auth:          authApp.NewAuthApp(identityDomain, userRepoInst, sessionRepoInst),
		User:          userApp.NewUserApp(userRepoInst, taskAppInst),
		Task:          taskAppInst,
		TaskComment:   taskAppInst,
		TaskCheckItem: taskAppInst,
		Project:       projectAppInst,
		Tag:           tagAppInst,
		Pomodoro:      pomodoroAppInst,
	}
}

// WireSSE 配置 SSE 会话验证
func WireSSE() {
	sessionRepo := identityRepo.NewSessionRepo(dbs.DB)
	sse.GetHub().SessionValidator = func(userId int64, token string) bool {
		return sessionRepo.IsSessionValid(context.TODO(), userId, token)
	}
}

// LoadCron 初始化定时任务
func LoadCron() {
	// 初始化定时任务
	cronService := cron.GetCronServiceImpl()
	// 添加定时任务 - 删除注销用户
	_, err := cronService.AddJob(
		"0 2 * * *",
		cron.NewDeleteDeactivedUserJob(15),
	)
	if err != nil {
		panic("删除注销用户定时任务添加失败：" + err.Error())
	}
	// 添加定时任务 - 任务提醒扫描
	_, err = cronService.AddJob("* * * * *", cron.NewReminderJob())
	if err != nil {
		panic("任务提醒扫描定时任务添加失败：" + err.Error())
	}
	// 启动定时任务
	cronService.Start()
}
