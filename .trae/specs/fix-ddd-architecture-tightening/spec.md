# Spec: 切断 application → infrastructure 依赖并补齐 Task 实体充血化

## 背景

两轮 DDD 整改后,项目已显著改善:

* 领域枚举/常量以 DB 为准(`task_enums.go`)

* 事务方案走 `context + TxManager` 端口

* 四域 DTO 与 application 自建 DTO 解耦(`application/<domain>/dto/`)

* Project/Task/TaskCheckItem 实体已充血化并接入单条生产路径

但仍有以下 DDD 契合度问题未处理:

### 1. application 反向依赖 infrastructure(架构硬伤)

5 个 appImpl 均直接 import `naotodoserver/infrastructure/context` 提取 userId,`task/appImpl.go` 还直接 import `naotodoserver/infrastructure/sse` 调用 `hub.Publish`。这违反"依赖方向单向向内"原则,DDD 战术上不可接受。

### 2. Task 实体方法仍是死代码

`Task.Archive()/Unarchive()/ToggleStar()` 仅出现在 `task_test.go` 中。生产路径 `UpdateTask` 对 `archivedAt/starMarkAt/givenUpAt` 仍走 `*string → NewNullableTimeByTimeStrPtr` 路径,实体状态迁移语义被绕过。

### 3. Task 实体不变量未生效

`Task.IsDatesValid()` 定义"归档/收藏/放弃时间必须晚于开始时间"规则,但全仓零调用点。`UpdateTask.Validate()` 只校验 end/start,该规则在写路径实际未生效。

### 4. Task 级联批量操作绕过实体

`infrastructure/persistence/task/repoImpl.go` 的 `SoftDeleteByProjectId/RestoreByProjectId/ArchiveByProjectId/UnarchiveByProjectId` 直接 `Updates(map)` 改 `archived_at/deleted_at`,与单条路径充血化后语义不一致。

## 目标

1. **切断 application → infrastructure**:userId 改为 App 接口显式参数,SSE 推送改用 `NotificationPublisher` 端口。
2. **Task 实体 Archive/Unarchive/ToggleStar 接入生产路径**:仅当 req 显式携带 archivedAt/starMarkAt/givenUpAt 时走读-改-写,空时保持原行为,零破坏。
3. **Task 不变量 IsDatesValid 挂到写路径**:`UpdateTask` 提交前对完整实体的"时间不变量"做最后校验。
4. **Task 级联批量操作内部走实体方法**:保持方法在 infra,但实现里先加载实体、调方法、再批量 Updates,语义与单条路径一致。

## 范围

### In scope

* 5 个 application appImpl + 5 个 app.go(改 userId 显式参数签名)

* 5 个 application appImpl 构造函数(去掉对 infrastructure/context 的 import)

* `infrastructure/context/userId.go` 仍保留,但仅供 controller 注入调用

* 全部 interfaces/controllers/\*.go(传入 userId 给 app)

* `domain/types/notification_publisher.go`(新增端口)

* `infrastructure/sse/notificationPublisher.go`(新增实现)

* `application/task/appImpl.go` `ProcessReminders`(改用端口)

* `infrastructure/cron/reminder.go` 构造 wiring(注入 publisher)

* `infrastructure/initialize.go` wiring

* `application/task/converters.go` `UpdateTaskReqToValueObject`(为挂 IsDatesValid 与读-改-写做适配)

* `domain/task/entities/task.go` 调整 `Archive/Unarchive` 签名:接受 nullable time 参数而非总是 time.Now,支持复用 req 字段

* `infrastructure/persistence/task/repoImpl.go` 4 个 ByProjectId 批量方法(内部调实体方法)

* `domain/task/entities/task_test.go` 补充 `Archive/Unarchive` 可控时间版本的测试

### Out of scope

* 子任务级联删除规则(已在 P1 风险表中,但属业务决策,需先与 PM 确认级联删 or 拦截)

* Tag/Pomodoro/Identity 领域服务透传式代理消除(优先级 P2)

* 删 Project 业务的"停用 N 天"规则上移(优先级 P2)

* application/user/avatarStorage 端口位置倒挂(优先级 P2,放下一轮)

## 设计决策(已与用户确认)

| 决策点            | 选择                                                                                     |
| -------------- | -------------------------------------------------------------------------------------- |
| userId 注入      | **App 接口签名改 userId 显式参数**(所有 app 接口签名同步改,controller 在调用前用 iCtx.GetUserId 提取后传入)        |
| SSE 推送         | **Application 引入 NotificationPublisher 接口,infrastructure/sse 实现**(端口放 `domain/types/`) |
| Task 实体充血化触发条件 | **仅当 req 显式携带时才走读-改-写**(req 字段为 nil 时保持原行为)                                            |
| Task 级联批量充血化位置 | **保持批量方法在 infra,内部加载实体调方法**(实现简单,与"实体方法在 domain"互补)                                    |
| 代码风格           | **严格遵循项目当前代码风格**(Chinese 注释、gofmt、CRLF 保留)                                             |

## 实施计划

### Part A:切断 application → infrastructure(架构修复)

#### Task 1: 引入 `NotificationPublisher` 端口

新增 `domain/types/notification_publisher.go`:

```go
// NotificationPublisher 通知发布端口
// 由 infrastructure/sse 实现,供 application 调用推送领域事件
type NotificationPublisher interface {
    // PublishReminder 推送任务提醒事件
    PublishReminder(ctx context.Context, userId int64, event ReminderEvent) error
}

// ReminderEvent 提醒事件载荷
type ReminderEvent struct {
    Type        string
    TaskId      string
    TaskName    string
    Description string
    RemindAt    string
}
```

新增 `infrastructure/sse/notificationPublisher.go`:

```go
// notificationPublisherImpl NotificationPublisher 实现(SSE Hub 适配)
type notificationPublisherImpl struct{ hub *Hub }

func NewNotificationPublisher() types.NotificationPublisher {
    return &notificationPublisherImpl{hub: GetHub()}
}

func (p *notificationPublisherImpl) PublishReminder(
    ctx context.Context, userId int64, event types.ReminderEvent,
) error {
    p.hub.Publish(userId, sse.ReminderEvent(event))
    return nil
}
```

保留 `sse.ReminderEvent`(SSE 层自己的 JSON tag 结构),`types.ReminderEvent` 是无 tag 的领域载荷;`sse.ReminderEvent(event)` 是值转换(sse.ReminderEvent 与 types.ReminderEvent 字段完全一致,做一次值拷贝即可)。

#### Task 2: 5 域 appImpl 改 userId 显式参数

以 task 为例,接口改动:

```go
// 旧
GetTaskById(ctx context.Context, taskId string, includeDeleted bool) (*dto.GetTaskRes, error)
CreateTask(ctx context.Context, req *dto.CreateTaskReq) (*dto.GetTaskRes, error)
// ...

// 新
GetTaskById(ctx context.Context, userId int64, taskId string, includeDeleted bool) (*dto.GetTaskRes, error)
CreateTask(ctx context.Context, userId int64, req *dto.CreateTaskReq) (*dto.GetTaskRes, error)
// ...
```

每个 appImpl:

* 删除 `import iCtx "naotodoserver/infrastructure/context"`

* 删除每个方法体开头"获取用户 ID"的几行(`userId := iCtx.GetUserId(ctx); if userId <= 0 { return ...ErrInvalidUserID }`)

* userId 改用入参,值由 controller 注入(在 controller 用 `iCtx.GetUserId(c)` 取出后再调 app)

* task appImpl 同步去掉 `import "naotodoserver/infrastructure/sse"`,改用 `types.NotificationPublisher`

构造函数改动:

* `NewTaskApp(taskDomain, taskRepo, checkItemRepo, commentRepo, publisher types.NotificationPublisher)`

#### Task 3: interfaces/controllers 适配 userId 注入

所有 controller 在调 app 前显式 `userId := iCtx.GetUserId(c)`(保留 `infrastructure/context/userId.go`,仅 controller 一处依赖),失败时返回 `responser.Fail(c, ...)`(遵循项目既有错误响应风格)。

#### Task 4: infrastructure/initialize.go wiring

* `infrastructure/cron/reminder.go` 构造时注入 `publisher`

* `initialize.go` 创建 `notificationPublisher`,注入到 `taskApp`

### Part B:Task 实体充血化补齐

#### Task 5: 改造 `Task.Archive/Unarchive` 接受时间参数

为支持读-改-写复用 req 字段,实体方法改为接受 `*string`(HTTP 时间字符串)与 `*time.Time` 两种内部重载,采用接受 `types.NullableTime` 的统一签名:

```go
// Archive 归档任务(可指定归档时间,nil 表示使用当前时间)
func (task *Task) Archive(at *time.Time) {
    if at == nil {
        now := time.Now()
        task.ArchivedAt = types.NewNullableTimeByTime(now)
        return
    }
    task.ArchivedAt = types.NewNullableTimeByTime(*at)
}

// Unarchive 取消归档任务
func (task *Task) Unarchive() {
    task.ArchivedAt = types.NewNullableTimeNull()
}

// ToggleStar 切换收藏状态(可指定收藏时间,nil 表示使用当前时间)
func (task *Task) ToggleStar(at *time.Time) {
    if at == nil {
        now := time.Now()
        task.StarMarkAt = types.NewNullableTimeByTime(now)
        return
    }
    task.StarMarkAt = types.NewNullableTimeByTime(*at)
}
```

更新 `projectEntity.Archive/Unarchive` 暂不动(Project 在 serviceImpl 中已用 `time.Now()` 调用,本次不在改造范围)。Task 实体测试同步更新以匹配新签名。

#### Task 6: UpdateTask 读-改-写改造

[application/task/appImpl.go UpdateTask](file:///c:/Users/LEE19/Projects/nao-todo-server/application/task/appImpl.go#L100-L144) 增加读-改-写分支,仅在 `req.ArchivedAt/StarMarkAt/GivenUpAt != nil` 时触发。读-改-写完成后将实体新值回写 `updateTaskValueObject`(替换原 `req.ArchivedAt` 字符串),保持与 `Update` 共用 update 字段。

具体规则:

* req.ArchivedAt != nil:加载实体,按 req 解析时间后调 `task.Archive(parsedTime)`,将 taskEntity.ArchivedAt 回写 vo

* req.StarMarkAt != nil:加载实体,按 req 解析后调 `task.ToggleStar(parsedTime)`,回写 vo

* req.GivenUpAt != nil:加载实体,按 req 解析后清空/设值(新增 `Task.GiveUp(at *time.Time)` 与 `Task.UngiveUp()`),回写 vo

* 任一分支触发读-改-写后,在 Update 前调 `taskEntity.IsDatesValid()` 校验不变量

注意:`req.ArchivedAt == nil` 表示"不动该字段",不视为"清空"。读-改-写与现有 `req.State` 状态机分支独立,可并存。

#### Task 7: Task 实体新增 GiveUp/UngiveUp

为支持 GivenUpAt 字段充血化,在 `Task` 实体新增:

```go
// GiveUp 放弃任务
func (task *Task) GiveUp(at *time.Time) {
    if at == nil {
        now := time.Now()
        task.GivenUpAt = types.NewNullableTimeByTime(now)
        return
    }
    task.GivenUpAt = types.NewNullableTimeByTime(*at)
}

// UngiveUp 取消放弃任务
func (task *Task) UngiveUp() {
    task.GivenUpAt = types.NewNullableTimeNull()
}
```

#### Task 8: Task 实体不变量挂到写路径

`application/task/appImpl.go UpdateTask` 在调用 `taskRepo.Update` 前,若已加载过 taskEntity(状态机或读-改-写分支触发过)则调 `taskEntity.IsDatesValid()`,失败返回 `errors.New("时间参数无效 - ...")`。

#### Task 9: infrastructure/persistence/task/repoImpl.go 4 个 ByProjectId 充血化

[repoImpl.go L641-L684](file:///c:/Users/LEE19/Projects/nao-todo-server/infrastructure/persistence/task/repoImpl.go#L641-L684) 的 4 个批量方法改造:

* `SoftDeleteByProjectId`:先 `Find` 该 user/project 下所有未删除 task,逐个加载实体调 `task.Delete()`(软删,设置 DeletedAt),收集要更新的行,最后批量 `Updates` 写回

* `RestoreByProjectId`:同上,`task.Restore()`(清空 DeletedAt)

* `ArchiveByProjectId`:同上,`task.Archive(time.Now())`

* `UnarchiveByProjectId`:同上,`task.Unarchive()`

为避免循环依赖(infra 调 domain 实体),实体在 domain/task/entities,可直接 import;infrastructure 本就可以 import domain。

具体改造:

```go
func (repo *TaskRepoImpl) ArchiveByProjectId(ctx context.Context, userId int64, projectId int64) error {
    db := dbs.DBFrom(ctx, repo.db)
    var models []models.Task
    if err := db.WithContext(ctx).
        Where("user_id = ? AND project_id = ?", userId, projectId).
        Find(&models).Error; err != nil {
        return err
    }
    if len(models) == 0 {
        return nil
    }
    taskEntities := make([]*entities.Task, 0, len(models))
    for i := range models {
        e := TaskModelToEntity(&models[i])
        e.Archive(nil)
        taskEntities = append(taskEntities, e)
    }
    return db.WithContext(ctx).
        Model(&models[0]).
        Where("user_id = ? AND project_id = ?", userId, projectId).
        Updates(map[string]interface{}{
            "archived_at": gorm.Expr("VALUES(archived_at)"), // 见说明
        }).Error
}
```

实际采用更直接的方式:遍历 entity 逐个 `Updates`:

```go
func (repo *TaskRepoImpl) ArchiveByProjectId(ctx context.Context, userId int64, projectId int64) error {
    db := dbs.DBFrom(ctx, repo.db)
    var taskModels []models.Task
    if err := db.WithContext(ctx).
        Where("user_id = ? AND project_id = ?", userId, projectId).
        Find(&taskModels).Error; err != nil {
        return err
    }
    for i := range taskModels {
        e := TaskModelToEntity(&taskModels[i])
        e.Archive(nil)
        taskModels[i].ArchivedAt = ArchivedAtToSqlNullTime(e.ArchivedAt)
    }
    if len(taskModels) == 0 {
        return nil
    }
    return db.WithContext(ctx).
        Save(&taskModels).Error
}
```

`ArchivedAtToSqlNullTime` 用项目现有 `ArchivedAt.ToSqlNullTime()` 工具(如不存在则新增,放 `infrastructure/persistence/task/converters.go`)。`TaskModelToEntity` 检查是否已存在,不存在则新增(放 `converters.go`)。

#### Task 10: 现有 Task 级联路径适配

`application/project/appImpl.go` 的 4 个级联方法([L145](file:///c:/Users/LEE19/Projects/nao-todo-server/application/project/appImpl.go#L145)、[L183](file:///c:/Users/LEE19/Projects/nao-todo-server/application/project/appImpl.go#L183)、[L226](file:///c:/Users/LEE19/Projects/nao-todo-server/application/project/appImpl.go#L226)、[L263](file:///c:/Users/LEE19/Projects/nao-todo-server/application/project/appImpl.go#L263)) 继续调用 `taskRepo.SoftDeleteByProjectId` 等,无需改动,语义由 infra 层下沉到实体方法。

## 验收标准

### 架构

1. `go build ./...` 0 错误
2. `go vet ./...` 0 告警
3. `go test ./...` 全部通过
4. `golangci-lint run` 无新增告警(允许既有 lll/CRLF 等已知项)
5. `application/**` 目录下所有 .go 文件不再 import `naotodoserver/infrastructure` 任何包(`grep -r "naotodoserver/infrastructure" application/` 零结果)
6. `application/user/avatarStorage.go` 端口位置倒挂 P2 不在本次范围,允许存在

### 行为

1. controller → app 调用全部走 userId 显式参数
2. `domain/task/entities/task.go` 的 `Archive/Unarchive/ToggleStar/GiveUp/UngiveUp` 在生产路径中至少各 1 处调用
3. `Task.IsDatesValid` 在 UpdateTask 写路径被调用(可被读-改-写分支顺带触发)
4. 4 个 ByProjectId 批量方法实现内含 `task.Archive`/`task.Unarchive`/`task.Delete`/`task.Restore` 调用
5. `application/task/appImpl.go ProcessReminders` 通过 `NotificationPublisher` 推送而非直接 `sse.GetHub().Publish`
6. cron/initialize wiring 正确注入 publisher
7. 现有功能行为零破坏:UpdateTask req 不携带 archivedAt/starMarkAt/givenUpAt 时,行为与改前完全一致

### 测试

1. `domain/task/entities/task_test.go` 新增 `Archive`/`Unarchive`/`ToggleStar`/`GiveUp`/`UngiveUp` 单元测试
2. `domain/task/entities/task_test.go` 新增 `IsDatesValid` 触发场景测试
3. `infrastructure/persistence/task/repoImpl.go` 4 个 ByProjectId 方法的端到端行为(集成测试可选,见下"测试策略")

## 测试策略

* **单元测试**:`domain/task/entities` 实体方法测试,与现有 `task_test.go` 风格一致(表驱动,Chinese 注释)

* **集成测试**:本轮不引入新的集成测试框架(项目当前也没有该基础设施);改用现有 `go test ./...` 覆盖,行为验证靠手动/接口调用

* **手动验证清单**:

  1. 启动 dev server,登录用户
  2. PUT 任务 `archivedAt` 字段,确认 DB `archived_at` 写入正确时间
  3. PUT 任务 `archivedAt=""` (清空),确认 DB `archived_at` 置 NULL
  4. PUT 任务 `starMarkAt` 字段,确认切换行为正确
  5. 创建/归档 Project,级联到下属 Task 的 `archived_at` 同步生效
  6. 触发一次 ProcessReminders(cron 或手动调),确认 SSE 推送正常工作

## 风险与回滚

* **风险 1:app 接口签名大改**——5 个 app interface + 全部 controller + wiring。缓解:按域逐个改 + 立刻 `go build` 验证,失败即停

* **风险 2:Task 实体方法签名变化**——`Archive/Unarchive/ToggleStar` 增加参数,破坏既有调用方。缓解:Task 实体当前只被 task\_test.go 引用,Project 实体方法暂不动,本轮改造无外部调用方

* **风险 3:4 个 ByProjectId 批量方法效率下降**——原来一次 SQL 完成,改造后 Find + 逐个 Save。缓解:同 user/project 下 task 数量有限(实测用户清单下任务数 < 1000),单 Save 走主键更新,无明显回归

* **回滚**:每个 Task 自包含,git revert 单 commit 即可

