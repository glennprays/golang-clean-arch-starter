# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go Clean Architecture boilerplate template using Fiber framework with Wire dependency injection. Requires Go 1.25+, Docker, and Make.

## Common Commands

```bash
# Run the API server
make run

# Run development services (PostgreSQL, Swagger UI)
make run-dev

# Stop development services
make stop-dev

# Refresh Swagger UI after updating docs
make swagger

# Rename module for new project
make rename RENAME_MODULE_TO=github.com/yourname/yourproject

# Regenerate Wire dependencies (after modifying wire.go)
go generate ./internal/infrastructure/...
```

## Architecture

### Clean Architecture Layers

```
cmd/api/main.go           # Entry point, Fiber server
config/                   # Configuration loading (Viper + struct tags)
internal/
├── apperror/             # Typed application error model (Kind + helpers)
├── domain/               # Core entities & value objects (innermost)
├── usecase/              # Application business rules
├── service/              # Reusable business logic
├── repository/           # Data access layer
├── handler/              # HTTP handlers (Fiber)
├── middleware/           # HTTP middleware (TraceID, recover, CORS, logger, error handler)
├── router/               # Route definitions
├── httperror/            # JSON contract for error responses (ErrorResponse/ErrorBody)
├── infrastructure/       # Wire DI, App struct
├── utils/                # Helper utilities
└── worker/               # Background jobs
pkg/
├── logger/               # Logger provider (wraps glennprays/log)
└── logctx/               # Trace ID propagation through context.Context
```

### Dependency Rule

Dependencies point inward:
- `domain` → no dependencies on other layers
- `apperror` → no dependencies on other layers (used everywhere errors cross a boundary)
- `usecase/service` → can import `domain`, `apperror`
- `repository` → can import `domain`, `apperror`
- `handler` → can import `usecase`, `service`, `domain`, `apperror`

### Key Components

**Application errors** (`internal/apperror/`):
- One `*apperror.Error` carries `Kind`, `Message`, optional `Cause` (logged, never sent to clients), optional `Details` (validation).
- Constructors: `apperror.NotFoundf(...)`, `apperror.BadRequest(...)`, `apperror.Internal(...).Wrap(err)`, etc.
- Sentinels for `errors.Is`: `apperror.ErrNotFound`, `apperror.ErrConflict`, ... — match by `Kind`.
- Design parallels `k8s.io/apimachinery/pkg/api/errors` and `gocloud.dev/gcerrors`.

**Infrastructure** (`internal/infrastructure/`):
- `app.go` - App struct holding dependencies
- `wire.go` / `wire_gen.go` - Wire DI configuration

**Error handling flow**:
- Usecases/services/handlers return `*apperror.Error` (often via the per-kind constructors).
- `middleware.ErrorHandler` (Fiber global handler) maps the error to a typed `httperror.ErrorResponse` with `code`, `message`, `trace_id`, and optional `details`. 5xx and unmapped errors are logged before the response.

**Trace ID propagation**:
- `middleware.TraceID` generates/validates a UUID per request and stores it in both `c.Locals` (Fiber) and `c.UserContext()` (Go context).
- Downstream code calls `logctx.TraceID(ctx)` to retrieve it for log correlation.

### API Endpoints

- `GET /api/v1/health` - Health check

### Development Environment

Docker Compose (`misc/develop/docker-compose.yml`):
- PostgreSQL 16.3 on port 5432
- Swagger UI on port 8080

Environment: `.env` (copy from `.env.example`). App runs on port 3000.
