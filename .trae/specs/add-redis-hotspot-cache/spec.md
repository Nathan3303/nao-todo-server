# Redis 热点数据缓存 Spec

## Why
鉴权链路与首页侧边栏存在大量高频、读多写少的 MySQL 查询（会话校验、用户资料/配置、项目/标签列表），每次请求直连数据库带来不必要的负载与延迟。引入 Redis 热点缓存可显著降低这些热路径的 DB 压力并提升响应速度。

## What Changes
- 新增一个可复用的缓存辅助组件（基于已有的 `dbs.RdsCli`），提供 JSON 序列化的 get/set/del 能力，供各仓储做 cache-aside。
- 为以下 3 个热点读路径增加 **cache-aside**（读时回填）+ **写操作主动失效** + **TTL 兜底**：
  1. **会话校验**：`IsSessionValid` 结果。
  2. **用户资料/配置**：`FindById` 用户资料、`GetConfig` 用户配置。
  3. **项目/标签列表**：项目 `GetByUserId`、标签 `Get`。
- 在 `initialize.go` 中把 `dbs.RdsCli` 注入相关仓储构造函数（沿用现有 `RateLimitRepo` 的注入风格）。
- 所有缓存均以用户维度隔离，缓存不可用（Redis 报错）时静默降级回源 MySQL，不影响主流程。

## Impact
- Affected specs: 会话校验、用户资料/配置、项目/标签列表
- Affected code:
  - 新增：`infrastructure/persistence/cache/`（缓存辅助组件）
  - 修改：`infrastructure/persistence/identity/sessionRepoImpl.go`、`userRepoImpl.go`
  - 修改：`infrastructure/persistence/project/repoImpl.go`
  - 修改：`infrastructure/persistence/tag/repoImpl.go`
  - 修改：`infrastructure/initialize.go`（注入 `dbs.RdsCli`）

## ADDED Requirements

### Requirement: 缓存辅助组件
系统 SHALL 提供一个基于 `dbs.RdsCli` 的缓存辅助组件，支持按字符串 key 进行带 TTL 的 JSON 值读取、写入与删除，并在 Redis 不可用时返回“未命中/降级”而非中断调用方。

#### Scenario: 写入并读取命中
- **WHEN** 调用方写入一个键值对并设置 TTL 后再次读取该键
- **THEN** 返回反序列化后的原始值且标记为命中

#### Scenario: 未命中
- **WHEN** 读取一个不存在的键
- **THEN** 返回“未命中”标记，调用方回源查询

#### Scenario: Redis 不可用降级
- **WHEN** Redis 连接出错
- **THEN** 读取返回“未命中”、写入/删除静默失败，调用方回源 MySQL 且请求正常完成

### Requirement: 会话校验热点缓存
系统 SHALL 对 `IsSessionValid` 的校验结果做缓存（key 以 userId+token 维度，带较短 TTL），并在会话写操作后失效。

#### Scenario: 校验结果命中
- **WHEN** 同一 userId+token 在 TTL 内重复校验
- **THEN** 返回缓存的有效性结果，不查询 MySQL

#### Scenario: 会话变更失效
- **WHEN** 会话被创建/更新 token/删除
- **THEN** 对应会话缓存被删除

### Requirement: 用户资料与配置热点缓存
系统 SHALL 对 `FindById`（用户资料）与 `GetConfig`（用户配置）结果做 cache-aside 缓存（key 以 userId 维度，带 TTL），并在相关写操作后失效。

#### Scenario: 资料/配置命中
- **WHEN** 在 TTL 内重复读取同一用户的资料或配置
- **THEN** 返回缓存结果，不查询 MySQL

#### Scenario: 资料/配置变更失效
- **WHEN** 更新昵称/头像/密码/配置或注销/激活/删除用户
- **THEN** 该用户的资料与配置缓存被删除

### Requirement: 项目与标签列表热点缓存
系统 SHALL 对项目列表 `GetByUserId` 与标签列表 `Get` 结果做 cache-aside 缓存（key 以 userId 维度，带 TTL），并在相关写操作后失效。

#### Scenario: 列表命中
- **WHEN** 在 TTL 内重复读取同一用户的项目或标签列表
- **THEN** 返回缓存结果，不查询 MySQL

#### Scenario: 列表变更失效
- **WHEN** 项目发生创建/更新/删除/恢复/归档/取消归档/批量更新，或标签发生创建/更新/删除/批量更新
- **THEN** 对应用户的项目或标签列表缓存被删除
