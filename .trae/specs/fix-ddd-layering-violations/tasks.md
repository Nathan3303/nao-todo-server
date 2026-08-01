# Tasks

- [x] Task 1: 统一任务枚举到领域层：将 `consts/` 中的任务状态、优先级、提醒重复、星期掩码映射迁入 `domain/task/entities`，修正枚举值错位，删除 `consts` 包。
  - [x] SubTask 1.1: 扩充 `domain/task/entities/task_enums.go`：将 `TaskState` 常量值改为 `TaskStatePending=1` / `TaskStateInProgress=2` / `TaskStateCompleted=3`，删除 `TaskStateGivenUp` 与 `TaskStateArchived`；`TaskPriority` 保留现有 `None=0/Low=1/Medium=2/High=3/Urgent=4` 值（与 `consts` 一致，无需改动）
  - [x] SubTask 1.2: 在同文件添加 `ParseTaskState(s string) (TaskState, bool)` 与 `(TaskState) String() string`，`ParseTaskPriority` / `(TaskPriority) String()`，遵循项目现有注释风格（`// 函数名 说明` + `@param` / `@return`）
  - [x] SubTask 1.3: 新增 `RemindRepeat` 类型（`RemindRepeatNone=0` / `Daily=1` / `Weekly=2` / `Monthly=3`）及 `ParseRemindRepeat` / `String()`；迁入星期掩码转换 `WeekdaysToBitmask([]uint8) uint8` 与 `BitmaskToWeekdays(uint8) []uint8`
  - [x] SubTask 1.4: 为上述转换函数编写单元测试 `task_enums_test.go`，覆盖合法值双向转换、非法字符串返回 false、星期掩码往返一致性
  - [x] SubTask 1.5: 改写 `application/task/converters.go`：移除 `consts` import，`consts.TodoStateMap[req.State]` 等 6 处引用改用新 API；删除本地的 `weekdaysToBitmask` / `bitmaskToWeekdays` 私有函数，改调 entities 中的版本
  - [x] SubTask 1.6: 改写 `infrastructure/persistence/task/scopes.go`：移除 `consts` import，`consts.TodoStateMap` / `consts.TodoPriorityMap` 两处改用 `entities.ParseTaskState` / `ParseTaskPriority`，跳过解析失败的取值
  - [x] SubTask 1.7: 删除 `consts/task.go` 与 `consts/user.go`（整个 `consts` 目录），运行 `go build ./...` 确认无残留引用

- [x] Task 2: 消除应用层的 gin 依赖：将头像文件上传改造为框架无关的接口，文件落盘下沉到基础设施层。
  - [x] SubTask 2.1: 定义头像存储端口（放在 `application/user`，理由：技术协作者且仅 user 应用服务使用，避免为文件 IO 污染纯净的 domain 层），提供保存文件与删除文件两个能力，入参为 `io.Reader` + 文件名，不含任何 HTTP 类型
  - [x] SubTask 2.2: 在 `infrastructure/storage/avatarStorage.go` 实现该端口（本地文件系统），迁入原 `appImpl.go` 的 `os.MkdirAll` / 文件写入 / `os.Remove` 逻辑，配置读取沿用 `conf.Conf.Uploads`
  - [x] SubTask 2.3: 修改 `application/user/app.go`：`UpdateAvatarByFile` 签名去掉 `*gin.Context`，改为接收文件流、文件名、文件大小；移除 `gin` import
  - [x] SubTask 2.4: 修改 `application/user/appImpl.go`：注入头像存储端口，`UpdateAvatarByFile` 保留大小与扩展名校验（业务规则）但改用端口落盘，更新失败时通过端口清理文件；移除 `gin` / `os` import（`path/filepath` 仅保留 `Ext` 用于扩展名校验）
  - [x] SubTask 2.5: 修改 `interfaces/controllers/user.go`：controller 负责 `ctx.FormFile("avatar")` 与打开文件流，将解析结果传给应用层，确保原有错误码 10081/10082 与响应结构不变
  - [x] SubTask 2.6: 更新 `infrastructure/initialize.go` 中 `userApp.NewUserApp` 的构造参数以注入新端口实现
  - [x] SubTask 2.7: 运行 `go build ./...` 与 `go vet ./...`，并 Grep 确认 `application/` 下无 `gin-gonic` 引用

- [x] Task 3: 建立事务管理端口：让应用层能将跨聚合操作包裹在单一事务中，仓储接口签名保持不变。
  - [x] SubTask 3.1: 在 `domain/types/tx.go` 新增 `TxManager` 接口，暴露 `Do(ctx context.Context, fn func(ctx context.Context) error) error`，签名不含 gorm 或任何数据库类型
  - [x] SubTask 3.2: 在 `infrastructure/persistence/dbs/tx.go` 实现 `TxManager`：用 `db.Transaction` 开启事务，将 `*gorm.DB` 以私有 key 注入 `context`（含嵌套复用保护）；同文件提供 `DBFrom(ctx, fallback *gorm.DB) *gorm.DB` helper，存在事务句柄时优先返回
  - [x] SubTask 3.3: 修改 `infrastructure/persistence/project/repoImpl.go`（Delete/Restore/Archive/Unarchive）、`project/preferenceRepoImpl.go`（Delete/Restore）、`task/repoImpl.go`（SoftDeleteByProjectId/RestoreByProjectId/ArchiveByProjectId/UnarchiveByProjectId），将 `r.db.WithContext(ctx)` 替换为 `dbs.DBFrom(ctx, r.db).WithContext(ctx)`
  - [x] SubTask 3.4: 修改 `infrastructure/initialize.go`：构造 `TxManager` 实例（后续任务注入 projectApp）

- [x] Task 4: 用事务包裹清单级联操作，消除静默失败。
  - [x] SubTask 4.1: 修改 `application/project/appImpl.go` 的 `Delete`：用 `txManager.Do` 包裹「领域服务删除 Project」+「`taskRepo.SoftDeleteByProjectId`」，级联失败返回错误而非 `log.Printf`
  - [x] SubTask 4.2: 同样改造 `Restore`、`Archive`、`Unarchive` 三处；`Archive`/`Unarchive` 改为经 `app.projectDomain`（领域服务新增对应委托方法，模式与 Delete/Restore 一致）
  - [x] SubTask 4.3: 处理 `HardDelete`：`interfaces/routers/projectRouter.go` 未挂载该路由，连同 `ProjectApp` 接口声明一并删除，禁止保留 `panic`
  - [x] SubTask 4.4: 移除因上述改动而不再使用的 `log` import；修复事务闭包内 4 处超长行（lll）

- [x] Task 5: 为已注销用户清理加上事务，消除孤儿数据风险。
  - [x] SubTask 5.1: 改写 `infrastructure/persistence/identity/userRepoImpl.go` 的 `DeleteDeactivatedUsers`：用 `r.db.Transaction` 包裹全部 12 次删除，保持现有依赖顺序，任一步 `Error` 非空即返回并回滚；空 userIds 提前返回
  - [x] SubTask 5.2: 将缓存失效（`r.cache.Del`）移到事务提交成功之后执行，避免回滚后缓存已被清空

- [x] Task 6: user 域应用层 DTO 解耦：切断 `application/user` 对 `interfaces/types` 的依赖。
  - [x] SubTask 6.1: 新建 `application/user/dto/`，为 `UserApp` 的每个方法定义对应的入参/出参结构体，字段类型使用 Go 原生类型与 `domain/types` 类型，不带 `json` / `form` tag
  - [x] SubTask 6.2: 修改 `application/user/app.go`：全部方法签名改用 `dto` 包类型，移除 `interfaces/types` import
  - [x] SubTask 6.3: 修改 `application/user/appImpl.go` 与 `converters.go`：内部改用 `dto` 类型，`converters.go` 只保留 app DTO ↔ domain entity 的转换，移除 `interfaces/types` import
  - [x] SubTask 6.4: 在 `interfaces/controllers/user.go` 中补充 `interfaces/types` ↔ `application/user/dto` 的转换，确保所有 JSON 字段名、错误码、响应结构与改造前完全一致
  - [x] SubTask 6.5: 检查 `application/user/appImpl.go` 中对 `taskApp.SyncTaskCommentUserProfile` 的调用与 `conf` 引用是否仍成立，头像 URL 拼接（`conf.Conf.Uploads.AvatarURL`）保持原行为

- [x] Task 7: auth 域应用层 DTO 解耦：切断 `application/auth` 对 `interfaces/types` 的依赖。
  - [x] SubTask 7.1: 新建 `application/auth/dto/`，为 `SignIn` / `SignUp` / `CheckIn` / `SignOut` 定义入参与出参结构体
  - [x] SubTask 7.2: 修改 `application/auth/app.go` 与 `appImpl.go` / `converters.go` 使用 `dto` 类型，移除 `interfaces/types` import
  - [x] SubTask 7.3: 在 `interfaces/controllers/identity.go` 中补充两侧 DTO 转换，确保登录/注册/登出/校验接口的 JSON 契约与错误码零变更
  - [x] SubTask 7.4: 确认 `RateLimit(ctx, clientIP, limit)` 与 `Validate(ctx, token)` 两个方法本身未使用 `interfaces/types`，无需改动

- [x] Task 8: 全量验证。
  - [x] SubTask 8.1: 运行 `go build ./...`、`go vet ./...`、`go test ./...`，全部通过
  - [x] SubTask 8.2: 运行 `golangci-lint run`，无新增告警（本次引入的超长行已修复；其余为仓库既有 CRLF gofmt 误报与既有 lll/prealloc）
  - [x] SubTask 8.3: Grep 验证：`application/` 下无 `gin-gonic`；`application/user/` 与 `application/auth/` 下无 `interfaces/types`；全项目无 `naotodoserver/consts`；`application/project/appImpl.go` 无 `panic("unimplemented")`、无 `log.Printf`

# Task Dependencies

- Task 4 依赖 Task 3（需要 `TxManager` 与 `DBFrom` 就绪）
- Task 5 可与 Task 3 并行（直接使用 `r.db.Transaction`，不依赖 `TxManager`）
- Task 6 依赖 Task 2（`UpdateAvatarByFile` 签名会被两个任务同时触及，先完成 gin 解耦再做 DTO 迁移，避免冲突）
- Task 8 依赖全部前置任务
- 可并行：Task 1、Task 2、Task 3、Task 5 之间无相互依赖；Task 7 与 Task 1/3/4/5 无依赖
