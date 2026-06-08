# 待办任务专注记录（Pomodoro）模块

## 概述

新增 Pomodoro（番茄专注记录）独立领域模块。提供**创建专注记录**和**查询专注记录（单条 + 列表）**两条执行链，其他操作暂不实现。

## 架构决策

### Pomodoro 作为独立领域模块

**Pomodoro 不归为 Task 子域，而是独立领域模块**（类似 project / tag），与 Task 仅通过 `taskId` 关联。

| 维度 | CheckItem / Comment（Task 子实体） | Pomodoro |
|---|---|---|
| 生命周期 | 完全依赖 Task，级联操作 | 独立生命周期，按 session/日期/任务多维度查询 |
| 查询语义 | 仅 `taskId → list` | `sessionId`, `startTime-endTime`, `taskId`, `type`, 分页, 排序 |
| 业务含义 | Task 的内聚属性（清单/评论） | 独立的专注计时记录 |
| 仓库接口内聚性 | 放 Task 接口内尚可 | 放进去会暴涨 Task 接口，违反 ISP |

### 错误码范围

`70000-70099`（现有范围：10000=auth, 20000=project, 30000=tag, 40000=task, 50000=event, 60000=comment）

## 数据库设计

### 表名

`pomodoros`（GORM 默认表名规则：模型名 `Pomodoro` → 蛇形复数 `pomodoros`）

### GORM 模型

```go
// infrastructure/persistence/models/pomodoro.go

package models

import "database/sql"

type Pomodoro struct {
    ModelBase
    UserId      int64        `gorm:"not null;index:idx_pomodoro_user_id"`
    SessionId   string       `gorm:"size:36;not null;index:idx_pomodoro_session_id"`
    Type        uint8        `gorm:"not null;default:0"`
    TaskId      int64        `gorm:"not null;index:idx_pomodoro_task_id"`
    TaskName    string       `gorm:"size:256;not null"`
    Description string       `gorm:"size:512"`
    StartAt     sql.NullTime `gorm:"null;index:idx_pomodoro_start_at"`
    EndAt       sql.NullTime `gorm:"null"`
    Duration    int          `gorm:"not null"`
        Note        string       `gorm:"type:text"`
}
```

### 字段说明

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `ID` | int64 | PK, Auto Snowflake | 继承 ModelBase，雪花算法生成 |
| `DeletedAt` | gorm.DeletedAt | 有索引 | 继承 ModelBase，软删除 |
| `CreatedAt` | time.Time | 自动 | 继承 ModelBase |
| `UpdatedAt` | time.Time | 自动 | 继承 ModelBase |
| `UserId` | int64 | NOT NULL, INDEX | 用户 ID（所有实体统一模式） |
| `SessionId` | string(36) | NOT NULL, INDEX | UUID v4，用于分组一次完整番茄工作法的会话 |
| `Type` | uint8 | NOT NULL, DEFAULT 0 | 专注类型（0=专注, 1=短休息, 2=长休息 等） |
| `TaskId` | int64 | NOT NULL, INDEX | 关联的任务 ID（雪花 ID） |
| `TaskName` | string(256) | NOT NULL | 反范式冗余任务名称，避免关联查询 |
| `Description` | string(512) | 可空 | 描述 |
| `StartAt` | datetime | NULL, INDEX | 专注开始时间（sql.NullTime，与 Task 模型一致） |
| `EndAt` | datetime | NULL | 专注结束时间 |
| `Duration` | int | NOT NULL | 专注时长（分钟） |
| `Note` | text | 可空 | 备注（最大 64KB，支持约 1000 中英文字符） |

### 索引

| 索引名 | 字段 | 说明 |
|--------|------|------|
| `idx_pomodoro_user_id` | UserId | 按用户查询 |
| `idx_pomodoro_session_id` | SessionId | 按会话分组查询 |
| `idx_pomodoro_task_id` | TaskId | 按任务查询 |
| `idx_pomodoro_start_at` | StartAt | 按时间范围查询 |

### GORM AutoMigrate

在 `infrastructure/persistence/dbs/mysql.go` 的 `DoMigration()` 中追加 `models.Pomodoro{}`。

## 领域层设计

### Entity

```go
// domain/pomodoro/entities/pomodoro.go

package entities

type Pomodoro struct {
    Id          int64
    UserId      int64
    SessionId   string
    Type        uint8
    TaskId      int64
    TaskName    string
    Description string
    StartAt     time.Time
    EndAt       time.Time
    Duration    int
    Note        string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   time.Time
}
```

### ValueObjects

仅创建操作用到的值对象（无 update 值对象）：

```go
// domain/pomodoro/valueobjects/createPomodoro.go

type CreatePomodoro struct {
    UserId      int64
    SessionId   string
    Type        uint8
    TaskId      int64
    TaskName    string
    Description string
    StartAt     *time.Time
    EndAt       *time.Time
    Duration    int
    Note        string
}

func NewCreatePomodoro(userId int64, ...) (*CreatePomodoro, error)
func (vo *CreatePomodoro) Validate() error
```

验证规则：
- `SessionId` 必填
- `TaskId` > 0
- `TaskName` 非空
- `StartAt`、`EndAt` 非空
- `Duration` > 0
- `StartAt` 不能晚于 `EndAt`

### Repository Interface

```go
// domain/pomodoro/repositories/pomodoro.go

type Pomodoro interface {
    Create(ctx context.Context, vo *valueobjects.CreatePomodoro) (*entities.Pomodoro, error)
    GetById(ctx context.Context, userId int64, id int64) (*entities.Pomodoro, error)
    List(ctx context.Context, userId int64, req *types.ListPomodoroReq) ([]*entities.Pomodoro, int64, error)
}
```

### Domain Service

```go
// domain/pomodoro/service/service.go

type PomodoroDomain interface {
    Create(ctx context.Context, userId int64, vo *valueobjects.CreatePomodoro) (*entities.Pomodoro, error)
    GetById(ctx context.Context, userId int64, id int64) (*entities.Pomodoro, error)
    List(ctx context.Context, userId int64, req *types.ListPomodoroReq) ([]*entities.Pomodoro, int64, error)
}
```

领域服务实现 `serviceImpl.go`：
- `Create`：调用 `vo.Validate()` → `repo.Create()`
- `GetById`：直接委托 `repo.GetById()`
- `List`：直接委托 `repo.List()`

## 应用层设计

```go
// application/pomodoro/app.go

type PomodoroApp interface {
    CreatePomodoro(ctx context.Context, req *types.CreatePomodoroReq) (*types.CreatePomodoroRes, error)
    GetPomodoro(ctx context.Context, req *types.GetPomodoroReq) (*types.GetPomodoroRes, error)
    ListPomodoro(ctx context.Context, req *types.ListPomodoroReq) ([]*types.GetPomodoroRes, int64, error)
}
```

### 数据流（以 Create 为例）

```
从 ctx 提取 userId
  → types.CreatePomodoroReq → valueobjects.CreatePomodoro（含字符串→int64/t.Parse 转换）
    → domain.PomodoroDomain.Create(ctx, userId, vo)
      → repo.Create(ctx, userId, vo)
        → entity → types.CreatePomodoroRes（含 int64→string 转换）
```

### Converters

```go
// application/pomodoro/converters.go
CreatePomodoroReqToVO(userId int64, req *types.CreatePomodoroReq) (*valueobjects.CreatePomodoro, error)
PomodoroEntityToCreateRes(e *entities.Pomodoro) *types.CreatePomodoroRes
PomodoroEntityToGetRes(e *entities.Pomodoro) *types.GetPomodoroRes
PomodoroEntitiesToGetReses(list []*entities.Pomodoro) []*types.GetPomodoroRes
```

## 基础设施层设计

### Repository Impl

```go
// infrastructure/persistence/pomodoro/repoImpl.go

type PomodoroRepoImpl struct {
    db *gorm.DB
}

func (r *PomodoroRepoImpl) Create(ctx, userId, vo) (*entities.Pomodoro, error)
func (r *PomodoroRepoImpl) GetById(ctx, userId, id) (*entities.Pomodoro, error)
func (r *PomodoroRepoImpl) List(ctx, userId, req) ([]*entities.Pomodoro, total, error)
```

- `Create`：VO → GORM Model → `db.Create()` → Model → Entity
- `GetById`：`db.Where("id = ? AND user_id = ?", id, userId).First()` → Model → Entity
- `List`：构建带 WHERE 条件的 `*gorm.DB` → `Count(&total)` → `Scopes(Paginate).Find()` → Models → Entities

### List 查询条件构建

| ListPomodoroReq 字段 | SQL 条件 |
|----------------------|----------|
| `SessionId` | `session_id = ?` |
| `StartTime` + `EndTime` | `start_at BETWEEN ? AND ?` |
| `TaskId` | `task_id = ?`（字符串→int64） |
| `TaskName` | `task_name LIKE ?` |
| `Type` | `type = ?` |
| `Sort` | `ORDER BY field dir` |
| `Page` / `Limit` | `OFFSET ? LIMIT ?` |

复用 `infrastructure/persistence/task/repoImpl.go` 中的 `PaginationVO2Scopes` 模式。

### Converters

```go
// infrastructure/persistence/pomodoro/converters.go
CreatePomodoroVOToModel(userId int64, vo *valueobjects.CreatePomodoro) *models.Pomodoro
PomodoroModel2Entity(m *models.Pomodoro) *entities.Pomodoro
PomodoroModels2Entities(list []*models.Pomodoro) []*entities.Pomodoro
```

## HTTP 层设计

### Types（已存在）

`interfaces/types/pomodoro.go` — 已有 `CreatePomodoroReq`/`Res`、`GetPomodoroReq`/`Res`、`ListPomodoroReq`，**无需修改**。

### Controllers

```go
// interfaces/controllers/pomodoro.go

// CreatePomodoroHandler 创建专注记录
// @code 7001x
func CreatePomodoroHandler(ctx *gin.Context)  // POST JSON → App.CreatePomodoro

// GetPomodoroHandler 获取单条专注记录
// @code 7002x
func GetPomodoroHandler(ctx *gin.Context)     // GET /:id → App.GetPomodoro

// ListPomodoroHandler 获取专注记录列表
// @code 7003x
func ListPomodoroHandler(ctx *gin.Context)    // GET /?query → App.ListPomodoro
```

错误码分配：
- Create: 70010=成功, 70011=参数错误, 70012=创建失败
- Get: 70020=成功, 70021=ID无效, 70022=查询失败
- List: 70030=成功, 70031=参数错误, 70032=查询失败

### Routers

```go
// interfaces/routers/pomodoroRouter.go

func UsePomodoroRouter(router *gin.RouterGroup) {
    g := router.Group("/pomodoros",
        middlewares.RateLimiter(48, "pomodoros"),
        middlewares.JWTValidator,
    )
    {
        g.POST("/", controllers.CreatePomodoroHandler)
        g.GET("/:id", controllers.GetPomodoroHandler)
        g.GET("/", controllers.ListPomodoroHandler)
    }
}
```

## 依赖注入

### application/services.go

```go
type Services struct {
    // ... existing fields ...
    Pomodoro pomodoroApp.PomodoroApp   // 新增
}
```

### infrastructure/initialize.go

```go
func LoadDomains() {
    // ... existing wiring ...

    application.App = &application.Services{
        // ... existing services ...
        Pomodoro: pomodoroApp.NewPomodoroApp(
            pomodoroService.NewPomodoroDomain(
                pomodoroRepo.NewPomodoroRepo(dbs.DB),
            ),
        ),
    }
}
```

## 实施清单

| # | 文件 | 操作 | 说明 |
|---|------|------|------|
| 1 | `domain/pomodoro/entities/pomodoro.go` | 创建 | Pomodoro 实体 |
| 2 | `domain/pomodoro/valueobjects/createPomodoro.go` | 创建 | 创建值对象 + 验证 |
| 3 | `domain/pomodoro/repositories/pomodoro.go` | 创建 | 仓库接口 |
| 4 | `domain/pomodoro/service/service.go` | 创建 | 领域服务接口 |
| 5 | `domain/pomodoro/service/serviceImpl.go` | 创建 | 领域服务实现 |
| 6 | `application/pomodoro/app.go` | 创建 | 应用服务接口 |
| 7 | `application/pomodoro/appImpl.go` | 创建 | 应用服务实现 |
| 8 | `application/pomodoro/converters.go` | 创建 | 应用层类型转换 |
| 9 | `infrastructure/persistence/models/pomodoro.go` | 创建 | GORM 模型 |
| 10 | `infrastructure/persistence/pomodoro/repoImpl.go` | 创建 | 仓库 GORM 实现 |
| 11 | `infrastructure/persistence/pomodoro/converters.go` | 创建 | 基础设施层类型转换 |
| 12 | `interfaces/controllers/pomodoro.go` | 创建 | HTTP 控制器 |
| 13 | `interfaces/routers/pomodoroRouter.go` | 创建 | 路由注册 |
| 14 | `application/services.go` | 修改 | 添加 Pomodoro 字段 |
| 15 | `infrastructure/initialize.go` | 修改 | 依赖注入装配 |
| 16 | `infrastructure/persistence/dbs/mysql.go` | 修改 | AutoMigrate 添加 Pomodoro 模型 |
| 17 | `interfaces/routers/routers.go` | 修改 | 注册 Pomodoro 路由 |
