# External Integrations

**Analysis Date:** 2026-06-08

## APIs & External Services

No external HTTP/API services are consumed. The application does not make outbound HTTP requests to any third-party service.

All data that requires remote access is served through direct database connections (Postgres and Oracle).

## Data Storage

**Databases (Primary — State Store):**
- **Postgres** — Application state store for users, activities, product verifications, and activity addresses.
  - Connection: `POSTGRES_URL` env var (or fallback `DATABASE_URL`)
  - Driver: `github.com/jackc/pgx/v5` via `database/sql` stdlib
  - Pool: Max 10 open connections (`PG_MAX_CONNS`)
  - Tables: `users`, `atividades`, `atividade_enderecos`, `produto_verificacao`
  - Auto-migrates on startup via `(a *App) migrate()` in `cmd/server/db.go:111`
  - Client file: `cmd/server/db.go` and `cmd/server/main.go:27-38`

**Databases (Secondary — Read-Only Lookup):**
- **Oracle** — Production source-of-truth database for product, company, and location lookups.
  - Connection: `ORACLE_URL` env var, OR built from `ORACLE_HOST`/`PORT`/`SERVICE`/`USER`/`PASSWORD`
  - Driver: `github.com/sijms/go-ora/v2` via `database/sql` stdlib
  - Pool: Max 5 connections (`ORACLE_MAX_CONNS`)
  - **Read-only enforced** — `(a *App) QueryContext()` and `(a *App) QueryRowContext()` in `cmd/server/db.go:15-27` guard with `isReadOnlySQL()`
  - Graceful degradation: On startup, Oracle ping failure is a warning (not fatal). Several API endpoints will return 500 if Oracle is unavailable.
  - Schemes queried: `CONSINCO.MRL_PRODUTOEMPRESA`, `CONSINCO.MAP_PRODUTO`, `CONSINCO.MAP_PRODCODIGO`, `CONSINCO.MRL_PRODLOCAL`, `CONSINCO.MAX_EMPRESA`, `CONSINCO.MRL_LOCAL`, `CONSINCO.ETLV_PRODUTO`
  - Oracle function used: `CONSINCO.fBuscaPrecoAtualPdv()`
  - Client file: `cmd/server/db.go`, `cmd/server/handlers.go`

**File Storage:**
- Local filesystem only (embedded assets). No external file storage (no S3, no blob storage).

**Caching:**
- None. No Redis, Memcached, or in-memory cache layer. Oracle and Postgres queries hit the database every time.

## Authentication & Identity

**Auth Provider:**
- **Custom implementation** — no external auth provider (no OAuth, no SSO, no OIDC).
- Authentication: HMAC-SHA256 signed session tokens stored in `token` cookie (HttpOnly, Secure, SameSite=Strict).
  - Implementation: `cmd/server/auth.go:21-60` (`currentUser`, `makeToken`)
- Password hashing: bcrypt via `golang.org/x/crypto` package.
- Session revocation: `last_token_at` column on `users` table — tokens issued before that timestamp are rejected.
  - Implementation: `cmd/server/auth.go:152-154` (`revokeSession`)
- CSRF protection: Double-submit cookie pattern with `csrf_token` cookie and `X-CSRF-Token` header.
  - Implementation: `cmd/server/auth.go:156-184` (`csrfMiddleware`)
- Rate limiting: In-memory per-IP rate limiter (5 attempts/minute) for login endpoints.
  - Implementation: `cmd/server/utils.go:183-225` (`rateLimiter`)

**Roles:**
- Three roles: `conferente`, `gerente`, `sysadmin`
- Role-based access enforced via `requireRole()` and `requireAPIRole()` middleware in `cmd/server/auth.go:86-131`

## Monitoring & Observability

**Error Tracking:**
- None. No Sentry, Datadog, or similar service.

**Logs:**
- Standard Go `log` package — structured console output only (stdout/stderr).
  - Request logging: `cmd/server/utils.go:130-137` — `METHOD PATH STATUS DURATION` format
  - Application errors and warnings: `log.Printf` throughout all files

**Health Check:**
- `GET /api/health` endpoint in `cmd/server/main.go:86` and `cmd/server/handlers.go:52`
  - Returns `{"status":"ok"}` or `{"status":"oracle_down"}`
  - Pings both Postgres and Oracle connections

## CI/CD & Deployment

**Hosting:**
- Not specified. Deployable as standalone Go binary. Listening on `:PORT` (default 3000).

**CI Pipeline:**
- None detected (no `.github/`, `.gitlab-ci.yml`, `Dockerfile`, or CI config in the main project; the `tmp/` directory has a `Dockerfile` but that is the old app reference copy, not to be modified).

**Containerization:**
- No `Dockerfile` in the active project.

## Environment Configuration

**Required env vars:**
| Variable | Purpose |
|---|---|
| `SESSION_SECRET` | HMAC key for session tokens (≥32 chars) |
| `POSTGRES_URL` (or `DATABASE_URL`) | Postgres connection string |

**Optional env vars:**
| Variable | Purpose | Default |
|---|---|---|
| `PORT` | HTTP listen port | `3000` |
| `APP_ENV` | Environment name | `development` |
| `SESSION_TTL` | Session duration | `8h` |
| `PG_MAX_CONNS` | Postgres pool size | `10` |
| `ORACLE_URL` | Oracle direct DSN | — |
| `ORACLE_HOST` | Oracle host | `localhost` |
| `ORACLE_PORT` | Oracle port | `1521` |
| `ORACLE_SERVICE` | Oracle service | `xe` |
| `ORACLE_USER` | Oracle user | — |
| `ORACLE_PASSWORD` | Oracle password | — |
| `ORACLE_MAX_CONNS` | Oracle pool size | `5` |
| `ORACLE_IDLE_CONNS` | Oracle idle pool | `3` |
| `ORACLE_IDLE_TIME` | Oracle idle timeout | `5m` |

**Secrets location:**
- `.env` file in project root (gitignored per `.gitignore:1`)
- `.env.example` committed to repo as a template (no secrets)

## Webhooks & Callbacks

**Incoming:**
- None. No webhook endpoints.

**Outgoing:**
- None. No webhook dispatches.

## Frontend Assets

All frontend assets are embedded in the Go binary at build time via `go:embed` (`cmd/server/main.go:206`). Assets served by the application server itself on dedicated routes defined in `cmd/server/main.go:86-93`. No CDN, no external asset host.

- `htmx.min.js` — Embedded HTMX library, served at `GET /htmx.min.js`
- CSS: `style.css`, `admin.css` — Served at `GET /style.css`, `GET /admin.css`
- JS: `shared.js`, `app.js`, `dashboard.js`, `admin.js`, `login.js`

---

*Integration audit: 2026-06-08*
