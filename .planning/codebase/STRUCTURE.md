# Codebase Structure

**Analysis Date:** 2026-06-08

## Directory Layout

```
go-simp/
├── cmd/
│   └── server/                    # Single main package — entire application
│       ├── main.go                # Entrypoint, route registration, template setup, go:embed
│       ├── models.go              # Core data structures (User, Activity, App, Config, etc.)
│       ├── handlers.go            # HTTP page handlers + inline API handlers (finalizar, etc.)
│       ├── api_handlers.go        # Dedicated JSON API handlers + response mappers
│       ├── auth.go                # Session tokens, RBAC middleware, CSRF protection
│       ├── db.go                  # Database queries, migrations, seeding, Oracle read guard
│       ├── utils.go               # Config loading, .env parser, logging, rate limiter, helpers
│       ├── main_test.go           # All test cases (no external test framework)
│       └── templates/             # Embedded HTML templates, CSS, JS (via go:embed)
│           ├── login.html         # Login page template
│           ├── home.html          # Home/welcome page
│           ├── atividades.html    # Activity scanning SPA (app.js driven)
│           ├── dashboard.html     # Dashboard SPA (dashboard.js driven)
│           ├── admin.html         # User admin SPA (admin.js driven)
│           ├── print.html         # Print view for activities
│           ├── style.css          # Main stylesheet (atividades layout)
│           ├── admin.css          # Admin/dashboard stylesheet
│           ├── shared.js          # Shared JS utilities (login, API helpers)
│           ├── app.js             # Atividades scanning SPA logic
│           ├── dashboard.js       # Dashboard SPA logic
│           ├── admin.js           # Admin panel SPA logic
│           ├── login.js           # Login page JS enhancement
│           ├── htmx.min.js        # Vendored HTMX library
│           └── components/        # HTMX template partials
│               ├── head.html      # HTML head section
│               ├── nav.html       # Navigation bar
│               ├── activities_table.html  # Activities table partial (HTMX target)
│               ├── activity_modal.html    # Activity detail modal partial
│               ├── users_section.html     # Admin user list partial
│               ├── user_row.html          # Single user table row partial
│               └── user_edit_row.html     # User edit form row partial
├── tmp/                           # REFERENCE COPY OF OLD APP — DO NOT MODIFY
│   ├── src/                       # Old TypeScript architecture
│   │   ├── client/                # Old client code
│   │   ├── server/                # Old server code
│   │   └── types.ts               # Old type definitions
│   ├── package.json
│   ├── tsconfig.json
│   ├── drizzle.config.ts
│   ├── Dockerfile
│   └── docker-compose.yml
├── .air.toml                      # Air hot-reload configuration
├── .env.example                   # Environment template (copy to .env)
├── go.mod                         # Go module definition (module go-simp, go 1.23)
├── go.sum                         # Go dependency checksums
├── AGENTS.md                      # Project agent instructions
└── README.md                      # Project readme
```

## Directory Purposes

**`cmd/server/`:**
- Purpose: Single executable Go package — all application code lives here
- Contains: 7 `.go` source files + 1 test file + embedded templates directory
- Key files: `cmd/server/main.go` (entrypoint + routes), `cmd/server/models.go` (App struct + data types)

**`cmd/server/templates/`:**
- Purpose: All frontend assets — HTML templates, CSS, JS, HTMX library
- Contains: 7 page templates, 7 component partials, 2 CSS files, 6 JS files, 1 vendored HTMX
- Embedded into binary via `go:embed` at `cmd/server/main.go:206`

**`cmd/server/templates/components/`:**
- Purpose: HTMX-targeted template partials (server-rendered fragments)
- Contains: Table rows, modal, user editing components

**`tmp/`:**
- Purpose: Reference copy of previous TypeScript-based application architecture
- Contains: Old Bun/Drizzle/Docker stack
- Constraint: **DO NOT MODIFY** — read-only reference only

## Key File Locations

**Entry Points:**
- `cmd/server/main.go:20`: `func main()` — application bootstrap, server start, graceful shutdown
- `cmd/server/main.go:85`: `func (a *App) routes(mux)` — all route registration in one method

**Configuration:**
- `cmd/server/utils.go:17`: `loadConfig()` — environment variable parsing into `Config` struct
- `cmd/server/utils.go:96`: `loadDotEnv()` — manual `.env` file parser (no godotenv)
- `.air.toml`: Hot-reload configuration for development
- `.env.example`: Environment variable template (copy to `.env`)

**Core Logic:**
- `cmd/server/handlers.go`: All page-rendering HTTP handlers (`loginPage`, `dashboardPage`, `adminPage`, `atividadesPage`, etc.)
- `cmd/server/api_handlers.go`: All JSON API handlers (`apiAdminUsersList`, `apiDashboardActivities`, etc.) + API response mappers
- `cmd/server/auth.go`: Session token creation/validation, RBAC middleware (`requireRole`, `requireAPIRole`), CSRF middleware
- `cmd/server/db.go`: All database logic — queries, migrations, seeding, Oracle read-only guard
- `cmd/server/models.go`: Core data structures (`App`, `Config`, `OracleReader`, all entity types)

**Data Models:**
- `cmd/server/models.go:9`: `Config` struct
- `cmd/server/models.go:22`: `App` struct (dependency container)
- `cmd/server/models.go:30`: `OracleReader` struct
- `cmd/server/models.go:34-101`: Entity types (`User`, `UserRow`, `Activity`, `ActivityFilters`, `FilterOptions`, `ProductVerification`)
- `cmd/server/models.go:103-129`: Oracle entity types (`OracleEmpresa`, `OracleLocal`, `OracleProduct`)
- `cmd/server/models.go:130-147`: Request types (`finalizeReq`)

**Testing:**
- `cmd/server/main_test.go`: All tests — standard `testing` package, no external test framework

**Middleware:**
- `cmd/server/auth.go:86`: `requireRole()` — page route RBAC
- `cmd/server/auth.go:112`: `requireAPIRole()` — API route RBAC
- `cmd/server/auth.go:156`: `csrfMiddleware()` — CSRF protection for mutating requests
- `cmd/server/utils.go:130`: `log()` — request logging middleware
- `cmd/server/utils.go:227`: `securityHeaders()` — security header middleware

**Template Functions:**
- `cmd/server/main.go:175`: `parseTemplates()` — template compilation with custom funcs (`date`, `rolePt`, `checked`, `rowUser`)

## Naming Conventions

**Files:**
- Go source: Snake case (`main.go`, `api_handlers.go`, `main_test.go`)
- Templates: Snake case with context suffix — `*.html`, `*.css`, `*.js` (e.g., `dashboard.html`, `admin.css`, `app.js`)
- Component templates: Underscore-separated with semantic names (`activity_modal.html`, `user_edit_row.html`)

**Functions:**
- Exported: None in this package (all functions are lowercase, package-internal)
- Handlers: camelCase with descriptive name — `dashboardPage`, `loginPost`, `apiProdutoEAN`, `adminCreateUser`
- Middleware: camelCase — `requireRole`, `csrfMiddleware`, `securityHeaders`
- Helpers: camelCase — `loadConfig`, `loadDotEnv`, `writeJSON`, `parseFilters`, `validRole`

**Variables:**
- camelCase throughout Go code — `sessionTTL`, `pgURL`, `activityID`
- Receiver: `a` for `*App`, `rl` for `*rateLimiter`
- Template parameter: Struct fields PascalCase by convention (exported from `map[string]any`)

**Types:**
- PascalCase structs — `User`, `Activity`, `OracleProduct`, `Config`, `OracleReader`
- No interface types defined (no interfaces in the entire codebase)
- API response types prefixed with `API` — `APIActivity`, `APIProductVerification`, `APIUser`
- Oracle types prefixed with `Oracle` — `OracleEmpresa`, `OracleLocal`, `OracleProduct`

**Packages:**
- Single package: `main` — no sub-packages or `internal/` modules

## Where to Add New Code

**New Feature (e.g., new API endpoint):**
1. Add request/response types to `cmd/server/api_handlers.go` (or `cmd/server/models.go` for shared types)
2. Add query function to `cmd/server/db.go` if new data access is needed
3. Add handler method on `*App` to `cmd/server/handlers.go` (page) or `cmd/server/api_handlers.go` (JSON API)
4. Register route in `func (a *App) routes(mux)` in `cmd/server/main.go`
5. Add role protection using `requireRole()` or `requireAPIRole()`
6. Add test to `cmd/server/main_test.go`

**New Template/Component:**
1. Create `.html` file in `cmd/server/templates/` or `cmd/server/templates/components/`
2. Define template with `{{define "template_name"}}`
3. Ensure it matches the `go:embed` pattern on `cmd/server/main.go:206` (currently `templates/*.html`, `templates/components/*.html`, etc.)
4. Call `a.render(w, "template_name", data)` from handler
5. For HTMX partials, the component template is rendered standalone

**New Client JS:**
1. Add `.js` file to `cmd/server/templates/`
2. Register route in `main.go` via `a.serveJS("filename.js")` (`main.go:89-94`)
3. Include `<script src="/filename.js">` in the HTML template

**New Middleware:**
1. Add middleware function in `cmd/server/auth.go` (auth/csrf) or `cmd/server/utils.go` (generic)
2. Wrap the handler in `main.go:66` middleware chain

**New Database Table:**
1. Add `CREATE TABLE IF NOT EXISTS` statement to `app.migrate()` in `cmd/server/db.go:111`
2. Add corresponding model struct in `cmd/server/models.go`
3. Add query/insert functions in `cmd/server/db.go`

## Special Directories

**`cmd/server/templates/`:**
- Purpose: All embeddable frontend assets
- Generated: No
- Committed: Yes
- Note: Embedded into binary — runtime file changes are not reflected until rebuild

**`tmp/`:**
- Purpose: Reference copy of the old TypeScript application
- Generated: No
- Committed: Yes
- Constraint: **DO NOT MODIFY** — read-only reference

**`bin/` and `.tmp/`:**
- Purpose: Build artifacts
- Generated: Yes (via `go build` / `air`)
- Committed: No (gitignored)

**`.planning/`:**
- Purpose: Codebase map documents and planning artifacts
- Generated: Yes (by `/gsd-map-codebase` and `/gsd-plan-phase`)
- Committed: Yes

---

*Structure analysis: 2026-06-08*
