<!-- refreshed: 2026-06-08 -->
# Architecture

**Analysis Date:** 2026-06-08

## System Overview

```text
┌───────────────────────────────────────────────────────────────────┐
│                       HTTP / Client Layer                          │
│  Browser (HTMX + JS SPA)  |  REST API clients (cURL, mobile)      │
│  `cmd/server/templates/*.html` | `cmd/server/templates/*.js`       │
└──────────────────────────────┬────────────────────────────────────┘
                               │
                               ▼
┌───────────────────────────────────────────────────────────────────┐
│                      Middleware Chain                              │
│  ┌────────┐  ┌──────────┐  ┌───────────────┐  ┌──────────────┐  │
│  │  log   │→ │ security │→ │ csrfMiddleware│→ │  ServeMux    │  │
│  │(utils) │  │ (Headers)│  │  (auth.ts)    │  │  (main.go)   │  │
│  └────────┘  └──────────┘  └───────────────┘  └──────┬───────┘  │
│                                                       │           │
└───────────────────────────────────────────────────────┼───────────┘
                                                         │
                                                         ▼
┌───────────────────────────────────────────────────────────────────┐
│                    Handler Layer (HTTP)                            │
│  ┌─────────────────────┐  ┌───────────────────────────────────┐   │
│  │  Page Handlers       │  │  API Handlers                     │   │
│  │  `handlers.go`       │  │  `api_handlers.go` + `handlers.go`│   │
│  │  /login, /dashboard, │  │  /api/auth/*, /api/empresas,      │   │
│  │  /atividades, /admin  │  │  /api/produtos/*, /api/atividades │   │
│  └──────────┬───────────┘  └──────────────┬────────────────────┘   │
│             │                             │                        │
└─────────────┼─────────────────────────────┼────────────────────────┘
               │                             │
               ▼                             ▼
┌───────────────────────────────────────────────────────────────────┐
│                  Application / Business Layer                      │
│  ┌─────────────────────┐  ┌───────────────────────────────────┐   │
│  │  Auth & Session     │  │  Data Access / Queries            │   │
│  │  `auth.go`          │  │  `db.go`                          │   │
│  │  token management,  │  │  Postgres + Oracle read operations │   │
│  │  RBAC, CSRF          │  │  migrations, seeders              │   │
│  └──────────┬───────────┘  └──────────────┬────────────────────┘   │
└─────────────┼─────────────────────────────┼────────────────────────┘
               │                             │
               ▼                             ▼
┌───────────────────────────────────────────────────────────────────┐
│                      Data Store Layer                              │
│  ┌──────────────────────┐  ┌────────────────────────────────┐     │
│  │  PostgreSQL (app)    │  │  Oracle (read-only, source)    │     │
│  │  users, atividades,  │  │  empresas, locais, produtos    │     │
│  │  produto_verificacao  │  │  CONSINCO schema              │     │
│  │  database/sql + pgx  │  │  database/sql + go-ora        │     │
│  └──────────────────────┘  └────────────────────────────────┘     │
└───────────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| `App` struct | Application container (config, DB conns, templates, rate limiter) | `cmd/server/models.go:22` |
| `Config` struct | All environment-derived configuration | `cmd/server/models.go:9` |
| `OracleReader` | Read-only Oracle proxy with SQL guard | `cmd/server/models.go:30` |
| Route registration | Maps method+path patterns to handler funcs | `cmd/server/main.go:85` |
| Page handlers | Render HTML templates for browser | `cmd/server/handlers.go` |
| API handlers | Return JSON responses for SPA/API clients | `cmd/server/api_handlers.go` |
| Auth middleware | Session token verification + role checking | `cmd/server/auth.go` |
| CSRF middleware | Origin check + token validation for mutating requests | `cmd/server/auth.go:156` |
| Rate limiter | Login brute-force protection | `cmd/server/utils.go:188` |
| DB migrations | Auto-create/alter Postgres tables on startup | `cmd/server/db.go:111` |
| Oracle queries | Read-only product/location lookups with SQL injection guard | `cmd/server/db.go:15` |
| Template engine | Go `html/template` with custom funcs | `cmd/server/main.go:175` |
| Middleware chain | log → securityHeaders → csrfMiddleware → mux | `cmd/server/main.go:66` |

## Pattern Overview

**Overall:** Flat package MVC variant — single `package main` with files organized by concern (models, handlers, auth, db). No framework, no dependency injection container. The `App` struct serves as the manual DI container holding all dependencies.

**Key Characteristics:**
- Single `package main` — all types and functions live in one flat namespace
- `App` struct as dependency container injected into every handler via method receivers
- HTMX for progressive enhancement on server-rendered pages; separate SPA-style JS apps for `/atividades` and `/dashboard` using REST API
- Custom session token (HMAC-signed JSON) with no third-party session library
- Two data sources: Postgres (app state, write path) and Oracle (read-only product data)
- Go 1.22+ routing with method+pattern `mux.HandleFunc("GET /path/{id}", handler)` — no router library

## Layers

**Template / Frontend Layer:**
- Purpose: HTML rendering and client-side interactivity
- Location: `cmd/server/templates/`
- Contains: `.html` Go templates, `.css` stylesheets, `.js` client scripts, vendored `htmx.min.js`
- Rendered by: `cmd/server/main.go:135` (`App.render` method)
- Client-side JS: Two separate SPA apps — `/atividades` uses `app.js`, `/dashboard` uses `dashboard.js`, `/admin` uses `admin.js`
- Pure server-rendered pages: `/login` uses `login.html` + `login.js`

**Handler Layer:**
- Purpose: HTTP request/response handling
- Location: `cmd/server/handlers.go` + `cmd/server/api_handlers.go`
- Contains: Page-rendering functions (return HTML), API JSON endpoints
- Depends on: `App` struct (for DB, templates, auth, config)
- Used by: `cmd/server/main.go` routes

**Auth Layer:**
- Purpose: Session management, RBAC, CSRF protection
- Location: `cmd/server/auth.go`
- Contains: `currentUser()`, `makeToken()`, `requireRole()`, `requireAPIRole()`, `csrfMiddleware()`
- Depends on: `App` struct (for DB queries, config)
- Used by: Route registration wraps most handlers with `requireRole()` or `requireAPIRole()`

**Data Access Layer:**
- Purpose: Database queries, migrations, seeding
- Location: `cmd/server/db.go`
- Contains: Postgres CRUD, Oracle read-only queries, table migrations, admin seed
- Depends on: `App.pg` (Postgres `*sql.DB`), `App.ora` (`*OracleReader`)
- Key detail: `OracleReader.QueryContext()` guards against non-SELECT/WITH queries (`cmd/server/db.go:15`)

**Configuration Layer:**
- Purpose: Environment loading, config struct building
- Location: `cmd/server/utils.go`
- Contains: `loadConfig()`, `loadDotEnv()`, rate limiter, logging middleware, security headers, JSON helpers

**Models Layer:**
- Purpose: Data structures shared across layers
- Location: `cmd/server/models.go`
- Contains: `Config`, `App`, `OracleReader`, `User`, `UserRow`, `Activity`, `ProductVerification`, `OracleEmpresa`, `OracleLocal`, `OracleProduct`, `finalizeReq`
- Also contains API-specific response types in: `cmd/server/api_handlers.go` (`APIActivity`, `APIProductVerification`, `APIUser`, `OracleProductResponse`)

## Data Flow

### Primary Request Path (Server-Rendered Page)

1. Browser sends request → `http.Server` (`main.go:64`)
2. Middleware chain processes: `mux` → `log` middleware → `securityHeaders` → `csrfMiddleware` → `http.ServeMux` routing
3. For protected routes, `requireRole()` middleware (`auth.go:86`) reads session cookie → validates HMAC token → loads user from DB → injects `*User` into request context
4. Route matches handler: e.g., `GET /dashboard` → `a.dashboardPage` (`handlers.go:76`)
5. Handler calls data layer: e.g., `a.listActivities(ctx, filters, limit)`, `a.listFilterOptions(ctx)`
6. Handler calls `a.render(w, "dashboard", data)` which executes Go template
7. Response returned through middleware chain

### SPA Data Flow (REST JSON API)

1. Client-side JS fetches `GET /api/dashboard/activities`
2. `requireAPIRole()` middleware (`auth.go:112`) validates token, checks role, injects user
3. API handler queries data, maps to API response types, calls `writeJSON()` (`utils.go:177`)
4. Client JS renders response into DOM (JSON-driven, not HTMX)

### HTMX Partial Rendering

1. HTMX triggers `hx-get="/dashboard/activities/{id}/details"` (`activities_table.html:1`)
2. Server handler `activityDetails()` (`handlers.go:94`) queries activity + products
3. Server renders `activity_modal.html` template fragment
4. HTMX swaps response into target DOM element

### Activity Finalization Flow

1. SPA client collects scan data, sends `POST /api/atividades/finalizar` with JSON body
2. `requireAPIRole` middleware (`auth.go:112`) authenticates
3. `apiFinalizar()` handler (`handlers.go:434`) begins Postgres transaction
4. Inserts into `atividades`, `atividade_enderecos`, `produto_verificacao` tables
5. Computes divergences, ruptures, replenishments from expected vs read products
6. Commits transaction
7. Returns JSON with activity ID, timestamps, and warning counts

**State Management:**
- Session state: HMAC-signed JSON token in `token` cookie (HttpOnly, Secure, SameSite=Strict)
- CSRF state: Random token in `csrf_token` cookie (SameSite=Lax), validated via `X-CSRF-Token` header for API calls
- No server-side session store — token contains user ID + expiry + issue time, with `last_token_at` revocation check
- Login rate limiting: In-memory `rateLimiter` struct with per-IP count and 1-minute reset window (`utils.go:188`)

## Key Abstractions

**`App` struct:**
- Purpose: Dependency container — holds config, DB connections, template engine, rate limiter
- Examples: `cmd/server/models.go:22`
- Pattern: Manual DI via struct field injection, all methods are `(a *App)` receivers

**`OracleReader` struct:**
- Purpose: Wraps `*sql.DB` for Oracle with read-only enforcement
- Examples: `cmd/server/models.go:30`, `cmd/server/db.go:15`
- Pattern: Decorator — intercepts queries, validates read-only before delegating

**HMAC Session Token:**
- Purpose: Custom stateless auth token (not JWT — no nonce, no standard claims)
- Examples: `cmd/server/auth.go:62`
- Pattern: `base64(json_payload) + "." + base64(HMAC-SHA256(payload, secret))`
- Contains: `{id, exp, iat}` — user ID, expiry, issued-at (for revocation)

**`requireRole` / `requireAPIRole`:**
- Purpose: Middleware factories for RBAC
- Examples: `cmd/server/auth.go:86`, `cmd/server/auth.go:112`
- Pattern: Closure — parses comma-separated allowed roles, returns `http.HandlerFunc` that validates session + role before passing to next handler

**Template rendering:**
- Purpose: Go `html/template` with `go:embed` and custom funcs
- Examples: `cmd/server/main.go:175`
- Pattern: Single `template.Template` compiled at startup from embedded FS, reused for all responses

## Entry Points

**`main()` — Application start:**
- Location: `cmd/server/main.go:20`
- Triggers: Process start
- Flow:
  1. Load `.env` file via `loadDotEnv()` (`utils.go:96`)
  2. Parse config via `loadConfig()` (`utils.go:17`)
  3. Open Postgres connection (`pgx` driver)
  4. Open Oracle connection (`go-ora` driver) with warning on failure
  5. Construct `App` struct with all dependencies
  6. Run auto-migrations (`app.migrate()`) — idempotent CREATE TABLE IF NOT EXISTS
  7. Seed admin user (`app.seedAdmin()`) — creates `admin` with random password on first run
  8. Register all routes via `app.routes(mux)` (`main.go:85`)
  9. Wrap mux in middleware chain: `app.csrfMiddleware(app.securityHeaders(app.log(mux)))`
  10. Start HTTP server with graceful shutdown on SIGINT/SIGTERM

**Routes — All registered in a single method:**
- Location: `cmd/server/main.go:85`
- Categories:
  - Static assets: CSS, JS, HTMX library (`style`, `adminStyle`, `serveJS`)
  - Page routes (server-rendered): `/`, `/login`, `/home`, `/atividades`, `/dashboard`, `/admin`
  - HTMX partials: `/dashboard/table`, `/dashboard/activities/{id}/details`, `/admin/users/section`, `/admin/users/{id}/edit`, `/admin/users/{id}/row`
  - API endpoints: `/api/auth/*`, `/api/empresas`, `/api/locais`, `/api/produtos/*`, `/api/atividades/*`, `/api/admin/users/*`, `/api/dashboard/*`
  - Health: `GET /api/health`

## Route Structure and Middleware Chain

```
Mux (http.ServeMux)
  │
  ├── Outer wrapper: a.log(mux)                   — logs method, path, status, duration
  │
  ├── Mid wrapper: a.securityHeaders(next)         — sets CSP, X-Content-Type-Options, X-Frame-Options, Referrer-Policy, HSTS
  │
  └── Inner wrapper: a.csrfMiddleware(next)        — checks Origin on POST/PATCH/DELETE, validates X-CSRF-Token for /api/*
       │
       └── Route matching
            ├── Public: /login (GET), /api/health (GET), static files
            ├── requireRole(roles): page handlers
            │   ├── "" (any authenticated) — /atividades
            │   ├── "gerente,sysadmin"     — /dashboard/*
            │   └── "sysadmin"             — /admin/*
            │
            └── requireAPIRole(roles): API handlers (receives *User as 3rd param)
                ├── "conferente,gerente,sysadmin" — /api/empresas, /api/produtos/*, /api/atividades/*
                ├── "gerente,sysadmin"            — /api/dashboard/*
                └── "sysadmin"                    — /api/admin/users/*
```

## Architectural Constraints

- **Threading:** Single-threaded Go HTTP server; each request handled in its own goroutine. Mutex-based rate limiter (`sync.Mutex`) for login attempts.
- **Global state:** No module-level singletons. All app state lives on the `App` struct. Rate limiter is the only stateful goroutine (background cleanup ticker in `newRateLimiter()`, `utils.go:193`).
- **Circular imports:** Not possible — single package prevents cycles.
- **Database connections:** Two separate `*sql.DB` pools — one for Postgres (write-capable), one for Oracle (read-only enforced).
- **Oracle read-only enforcement:** `isReadOnlySQL()` (`db.go:29`) strips comments and string literals, then checks for DML keywords. This is a defense-in-depth measure on top of using a read-only DB account.
- **No ORM:** Raw SQL queries throughout `db.go` with manual scanning.
- **Template embedding:** All templates compiled into binary via `go:embed` at `cmd/server/main.go:206`. No runtime template loading.
- **Error handling:** All handler errors are logged via `log.Printf` and return generic user-facing messages ("Erro interno do servidor"). No structured error types.
- **CSRF gap:** The `/login` and `/api/auth/login` endpoints are explicitly excluded from CSRF protection (`auth.go:159`).

## Anti-Patterns

### No structured error types

**What happens:** All errors are `error` interfaces or strings. Error propagation is manual (`if err != nil { log.Printf(...); http.Error(...); return }`) with no centralized error handling.
**Why it's wrong:** Mixed logging + HTTP response concerns in every handler. Error messages in Portuguese leak business logic but not stack traces.
**Do this instead:** Define HTTP error response helpers and use Go 1.24+ `errors.Join` or sentinel errors.

### Flat package, no internal modules

**What happens:** All code in `package main` with no `internal/` packages.
**Why it's wrong:** No explicit dependency boundaries. Any function can call any other function. Tests import the same flat package.
**Do this instead:** For future growth, split into `internal/auth`, `internal/db`, `internal/handler` packages.

### Mixed rendering strategies

**What happens:** `/dashboard` and `/admin` have both server-rendered HTML pages (initial load) AND JSON API endpoints (SPA interactions). Some pages use HTMX (`activities_table.html`), others use vanilla JS. The `/login` page is pure server-rendered with JS enhancement.
**Why it's wrong:** Multiple client rendering strategies increase maintenance burden. Template partials (`activities_table.html`) duplicate the table structure that `dashboard.js` also renders from JSON.
**Do this instead:** Choose one client rendering approach — either pure server-rendered with HTMX everywhere, or full SPA with one JS entry point.

## Error Handling

**Strategy:** Ad-hoc per-handler error handling. Errors are logged server-side, generic messages sent to client.

**Patterns:**
- Page handlers: `a.render()` wraps template execution errors with generic 500 (`main.go:135`)
- API handlers: `writeJSON()` with HTTP status code + simple `{"error": "..."}` body
- Data layer: Errors propagated upward via `error` return values; Oracle errors are logged as warnings (non-fatal)
- Panic handling: None explicit — Go http.Server has default panic recovery

## Cross-Cutting Concerns

**Logging:** Standard `log.Printf` throughout. No structured logging, no log levels beyond `log.Printf` / `log.Fatal`. Request logging via `log` middleware (`utils.go:130`) captures method, path, status, duration.

**Validation:** Minimal. `validRole()` checks allowed roles (`utils.go:148`). Password length check (≥8) in user creation. Input validation is performed client-side.

**Authentication:** Custom HMAC token in HttpOnly cookie. Session revocation via `last_token_at` timestamp in Postgres. Rate limiting on login endpoints (5 attempts/minute per IP).

---

*Architecture analysis: 2026-06-08*
