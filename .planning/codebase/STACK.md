# Technology Stack

**Analysis Date:** 2026-06-08

## Languages

**Primary:**
- Go 1.23 — All application code is in `/home/bblosangeles/go_simp/cmd/server/` (single `main` package). No sub-packages.

**Secondary:**
- HTML/Go Templates — Server-rendered HTML with Go `html/template`, embedded via `go:embed`
- CSS — Two stylesheets: `style.css` and `admin.css`
- JavaScript — Client-side JS files (HTMX-driven, no framework): `shared.js`, `app.js`, `dashboard.js`, `admin.js`, `login.js`

## Runtime

**Environment:**
- Compiled Go binary (self-contained, no runtime dependency). Uses `net/http` HTTP/1.1 server.

**Package Manager:**
- Go Modules (`go.mod` / `go.sum` at project root)

**Lockfile:**
- `go.sum` present — standard Go checksum database

## Frameworks

**Core:**
- None — Uses Go standard library `net/http` exclusively. No external web framework.

**Templating:**
- Go `html/template` — templates embedded via `go:embed` directive in `cmd/server/main.go:206`:
  ```go
  //go:embed templates/*.html templates/components/*.html templates/*.css templates/*.js
  var templatesFS embed.FS
  ```

**Testing:**
- Go standard `testing` package — all tests in `cmd/server/main_test.go`. No external test frameworks.

**Frontend:**
- HTMX (embedded at `cmd/server/templates/htmx.min.js`) — served as a static file route. No npm/node build step.
- No JS frameworks (React, Vue, etc.)

## Key Dependencies

**Critical — Postgres Driver:**
- `github.com/jackc/pgx/v5 v5.7.2` — Database driver for Postgres, used via `database/sql` stdlib (stdlib driver imported at `cmd/server/main.go:16`: `_ "github.com/jackc/pgx/v5/stdlib"`)
- Transitive deps: `pgpassfile`, `pgservicefile`, `puddle/v2` (connection pool)

**Critical — Oracle Driver:**
- `github.com/sijms/go-ora/v2 v2.8.24` — Oracle database driver, used via `database/sql` (imported at `cmd/server/main.go:17`: `_ "github.com/sijms/go-ora/v2"`)
- Oracle connection is **read-only** guarded by `isReadOnlySQL()` in `cmd/server/db.go:29`

**Critical — Cryptography:**
- `golang.org/x/crypto v0.37.0` — Used for `bcrypt` password hashing in `cmd/server/auth.go` and `cmd/server/handlers.go`

**Transitive:**
- `golang.org/x/sync v0.13.0` — Used by pgx pool
- `golang.org/x/text v0.24.0` — Used by pgx

## Configuration

**Environment:**
- `.env` file loaded manually via `loadDotEnv()` in `cmd/server/utils.go:96` (custom key-value parser, NOT godotenv)
- Environment variables override `.env` values (checked by `os.LookupEnv` before `os.Setenv`)

**Key Configs (all via env vars):**

| Env Variable | Default | Required | Purpose |
|---|---|---|---|
| `PORT` | `3000` | No | HTTP listen port |
| `APP_ENV` | `development` | No | Environment name; enables HSTS in production |
| `SESSION_SECRET` | — | Yes (min 32 chars) | HMAC signing key for session tokens |
| `SESSION_TTL` | `8h` | No | Session cookie lifetime (Go duration) |
| `POSTGRES_URL` / `DATABASE_URL` | — | Yes | Postgres connection string |
| `PG_MAX_CONNS` | `10` | No | Postgres pool max open connections |
| `ORACLE_URL` | — | No (see individual vars) | Oracle connection string (wins over individual vars) |
| `ORACLE_HOST` | `localhost` | No | Oracle host |
| `ORACLE_PORT` | `1521` | No | Oracle port |
| `ORACLE_SERVICE` | `xe` | No | Oracle service name |
| `ORACLE_USER` | — | No | Oracle username (required if no ORACLE_URL) |
| `ORACLE_PASSWORD` | — | No | Oracle password (required if no ORACLE_URL) |
| `ORACLE_MAX_CONNS` | `5` | No | Oracle pool max open connections |
| `ORACLE_IDLE_CONNS` | `3` | No | Oracle pool max idle connections |
| `ORACLE_IDLE_TIME` | `5m` | No | Oracle conn max idle time (Go duration) |

**Build:**
- No build config file (pure Go, no build tags or CGo used)
- Hot-reload via `air` — config at `.air.toml` (builds to `.tmp/main`, watches `.go`, `.html`, `.tpl`, `.tmpl`)

## Platform Requirements

**Development:**
- Go 1.23+
- Air (optional, for hot reload): `go install github.com/air-verse/air@latest`
- Postgres instance (local or remote)
- Oracle instance (optional, some features degrade gracefully)

**Production:**
- Compiled Go binary — architecture: `linux/amd64` (no platform-specific deps detected)
- Postgres database
- Oracle database (optional — app runs without it, health check reports `oracle_down`)

---

*Stack analysis: 2026-06-08*
