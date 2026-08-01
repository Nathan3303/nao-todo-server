# Checklist

## 1. 任务枚举统一到领域层

- [x] `domain/task/entities/task_enums.go` 中 `TaskStatePending=1` / `TaskStateInProgress=2` / `TaskStateCompleted=3`，与数据库现存数据语义一致
- [x] `TaskStateGivenUp` 与 `TaskStateArchived` 已删除，全项目无残留引用
- [x] `ParseTaskState` / `ParseTaskPriority` / `ParseRemindRepeat` 对非法输入返回失败标识，不静默降级为合法值
- [x] `(TaskState) String()` / `(TaskPriority) String()` / `(RemindRepeat) String()` 输出的字符串与改造前 HTTP 响应完全一致（`todo` / `in-progress` / `done`、`low` / `medium` / `high` / `urgent`、`none` / `daily` / `weekly` / `monthly`）
- [x] `WeekdaysToBitmask` 与 `BitmaskToWeekdays` 位于 `domain/task/entities`，往返转换结果一致
- [x] `task_enums_test.go` 存在且覆盖双向转换、非法输入、星期掩码往返三类场景，测试通过
- [x] `application/task/converters.go` 已移除 `consts` import 与本地私有的星期转换函数
- [x] `infrastructure/persistence/task/scopes.go` 已移除 `consts` import，解析失败的取值被跳过而非当作 0 处理
- [x] `consts/` 目录已删除，全项目 Grep `naotodoserver/consts` 零命中

## 2. 应用层不再依赖 gin

- [x] 头像存储端口定义中不含任何 HTTP 或 gin 类型，入参为 `io.Reader` + 文件名 + 大小
- [x] 基础设施层提供本地文件系统实现，`os.MkdirAll` / 文件写入 / `os.Remove` 均在该实现内
- [x] `application/user/app.go` 与 `appImpl.go` 均无 `github.com/gin-gonic/gin` import
- [x] Grep 验证 `application/` 全目录无 `gin-gonic` 引用
- [x] `application/user/appImpl.go` 无 `os` 直接调用（`path/filepath` 仅保留 `Ext` 做扩展名校验，属纯字符串处理）
- [x] 文件大小上限与扩展名白名单校验仍保留在应用层（业务规则未丢失）
- [x] 数据库更新失败时已落盘文件被清理的行为保持不变
- [x] `interfaces/controllers/user.go` 中头像上传的错误码 10081 / 10082 与响应结构未变更
- [x] `infrastructure/initialize.go` 已注入头像存储实现，程序可正常启动

## 3. 事务管理端口

- [x] `domain/types/` 中定义了 `TxManager` 接口，签名不含 gorm 或任何数据库类型
- [x] `infrastructure/persistence/dbs/` 提供 `TxManager` 实现与 `DBFrom` helper
- [x] `DBFrom` 在 context 中存在事务句柄时返回该句柄，否则返回传入的 fallback
- [x] 所有仓储接口（`domain/*/repositories/*.go`）签名未因事务改造而变更
- [x] `infrastructure/initialize.go` 已构造并注入 `TxManager`

## 4. 清单级联操作原子性

- [x] `application/project/appImpl.go` 的 `Delete` / `Restore` / `Archive` / `Unarchive` 四个方法均由 `txManager.Do` 包裹
- [x] 四个方法中级联失败时返回错误，不存在 `log.Printf` 后继续返回 `nil` 的路径
- [x] `Archive` 不再直接调用 `app.repo.Archive`，改为经 `app.projectDomain.Archive`（领域服务委托 repo，与 Delete/Restore 模式一致）
- [x] `HardDelete` 不再包含 `panic("unimplemented")`，方法及接口声明已删除（无路由挂载）
- [x] 因改动而失效的 `log` import 已移除

## 5. 已注销用户清理事务化

- [x] `DeleteDeactivatedUsers` 的全部删除操作包裹在单一 `db.Transaction` 内
- [x] 原有的 11 张关联表删除顺序保持不变
- [x] 任一删除步骤返回错误时整个事务回滚并向上返回错误
- [x] 缓存失效操作在事务提交成功之后执行

## 6. user 域 DTO 解耦

- [x] `application/user/dto/` 已建立，结构体字段不含 `json` / `form` tag
- [x] `application/user/` 下所有 Go 文件 Grep `naotodoserver/interfaces/types` 零命中
- [x] `UserApp` 接口全部方法签名仅使用 `dto` 包类型或 `domain/types` 类型
- [x] `interfaces/controllers/user.go` 承担 HTTP DTO ↔ app DTO 转换
- [x] 用户昵称、资料、密码、头像、注销、恢复、配置各接口的 JSON 字段名与错误码与改造前一致
- [x] 头像 URL 拼接行为（`conf.Conf.Uploads.AvatarURL`）未改变
- [x] cron 调用的 `DeleteDeactivatedUsers(ctx, dayOffset)` 仍可正常工作

## 7. auth 域 DTO 解耦

- [x] `application/auth/dto/` 已建立
- [x] `application/auth/` 下所有 Go 文件 Grep `naotodoserver/interfaces/types` 零命中
- [x] `AuthApp` 接口全部方法签名仅使用 `dto` 包类型或 `domain/types` 类型
- [x] `interfaces/controllers/identity.go` 承担 HTTP DTO ↔ app DTO 转换
- [x] 登录、注册、登出、登录状态检查各接口的 JSON 契约与错误码与改造前一致

## 8. 整体验证

- [x] `go build ./...` 通过
- [x] `go vet ./...` 无告警
- [x] `go test ./...` 全部通过（含改造前已有的 domain 层单元测试）
- [x] `golangci-lint run` 无新增告警（本次引入的 `project/appImpl.go` 超长行已修复；其余告警为仓库既有问题：全仓 CRLF 行尾导致的 gofmt 误报、`task/repoImpl.go` 函数签名超长、`tag/appImpl.go` prealloc，均在改动前已存在）
- [x] 新增与修改的代码遵循项目现有注释风格与命名风格
- [x] 本次改动未触及 `task` / `project` / `tag` / `pomodoro` 四域对 `interfaces/types` 的依赖（超出本 spec 范围）
