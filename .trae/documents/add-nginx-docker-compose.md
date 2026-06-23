# 计划：为 Docker Compose 添加 Nginx 反向代理服务

## 摘要

在现有 `docker-compose.yml` 中新增 `nginx` 服务，实现基于域名的 HTTPS 反向代理，将请求转发到后端 `app` 服务。同时更新 `.example/nginx.conf.example` 模板以适配 Docker 内部网络。

## 当前状态分析

- **docker-compose.yml**：现有 3 个服务（mysql、redis、app），均在 `naotodo-backend` 桥接网络中
- **app 服务**：监听 3302 端口，对外暴露 `3302:3302`
- **nginx 配置模板**：`.example/nginx.conf.example` 中 `proxy_pass` 指向 `http://127.0.0.1:3302`（宿主机地址），需要改为 Docker 服务名 `http://app:3302`
- **CORS**：`interfaces/routers/routers.go` 已配置 `AllowOrigins` 包含 `https://todo.nathanao.space`

## 变更内容

### 1. `docker-compose.yml` — 新增 nginx 服务

在 `app` 服务之后新增 `nginx` 服务定义：

```yaml
    nginx:
        image: nginx:stable-alpine
        container_name: naotodo-nginx
        ports:
            - "80:80"
            - "443:443"
        volumes:
            - /opt/nao-todo-server/nginx/conf/nginx.conf:/etc/nginx/nginx.conf:ro
            - /opt/nao-todo-server/nginx/ssl:/etc/nginx/ssl:ro
            - /opt/nao-todo-server/nginx/logs:/var/log/nginx
        depends_on:
            - app
        mem_limit: 64m
        memswap_limit: 64m
        cpus: 0.25
        networks:
            - naotodo-backend
        restart: unless-stopped
```

**说明**：
- 使用 `nginx:stable-alpine` 轻量镜像
- 暴露 80（HTTP）和 443（HTTPS）端口
- 配置文件以只读方式挂载自 `/opt/nao-todo-server/nginx/conf/nginx.conf`
- SSL 证书以只读方式挂载自 `/opt/nao-todo-server/nginx/ssl/`
- 日志目录可写挂载，方便排查问题
- 加入 `naotodo-backend` 网络，可直接通过服务名 `app` 访问后端
- 资源限制：64MB 内存、0.25 CPU，匹配 nginx 轻量特性

### 2. `.example/nginx.conf.example` — 适配 Docker 内部网络

将两处 `proxy_pass http://127.0.0.1:3302` 改为 `proxy_pass http://app:3302`，因为 nginx 与 app 在同一 Docker 网络中，应使用服务名通信。

同时更新顶部注释，说明 Docker 部署方式。

### 3. 无需变更的部分

- **app 端口暴露**：保留 `3302:3302` 映射，不影响现有直接访问方式
- **CORS 配置**：无需修改，已包含所需域名
- **其他服务**：无需修改

## 假设与决策

1. **配置文件路径**：nginx 主配置文件挂载到 `/opt/nao-todo-server/nginx/conf/nginx.conf`，遵循项目现有的 `/opt/nao-todo-server/` 数据目录惯例
2. **SSL 证书路径**：证书文件放在 `/opt/nao-todo-server/nginx/ssl/`，与模板中 `/etc/nginx/ssl/` 对应
3. **模板保留占位符**：`nginx.conf.example` 中 `your-domain.com` 等占位符保持不变，由部署者自行替换
4. **app 端口保留对外暴露**：保持向后兼容，允许不经过 nginx 直接访问（如本地开发）

## 部署步骤（供参考，不在本次变更范围内）

1. 创建目录结构：
   ```
   mkdir -p /opt/nao-todo-server/nginx/{conf,ssl,logs}
   ```
2. 复制并编辑 nginx 配置：
   ```
   cp .example/nginx.conf.example /opt/nao-todo-server/nginx/conf/nginx.conf
   # 编辑替换 your-domain.com 为实际域名
   ```
3. 放置 SSL 证书到 `/opt/nao-todo-server/nginx/ssl/`
4. 重启 docker compose：`docker compose up -d`

## 验证

1. `docker compose config` 无语法错误
2. `docker compose up -d nginx` 启动成功
3. `curl -H "Host: your-domain.com" http://localhost/api/ping` 返回正常响应
