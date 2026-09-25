# 阶段一：数据层打通

## 实现内容

- 数据库 model 新增 `RemindRepeat`、`RemindTime`、`RemindWeekdays` 字段
- domain entity/VO 补全提醒相关字段
- infrastructure converters 补全 model↔entity 映射，修复 `RemindAt` 和 `ParentTaskId` 漏映射
- API 类型新增提醒字段：`CreateTaskReq`/`UpdateTaskReq`/`GetTaskRes`
- 新增 consts 映射：`RemindRepeatMap`、`WeekdayBitmask`（星期↔位掩码）
- 新增 Snooze 端点：controller + types + router + application + domain service + repo

## 受影响文件

- `consts/task.go`
- `domain/task/entities/task.go`
- `domain/task/valueobjects/createTask.go`
- `domain/task/valueobjects/updateTask.go`
- `domain/task/service/service.go`
- `domain/task/service/serviceImpl.go`
- `domain/task/repositories/task.go`
- `infrastructure/persistence/models/task.go`
- `infrastructure/persistence/task/converters.go`
- `infrastructure/persistence/task/repoImpl.go`
- `application/task/app.go`
- `application/task/appImpl.go`
- `application/task/converters.go`
- `interfaces/types/task.go`
- `interfaces/types/snooze.go`
- `interfaces/controllers/task.go`
- `interfaces/routers/taskRouter.go`
