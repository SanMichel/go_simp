# Technology Stack

**Analysis Date:** 2026-06-05

## Languages

**Primary:**
- Go 1.23.0 - All server-side logic, API handlers, HTML rendering, database access. Single `main` package in `cmd/server/`.

**Frontend:**
- JavaScript (ES5) - Plain JS files served as static assets via `go:embed`. No bundler, no framework. HTMX library bundled.
- HTML (Go `html/template`) - Server-rendered templates with `{{define}}`/`{{template}}` blocks. Embedded via `go:embed`.
- CSS - Two stylesheets served as static assets via `go:embed`.

## Runtime

**Environment:**
- Go compiled binary (single executable). No external Go runtime dependency.

**Package Manager:**
- Go modules (`go.mod` / `go.sum`)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Go standard library `net/http` - HTTP server, routing, middleware. No external web framework.
- Go `database/sql` - Database abstraction layer for both Postgres and Oracle.
- `html/template` - Server-side HTML rendering with custom `FuncMap`.

**Frontend:**
- HTMX (bundled as `htmx.min.js` in templates) - AJAX-driven page updates, partial rendering. No frontend framework.

**Testing:**
- Go standard `testing` package - All tests in `cmd/server/main_test.go`. No external test framework.
- `net/http/httptest` - HTTP handler testing.

**Build/Dev:**
- Air (`go install github.com/air-verse/air@latest`) - Hot reload during development. Config: `.air.toml`.

## Key Dependencies

**Critical (from `go.mod`):**
- `github.com/jackc/pgx/v5 v5.7.2` - Postgres driver for `database/sql`. Registered as `"pgx"` driver. Runtime connection pool configurable via env vars.
- `github.com/sijms/go-ora/v2 v2.8.24` - Oracle driver for `database/sql`. Registered as `"oracle"` driver. READ-ONLY enforced by `isReadOnlySQL()` guard.
- `golang.org/x/crypto v0.37.0` - `bcrypt` for password hashing (user auth).

**Infrastructure:**
- No external Go HTTP framework. No router library. No ORM. No logging library.
- All middleware (logging, CSRF, security headers, rate limiting) implemented in ~230 lines of hand-written code in `cmd/server/utils.go`.
- Session tokens: HMAC-SHA256 signed cookies (custom implementation in `cmd/server/auth.go`).

## Configuration

**Environment (loaded from `.env` via custom `loadDotEnv()` in `cmd/server/utils.go`):**
| Variable | Required | Default | Purpose |
|---|---|---|---|
| `PORT` | No | `3000` | HTTP listen port |
| `APP_ENV` | No | `development` | Controls HSTS header (prod only) |
| `SESSION_SECRET` | Yes | — | HMAC key for session tokens (≥32 chars) |
| `POSTGRES_URL` / `DATABASE_URL` | Yes | — | Postgres connection string |
| `ORACLE_URL` | No | built from `ORACLE_HOST/PORT/SERVICE/USER/PASSWORD` | Oracle direct connection URL |
| `ORACLE_HOST` | No | `localhost` | Oracle host |
| `ORACLE_PORT` | No | `1521` | Oracle port |
| `ORACLE_SERVICE` | No | `xe` | Oracle service name |
| `ORACLE_USER` | No | — | Oracle user |
| `ORACLE_PASSWORD` | No | — | Oracle password |
| `SESSION_TTL` | No | `8h` | Session cookie TTL (Go duration) |
| `PG_MAX_CONNS` | No | `10` | Postgres max open connections |
| `ORACLE_MAX_CONNS` | No | `5` | Oracle max open connections |
| `ORACLE_IDLE_CONNS` | No | `1` | Oracle max idle connections |

**Build:**
- `.air.toml` - Air hot-reload config: builds to `.tmp/main`, watches `.go`, `.html`, `.tpl`, `.tmpl` extensions.
- `.gitignore` - ignores `.env`, `tmp/`, `bin/`, `.tmp/`.

## Platform Requirements

**Development:**
- Go 1.23+
- Air (optional, for hot reload)
- Postgres instance
- Oracle instance (optional for dev, app warns on ping failure but continues)

**Production:**
- Compiled binary — deploy as single executable
- Postgres database (required)
- Oracle database (required for product/company/local lookups)
- No external web server required (Go serves directly)

---

*Stack analysis: 2026-06-05*
