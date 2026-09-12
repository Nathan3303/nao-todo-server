package infrastructure

import (
	"context"
	"naotodoserver/application"
	authApp "naotodoserver/application/auth"
	counts "naotodoserver/application/counts"
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
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/cron"
	"naotodoserver/infrastructure/events"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/dbs"
	identityRepo "naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/infrastructure/persistence/models"
	pomodoroRepo "naotodoserver/infrastructure/persistence/pomodoro"
	projectRepo "naotodoserver/infrastructure/persistence/project"
	tagRepo "naotodoserver/infrastructure/persistence/tag"
	taskRepo "naotodoserver/infrastructure/persistence/task"
	"naotodoserver/infrastructure/sse"
	"naotodoserver/infrastructure/storage"
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
func LoadDomains() *application.Services {
	// 初始化缓存辅助组件
	cacheInst := cache.NewCache(dbs.RdsCli)
	// 初始化事务管理器
	txManager := dbs.NewTxManager(dbs.DB)
	// 初始化身份领域模型
	userRepoInst := identityRepo.NewUserRepo(dbs.DB, cacheInst)
	sessionRepoInst := identityRepo.NewSessionRepo(dbs.DB, cacheInst)
	identityDomain := identityService.NewIdentityDomain(
		identityRepo.NewJWTRepo(),
		sessionRepoInst,
		identityRepo.NewRateLimitRepo(dbs.RdsCli),
	)
	// 初始化通知发布端口(SSE Hub 适配)
	notificationPublisher := sse.NewNotificationPublisher()
	// 初始化任务领域模型
	taskRepoInst := taskRepo.NewTaskRepo(dbs.DB)
	taskDomain := taskService.NewTaskDomain(taskRepoInst, taskRepoInst)
	// 初始化项目领域模型（先于任务应用装配：计数订阅者需跨仓注入 taskRepo + projectRepo）
	projectRepoInst := projectRepo.NewProjectRepo(dbs.DB, cacheInst)
	projectPreferenceRepoInst := projectRepo.NewProjectPreferenceRepo(dbs.DB)
	// 初始化计数事件总线 + 订阅者（领域统计属性联动，ADR 2026-09-12 §4.1/§4.2）
	countBus := events.NewInMemoryBus()
	countUpdater := counts.NewCountUpdater(taskRepoInst, projectRepoInst)
	countBus.Subscribe(countUpdater.HandleCountEvent)
	taskAppInst := taskApp.NewTaskApp(
		taskDomain,
		taskRepoInst,
		taskRepoInst,
		taskRepoInst,
		notificationPublisher,
		txManager,
		countBus,
	)
	// 初始化番茄领域模型
	pomodoroRecordRepoInst := pomodoroRepo.NewPomodoroRecordRepo(dbs.DB)
	pomodoroRepoInst := pomodoroRepo.NewPomodoroRepo(dbs.DB)
	pomodoroDomain := pomodoroService.NewPomodoroDomain(
		pomodoroRecordRepoInst,
		pomodoroRepoInst,
	)
	pomodoroAppInst := pomodoroApp.NewPomodoroApp(
		pomodoroDomain,
		pomodoroRecordRepoInst,
		pomodoroRepoInst,
	)
	projectDomain := projectService.NewProjectDomain(
		projectRepoInst,
		projectPreferenceRepoInst,
	)
	projectAppInst := projectApp.NewProjectApp(
		projectDomain,
		txManager,
		projectRepoInst,
		projectPreferenceRepoInst,
		taskRepoInst, // 注入 Task 仓库，用于级联操作
		countBus,     // 注入计数事件总线（E7 级联重算）
	)
	// 初始化标签领域模型
	tagRepoInst := tagRepo.NewTagRepo(dbs.DB, cacheInst)
	tagPreferenceRepoInst := tagRepo.NewTagPreferenceRepo(dbs.DB)
	tagDomain := tagService.NewTagDomain(
		tagRepoInst,
		tagPreferenceRepoInst,
	)
	tagAppInst := tagApp.NewTagApp(
		tagDomain,
		tagRepoInst,
		tagPreferenceRepoInst,
		txManager,
		taskDomain, // 注入 Task 领域服务，用于删除标签时级联清理任务引用
	)
	// 初始化项目领域模型
	return &application.Services{
		Auth: authApp.NewAuthApp(
			identityDomain,
			userRepoInst,
			sessionRepoInst,
		),
		User: userApp.NewUserApp(
			userRepoInst,
			sessionRepoInst,
			taskAppInst,
			storage.NewAvatarStorage(),
		),
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
	sessionRepo := identityRepo.NewSessionRepo(dbs.DB, cache.NewCache(dbs.RdsCli))
	sse.GetHub().SessionValidator = func(userId int64, token string) bool {
		return sessionRepo.IsSessionValid(context.TODO(), types.UserID(userId), token)
	}
}

// LoadCron 初始化定时任务
func LoadCron(svc *application.Services) {
	// 初始化定时任务
	cronService := cron.GetCronServiceImpl()
	// 添加定时任务 - 删除注销用户
	_, err := cronService.AddJob(
		"0 4 * * *",
		cron.NewDeleteDeactivedUserJob(7, svc.User),
	)
	if err != nil {
		panic("删除注销用户定时任务添加失败：" + err.Error())
	}
	// 添加定时任务 - 任务提醒扫描
	_, err = cronService.AddJob("* * * * *", cron.NewReminderJob(svc.Task))
	if err != nil {
		panic("任务提醒扫描定时任务添加失败：" + err.Error())
	}
	// 启动定时任务
	cronService.Start()
}
