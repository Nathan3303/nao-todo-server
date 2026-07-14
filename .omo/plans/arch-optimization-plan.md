# 架构优化落地方案

## 总览

针对四个问题进行增量式重构，每个问题可独立执行、独立验证。

| 编号 | 问题 | 改动量 | 风险 | 建议顺序 |
|------|------|--------|------|----------|
| #4 | 拆分 Task Repository 接口 | 中 | 低（接口分拆，实现复用） | 1 |
| #6 | 类型安全 ID | 大 | 中（批量签名变更，机械替换） | 2 |
| #5 | Converter 去重 | 中 | 低（新增 helper，不改逻辑） | 3 |
| #2 | 移除全局单例 `App` | 中 | 中（启动流程变化） | 4 |

---

## Issue #4 — 拆分 Task Repository 接口

### 现状

`domain/task/repositories/task.go` 的 `Task` 接口包含 25 个方法，同时承载 Task、CheckItem、Comment 三个子实体的 CRUD：

```
Task (interface)
├── Task 方法 (12个)  — GetById, Create, Update, Delete, Restore, List, Snooze, ...
├── CheckItem 方法 (7个) — GetCheckItemById, CreateCheckItem, UpdateCheckItem, ...
└── Comment 方法 (7个)  — GetCommentById, CreateComment, UpdateComment, ...
```

只要依赖任一类方法，调用方必须 mock 全部 25 个方法。

### 方案

拆分为三个独立的接口，各自放在独立文件：

```
domain/task/repositories/
├── task.go       → 仅保留 Task 方法
├── checkitem.go  → TaskCheckItem 接口
└── comment.go    → TaskComment 接口
```

### 具体步骤

**Step 4.1 — 新建接口文件**

新建 `domain/task/repositories/checkitem.go`：

```go
package repositories

type TaskCheckItem interface {
    GetCheckItemById(ctx, userId, checkItemId) (*entities.TaskCheckItem, error)
    CreateCheckItem(ctx, userId, vo) (*entities.TaskCheckItem, error)
    UpdateCheckItem(ctx, userId, checkItemId, vo) error
    DeleteCheckItem(ctx, userId, checkItemId) error
    ListCheckItems(ctx, userId, taskId) ([]*entities.TaskCheckItem, error)
    GetMaxCheckItemSortId(ctx, userId, taskId) uint16
    BatchUpdateCheckItems(ctx, userId, vos) ([]*entities.TaskCheckItem, error)
}
```

新建 `domain/task/repositories/comment.go`：

```go
package repositories

type TaskComment interface {
    GetCommentById(ctx, userId, commentId) (*entities.TaskComment, error)
    CreateComment(ctx, userId, vo) (*entities.TaskComment, error)
    UpdateComment(ctx, userId, commentId, vo) error
    DeleteComment(ctx, userId, commentId) error
    ListComments(ctx, userId, taskId) ([]*entities.TaskComment, error)
    SyncCommentUserProfile(ctx, userId, nickname, avatar) error
}
```

**Step 4.2 — 裁剪 `task.go`**

删除 CheckItem 和 Comment 相关方法声明。

**Step 4.3 — Infrastructure 实现**

`TaskRepoImpl` 同时实现三个接口（已实现全部方法，只需确保方法签名匹配）。

在 `infrastructure/persistence/task/repoImpl.go` 顶部增加断言（编译时验证）：

```go
var _ repositories.Task = (*TaskRepoImpl)(nil)
var _ repositories.TaskCheckItem = (*TaskRepoImpl)(nil)
var _ repositories.TaskComment = (*TaskRepoImpl)(nil)
```

**Step 4.4 — Domain Service 适配**

`TaskDomainImpl` 目前通过 `taskRepo repositories.Task` 调用 `CreateCheckItem`。拆分后需要增加 `checkItemRepo repositories.TaskCheckItem` 字段，`CreateCheckItem` 方法改用 `checkItemRepo`。

```go
type TaskDomainImpl struct {
    taskRepo       repositories.Task
    checkItemRepo  repositories.TaskCheckItem
}
```

修改 `NewTaskDomain` 签名，增加 `checkItemRepo` 参数。

**Step 4.5 — Application 层适配**

`TaskAppImpl` 结构体增加 `checkItemRepo` 和 `commentRepo` 字段：

```go
type TaskAppImpl struct {
    taskDomain     service.TaskDomain
    taskRepo       repositories.Task
    checkItemRepo  repositories.TaskCheckItem
    commentRepo    repositories.TaskComment
}
```

修改 `NewTaskApp` 构造函数签名（增加两个参数），将 CheckItem 和 Comment 方法的实现从 `impl.taskRepo.Xxx` 改为 `impl.checkItemRepo.Xxx` / `impl.commentRepo.Xxx`。

**Step 4.6 — 初始化适配**

在 `infrastructure/initialize.go` 中：

```go
taskRepoInst := taskRepo.NewTaskRepo(dbs.DB)
// 同一个实例实现了三个接口
taskDomain := taskService.NewTaskDomain(taskRepoInst, taskRepoInst)
taskAppInst := taskApp.NewTaskApp(taskDomain, taskRepoInst, taskRepoInst, taskRepoInst)
```

**验证方式：** `go build ./...` 通过即完成。

---

## Issue #6 — 类型安全 ID

### 现状

`int64` 贯穿所有层，userId / projectId / taskId 在函数签名中无法区分。

### 方案

在 `domain/types/` 下新建 `id.go`，定义命名类型：

```go
package types

type (
    UserID      int64
    ProjectID   int64
    TaskID      int64
    TagID       int64
    CommentID   int64
    CheckItemID int64
    EventID     int64
)
```

### 变更策略

高 ROI + 低侵入原则：

| 层级 | 变更程度 | 说明 |
|------|----------|------|
| domain/entities | **全局替换** | `UserId int64` → `UserId types.UserID` |
| domain/repositories | **全局替换** | 接口签名参数类型 |
| domain/service | **全局替换** | 接口和实现签名 |
| domain/valueobjects | **全局替换** | VO 结构体字段 |
| application | **全局替换** | app 接口和实现签名、converter 签名 |
| infrastructure/persistence | **仅边界处** | repo 实现中接收 typed ID，传给 GORM model 时 `int64(id)` 转换 |
| infrastructure/cron | **按需** | cron 任务中 ID 参数 |
| interfaces | **不做大改** | controller 层保持 `string`（用户输入解析），parse 后转为 typed ID |

核心转换点（在 infrastructure 层）：

```go
// 正向：typed ID → GORM int64
whereCond.UserId = int64(userId)

// 反向（Snowflake ID → typed ID）
entity.UserId = types.UserID(model.UserId)
```

**注意**：`domain/types/EntityBase` 中的 `Id int64` 需要改为 `Id types.SnowflakeID` 或保持现状（GORM 层面兼容）。最安全的做法：在 EntityBase 中保持 `int64`，在各个实体的 typed 字段上加显式类型。

实际上更推荐的思路：**不为 EntityBase 的 Id 加类型**（因为 entity 不知道自己是哪种 ID，它是泛化的），只为函数参数和业务字段加类型。

> 修正方案：EntityBase 保持 `Id int64`，只为业务参数（方法签名中的 `userId`, `projectId` 等）加类型。

### 具体步骤

**Step 6.1 — 创建 `domain/types/id.go`**

定义所有 ID 类型及辅助方法（`String()`、`Parse()` 等）。

**Step 6.2 — 批量替换 domain 层**

逐步替换 domain 下所有仓库接口、实体字段、值对象、服务接口中的 `int64` ID 参数：

- `userId int64` → `userId types.UserID`
- `projectId int64` → `projectId types.ProjectID`
- `taskId int64` → `taskId types.TaskID`
- 以此类推

**Step 6.3 — 修改 application 层**

应用层接口和实现签名同步更新。

**Step 6.4 — 修改 infrastructure 边界**

Repository 实现中：接收 typed ID → `int64(id)` 传给 GORM。Model → Entity 转换时：`types.UserID(model.UserId)`。

**Step 6.5 — 修改 interfaces 层**

Controller 中 `idutil.ParseID()` 返回的 `int64` 直接转为 typed ID。

**验证方式：** `go build ./...` + `go vet ./...`。

---

## Issue #5 — Converter 去重

### 现状

三组重复模式：

1. **Entity ↔ Res converter 重复** — 如 `ProjectEntityToCreateRes` 和 `ProjectEntityToGetRes` 字段赋值完全一样
2. **Model ↔ Entity converter 模板代码** — 每个模块一套手写赋值循环
3. **Task Update 双重转换** — `UpdateTaskValueObjectToModel` 和 `UpdateTaskValueObjectToMap` 做相同的条件判断

### 方案

分三个子任务解决：

### Step 5.1 — 消除 Entity→Res 重复

**涉及模块：** Project、Tag（Task 的 Converter 更复杂，暂不简化）

对 Project：
- `ProjectEntityToCreateRes` 和 `ProjectEntityToGetRes` 合并为一个 `ProjectEntityToRes`
- `GetProjectRes` 已经 `type GetProjectRes CreateProjectRes`，只需要一个转换函数返回 `*GetProjectRes`

**操作：** 删除 `ProjectEntityToCreateRes`，保留 `ProjectEntityToGetRes`（因为 `GetProjectRes` 和 `CreateProjectRes` 共享底层结构），`Create` 方法中直接调用 `ProjectEntityToGetRes` 再类型断言到 `*CreateProjectRes`。

简化后：
```go
// application/project/converters.go 精简约 15 行
func ProjectEntityToRes(e *entities.Project) *types.GetProjectRes { ... }
// Create 直接用类型转换
res := (*types.CreateProjectRes)(ProjectEntityToRes(entity))
```

### Step 5.2 — 通用结构体拷贝 Helper

在 `domain/types/` 下新建 `mapper.go`，提供通用拷贝函数：

```go
package types

// CopyStruct 浅拷贝同名字段（用于简单的 Entity ↔ Model 转换）
// 仅拷贝可导出的、类型完全匹配的字段。跳过不匹配的字段。
func CopyStruct(src, dst any) { ... }
```

基于 `reflect` 实现，仅处理简单类型字段（string, int64, uint8, uint16 等），遇到复杂类型（Nested struct, map, slice）跳过。**适用范围**：Project、Tag、ProjectPreference 等字段简单的模块。

使用示例（`infrastructure/persistence/project/converters.go`）：

```go
func ProjectModelToEntity(m *models.Project) *entities.Project {
    e := &entities.Project{}
    types.CopyStruct(m, e)
    return e
}
```

配合同等简化 `EntityToModel`，可将 Project 和 Tag 的 Model↔Entity 转换器从 ~30 行压缩到 ~5 行。

> ⚠️ **Task 的 Model/Entity 字段不一致**（Model 用 `sql.NullTime`，Entity 用 `types.NullableTime`），不走通用拷贝，保持手写。

### Step 5.3 — 消除 Task Update 双重转换

现状：`UpdateTaskValueObjectToModel`（106-107行）和 `UpdateTaskValueObjectToMap`（110-195行）包含完全相同的条件判断逻辑（约 40 个 if 块）。

方案：**删除 `UpdateTaskValueObjectToModel`**（不再使用），只保留 `UpdateTaskValueObjectToMap`。Task Update 流程改为直接使用 Map 方式更新（GORM 的 `Updates(map)` 效果等同且更灵活）。

检查 `UpdateTaskValueObjectToModel` 的所有调用方，替换为 `UpdateTaskValueObjectToMap`。

### 验证方式

`go build ./...` + 对比 Project/Tag 的 converter 测试确保映射正确。

---

## Issue #2 — 移除全局单例 `application.App`

### 现状

依赖链：
```
infrastructure.LoadDomains() → application.App = &Services{...}
                                  ↑ 全局变量
routers.InitRouters() → application.App.XXX → controller 构造函数
middlewares          → application.App.Auth
cron jobs            → application.App.XXX
```

### 方案

去掉全局变量，改为构造时显式传参。分为四个独立子步骤：

### Step 2.1 — LoadDomains 返回 `*Services`

```go
// infrastructure/initialize.go
func LoadDomains() *application.Services {
    // ... 全部组装逻辑不变 ...
    return &application.Services{
        Auth:          ...,
        User:          ...,
        Task:          taskAppInst,
        TaskComment:   taskAppInst,
        TaskCheckItem: taskAppInst,
        Project:       projectAppInst,
        Tag:           tagAppInst,
        Pomodoro:      pomodoroAppInst,
    }
}
```

删除 `application/services.go` 中的 `var App *Services`。

### Step 2.2 — 修改 main.go

```go
func main() {
    conf.InitConfig()
    runtime.GOMAXPROCS(...)
    infrastructure.LoadLogger()
    infrastructure.LoadDBs()
    svc := infrastructure.LoadDomains()   // ← 接收返回值
    infrastructure.WireSSE(svc)            // ← 传参
    infrastructure.LoadCron(svc)           // ← 传参
    router := routers.InitRouters(svc)     // ← 传参
    // ...
}
```

### Step 2.3 — 改造 Routers

`InitRouters` 签名改为：

```go
func InitRouters(svc *application.Services) *gin.Engine {
    // ...
    authCtrl := controllers.NewAuthController(svc.Auth)
    userCtrl := controllers.NewUserController(svc.User)
    // ...
}
```

**Middleware 改造**：`jwtValidator.go` 和 `rateLimit.go` 目前直接读 `application.App.Auth`。改为函数工厂（闭包）或 Middleware 结构体：

```go
// interfaces/middlewares/jwtValidator.go
func JWTAuthMiddleware(authApp authApp.AuthApp) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        userId, err := authApp.Validate(ctx, jwtString)
        // ...
    }
}
```

在 `InitRouters` 中注入：

```go
v1.Use(middlewares.JWTAuthMiddleware(svc.Auth))
```

### Step 2.4 — 改造 Cron Jobs

Cron job 改为接收依赖：

```go
// infrastructure/cron/deleteInactiveProject.go
type DeleteDeactivedProjectJob struct {
    DayOffset   int8
    ProjectApp  projectApp.ProjectApp  // ← 新增字段
}

func NewDeleteDeactivedProjectJob(dayOffset int8, projectApp projectApp.ProjectApp) *DeleteDeactivedProjectJob {
    return &DeleteDeactivedProjectJob{
        DayOffset:  dayOffset,
        ProjectApp: projectApp,
    }
}

func (ddp *DeleteDeactivedProjectJob) Run() {
    err := ddp.ProjectApp.DeleteDeactivatedProjects(context.TODO(), ddp.DayOffset)
    // ...
}
```

在 `LoadCron` 中传入：

```go
func LoadCron(svc *application.Services) {
    // ...
    cronService.AddJob("0 2 * * *", cron.NewDeleteDeactivedUserJob(15, svc.User))
    cronService.AddJob("* * * * *", cron.NewReminderJob(svc.Task))
    // ...
}
```

### 验证方式

`go build ./...` 且 `go run cmd/main.go` 能正常启动，API 响应正常。

---

## 执行顺序 & 并行度

```
Step 4 (Repo拆分) ──→ Step 6 (类型ID) ──→ Step 5 (Converter) ──→ Step 2 (移除全局)
      独立             独立（机械替换）       独立（helper 新增）      依赖前面步骤
```

**实际执行建议**：Step 4, 5, 6 可分别由三个子代理并行执行（互不依赖），最后 Step 2 由主流程串行。

---

## 回滚策略

每一步完成后运行 `go build ./...` && `go vet ./...`。如某一步失败，git checkout 该步骤涉及的文件即可回滚。建议为每个步骤创建独立 commit。
