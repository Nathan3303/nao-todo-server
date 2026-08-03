# Tasks: 切断 application → infrastructure 依赖并补齐 Task 实体充血化

## Part A:切断 application → infrastructure(架构修复)

- [x] **Task 1**:新增 `domain/types/notification_publisher.go`
  - 定义 `NotificationPublisher` 接口与 `ReminderEvent` 载荷
  - 文件路径:`domain/types/notification_publisher.go`
  - 验收:`go build ./domain/types/...` 通过
- [x] **Task 2**:新增 `infrastructure/sse/notificationPublisher.go`
  - 实现 `NotificationPublisher`,内部用 `sse.GetHub().Publish`
  - 文件路径:`infrastructure/sse/notificationPublisher.go`
  - 验收:`go build ./infrastructure/sse/...` 通过
- [x] **Task 3**:改造 `application/task/appImpl.go` 切到端口
  - 删除 `import "naotodoserver/infrastructure/sse"`
  - `NewTaskApp` 构造函数增加 `publisher domaintypes.NotificationPublisher` 参数
  - `ProcessReminders` 改用 `publisher.PublishReminder`
  - 验收:task 域 app 编译通过
- [x] **Task 4**:Task 域 app 接口/实现 userId 显式参数化
  - `application/task/app.go` + `appImpl.go` 全部方法加 `userId int64` 入参
  - 删除 `import iCtx` 与每个方法开头的"获取用户 ID"块
  - 验收:task app 接口零 `iCtx` 引用
- [x] **Task 5**:Project 域 userId 显式参数化
  - `application/project/app.go` + `appImpl.go` 同上
- [x] **Task 6**:Tag 域 userId 显式参数化
  - `application/tag/app.go` + `appImpl.go` 同上
- [x] **Task 7**:Pomodoro 域 userId 显式参数化
  - `application/pomodoro/app.go` + `appImpl.go` 同上
- [x] **Task 8**:User/Auth 域 userId 显式参数化
  - `application/user/app.go` + `appImpl.go`(auth 域方法不依赖 userId,跳过)
- [x] **Task 9**:controllers 适配 userId 注入
  - 全部 `interfaces/controllers/*.go` 在调 app 前 `userId := iCtx.GetUserId(ctx.Request.Context())`,失败时 `Failure` 响应
  - 验收:`go build ./...` 通过
- [x] **Task 10**:infrastructure wiring
  - `infrastructure/initialize.go` 创建 `notificationPublisher`,注入到 `NewTaskApp`
  - 验收:`go build ./...` 通过
- [x] **Task 11**:验证 application 完全无 infrastructure import
  - `grep -r "naotodoserver/infrastructure" application/` 应零结果
  - 验收:命令执行零结果

## Part B:Task 实体充血化补齐

- [x] **Task 12**:改造 `Task.Archive/Unarchive/ToggleStar` 签名
  - `Archive(at *time.Time)` / `Unarchive()` / `ToggleStar(at *time.Time)`
  - `at == nil` 表示用 `time.Now()`
  - 文件路径:`domain/task/entities/task.go`
- [x] **Task 13**:新增 `Task.GiveUp/UngiveUp` 实体方法
  - 文件路径:`domain/task/entities/task.go`
  - 与 `Archive` 风格一致
- [x] **Task 14**:更新 `task_test.go` 实体单元测试
  - 更新 `TestTaskArchive` / `TestTaskToggleStar`(新签名 nil 入参)
  - 新增 `TestTaskGiveUp` / `TestTaskUngiveUp`
  - 文件路径:`domain/task/entities/task_test.go`
- [x] **Task 15**:UpdateTask 读-改-写改造
  - `application/task/appImpl.go UpdateTask` 增加三个分支:`req.ArchivedAt != nil` / `req.StarMarkAt != nil` / `req.GivenUpAt != nil`
  - 各分支:共用一次 `GetById` → 解析时间 → 调实体方法 → 回写 `vo`
  - 验收:仅当 req 携带上述字段时触发 GetById(性能),不携带时与改前零差异
- [x] **Task 16**:UpdateTask 挂 IsDatesValid
  - 读-改-写完成后调 `taskEntity.IsDatesValid()`,失败返回错误
  - 文件路径:`application/task/appImpl.go`
- [x] **Task 17**:infrastructure/persistence/task 4 个 ByProjectId 充血化
  - `SoftDeleteByProjectId/RestoreByProjectId/ArchiveByProjectId/UnarchiveByProjectId` 改为 Find → 调实体方法 → Save
  - 文件路径:`infrastructure/persistence/task/repoImpl.go`
  - 验收:行为等价,语义由"infra 直改列"变为"infra 走实体方法"
- [x] **Task 18**:验证全仓 build/test/lint
  - `go build ./...` 0 错误
  - `go vet ./...` 0 告警
  - `go test -count=1 ./...` 全通过
  - `golangci-lint run` 零新增告警
  - 验证 grep `naotodoserver/infrastructure` 在 `application/` 下零结果

## 文档

- [x] **Task 19**:更新 checklist.md 勾选所有完成项
- [x] **Task 20**:commit
  - commit message:`refactor(ddd架构收紧): 切断 application→infrastructure 依赖并补齐 Task 实体充血化`

## 状态

- [x] spec.md 已批准
- [x] tasks.md 已批准
- [x] 实施中 → 完成
- [x] 完成
