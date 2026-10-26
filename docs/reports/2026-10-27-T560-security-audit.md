# T560 后端安全审计报告 · nao-todo-server

> 任务号：T560 · 角色：arch-designer（咨询/评审，零代码改动）
> 审计基线：**远端 main `c834404`**（`c83440484e5aa777da90a9cc38b8647f0940f9a0`，含 #58/#59/#60/#61 已合并修复）
> 报告日期：2026-10-27 · 交付对象：PM
> 本报告为**只读审计**产物：不修改代码、不开 PR、不升级依赖、不对生产发起任何请求。全文不含真实密钥/token/口令（一律 `xxx` 占位），不提供可直接利用的完整载荷。

---

## 1. 报告头

### 1.1 审计范围

| 维度 | 覆盖内容 |
| --- | --- |
| 代码 | 远端 main 全量 Go 源码（274 个 `.go`，其中 203 个非测试文件）：`cmd/` `conf/` `interfaces/` `application/` `domain/` `infrastructure/` |
| 认证与会话 | JWT 签发/验签/算法/过期、会话创建/轮换/并发 CAS/登出/绝对期 7 天、会话列表与下线、过期清理 cron |
| 越权 | IDOR/横向/纵向；中间件→控制器→应用→仓储全链路 `user_id` 约束；请求体是否可注入 `userId`（mass assignment） |
| 注入 | ORM 动态列（`sort`/`order by`）、raw SQL、LIKE 拼接、同步删除表名、JSON 列 |
| 上传/文件 | 头像上传（类型/大小/文件名/路径穿越）、头像读取鉴权、URL 方式头像 |
| Web 边界 | CORS/Origin 判定、CSRF 可行性、安全响应头、TLS 配置、可信代理 |
| 限流/暴力破解 | 登录/注册/检入/登出/改密/业务接口桶、原子性、故障策略 |
| 口令 | 算法/成本/盐/时序/策略/上限 |
| 敏感信息 | 仓库内配置与部署文件、默认值、错误码信息量、日志、缓存键、容器镜像 |
| 依赖与供应链 | `go.mod`/`go.sum` 版本清单、CI 门禁、GitHub Actions 固定方式 |
| SSRF/并发/DoS | 服务端出站请求、余额式/唯一键/CAS 竞态、无界查询/超时/批量接口 |
| 可观测性 | 请求日志、安全事件审计、审计留痕 |
| 文档 | `docs/plans/server-deploy.md`、头像/会话前端落地报告（作为客户端假设来源） |

### 1.2 方法

1. 静态读码 + 数据流追踪：`routers → middlewares（ClientInfo/JWTValidator/RateLimiter）→ controllers → application → domain → repositories`。
2. 配置与部署文件核对（`conf/config.yaml`、`conf/config.go` 的 env 覆盖逻辑、`Dockerfile`、`docker-compose*.yml`、`.github/workflows/ci.yml`）。
3. 定向搜索：`Order(`/`Raw(`/`Exec(`/`Sprintf` 拼 SQL、`io.ReadAll`、`Unscoped()`、`InsecureSkipVerify`、`http.Client`、`md5`/`math/rand`、`RowsAffected`、`err.Error()` 回显。
4. 与 #58/#59/#60/#61 的修复提交逐条对照，单列残余面（§4）。
5. **仅只读验证**：`git fetch`/`git show`/`git log`/`grep`/行号核对；未启动服务、未连接生产、未做任何攻击性探测。

### 1.3 严重度口径

| 级别 | 判定标准 |
| --- | --- |
| **P0** | 无需有效凭证（或仅需任意注册账号即可批量）读取/篡改他人数据，或直接接管账号 |
| **P1** | 需一个低门槛前置（单个部署配置错误、单次资源耗尽请求）即可造成大范围数据/凭据泄露或全站不可用 |
| **P2** | 需已持有有效会话/特定部署形态，影响限于该用户、可用性或防护缺失 |
| **P3** | 信息泄露、纵深防御与最佳实践缺失，单独不足以构成入侵 |

**证据类型**只允许两种：`读码确证`（代码路径可确定）与 `待运行时验证`（需部署态/客户端/工具确认）。二者在每条发现中分别标注，禁止把推断写成确证。

### 1.4 限制

- 无动态测试/渗透测试：按 PM 约束不对生产（`todobe.nathanao.space`）发起任何请求；本机未启动 MySQL/Redis 做复现窗口。
- 客户端（Web/桌面/移动）代码不在本仓，凡涉及「客户端如何渲染/存储 token」的结论均标 `待运行时验证`。
- 主机/网络层（防火墙实际规则、TLS 私钥权限、`/opt/...` 下的 my.cnf/redis.conf、日志平台）不在仓库内，未审计。
- 依赖漏洞未跑扫描器（本机无 `govulncheck`，且任务禁止 `go get`/升级），仅做版本静态核对。

### 1.5 未覆盖面

1. 运行时验证：限流实际命中、OOM 阈值、反代日志是否记录 query、生产 `.env` 实际取值。
2. 生产部署态：① **公网可达性已由 PM 外部 TCP 探测确认**（`3306/6379/3301/3302/5432` 均 closed/filtered，见 SEC-02「现状」）；② 仍未覆盖：主机内部监听面与容器端口发布、容器是否已改非 root、是否前置 nginx——这三项已转为用户可自行执行的只读检查（附录 B.1/B.2/B.8/B.9）。
3. 客户端安全：token 存储位置（localStorage/Cookie）、头像与评论字段的渲染方式（是否 `v-html`/`href`）。
4. 依赖 CVE 逐条确认（需 `govulncheck`/OSV 数据源）。
5. CI/CD secrets、分支保护、镜像仓库权限、备份与密钥轮换流程。
6. 历史提交全文扫描（仅核对了当前基线中 `git ls-files` 命中的配置文件；未对全部历史 blob 做密钥扫描）。

### 1.6 结论摘要

- **未发现 P0**：认证链路（固定 HS256、会话表双重校验、7 天绝对期、checkin 原地轮换 + CAS、登出 RowsAffected）经读码核对无绕过路径；单纯伪造 JWT **不足以**越权，因为 `Validate` 还要求 `(user_id, token)` 在 `user_sessions` 中存在且未过期。
- 主要风险集中在三处：**资源上限缺失（可用性）**、**凭据/部署配置的静默回退**、**横向信息暴露（头像/token in URL/错误详情）**。
- 业务数据访问控制整体一致：抽查任务/清单/标签/检查项/评论/番茄全链路均以服务端 `userId` 约束，请求 DTO 无 `userId` 字段，未发现跨租户读写。

### 1.7 复跑与交接（只读）

```bash
# 1) 确认审计基线（应输出 c834404…）
cd /home/nathan/Projects/nao-todo-server
git fetch origin && git log --oneline -1 origin/main

# 2) 关键点复检（只读）
grep -n "WithValidMethods\|WithoutClaimsValidation" infrastructure/auth/jwt.go
grep -n "SET password\|DeleteByUserIdExceptToken" -r application infrastructure
grep -n "io.ReadAll\|Limit(limit)\|Offset(" -r interfaces infrastructure | grep -v _test
grep -n "ctx.Query(\"token\")" interfaces/middlewares/jwtValidator.go
grep -n "fail-open" -r infrastructure
git show origin/main:conf/config.yaml | grep -n "jwtSecret"   # 只核对存在性，勿打印/外传

# 3) 修复后回归（建议由 RD 在独立分支执行）
go test ./... -count=1
go test -tags integration -p 1 ./...
# 新增用例建议落点：
#   infrastructure/persistence/identity/sessionRepoImpl_integration_test.go（改密撤销、DeleteById RowsAffected）
#   interfaces/routers/*_test.go（query token 收窄、限流/上限）
#   interfaces/controllers/*_test.go（错误响应白名单、请求体上限）
```

- 本报告仅为审计结论，**不含实现改动**；修复请由 PM 派发 RD，按 SEC 编号拆单。
- 建议修复顺序：SEC-02（凭据/部署，先轮换再改代码）→ SEC-01（资源上限）→ SEC-03/04/05（认证生命周期与爆破）→ SEC-06/07/12（信息暴露面）→ 其余加固。
- 报告未提交 git（按 PM 要求统一走 PR）。

---

## 2. 汇总表

| 严重度 | 数量 | 编号 |
| --- | --- | --- |
| **P0** | **0** | — |
| **P1** | **2** | SEC-01、SEC-02 |
| **P2** | **10** | SEC-03 ～ SEC-12 |
| **P3** | **6** | SEC-13 ～ SEC-18 |
| 合计 | **18** | — |

| 编号 | 严重度 | 类别 | 一句话 |
| --- | --- | --- | --- |
| SEC-01 | P1 | DoS/无界输入 | `/sync/push` 无上限读 body、批量与 `limit` 无上限，单请求可打挂单实例 |
| SEC-02 | P1 | 凭据/安全默认值 | 仓库内提交可用默认凭据且 env 为空时静默回退；compose 公开 3306/6379 且 root 连库 |
| SEC-03 | P2 | 会话生命周期 | 改密不撤销既有会话，旧 token 最长再活 7 天 |
| SEC-04 | P2 | 暴力破解 | 仅 IP 维度限流，无账户锁定，可分布式绕过；Redis 故障 fail-open |
| SEC-05 | P2 | 口令策略 | 注册口令无长度/强度校验（1 字符可注册），与改密策略不一致 |
| SEC-06 | P2 | 凭据泄露面 | 任意路由优先接受 `?token=`，7 天会话 token 可落入 URL/日志/分享 |
| SEC-07 | P2 | 横向越权/枚举 | 头像文件名为 `{userId}.ext`，任意登录用户可读任意用户头像且无速率限制 |
| SEC-08 | P2 | 上传校验 | 头像仅按扩展名判定；URL 方式接受任意字符串（含外链/非 http scheme） |
| SEC-09 | P2 | 容器加固 | 运行镜像为 golang 全量、root 运行、无 `.dockerignore`、无 cap/只读限制 |
| SEC-10 | P2 | 限流可靠性 | Redis 异常即放行，限流整体失效（含登录/改密） |
| SEC-11 | P2 | 部署/限流 | 未配置可信代理；按文档前置 nginx 后全站共享单一 IP 桶 → 登录被单点打满 |
| SEC-12 | P2 | 信息泄露 | 多处回显 `err.Error()`，泄露 DB/路径/主键值；重复 ID 形成跨租户存在性 oracle |
| SEC-13 | P3 | 用户枚举 | 注册文案 + 登录时序差 + 头像 404 可批量确认账号存在 |
| SEC-14 | P3 | 传输/缓存 | 无 HSTS/CSP；PII 响应无 `Cache-Control: no-store` |
| SEC-15 | P3 | 审计/日志 | 无安全事件审计；失败日志含邮箱；会话 token 明文入库并作为 Redis key |
| SEC-16 | P3 | 数据完整性 | 客户端可控 `createdAt/updatedAt/deletedAt` 参与 LWW 与增量游标 |
| SEC-17 | P3 | 供应链 | 依赖 EOL/停更；CI 无漏洞扫描门禁；Actions 未按 SHA 固定 |
| SEC-18 | P3 | 可用性 | 仅 TLS 分支设置服务器超时；HTTP 分支无超时（Slowloris） |

---

## 3. 逐条发现

### SEC-01 · P1 · DoS / 无界输入

- **文件:行**：`interfaces/controllers/sync.go:104`、`interfaces/controllers/sync.go:128`；`infrastructure/utils/query/pagination.go:12-16`；`infrastructure/persistence/task/repoImpl.go:348-353`；`interfaces/types/sync.go:143`
- **机制**：`/api/sync/push` 用 `io.ReadAll(ctx.Request.Body)` 一次性读入整个请求体且无上限，随后按 `tasks/…/deletions` 数组长度逐条落库；列表与增量拉取的 `limit` 只有 `<=0 取默认值` 的下限兜底，没有上限，`page` 直接换算成 `OFFSET`。
- **前置/可利用条件**：任意有效账号即可（注册开放，仅 8 次/分钟/IP）；单次请求即可触发。
- **影响**：`mem_limit: 512m` 的单实例可被一个请求打满内存/触发 OOM 重启；超大 `limit`/深分页可长时间占用连接池（`SetMaxOpenConns(30)`），全站 5xx。
- **风险量化（PM 追问）**：主要路径是**单请求 OOM 重启**——`io.ReadAll` 将整个 body 驻留内存，JSON 解码再放大 2–3×，而容器 `mem_limit=512m` 且 `restart: unless-stopped`，会被反复重启（每次重启期间全站不可用）；**连接池耗尽是次要路径**（30 条连接，由超大 `limit`/深 `page` 或多表批量数组长时间占用），表现是全站延迟/超时而非进程退出。前置仅为**任意有效账号**（注册开放、8 次/分钟/IP），不需要管理员权限或特殊网络位置。
- **证据类型**：`读码确证`（无 `MaxBytesReader`、无上限 clamp）；OOM 阈值与恢复时长 `待运行时验证`。
- **最小修法**：全局 `http.MaxBytesReader`（如 1–2 MB；上传路由单独放宽）；`limit` 上限（如 ≤500）与批量条数上限（如 ≤500）；超限返回 413/400 而非静默截断。
- **验证/回归方法**：集成测试向 `/sync/push` 发送 10 MB body 断言 413 且进程 RSS 平稳；`GET /api/tasks?limit=100000` 断言被夹取或 400；`page=10^9` 断言拒绝。

### SEC-02 · P1 · 凭据与安全默认值

- **文件:行**：`conf/config.yaml:5`（JWT 默认密钥）、`conf/config.yaml:16`（MySQL 口令）、`conf/config.yaml:24`（Redis 口令）；`conf/config.go:160-166`；`docker-compose.yml:5-6`、`:41-42`、`:76`；`Dockerfile:16-18`
- **机制**：仓库内提交了**非占位符**的 JWT/MySQL/Redis 凭据（值此处以 `xxx` 代替）；`overrideWithEnv` 只在 env **非空**时覆盖，`docker-compose.yml` 的 `JWT_SECRET=${JWT_SECRET}` 在 `.env` 缺该项时传入空串 → 静默回落到仓库内默认值，无启动校验；compose 同时把 3306/6379 发布到所有网卡，并以 MySQL `root` 连库。
- **前置/可利用条件**：① 部署 `.env` 缺少 `JWT_SECRET`/`MYSQL_ROOT_PASSWORD`/`REDIS_PASSWORD` 任一项；或 ② 主机防火墙/安全组未拦截 3306/6379 且线上口令与仓库值一致（当前实测已被拦截，见「现状」）。
- **现状（PM 复核，2026-10-27）**：PM 从本机对生产 `8.138.194.53` 做 TCP 连接测试，`3306 / 6379 / 3301 / 3302 / 5432` **全部 closed/filtered** ⇒ 当前公网**未暴露** DB/Redis（防火墙在挡），**P0 升级条件当前不成立，维持 P1**。该结论来自 PM 侧的外部 TCP 探测（仅建连测试，**未做连接后交互、未尝试任何凭据**）；本报告未对生产发起任何请求。
- **影响**：JWT 默认密钥单独泄露**不足以**伪造可用 token——`Validate` 还要求 `(user_id, token)` 命中 `user_sessions` 行，仓库默认 secret 签出的新 token 无对应行，只会得到 10041/10022。真正的高危链条是 **DB/Redis 可达 + 口令一致** → 读取 `user_sessions`（token 明文）即可接管任意账号，读取 `users.password` 可离线爆破；该链条当前被防火墙阻断，故 P0 不成立。
- **证据类型**：仓库内容与回退分支 `读码确证`；线上 env 是否覆盖 `待运行时验证`；端口暴露情况以 PM 外部探测为准（本条记为 `读码确证 + PM 实测`，本报告自身未探测）。
- **最小修法**：启动 fail-fast：拒绝空/默认/长度 <32 字节的 JWT 密钥、拒绝默认 DB 口令；`config.yaml` 只留占位符；compose 端口改 `127.0.0.1:3306:3306` / `127.0.0.1:6379:6379`；应用改用最小权限 DB 账号；轮换并作废仓库内曾出现的凭据。
- **验证/回归方法**：单测 `overrideWithEnv` 在 `JWT_SECRET` 缺失/为默认值时 panic 退出；`docker compose config` 断言端口绑定 `127.0.0.1`；CI 增加 `gitleaks`/`trufflehog` 扫描。

### SEC-03 · P2 · 会话生命周期（改密不撤销）

- **文件:行**：`application/user/appImpl.go:74-92`；`infrastructure/persistence/identity/userRepoImpl.go:137-166`；`interfaces/controllers/user.go:213`
- **机制**：改密仅 `UPDATE users SET password`（bcrypt cost 12）并清资料缓存，完全不触碰 `user_sessions`；而鉴权只校验 JWT 未过期 + 会话行存在且未过期。
- **前置/可利用条件**：攻击者已持有该用户任一有效 token（token 泄露、共享设备、日志外泄等）。
- **影响**：受害者「改密自救」无效，攻击者访问权最长延续至会话绝对期（7 天）；结合 SEC-04，攻击者还可在线爆破旧口令后自行改密实现持久化。
- **证据类型**：`读码确证`。
- **最小修法**：改密成功后调用既有 `DeleteByUserIdExceptToken(userId, currentToken)`（或全部删除并要求重新登录），并写安全审计事件。
- **验证/回归方法**：集成测试建两个会话 → 改密 → 旧 token 请求断言 10041，当前 token 仍可用；改密后 `user_sessions` 行数断言。

### SEC-04 · P2 · 限流与暴力破解

- **文件:行**：`interfaces/routers/authRouter.go:19-22`；`interfaces/middlewares/rateLimit.go:14`；`infrastructure/persistence/identity/rateLimitRepoImpl.go:26-60`
- **机制**：限额键为 `ClientIP + ":" + 桶名`，固定 60 秒窗口（Lua 原子），**没有账户/邮箱维度**的失败计数与锁定；换 IP（IPv6 隐私地址、代理池、移动网络）即可重置配额；Redis 异常时 `Allow` 返回放行。
- **前置/可利用条件**：分布式来源，或 Redis 抖动/被打挂（叠加 SEC-10）。
- **影响**：登录在线口令猜测（8 次/分钟/IP）、改密旧口令爆破（16 次/分钟/IP，成功即接管）；每次尝试都触发 bcrypt(cost 12)，构成 CPU 放大。
- **证据类型**：`读码确证`；实际绕过成本 `待运行时验证`。
- **最小修法**：增加 per-account 失败计数 + 指数退避/临时锁定（Redis 键含 email 哈希，避免 PII）；改密单独更严桶并要求二次验证；认证类接口 fail-closed。
- **验证/回归方法**：单测矩阵「同 IP 不同 email / 不同 IP 同 email」；Redis 断开时断言登录被拒（或降级只读）；锁定后正确口令仍拒绝直至窗口结束。

### SEC-05 · P2 · 口令策略

- **文件:行**：`interfaces/types/identity.go:17-21`；`domain/identity/valueobjects/createUser.go:19-34`、`:36-46`；对照 `interfaces/controllers/user.go:213`
- **机制**：注册 `Password` 只有 `binding:"required"`，`CreateUser.Validate()` 只判非空；改密路径要求 8–32 字符，两条路径策略不一致。bcrypt 上限 72 字节由库返回错误，但无显式校验与提示。
- **前置/可利用条件**：目标用户自行选择弱口令；注册接口开放。
- **影响**：1 字符口令可注册，撞库/猜测即接管；与 SEC-04 叠加显著降低爆破成本。
- **证据类型**：`读码确证`。
- **最小修法**：注册复用 8–32 字符（按 rune）+ 常见弱口令黑名单；拒绝 >72 字节并返回明确错误码。
- **验证/回归方法**：单测 `"1"`、`"1234567"`、73 字节口令均被拒；与改密路径共用同一校验函数（避免再次漂移）。

### SEC-06 · P2 · 凭据经 URL 传输

- **文件:行**：`interfaces/middlewares/jwtValidator.go:14-18`；`docs/avatar-access-frontend-report.md:29-33`
- **机制**：`getJwtString` 对**所有**受保护路由优先取 `ctx.Query("token")`，其次才看 `Authorization`；头像 `<img>` 与 EventSource 依赖该机制。
- **前置/可利用条件**：存在记录 query 的日志链路（反代/CDN/APM）、浏览器历史/分享链接、或客户端扩展读取；服务端自身 `RequestLogger` 只记录 `URL.Path`（不含 query），已确认不是泄露点。
- **影响**：有效期最长 7 天的会话 token 落入日志/截图/分享 → 直接账号接管；同时扩大 SEC-07 的批量抓取可行性。
- **证据类型**：`读码确证`（服务端接受）；日志链路实际形态 `待运行时验证`。
- **最小修法**：仅对 `/static/uploads/avatars/*` 与 `/api/sse/*` 允许 query token，其余路由一律只认 `Authorization`；如需长期保留，改为一次性短 TTL 的下载票据（由登录态换取）。
- **验证/回归方法**：单测 `GET /api/tasks?token=...` 断言 10041、带 Header 断言通过；反代 access log 样例断言不含 `token=`。

### SEC-07 · P2 · 头像横向读取与账号枚举

- **文件:行**：`interfaces/routers/routers.go:95-99`；`interfaces/controllers/user.go:335-355`；`infrastructure/storage/avatarStorage.go:16-24`；`docs/avatar-access-frontend-report.md:5-9`
- **机制**：头像文件名恒为 `{userId}.{ext}`（`application/user/appImpl.go:144`），读取端点只校验「已登录」，不校验归属；该路由挂在 `/api` 组之外，**没有 RateLimiter、也没有 RequestLogger**。
- **前置/可利用条件**：任意有效账号 + 可猜测/枚举的雪花用户 ID。
- **影响**：跨用户批量抓取头像（人脸等 PII）；`200 / 404` 差异可确认账号存在；无限流使批量抓取成本极低。
- **证据类型**：`读码确证`（该设计在前端落地文档中已被明确声明为「任意已登录用户可查看任意头像」，属已知取舍，但仍构成跨用户读取）。
- **最小修法**：文件名改为不可猜测随机串并落库映射；或读取时校验「本人 / 同任务评论参与者」；路由挂限流与访问日志。
- **验证/回归方法**：两账号交叉请求断言 403/404；并发压测断言触发 10051；枚举 100 个不存在 ID 断言被限流。

### SEC-08 · P2 · 上传内容与 URL 头像校验不足

- **文件:行**：`application/user/appImpl.go:123-158`；`interfaces/controllers/user.go:258-306`
- **机制**：文件上传仅以 `filepath.Ext` 白名单（jpg/jpeg/png）判定，不校验魔数/真实 Content-Type、不重编码；JSON 方式 `avatarURL` 接受任意字符串（外链、任意 scheme），服务端不做协议/域名约束。
- **前置/可利用条件**：已登录；受害者客户端渲染该字段（渲染方式 `待运行时验证`）。
- **影响**：非图片内容以 `image/*` 落盘（`X-Content-Type-Options: nosniff` 已缓解浏览器嗅探，但依赖客户端/中间层行为）；外链头像使每个查看者向第三方发起请求（IP/行为跟踪）；若客户端以 `href`/`v-html` 渲染则可能触发 XSS。
- **证据类型**：服务端无校验 `读码确证`；最终影响 `待运行时验证`（客户端渲染路径）。
- **最小修法**：读取前 512 字节做 `http.DetectContentType` 魔数校验并按内容重编码；URL 方式仅允许 `https://` 且走域名白名单（或直接禁用以外的 scheme/外链）。
- **验证/回归方法**：上传「HTML 内容 + .png 后缀」断言 400；提交 `javascript:`/`data:`/`file:` URL 断言 400；正常 jpg/png 回归通过。

### SEC-09 · P2 · 容器与镜像加固

- **文件:行**：`Dockerfile:1-22`；`docker-compose.yml:63-95`
- **机制**：运行阶段 `FROM golang:1.25.0`（含编译器、shell、包管理器），无 `USER`（root 运行），`COPY . .` 且仓库**无 `.dockerignore`**（`.env`、`.git`、`logs/`、`uploads/` 进入构建上下文），compose 未设置 `read_only`/`cap_drop`/`no-new-privileges`。
- **前置/可利用条件**：需要先有代码执行或任意文件写入（本条为纵深防御项）。
- **影响**：一旦被利用，攻击者获得 root + 工具链，可横向/持久化；构建上下文中的本地 `.env` 可能在构建缓存/镜像层留痕。
- **证据类型**：`读码确证`（Dockerfile/compose 静态事实）。
- **最小修法**：运行阶段换 `gcr.io/distroless/static` 或 `scratch` + 非 root UID；新增 `.dockerignore`（`.env`、`.git`、`logs`、`uploads`、`docs`）；compose 加 `read_only: true`、`tmpfs: /tmp`、`cap_drop: [ALL]`、`security_opt: [no-new-privileges:true]`。
- **验证/回归方法**：`docker run --rm <image> id` 断言非 uid 0；`docker history` 断言无 `.env`；可选 trivy 基线对比。

### SEC-10 · P2 · 限流 fail-open

- **文件:行**：`infrastructure/persistence/identity/rateLimitRepoImpl.go:52-60`
- **机制**：Redis `Eval` 出错时记录告警并 `return true, nil`（放行），覆盖登录/注册/改密/全部业务桶。
- **前置/可利用条件**：Redis 不可用、超时或被打挂（可与 SEC-01 组合）。
- **影响**：限流整体失效，SEC-04/05 的爆破与 SEC-01 的批量请求失去最后一道节流。
- **证据类型**：`读码确证`（策略选择在注释中已声明，属有意取舍，但缺少分接口差异化）。
- **最小修法**：认证类接口 fail-closed；业务接口可保留 fail-open 但必须告警 + 本地内存令牌桶兜底。
- **验证/回归方法**：单测注入 Redis 错误，断言 `signin` 拒绝、`tasks` 放行且产生告警日志。

### SEC-11 · P2 · 可信代理与限流全局化

- **文件:行**：`interfaces/routers/routers.go:35`；`interfaces/middlewares/rateLimit.go:14`；`docs/plans/server-deploy.md:313-320`
- **机制**：`router.SetTrustedProxies(nil)` 表示不信任任何代理，`ClientIP()` 取 `RemoteAddr`；而部署文档推荐前置 nginx 并注入 `X-Real-IP`/`X-Forwarded-For`，此时应用看到的源 IP 恒为反代地址。
- **前置/可利用条件**：现网（或未来）按文档前置反向代理/负载均衡。
- **影响**：8 次/分钟的登录桶成为**全站共享**，单个攻击者即可让所有人无法登录/改密（可用性 DoS）；审计日志中的 IP 全部失真。
- **证据类型**：`读码确证`（代码与文档组合矛盾）；现网是否前置反代 `待运行时验证`。
- **最小修法**：若前置反代：`SetTrustedProxies([反代所在内网 CIDR])` 并确保仅内网可达；若坚持直连：删除部署文档中的 nginx 段并显式声明「不得前置反代」。
- **验证/回归方法**：集成测试在「信任 / 不信任」两态下带 `X-Forwarded-For` 断言 `ClientIP`；反代后压测断言不同真实 IP 使用不同桶。

### SEC-12 · P2 · 错误详情回显与跨租户存在性 oracle

- **文件:行**：`interfaces/controllers/user.go:190`、`:320`、`:384`、`:419`、`:444`、`:479`；`interfaces/controllers/sync.go:139-141`、`:158`、`:191`
- **机制**：多处把 `err.Error()` 原样放入响应（`Error`/`Data` 字段），包括仓储/驱动错误、文件系统路径、MySQL 重复键消息；`/sync/push` 逐条回传 raw error。
- **前置/可利用条件**：任意已登录账号；构造超长字段、非法文件名、与已存在 ID 冲突的同步条目。
- **影响**：泄露表名/列名/文件绝对路径/主键值；「Duplicate entry '<id>' for key 'PRIMARY'」可判断**他人**雪花 ID 是否存在（跨租户存在性 oracle），为后续定向攻击提供信息。
- **证据类型**：`读码确证`（回显点）；具体报文形态 `待运行时验证`（依赖 DB 严格模式与驱动版本）。
- **最小修法**：统一错误映射层：领域错误 → 稳定 `code`；raw error 只写日志（带 `request_id`）；重复键统一返回 `conflict` 且不带键值。
- **验证/回归方法**：golden test 断言错误响应只含白名单 code/message；构造重复 ID 与超长字段断言响应不含 `Duplicate`/`PRIMARY`/`/opt/`/SQL 关键字。

### SEC-13 · P3 · 用户枚举

- **文件:行**：`interfaces/controllers/identity.go:148`（注册「该邮箱已被注册」）；`application/auth/appImpl.go:36-52`（登录：未知邮箱直接返回、跳过 bcrypt，产生时序差）；`interfaces/controllers/identity.go:100-116`（失败日志与统一文案）；配合 SEC-07 头像 404
- **机制**：登录响应文案已统一，但「用户不存在」与「口令错误」在耗时上可区分（仅在用户存在时执行 bcrypt）；注册接口显式区分邮箱是否已存在；头像端点 200/404 也可作为旁路。
- **前置/可利用条件**：无需登录（登录/注册接口本身就开放）。
- **影响**：批量确认邮箱是否注册，为撞库、钓鱼、定向攻击提供目标清单（PII 关联）。
- **证据类型**：`读码确证`（分支与 bcrypt 调用位置）；时序差幅度 `待运行时验证`。
- **最小修法**：登录对未知邮箱执行一次固定 dummy bcrypt；注册改为「若可注册将发送验证邮件」的统一文案。
- **验证/回归方法**：时序基准测试（已知/未知邮箱 p95 差 < 10%）；注册响应文案快照测试。

### SEC-14 · P3 · 安全响应头与缓存策略

- **文件:行**：`interfaces/routers/routers.go:14-27`；`cmd/main.go:37-54`
- **机制**：已设置 `X-Frame-Options`/`X-Content-Type-Options`/`Referrer-Policy`，但**无 HSTS**、无 CSP；`/api/user/profile`、`/api/user/sessions`、`/api/user/config` 等含 PII 的响应未设置 `Cache-Control: no-store`（仅头像显式 `private, immutable`）。
- **前置/可利用条件**：共享终端/中间缓存/降级到 HTTP 的场景。
- **影响**：协议降级与缓存残留导致 PII/会话元数据泄露；浏览器侧缺少纵深防御。
- **证据类型**：`读码确证`。
- **最小修法**：TLS 分支加 `Strict-Transport-Security: max-age=31536000; includeSubDomains`；API 默认 `Cache-Control: no-store`（头像/静态资源单独覆盖）。
- **验证/回归方法**：httptest 断言各端点响应头；回归确认头像仍为 `private, max-age=31536000, immutable`。

### SEC-15 · P3 · 审计与敏感数据落盘

- **文件:行**：`interfaces/controllers/identity.go:107-114`；`infrastructure/persistence/cache/cache.go:55-57`；`infrastructure/persistence/models/user.go:94`
- **机制**：成功登录、改密、会话下线/全下线、账号注销均无结构化安全审计；失败登录日志写入邮箱明文（无 IP 关联字段）；会话 token 明文存 `user_sessions.token`，且直接拼进 Redis key（`cache:session:<uid>:<JWT>`）。
- **前置/可利用条件**：需要日志平台/Redis/DB 的读取权限（内部或旁路泄露）。
- **影响**：安全事件无法闭环取证（谁在何时改密/踢下线）；日志与 Redis 监控（slowlog、`KEYS`、备份）成为 token/PII 的二次泄露面。
- **证据类型**：`读码确证`。
- **最小修法**：新增安全事件审计（事件类型/user_id/IP/UA/结果，**不含 token**）；Redis key 与会话表改存 token 的 keyed hash；日志对邮箱做脱敏或仅存哈希。
- **验证/回归方法**：审计事件快照测试；`redis-cli --scan` 断言不出现 `eyJ` 前缀；日志样例断言无完整邮箱。

### SEC-16 · P3 · 客户端可控时间戳与增量游标

- **文件:行**：`interfaces/types/task.go:49-54`；`infrastructure/persistence/task/repoImpl.go:100-141`；`interfaces/types/sync.go:141-143`
- **机制**：create/upsert 接受客户端 `createdAt/updatedAt/deletedAt`；LWW 判定使用客户端时间（仅写入时把 `updated_at` 覆盖为服务端 now）；`/sync/pull` 的 keyset 游标同样来自客户端回传值。
- **前置/可利用条件**：正常客户端或篡改客户端（仅影响自身数据）。
- **影响**：推送未来时间戳可使自身后续真实变更落在游标之前 → 增量拉取长期漏数据/不同步；也可让自身写入长期被判 `noop/stale`。非跨租户，但属数据完整性风险。
- **证据类型**：`读码确证`（判定与游标来源）；具体漏数据形态 `待运行时验证`。
- **最小修法**：服务端钳制客户端时间（不得晚于 `now + 允许偏移`，如 5 分钟）；pull 游标上限钳到 `now`。
- **验证/回归方法**：单测未来 `updatedAt` 被钳制/拒绝；双端同步回归用例（先推未来时间行，再改同一条，断言仍能拉取到）。

### SEC-17 · P3 · 依赖与供应链

- **文件:行**：`go.mod:10`（`go-redis/redis/v8`，上游已停止维护）、`:12`（`file-rotatelogs` 停更）、`:13`（`lfshook` 2018 年后无提交）；`.github/workflows/ci.yml:32-33`、`:53`、`:90-91`（Actions 使用 tag 而非 commit SHA）
- **机制**：依赖清单无已知 CVE 扫描门禁（CI 仅 build/vet/gofmt/lint/unit/integration），Actions 以可移动 tag 引用。
- **前置/可利用条件**：上游新披露漏洞；或第三方 Action tag 被重指（供应链投毒）。
- **影响**：漏洞无法在合并前暴露；CI 具备仓库读权限，tag 被篡改可影响构建产物。
- **证据类型**：版本与 CI 配置 `读码确证`；是否存在可利用 CVE `待运行时验证`（需 govulncheck/OSV）。
- **最小修法**：CI 增加 `govulncheck ./...`（high 阻断）；Actions 固定到 commit SHA（可用 dependabot 维护）；go-redis v8 → v9 单独立项迁移。
- **验证/回归方法**：新增 job 首跑建立基线；`go list -m -u all` 报告输出归档；SHA 固定后 CI 仍全绿。

### SEC-18 · P3 · HTTP 分支缺少服务器超时

- **文件:行**：`cmd/main.go:49-66`
- **机制**：只有 TLS 分支构造了带 `ReadTimeout/WriteTimeout/IdleTimeout` 的 `http.Server`；HTTP 分支直接 `router.Run(addr)`（默认无超时）。另外 `WriteTimeout: 15s` 对 SSE 长连接不友好（功能性副作用）。
- **前置/可利用条件**：以非 TLS 方式运行（本地、调试、误配未挂证书），且可达端口。
- **影响**：Slowloris 类慢连接可长期占用连接；SSE 每 15 秒被写超时打断（客户端频繁重连）。
- **证据类型**：`读码确证`。
- **最小修法**：两分支共用同一 `http.Server`（加 `ReadHeaderTimeout`），HTTP 分支也设超时；SSE 路由单独豁免/提高写超时。
- **验证/回归方法**：慢速发送请求头断言在超时后被断开；SSE 连接保持 >60 秒不断流。

---

## 4. 已修项（#58 / #60 / #61）的残余风险

> 已合并修复本身经读码确认有效，以下仅为**残余面**，不与已修项重复计数，也不建议回退修复。

| 来源 | 残余风险 | 位置 | 建议 |
| --- | --- | --- | --- |
| #60（登出校验 RowsAffected） | `DeleteById`（下线指定会话）**仍未校验 `RowsAffected`**：先 `First` 再 `Delete` 存在竞态窗口，命中 0 行仍返回成功（10030），与 #60 同类语义缺陷，但只影响「下线指定设备」提示 | `infrastructure/persistence/identity/sessionRepoImpl.go:188-223`；`interfaces/controllers/session.go:77-120` | 与 `Delete` 同口径：`RowsAffected==0` 返回 `ErrSessionNotFound` |
| #60 同族 | `DeleteByUserId`（`:226`）/ `DeleteByUserIdExceptToken`（`:258`）不校验 `RowsAffected`（删除 0 行返回成功）；作为幂等可接受，但「退出其他设备」提示可能与事实不符 | `sessionRepoImpl.go:226-300` | 至少在响应中返回实际下线数量 |
| #58（signin 复活软删会话） | 复活依赖 `(user_id, device_id)` 唯一索引；老客户端不传 `X-Device-Id` 时每次登录**新增**会话行，无数量上限（与 SEC-01 叠加可放大 DB 增长）；`device_id` 由客户端任意指定，可用来顶掉自己同设备的旧会话（仅自身可用性） | `sessionRepoImpl.go:40-107`；`interfaces/middlewares/clientInfo.go:51-54` | 对空 `device_id` 的会话数设上限/或强制要求；`assertSessionVisible` 可追加 device 维度校验 |
| #56 会话清理（cron） | 硬删 `deleted_at < now-24h` 的会话后，失去「该 token 曾被主动登出」的审计痕迹；SSE 的 `IsSessionValid` 有 5 分钟正向缓存，登出/删除后 SSE 通道最长 5 分钟内仍可能被判有效（**API 鉴权走 DB，不受影响**） | `sessionRepoImpl.go:438-464`；`infrastructure/initialize.go:147-149`；`infrastructure/persistence/identity/sessionRepoImpl.go` 清理段 | 审计需求若要求留痕，改为「保留已登出会话 ≥ 审计期」或删除前写入审计表 |

---

## 5. 附录 A：排除项与误报口径

### A.1 排除项（已核对为「未发现」，非漏报）

| 检查项 | 结论 | 依据 |
| --- | --- | --- |
| SQL / ORM 注入（`sort` / `order by` / 动态列 / raw SQL） | **未发现** | `infrastructure/utils/query/sorting.go:36-55` 字段与方向双白名单，未命中即忽略；`SyncOrder`/`ByKeysetCursor` 固定列；`dbs/mysql.go` 的 raw SQL 仅启动迁移且参数化；`sessionRepoImpl` 清理 raw DELETE 参数化 + 常量列 |
| 同步删除表名注入 | **未发现** | `interfaces/controllers/sync.go:255-278` 使用 `switch d.Table` 白名单分派，未知表直接报错 |
| LIKE 拼接注入 | **未发现** | `task/scopes.go:33-100`、`repoImpl.go:1139` 全部使用 `?` 占位符（通配符仅影响匹配范围，不构成注入） |
| IDOR / 跨租户读写 | **未发现** | 抽查任务/子任务/检查项/评论/清单/偏好/标签/番茄/会话全部仓储查询均带 `user_id`；`GetById`/`Update`/`Delete`/`Restore`/`BatchUpdate` 均 `(id + user_id)`；SSE Hub 按 `userId` 隔离频道 |
| Mass assignment（请求体注入 userId/role） | **未发现** | `interfaces/types/*` 无 `userId`/`role` 字段；服务端一律从上下文取 `userId` 传入应用层 |
| 路径穿越（头像） | **未发现** | `avatarStorage.go:18-24` 正则 `^[0-9]{1,20}\.(jpg\|jpeg\|png)$` 锚定整串；保存文件名由服务端 `{userId}{ext}` 生成 |
| SSRF | **未发现** | 服务端无出站 HTTP 客户端（`http.Get/Post/Client` 零命中）；头像 URL 不做服务端拉取；ip2region 读本地 xdb |
| CSRF | **不适用** | 无 Cookie 会话，凭证只经 `Authorization`（及 query）；CORS 精确白名单 + 回显 origin |
| JWT 算法混淆 / `none` | **未发现** | `infrastructure/auth/jwt.go:111` `WithValidMethods([]string{"HS256"})`；签名密钥固定 |
| 令牌伪造即越权 | **未发现** | `Validate` 除验签+exp 外还要求 `(user_id, token)` 存在于 `user_sessions` 且未过期（`application/auth/appImpl.go:204-231`）；**这是本条不评 P0 的直接原因** |
| 会话轮换竞态 | **未发现** | `UpdateToken` CAS（`user_id + old_token`）+ `RowsAffected` 校验（`sessionRepoImpl.go:403-425`） |
| 绝对有效期 | **符合设计** | `sessionRepoImpl.go:50` 固定 `now + 7d`，checkin 不延长 |
| 口令存储 | **符合基线** | bcrypt cost 12（`userRepoImpl.go:150`、`createUser.go:40`），自带盐，`users.password size:64` 容纳 60 字符哈希；无明文/无弱哈希 |
| 越权删除他人会话 | **未发现** | `Delete`/`DeleteById`/`DeleteByUserId*` 均带 `user_id` |
| 客户端指定时间戳导致跨用户影响 | **未发现** | 时间戳只作用于本人行（见 SEC-16 的自身完整性影响） |

### A.2 误报口径（本报告如何判定「不算」）

1. **「JWT 默认密钥入库」不直接定 P0**：鉴权是「JWT + 会话行」双校验，知道密钥也无法凭空造出可用 token。故按「凭据泄露 + 配置静默回退」计 P1，并写明升级为 P0 的**前置条件**（DB/Redis 可达且口令一致）。这样既不夸大，也不漏掉真实攻击链。
2. **「任意登录用户可看任意头像」已有文档声明**：`docs/avatar-access-frontend-report.md` 明确写了该取舍，因此不按「未知设计缺陷」上报；但因文件名可直接枚举 + 路由无限流，仍计 **P2（横向读取 + 枚举）**，并标注「设计已声明」。
3. **`?token=` 不直接定 P1**：服务端自身日志只记 `URL.Path`，不记 query；能否泄露取决于反代/CDN/客户端。故计 P2 并标注 `待运行时验证`。
4. **`fail-open` 不按漏洞定级为 P1**：代码注释已声明「避免限流服务故障导致全站拒绝」，属可用性取舍；但覆盖了认证接口，故计 P2 并给分接口差异化修法。
5. **CORS 放行 Vercel 预览域不误报**：正则锚定 `nathan3303s-projects` 后缀、拒绝 `*.vercel.app`、拒绝 `vercel.app.evil.com`，已有单测矩阵（`interfaces/routers/cors_test.go`）。
6. **`ByTags`/`ByTaskName` 的 LIKE 通配符不误报为注入**：用户可放大匹配范围，但仍受 `user_id` 约束，且参数化。
7. **`config.yaml:9 debug: true` 不误报为「生产开 debug」**：`conf/config.go:197` 用 `APP_DEBUG == "true"` 无条件覆盖，yaml 值实际是死配置（仅作观察记录，不单列发现）。
8. **`WriteTimeout: 15s` 不误报为安全漏洞**：它是 SSE 的功能性问题，归入 SEC-18 的附带说明，不单独计级。
9. **会话 token 明文入库不按 P0**：这是「可撤销的 JWT 会话」必然形态（需原样比对），泄露前提是 DB 已被读取；归入 SEC-15（硬化）而非越权缺陷。
10. **不把「需要受害者客户端配合」的影响写成确证**：SEC-06/SEC-08/SEC-16 的最终影响均标 `待运行时验证`，不写成已复现。

---

## 6. 附录 B · 生产自查清单（只读命令 · 均可在 VPS 上直接粘贴）

> **自查纪律**：本节所有命令均为**只读**（`docker ps/inspect/stats`、`printenv`、`SHOW`/`CONFIG GET`、`grep`、`ss`、`curl -I`、`openssl s_client`）。**不要**执行 `docker compose down/up`、`CONFIG SET`、`FLUSHALL`、`ALTER/SET` 等写操作；输出含口令/密钥的命令请**不要**截图外传。命令中的容器名/路径按现网命名（`naotodo-server`/`naotodo-mysql`/`naotodo-redis`、`/opt/nao-todo-server`）编写，不符时替换即可。
>
> 补充事实：PM 外部探测显示 `3306/6379/3301/3302/5432` 均为 closed/filtered；若域名仍可访问，说明入口经 CDN/反代或端口映射不同，请以实际入口核对 B.9/B.10。

### B.1 容器端口发布（SEC-02）

```bash
docker ps --format 'table {{.Names}}\t{{.Ports}}'
```
- **期望**：`naotodo-mysql` / `naotodo-redis` 的映射必须带 `127.0.0.1:` 前缀；出现 `0.0.0.0:` 或 `[::]:` 即不符。
- **不符处置**：compose 改 `127.0.0.1:3306:3306` / `127.0.0.1:6379:6379`（变更，需窗口）。

### B.2 主机监听面（SEC-02）

```bash
sudo ss -ltnp | grep -E ':(3306|6379|3302)\b'
```
- **期望**：3306/6379 只出现在 `127.0.0.1`（或不存在）；3302 可为 `0.0.0.0`（对外服务）。
- **不符处置**：优先改 compose 绑定 + 主机防火墙（ufw/安全组）双保险。

### B.3 应用关键 env 是否为空/过短（SEC-02，不回显值）

```bash
docker exec naotodo-server sh -c 'for k in JWT_SECRET MYSQL_PASSWORD REDIS_PASSWORD; do v=$(printenv "$k"); printf "%s len=%s\n" "$k" "${#v}"; done'
```
- **期望**：三项 `len` 均 > 0，且 `JWT_SECRET` 的 `len` ≥ 32。
- **不符处置**：补齐 `.env` 后重建应用容器；**空值即命中「静默回退到仓库默认密钥」**。

### B.4 线上 JWT 密钥是否等于仓库默认值（SEC-02，哈希比对，不打印密钥）

```bash
docker exec naotodo-server printenv JWT_SECRET | tr -d '\n' | sha256sum
git -C /opt/nao-todo-server show origin/main:conf/config.yaml | awk -F': ' '/^ *jwtSecret:/{print $2}' | tr -d ' \r' | sha256sum
```
- **期望**：两行哈希**不同**（若部署目录不是 git clone，跳过第二条并改用发行包内的 `conf/config.yaml`）。
- **不符处置**：立即轮换 `JWT_SECRET`（会强制全员重新登录），并在下次发布前清掉仓库内默认值。

### B.5 MySQL 账号与传输（SEC-02）

```bash
docker exec naotodo-mysql sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -N -e "SELECT @@bind_address, @@require_secure_transport; SELECT user,host FROM mysql.user WHERE user=\"root\";"' 2>/dev/null
```
- **期望**：`bind_address` 为 `127.0.0.1` 或容器内网地址（**不是** `0.0.0.0`）；`require_secure_transport=ON`；root 只允许 `localhost`/内网；应用不应长期用 root。
- **不符处置**：`my.cnf` 改 `bind-address=127.0.0.1`（容器网络内仍可用服务名访问）、评估开启 `require_secure_transport`、为应用建最小权限账号。

### B.6 Redis 绑定/口令/保护模式（SEC-02、SEC-10）

```bash
PW=$(docker exec naotodo-server printenv REDIS_PASSWORD)
docker exec naotodo-redis redis-cli -a "$PW" --no-auth-warning CONFIG GET bind
docker exec naotodo-redis redis-cli -a "$PW" --no-auth-warning CONFIG GET protected-mode
docker exec naotodo-redis redis-cli -a "$PW" --no-auth-warning CONFIG GET requirepass | tail -n1 | wc -c
```
- **期望**：`bind` 为 `127.0.0.1`（部署文档示例给的 `bind 0.0.0.0` **不符**）；`protected-mode yes`；`requirepass` 长度 > 1（非空，**不要**把输出贴到聊天/工单）。
- **不符处置**：改 `bind 127.0.0.1` + `requirepass`（用容器网络访问，不需要对外监听）。

### B.7 内存上限与当前占用（SEC-01）

```bash
docker inspect -f '{{.Name}} mem={{.HostConfig.Memory}} restart={{.HostConfig.RestartPolicy.Name}}' naotodo-server naotodo-mysql
docker stats --no-stream --format '{{.Name}} {{.MemUsage}}'
```
- **期望**：app `mem=536870912`（512m）；记录当前 `MemUsage` 作为基线，异常增长需排查。
- **不符处置**：应用侧加 `http.MaxBytesReader` + `limit/批量` 上限（RD）；有反代时可临时加 `client_max_body_size 2m`，直连部署则必须代码侧限制。

### B.8 容器权限（SEC-09）

```bash
docker inspect -f 'user={{.Config.User}} ro={{.HostConfig.ReadonlyRootfs}} caps={{.HostConfig.CapDrop}} sec={{.HostConfig.SecurityOpt}}' naotodo-server
docker exec naotodo-server id -u
```
- **期望**：`user` 非空且不为 `root`/`0`；`ro=true`；`caps=[ALL]`；`sec` 含 `no-new-privileges`；`id -u` 输出非 0。
- **不符处置**：换 distroless + 非 root UID，加 `read_only`/`tmpfs`/`cap_drop`/`no-new-privileges`（镜像与 compose 变更）。

### B.9 是否存在反代与真实 IP 配置（SEC-11）

```bash
command -v nginx >/dev/null && sudo nginx -T 2>/dev/null | grep -nE 'proxy_pass|set_real_ip_from|real_ip_header|client_max_body_size' || echo 'no nginx'
ss -tnp state established '( sport = :3302 )' | head
```
- **期望**：**无 nginx（直连）**，或「有 nginx + `set_real_ip_from` + `X-Forwarded-For` + 应用侧 `SetTrustedProxies` 同步配置」三件套齐全。
- **不符处置**：只配一半会让全站共享同一个 IP 限流桶（SEC-11）——要么补齐两端口径，要么确认直连并在文档删除反代段。

### B.10 日志里是否出现会话 token（SEC-06）

```bash
sudo grep -c 'token=' /var/log/nginx/*.log 2>/dev/null | head
grep -c 'token=' /opt/nao-todo-server/app/logs/app.log* 2>/dev/null
```
- **期望**：均为 `0`（应用自身只记 `URL.Path`，正常应为 0；反代若记录 query 会 > 0）。
- **不符处置**：若 > 0 ⇒ 视为已泄露，**吊销全部会话**（登录后「退出其他全部设备」或改密）+ 反代日志脱敏/改日志格式。

### B.11 头像目录文件名与内容（SEC-07、SEC-08）

```bash
sudo find /opt/nao-todo-server/app/uploads/avatars -type f -printf '%f\n' | grep -vE '^[0-9]{1,20}\.(jpg|jpeg|png)$' | head
sudo find /opt/nao-todo-server/app/uploads/avatars -type f -printf '%s %p\n' | sort -rn | head -5
```
- **期望**：第一条输出为空（文件名合法）；第二条最大文件 ≤ 2 MB 量级。
- **不符处置**：异常文件人工核查来源；服务端补魔数校验与重编码（RD）。

### B.12 安全响应头（SEC-14）

```bash
curl -sSI https://todobe.nathanao.space/api/ping | grep -iE 'strict-transport-security|cache-control|x-content-type-options|x-frame-options'
```
- **期望**：至少含 `x-content-type-options: nosniff`；`strict-transport-security` 建议存在（当前 SEC-14 判定为缺失）。
- **不符处置**：由 RD 在 TLS 分支补 HSTS，API 默认 `Cache-Control: no-store`。

### B.13 部署目录与仓库版本一致性（SEC-02 前提）

```bash
git -C /opt/nao-todo-server status --porcelain
git -C /opt/nao-todo-server rev-parse --short HEAD
```
- **期望**：`status` 无输出（无本地漂移）；`HEAD` 与本次发布 tag/commit 一致（本次审计基线 `c834404`）。
- **不符处置**：先确认「线上隐式改动」内容（尤其 compose/env），再决定纳入仓库还是回退。

### B.14 TLS 证书有效期（SEC-14，端口按实际入口替换）

```bash
openssl s_client -connect todobe.nathanao.space:443 -servername todobe.nathanao.space </dev/null 2>/dev/null | openssl x509 -noout -subject -dates
```
- **期望**：`notAfter` 未过期；`subject/SAN` 与域名一致。
- **不符处置**：续签证书并 `docker compose restart app`（变更，需窗口）。

### B.15 日志 PII 与登录失败留痕（SEC-13、SEC-15）

```bash
grep -c 'UserSignIn rejected' /opt/nao-todo-server/app/logs/app.log* 2>/dev/null
ls -l /opt/nao-todo-server/app/logs/ | head
```
- **期望**：失败统计可读、日志轮转存在（`maxAge=30`）；存量中包含邮箱（SEC-15 已知）。
- **不符处置**：无需立即处置；若要收敛 PII，改为记录邮箱哈希 + 增加安全事件审计字段（RD）。

### B.16 自查结果回填建议

| 检查项 | 关联发现 | 结果（✅/❌） | 备注 |
| --- | --- | --- | --- |
| B.1/B.2 端口发布与监听 | SEC-02 |  |  |
| B.3/B.4 env 与默认密钥 | SEC-02 |  |  |
| B.5/B.6 MySQL/Redis 硬化 | SEC-02、SEC-10 |  |  |
| B.7 内存与基线 | SEC-01 |  |  |
| B.8 容器权限 | SEC-09 |  |  |
| B.9 反代与真实 IP | SEC-11 |  |  |
| B.10 日志 token | SEC-06 |  |  |
| B.11 头像文件 | SEC-07、SEC-08 |  |  |
| B.12/B.14 安全头与证书 | SEC-14 |  |  |
| B.13 部署漂移 | SEC-02 |  |  |
| B.15 日志 PII | SEC-13、SEC-15 |  |  |

> 自查只回答「现网是否命中前置条件」，不改变 §3 的定级；命中 ❌ 的项按对应「不符处置」拆单给 RD/运维。


---

**审计人**：arch-designer（T560） · **审计基线**：`c834404` · **P0/P1/P2/P3 = 0/2/10/6** · **证据纪律**：`读码确证` 与 `待运行时验证` 已逐条区分。
