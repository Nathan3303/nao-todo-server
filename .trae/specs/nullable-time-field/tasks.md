# 可空时间字段更新任务列表

## 任务 1：创建 `NullableTime` 自定义类型
- [x] 在 `infrastructure/utils/` 目录下创建 `nullableTime.go` 文件
  - [x] 定义 `NullableTime` 结构体，包含 `Valid`（是否设置）、`IsNull`（是否为 NULL）、`Time`（时间值）字段
  - [x] 实现 `NewNullableTimeWithTime(time.Time) *NullableTime` 函数：创建有具体时间值的对象
  - [x] 实现 `NewNullableTimeNull() *NullableTime` 函数：创建设置为 NULL 的对象
  - [x] 实现 `ToSqlNullTime() sql.NullTime` 方法：转换为 sql.NullTime 类型
  - [x] 实现 `ShouldUpdate() bool` 方法：判断是否需要更新该字段
  - [x] 实现 `IsSetToNull() bool` 方法：判断是否设置为 NULL

## 任务 2：更新任务值对象（UpdateTask）
- [x] 更新 `domain/task/valueobjects/updateTask.go`
  - [x] 将 `StartAt`、`EndAt`、`ArchivedAt`、`StarMarkAt`、`GivenUpAt` 字段类型从 `sql.NullTime` 改为 `*NullableTime`
  - [x] 更新 `NewUpdateTask` 函数参数类型
  - [x] 更新 `Validate` 方法中对时间字段的验证逻辑

## 任务 3：更新应用层转换逻辑
- [x] 更新 `application/task/converters.go`
  - [x] 修改 `UpdateTaskReqToValueObject` 函数中的时间转换逻辑
  - [x] 使用新的转换函数处理 `*string` 到 `*NullableTime` 的转换
  - [x] 区分三种情况：nil（不更新）、空字符串/null（设置为NULL）、有效时间字符串（设置为具体时间）

## 任务 4：更新持久化层转换逻辑
- [x] 更新 `infrastructure/persistence/task/converters.go`
  - [x] 修改 `UpdateTaskValueObjectToMap` 函数
  - [x] 对于 `*NullableTime` 类型的字段：
    - 如果 `ShouldUpdate()` 返回 false：不加入更新 map
    - 如果 `IsSetToNull()` 返回 true：将值设为 nil
    - 否则：将 `ToSqlNullTime()` 的结果加入更新 map

## 任务 5：在 utils 中新增转换函数
- [x] 更新 `infrastructure/utils/timeParser.go`
  - [x] 新增 `StringPtr2NullableTime(s *string) *NullableTime` 函数
  - [x] 实现转换逻辑：
    - s == nil：返回 nil（不更新）
    - *s == "" 或 *s == "null"：返回 NewNullableTimeNull()
    - 其他情况：尝试解析时间字符串，成功则返回带时间的 NullableTime，失败返回 nil 或错误

## 任务 6：验证和测试
- [x] 编译代码确保没有语法错误
- [x] 验证三种场景：
  - [x] 不传时间字段：字段不更新
  - [x] 传 null：字段被设置为 NULL
  - [x] 传有效时间字符串：字段被设置为对应时间

## 任务依赖关系
- 任务 2、3、4、5 依赖于任务 1
- 任务 6 依赖于任务 1-5 全部完成
