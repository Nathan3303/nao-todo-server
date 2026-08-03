# 头像文件访问权限落地计划

## 摘要

当前头像文件保存在 `uploads/avatars/` 目录，通过 `router.Static("/static/uploads", "./uploads")` 以开放静态目录方式提供，任何未登录请求都能直接访问任意用户头像文件。

本次改动将头像访问改为**需登录（JWT）鉴权**的接口提供：任何已登录用户可查看任意头像，未登录用户无法访问（产品决策：头像在任务评论等共享场景需对登录用户可见，详见「假设与决策」）。

**核心策略：保持 URL 路径** **`/static/uploads/avatars/{filename}`** **不变**，仅把「静态目录直出」替换为「带 JWT 鉴权的路由处理」。

* 数据库中存量 `users.avatar` / `task_comments.avatar` 值无需任何迁移

* 前端 `<img src>` 的 URL 结构不变，只需附加 `?token=<JWT>`（JWTValidator 已支持 query 参数，SSE 已采用同款模式）

* 生产 nginx 中 `/static/uploads` 的 proxy 段继续有效

## 现状分析

* 上传：`UpdateAvatarByFile`（[appImpl.go](file:///c:/Users/LEE19/Projects/nao-todo-server/application/user/appImpl.go#L122-L161)）生成文件名 `{userId}.{ext}`（仅 `.jpg/.jpeg/.png`），经 `avatarStorageImpl.Save` 写入 `uploads/avatars/`，返回 `/static/uploads/avatars/{userId}.{ext}` 存入 DB。

* 开放暴露点：[routers.go](file:///c:/Users/LEE19/Projects/nao-todo-server/interfaces/routers/routers.go#L73-L75) `router.Static("/static/uploads", "./uploads")`。

* 存储端口：`AvatarStorage`（[avatarStorage.go](file:///c:/Users/LEE19/Projects/nao-todo-server/application/user/avatarStorage.go)）只有 `Save` / `Delete`，没有读取能力。

* 鉴权能力：`JWTValidator`（[jwtValidator.go](file:///c:/Users/LEE19/Projects/nao-todo-server/interfaces/middlewares/jwtValidator.go)）支持 `Authorization: Bearer` 头或 `?token=` 查询参数，校验后把 `userId` 写入 context。

* 头像使用场景：用户本人资料页 + 任务评论（头像 URL 冗余存储于 `task_comments` 表，任何登录用户可看他人头像）。

* `uploads` 目录当前仅存放头像（任务评论的 `Attachments` 是外部 URL 字符串，不落地本地）。

## 变更清单

### 1. 存储端口新增读取能力（3 个文件）

**`application/user/avatarStorage.go`** — 端口新增方法：

```go
// Open 打开已存在的头像文件流，供 HTTP 响应
Open(ctx context.Context, filename string) (io.ReadCloser, error)
```

**`infrastructure/storage/avatarStorage.go`** — 实现 `Open`：

* 新增包级校验函数 `isSafeAvatarFilename(filename string) bool`，正则 `^[0-9]{1,20}\.(jpg|jpeg|png)$`，杜绝路径穿越（`..`、`/`、反斜杠均不匹配），且只放行上传白名单格式。

* `Open`：校验通过后 `os.Open(filepath.Join(avatarDir(), filename))`，失败返回错误。

**`application/user/app.go`** **+** **`appImpl.go`** — 应用层新增：

```go
// 接口
GetAvatar(ctx context.Context, filename string) (io.ReadCloser, error)
// 实现：直接委托 u.avatarStorage.Open(ctx, filename)
```

### 2. 控制器新增头像响应处理器

**`interfaces/controllers/user.go`** — 新增 `GetAvatar`：

1. `filename := ctx.Param("filename")`
2. `f, err := c.userApp.GetAvatar(ctx.Request.Context(), filename)`；失败 → `FailureByHttpStatus(ctx, http.StatusNotFound, ...)`（404，不暴露文件是否存在细节）
3. 按扩展名设置 Content-Type（`mime.TypeByExtension`，`.jpg/.jpeg`→`image/jpeg`，`.png`→`image/png`；空则 `application/octet-stream`）
4. 设置缓存头 `Cache-Control: private, max-age=31536000, immutable`（文件名不变则内容不可变，但受鉴权约束故用 `private` 避免共享缓存）
5. `ctx.DataFromReader(http.StatusOK, -1, contentType, f, extraHeaders)` 流式输出

### 3. 路由：移除静态目录，注册鉴权路由

**`interfaces/routers/routers.go`**：

* 删除 `router.Static("/static/uploads", "./uploads")`（当前 L73-75）。

* 在函数末尾（`return router` 前，复用已声明的 `userCtrl`）注册：

```go
router.GET(
    "/static/uploads/avatars/:filename",
    middlewares.JWTValidator(svc.Auth),
    userCtrl.GetAvatar,
)
```

### 4. nginx 参考配置注释更新

**`.example/nginx.sub.conf.example`** — 在 `/static/uploads` location 块中，把「直接服务静态文件（alias）」分支的注释改为**醒目警告：不得启用** **`alias`** **直连静态目录，会绕过 JWT 鉴权**；保留当前 proxy 到 Go 的方式。

### 5. 不改动项（明确说明）

* 头像上传逻辑（`UpdateAvatarByFile`、`Save`、`Delete`）、`users.avatar` 存储格式：不变。

* DB 数据：**零迁移**。

* `conf` 配置（`Uploads.StaticPath` 保持 `/static/uploads`）：不变。

* 全局 JWTValidator 的失败响应（HTTP 200 + JSON body）为项目统一风格，本次不改；对 `<img>` 而言结果是坏图 → 前端 `onerror` 兜底。

## 前端配合（另一仓库，需同步）

* 渲染头像 `<img>` 时附加 `?token=<JWT>`（复用 SSE 已有的 query token 模式）；或改用带 `Authorization` 头的 fetch/blob 加载。

* 图片加载失败（401/404）时显示默认头像兜底。

## 假设与决策

1. **登录用户可查看任意头像，不做所有权校验**（用户已确认）——仅需鉴权，不校验 filename 是否属于当前用户。
2. **保持 URL 路径不变**：避免 DB 迁移与前端 URL 结构调整，改动面最小。
3. **文件名正则白名单**是唯一的安全边界（防路径穿越 + 只读本地头像文件）；不存在的文件自然 404。
4. 注销用户 7 天后的文件清理现状（`DeleteDeactivatedUsers` 不删文件）不在本次范围；如需可另立任务。
5. 头像 GET 暂不加 RateLimiter（此前静态目录也无）；如需防滥用可后续在路由上追加。

## 验证步骤

1. `go build ./...`、`go vet ./...` 通过。
2. 启动服务（`fresh` 或 `go run ./cmd`）。
3. 未登录 `GET /static/uploads/avatars/{id}.jpg` → 返回 `{code:10041}`（不返回图片）。
4. 登录后 `GET /static/uploads/avatars/{id}.jpg?token=<jwt>` → 200，Content-Type `image/jpeg`，图片内容正确。
5. 路径穿越尝试 `GET /static/uploads/avatars/..%2F..%2Fconf%2Fconfig.yaml?token=<jwt>` → 404。
6. 非法文件名 `GET /static/uploads/avatars/abc.jpg?token=<jwt>` → 404。
7. 确认 `/static/uploads/` 下其他路径不再被静态直出（404）。

