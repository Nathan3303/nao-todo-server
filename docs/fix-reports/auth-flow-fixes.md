# 认证接口执行流程修复报告

## 1. 概述

- **范围**：用户认证相关接口（`/api/auth/*`：signin / signup / checkin / signout / validate）及其支撑的领域服务、基础设施实现、中间件。
- **触发背景**：对认证链路执行流程的代码走查共发现 16 项问题，本次修复其中 **8 项**（高优先级 4 项 + 无争议健壮性 4 项）；其余为产品行为权衡或范围外，列于附录。
- **结论**：所有改动通过 `go build`、`go vet`、`go test ./...`（全绿），未破坏既有行为（响应码、响应文案、限流阈值 8 次/分均保持不变）。

## 2. 修复清单

| #   | 问题                                                           | 严重度 | 修复方式                                                                                           | 涉及文件                                                                                                                   |
| --- | -------------------------------------------------------------- | ------ | -------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| 1   | 注册时 DB 故障被误判为「邮箱可用」                             | 高     | `FindByEmail` 返回领域错误 `ErrUserNotFound`，`SignUp` 按 `errors.Is` 区分「用户不存在」与真实错误 | `userRepoImpl.go`、`appImpl.go`                                                                                            |
| 2   | 注册「查重 + 创建」非原子，并发注册竞态                        | 高     | `CreateByVO` 识别 MySQL 唯一索引冲突（1062）为 `ErrEmailExists`，由 DB 兜底                        | `userRepoImpl.go`、`appImpl.go`                                                                                            |
| 3   | 登录内部错误被统一文案吞掉、无日志                             | 高     | 按错误类型分级记录日志（Warn：业务拒绝；Error：基础设施故障）                                      | `controllers/identity.go`                                                                                                  |
| 4   | 限流 `Get`+`Incr` 非原子，并发计数竞态                         | 高     | 接口合并为原子 `Allow`，Redis Lua 脚本「检查 + 计数」                                              | `rateLimit.go`、`service.go`、`serviceImpl.go`、`app.go`、`appImpl.go`、`rateLimitRepoImpl.go`、`middlewares/rateLimit.go` |
| 5   | 限流计数 `int8` 溢出导致限流失效                               | 高     | 限流全链路 `int8 → int64`                                                                          | 同上                                                                                                                       |
| 6   | 限流 Redis 故障时静默失效（fail-open 不可观测）                | 中     | 故障放行时记录 Warn 告警                                                                           | `rateLimitRepoImpl.go`                                                                                                     |
| 7   | `/auth` 全组共享一个限流桶，checkin 挤占登录额度               | 中     | `/auth/signin` 独立限流桶（`auth-signin`），阈值不变                                               | `routers/authRouter.go`                                                                                                    |
| 8   | `Authorization` 头解析脆弱（大小写敏感、多余空白导致验签失败） | 中     | `SplitN` + `EqualFold` + `TrimSpace`，兼容 `Bearer`/`bearer`                                       | `middlewares/jwtValidator.go`                                                                                              |
| 9   | 登出删除会话未限定 `user_id`，仅按 token 删除                  | 中     | `Delete` 改为 `user_id + token` 双条件                                                             | `sessionRepoImpl.go`                                                                                                       |
| 10  | 会话表无唯一约束，并发登录产生孤儿会话行                       | 中     | `UserId` 改 `uniqueIndex` + `Create` 改 `clause.OnConflict` upsert                                 | `models/user.go`、`sessionRepoImpl.go`                                                                                     |
| 11  | 邮箱格式无校验，`abc` 也能注册                                 | 中     | `SignInReq`/`SignUpReq` 的 `Email` 加 `binding:"required,email"`                                   | `types/identity.go`                                                                                                        |

## 3. 各问题详情

### 3.1 注册：DB 故障误判与并发竞态（#1、#2）

**根因**：`SignUp` 仅以 `err == nil` 判断邮箱是否已存在（`application/auth/appImpl.go`）。但 `FindByEmail` 返回两类错误——「用户不存在」与 DB 故障透传；DB 故障时上层会误以为邮箱可用并继续走创建分支，把真实故障掩盖成「注册失败」。同时「查重 → 创建」两步非原子，并发注册同一邮箱时两次都通过查重。

**修复**：

- `userRepoImpl.go`：`gorm.ErrRecordNotFound` 分支返回领域错误 `domerr.ErrUserNotFound`（原为 `errors.New("用户不存在")`，无法被 `errors.Is` 识别）。
- `appImpl.go` `SignUp`：

```go
_, err := as.userRepo.FindByEmail(ctx, signUpInput.Email)
switch {
case err == nil:
    return domerr.ErrEmailExists
case !errors.Is(err, domerr.ErrUserNotFound):
    // DB 故障等真实错误直接透传，避免被误判为"邮箱可用"
    return fmt.Errorf("auth.SignUp.FindByEmail: %w", err)
}
```

- `userRepoImpl.go` `CreateByVO` 捕获 MySQL 唯一索引冲突（`*mysql.MySQLError` Number 1062）返回 `ErrEmailExists`，并发注册由 `users.email` 唯一索引兜底，用户收到明确提示而非「注册失败」。

### 3.2 登录：内部错误无日志（#3）

**根因**：`UserSignIn` 为防用户枚举统一返回「邮箱或密码错误」，但服务端也未记录真实原因，线上故障无法排查。

**修复**：`interfaces/controllers/identity.go` 按错误类型分级记录：

- `ErrUserNotFound` / `ErrPasswordMismatch` → `Warn`（业务拒绝，可能是枚举探测）；
- 其余（DB 故障、JWT 生成失败、会话创建失败）→ `Error`。

仅记录邮箱、不记录密码；响应文案保持统一（防枚举语义不变）。

### 3.3 限流：原子性、溢出、可观测性、分桶（#4-#7）

**根因**：

- `CheckRateLimit` 先 `Get` 再 `Incr`，两步非原子，并发请求可同时读到超限前计数而全部放行；
- 计数用 `int8` 接收 Redis 计数，超过 127 溢出为负数导致判断失效；
- Redis 故障时静默 fail-open，限流失效无从察觉；
- `/auth` 全组（signin/signup/checkin/signout/validate）共享一个 8 次/分桶，高频 checkin 会挤占登录额度。

**修复**：

- `domain/identity/repositories/rateLimit.go`：接口由 `Get`/`Incr` 合并为 `Allow(ctx, key, limit int64) (bool, error)`；
- `rateLimitRepoImpl.go`：Redis Lua 脚本原子执行「检查 + 计数」，超限请求不计数，首次计数时设置 60s 窗口过期：

```lua
local c = redis.call('GET', KEYS[1])
if c and tonumber(c) >= tonumber(ARGV[1]) then
  return -1
end
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[2])
end
return n
```

- 全链路签名 `int8 → int64`：`AuthApp.RateLimit`、`IdentityDomain.CheckRateLimit`、`RateLimit.Allow`、`RateLimiter` 中间件；
- Redis 不可用时仍放行（fail-open，保持原语义），但记录 `Warn` 告警；
- `routers/authRouter.go`：`/auth/signin` 独立限流桶 `auth-signin`（阈值仍为 8 次/分），其余接口共享 `auth` 桶。

### 3.4 JWT 头解析（#8）

**根因**：`getJwtString` 用 `strings.Split(jwtRaw, "Bearer ")` + `append` 取 `[1]`，大小写敏感（`bearer` 不识别）、`"Bearer  xxx"`（双空格）会带前导空格导致验签失败。

**修复**：`interfaces/middlewares/jwtValidator.go`

```go
parts := strings.SplitN(authHeader, " ", 2)
if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
    return strings.TrimSpace(parts[1])
}
return ""
```

兼容大小写与多余空白；无 Bearer scheme（裸 token / 其他 scheme）仍返回空串拒绝，行为与之前一致。

### 3.5 登出删除会话未限定用户（#9）

**根因**：`sessionRepoImpl.Delete` 的 `deleteCond` 只含 `Token` 字段，`userId` 参数未参与 WHERE。

**修复**：`deleteCond := &models.UserSession{UserId: int64(userId), Token: token}`，双条件删除。

### 3.6 会话表唯一约束与并发孤儿行（#10）

**根因**：模型注释标注「目标用于多设备登录」，但 `Create` 实际是「一用户一会话」覆盖逻辑；并发登录时两个请求都先查到「无记录」再插入，产生两行会话，`FindByUserIdAndToken` 的 `First` 只取其一，另一行成为孤儿。

**修复**：

- `models/user.go`：`UserSession.UserId` 由 `index` 改为 `uniqueIndex:idx_user_session_user_id`；
- `sessionRepoImpl.go` `Create`：合并更新/插入为 `clause.OnConflict` upsert（冲突列 `user_id`，`DoUpdates` token/expired_at/ip4/region/device_type），并保留先查旧记录以清理旧 token 缓存。

> 注意：`OnConflict` 依赖唯一索引生效，`AutoMigrate` 在启动时创建，见 §5 部署注意事项。

### 3.7 邮箱格式校验（#11）

**根因**：`SignInReq`/`SignUpReq` 的 `Email` 仅有 `binding:"required"`，非法邮箱也能进入业务层。

**修复**：`interfaces/types/identity.go` 的 `Email` 改为 `binding:"required,email"`，非法格式在绑定阶段即返回参数错误（10001/10011），不影响已注册用户。

## 4. 变更文件清单

```
 application/auth/app.go                          |  2 +-
 application/auth/appImpl.go                      | 13 ++++-
 domain/identity/repositories/rateLimit.go         |  4 +-
 domain/identity/service/service.go                |  2 +-
 domain/identity/service/serviceImpl.go            | 13 +++--
 go.mod                                            |  2 +-  (go-sql-driver/mysql: indirect → direct)
 infrastructure/persistence/identity/rateLimitRepoImpl.go | 59 +++++++++++++---------
 infrastructure/persistence/identity/sessionRepoImpl.go  | 51 +++++++------------
 infrastructure/persistence/identity/userRepoImpl.go     | 15 +++++-
 infrastructure/persistence/models/user.go         |  3 +-
 interfaces/controllers/identity.go                |  8 +++
 interfaces/middlewares/jwtValidator.go            | 11 ++--
 interfaces/middlewares/rateLimit.go               |  2 +-
 interfaces/routers/authRouter.go                  | 15 +++---
 interfaces/types/identity.go                      |  4 +-
 15 files changed, 119 insertions(+), 85 deletions(-)
```

## 5. 部署注意事项

`user_sessions.user_id` 唯一索引由启动时 `DoMigration`（`AutoMigrate`）创建，但 `DoMigration` 对迁移错误是**静默忽略**的（既有行为）。若线上库已存在同一 `user_id` 的多行会话（历史并发孤儿数据），索引会创建失败且无报错。**上线前建议手动清理一次**：

```sql
DELETE s1 FROM user_sessions s1
JOIN user_sessions s2
  ON s1.user_id = s2.user_id AND s1.id < s2.id;
```

清理后重启服务即可自动加索引。

## 6. 验证

```bash
go build ./... && go vet ./... && go test ./...
```

- `go build`：无错误；
- `go vet`：无告警；
- `go test ./...`：全部通过（含 `domain/project`、`domain/task`、`textutils`、`domain/types` 等既有测试）。

## 附录：本次未修复的问题（范围外 / 产品权衡）

| #   | 问题                                                                | 原因                                                             |
| --- | ------------------------------------------------------------------- | ---------------------------------------------------------------- |
| A1  | 注册时创建 `UserConfig` 的错误被忽略且不在事务内                    | 已被 `GetConfig` 懒创建兜底，改动涉及注册事务化，超出本次范围    |
| A2  | 登录响应时间差可侧信道枚举邮箱（bcrypt 仅对已存在用户执行）         | 权衡项，需引入虚拟哈希，改动较大                                 |
| A3  | 注销（deactived）用户仍可登录并正常使用全部业务接口                 | 属产品语义决策（当前设计允许等待期内反悔），需明确产品规则后处理 |
| A4  | CheckIn 换发 token 不刷新会话 `expired_at`，会话硬上限 7 天         | 属安全设计权衡，需确认产品预期后再定                             |
| A5  | CheckIn 的 `DeviceType` 参数未使用、轮换无设备/IP 环境校验          | 需要新的会话记录模型支持，超出本次范围                           |
| A6  | token 通过 query 参数传递（`?token=`，兼容 SSE），会进入 URL/日志   | 改动会影响 SSE 客户端，需前后端联调                              |
| A7  | `Validate` 每次请求都查 DB 会话表，`IsSessionValid` 缓存仅 SSE 使用 | 纯性能优化，无正确性问题                                         |
| A8  | JWT `Payload` 字段明文存邮箱，且从未被业务使用                      | 移除会导致存量 token 兼容问题，需协调客户端重新登录              |
