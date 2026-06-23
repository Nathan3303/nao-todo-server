# 计划：开发环境不部署 Nginx，生产环境部署 Nginx

## 摘要

利用 Docker Compose 的多文件覆盖机制，将 nginx 服务从 `docker-compose.yml` 中移出，放入新建的 `docker-compose.prod.yml`。开发环境只使用基础 compose 文件，生产环境通过 `-f` 叠加两个文件启动。

## 当前状态分析

- `docker-compose.yml` 包含 4 个服务：mysql、redis、app、nginx
- 没有 Makefile 或其他编排脚本
- 没有 `docker-compose.override.yml` 或其他 compose 文件
- `.example/.env.example` 中没有环境区分变量

## 变更内容

### 1. `docker-compose.yml` — 移除 nginx 服务

删除第 98-115 行的 `nginx` 服务定义。其余服务（mysql、redis、app）保持不变。

### 2. 新建 `docker-compose.prod.yml` — 仅包含 nginx 服务

```yaml
services:
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

与当前 nginx 定义完全一致，无需修改。

### 3. `.example/nginx.conf.example` — 更新部署说明

更新顶部注释，区分开发/生产环境的启动命令。

### 4. 无需变更的部分

- app 端口 `3302:3302` 保留对外暴露，开发环境可直接访问
- 其他服务无需修改
- CORS 配置无需修改

## 使用方式

**开发环境**（无 Nginx）：
```bash
docker compose up -d
# 直接访问 http://localhost:3302
```

**生产环境**（带 Nginx HTTPS）：
```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
# 通过 https://your-domain.com 访问
```

## 假设与决策

1. **app 端口保留对外暴露**：生产环境中 `3302:3302` 仍然暴露，可通过防火墙限制外部访问。如需在生产环境关闭直接暴露，可在 `docker-compose.prod.yml` 中覆盖 app 的 ports 配置，但本次不做此变更。
2. **不引入 Makefile**：保持项目简洁，使用原生命令行。
3. **nginx 配置模板不变**：`proxy_pass http://app:3302` 已在之前改为 Docker 服务名，无需再改。

## 验证

1. `docker compose config` 确认 nginx 不在基础配置中
2. `docker compose -f docker-compose.yml -f docker-compose.prod.yml config` 确认叠加后 nginx 正确出现
3. 开发环境 `docker compose up -d` 只启动 mysql、redis、app 三个服务
