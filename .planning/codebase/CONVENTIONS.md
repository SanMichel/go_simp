# Coding Conventions

**Analysis Date:** 2026-06-05

## Naming Patterns

**Files:**
- Go source: `snake_case.go` — e.g., `main.go`, `handlers.go`, `api_handlers.go`, `main_test.go`
- HTML templates: `snake_case.html` — e.g., `login.html`, `atividades.html`, `dashboard.html`
- Component templates: `snake_case.html` in `templates/components/` — e.g., `nav.html`, `user_row.html`
- CSS: `snake_case.css` — `style.css`, `admin.css`
- JS: `snake_case.js` — `shared.js`, `app.js`, `dashboard.js`

**Functions:**
- Exported: `PascalCase` — used only for Go's `main()` package function
- Unexported methods on `*App`: `camelCase` — e.g., `findUserByID`, `listActivities`, `parseFilters`, `redirectByRole`
- Unexported standalone helpers: `camelCase` — e.g., `validRole`, `firstNonEmpty`, `intQuery`, `randomString`, `getenv`, `loadDotEnv`, `newRateLimiter`
- HTTP handlers: `camelCase` — e.g., `loginPage`, `loginPost`, `healthCheck`, `home`, `adminPage`
- API handlers: `apiPascalCase` — e.g., `apiMe`, `apiLogin`, `apiEmpresas`, `apiDashboardActivities`
- Template mapping/conversion: `mapPascalCase` — e.g., `mapActivity`, `mapProduct`, `mapUser`

**Variables:**
- Go: `camelCase` for local vars and parameters — e.g., `req`, `cfg`, `mux`, `rec`, `handler`, `dataFim`
- Single-letter for loop indices: `i` in `for` loops over slices
- Short abbreviations: `u` for `*User`, `p` for `ProductVerification`/`OracleProduct`, `a` for `*App` receiver, `rl` for `*rateLimiter`
- Error vars: `err` (single assignment), `errOra` (when a second error var is needed in the same scope)

**Types:**
- Struct types: `PascalCase` — e.g., `User`, `Activity`, `Config`, `App`, `OracleReader`, `ProductVerification`, `FilterOptions`, `OracleEmpresa`, `OracleLocal`, `OracleProduct`
- API response types: `APIPascalCase` — e.g., `APIActivity`, `APIProductVerification`, `APIUser`, `APIFilterOptions`
- Internal DTOs: `camelCase` lowercase — e.g., `finalizeReq`, `Bundle` (local to functions), `rateEntry`, `logWriter`
- Inline structs: defined locally in functions without a named type — e.g., in `printActivities`, `apiProdutosLocal`, `apiDashboardFilters`

**JSON tags:**
- UPPERCASE for Oracle-sourced structs — e.g., `json:"NROEMPRESA"`, `json:"SEQLOCAL"`, `json:"CODACESSO"`
- Mixed case (snake_case) for app-internal API types — e.g., `json:"id"`, `json:"username"`, `json:"dataFim"`, `json:"seqproduto"`, `json:"atividade_id"`
- All lowercase for generic API response maps — e.g., `"error"`, `"message"`, `"success"`

## Code Style

**Formatting:**
- Standard `gofmt` (no `.golangci.yml`, no `.editorconfig`, no `.prettierrc` — project relies on Go's built-in formatting only)
- Import groups: standard library first, followed by third-party, separated by a blank line. No grouping for internal imports since this is a single `package main`.

**Linting:**
- No golangci-lint, no eslint, no biome — zero linting tools detected
- Imports are explicit with `_` prefix for driver-only imports: `_ "github.com/jackc/pgx/v5/stdlib"`, `_ "github.com/sijms/go-ora/v2"`

## Import Organization

**Order:**
1. Standard library packages (alphabetically: `context`, `crypto/*`, `database/sql`, `embed`, `encoding/*`, `errors`, `fmt`, `html/template`, `log`, `net/http`, `os`, `path/filepath`, `strconv`, `strings`, `sync`, `testing`, `time`)
2. Third-party packages (alphabetically): `github.com/jackc/pgx/...`, `github.com/sijms/go-ora/...`, `golang.org/x/crypto/...`

**Path Aliases:**
- Import alias `go_ora "github.com/sijms/go-ora/v2"` is used in `utils.go` to call `go_ora.BuildUrl`
- No other import aliases used

## Error Handling

**Patterns:**
```go
// Standard if-err-return (most common pattern)
if err != nil {
    log.Printf("error: %v", err)
    writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Erro interno do servidor"})
    return
}

// Handler pattern — log then render error
if err != nil {
    log.Printf("error: %v", err)
    http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
    return
}

// Template render error — caught via a.render wrapper
func (a *App) render(w http.ResponseWriter, name string, data any) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := a.tpl.ExecuteTemplate(w, name, data); err != nil {
        log.Printf("error: %v", err)
        http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
    }
}

// Fatal on startup — log.Fatal used for unrecoverable startup errors
if err != nil {
    log.Fatal("postgres ping: ", err)
}
```

**Key conventions:**
- All errors are logged via `log.Printf("error: %v", err)` — never silent drops. Exception: `json.NewDecoder(r.Body).Decode(&req)` in `apiAdminUserCreate` silently discards decode error (potential bug).
- HTTP handlers return user-facing error messages in Portuguese (e.g., `"Erro interno do servidor"`, `"Usuário ou senha incorretos."`)
- API handlers return Portuguese JSON error responses (e.g., `map[string]string{"error": "Não autorizado"}`)
- `_` is used to discard errors only in deliberately best-effort operations: row scans inside loops (`rows.Scan(...) == nil`), cookie writes, deferred rollback, bulk updates
- No `defer` error checking pattern (deferred `Rollback` error is discarded)

## Logging

**Framework:** Go's `log` standard library — no third-party logger.

**Patterns:**
```go
// Startup / fatal
log.Fatal(err)

// Warning (non-fatal startup)
log.Printf("warning: oracle ping failed")

// Route requests (middleware)
log.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start).Truncate(time.Millisecond))

// Errors in handlers
log.Printf("error: %v", err)

// Server ready
log.Printf("server ready on http://localhost%s", addr)
```

- No structured logging, no log levels (only `log.Printf` and `log.Fatal`)
- Middleware wraps `http.ResponseWriter` with `logWriter` to capture response status code

## Comments

**When to Comment:**
- SQL comments are written in Portuguese at the point of use for complex queries (no block comments above, just inline)
- No JSDoc/TSDoc-style documentation comments — Go standard `//` comments only
- No exported function documentation (this is a single `package main` with only unexported functions)
- HTML template comments use HTML `<!-- ── Section ──────────────────────── -->` style for visual section separation
- No TODO/FIXME/HACK comments present
- One code comment exists: `// FilterOptions can just have lowercase JSON tags` in `api_handlers.go:173`

**Go doc comments:** Not used — all functions are unexported within a single `main` package.

## Function Design

**Size:**
- Most handler functions are 3-15 lines (simple page renders)
- Largest function: `listActivities` (81 lines, `db.go:204-285`) — dynamic SQL builder
- `apiFinalizar` (68 lines, `handlers.go:383-452`) — second largest

**Parameters:**
- HTTP handlers follow `func(w http.ResponseWriter, r *http.Request)` or `func(w http.ResponseWriter, r *http.Request, u *User)` (for API handlers with auth)
- Database methods accept `ctx context.Context` as first parameter — e.g., `findUserByID(ctx, id)`, `listActivities(ctx, f, limit)`
- Helper functions use positional parameters — no Config struct for function options
- `serveJS` returns `http.HandlerFunc` via closure pattern

**Return Values:**
- Error is always the last return value — standard Go convention
- Tuple returns `(T, error)` for most data retrieval functions
- Mutation functions return only `error` or nothing
- `writeJSON` and `render` handle their own errors internally, return nothing

## Module Design

**Exports:**
- **No exported symbols** — every type, function, method, variable and constant is unexported (lowercase). The entire app is `package main` with no public API surface.
- Only the `main()` function is exported (by Go convention for program entrypoint)

**Barrel Files:**
- Not used — single `package main` with no sub-packages
- Files are organized by concern (handlers, auth, db, models, utils, api_handlers) but all in the same package

**File Organization:**
- `main.go`: Entrypoint, route registration, `render`, static file serving, template parsing, `go:embed` FS
- `models.go`: All data structures (`Config`, `App`, `User`, `Activity`, `ProductVerification`, Oracle types, request DTOs)
- `utils.go`: Config loading (`loadConfig`, `getenv`, `loadDotEnv`), middleware (`log`, `securityHeaders`, `csrfMiddleware`), helper utilities (`writeJSON`, `parseFilters`, `validRole`, `firstNonEmpty`, `intQuery`, `newRateLimiter`/`allow`)
- `auth.go`: Session management (`currentUser`, `makeToken`, `randomString`, `revokeSession`), role middleware (`requireRole`, `requireAPI`, `redirectByRole`), CSRF middleware (`setCSRFCookie`, `clearCSRFCookie`, `csrfMiddleware`)
- `handlers.go`: Page handlers (`home`, `loginPage`, `loginPost`, `logout`, `atividadesPage`, `dashboardPage`, etc.), print handlers, API handlers that mix auth under `*User` (`apiMe`, `apiLogin`, `apiLogout`, `apiEmpresas`, etc.)
- `api_handlers.go`: API-only handlers (`apiAdminUsersList`, `apiAdminUserCreate`, `apiAdminUserUpdate`, `apiDashboardFilters`, `apiDashboardActivities`, `apiDashboardActivityDetails`, `apiDashboardBulkDetails`, `apiDashboardBulkPrint`), mapper functions (`mapActivity`, `mapProduct`, `mapUser`), API-level DTOs (`APIActivity`, `APIProductVerification`, `APIUser`)
- `db.go`: Oracle reader wrapper (`OracleReader.QueryContext`, `OracleReader.QueryRowContext`, `isReadOnlySQL`, `removeSQLComments`), migrations (`migrate`), seeder (`seedAdmin`), database queries (`findUserByUsername`, `findUserByID`, `listUsers`, `listActivities`, `listFilterOptions`, `activityDetailsData`, `findAddressByCode`, `findFullProductByCode`)
- `main_test.go`: All test functions

**HTML Templates:**
- Templates use `{{define "name"}}...{{end}}` named block syntax
- Components are in `templates/components/` and loaded via `ParseFS(templatesFS, "templates/*.html", "templates/components/*.html")`
- Template functions registered in `template.FuncMap`: `rowUser`, `date`, `rolePt`, `checked`

---

*Convention analysis: 2026-06-05*
