# 实体充血化与四域 DTO 解耦 Spec

## Why

上一批整改（fix-ddd-layering-violations）修复了枚举、gin 泄漏、事务边界，并将 user/auth 两域的 DTO 与 `interfaces/types` 解耦。但领域层仍存在两个突出问题：

1. **实体贫血**：`Project` 实体零方法，归档/删除等状态迁移的 SQL 全部散落在基础设施层（`repoImpl.go` 直接 `UPDATE archived_at`）；`Task` 的状态变更由应用层转换器直接透传 `State` 字段，无状态机校验；`TaskCheckItem` 的 `MarkDone/MarkUndone/ToggleDone/IsCompleted` 是全项目零调用的死代码方法。
2. **剩余四域反向依赖**：`task`/`project`/`tag`/`pomodoro` 的 application 层仍直接依赖 `interfaces/types`（14 个文件），依赖方向违反 DDD 分层。

## What Changes

**Part A：实体充血化（第一批，从状态迁移开始）**

- `Project` 实体新增状态迁移方法 `Archive()/Unarchive()/Delete()/Restore()`，领域服务（`ProjectDomainImpl`）从"直接委托 repo 执行 UPDATE"改为"加载实体 → 应用实体方法 → 持久化实体状态"。
- `Task` 实体新增 `ChangeState(next)/Archive()/Unarchive()/ToggleStar()` 状态迁移方法；`UpdateTask` 的 `State` 变更路径改为读-改-写（加载实体 → 状态机迁移 → 落库），并补齐 `CompletedAt` 的落库支持（此前传 `state=completed` 不写 `completed_at`，属隐性缺陷）。
- `TaskCheckItem` 的死代码方法 `MarkDone/MarkUndone/ToggleDone/IsCompleted` 接入 `UpdateTaskCheckItem` / `BatchUpdateTaskCheckItems` 生产路径（读-改-写）。
- 新增/改造均保持既有 HTTP 契约（JSON 字段名、错误码、PATCH 语义）不变。

**Part B：application 自建 DTO，切断对 `interfaces/types` 的依赖（按域渐进）**

- 对 `tag` → `pomodoro` → `project` → `task` 四域复用 user/auth 样板：新建 `application/<domain>/dto/`（无 json/form tag），app 接口签名改用 dto 类型，HTTP DTO ↔ app DTO 转换下沉到 controller，app DTO ↔ domain VO 转换保留在应用层 converters。

## Impact

- Affected specs：无（新增能力域）
- Affected code：
  - `domain/project/entities/project.go`、`domain/project/service/service.go|serviceImpl.go`、`domain/project/repositories/project.go`（新增 `UpdateState`）
  - `infrastructure/persistence/project/repoImpl.go`
  - `domain/task/entities/task.go`、`domain/task/valueobjects/updateTask.go`、`application/task/appImpl.go|converters.go`
  - `infrastructure/persistence/task/converters.go`（`UpdateTaskValueObjectToMap` 支持 CompletedAt）
  - `application/{tag,pomodoro,project,task}/` 全部（新建 dto 包，改造 app/appImpl/converters）
  - `interfaces/controllers/{tag,pomodoro,project,task}.go`（承担 HTTP ↔ app 转换）

## ADDED Requirements

### Requirement: Project 实体状态迁移方法

`Project` 实体 SHALL 提供 `Archive()/Unarchive()/Delete()/Restore()` 四个状态迁移方法，直接修改自身 `ArchivedAt`/`DeactivedAt` 字段，行为与现有 API 一致（幂等、重设或清空时间戳）。

#### Scenario: 归档清单

- **WHEN** 领域服务对未归档清单调用实体 `Archive()`
- **THEN** 实体 `ArchivedAt` 被设置为当前时间

#### Scenario: 取消归档清单

- **WHEN** 领域服务调用实体 `Unarchive()`
- **THEN** 实体 `ArchivedAt` 被清空（`NullableTime` 无效/为空）

### Requirement: Project 领域服务经实体方法执行状态迁移

`ProjectDomainImpl` 的 `Archive/Unarchive/Delete/Restore` SHALL 先通过 `repo.GetById` 加载实体，调用实体状态迁移方法，再通过新增仓储方法 `UpdateState` 持久化实体状态字段。基础设施层不再存在独立的 `Archive/Unarchive/Delete/Restore` SQL 方法（由 `UpdateState` 取代）。

#### Scenario: 归档清单（领域服务）

- **WHEN** 应用层在事务内调用 `projectDomain.Archive(ctx, userId, projectId)`
- **THEN** 事务内完成实体加载、`entity.Archive()`、`repo.UpdateState` 持久化；`RowsAffected == 0` 时返回"清单不存在"错误

### Requirement: Task 实体状态迁移与状态机

`Task` 实体 SHALL 提供 `ChangeState(next TaskState) error`，仅允许迁移至合法值（Pending/InProgress/Completed），Completed 时设置 `CompletedAt=now`，离开 Completed 时清空 `CompletedAt`，非法值返回错误；同状态迁移幂等。

#### Scenario: 完成任务

- **WHEN** 用户通过更新接口将任务 State 改为 `completed`
- **THEN** 应用层加载实体、调用 `entity.ChangeState(TaskStateCompleted)`，落库字段包含 `state=3` 与 `completed_at=now`

#### Scenario: 回退已完成任务

- **WHEN** 用户将已完成任务的 State 改回 `in-progress`
- **THEN** 实体 `CompletedAt` 被清空，`State` 落库为 2

### Requirement: TaskCheckItem 死代码方法接入生产路径

`UpdateTaskCheckItem` 与 `BatchUpdateTaskCheckItems` SHALL 在请求携带 `IsDone` 时采用读-改-写：加载实体，通过 `IsCompleted()` 比较当前值，通过 `MarkDone()/MarkUndone()`（或 `ToggleDone()`）迁移，再持久化。`MarkDone/MarkUndone/ToggleDone/IsCompleted` 不再是无引用的死代码。

#### Scenario: 勾选检查项

- **WHEN** 用户更新检查项且 `IsDone=true`、实体当前未完成
- **THEN** 应用层调用 `entity.MarkDone()` 后落库，响应/列表字段与改造前一致

### Requirement: tag 域应用层 DTO 解耦

`application/tag/dto/` SHALL 建立，`TagApp` 接口全部方法签名仅使用 dto 类型；`application/tag/` 下所有文件 Grep `naotodoserver/interfaces/types` 零命中；controller 承担 HTTP DTO ↔ app DTO 转换，JSON 契约与错误码不变。

### Requirement: pomodoro 域应用层 DTO 解耦

`application/pomodoro/dto/` SHALL 建立，`PomodoroApp` 接口签名仅使用 dto 类型；`application/pomodoro/` 下 Grep 零命中；controller 承担转换，JSON 契约与错误码不变。

### Requirement: project 域应用层 DTO 解耦

`application/project/dto/` SHALL 建立，`ProjectApp` 接口签名仅使用 dto 类型；`application/project/` 下 Grep 零命中；controller 承担转换，JSON 契约与错误码不变；`DeleteDeactivatedProjects`（cron 调用）签名同步更新。

### Requirement: task 域应用层 DTO 解耦

`application/task/dto/` SHALL 建立（含 Task/TaskCheckItem/TaskComment 三套 DTO），`TaskApp` 与 `TaskCheckItemApp` 接口签名仅使用 dto 类型；`application/task/` 下 Grep 零命中；controller 承担转换，JSON 契约与错误码不变。

## MODIFIED Requirements

### Requirement: UpdateTask 值对象支持 CompletedAt

`valueobjects.UpdateTask` 及其构造器、`infrastructure/persistence/task/converters.go` 的 `UpdateTaskValueObjectToMap` SHALL 支持 `CompletedAt` 字段落库，供实体状态机产生的完成时间持久化。**补充说明**：此前 `state=completed` 不写 `completed_at`，本次补齐属隐性缺陷修复，非 API 破坏。

### Requirement: UpdateTask 状态变更读-改-写

`application/task/appImpl.go` 的 `UpdateTask` SHALL 在请求携带 `State` 时，先 `repo.GetById` 加载实体并调用 `entity.ChangeState`，再以实体迁移后的 `State`/`CompletedAt` 构造更新值对象落库；未携带 `State` 时保持原路径（不额外读库）。

#### Scenario: 更新任务名称（不涉及状态）

- **WHEN** 用户仅更新任务名称
- **THEN** 不触发读库与状态机，行为与改造前一致

### Requirement: Project 仓储接口调整

`repositories.Project` SHALL 新增 `UpdateState(ctx, userId, projectId int64, archivedAt, deactivedAt types.NullableTime) error`；原 `Archive/Unarchive/Delete/Restore` 方法在确认无其他调用方后 SHALL 从接口与实现中移除（其职责并入 `UpdateState`）。

## REMOVED Requirements

### Requirement: 基础设施层直改状态列的仓储方法

**Reason**：状态迁移逻辑应从基础设施层上移到领域实体，实体方法成为唯一权威。
**Migration**：`repoImpl.go` 中的 `Archive/Unarchive/Delete/Restore` 由 `UpdateState`（仅更新 `ShouldUpdate` 的状态字段，保留"清单不存在"错误语义）替代；领域服务改为加载实体后应用实体方法再持久化。
