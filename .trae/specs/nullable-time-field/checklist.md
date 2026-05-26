# 可空时间字段更新检查清单

## 设计和类型定义检查
- [x] `NullableTime` 结构体正确定义，包含 `Valid`、`IsNull`、`Time` 三个字段
- [x] `NewNullableTimeWithTime` 函数正确创建有时间值的对象
- [x] `NewNullableTimeNull` 函数正确创建设置为 NULL 的对象
- [x] `ToSqlNullTime` 方法正确转换为 sql.NullTime 类型
- [x] `ShouldUpdate` 方法正确判断是否需要更新
- [x] `IsSetToNull` 方法正确判断是否设置为 NULL

## 值对象更新检查
- [x] UpdateTask 值对象中所有时间字段（StartAt、EndAt、ArchivedAt、StarMarkAt、GivenUpAt）类型已从 `sql.NullTime` 改为 `*NullableTime`
- [x] NewUpdateTask 函数参数类型已更新
- [x] Validate 方法中的时间验证逻辑已更新，正确处理 `*NullableTime` 类型

## 转换逻辑检查
- [x] `StringPtr2NullableTime` 函数正确实现，能处理三种情况：
  - 输入为 nil 时返回 nil（不更新）
  - 输入为空字符串或 "null" 时返回 NULL 状态的 NullableTime
  - 输入为有效时间字符串时返回带时间值的 NullableTime
- [x] application 层 converter 中 UpdateTaskReqToValueObject 函数使用了新的转换函数
- [x] 持久化层 converter 中 UpdateTaskValueObjectToMap 函数正确处理 *NullableTime 类型
- [x] 持久化层转换逻辑中，对于设置为 NULL 的字段，在更新 map 中设置为 nil

## 功能验证检查
- [x] 代码能够正常编译，无语法错误
- [x] 当前端不传时间字段时，该字段不会被更新（保持原值）
- [x] 当前端传 null 时，该字段被正确设置为 NULL
- [x] 当前端传有效时间字符串时，该字段被正确设置为对应的时间值
- [x] 时间字段的验证逻辑（如结束时间晚于开始时间）仍然正常工作
