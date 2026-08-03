# Checklist: 切断 application → infrastructure 依赖并补齐 Task 实体充血化

## Part A:切断 application → infrastructure

- [x] Task 1: `domain/types/notification_publisher.go` 新增端口
- [x] Task 2: `infrastructure/sse/notificationPublisher.go` 新增实现
- [x] Task 3: `application/task/appImpl.go` 切到 publisher 端口
- [x] Task 4: task 域 userId 显式参数化
- [x] Task 5: project 域 userId 显式参数化
- [x] Task 6: tag 域 userId 显式参数化
- [x] Task 7: pomodoro 域 userId 显式参数化
- [x] Task 8: user/auth 域 userId 显式参数化(auth 域方法不依赖 userId,仅 user 域改造)
- [x] Task 9: controllers 适配 userId 注入
- [x] Task 10: infrastructure wiring(`initialize.go` 注入 `notificationPublisher`)
- [x] Task 11: 验证 application 完全无 infrastructure import(`grep` 零结果)

## Part B:Task 实体充血化补齐

- [x] Task 12: `Task.Archive/Unarchive/ToggleStar` 签名改造(Archive/ToggleStar 接受 `at *time.Time`)
- [x] Task 13: `Task.GiveUp/UngiveUp` 实体方法
- [x] Task 14: `task_test.go` 单元测试(新增 `TestTaskGiveUp` / `TestTaskUngiveUp`,更新 `TestTaskArchive` / `TestTaskToggleStar`)
- [x] Task 15: UpdateTask 读-改-写(`req.ArchivedAt/StarMarkAt/GivenUpAt != nil` 触发,共用一次 `GetById`)
- [x] Task 16: UpdateTask 挂 IsDatesValid(读-改-写完成后校验)
- [x] Task 17: 4 个 ByProjectId 充血化(SoftDeleteByProjectId / RestoreByProjectId / ArchiveByProjectId / UnarchiveByProjectId 内部走实体方法)
- [x] Task 18: build/vet/test/lint 全绿

## 验证

- [x] `go build ./...` 0 错误
- [x] `go vet ./...` 0 告警
- [x] `go test -count=1 ./...` 全通过(含 task 实体测试)
- [x] `golangci-lint run` 零新增告警(残留 3 处 gofmt 误报为仓库既有 CRLF 问题,git status 确认未改动)
- [x] `grep -r "naotodoserver/infrastructure" application/` 零结果
- [x] `application/task/appImpl.go ProcessReminders` 通过 `NotificationPublisher` 推送
- [x] 4 个 ByProjectId 批量方法实现内含 `task.Archive` / `task.Unarchive` 实体方法调用
- [x] `Task.IsDatesValid` 在 UpdateTask 写路径被调用(读-改-写分支顺带触发)

## 文档

- [x] checklist.md 勾选完成
- [x] commit 提交
