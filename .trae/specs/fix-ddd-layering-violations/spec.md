# 修复 DDD 分层违规（第一批）Spec

## Why

项目 DDD 分层骨架标准（`domain/` 零 gorm/gin/redis 依赖），但存在四类高危问题：任务状态枚举在 `consts/` 与 `domain/` 中存在**值错位**（正确性缺陷）、`gin.Context` 泄漏进应用层公开接口、应用层反向依赖 `interfaces/types`、跨聚合级联操作无事务且静默吞错。本次集中处理这四项风险最高、改动边界最清晰的问题。

## What Changes

### 1. 统一任务状态与优先级枚举（正确性修复）

- `consts/task.go` 定义 `"todo":1, "in-progress":2, "done":3`，而 `domain/task/entities/task_enums.go` 定义 `TaskStatePending=0, TaskStateInProgress=1, TaskStateCompleted=2`，**整体错位一位**。DB 模型注释（`infrastructure/persistence/models/task.go:26-28`）与 `consts` 一致，说明 DB 已存数据遵循 `consts` 语义。
- 决策：**以 DB 现状（1/2/3）为准**，修正 `domain` 常量值，零数据迁移、零 API 变更。
- 将 `consts/task.go` 的字符串↔数值映射迁入 `domain/task/entities`，改为挂在枚举类型上的 `ParseTaskState(string)` / `(TaskState) String()` 方法。
- 删除经确认无效的 `TaskStateGivenUp` / `TaskStateArchived`（放弃与归档实际由 `GivenUpAt` / `ArchivedAt` 时间戳表达，`state` 列无对应取值）。
- 提醒重复类型（`RemindRepeatMap`）同样迁入 `domain/task/entities`，定义 `RemindRepeat` 类型。
- 星期位掩码（`WeekdayBitmask`）是任务提醒的领域规则，迁入 `domain/task/entities`。
- 删除空文件 `consts/user.go`，`consts/` 包整体移除。
- **BREAKING**（内部）：`consts` 包被删除，所有引用点需改为 domain 枚举 API。对外 HTTP 契约不变。

### 2. 消除 `gin.Context` 对应用层的泄漏

- `application/user/app.go:37-40` 的 `UpdateAvatarByFile(ctxRaw *gin.Context, ctx context.Context)` 是全项目唯一的 Web 框架泄漏点。
- controller 层负责解析 multipart 表单，向应用层传递与框架无关的参数。
- 文件落盘操作抽象为端口，实现下沉到 `infrastructure/`。
- **BREAKING**（内部）：`UserApp.UpdateAvatarByFile` 签名变更。

### 3. 切断 `application` 对 `interfaces/types` 的依赖（限 user + auth 两域）

- 全项目 20 处 `application/*` import `interfaces/types`，导致应用层被 HTTP DTO 绑死，cron / SSE / 未来 gRPC 无法干净复用。
- 本次仅改造 `user` 与 `auth` 两域，建立可复制的模式样板；`task` / `project` / `tag` / `pomodoro` 四域另开 spec 推进。
- 引入 `application/user/dto` 与 `application/auth/dto`，转换职责拆为两层：`interfaces` DTO ↔ app DTO 由 controller 负责，app DTO ↔ domain VO 由应用层 converters 负责。
- **BREAKING**（内部）：`UserApp` / `AuthApp` 全部方法签名变更。

### 4. 为跨聚合级联操作建立事务边界

- `application/project/appImpl.go` 的 `Delete` / `Restore` / `Archive` / `Unarchive` 四处均为「先改 Project，再级联改 Task」的两步非原子操作，且级联失败仅 `log.Printf` 不回滚（见 `appImpl.go:150-153, 185-187, 235-237, 268-270`），数据可永久停留在「Project 已删、Task 未删」的脏状态。
- `infrastructure/persistence/identity/userRepoImpl.go:292-366` 的 `DeleteDeactivatedUsers` 跨 11 张表删除且全程无事务，任一步失败即产生孤儿数据。
- 决策：**context 传递 tx + TxManager 端口**。`domain` 定义 `TxManager` 接口，`infrastructure` 用 `gorm.Transaction` 实现并将 `*gorm.DB` 注入 context；各 repoImpl 增加取值 helper，优先使用 context 中的事务句柄。仓储接口签名零变更。
- `application/project/appImpl.go:202-207` 的 `HardDelete` 当前直接 `panic("unimplemented")`，一并处理。

## Impact

- **Affected specs**: 任务状态管理、任务提醒、用户头像上传、用户注销、清单删除/恢复/归档、登录认证
- **Affected code**:
  - 删除：`consts/`
  - 新增：`domain/task/entities/task_enums.go`（扩充）、`domain/types/tx.go`、`infrastructure/persistence/dbs/tx.go`、`application/user/dto/`、`application/auth/dto/`、头像存储端口与实现
  - 修改：`application/task/converters.go`、`infrastructure/persistence/task/scopes.go`、`application/user/*`、`application/auth/*`、`application/project/appImpl.go`、`interfaces/controllers/user.go`、`interfaces/controllers/identity.go`、`infrastructure/persistence/identity/userRepoImpl.go`、各 `repoImpl.go`、`infrastructure/initialize.go`

## ADDED Requirements

### Requirement: 任务状态枚举单一权威

系统 SHALL 在 `domain/task/entities` 中提供任务状态、优先级、提醒重复类型的唯一权威定义，包含字符串与数值之间的双向转换能力。

#### Scenario: 状态字符串转数值
- **WHEN** 应用层收到 `state="todo"` 的请求
- **THEN** 转换结果为 `TaskStatePending`，其底层数值为 `1`，与数据库现存数据语义一致

#### Scenario: 状态数值转字符串
- **WHEN** 从数据库读出 `state=3` 的任务
- **THEN** 响应中 `state` 字段为 `"done"`

#### Scenario: 非法状态字符串
- **WHEN** 传入未定义的状态字符串
- **THEN** 转换函数返回零值与失败标识，调用方可据此拒绝请求，不得静默转为某个合法状态

### Requirement: 提醒周期与星期掩码归属领域层

系统 SHALL 将提醒重复类型与星期位掩码的转换规则定义在 `domain/task/entities` 中。

#### Scenario: 星期数组与掩码互转
- **WHEN** 传入星期数组 `[1, 3, 5]`
- **THEN** 得到对应位掩码，且反向转换可还原为同一集合

### Requirement: 头像文件存储端口

系统 SHALL 定义与 Web 框架无关的头像存储抽象，由基础设施层提供本地文件系统实现。

#### Scenario: 应用层不感知 gin
- **WHEN** 检索 `application/` 目录下所有 Go 文件的 import
- **THEN** 不存在任何对 `github.com/gin-gonic/gin` 的引用

### Requirement: 事务管理端口

系统 SHALL 提供事务管理抽象，使应用层能够将多个仓储操作包裹在单一事务中，且仓储接口签名不因此变更。

#### Scenario: 级联操作原子性
- **WHEN** 删除清单时级联软删除任务的步骤失败
- **THEN** 清单自身的删除一并回滚，接口返回错误，数据库中清单与任务状态保持删除前的一致状态

#### Scenario: 事务内仓储复用
- **WHEN** 应用层在一个事务中先后调用 Project 仓储与 Task 仓储
- **THEN** 两次调用使用同一数据库事务句柄，无需为仓储方法额外传入事务参数

## MODIFIED Requirements

### Requirement: 用户头像文件上传

系统 SHALL 支持用户通过 multipart 表单上传头像文件，校验文件大小与类型后落盘，并更新用户头像地址与历史评论中的头像快照。

接口层负责从 HTTP 请求中解析出文件内容与原始文件名；应用层接收与框架无关的参数，不得直接访问 HTTP 请求对象、不得直接执行文件系统调用。

#### Scenario: 上传合法头像
- **WHEN** 用户以 multipart 表单提交 JPG/JPEG/PNG 且不超过配置上限的文件
- **THEN** 文件保存成功，返回可访问的头像 URL，评论中的头像同步更新

#### Scenario: 文件类型不支持
- **WHEN** 用户上传扩展名非 JPG/JPEG/PNG 的文件
- **THEN** 请求被拒绝并返回类型不支持的错误，不产生任何落盘文件

#### Scenario: 落盘成功但更新数据库失败
- **WHEN** 头像文件已保存但用户头像字段更新失败
- **THEN** 已保存的文件被清理，接口返回错误

### Requirement: 清单删除与恢复

系统 SHALL 在删除清单时级联软删除其下所有任务，在恢复清单时级联恢复其下所有任务；两者 SHALL 在同一事务内完成。

#### Scenario: 删除清单成功
- **WHEN** 用户删除一个包含任务的清单
- **THEN** 清单及其偏好设置、其下所有任务在同一事务内被软删除

#### Scenario: 级联步骤失败
- **WHEN** 级联软删除任务的过程中发生数据库错误
- **THEN** 整个事务回滚，接口返回错误，清单未被删除

### Requirement: 清单归档与取消归档

系统 SHALL 在归档清单时级联归档其下所有任务，在取消归档时级联取消归档；两者 SHALL 在同一事务内完成，且失败时返回错误而非静默记录日志。

#### Scenario: 归档清单成功
- **WHEN** 用户归档一个包含任务的清单
- **THEN** 清单与其下所有任务在同一事务内被标记归档

### Requirement: 定时清理已注销用户

系统 SHALL 在删除已注销用户时，于单一事务内按依赖顺序清理其在所有关联表中的数据。

#### Scenario: 清理过程中某表删除失败
- **WHEN** 清理过程中任一表的删除操作失败
- **THEN** 整个清理事务回滚，不产生孤儿数据，错误被返回给调用方

### Requirement: 应用层接口与 HTTP 契约解耦（user + auth 域）

`UserApp` 与 `AuthApp` 的方法签名 SHALL 仅使用应用层自有 DTO 或领域类型，不得引用 `interfaces/types` 中的类型。

#### Scenario: 依赖方向检查
- **WHEN** 检索 `application/user/` 与 `application/auth/` 下所有 Go 文件的 import
- **THEN** 不存在对 `naotodoserver/interfaces/types` 的引用

#### Scenario: HTTP 契约保持不变
- **WHEN** 客户端调用登录、注册、获取用户资料、更新头像等接口
- **THEN** 请求与响应的 JSON 字段名、类型、错误码与改造前完全一致

## REMOVED Requirements

### Requirement: consts 包作为枚举映射来源

**Reason**: 枚举映射属于领域概念，置于技术包 `consts/` 导致与 `domain` 中的枚举定义并存且值错位，是数据正确性隐患的根源。`consts/user.go` 更是空文件。

**Migration**: 全部映射迁入 `domain/task/entities`，改为枚举类型上的方法。`consts` 包的两个引用点（`application/task/converters.go`、`infrastructure/persistence/task/scopes.go`）改用新 API 后删除整个包。

### Requirement: 任务状态包含「已放弃」与「已归档」取值

**Reason**: `state` 列实际不存储这两种取值，放弃与归档由 `GivenUpAt` / `ArchivedAt` 时间戳字段独立表达，查询也走 `ByTaskGivenUpFlag` 等独立 scope。保留这两个常量会误导后续开发者。

**Migration**: 删除 `TaskStateGivenUp` 与 `TaskStateArchived` 常量。若后续确需以状态机表达，另行设计。
