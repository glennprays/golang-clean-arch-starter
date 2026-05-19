# Golang Clean Architecture Starter

A production-leaning starter for HTTP services in Go. Fiber for routing, Wire
for compile-time DI, structured logging with per-request trace IDs, and a
typed error model with a unified response envelope. It is opinionated about
the bones (error handling, observability, request shape) and deliberately
unopinionated about what you build on top.

> Status: the HTTP and observability surface is feature-complete and tested.
> The data layer (`pgx`/`sqlc`/migrations) is intentionally **not** wired
> yet — see [What's not included](#whats-not-included).

---

## What you get

- **Typed application errors** (`internal/apperror`). One `Kind` enum, per-kind
  constructors (`apperror.NotFoundf(...)`, `.Wrap(cause)`), nil-safe accessors,
  `errors.Is/Unwrap` support, and structured `Details[]` for field-level
  validation. Pattern mirrors [`k8s.io/apimachinery/pkg/api/errors`][k8s-err]
  and [`gocloud.dev/gcerrors`][gocloud-err].
- **Unified response envelope** for every endpoint:
  ```jsonc
  // success
  { "data": { ... }, "trace_id": "f47ac10b-..." }
  // error
  { "error": { "code": "NOT_FOUND", "message": "...", "details": [...] },
    "trace_id": "f47ac10b-..." }
  ```
- **Trace ID propagation** through both `c.Locals` (for Fiber middleware) and
  `context.Context` (for usecases/services/repositories), so every log line
  carries a `trace_id` that matches the `X-Trace-ID` response header.
- **Global error handler** that logs every 5xx and every unmatched error with
  its trace ID before responding — no more silent 500s.
- **Configuration with validation** via Viper + `go-playground/validator/v10`.
  Required fields and enum constraints are checked at boot; production refuses
  to start with a wildcard CORS origin.
- **Request body validation** through a one-line generic helper:
  `dto, err := httperror.BindAndValidate[CreateUserDTO](c)`.
- **Graceful shutdown** on `SIGINT`/`SIGTERM` with an explicit timeout context.
- **Explicit server limits** (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`,
  `BodyLimit`) — Fiber's defaults are not appropriate for production.
- **Wire DI** for compile-time dependency injection — no reflection, no magic.
- **Unit tests + `golangci-lint`** baseline (`make test`, `make lint`) and
  a `make rename` target to re-namespace the module for your project.

[k8s-err]: https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors
[gocloud-err]: https://pkg.go.dev/gocloud.dev/gcerrors

---

## Quick start

### Prerequisites

- Go 1.25+
- Docker (for the local Postgres / Swagger UI containers)
- Make

### 1. Clone and rename

```bash
git clone git@github.com:glennprays/golang-clean-arch-starter.git my-service
cd my-service
make rename RENAME_MODULE_TO=github.com/yourname/my-service
```

`make rename` rewrites the module path in `go.mod` and every `.go` import.
Verify, then optionally reset git history:

```bash
rm -rf .git && git init && git add . && git commit -m "Initial commit"
```

### 2. Configure

```bash
cp .env.example .env
# edit .env if you want non-defaults
```

`DB_PASSWORD` and `DB_NAME` are required even though the data layer isn't
wired — config validation enforces them so future code can rely on them
being present. See [Configuration](#configuration).

### 3. Run

```bash
make run-dev   # starts Postgres + Swagger UI in Docker
make run       # runs the API once on $APP_PORT (default 3000)
make dev       # runs the API with hot reload via air (recompiles on file save)
```

Hit it:

```bash
curl -i http://localhost:3000/api/v1/health
```

```
HTTP/1.1 200 OK
X-Trace-Id: 24585beb-2bfb-4a60-b0b5-b4bb1f05ab26
Content-Type: application/json

{"data":{"status":"ok","timestamp":"2026-..."},"trace_id":"24585beb-..."}
```

---

## Project layout

```
cmd/api/main.go              # Entry point: Fiber server, graceful shutdown
config/                      # Viper config with struct-tag defaults + validation
internal/
├── apperror/                # Typed error model (Kind enum + constructors)
├── domain/                  # Pure entities & value objects (innermost)
│   ├── entity/
│   └── valueobject/
├── usecase/                 # Application business rules
├── service/                 # Reusable business logic
├── repository/              # Data access (interfaces consumed by usecase)
├── handler/                 # HTTP handlers (Fiber)
├── middleware/              # TraceID, recover, CORS, logger, error_handler
├── router/                  # Route definitions + middleware registration
├── httperror/               # Response envelope (Envelope, OK, BindAndValidate)
├── infrastructure/          # Wire DI: providers + App struct
├── params/                  # Request/response DTOs (placeholders)
├── utils/                   # Helpers
└── worker/                  # Background jobs (placeholder)
pkg/
├── logger/                  # Logger provider (wraps glennprays/log → zap)
└── logctx/                  # Trace ID through context.Context
migrations/                  # SQL migrations (placeholder — see deferred section)
docs/                        # Swagger / OpenAPI
misc/develop/                # Docker compose for local dev
```

### The dependency rule

Dependencies only point inward — outer layers depend on inner ones, never the
reverse.

| Layer                         | May import                              |
| ----------------------------- | --------------------------------------- |
| `domain`                      | (nothing else in this repo)             |
| `apperror`                    | (nothing else in this repo)             |
| `usecase`, `service`          | `domain`, `apperror`                    |
| `repository`                  | `domain`, `apperror`                    |
| `handler`                     | `usecase`, `service`, `domain`, `apperror`, `httperror` |
| `middleware`, `router`        | `handler`, `httperror`, `apperror`, `config` |
| `infrastructure`              | (everything — assembles the graph via Wire) |

`pkg/logctx` and `pkg/logger` are leaf packages and can be imported from any
layer.

---

## Patterns this starter establishes

### Returning errors

Don't return raw Go errors out of usecases or handlers — use `apperror`:

```go
import "github.com/yourname/my-service/internal/apperror"

func (uc *GetUserUseCase) Execute(ctx context.Context, id int) (*domain.User, error) {
    u, err := uc.repo.FindByID(ctx, id)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, apperror.NotFoundf("user %d not found", id)
    }
    if err != nil {
        return nil, apperror.Internal("fetch user failed").Wrap(err)
    }
    return u, nil
}
```

Why this shape:

- **Self-documenting at the call site** — `apperror.NotFoundf` names the
  intent in the function name, no `serviceErr`/`appErr` argument pair to
  remember.
- **`errors.Is(err, apperror.ErrNotFound)` works** — handy in usecases that
  branch on category. The `Is` method compares by `Kind`, not pointer
  identity.
- **`.Wrap(err)` attaches the cause** — the cause goes to the logs but is
  *never* sent to the client. Clients see only the safe `Message`.
- **The global error handler does the rest** — status code, JSON shape,
  logging at 5xx, request log line. Handlers just `return err`.

Available kinds: `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`,
`Conflict`, `Validation`, `Internal`. Each has plain (`NotFound`) and
formatted (`NotFoundf`) variants, plus a sentinel (`ErrNotFound`) for
`errors.Is`.

### Validating request bodies

```go
import "github.com/yourname/my-service/internal/httperror"

type CreateUserDTO struct {
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age"   validate:"required,min=18"`
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
    dto, err := httperror.BindAndValidate[CreateUserDTO](c)
    if err != nil {
        return err  // global handler emits VALIDATION_FAILED + details[]
    }
    // ... call usecase ...
}
```

A failed validation produces a `400` with `code: "VALIDATION_FAILED"` and a
`details[]` array where each entry names the JSON field, the failed rule, and
a short message:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "request validation failed",
    "details": [
      { "field": "email", "rule": "email",    "message": "failed on 'email' rule" },
      { "field": "age",   "rule": "min",      "message": "failed on 'min=18' rule" }
    ]
  },
  "trace_id": "..."
}
```

Field names in `details[]` come from the `json` struct tag, not the Go field
name — clients see the same identifier they sent.

### Returning success

Use the response helper so the envelope stays consistent:

```go
func (h *UserHandler) Get(c *fiber.Ctx) error {
    user, err := h.usecase.GetByID(c.UserContext(), id)
    if err != nil {
        return err
    }
    return httperror.OK(c, user)  // wraps in { data, trace_id }
}
```

### Returning paginated lists

`OKPaginated` carries both the items and the pagination metadata in
the same envelope. The `Page` struct supports two schemes — pick one
per endpoint:

```go
// Cursor (Stripe-shaped):
return httperror.OKPaginated(c, users, httperror.Page{
    NextCursor: nextID,
})

// Offset (classic):
return httperror.OKPaginated(c, users, httperror.Page{
    Page: 2, PageSize: 20, Total: 137,
})
```

Response:
```json
{
  "data": [ { "...": "..." } ],
  "page": { "next_cursor": "user_abc123" },
  "trace_id": "f47ac10b-..."
}
```

### Logging with trace IDs

The `TraceID` middleware generates (or validates and reuses) a UUID per
request and threads it through both `c.Locals` and `context.Context`.
Downstream code pulls it from context for log correlation:

```go
import "github.com/yourname/my-service/pkg/logctx"

func (s *UserService) Refresh(ctx context.Context) error {
    s.logger.Info(logctx.TraceID(ctx), "refreshing user cache", nil)
    // ...
}
```

Every request emits exactly one structured `"http request"` log line — from
the HTTP-logger middleware on success, from the global error handler on
errors — both reporting the actual response status that went out on the wire.

### Testing handlers

`internal/testkit.NewTestApp` builds a Fiber app with the production
middleware stack (TraceID, ErrorHandler) and a quiet logger. Pass a
register callback to mount the routes under test:

```go
func TestHealthHandler(t *testing.T) {
    h := handler.NewHealthHandler()
    app := testkit.NewTestApp(t, func(app *fiber.App) {
        app.Get("/health", h.Check)
    })

    resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
    // assert status, X-Trace-Id header, JSON envelope...
}
```

See `internal/handler/health_test.go` for a worked example that asserts
both the envelope shape and that `trace_id` matches between header and body.

---

## Configuration

Loaded from environment variables (with `.env` support in development).
Required fields fail at boot via `validate.Struct(cfg)`.

| Variable                 | Default                              | Validation                         |
| ------------------------ | ------------------------------------ | ---------------------------------- |
| `ENV`                    | `development`                        | one of `development \| staging \| production` |
| `APP_NAME`               | `golang-clean-architecture`          | required                           |
| `APP_PORT`               | `3000`                               | 1–65535                            |
| `LOG_LEVEL`              | `debug`                              | required                           |
| `LOG_OUTPUT`             | `stdout`                             | required                           |
| `DB_HOST`                | `localhost`                          | required                           |
| `DB_PORT`                | `5432`                               | 1–65535                            |
| `DB_USER`                | `postgres`                           | required                           |
| `DB_PASSWORD`            | —                                    | **required, no default**           |
| `DB_NAME`                | —                                    | **required, no default**           |
| `CORS_ALLOWED_ORIGINS`   | `*`                                  | required; **refused** when `ENV=production` and value is `*` |

The boot guard for CORS in production is intentional — a wildcard origin is
almost always a misconfiguration in prod.

---

## Common commands

```bash
make run          # go run cmd/api/main.go
make dev          # hot reload via air (recompiles on save)
make run-dev      # docker compose: Postgres (5432) + Swagger UI (8080)
make stop-dev     # tear those down
make swagger      # restart the Swagger UI container after editing docs/

make test         # go test ./... -race -count=1
make lint         # golangci-lint run ./...   (requires golangci-lint built on Go 1.25+)
make tidy         # go mod tidy
make hooks        # install pre-commit hooks via lefthook (one-time setup)

make generate     # wire gen ./internal/infrastructure/...   (after editing wire.go)
make rename RENAME_MODULE_TO=github.com/you/proj
```

`make dev` requires [`air`](https://github.com/air-verse/air); `make hooks`
requires [`lefthook`](https://github.com/evilmartians/lefthook). Install both:

```bash
go install github.com/air-verse/air@latest
go install github.com/evilmartians/lefthook@latest
```

A GitHub Actions workflow at `.github/workflows/ci.yml` runs `go build`,
`go vet`, `make test`, `golangci-lint`, and `docker build` on every PR.

---

## API endpoints

| Method | Path                 | Notes                                                                            |
| ------ | -------------------- | -------------------------------------------------------------------------------- |
| GET    | `/api/v1/health`     | Liveness check; returns `status` and `timestamp` in the unified envelope.        |
| GET    | `/api/v1/version`    | Build metadata: `version`, `commit`, `build_time` (embedded via `-ldflags`).     |
| GET    | `/debug/pprof/*`     | Standard Go profiling endpoints. Mounted **only when `ENV != production`**.     |

Build metadata is populated by `make build` (or `docker build`), which passes
`VERSION`, `COMMIT`, and `BUILD_TIME` through `-ldflags`. `go run` / `make run`
leaves the defaults (`"dev"` / `"unknown"`) which is fine for local work.

Add new routes in `internal/router/router.go` — the `setupHealthRoutes`
function is the template.

---

## What's not included

This starter ships the HTTP and observability bones. The following are
deliberately deferred so you can pick the right tool for your service —
adding them is the next step you'll likely take.

- **Database layer**: no driver, no connection pool, no ORM. Recommended:
  [`pgx/v5`](https://github.com/jackc/pgx) for the driver and
  [`sqlc`](https://sqlc.dev) for type-safe queries (avoid ORMs in Clean
  Architecture — they fight the repository pattern).
- **Migrations**: `migrations/` directory exists; the runner does not.
  Recommended: [`golang-migrate`](https://github.com/golang-migrate/migrate)
  as a separate CLI step, not on app boot.
- **Deep health checks**: `/api/v1/health` is liveness only. A future
  `/api/v1/health/ready` should ping the DB and any downstream dependencies.
- **Auth**: no JWT/session/OAuth scaffold — too dependent on your auth model
  to ship blind. Add `gofiber/contrib/jwt` or similar when the first protected
  endpoint lands.
- **Rate limiting & security headers**: Fiber has `middleware/limiter` and
  `gofiber/helmet`. Add when you have a public endpoint.
- **Metrics / distributed tracing**: trace IDs in logs are usually enough at
  starter scale. Add `prometheus/client_golang` and OpenTelemetry when you
  have a real reason.

---

## Tech stack

| Concern                | Choice                                                                 |
| ---------------------- | ---------------------------------------------------------------------- |
| HTTP framework         | [Fiber v2](https://gofiber.io/) (fasthttp under the hood)             |
| Dependency injection   | [Wire](https://github.com/google/wire) — compile-time, no reflection  |
| Configuration          | [Viper](https://github.com/spf13/viper) + [godotenv](https://github.com/joho/godotenv) + [creasty/defaults](https://github.com/creasty/defaults) |
| Validation             | [`go-playground/validator/v10`](https://github.com/go-playground/validator) |
| Logging                | `github.com/glennprays/log` (structured JSON, Zap underneath)         |
| Testing                | [`testify`](https://github.com/stretchr/testify), `-race -count=1`    |
| Linting                | [`golangci-lint`](https://golangci-lint.run/) — errcheck, errorlint, govet, ineffassign, staticcheck, unused, gosec, contextcheck |
| Local dev services     | Docker Compose (Postgres 16, Swagger UI)                              |

---

## Contributing / extending

When you add a new feature, you typically touch:

1. `internal/domain/entity/` — the entity (e.g., `User`).
2. `internal/repository/` — interface (in `usecase` or alongside the
   implementation) and the implementation.
3. `internal/usecase/` — the use case.
4. `internal/handler/` — the handler. Use `httperror.BindAndValidate[T]` for
   request parsing.
5. `internal/router/router.go` — wire the route in.
6. `internal/infrastructure/wire.go` — register providers; run `make generate`.

Return `apperror` from anywhere below the handler; let the global error
handler do the HTTP mapping.
