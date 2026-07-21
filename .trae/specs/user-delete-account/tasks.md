# 用户注销功能 - 实现计划

## [x] Task 1: 新增 DeleteUser 请求类型定义和响应字段扩展
- **Priority**: high
- **Depends On**: None
- **Description**:
  - 在 `interfaces/types/user.go` 中新增 `DeleteUserReq` 请求结构体，用于接收用户密码
  - 在 `GetUserProfileRes` 中新增 `DeactivedAt` 字段，用于返回待注销用户的预计删除时间
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-5
- **Test Requirements**:
  - `programmatic` TR-1.1: 验证 `DeleteUserReq` 结构体包含 `Password` 字段且为必填
  - `programmatic` TR-1.2: 验证 `GetUserProfileRes` 结构体包含 `DeactivedAt` 字段
- **Notes**: 参照现有 `DeactiveUserReq` 的定义方式

## [x] Task 2: 仓储层实现删除用户所有会话
- **Priority**: high
- **Depends On**: None
- **Description**: 在 `domain/identity/repositories/session.go` 和 `infrastructure/persistence/identity/sessionRepoImpl.go` 中新增 `DeleteByUserId` 方法，用于删除指定用户的所有会话
- **Acceptance Criteria Addressed**: AC-1, AC-3
- **Test Requirements**:
  - `programmatic` TR-2.1: 删除用户所有会话后，该用户的所有会话记录不存在于数据库
  - `programmatic` TR-2.2: 删除用户所有会话后，相关缓存被清除
- **Notes**: 需要删除数据库中该用户的所有会话记录，并清除对应的缓存

## [x] Task 3: 应用层实现用户注销业务逻辑
- **Priority**: high
- **Depends On**: Task 1, Task 2
- **Description**: 在 `application/user/app.go` 和 `application/user/appImpl.go` 中新增 `DeleteUser` 方法，实现：
  - 验证用户身份（从上下文获取 userId）
  - 验证用户密码
  - 标记用户为待注销状态（调用 `userRepo.Deactive`）
  - 删除用户所有会话（调用 `sessionRepo.DeleteByUserId`）
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3
- **Test Requirements**:
  - `programmatic` TR-3.1: 密码正确时，调用 `DeleteUser` 返回 nil（成功），`DeactivedAt` 被设置
  - `programmatic` TR-3.2: 密码错误时，调用 `DeleteUser` 返回 `ErrPasswordMismatch`
  - `programmatic` TR-3.3: userId 无效时，调用 `DeleteUser` 返回 `ErrInvalidUserID`
  - `programmatic` TR-3.4: 用户注销后，所有会话被删除
- **Notes**: 需要注入 `sessionRepo` 依赖

## [x] Task 4: 应用层实现取消注销业务逻辑
- **Priority**: high
- **Depends On**: Task 1
- **Description**: 在 `application/user/app.go` 和 `application/user/appImpl.go` 中修改 `ActiveUser` 方法，实现：
  - 验证用户身份（从上下文获取 userId）
  - 验证用户密码
  - 检查用户是否处于待注销状态
  - 恢复用户正常状态（清除 `DeactivedAt`）
- **Acceptance Criteria Addressed**: AC-4
- **Test Requirements**:
  - `programmatic` TR-4.1: 处于待注销状态的用户，密码正确时调用 `ActiveUser` 返回 nil（成功），`DeactivedAt` 被清除
  - `programmatic` TR-4.2: 密码错误时，调用 `ActiveUser` 返回 `ErrPasswordMismatch`
  - `programmatic` TR-4.3: 用户未处于待注销状态时，调用 `ActiveUser` 返回错误
- **Notes**: 现有 `ActiveUser` 方法需要修改错误提示信息

## [x] Task 5: 更新用户个人信息响应
- **Priority**: high
- **Depends On**: Task 1
- **Description**: 在 `application/user/converters.go` 中修改 `UserEntity2Res` 方法，将 `DeactivedAt` 字段转换到响应中
- **Acceptance Criteria Addressed**: AC-5
- **Test Requirements**:
  - `programmatic` TR-5.1: 正常用户的 `GetProfile` 响应中 `DeactivedAt` 为空
  - `programmatic` TR-5.2: 待注销用户的 `GetProfile` 响应中包含 `DeactivedAt` 时间（预计删除时间=DeactivedAt+7天）
- **Notes**: 预计删除时间 = `DeactivedAt` + 7天

## [x] Task 6: 控制器层实现用户注销接口
- **Priority**: high
- **Depends On**: Task 3
- **Description**: 在 `interfaces/controllers/user.go` 中新增 `DeleteUser` 控制器方法，处理 HTTP 请求：
  - 绑定请求参数
  - 调用应用层 `DeleteUser` 方法
  - 返回成功或失败响应
- **Acceptance Criteria Addressed**: AC-1, AC-2
- **Test Requirements**:
  - `programmatic` TR-6.1: POST /user/delete 请求，密码正确时返回状态码 200 和成功消息
  - `programmatic` TR-6.2: POST /user/delete 请求，密码错误时返回状态码 400 和错误消息
- **Notes**: 响应码范围使用 1013x

## [x] Task 7: 路由层注册用户注销接口
- **Priority**: high
- **Depends On**: Task 6
- **Description**: 在 `interfaces/routers/userRouter.go` 中注册 `POST /user/delete` 路由，应用 JWT 验证和限流中间件
- **Acceptance Criteria Addressed**: AC-3
- **Test Requirements**:
  - `programmatic` TR-7.1: 未登录用户调用注销接口返回认证失败
  - `programmatic` TR-7.2: 已登录用户调用注销接口能正确路由到控制器
- **Notes**: 使用 POST HTTP 方法

## [x] Task 8: 认证验证时检查用户是否已注销
- **Priority**: high
- **Depends On**: Task 3
- **Description**: 修改 `application/auth/appImpl.go` 的 `Validate` 方法，在验证 JWT 和会话后，检查用户是否处于待注销状态
- **Acceptance Criteria Addressed**: AC-3
- **Test Requirements**:
  - `programmatic` TR-8.1: 已注销用户使用旧令牌调用受保护接口返回认证失败，提示用户已注销
  - `programmatic` TR-8.2: 正常用户使用有效令牌调用受保护接口成功通过验证
- **Notes**: 通过 `userRepo.FindById` 获取用户后检查 `IsDeactived()` 方法

## [x] Task 9: 检查定时任务清理逻辑
- **Priority**: medium
- **Depends On**: None
- **Description**: 检查 `infrastructure/cron/deleteInactiveUser.go` 中的定时任务逻辑，确保：
  - 定时任务正确删除 `DeactivedAt` 超过7天的用户
  - 删除用户时级联删除关联数据
- **Acceptance Criteria Addressed**: AC-6
- **Test Requirements**:
  - `programmatic` TR-9.1: `DeactivedAt` 超过7天的用户被定时任务删除
  - `programmatic` TR-9.2: 删除用户时关联数据（任务、项目、标签等）被级联删除
- **Notes**: 现有定时任务 `DeleteDeactivatedUsers` 需要验证逻辑正确性

## [x] Task 10: 数据库模型层检查关联关系和级联删除配置
- **Priority**: medium
- **Depends On**: None
- **Description**: 检查 `infrastructure/persistence/models/` 目录下的模型定义，确保用户删除时能正确清理关联数据（任务、项目、标签等）
- **Acceptance Criteria Addressed**: AC-6
- **Test Requirements**:
  - `human-judgement` TR-10.1: 模型定义中用户 ID 字段设置了外键约束和级联删除
  - `programmatic` TR-10.2: 删除用户时关联数据被级联删除
- **Notes**: 需要检查 task.go、project.go、tag.go、pomodoro.go 等模型文件