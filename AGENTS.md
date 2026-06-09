# go-simp

Single `main` package in `cmd/server/`. No sub-packages, no internal modules.

## Stack

- Go `net/http` + `database/sql` with pgx (Postgres) + go-ora (Oracle, **read-only**)
- HTMX + Go templates — templates are separate `.html` files in `cmd/server/templates/` embedded using `go:embed`. CSS is in `templates/style.css` and `templates/admin.css`.

## Quick start

```bash
cp .env.example .env   # EDIT .env first — POSTGRES_URL is required
go mod tidy
go run ./cmd/server    # or: air (hot reload via .air.toml)
```

On first start, auto-migrates tables and seeds `admin` as `sysadmin` with random password (logged on first start).

## Tests

```bash
go test ./cmd/server
```

Test helper factories are in `cmd/server/testhelper.go`. Integration tests requiring a Postgres instance are gated by `TEST_POSTGRES_URL` env var (gracefully skipped with `t.Skip` when absent).

## Codebase Map

The application is modularized into multiple files within the single `main` package in the `cmd/server/` directory for better readability:

- `cmd/server/main.go`: Application entrypoint, route configuration, and `go:embed` filesystem initialization for templates.
- `cmd/server/models.go`: Core data structures (`User`, `Activity`, `ProductVerification`, etc.) and the main `App` and `Config` struct definitions.
- `cmd/server/handlers.go`: Page/entrypoint HTTP handlers (home, login, logout, healthCheck, atividadesPage, apiMe, apiLogin, apiLogout) — ~120 lines after domain split.
- `cmd/server/activity_handlers.go`: Scanning/activity API handlers (apiEmpresas, apiFinalizar, etc.).
- `cmd/server/dashboard_handlers.go`: Dashboard page and partial handlers.
- `cmd/server/admin_handlers.go`: Admin CRUD handlers.
- `cmd/server/api_handlers.go`: Admin + dashboard API handlers (user CRUD, dashboard data, bulk print).
- `cmd/server/errors.go`: `AppError` type, error code constants, `handleError()` dispatcher, `requestIDMiddleware`, `getRequestID()`.
- `cmd/server/validation.go`: `Validator` type with chainable methods (Required, MinLength, ValidRole, Positive).
- `cmd/server/testhelper.go`: Test helper factories — `testDB()`, `testApp()`, `testUser()`, `testToken()`, `cleanupTestData()`.
- `cmd/server/auth.go`: Authentication logic, session cookie management, and role-based access control middleware.
- `cmd/server/db.go`: Database connectivity, auto-migrations, seeders, and specialized data queries (Postgres and Oracle).
- `cmd/server/utils.go`: Utility functions — environment config, `writeJSON` (buffer-first), `recoveryMiddleware`, slog-based log middleware.
- `cmd/server/main_test.go`: All test cases.
- `cmd/server/templates/`: Directory containing all HTML templates. They are compiled directly into the binary using `go:embed`.
- `tmp/`: A reference copy of the old application architecture. Do not modify.

## Key gotchas

- `.env` is loaded manually via `loadDotEnv()` — not godotenv.
- Oracle connection is guarded by `isReadOnlySQL()` — only `SELECT`/`WITH` pass.
- Templates are located in `cmd/server/templates/` and embedded into the binary. To add or modify a template, create/edit a file there and ensure it is defined using `{{define "my_template"}}`.
- CSS remains inside `main.go` as a `const` block.
- `tmp/` is a reference copy of the old app — **do not modify**
- `bin/` and `.tmp/` are build artifacts (gitignored).
- Roles: `conferente`, `gerente`, `sysadmin`

## Important
- On any change of structure or behavior of the codebase, always update `AGENTS.md` to keep the information up to date.
