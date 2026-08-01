# Tasks

- [x] Task 1: Project 实体状态迁移方法 + 领域服务接入
  - [x] 1.1 `domain/project/entities/project.go` 新增 `Archive()/Unarchive()/Delete()/Restore()`：直接修改 `ArchivedAt`/`DeactivedAt`（归档=当前时间、取消归档/恢复=清空），行为幂等，与现有 API 一致
  - [x] 1.2 `domain/project/repositories/project.go` 新增 `UpdateState(ctx, userId, projectId int64, archivedAt, deactivedAt types.NullableTime) error`
  - [x] 1.3 `infrastructure/persistence/project/repoImpl.go` 实现 `UpdateState`（仅更新 `ShouldUpdate()` 的状态字段，`RowsAffected==0` 返回"清单不存在"，失败后失效缓存）
  - [x] 1.4 确认 `repo.Archive/Unarchive/Delete/Restore` 无其他调用方后，从接口与实现中移除，删除对应私有 SQL 辅助（如无复用）
  - [x] 1.5 `domain/project/service/serviceImpl.go` 的 `Archive/Unarchive/Delete/Restore` 改为：`repo.GetById` 加载实体 → 实体状态迁移方法 → `repo.UpdateState` 持久化；`Delete` 保留同步删除偏好逻辑
  - [x] 1.6 新增 `domain/project/entities/project_test.go`：四方法的时间戳设置/清空行为
  - 验证：`go build ./...`；`go test ./domain/project/...`；归档/恢复/删除/取消归档接口行为与改造前一致

- [x] Task 2: Task 实体状态迁移 + UpdateTask 读-改-写
  - [x] 2.1 `domain/task/entities/task.go` 新增 `ChangeState(next TaskState) error`：同状态幂等；进入 Completed 设置 `CompletedAt=now`；离开 Completed 清空 `CompletedAt`；非法值返回错误
  - [x] 2.2 `Task` 新增 `Archive()/Unarchive()/ToggleStar()`：设置/清空 `ArchivedAt`、切换 `StarMarkAt`（幂等）
  - [x] 2.3 `domain/task/valueobjects/updateTask.go` 新增 `CompletedAt types.NullableTime` 字段 + `NewUpdateTask` 参数透传
  - [x] 2.4 `infrastructure/persistence/task/converters.go` 的 `UpdateTaskValueObjectToMap` 支持 `CompletedAt` 落库（复用 StartAt/EndAt 的 ShouldUpdate/IsSetToNull 分支）
  - [x] 2.5 `application/task/converters.go` 的 `UpdateTaskReqToValueObject` 透传 `CompletedAt`（默认不设置）
  - [x] 2.6 `application/task/appImpl.go` 的 `UpdateTask`：请求携带 `State` 时先 `repo.GetById`（includeDeleted=false）加载实体 → `entity.ChangeState(parsed)` → 以实体迁移后 `State` + `CompletedAt` 覆盖构造 VO → `repo.Update`；未携带 `State` 时保持原路径
  - [x] 2.7 新增 `domain/task/entities/task_test.go`：状态机合法/非法迁移、CompletedAt 设置与清空、Archive/Unarchive/ToggleStar
  - 验证：`go build ./...`；`go test ./domain/task/entities/...`；更新任务 State 接口落库 `state` 与 `completed_at` 正确
  - 依赖：无（与 Task 1 并行）

- [x] Task 3: TaskCheckItem 死代码方法接入生产路径
  - [x] 3.1 `application/task/appImpl.go` 的 `UpdateTaskCheckItem`：请求 `IsDone != nil` 时 `GetCheckItemById` 加载实体，`IsCompleted()` 与目标不同则 `MarkDone()`/`MarkUndone()`，以实体新值组装 VO 后 `repo.UpdateCheckItem`
  - [x] 3.2 `BatchUpdateTaskCheckItems` 同样接入：每项 `IsDone != nil` 时经实体方法迁移（可先 `ListCheckItems` 一次加载后按 ID 应用）
  - [x] 3.3 删除上述路径中不再使用的私有转换逻辑（如有）
  - 验证：`go build ./...`；勾选/取消勾选检查项（单个与批量）落库正确，响应字段不变；`ToggleDone` 在迁移路径中真实被调用
  - 依赖：无（与 Task 1/2 并行）

- [x] Task 4: tag 域应用层 DTO 解耦
  - [x] 4.1 新建 `application/tag/dto/`（GetTagRes/CreateTagReq|Res/UpdateTagReq/ListTagRes/BatchUpdateTagReq|Res/GetTagPreferenceRes/UpdateTagPreferenceReq 等），无 json/form tag
  - [x] 4.2 `application/tag/app.go|appImpl.go|converters.go` 签名改用 dto，converter 中 app DTO ↔ domain VO 转换保留在应用层
  - [x] 4.3 `interfaces/controllers/tag.go` 承担 HTTP DTO ↔ app DTO 转换，JSON 契约与错误码不变
  - 验证：`application/tag/` Grep `naotodoserver/interfaces/types` 零命中；`go build ./...`
  - 依赖：无（独立，可与 Part A 并行）

- [x] Task 5: pomodoro 域应用层 DTO 解耦
  - [x] 5.1 新建 `application/pomodoro/dto/`（Create/Get/List PomodoroRecord、Create/Get/Update/Delete/Archive/Unarchive/List Pomodoro），无 json/form tag
  - [x] 5.2 `application/pomodoro/app.go|appImpl.go|converters.go` 签名改用 dto
  - [x] 5.3 `interfaces/controllers/pomodoro.go` 承担 HTTP DTO ↔ app DTO 转换，JSON 契约与错误码不变
  - 验证：`application/pomodoro/` Grep 零命中；`go build ./...`
  - 依赖：无（独立）

- [ ] Task 6: project 域应用层 DTO 解耦
  - [ ] 6.1 新建 `application/project/dto/`（GetProjectRes/CreateProjectReq|Res/UpdateProjectReq/BatchUpdateProjectReq|Res/ListProjectRes/GetProjectPreferenceRes/UpdateProjectPreferenceReq 等），无 json/form tag
  - [ ] 6.2 `application/project/app.go|appImpl.go|converters.go` 签名改用 dto；`DeleteDeactivatedProjects` 签名同步（cron 调用点同步更新）
  - [ ] 6.3 `interfaces/controllers/project.go` 承担 HTTP DTO ↔ app DTO 转换，JSON 契约与错误码不变
  - 验证：`application/project/` Grep 零命中；`go build ./...`
  - 依赖：Task 1（同文件被实体充血化改动）

- [x] Task 7: task 域应用层 DTO 解耦
  - [x] 7.1 新建 `application/task/dto/`：Task（GetTaskRes/CreateTaskReq/UpdateTaskReq/ListTaskReq|Res/SnoozeTaskReq|Res/Pagination）+ TaskCheckItem（Get/Create/Update/Batch）+ TaskComment（TaskCommentRes/Create/Update）三套，无 json/form tag
  - [x] 7.2 `application/task/app.go|appImpl.go|converters.go|comment_app.go|checkitem_app.go` 签名改用 dto
  - [x] 7.3 `interfaces/controllers/task.go` 承担 HTTP DTO ↔ app DTO 转换，JSON 契约与错误码不变
  - 验证：`application/task/` Grep 零命中；`go build ./...`
  - 依赖：Task 2、Task 3（同文件被实体充血化改动）

- [x] Task 8: 全量验证
  - [x] 8.1 `go build ./...`、`go vet ./...`、`go test ./...` 全部通过
  - [x] 8.2 `golangci-lint run`（项目内缓存）无新增告警
  - [x] 8.3 Grep 验证：`application/{tag,pomodoro,project,task}/` 无 `naotodoserver/interfaces/types`；`MarkDone|MarkUndone|IsCompleted|ChangeState` 在 `application/` 有真实调用点
  - [x] 8.4 抽查各域 controller 转换函数：JSON 字段名、错误码与 `interfaces/types` 原契约一致
  - [x] 8.5 新增/修改代码遵循项目现有注释与命名风格（中文注释、`@param`/`@return` 风格）
  - 依赖：Task 1-7

# Task Dependencies

- Task 6 (project DTO) depends on Task 1 (project 实体充血化)
- Task 7 (task DTO) depends on Task 2、Task 3 (task 实体充血化)
- Task 8 依赖全部任务
- Task 1 / Task 2 / Task 3 / Task 4 / Task 5 相互独立，可并行
