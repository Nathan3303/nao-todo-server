# 安全加固操作手册（部署前必读）

> ## ⚠️ 部署前置提醒：先轮换凭据，再部署新版本
>
> 自本版本（Issue #63 / T561）起，服务端启动时会执行**凭据 fail-fast 校验**（`conf.ValidateSecurity`，SEC-02）。
> 只要生产 `.env` 里以下任一情况成立，**新版本会直接拒绝启动并退出**：
>
> - `JWT_SECRET` 为空、长度 < 32 字节、或仍是仓库 `conf/config.yaml` 里的默认值；
> - `MYSQL_PASSWORD` / `REDIS_PASSWORD` 为空或仍是仓库默认值。
>
> 因此升级顺序必须是 **① 轮换凭据 → ② 重启/部署新版本**。若先部署再补 env，会出现「容器反复重启、全站不可用」的停机窗口。
> 另：轮换 `JWT_SECRET` 会使**所有存量登录态失效**（会话 token 由旧密钥签发），用户需重新登录 —— 请选择低峰窗口并提前告知。

---

## 1. 适用范围与边界

本手册对应安全审计报告 `docs/reports/2026-10-27-T560-security-audit.md` 的 **SEC-01 / SEC-02** 两项 P1。

| 项 | 本仓库已内置（代码侧） | 需用户在自己的部署上执行（配置侧） |
| --- | --- | --- |
| SEC-01 无界输入 | 请求体上限中间件、`limit`/`page` 上限、`/sync/push` 批量上限 | 无需操作（可按需调大 env） |
| SEC-02 默认凭据 | 启动 fail-fast 校验 | **轮换凭据**；收口端口绑定；改用最小权限库账号 |

> ⛔ **本仓库不修改 `docker-compose.yml`**（用户工作区存在未提交改动）。下文 compose 片段请**由用户在自己的 compose 文件上照抄执行**，仓库侧不代改。
> 部署目录按现网约定写作 `/opt/nao-todo-server`，与实际不符时自行替换。

---

## 2. 凭据轮换（应用侧 fail-fast）

### 2.1 要求

| 变量 | 要求 | 生成 / 设置方式 |
| --- | --- | --- |
| `JWT_SECRET` | ≥ 32 字节随机串，且不等于仓库默认值 | `openssl rand -hex 32` |
| `MYSQL_PASSWORD` | 强口令（建议 ≥ 24 字符随机串），且不等于仓库默认值 | `openssl rand -hex 24` |
| `REDIS_PASSWORD` | 强口令（建议 ≥ 24 字符随机串），且不等于仓库默认值 | `openssl rand -hex 24` |

校验失败时日志只输出「**哪个变量不合格 + 期望**」，**不会打印任何密钥/口令值**，可安全截图给运维。

### 2.2 操作步骤

```bash
# ① 在部署目录生成新值（先落盘到 .env，不要贴到聊天/工单）
cd /opt/nao-todo-server

NEW_JWT=$(openssl rand -hex 32)
NEW_MYSQL=$(openssl rand -hex 24)
NEW_REDIS=$(openssl rand -hex 24)

# ② 备份现有 .env（含旧凭据，妥善保管，排查时可回滚）
cp .env ".env.bak.$(date +%Y%m%d%H%M%S)"

# ③ 写回 .env（示例用 sed；也可手改）
sed -i "s|^JWT_SECRET=.*|JWT_SECRET=${NEW_JWT}|"     .env
sed -i "s|^MYSQL_ROOT_PASSWORD=.*|MYSQL_ROOT_PASSWORD=${NEW_MYSQL}|" .env
sed -i "s|^REDIS_PASSWORD=.*|REDIS_PASSWORD=${NEW_REDIS}|"     .env

# ④ 同步修改 Redis 的 requirepass（否则应用带新口令连不上 Redis）
#    编辑 /opt/nao-todo-server/redis/conf/redis.conf 的 requirepass 行
#    MySQL 口令由 MYSQL_ROOT_PASSWORD 在容器初始化时生效；已初始化的数据卷需额外 ALTER USER：
#    docker exec -it naotodo-mysql mysql -uroot -p'<旧口令>' \
#      -e "ALTER USER 'root'@'%' IDENTIFIED BY '<新口令>'; FLUSH PRIVILEGES;"
```

> MySQL 官方镜像的 `MYSQL_ROOT_PASSWORD` **只在数据卷首次初始化时生效**；已有 `mysql-data` 的数据卷必须用 `ALTER USER` 改密，否则容器仍是旧口令。
> Redis 同理：`requirepass` 写在挂载的 `redis.conf` 里，改完 `docker compose restart redis` 生效。
> 若 compose 的 redis healthcheck 是裸 `redis-cli ping`，带口令后会返回 `NOAUTH` 判为 unhealthy（进而阻塞 app 启动），
> 需同步改成 `["CMD-SHELL", "redis-cli --no-auth-warning -a \"$REDIS_PASSWORD\" ping"]`（口令经 env 注入）。

### 2.3 验证

```bash
# 应用能读到非空且足够长的密钥（只打印长度，不打印值）
docker exec naotodo-server sh -c \
  'for k in JWT_SECRET MYSQL_PASSWORD REDIS_PASSWORD; do
     v=$(printenv "$k"); printf "%s len=%s\n" "$k" "${#v}"; done'
# 期望：三项 len 均 > 0，JWT_SECRET len ≥ 32

# 确认线上 JWT 密钥不是仓库默认值（哈希比对，不打印密钥）
docker exec naotodo-server printenv JWT_SECRET | tr -d '\n' | sha256sum
git -C /opt/nao-todo-server show origin/main:conf/config.yaml \
  | awk -F': ' '/^ *jwtSecret:/{print $2}' | tr -d ' \r' | sha256sum
# 期望：两个哈希不同
```

### 2.4 回滚

- 应用侧：恢复备份的 `.env`（`.env.bak.*`）→ 重启容器。**不推荐**长期保留仓库默认凭据（新版本也会拒绝启动）。
- 若新 JWT 导致客户端集中掉线且需要临时恢复登录态：只能回滚 `JWT_SECRET` 到旧值（旧 token 仍有效），但这会重新引入审计发现，须尽快再轮换。

---

## 3. 端口绑定收口（DB / Redis 不对外发布）

> 具体改动由用户在自己的 compose 上执行。目标：MySQL/Redis **只监听回环**，应用容器仍通过内部网络访问，不受影响。

### 3.1 片段（照抄到自己的 `docker-compose.yml`）

```yaml
services:
    mysql:
        ports:
            # ❌ 原： - "3306:${MYSQL_PORT}"   （等价 0.0.0.0:3306，所有网卡可达）
            - "127.0.0.1:3306:3306"
    redis:
        ports:
            # ❌ 原： - "6379:${REDIS_PORT}"   （等价 0.0.0.0:6379）
            - "127.0.0.1:6379:6379"
    app:
        ports:
            - "3302:${APP_PORT}" # 应用端口保持对外，无需改动
```

若想保留原有的 env 变量风格，可写成 `"127.0.0.1:3306:${MYSQL_PORT}"` / `"127.0.0.1:6379:${REDIS_PORT}"`，效果相同（前提是 `.env` 中这两个变量为 3306 / 6379）。

### 3.2 迁移步骤

1. 低峰窗口执行；先 `docker compose config > /tmp/compose.rendered.yaml` 留档。
2. 修改 compose 的两个 `ports` 行（仅加 `127.0.0.1:` 前缀）。
3. `docker compose up -d`（compose 会重建受影响的 mysql/redis 容器；数据在绑定卷中，不会丢）。
4. 验证：见 3.3。
5. 同步检查主机防火墙/云安全组：3306/6379 不应有公网放行规则（双保险）。

### 3.3 验证

```bash
# ① 端口发布必须带 127.0.0.1 前缀
docker ps --format 'table {{.Names}}\t{{.Ports}}'
# 期望：naotodo-mysql / naotodo-redis 显示 127.0.0.1:3306->... / 127.0.0.1:6379->...

# ② 主机监听面
sudo ss -ltnp | grep -E ':(3306|6379|3302)\b'
# 期望：3306/6379 只在 127.0.0.1；3302 可为 0.0.0.0（对外服务）

# ③ 应用侧连通性（容器内网络，必须仍正常；golang 基础镜像自带 bash）
docker exec naotodo-server bash -c \
  '</dev/tcp/mysql/3306 && </dev/tcp/redis/6379 && echo CONNECT_OK'
```

### 3.4 回滚

把 `ports` 改回原样 → `docker compose up -d`。回滚会重新把 DB/Redis 暴露到所有网卡，仅限排障窗口内短时使用，事后必须改回。

---

## 4. 应用改用最小权限库账号（不再用 root 连库）

> 前提事实：应用启动时会执行 `AutoMigrate`（`infrastructure/persistence/dbs/mysql.go`），**需要 schema 级 DDL 权限**；
> 该迁移的错误目前被静默忽略，若账号缺 DDL 权限，会表现为「表结构不更新但不报错」。因此授权必须包含建表/改表权限。

### 4.1 建账号片段（在 MySQL 容器内执行）

```sql
-- 口令用 openssl rand -hex 32 生成，不要复用 root 口令
CREATE USER 'naotodo_app'@'%' IDENTIFIED BY '<强口令>';
-- schema 级授权（不是 *.*）：DML + 应用内 AutoMigrate 所需的 DDL
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, REFERENCES
  ON naotodo.* TO 'naotodo_app'@'%';
FLUSH PRIVILEGES;
```

### 4.2 compose 片段（照抄到自己的 `docker-compose.yml` 的 app 服务）

```yaml
    app:
        environment:
            # ❌ 原： - MYSQL_USER=root
            - MYSQL_USER=${MYSQL_USER}          # 在 .env 中设为 naotodo_app
            - MYSQL_PASSWORD=${MYSQL_APP_PASSWORD} # 与 4.1 的 <强口令> 一致
```

> 注意：仓库 compose 把 `MYSQL_USER` 写死为 `root`，且复用了 `MYSQL_ROOT_PASSWORD` 作为应用口令；
> 请在自己的文件里改成独立的 `MYSQL_APP_PASSWORD`，避免应用进程持有 root 口令。

### 4.3 迁移步骤

1. 按 4.1 建账号并授权；按 4.2 改 compose 与 `.env`（新增 `MYSQL_USER`、`MYSQL_APP_PASSWORD`）。
2. `docker compose up -d app` 重建应用容器。
3. 验证：见 4.4；确认日志无 DB 权限错误、`AutoMigrate` 后表结构未变。
4. 观察一个完整发布周期（含一次业务写入与一次重启）后，再把 root 口令从 `.env` 中移除。

### 4.4 验证

```bash
# 应用实际使用的库账号（应为 naotodo_app，不是 root）
docker exec naotodo-server printenv MYSQL_USER
docker exec naotodo-mysql sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -N -e \
  "SELECT user,host FROM mysql.user WHERE user=\"naotodo_app\";"'
# 期望：存在 naotodo_app 行；root 不再被应用使用
```

### 4.5 回滚

`MYSQL_USER=root` + `MYSQL_PASSWORD=${MYSQL_ROOT_PASSWORD}` → `docker compose up -d app`。
账号可保留（不删），待下次窗口重试。

---

## 5. 请求体 / 分页 / 批量上限（已内置，可按需调大）

| 能力 | 默认值 | 调整方式 | 超限行为 |
| --- | --- | --- | --- |
| 全局请求体上限 | 2 MB | env `MAX_REQUEST_BODY_BYTES`（字节） | HTTP **413**（不静默截断） |
| 头像上传请求体上限 | 5 MB | env `MAX_UPLOAD_BODY_BYTES`（字节） | HTTP **413** |
| 列表 `limit` | ≤ 500 | 代码常量 `query.MaxLimit` | 自动夹取到 500（不拒绝请求） |
| 列表 `page` | ≤ 10000 | 代码常量 `query.MaxPage` | HTTP **400**（深分页拒绝） |
| `/sync/push` 单表条数 | ≤ 500 | 代码常量 `query.MaxLimit` | HTTP **400** + 可读错误（含表名/实际条数） |
| `/sync/pull` 单表 `limit` | ≤ 500 | 同上 | 自动夹取到 500 |

调大示例（不建议超过 8 MB，容器 `mem_limit: 512m`）：

```yaml
    app:
        environment:
            - MAX_REQUEST_BODY_BYTES=4194304  # 4 MB
            - MAX_UPLOAD_BODY_BYTES=10485760  # 10 MB
```

> 回滚：上限无「关闭开关」。如遇客户端大批量同步被 400 拒绝，优先把 `query.MaxLimit` 调大并重新发布（而不是关掉校验）；
> 413 则通过上面的 env 调大（需重启容器）。

---

## 6. 上线前自查清单

- [ ] `.env` 中 `JWT_SECRET` / `MYSQL_PASSWORD` / `REDIS_PASSWORD` 已轮换（非空、非默认、JWT ≥ 32 字节）
- [ ] 已确认轮换 `JWT_SECRET` 会导致全员重新登录，并已选好窗口
- [ ] `docker ps` 中 3306 / 6379 映射均带 `127.0.0.1:` 前缀
- [ ] `sudo ss -ltnp` 中 3306 / 6379 只监听回环
- [ ] 应用已改用 `naotodo_app`（非 root）连库，且启动日志无权限错误
- [ ] 新版本启动日志无 fail-fast 报错（若报错，按提示修 env 后重启，不要跳过）

---

## 7. 附：相关代码位置

| 内容 | 位置 |
| --- | --- |
| 启动 fail-fast 校验 | `conf/security.go`（`ValidateSecurity`）、`cmd/main.go` |
| 请求体上限中间件 | `interfaces/middlewares/bodyLimit.go`、`interfaces/routers/routers.go` |
| 分页/批量上限 | `infrastructure/utils/query/pagination.go`、`interfaces/controllers/sync.go` |
| 审计原始报告 | `docs/reports/2026-10-27-T560-security-audit.md` |
