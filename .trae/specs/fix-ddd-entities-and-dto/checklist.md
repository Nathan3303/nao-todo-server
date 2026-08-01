# Checklist

## 1. Project 实体状态迁移

- [x] `domain/project/entities/project.go` 存在 `Archive()/Unarchive()/Delete()/Restore()`，直接修改 `ArchivedAt`/`DeactivedAt`，行为幂等
- [x] `repositories.Project` 新增 `UpdateState`，实现仅更新 `ShouldUpdate()` 的状态字段（含 IsNull 清空分支）
- [x] `repoImpl.go` 的 `Archive/Unarchive/Delete/Restore` 已移除（接口与实现），无残留引用
- [x] `ProjectDomainImpl.Archive/Unarchive/Delete/Restore` 均经 `repo.GetById` → 实体方法 → `repo.UpdateState`，`Delete` 仍同步删除偏好
- [x] `RowsAffected==0` 时 `UpdateState` 返回"清单不存在"错误（保持原行为）
- [x] `project_test.go` 覆盖四个状态迁移方法，测试通过

## 2. Task 实体状态迁移

- [x] `Task.ChangeState(next)` 实现状态机：同状态幂等、进入 Completed 设置 `CompletedAt`、离开 Completed 清空、非法值报错
- [x] `Task.Archive()/Unarchive()/ToggleStar()` 存在且幂等
- [x] `valueobjects.UpdateTask` 与 `NewUpdateTask` 支持 `CompletedAt`
- [x] `UpdateTaskValueObjectToMap` 支持 `CompletedAt` 落库（ShouldUpdate/IsSetToNull 分支）
- [x] `UpdateTask` 携带 `State` 时走读-改-写（`repo.GetById` → `entity.ChangeState` → 以迁移后 State/CompletedAt 落库）；未携带 `State` 时不额外读库
- [x] `task_test.go` 覆盖状态机与时间戳维护，测试通过

## 3. TaskCheckItem 死代码接入

- [x] `UpdateTaskCheckItem` 与 `BatchUpdateTaskCheckItems` 在 `IsDone != nil` 时经 `IsCompleted()`/`MarkDone()`/`MarkUndone()` 迁移
- [x] `MarkDone|MarkUndone|IsCompleted|ChangeState` 在 `application/` 存在真实调用点（Grep 命中 application/task/appImpl.go L128/L398/L400/L402/L494/L496/L498）
- [x] 勾选/取消勾选（单个与批量）落库正确，响应字段与改造前一致

## 4. tag 域 DTO 解耦

- [x] `application/tag/dto/` 已建立，字段不含 json/form tag
- [x] `application/tag/` 下 Grep `naotodoserver/interfaces/types` 零命中
- [x] `TagApp` 接口签名仅使用 dto 类型或 `domain/types` 类型
- [x] `interfaces/controllers/tag.go` 承担 HTTP ↔ app 转换，JSON 契约与错误码不变（含 GetTagPreference 响应经 `toGetTagPreferenceRes` 转换）

## 5. pomodoro 域 DTO 解耦

- [x] `application/pomodoro/dto/` 已建立，字段不含 json/form tag
- [x] `application/pomodoro/` 下 Grep `naotodoserver/interfaces/types` 零命中
- [x] `PomodoroApp` 接口签名仅使用 dto 类型或 `domain/types` 类型
- [x] `interfaces/controllers/pomodoro.go` 承担转换，JSON 契约与错误码不变

## 6. project 域 DTO 解耦

- [x] `application/project/dto/` 已建立，字段不含 json/form tag
- [x] `application/project/` 下 Grep `naotodoserver/interfaces/types` 零命中
- [x] `ProjectApp` 接口签名仅使用 dto 类型或 `domain/types` 类型
- [x] `DeleteDeactivatedProjects` 签名更新后 cron 调用点同步可用（签名未变，`infrastructure/cron/deleteInactiveProject.go` 兼容）
- [x] `interfaces/controllers/project.go` 承担转换，JSON 契约与错误码不变

## 7. task 域 DTO 解耦

- [x] `application/task/dto/` 已建立（Task/TaskCheckItem/TaskComment 三套），字段不含 json/form tag
- [x] `application/task/` 下 Grep `naotodoserver/interfaces/types` 零命中
- [x] `TaskApp` 与 `TaskCheckItemApp` 接口签名仅使用 dto 类型或 `domain/types` 类型
- [x] `interfaces/controllers/task.go`（含 `event.go`、`comment.go` 调用点）承担转换，JSON 契约与错误码不变

## 8. 整体验证

- [x] `go build ./...` 通过
- [x] `go vet ./...` 无告警
- [x] `go test ./...` 全部通过（含新增 `domain/project/entities`、`domain/task/entities` 测试）
- [x] `golangci-lint run` 无新增告警（本次引入的 controller 转换函数超长行、`serviceImpl.go` UpdateState 超长行均已修复；剩余为仓库既有问题：全仓 CRLF gofmt 误报、`task/repoImpl.go` 函数签名超长、`tag/appImpl.go` prealloc）
- [x] 新增与修改代码遵循项目现有注释与命名风格
- [x] `interfaces/types/` 下四域 HTTP 契约文件未做破坏性修改（仅 controller 转换引用；抽查确认 controller 输出均经转换函数，无 dto 直出丢失 json tag）
