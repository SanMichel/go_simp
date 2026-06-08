# Coding Conventions

**Analysis Date:** 2026-06-08

## Project Structure Convention

**Single `main` package architecture.** All Go source files in `cmd/server/` share `package main`. No sub-packages, no `internal/` module, no separate library packages.

**Files:**
| File | Purpose | Key Types |
|------|---------|-----------|
| `cmd/server/main.go` | Entrypoint, routes, HTTP server setup, template parsing | — |
| `cmd/server/models.go` | Core data structures, `App`/`Config` structs | `App`, `Config`, `User`, `Activity`, `ProductVerification`, Oracle types |
| `cmd/server/handlers.go` | Page + API HTTP handlers | — |
| `cmd/server/api_handlers.go` | API-only handlers (admin CRUD, dashboard API) | `APIActivity`, `APIProductVerification`, `OracleProductResponse` |
| `cmd/server/auth.go` | Session tokens, CSRF, role middleware | — |
| `cmd/server/db.go` | DB connectivity, migrations, queries | `OracleReader` |
| `cmd/server/utils.go` | Config loading, logging middleware, rate limiter | `rateLimiter` |
| `cmd/server/main_test.go` | All tests (single file) | — |

## Naming Conventions

**Files:** `snake_case.go` — descriptive single-word or compound names (`handlers.go`, `api_handlers.go`, `db.go`, `auth.go`, `utils.go`, `models.go`).

**Types:** PascalCase exported types:
- `Config`, `User`, `Activity`, `App`, `OracleReader`, `ProductVerification`, `FilterOptions`, `ActivityFilters`
- Oracle DB response types: `OracleEmpresa`, `OracleLocal`, `OracleProduct`
- API response types: `APIActivity`, `APIProductVerification`, `APIUser`, `OracleProductResponse`

**Functions/Methods:**
- Receiver methods on `*App`: `camelCase` — `app.home`, `app.loginPost`, `app.apiFinalizar`, `app.listActivities`
- Standalone functions: `camelCase` — `parseTemplates()`, `loadConfig()`, `loadDotEnv()`, `writeJSON()`, `parseFilters()`, `validRole()`, `firstNonEmpty()`, `isReadOnlySQL()`, `removeSQLComments()`
- Tests: `camelCase` — `TestTemplatesParse`, `TestLoadDotEnv`, `TestOracleReadOnlySQLGuard`, `TestMakeToken`

**Variables:** Short, idiomatic Go:
- Context: `ctx`
- Request/Response: `w`, `r`
- Database: `pg`, `ora`
- Config: `cfg`
- User: `u`
- Query arguments: `args`
- Loop iteration: `i`, `v`, `x`
- Nullable DB fields: `desc`, `mdv`, `ddv`, `rua`, `predio`, `expRua`, `expPredio`, `data`

**Private fields:** lowercase (Go convention) — `a.cfg`, `a.pg`, `a.ora`, `a.tpl`, `a.loginLimiter`

**JSON tags:** snake_case — `json:"nroempresa"`, `json:"desccompleta"`, `json:"dataFim"`, `json:"codacesso"`, `json:"precoVenda"`, `json:"diasEstoque"`

**Oracle column tags:** UPPER_SNAKE_CASE for some Oracle-mapped types — `json:"NROEMPRESA"`, `json:"NOMEREDUZIDO"`, `json:"SEQLOCAL"`, `json:"LOCAL"`. This is inconsistent with the snake_case tags used elsewhere.

**Template name references:** lowercase single-word (`"login"`, `"home"`, `"dashboard"`, `"atividades"`, `"admin"`, `"print"`).

## Code Style

**Go version:** 1.23 (`go.mod` line 3)

**Formatting:** Standard `gofmt` (enforced by Go toolchain). No `.editorconfig` or `.golangci.yml` detected.

**Linting:** No linter config file detected. No ESLint, Prettier, or biome configs in repo.

**Import organization:** Standard Go convention — standard library first, then third-party packages, blank imports for drivers:
```go
import (
    "context"
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    // ... stdlib

    _ "github.com/jackc/pgx/v5/stdlib"
    _ "github.com/sijms/go-ora/v2"
    "golang.org/x/crypto/bcrypt"
)
```

**Line length:** No enforced limit. Some handler functions are long (e.g., `apiFinalizar` in `handlers.go:434-535` is ~100 lines).

**Blank lines:** Used sparingly between function declarations. No blank line between closely related functions (e.g., `validRole()` and `firstNonEmpty()` in `utils.go`).

**Error handling:**
- `log.Fatal` on startup failures (`utils.go:21`, `main.go:29`, `main.go:37`)
- `log.Printf` for non-fatal warnings/errors: `log.Printf("error: %v", err)`, `log.Printf("warn: ...")`
- `http.Error(w, msg, statusCode)` for page handler errors
- `writeJSON(w, status, map[string]string{"error": msg})` for API handler errors
- Errors silently discarded with `_ =` when intentionally ignored (e.g., `_, _ = w.Write(b)` in `main.go:150`, `defer tx.Rollback()` in `handlers.go:453`)
- Oracle query errors in `activityDetailsData` are caught, logged, and return partial results rather than failing (`db.go:332-333`)

**Logging framework:** Go standard `log` package. No structured logging, no zerolog/zap/logrus.

**Logging patterns:**
- `log.Printf("error: %v", err)` — handler-level errors
- `log.Printf("warn: %s", msg)` — non-fatal issues (Oracle ping, failed impresso update)
- `log.Printf("server ready on http://localhost%s", srv.Addr)` — startup info
- `log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start))` — request logging via middleware (`utils.go:135`)
- Portuguese messages for user-facing errors, English for internal logs

## HTML Template Conventions

**Engine:** Go `html/template` (not `text/template`) via `template.ParseFS` (`main.go:203`).

**Template files:** Located in `cmd/server/templates/` and `cmd/server/templates/components/`, embedded via `//go:embed templates/*.html templates/components/*.html templates/*.css templates/*.js` (`main.go:206`).

**Template definitions:** Each file uses `{{define "name"}}...{{end}}` to name the template. Names match the route's rendered view:
- `templates/login.html` → `{{define "login"}}`
- `templates/dashboard.html` → `{{define "dashboard"}}`
- `templates/atividades.html` → `{{define "atividades"}}`
- `templates/admin.html` → `{{define "admin"}}`
- `templates/home.html` → `{{define "home"}}`
- `templates/print.html` → `{{define "print"}}`
- `templates/components/head.html` → `{{define "head"}}`
- `templates/components/nav.html` → `{{define "nav"}}`
- `templates/components/activities_table.html` → `{{define "activities_table"}}`
- `templates/components/activity_modal.html` → `{{define "activity_modal"}}`
- `templates/components/users_section.html` → `{{define "users_section"}}`
- `templates/components/user_row.html` → `{{define "user_row"}}`
- `templates/components/user_edit_row.html` → `{{define "user_edit_row"}}`

**Template functions:** Registered via `template.FuncMap` in `parseTemplates()` (`main.go:176-202`):
- `"rowUser"` — wraps `UserRow` for nested template context
- `"date"` — formats `time.Time` as `02/01/2006 15:04` or `"-"` if zero
- `"rolePt"` — translates role enum to Portuguese label
- `"checked"` — bool to `"S"`/`"N"`

**Sub-template inclusion:**
- `{{template "head" .}}` — included in `home.html`
- `{{template "user_row" (rowUser .)}}` — included in `users_section.html`

**Compiled JS/CSS:** CSS and JS files are embedded and served at runtime via handlers (`main.go:142-173`). CSS is in separate `.css` files (not const blocks as the old README suggested).

**HTMX usage:** HTMX is served from `templates/htmx.min.js` (embedded). Components use `hx-get`, `hx-post`, `hx-target`, `hx-on:click` attributes (`activities_table.html`).

**Language:** All templates use `lang="pt-BR"`.

## Route and Handler Patterns

**Router:** Go 1.22+ `http.ServeMux` with method-based routing and path parameters:
```go
mux := http.NewServeMux()
mux.HandleFunc("GET /api/health", a.healthCheck)
mux.HandleFunc("GET /dashboard/activities/{id}/details", a.requireRole("gerente,sysadmin", a.activityDetails))
```

**Route groups:**
- Page routes: `/home`, `/login`, `/atividades`, `/dashboard`, `/admin`
- API routes (prefix `/api/`): `/api/auth/*`, `/api/empresas`, `/api/locais`, `/api/produtos/*`, `/api/atividades/*`, `/api/admin/*`, `/api/dashboard/*`
- Static routes: `/style.css`, `/admin.css`, `/shared.js`, `/htmx.min.js`, `/app.js`, `/dashboard.js`, `/admin.js`, `/login.js`

**Middleware stack pattern** (`main.go:66`):
```go
Handler: app.csrfMiddleware(app.securityHeaders(app.log(mux)))
```
Middleware is stacked by wrapping: innermost wraps the mux, outermost is the server handler.

**Auth middleware:** Two variants:
- `requireRole(roles, handler)` — for pages, redirects on failure (`auth.go:86-110`)
- `requireAPIRole(roles, handler)` — for API, returns JSON error on failure (`auth.go:112-131`)

**Handler user injection:** Authenticated user stored in request context via `context.WithValue(r.Context(), ctxUser, u)` (`auth.go:107`). Retrieved with `r.Context().Value(ctxUser).(*User)` in handlers.

**API handler signature:** `requireAPIRole` wraps handlers with signature `func(w http.ResponseWriter, *http.Request, *User)`.

**Redirection pattern:** `a.redirectByRole(w, r, role)` routes authenticated users based on role: `sysadmin` → `/admin`, `gerente` → `/dashboard`, default → `/atividades` (`auth.go:132-141`).

**Render pattern:**
```go
a.render(w, "templateName", map[string]any{"Key": value})
```
All page handlers use this pattern. `render()` sets Content-Type, executes template, logs error on failure, returns Portuguese "Erro interno do servidor" (`main.go:135-141`).

**JSON response pattern:**
```go
writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
```
All API handlers use `writeJSON()` which sets Content-Type, writes status, encodes JSON (`utils.go:177-181`).

## Database Query Patterns

**Postgres (app database):** `*sql.DB` via `pgx` driver, uses `$1`, `$2` parameter placeholders:
```go
a.pg.QueryRowContext(ctx, `SELECT id FROM users WHERE username=$1`, username).Scan(&id)
a.pg.ExecContext(ctx, `INSERT INTO users (username, password, role) VALUES ($1,$2,$3)`, username, string(hash), role)
```

**Oracle (read-only source):** Custom `OracleReader` wrapper around `*sql.DB` via `go-ora` driver, uses `:1`, `:2` bind variables:
```go
a.ora.QueryContext(ctx, `SELECT ml.SEQLOCAL FROM CONSINCO.MRL_LOCAL ml WHERE ml.STATUS = 'A' AND ml.NROEMPRESA = :1`, empresa)
```

**Oracle guard:** All Oracle queries pass through `isReadOnlySQL()` which rejects anything that isn't `SELECT` or `WITH` (`db.go:15-27`). DML keywords (INSERT, UPDATE, DELETE, DROP, etc.) are blocked.

**Transaction pattern:**
```go
tx, err := a.pg.BeginTx(r.Context(), nil)
defer tx.Rollback()
// ... tx.ExecContext / tx.QueryRowContext ...
tx.Commit()
```
(`handlers.go:448-533`, `apiFinalizar`)

**Dynamic query building:** `listActivities` uses string concatenation with `fmt.Sprintf` for WHERE clauses and ORDER BY:
```go
where = append(where, fmt.Sprintf("a.impresso=$%d", len(args)))
```
(`db.go:210-298`)

**SQL comment removal:** `removeSQLComments()` strips `--` inline and `/* */` block comments from Oracle queries before checking for DML keywords (`db.go:67-109`).

**NULL handling:** `sql.NullString`, `sql.NullTime`, `sql.NullFloat64` used extensively in models and mapped to `*string`, `*time.Time`, `*float64` in API responses via mapping functions (`api_handlers.go`).

## Configuration Management

**Env-based config:** `loadConfig()` reads from environment variables with defaults (`utils.go:17-87`). Falls back to `.env` file via `loadDotEnv()`.

**Required vars:** `SESSION_SECRET` (min 32 chars), `POSTGRES_URL` (or `DATABASE_URL`).

**Optional vars with defaults:** `PORT` (3000), `APP_ENV` (development), `SESSION_TTL` (8h), `PG_MAX_CONNS` (10), `ORACLE_MAX_CONNS` (5), `ORACLE_IDLE_CONNS` (3), `ORACLE_IDLE_TIME` (5m).

**Oracle connection:** Built via `go_ora.BuildUrl()` from individual env vars, or overridden by `ORACLE_URL` (`utils.go:31-42`).

**Dotenv loader:** Custom `loadDotEnv()` function reads `.env` files manually — no godotenv dependency. Supports quoted values, comments, CRLF, and skips existing env vars (`utils.go:96-129`).

## Error Handling

**Patterns used:**
1. **Return error, check upstream** — standard Go: `err := doSomething(); if err != nil { ... }`
2. **log.Fatal on startup** — irrecoverable initialization failures
3. **log.Printf + graceful fallback** — non-critical failures (Oracle ping, optional queries)
4. **Silent discard with `_ =`** — deferred operations, writes that can't fail usefully

**User-facing errors in Portuguese:**
- "Erro interno do servidor" — generic 500
- "Usuário ou senha incorretos." — login failure
- "Muitas tentativas. Aguarde 1 minuto." — rate limit
- "Dados inválidos." — validation
- "Não é possível editar o próprio usuário" — self-edit prevention
- "CSRF token ausente" / "CSRF token inválido" — CSRF failures
- "JSON inválido" — bad JSON payload

## Rate Limiting

**Simple in-memory rate limiter** (`utils.go:183-225`):
- Tracks attempts per IP address
- Allows 5 requests per minute per IP
- Background goroutine purges expired entries every minute
- Used for login endpoints (`/login` and `/api/auth/login`)

## Session & Auth

**Custom token format** (not JWT, not OAuth): `base64(payload).base64(hmac-sha256)` where payload is JSON with `id`, `exp`, `iat` fields (`auth.go:62-77`).

**Session revocation:** `last_token_at` column on `users` table. If token's `iat` < `last_token_at`, the session is considered revoked (`auth.go:56-58`).

**CSRF protection:** Cookie-based: `csrf_token` cookie set on login, `X-CSRF-Token` header required for POST/PATCH/DELETE to `/api/*` endpoints. Login page (`/login`) is excluded (`auth.go:156-184`).

**Security headers middleware:** CSP, X-Content-Type-Options, X-Frame-Options, Referrer-Policy, HSTS (production only) (`utils.go:227-238`).

---

*Convention analysis: 2026-06-08*
