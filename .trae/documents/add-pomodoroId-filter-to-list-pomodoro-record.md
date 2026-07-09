# 为 List PomodoroRecord 增加 PomodoroId 筛选

## 概述

在 `List PomodoroRecord` 的完整执行链上增加 `PomodoroId` 筛选参数，支持通过常用专注 ID 查询其所有专注记录。

## 当前状态分析

`PomodoroRecord` 数据库模型 ([models/pomodoro.go](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/models/pomodoro.go#L52)) 已有 `PomodoroId int64` 字段和索引，但整个查询链路上缺少对该字段的支持：

| 层级 | 文件 | 当前状态 |
|------|------|----------|
| 实体 | `entities/pomodoroRecord.go` | 无 `PomodoroId` 字段 |
| 查询 VO | `valueobjects/queryPomodoroRecord.go` | 无 `PomodoroId` 字段 |
| 请求类型 | `types/pomodoro.go` (`ListPomodoroRecordReq`) | 无 `pomodoroId` 查询参数 |
| Scope | `scopes.go` | 无 `ByPomodoroRecordPomodoroId` |
| 仓储实现 | `pomodoroRecordRepoImpl.go` | List 方法未调用 PomodoroId scope |
| 应用层转换 | `application/pomodoro/converters.go` | 未传递 PomodoroId |
| 持久层转换 | `infrastructure/persistence/pomodoro/converters.go` | Model→Entity 未映射 PomodoroId |

## 变更方案

遵循现有 `TaskId` 筛选的完整模式，在 7 个文件中添加 `PomodoroId` 支持。

### 1. 实体层 - 添加 PomodoroId 字段

**文件**: [domain/pomodoro/entities/pomodoroRecord.go](file:///home/nathan/Projects/nao-todo-server/domain/pomodoro/entities/pomodoroRecord.go)

在 `PomodoroRecord` 结构体中添加 `PomodoroId int64` 字段（放在 `UserId` 之后，与 DB 模型字段顺序一致）。

### 2. 查询值对象 - 添加 PomodoroId 查询字段

**文件**: [domain/pomodoro/valueobjects/queryPomodoroRecord.go](file:///home/nathan/Projects/nao-todo-server/domain/pomodoro/valueobjects/queryPomodoroRecord.go)

- `QueryPomodoroRecord` 结构体添加 `PomodoroId int64` 字段
- `NewQueryPomodoroRecord` 构造函数添加 `pomodoroId int64` 参数

### 3. 请求类型 - 添加 pomodoroId 查询参数

**文件**: [interfaces/types/pomodoro.go](file:///home/nathan/Projects/nao-todo-server/interfaces/types/pomodoro.go)

`ListPomodoroRecordReq` 结构体添加 `PomodoroId string` form 字段（与 `TaskId` 同为 string 类型，在应用层转换时解析为 int64）。

### 4. Scope 函数 - 添加 ByPomodoroRecordPomodoroId

**文件**: [infrastructure/persistence/pomodoro/scopes.go](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/pomodoro/scopes.go)

新增 `ByPomodoroRecordPomodoroId` 函数，与 `ByPomodoroRecordTaskId` 模式一致：当 `pomodoroId <= 0` 时跳过，否则添加 `WHERE pomodoro_id = ?` 条件。

### 5. 仓储实现 - List 方法添加 scope 调用

**文件**: [infrastructure/persistence/pomodoro/pomodoroRecordRepoImpl.go](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/pomodoro/pomodoroRecordRepoImpl.go)

在 `List` 方法的 `Scopes` 链中添加 `ByPomodoroRecordPomodoroId(q.PomodoroId)`（放在 `ByPomodoroRecordSessionId` 之后，与其他 ID 类筛选相邻）。

### 6. 应用层转换器 - 传递 PomodoroId

**文件**: [application/pomodoro/converters.go](file:///home/nathan/Projects/nao-todo-server/application/pomodoro/converters.go)

`ListPomodoroRecordReqToQueryVO` 函数中解析 `req.PomodoroId` 为 int64（与 TaskId 解析逻辑一致），传入 `NewQueryPomodoroRecord`。

### 7. 持久层转换器 - Model→Entity 映射 PomodoroId

**文件**: [infrastructure/persistence/pomodoro/converters.go](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/pomodoro/converters.go)

`PomodoroRecordModel2Entity` 函数中添加 `e.PomodoroId = m.PomodoroId` 映射。

## 不变更的部分

- **仓储接口** ([repositories/pomodoroRecord.go](file:///home/nathan/Projects/nao-todo-server/domain/pomodoro/repositories/pomodoroRecord.go))：`List` 方法签名无需修改，因为 `QueryPomodoroRecord` 已包含新字段。
- **控制器** ([controllers/pomodoro.go](file:///home/nathan/Projects/nao-todo-server/interfaces/controllers/pomodoro.go))：`ShouldBindQuery` 自动绑定新 form 字段，无需修改。
- **路由** ([routers/pomodoroRouter.go](file:///home/nathan/Projects/nao-todo-server/interfaces/routers/pomodoroRouter.go))：无需修改。
- **响应类型**：`GetPomodoroRecordRes` / `CreatePomodoroRecordRes` 不添加 `PomodoroId`，用户未要求返回该字段。

## 验证步骤

1. `go build ./...` 编译通过
2. 调用 `GET /pomodoro-records?pomodoroId=123` 验证可按 PomodoroId 筛选
3. 不传 `pomodoroId` 时行为不变（向后兼容）