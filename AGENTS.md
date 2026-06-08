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

All tests are in `cmd/server/main_test.go`. No external framework, no integration prerequisites.

## Codebase Map

The application is modularized into multiple files within the single `main` package in the `cmd/server/` directory for better readability:

- `cmd/server/main.go`: Application entrypoint, route configuration, and `go:embed` filesystem initialization for templates.
- `cmd/server/models.go`: Core data structures (`User`, `Activity`, `ProductVerification`, etc.) and the main `App` and `Config` struct definitions.
- `cmd/server/handlers.go`: HTTP route handlers for pages and API endpoints (e.g., `/login`, `/dashboard`, `/api/*`).
- `cmd/server/auth.go`: Authentication logic, session cookie management, and role-based access control middleware.
- `cmd/server/db.go`: Database connectivity, auto-migrations, seeders, and specialized data queries (Postgres and Oracle).
- `cmd/server/utils.go`: Utility functions such as environment configuration parsing, logging middleware, and JSON responses.
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
