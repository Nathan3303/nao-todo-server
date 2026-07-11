# 为 Task 新增 SortId 字段实现计划

## Summary

为 Task 领域（实体 / GORM 模型 / 值对象 / DTO / 转换器 / 仓库 / 领域服务 / 应用层）新增 `SortId uint16` 字段，用于任务列表排序。实现参考同项目已有的三处 SortId 实现（Project、Tag、以及同领域的 TaskCheckItem），其中 **TaskCheckItem 与 Task 同属 task 领域、复用同一批文件，是最贴近的参考**。

字段语义与既有实现保持一致：
- 创建任务时 `SortId = 当前用户最大 SortId + 1`（基数 255，无记录则从 256 起）。
- 支持通过 `UpdateTask` 单条更新 SortId。
- 列表可通过查询参数 `sort=sortId:asc` 按 SortId 排序（现有通用排序 Scope 已支持，无需改仓库 List 逻辑）。

## Current State Analysis

- 架构分层：`interfaces/controllers` → `interfaces/types`(DTO) → `application/task`(app + converters) → `domain/task/service`(领域服务) → `domain/task/repositories`(接口) → `infrastructure/persistence/task`(repoImpl + converters) → `infrastructure/persistence/models`(GORM 模型)。ORM 为 GORM(MySQL)，表结构由 `infrastructure/persistence/dbs/mysql.go` 的 `DoMigration()` 中 `DB.AutoMigrate(models.Task{}, ...)` 自动迁移。
- Task 当前**无** SortId 字段。
- TaskCheckItem 已完整实现 SortId：
  - 模型 [task.go:103](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/models/task.go#L99-L103) `SortId uint16 gorm:"default:0"`
  - 领域服务初始化 [serviceImpl.go:73](file:///home/nathan/Projects/nao-todo-server/domain/task/service/serviceImpl.go#L68-L75) `vo.SortId = repo.GetMaxCheckItemSortId(...) + 1`
  - 仓库 [repoImpl.go:415](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/task/repoImpl.go#L415-L423) `GetMaxCheckItemSortId`
  - 转换器 [converters.go:240/258/298](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/task/converters.go#L233-L300)
  - DTO [task.go:96/114](file:///home/nathan/Projects/nao-todo-server/interfaces/types/task.go#L90-L115) `SortId uint16/*uint16 json:"sortId"`
- **关键差异**：Task 的创建路径 [appImpl.go:59-80](file:///home/nathan/Projects/nao-todo-server/application/task/appImpl.go#L59-L80) 直接调用 `taskRepo.Create(...)`，**绕过领域服务**；`TaskDomain` 接口无 `CreateTask`。而 SortId 初始化按项目惯例应在领域服务完成。
- 列表排序 [repoImpl.go:181](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/task/repoImpl.go#L160-L206) 走通用 `query.Sort(q.Sort)`，字段名驼峰转下划线（`sortId`→`sort_id`），因此前端传 `sort=sortId:asc` 即可排序，无需改 List。

## Assumptions & Decisions

由于当前模式无法向用户提问，采用以下决策（均贴合既有模式、最小改动）：

1. **初始化落点：走领域服务**。在 `TaskDomain` 接口新增 `CreateTask` 方法，把 `SortId = GetMaxSortId(userId) + 1` 放入 `TaskDomainImpl.CreateTask`，并让 `TaskAppImpl.CreateTask` 改为调用领域服务。与 Project/Tag/CheckItem 完全一致。`Copy`（[serviceImpl.go:20-50](file:///home/nathan/Projects/nao-todo-server/domain/task/service/serviceImpl.go#L20-L50)）复用同一初始化逻辑，使复制出的任务也获得新 SortId。
2. **列表排序：仅支持查询参数**。不改 `List` 默认排序逻辑，前端传 `sort=sortId:asc` 即可。保持最小改动（Task 已有丰富的多字段排序场景，不宜强加默认 sort_id 排序）。
3. **批量重排：本次不新增**。仅通过 `UpdateTask` 支持单条 SortId 更新。若后续需要拖拽批量重排，可另行参考 `BatchUpdateProject` 扩展。
4. `SortId` 类型统一为 `uint16`，JSON tag 统一 `json:"sortId"`，GORM tag `gorm:"default:0"`。
5. `GetMaxSortId` 只按 `user_id` 维度聚合（与 Project/Tag 一致；Task 不按 project 维度分组）。
6. 更新时 SortId 采用指针 `*uint16` + map 写入，允许显式更新；不加“不能为 0”的校验（Task 侧无强约束需求，保持简单）。

## Proposed Changes

### 1. GORM 模型：新增 SortId 列
文件 [models/task.go](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/models/task.go#L74-L77)
在 `Task` 结构体末尾（`RemindWeekdays` 之后）新增：
```go
// 排序 ID
// 用于任务列表排序，采用大间距整数，基数 255，创建时取当前用户最大值 +1
SortId uint16 `gorm:"default:0"`
```
`AutoMigrate` 会自动为 task 表新增 `sort_id` 列，无需手写迁移。

### 2. 领域实体：新增 SortId
文件 [entities/task.go](file:///home/nathan/Projects/nao-todo-server/domain/task/entities/task.go#L12-L32)
在 `Task` 结构体新增 `SortId uint16`（放在 `RemindWeekdays` 之后）。

### 3. 值对象
- 创建 [valueobjects/createTask.go](file:///home/nathan/Projects/nao-todo-server/domain/task/valueobjects/createTask.go)：`CreateTask` 结构体新增 `SortId uint16`。`NewCreateTask` 构造器签名**不改**（SortId 由领域服务赋值，与 CheckItem 的做法一致——CheckItem 的 `NewCreateTaskCheckItem` 也不含 SortId 参数）。
- 更新 [valueobjects/updateTask.go](file:///home/nathan/Projects/nao-todo-server/domain/task/valueobjects/updateTask.go)：`UpdateTask` 结构体新增 `SortId *uint16`；`NewUpdateTask` 末尾新增参数 `sortId *uint16` 并赋值 `vo.SortId = sortId`。

### 4. 仓库接口与实现
- 接口 [repositories/task.go](file:///home/nathan/Projects/nao-todo-server/domain/task/repositories/task.go#L11-L20)：在 Task 段新增 `GetMaxSortId(ctx context.Context, userId int64) uint16`。
- 实现 [persistence/task/repoImpl.go](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/task/repoImpl.go#L54-L68)：仿 `GetMaxCheckItemSortId` 新增：
```go
func (taskRepo *TaskRepoImpl) GetMaxSortId(ctx context.Context, userId int64) uint16 {
	var maxSortId uint16 = 255
	taskRepo.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ?", userId).
		Pluck("MAX(sort_id)", &maxSortId)
	return maxSortId
}
```
`Create` / `List` 逻辑不改（List 靠通用 Sort 支持 `sortId`）。

### 5. 转换器（persistence/task/converters.go）
文件 [converters.go](file:///home/nathan/Projects/nao-todo-server/infrastructure/persistence/task/converters.go)
- `CreateTaskValueObjectToModel`（L15-34）：新增 `SortId: createTaskValueObject.SortId,`。
- `UpdateTaskValueObjectToMap`（L106-189）：末尾新增
```go
if updateTaskValueObject.SortId != nil {
	updateMap["SortId"] = *updateTaskValueObject.SortId
}
```
- （可选，保持对称）`UpdateTaskValueObjectToModel`（L39-101）：新增 `if vo.SortId != nil { m.SortId = *vo.SortId }`。该函数当前似未被 Update 使用（Update 走 map），按现有风格补齐即可。
- `TaskModel2Entity`（L194-218）：新增 `e.SortId = m.SortId`。

### 6. 领域服务：新增 CreateTask
- 接口 [service/service.go](file:///home/nathan/Projects/nao-todo-server/domain/task/service/service.go#L11-L31)：`TaskDomain` 的 `--- Task ---` 段新增：
```go
CreateTask(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTask,
) (*entities.Task, error)
```
- 实现 [service/serviceImpl.go](file:///home/nathan/Projects/nao-todo-server/domain/task/service/serviceImpl.go#L19-L50)：新增
```go
func (d *TaskDomainImpl) CreateTask(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTask,
) (*entities.Task, error) {
	vo.SortId = d.taskRepo.GetMaxSortId(ctx, userId) + 1
	return d.taskRepo.Create(ctx, userId, vo)
}
```
- 同文件 `Copy`（L49）：把 `return d.taskRepo.Create(ctx, userId, &vo)` 改为 `return d.CreateTask(ctx, userId, &vo)`，使复制任务也初始化 SortId。

### 7. 应用层
文件 [application/task/appImpl.go](file:///home/nathan/Projects/nao-todo-server/application/task/appImpl.go#L59-L80)
`CreateTask` 中把 `taskApp.taskRepo.Create(ctx, userId, createTaskValueObject)` 改为 `taskApp.taskDomain.CreateTask(ctx, userId, createTaskValueObject)`。

### 8. 应用层转换器
文件 [application/task/converters.go](file:///home/nathan/Projects/nao-todo-server/application/task/converters.go)
- `TaskEntityToGetRes`（L42-68）：新增 `res.SortId = taskEntity.SortId`。
- `UpdateTaskReqToValueObject`（L114-169）：`NewUpdateTask(...)` 调用末尾新增实参 `req.SortId`（透传指针）。
- `CreateTaskReqToValueObject`：无需改（SortId 由领域服务生成，不从请求读取）。

### 9. 接口 DTO
文件 [interfaces/types/task.go](file:///home/nathan/Projects/nao-todo-server/interfaces/types/task.go)
- `GetTaskRes`（L4-22）：新增 `SortId uint16 \`json:"sortId"\``。
- `UpdateTaskReq`（L42-59）：新增 `SortId *uint16 \`json:"sortId"\``。
- `CreateTaskReq`：不加（创建时后端自动生成）。

## Verification

1. 编译：`go build ./...` 通过（重点检查 `NewUpdateTask` 新增参数后所有调用点、`TaskDomain` 接口新方法的实现完整性）。
2. `go vet ./...` 无新增告警。
3. 运行时（手动/接口）验证：
   - 连续创建多个任务，`sortId` 依次递增（256、257…）。
   - `GET /tasks?sort=sortId:asc`（或对应列表接口的 sort 参数）返回按 `sortId` 升序。
   - 通过更新接口传 `{"sortId": 300}`，`GET` 单任务返回 `sortId=300`。
   - 复制任务后新任务 `sortId` 为当前最大值 +1。
4. 确认 `DoMigration` 启动后 task 表已新增 `sort_id` 列（AutoMigrate 自动完成）。

## 变更文件清单

1. `infrastructure/persistence/models/task.go`
2. `domain/task/entities/task.go`
3. `domain/task/valueobjects/createTask.go`
4. `domain/task/valueobjects/updateTask.go`
5. `domain/task/repositories/task.go`
6. `infrastructure/persistence/task/repoImpl.go`
7. `infrastructure/persistence/task/converters.go`
8. `domain/task/service/service.go`
9. `domain/task/service/serviceImpl.go`
10. `application/task/appImpl.go`
11. `application/task/converters.go`
12. `interfaces/types/task.go`
