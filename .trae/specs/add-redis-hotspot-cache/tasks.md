# Tasks

- [x] Task 1: 新增缓存辅助组件
  - 在 `infrastructure/persistence/cache/` 下新增一个基于 `dbs.RdsCli` 的缓存辅助组件（如 `cache.go`），提供泛型或 JSON 序列化的 `Get`、`Set(带 TTL)`、`Del` 方法，以及统一的 key 构造辅助（含用户维度前缀）。
  - Redis 出错时：`Get` 返回未命中、`Set`/`Del` 静默失败，绝不 panic 或阻断调用方。
  - 沿用项目现有代码风格（中文注释、`go-redis/v8`、错误处理方式参考 `RateLimitRepoImpl`）。
  - verify: `go build ./...` 通过。

- [x] Task 2: 会话校验热点缓存
  - 修改 `infrastructure/persistence/identity/sessionRepoImpl.go`：`IsSessionValid` 结果做缓存（userId+token key，较短 TTL）；`Create`/`UpdateToken`/`Delete` 后失效对应缓存。
  - verify: `go build ./...` 通过。

- [x] Task 3: 用户资料与配置热点缓存
  - 修改 `infrastructure/persistence/identity/userRepoImpl.go`：`FindById` 与 `GetConfig` 做 cache-aside（userId key，TTL）；`UpdateNickname`/`UpdateAvatar`/`UpdatePassword`/`UpdateConfig`/`Deactive`/`Active`/`Delete` 后失效对应缓存。
  - verify: `go build ./...` 通过。

- [x] Task 4: 项目与标签列表热点缓存
  - 修改 `infrastructure/persistence/project/repoImpl.go`：`GetByUserId` 做 cache-aside；`Create`/`Update`/`Delete`/`Restore`/`Archive`/`Unarchive`/`BatchUpdate` 后失效。
  - 修改 `infrastructure/persistence/tag/repoImpl.go`：`Get` 做 cache-aside；`Create`/`Update`/`Delete`/`BatchUpdate` 后失效。
  - verify: `go build ./...` 通过。

- [x] Task 5: 依赖注入接线
  - 修改 `infrastructure/initialize.go`：把 `dbs.RdsCli` 注入 session/user/project/tag 各仓储构造函数（沿用 `NewRateLimitRepo(dbs.RdsCli)` 风格），并同步更新各 `NewXxxRepo` 构造函数签名。
  - verify: `go build ./...` 通过；`go vet ./...` 无新增告警。

# Task Dependencies
- Task 2、Task 3、Task 4 均依赖 Task 1（缓存辅助组件）。
- Task 5 依赖 Task 2~4（构造函数签名变更后统一接线）。
- Task 2、Task 3、Task 4 之间相互独立，可并行开发。
