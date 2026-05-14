# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the server
go run cmd/main.go

# Hot reload (requires fresh: go install github.com/pilu/fresh@latest)
fresh

# Production build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o nao-todo-server cmd/main.go

# Format code
go fmt ./...

# Vet code
go vet ./...
```

No tests or lint configuration exists yet.

## Architecture

This is a Go (1.24.6) task management backend using DDD with 4 layers. Module path is `naotodoserver`.

### Layer flow

```text
interfaces/          → HTTP handlers, routers, middleware, request/response types
    ↓ calls
application/         → Use-case orchestration, DTO↔value-object conversion
    ↓ calls
domain/              → Entities, value objects, repository interfaces, domain service interfaces
    ↑ implemented by
infrastructure/      → DB/GORM models, Redis, logging, cron, persistence repositories
```

### Module structure (applies to all 7 domain modules: auth, user, project, tag, task, event, comment)

Each module has the same structure across layers:

- **domain/<module>/** — `entities/`, `valueobjects/`, `repositories/` (interfaces), `service/` (domain service interface + impl)
- **application/<module>/** — `app.go` (interface + singleton), `appImpl.go` (orchestration logic), `converters.go` (request ↔ VO mapping)
- **infrastructure/persistence/<module>/** — `repoImpl.go` (GORM repository impl), `converters.go` (model ↔ entity ↔ VO mapping)
- **interfaces/controllers/<module>.go** — HTTP handlers calling `*.App.*`
- **interfaces/routers/<module>Router.go** — route registration
- **interfaces/types/<module>.go** — request/response DTOs

### Wiring (startup order in `infrastructure/initialize.go`)

1. `LoadLogger()` — Logrus with daily rotation
2. `LoadDBs()` — MySQL (GORM + AutoMigrate), Redis, Snowflake node
3. `LoadDomains()` — Each `*App.RegistDomainImpl()` wires domain service → app singleton via `sync.Once`
4. `LoadCron()` — Cron jobs (e.g., periodic cleanup of deactivated users)
5. `routers.InitRouters()` — Gin engine setup, CORS, static files, routes

### Key patterns

- **Singletons**: Each application module exposes an `App` variable initialized via `sync.Once` in `RegistDomainImpl()`. Example: `task.App.CreateTask()`, `auth.App.Validate()`.
- **User context**: JWT middleware extracts userId → stored in `context.Context` via `infrastructure/context.SetUserId()`. Application layer reads it with `context.GetUserId(ctx)`.
- **ID handling**: All IDs are int64 (Snowflake) internally, converted to string in API responses.
- **Partial updates**: Use `*string` pointer fields in request types. `NullableString` (with custom JSON unmarshaling) handles three states: not present, null, and value — used for optional datetime fields.
- **Error codes**: Each module has its own range (10xxx auth, 20xxx projects, 30xxx tags, 40xxx tasks, 50xxx events, 60xxx comments). Codes ending in 0 = success, others = specific failures.
- **Response format**: Always `{ "code": N, "message": "...", "data": ..., "pagination": ... }`. Even errors return HTTP 200 with error code in body. Controllers use `Success()` / `Failure()` helpers from `responser.go`.
- **Value object constructors**: `NewXxx()` functions validate and return `(*ValueObject, error)`.

### Auth flow

JWT-based: login → JWT token + session created in Redis → subsequent requests include `Authorization: Bearer <token>` → `JWTValidator` middleware validates and injects userId into context.

### Database

GORM with MySQL. `ModelBase` (embedded in all models) provides auto-generated Snowflake ID via `BeforeCreate` hook. `AutoMigrate` runs on startup. Models use `sql.NullTime` for nullable timestamps, converted to `*time.Time` in domain entities.

## Git commit style

Follow the format in `.trae/rules/git-commit-message.md`:

```text
feat|fix|chore|change(功能点或变更点): 描述变更

变更点：
- 具体变更描述

其他（可选）:
- 非功能点变更说明
```
