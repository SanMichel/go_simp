<!-- refreshed: 2026-06-05 -->
# Architecture

**Analysis Date:** 2026-06-05

## System Overview

```text
┌──────────────────────────────────────────────────────────────────┐
│                       HTTP Layer                                 │
│  net/http.ServeMux routing + middleware stack                    │
│  `cmd/server/main.go`                                            │
│     ├── securityHeaders()                                        │
│     ├── csrfMiddleware()                                         │
│     └── log()                                                    │
├──────────────────┬──────────────────┬────────────────────────────┤
│   Page Handlers  │   API Handlers   │    Static Assets           │
│  `handlers.go`   │ `api_handlers.go`│    `main.go` (serveJS)     │
│                  │                  │                             │
│  /login          │  /api/auth/*     │  /style.css                │
│  /home           │  /api/empresas   │  /app.js                   │
│  /atividades     │  /api/produtos/* │  /htmx.min.js             │
│  /dashboard      │  /api/atividades/*│                            │
│  /admin          │  /api/dashboard/*│                            │
│  /dashboard/print│  /api/admin/*    │                            │
└────────┬─────────┴────────┬─────────┴──────────┬─────────────────┘
         │                  │                     │
         ▼                  ▼                     │
┌─────────────────────────────────────────────────┴──┐
│                 Application Layer                    │
│  `cmd/server/auth.go`  — auth, sessions, RBAC       │
│  `cmd/server/db.go`    — queries, migrations        │
│  `cmd/server/utils.go` — config, logging, helpers   │
│  `cmd/server/models.go`— struct definitions         │
└─────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────┐
│  Data Stores                                         │
│                                                      │
│  PostgreSQL (pgx)       Oracle (go-ora, read-only)   │
│  ─────────────────      ─────────────────────        │
│  users                  MAX_EMPRESA (companies)      │
│  atividades             MRL_LOCAL (locations)        │
│  atividade_enderecos    MAP_PRODUTO (products)       │
│  produto_verificacao    MRL_PRODUTOEMPRESA           │
│                         MRL_PRODLOCAL                │
│                         MAP_PRODCODIGO               │
└─────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| App struct | Shared state container (cfg, pg, ora, tpl, loginLimiter) | `cmd/server/models.go:21` |
| Config struct | All environment-derived config | `cmd/server/models.go:9` |
| OracleReader | Wraps Oracle DB with read-only guard | `cmd/server/models.go:29`, `cmd/server/db.go:14` |
| routes() | Registers all URL patterns and middleware chains | `cmd/server/main.go:65` |
| render() | Renders Go templates to HTTP response | `cmd/server/main.go:114` |
| Page handlers | Handle HTML page requests (login, dashboard, admin) | `cmd/server/handlers.go` |
| API handlers | Handle JSON API requests with authenticated user | `cmd/server/api_handlers.go` |
| currentUser() | Extracts and validates session cookie | `cmd/server/auth.go:17` |
| requireRole() | Middleware factory for role-gated HTML routes | `cmd/server/auth.go:82` |
| requireAPI() | Middleware for authenticated API routes | `cmd/server/auth.go:107` |
| csrfMiddleware() | CSRF protection for POST/PATCH/DELETE | `cmd/server/auth.go:140` |
| migrate() | Auto-migrates Postgres schema on startup | `cmd/server/db.go:110` |
| seedAdmin() | Seeds default admin user on startup | `cmd/server/db.go:160` |
| loadConfig() | Reads env vars into Config struct | `cmd/server/utils.go:17` |
| loadDotEnv() | Custom .env file parser (no godotenv) | `cmd/server/utils.go:89` |
| Templates | Embedded HTML templates + JS + CSS via go:embed | `cmd/server/main.go:185` |

## Pattern Overview

**Overall:** Monolithic single-package Go server with flat file organization.

**Key Characteristics:**
- Single `package main` — all files share the same namespace; no internal packages or modules
- Method receivers on `*App` struct — all handlers and data layer methods hang off a central struct that holds dependencies (DI-light via manual wire-up in `main()`)
- Middleware chain wrapping pattern — `securityHeaders()` → `csrfMiddleware()` → `log()` → mux
- Handler signature polymorphism — HTML handlers use `http.HandlerFunc`, API handlers use custom `func(http.ResponseWriter, *http.Request, *User)` wrapped by `requireAPI()`
- Embedded static assets — all templates, CSS, JS are embedded into the binary via `//go:embed`

## Layers

**HTTP Routing Layer:**
- Purpose: Maps URL patterns to handlers, applies middleware
- Location: `cmd/server/main.go` (routes method, lines 65-113)
- Contains: Route definitions using Go 1.22 `mux.HandleFunc("METHOD /path", handler)` syntax
- Depends on: All handler functions defined in `handlers.go`, `api_handlers.go`
- Used by: `main()` at startup

**Middleware Stack:**
- Purpose: Wraps every request with logging, security headers, CSRF protection, rate limiting
- Location: `cmd/server/main.go:63` and `cmd/server/auth.go:140`, `cmd/server/utils.go:123,220`
- Order: `securityHeaders` → `csrfMiddleware` → `log` → handler
- CSRF skips: `/login` and `/api/auth/login` endpoints
- Rate limiting: Applies only to login endpoints (5 attempts/min/IP via `rateLimiter`)

**Handler Layer (Page Handlers):**
- Purpose: Serve HTML pages rendered from Go templates
- Location: `cmd/server/handlers.go`
- Contains: Login flow, dashboard, activities, admin CRUD, printing
- Depends on: `App` methods for DB queries, auth, template rendering
- Used by: Routes defined in `main.go`

**Handler Layer (API Handlers):**
- Purpose: Serve JSON API responses for client-side JS (HTMX)
- Location: `cmd/server/api_handlers.go`
- Contains: Product lookup, activity finalization, dashboard data, admin CRUD (JSON)
- Signature: All take `(w, r, u *User)` — authenticated user injected by `requireAPI`
- Used by: Routes defined in `main.go`

**Data Access Layer:**
- Purpose: Database queries, schema migrations, seeding
- Location: `cmd/server/db.go`
- Contains: Postgres queries (migrate, CRUD for users/activities/products), Oracle queries via `OracleReader`
- Dependencies: `database/sql` with pgx driver for Postgres, go-ora driver for Oracle
- Key constraint: Oracle is read-only — `isReadOnlySQL()` guard rejects non-SELECT/WITH queries

**Models Layer:**
- Purpose: Struct definitions for domain objects and configuration
- Location: `cmd/server/models.go`, plus API-specific DTOs in `cmd/server/api_handlers.go`
- Contains: `App`, `Config`, `User`, `Activity`, `ProductVerification`, `OracleProduct`, `finalizeReq`, API DTOs (APIActivity, APIProductVerification, APIUser)
- Maps: `mapActivity()`, `mapProduct()`, `mapUser()` convert internal models to API-safe DTOs with nullable fields

**Utilities Layer:**
- Purpose: Configuration loading, logging middleware, helpers, rate limiter
- Location: `cmd/server/utils.go`
- Contains: `loadConfig()`, `loadDotEnv()`, `log` middleware, `rateLimiter`, `securityHeaders`, `writeJSON`, `parseFilters`

**Authentication Layer:**
- Purpose: Session management with HMAC-signed tokens
- Location: `cmd/server/auth.go`
- Contains: `currentUser()`, `makeToken()`, `requireRole()`, `requireAPI()`, `csrfMiddleware()`, `revokeSession()`
- Token format: `base64(payload).base64(HMAC-SHA256(payload))` — custom JWT-like format, no external JWT library
- Session revocation: Updating `last_token_at` on user row invalidates tokens with `iat` before that time

## Data Flow

### Primary Request Path (Page Request)

1. HTTP request enters `http.ListenAndServe` → `securityHeaders()` middleware sets CSP, HSTS, X-Frame-Options (`cmd/server/utils.go:220`)
2. → `csrfMiddleware()` checks CSRF on POST/PATCH/DELETE (skips for `/login` and `/api/auth/login`) (`cmd/server/auth.go:140`)
3. → `log()` middleware records method, path, status, duration (`cmd/server/utils.go:123`)
4. → `ServeMux` matches route → calls handler
5. Route handlers protected by `requireRole()` first validate session cookie → call `currentUser()` → check role (`cmd/server/auth.go:82`)
6. Handler calls `App` methods (e.g., `listActivities()`, `findUserByID()`) → executes SQL → renders template
7. `render()` executes template from `templatesFS` → writes HTML response (`cmd/server/main.go:114`)

### API Request Path

1. Same middleware stack (securityHeaders → csrf → log)
2. Route calls `requireAPI()` → validates session cookie via `currentUser()` → injects `*User` into handler (`cmd/server/auth.go:107`)
3. API handler reads JSON body, calls DB methods, returns `writeJSON()` response (`cmd/server/api_handlers.go`)

### Login Flow

1. `GET /login` → `loginPage()` → checks existing session, renders login template (`cmd/server/handlers.go:20`)
2. `POST /login` → `loginPost()` → rate limiter check → `findUserByUsername()` → bcrypt compare → `makeToken()` → set cookie + CSRF cookie → redirect by role (`cmd/server/handlers.go:28`)
3. `POST /logout` → `logout()` → `revokeSession()` → clear cookies → redirect (`cmd/server/handlers.go:62`)

### Activity Finalization Flow

1. HTMX client POSTs JSON to `/api/atividades/finalizar` with products, addresses, company data (`cmd/server/handlers.go:383`)
2. Server begins Postgres transaction → inserts `atividades` row → inserts `atividade_enderecos` rows → inserts `produto_verificacao` rows for each expected product (status defaults to "RUPTURA" if not read) → commits
3. Returns JSON with activity ID and timestamp

### Oracle Lookup Flow

1. JavaScript/HTMX calls `/api/produtos/ean/{codigo}` or `/api/produtos/consulta/{codigo}` (`cmd/server/handlers.go:316`)
2. Handler calls Oracle via `OracleReader.QueryContext()` → SQL is validated by `isReadOnlySQL()` before execution (`cmd/server/db.go:14`)
3. Result mapped to `OracleProduct` struct → returned as JSON

**State Management:**
- Session state: HMAC-signed cookie (`token`) with embedded expiry + session revocation timestamp
- DB state: Postgres for app data, Oracle for product/company/location reference data (read-only)
- In-memory state: `rateLimiter` holds login attempt counts per IP (auto-cleanup goroutine every minute)

## Key Abstractions

**`App` struct:**
- Purpose: Dependency container — all handler methods, DB access, config, and template engine hang off this struct
- Examples: `a.findUserByUsername()`, `a.listActivities()`, `a.render()`, `a.currentUser()`
- Location: `cmd/server/models.go:21`
- Pattern: Single shared instance created in `main()` — manual dependency injection via struct fields

**`OracleReader` struct:**
- Purpose: Wraps `*sql.DB` for Oracle with read-only enforcement
- Location: `cmd/server/models.go:29`, `cmd/server/db.go:14`
- Pattern: Decorator — overrides `QueryContext` and `QueryRowContext` to validate SQL before delegation

**`rateLimiter`:**
- Purpose: Per-IP rate limiting for login endpoints
- Location: `cmd/server/utils.go:181`
- Pattern: Mutex-protected map with background goroutine for entry cleanup

**Middleware:**
- Purpose: Cross-cutting concerns (logging, security, CSRF, role gating)
- Location: `cmd/server/auth.go`, `cmd/server/utils.go`
- Pattern: Higher-order functions returning `http.Handler` or `http.HandlerFunc`

**API DTO mapping:**
- Purpose: Separate internal models from JSON-serializable API types (null-handling)
- Location: `cmd/server/api_handlers.go:50-101`
- Pattern: `map*` functions — `mapActivity()`, `mapProduct()`, `mapUser()` convert DB models to `API*` types

## Entry Points

**`main()`:**
- Location: `cmd/server/main.go:18`
- Triggers: `go run ./cmd/server` or compiled binary
- Responsibilities: Load config → connect Postgres → connect Oracle → wire App → auto-migrate → seed admin → setup routes → start HTTP server

## Architectural Constraints

- **Threading:** Single-threaded Go HTTP server. Postgres connection pool (`SetMaxOpenConns`), Oracle connection pool (`SetMaxOpenConns`). Rate limiter uses `sync.Mutex` for map access. Background goroutine for rate limiter cleanup.
- **Global state:** None. All state is encapsulated in the `App` struct created in `main()`. The `rateLimiter` background goroutine accesses the `entries` map via mutex.
- **Circular imports:** Not possible — single `package main` with no sub-packages.
- **Oracle read-only constraint:** `isReadOnlySQL()` in `cmd/server/db.go:28` strictly enforces read-only on Oracle connections. DML/DCL/DDL keywords cause query rejection. The guard handles comments, string literals, and nested statements.
- **Template rendering:** All templates are parsed at startup via `template.ParseFS()` and embedded in the binary. No runtime template reloading.

## Anti-Patterns

### Flat Package Structure

**What happens:** All Go source files are in `package main` under `cmd/server/`. There are no internal packages, no `internal/`, no domain separation.
**Why it's wrong:** Creates implicit coupling between all components. Any function can call any other function in any file. No compiler-enforced boundaries between layers (e.g., handler calling another handler).
**Do this instead:** Split into packages like `internal/auth/`, `internal/handlers/`, `internal/db/`, `internal/models/`. See `cmd/server/utils.go:220` and `cmd/server/handlers.go` — they share the same namespace.

### Inline DTO Structs

**What happens:** Several API handler functions define anonymous structs inline for request bodies (e.g., `cmd/server/handlers.go:244`, `cmd/server/api_handlers.go:118`).
**Why it's wrong:** Duplicated across form-based handlers (`handlers.go`) and JSON API handlers (`api_handlers.go`). Hard to reuse, no single source of truth for validation rules.
**Do this instead:** Define shared request/response types in a models file.

### Error Handling — Mixed Patterns

**What happens:** Some handlers use `log.Printf("error: %v", err)` and return a generic message to the client. Others use `writeJSON(w, http.StatusInternalServerError, ...)` directly. No centralized error handler.
**Why it's wrong:** Inconsistent user-facing error messages, risk of leaking stack traces or DB details.
**Do this instead:** Use a helper like `func (a *App) internalError(w, err)` that logs and returns a consistent response.

## Error Handling

**Strategy:** Inline error checks per handler function. No centralized error handler.

**Patterns:**
- `if err != nil { log.Printf(...); http.Error(...) }` — used in page handlers (`handlers.go`)
- `if err != nil { writeJSON(w, status, errorMap) }` — used in API handlers (`api_handlers.go`)
- `if err != nil { log.Fatal(err) }` — used in `main()` and `loadConfig()` for fatal startup errors

## Cross-Cutting Concerns

**Logging:** Standard `log.Printf` — no structured logging. Requests logged via `log` middleware: `"GET /path 200 5ms"`. Errors logged with `"error: %v"` prefix.

**Validation:** Minimal. Password minimum length checked (8 chars). Role validated against allowed set. Activity `finalizeReq.Empresa` accepts `any` and is stringified with `fmt.Sprint`.

**Authentication:** Custom HMAC-SHA256 token in cookie. Session secret required (≥32 chars). Session TTL configurable (default 8h). Session revocation via `last_token_at` column.

**CSRF:** Double-submit cookie pattern for API routes. Origin header check for browser requests. `/login` and `/api/auth/login` are exempted.

---

*Architecture analysis: 2026-06-05*
