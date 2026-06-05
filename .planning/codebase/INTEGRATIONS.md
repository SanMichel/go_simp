# External Integrations

**Analysis Date:** 2026-06-05

## APIs & External Services

**Oracle Database (Product/Company/Location data):**
- Used for product lookups by EAN/code, company listings, location listings, and product details (stock, pricing, sales velocity).
- Read-only — enforced by `isReadOnlySQL()` guard in `cmd/server/db.go:28-64`. Only `SELECT`/`WITH` queries pass. DML statements are rejected.
- Driver: `github.com/sijms/go-ora/v2 v2.8.24` registered as `"oracle"` in `cmd/server/main.go:15`.
- Connection: `ORACLE_URL` env var, or built from `ORACLE_HOST/PORT/SERVICE/USER/PASSWORD`.
- Queries target `CONSINCO` schema tables: `MAX_EMPRESA`, `MRL_LOCAL`, `MRL_PRODUTOEMPRESA`, `MAP_PRODUTO`, `MAP_PRODCODIGO`, `MRL_PRODLOCAL`, `ETLV_PRODUTO`, and function `fBuscaPrecoAtualPdv`.

## Data Storage

**Databases:**

**PostgreSQL (Application state):**
- Stores: users, activities (atividades), activity addresses (atividade_enderecos), product verifications (produto_verificacao).
- Driver: `github.com/jackc/pgx/v5 v5.7.2` via `database/sql` stdlib interface, registered as `"pgx"`.
- Connection: `POSTGRES_URL` or `DATABASE_URL` env var.
- Connection pool: configurable via `PG_MAX_CONNS` (default 10), conn max lifetime 1 hour.
- Auto-migration on startup: `App.migrate()` in `cmd/server/db.go:110-158` creates tables and indexes idempotently.
- Schema (4 tables): `users`, `atividades`, `atividade_enderecos`, `produto_verificacao` — all with `BIGSERIAL` PKs, `TIMESTAMPTZ` for time fields.

**Oracle (Read-only reference):**
- Stores: enterprise configuration, product catalog, stock levels, pricing, locations.
- Read-only enforced programmatically.
- Connection pool: configurable via `ORACLE_MAX_CONNS` (default 5) and `ORACLE_IDLE_CONNS` (default 1), conn max lifetime 1 hour.

**File Storage:**
- Local filesystem only. No external file storage (S3, etc.).
- Static assets (CSS, JS, HTML templates) embedded in binary via `go:embed` in `cmd/server/main.go:185-186`.

**Caching:**
- None. No Redis, Memcached, or in-memory cache beyond application structs.

## Authentication & Identity

**Auth Provider:**
- **Custom implementation** — no third-party auth provider (no OAuth, no SSO, no social login).
- Session tokens: HMAC-SHA256 signed JSON payloads stored in `token` cookie (HttpOnly, SameSite=Strict, Secure).
- Token format: `base64(payload).base64(signature)` — custom, not JWT (no `nonce` field, simple `{id, exp, iat}`).
- Password hashing: bcrypt via `golang.org/x/crypto/bcrypt`.
- Session revocation: `last_token_at` column on `users` table invalidates tokens issued before that timestamp.
- Rate limiting: in-memory per-IP, 5 attempts/minute (`rateLimiter` in `cmd/server/utils.go:181-218`).
- CSRF protection: custom middleware using `csrf_token` cookie + `X-CSRF-Token` header for API routes (`cmd/server/auth.go:140-167`).
- Roles: `conferente` (base), `gerente` (manager), `sysadmin` (admin). Enforced by `requireRole` middleware.

## Monitoring & Observability

**Error Tracking:**
- None. Errors are logged via `log.Printf` to stdout only.

**Logs:**
- Standard library `log` package — structured-ish format: `"METHOD /path STATUS duration"` via `log` middleware in `cmd/server/utils.go:123-130`.
- Build error logs: `build-errors.log` (Air output, gitignored).

**Health Check:**
- `GET /api/health` — returns `{"status": "ok"}` or `{"status": "oracle_down"}` (if Oracle ping fails). Postgres not checked in health endpoint.

## CI/CD & Deployment

**Hosting:**
- Self-hosted Go binary. Single executable, no container configuration found.

**CI Pipeline:**
- Not detected. No CI config files (GitHub Actions, GitLab CI, CircleCI, etc.).

## Environment Configuration

**Required env vars (app will `log.Fatal` if missing):**
- `SESSION_SECRET` — HMAC key, minimum 32 characters
- `POSTGRES_URL` or `DATABASE_URL` — Postgres connection string

**Optional but functionally required:**
- Oracle connection params (`ORACLE_URL` or `ORACLE_HOST/PORT/SERVICE/USER/PASSWORD`) — without these, product lookups will fail. App logs a warning but continues.

**Secrets location:**
- `.env` file (gitignored, loaded at startup by `loadDotEnv()` in `cmd/server/utils.go:89-122`)
- Environment variables override `.env` values

## Webhooks & Callbacks

**Incoming:**
- None. No webhook endpoints.

**Outgoing:**
- None. The app does not call external APIs.

---

*Integration audit: 2026-06-05*
