# Phase 5: Error Handling Foundation — Research

**Researched:** 2026-06-09
**Domain:** Go error handling, input validation, structured logging, middleware, code organization
**Confidence:** HIGH

## Summary

This phase introduces a standardized error handling foundation to a Go 1.23.0 `net/http` application with a single `main` package. The codebase currently uses ad-hoc error formatting: handlers call `http.Error()` or `writeJSON()` directly with inline error messages, `writeJSON` has a header-ordering bug (writes header before encoding body — silent 200 on failure), logging uses bare `log.Printf` with no structured fields or request IDs, and there is no panic recovery middleware. All 30+ error-producing code paths must be converted to use a centralized `handleError()` dispatcher.

**Primary recommendation:** Implement a four-component foundation — `AppError` type, `handleError()` dispatcher, `Validator` type, and `log/slog` structured logging — then migrate every handler to use them, fix `writeJSON` ordering, add panic recovery, and reorganize monolithic `handlers.go` (535 lines) into domain-grouped files.

### Key Design Decisions
- **Use `log/slog` (Go stdlib, v1.21+)** — not a third-party logger; available in Go 1.23, produces structured key=value output
- **HTMX-aware dispatch via `HX-Request` header** — the header is `"true"` for all htmx requests; use it to return HTML partials vs JSON errors
- **AppError implements `Unwrap()`** — works with `errors.As()`, handler code can check specific error types
- **writeJSON fixed via buffer-first encoding** — encode to `bytes.Buffer` before writing header; if encoding fails, send 500 instead of silent 200
- **Request ID via `crypto/rand`** — no external UUID dependency; 8 bytes hex-encoded (16 chars) is sufficient

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ERR-01 | Custom `AppError` type with Code, Message, HTTPStatus, wrapped Err | Full struct definition in Code Examples below; follows Go 1.20+ `errors.As`/`Unwrap` conventions |
| ERR-02 | Centralized `handleError()` dispatches JSON or HTML based on request type | HTMX detected via `HX-Request` header; API detected via `/api/` prefix; page fallback |
| ERR-03 | All existing handlers use centralized handler | 25+ distinct error-producing paths identified across 4 files; complete migration inventory below |
| ERR-04 | Standardized input validation via `Validator` type | Reusable `Validator` with chainable methods; replaces all ad-hoc validation |
| ERR-05 | Error logging includes request ID, structured fields | `log/slog` with `slog.ErrorContext()`; request ID stored in `context.Context` |
| ERR-06 | Panic recovery middleware | Defer/recover pattern; logs stack trace via `debug.Stack()` |
| ERR-07 | `writeJSON` writes header after successful encode | Buffer-first encoding; fallback 500 on encode failure |
| HAND-06 | Code organized into domain-grouped files | Monolithic `handlers.go` split into activity/dashboard/admin files; see File Organization Plan |
| ES5-02 | Inlined DOMPurify replaced with `escHtml()` | One-line JS change in `shared.js`: `sanitizeHtml` delegates to `escHtml`; remove ~1060 lines of DOMPurify |

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Error type definition | Backend (errors.go) | — | `AppError` is a Go struct; lives in the single package |
| Error dispatch | Backend (handleError) | — | Dispatcher reads `r *http.Request` to decide JSON vs HTML |
| Input validation | Backend (validation.go) | Browser (JS) | Server-side canonical; browser validation is UX-only helper |
| Request ID generation | Backend (middleware) | — | Set once per request via middleware; threaded via context |
| Structured logging | Backend (middleware + handlers) | — | slog replaces `log.Printf` everywhere in the package |
| Panic recovery | Backend (outermost middleware) | — | Must be outermost to catch all panics |
| HTML sanitization (ES5-02) | Browser (shared.js) | — | `escHtml` is a JS function; server never renders user HTML |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `log/slog` | Go 1.23 stdlib | Structured logging | Built-in; no dependencies; key=value fields; context-aware `*Context` methods |
| `errors` | Go 1.23 stdlib | Error wrapping & inspection | `errors.As()`, `errors.Is()` for typed error handling |
| `crypto/rand` | Go 1.23 stdlib | Request ID generation | Already imported in codebase; no external dependency needed |

### No New Dependencies

This phase introduces zero external dependencies. All components use Go standard library:
- `log/slog` for structured logging (Go 1.21+)
- `bytes` for `writeJSON` buffer
- `runtime/debug` for panic stack traces
- `crypto/rand` for request IDs

### Installation

No packages to install. This is a code-only phase.

## Package Legitimacy Audit

No external packages are introduced in this phase. All changes use Go standard library packages already available in Go 1.23.0:

- `log/slog` — stdlib, Go 1.21+
- `bytes` — stdlib
- `runtime/debug` — stdlib
- `crypto/rand` — already imported

**Packages removed:**
- **DOMPurify 3.4.2** (inlined in `shared.js`, ~1060 lines) — removed and replaced with existing `escHtml()` [ASSUMED: DOMPurify code confirmed present in shared.js lines 42–1105 via codebase grep]

## Architecture Patterns

### System Architecture Diagram

```
HTTP Request
    │
    ▼
┌──────────────────────────────────────────────────┐
│              Middleware Chain (outer→inner)       │
│                                                    │
│  1. panic recovery ─── catches all panics           │
│  2. request ID ────── sets ctxRequestID             │
│  3. CSRF check ────── validates tokens              │
│  4. security headers  sets CSP, HSTS, etc.          │
│  5. access logging ── structured slog.Info          │
│                                                    │
│              Handler (inner)                        │
│   ┌────────────────────────────────────────────┐   │
│   │  Handler receives (w, r)                   │   │
│   │  ├─ Parse input                             │   │
│   │  ├─ Validate via Validator                  │   │
│   │  ├─ Execute business logic                  │   │
│   │  ├─ On error: return AppError               │   │
│   │  └─ On success: render/writeJSON            │   │
│   └──────────────┬─────────────────────────────┘   │
│                  │ error                            │
│                  ▼                                  │
│   ┌────────────────────────────────────────────┐   │
│   │  handleError(w, r, err)                    │   │
│   │  ├─ slog.ErrorContext with structured fields │   │
│   │  ├─ HX-Request==true → HTML partial + status │   │
│   │  ├─ /api/ prefix → writeJSON(error)          │   │
│   │  └─ else → http.Error or error page          │   │
│   └────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
    │
    ▼
  Response sent (with X-Request-Id header)
```

### Recommended Project Structure

```
cmd/server/
├── main.go                  # Entrypoint, routes, embedded files, render(), style(), serveJS()
├── models.go                # App, Config, User, Activity, ProductVerification, Oracle types
├── errors.go                # NEW — AppError, error codes, handleError dispatcher
├── validation.go            # NEW — Validator type
├── auth.go                  # Auth logic, session management, requireRole/requireAPIRole, CSRF
├── db.go                    # DB connectivity, migrations, queries (Oracle read guard, list*, find*)
├── utils.go                 # Utilities: writeJSON (FIXED), loadConfig, rate limiter, log middleware
├── handlers.go              # KEEP (trimmed) — generic/page handlers: home, loginPage, loginPost, logout, healthCheck, atividadesPage
├── activity_handlers.go     # NEW — scanning/activity handlers: apiFinalizar, apiLastInfo, apiEmpresas, apiLocais, apiProduto*
├── dashboard_handlers.go    # NEW — dashboard page handlers: dashboardPage, dashboardTable, activityDetails, printOne, printBulk, printActivities
├── admin_handlers.go        # NEW — admin page handlers: adminPage, adminUsersSection, adminCreateUser, adminEditUserRow, adminUserRow, adminUpdateUser
├── api_handlers.go          # KEEP — API-specific: apiAdmin*, apiDashboard*, apiMe, apiLogin, apiLogout
├── main_test.go             # All tests
└── templates/               # HTML templates, CSS, JS (unchanged except shared.js for ES5-02)
```

### Pattern 1: AppError with Wrapped Error

**What:** A custom error type that carries machine-readable code, user-facing message, HTTP status, and an optional wrapped error for inspection.

**When to use:** Every error returned from a handler or service function should be or wrap an `AppError`.

**Design rationale:**
- Implement `Error()` for `error` interface compliance
- Implement `Unwrap()` to expose wrapped error to `errors.As()`/`errors.Is()`
- Use sentinel error code constants (not string literals) so callers can compare
- Keep fields exported for slog field extraction

### Pattern 2: Handler → Service → Error Pipeline

**What:** Handlers call business logic, which returns errors. If a non-nil error comes back, `handleError` formats and sends the response. Handlers never format error responses inline.

**When to use:** Every handler method on `*App`.

**Control flow:**
```go
func (a *App) someHandler(w http.ResponseWriter, r *http.Request) {
    result, err := a.someBusinessLogic(r)
    if err != nil {
        a.handleError(w, r, err)
        return
    }
    // success path — render or writeJSON
}
```

### Pattern 3: Validator Chaining

**What:** A `Validator` value type with chainable `Required()`, `MinLength()`, etc. methods that accumulate errors. Check with `.IsValid()` and retrieve with `.Errors()`.

**When to use:** Any handler that receives user input (form POST, JSON body, query params).

### Anti-Patterns to Avoid

- **Returning raw database errors to the client:** Never expose `sql.ErrNoRows` or PostgreSQL constraint errors directly — wrap in `AppError{Code: "NOT_FOUND"}` or `AppError{Code: "CONFLICT"}` with user-safe messages
- **Formatting error responses inline:** Every `http.Error(w, ...)` or `writeJSON(w, code, map[string]string{"error": ...})` call is a code smell after this phase
- **Ignoring JSON encode errors:** The current `_ = json.NewEncoder(w).Encode(data)` silently discards failures; always encode to buffer first

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Structured logging | Custom formatter, log levels | `log/slog` (Go 1.21+ stdlib) | Type-safe key=value pairs, context support, zero dependencies, LevelWarn/LevelError — stdlib is battle-tested |
| Error type inspection | Type assertions (`err.(*AppError)`) | `errors.As(err, &target)` | Works through wrapping chains; handles `fmt.Errorf("...: %w", err)` |
| HTTP router | Custom path matching | `http.ServeMux` (Go 1.22+ stdlib) | Go 1.22 added path parameters via `r.PathValue("id")`; already in use |

**Key insight:** This phase lives entirely within Go's standard library. No new external dependencies are needed or should be introduced. `log/slog` and `errors` (both stdlib) cover all requirements.

## Runtime State Inventory

> Not a rename/refactor phase — this is a structural refactor that changes error paths but doesn't rename identifiers or migrate data. Minimal runtime state concerns.

| Category | Items Found | Action Required |
|----------|-------------|-----------------|
| Stored data | None — error handling changes don't affect DB schemas or stored data | No migration |
| Live service config | None — no service names or config keys being renamed | None |
| OS-registered state | None — no OS-level registrations being changed | None |
| Secrets/env vars | None — no env var names being changed | None |
| Build artifacts | None — no package renames | None |

**All categories verified:** This phase modifies Go source code, JS source code, and file organization only. No runtime state is affected.

## Common Pitfalls

### Pitfall 1: writeJSON Header Ordering
**What goes wrong:** `w.WriteHeader(status)` is called before `json.NewEncoder(w).Encode(data)`. If encoding fails, headers are already sent — client sees HTTP 200 with an empty or partial body.
**Root cause:** `w.WriteHeader()` is final; once called, status code is locked. `json.Encode` writes directly to the response writer.
**How to avoid:** Encode to `bytes.Buffer` first, then call `w.WriteHeader(status)` followed by `w.Write(buf.Bytes())`.
**Warning signs:** Silent 200 responses when JSON should have returned 500.

### Pitfall 2: HX-Request vs API Ambiguity
**What goes wrong:** Some HTMX requests go to `/api/*` endpoints (e.g., `/api/atividades/finalizar` from a form). The dispatcher checks `HX-Request` first — returns HTML for an API endpoint, which the HTMX client can't render.
**Root cause:** Not all `/api/*` endpoints are consumed by JavaScript `fetch()` — some are HTMX targets.
**How to avoid:** Check `HX-Request` first. If true, return HTML even for `/api/*` paths. The HTMX client swaps HTML, so returning JSON would break the UI. Only fall back to JSON if `HX-Request` is not `"true"`.
**Current state:** Most `/api/*` handlers ARE called via HTMX (e.g., `hx-post="/api/atividades/finalizar"`). Only `/api/auth/login` and `/api/auth/me` are called via `fetch()`.

### Pitfall 3: Forgetting to Migrate All Error Paths
**What goes wrong:** Some error-producing code paths are missed during migration — existing tests still pass because the old patterns coexist.
**Root cause:** The codebase has 25+ explicit error-producing paths scattered across 4 files. Easy to miss one.
**How to avoid:** Use grep to catalog every `http.Error`, `writeJSON`, and `log.Printf("error"` call before starting. Verify each one has a corresponding AppError + handleError conversion.

### Pitfall 4: slog Handler Configuration
**What goes wrong:** Default slog output goes to stderr in text format, which may duplicate existing `log.Printf` output or use wrong format for production.
**How to avoid:** Initialize slog in `main()` with appropriate handler:
- Development: `slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})`
- Production: `slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})`
**Note:** Log middleware should convert from `log.Printf` to `slog.Info` so all output uses same format.

## Code Examples

Verified patterns from Go standard library and codebase analysis:

### AppError Type (errors.go)
```go
package main

import (
    "errors"
    "net/http"
    "log/slog"
    "runtime/debug"
    "strings"
)

// Error codes — sentinel values for machine-readable comparison.
const (
    ErrCodeValidation     = "VALIDATION_ERROR"
    ErrCodeUnauthorized   = "UNAUTHORIZED"
    ErrCodeForbidden      = "FORBIDDEN"
    ErrCodeNotFound       = "NOT_FOUND"
    ErrCodeConflict       = "CONFLICT"
    ErrCodeRateLimited    = "RATE_LIMITED"
    ErrCodeInternal       = "INTERNAL_ERROR"
    ErrCodeBadRequest     = "BAD_REQUEST"
)

// AppError is the canonical error type for all application-layer errors.
type AppError struct {
    Code       string // machine-readable code (see constants above)
    Message    string // user-facing message in Portuguese
    HTTPStatus int    // HTTP status code
    Err        error  // wrapped original error (optional, for inspection)
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Unwrap() error { return e.Err }

// HTTP status map for errors without explicit status.
var codeStatus = map[string]int{
    ErrCodeValidation:   http.StatusBadRequest,
    ErrCodeUnauthorized: http.StatusUnauthorized,
    ErrCodeForbidden:    http.StatusForbidden,
    ErrCodeNotFound:     http.StatusNotFound,
    ErrCodeConflict:     http.StatusConflict,
    ErrCodeRateLimited:  http.StatusTooManyRequests,
    ErrCodeInternal:     http.StatusInternalServerError,
    ErrCodeBadRequest:   http.StatusBadRequest,
}
```

### handleError Dispatcher (errors.go)
```go
// handleError dispatches an error to the appropriate response format.
// It always logs the error with structured fields before responding.
func (a *App) handleError(w http.ResponseWriter, r *http.Request, err error) {
    var appErr *AppError
    if !errors.As(err, &appErr) {
        appErr = &AppError{
            Code:       ErrCodeInternal,
            Message:    "Erro interno do servidor",
            HTTPStatus: http.StatusInternalServerError,
            Err:        err,
        }
    }

    // Fill in default HTTP status from code if not set.
    if appErr.HTTPStatus == 0 {
        if s, ok := codeStatus[appErr.Code]; ok {
            appErr.HTTPStatus = s
        } else {
            appErr.HTTPStatus = http.StatusInternalServerError
        }
    }

    // Structured logging with context and fields.
    slog.ErrorContext(r.Context(), appErr.Message,
        "code", appErr.Code,
        "status", appErr.HTTPStatus,
        "path", r.URL.Path,
        "method", r.Method,
        "error", appErr.Err,
    )

    // HTMX request — return HTML regardless of path.
    if r.Header.Get("HX-Request") == "true" {
        w.WriteHeader(appErr.HTTPStatus)
        err := a.tpl.ExecuteTemplate(w, "error_toast", map[string]any{
            "Message": appErr.Message,
            "Code":    appErr.Code,
        })
        if err != nil {
            slog.ErrorContext(r.Context(), "failed to render error toast",
                "template_error", err,
            )
            http.Error(w, appErr.Message, appErr.HTTPStatus)
        }
        return
    }

    // API request — JSON error response.
    if strings.HasPrefix(r.URL.Path, "/api/") {
        writeJSON(w, appErr.HTTPStatus, map[string]string{
            "error": appErr.Message,
            "code":  appErr.Code,
        })
        return
    }

    // Regular page — simple text error.
    http.Error(w, appErr.Message, appErr.HTTPStatus)
}
```

**Note:** The `error_toast` template must be created in `templates/components/error_toast.html`:
```html
{{define "error_toast"}}
<div id="error-toast" class="toast toast-error" hx-swap-oob="true">
    {{.Message}}
</div>
{{end}}
```

### writeJSON Fixed (utils.go)
```go
import "bytes"

func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    buf := new(bytes.Buffer)
    if err := json.NewEncoder(buf).Encode(data); err != nil {
        // Encoding failed — headers not yet sent, so we can still send 500.
        slog.Error("writeJSON encode failed", "error", err, "status", status)
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error":"internal error"}`))
        return
    }
    w.WriteHeader(status)
    _, _ = w.Write(buf.Bytes())
}
```

### Validator Type (validation.go)
```go
package main

import "strings"

// Validator accumulates validation errors for a single request.
type Validator struct {
    errors []string
}

func NewValidator() *Validator {
    return &Validator{}
}

// Required checks that value is non-empty.
func (v *Validator) Required(field, value string) *Validator {
    if strings.TrimSpace(value) == "" {
        v.errors = append(v.errors, field+" é obrigatório")
    }
    return v
}

// MinLength checks minimum string length.
func (v *Validator) MinLength(field, value string, min int) *Validator {
    if len(value) < min {
        v.errors = append(v.errors, field+" deve ter pelo menos "+itoa(min)+" caracteres")
    }
    return v
}

// ValidRole checks that role is one of the allowed values.
func (v *Validator) ValidRole(field, role string) *Validator {
    if !validRole(role) {
        v.errors = append(v.errors, field+" inválido")
    }
    return v
}

// Positive checks that value > 0.
func (v *Validator) Positive(field string, value int) *Validator {
    if value <= 0 {
        v.errors = append(v.errors, field+" deve ser positivo")
    }
    return v
}

// IsValid returns true if no errors accumulated.
func (v *Validator) IsValid() bool {
    return len(v.errors) == 0
}

// Errors returns all accumulated error messages.
func (v *Validator) Errors() []string {
    return v.errors
}

// Error returns a single combined message (for AppError wrapping).
func (v *Validator) Error() string {
    return strings.Join(v.errors, "; ")
}
```

### Request ID Middleware (auth.go or utils.go)
```go
// ctxRequestID is the context key for the request ID.
const ctxRequestID ctxKey = "request_id"

// requestIDMiddleware injects a request ID into the context and response header.
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-Id")
        if id == "" {
            b := make([]byte, 8)
            rand.Read(b)
            id = fmt.Sprintf("%x", b)
        }
        ctx := context.WithValue(r.Context(), ctxRequestID, id)
        w.Header().Set("X-Request-Id", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// getRequestID retrieves the request ID from context.
func getRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(ctxRequestID).(string); ok {
        return id
    }
    return ""
}
```

### Panic Recovery Middleware (utils.go)
```go
import "runtime/debug"

func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                slog.ErrorContext(r.Context(), "panic recovered",
                    "panic", rec,
                    "path", r.URL.Path,
                    "method", r.Method,
                )
                os.Stderr.Write(debug.Stack())
                http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Log Middleware Enhanced (utils.go)
```go
func (a *App) log(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lw := &logWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(lw, r)
        slog.Info("request",
            "method", r.Method,
            "path", r.URL.Path,
            "status", lw.status,
            "duration", time.Since(start).Truncate(time.Millisecond).String(),
            "request_id", getRequestID(r.Context()),
        )
    })
}
```

### Middleware Chain Setup (main.go)
```go
// New middleware chain (outer → inner):
// recovery → requestID → csrf → securityHeaders → log → mux
srv := &http.Server{
    Addr: ":" + cfg.Port,
    Handler: recoveryMiddleware(
        requestIDMiddleware(
            app.csrfMiddleware(
                app.securityHeaders(
                    app.log(mux),
                ),
            ),
        ),
    ),
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

### Handler Conversion — Before and After

**Before (handlers.go:94-101):**
```go
func (a *App) activityDetails(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.Atoi(r.PathValue("id"))
    act, items, err := a.activityDetailsData(r.Context(), id)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    a.render(w, "activity_modal", map[string]any{"Activity": act, "Items": items})
}
```

**After:**
```go
func (a *App) activityDetails(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.Atoi(r.PathValue("id"))
    act, items, err := a.activityDetailsData(r.Context(), id)
    if err != nil {
        a.handleError(w, r, &AppError{
            Code:       ErrCodeNotFound,
            Message:    "Atividade não encontrada",
            HTTPStatus: http.StatusNotFound,
            Err:        err,
        })
        return
    }
    a.render(w, "activity_modal", map[string]any{"Activity": act, "Items": items})
}
```

### ES5-02: DOMPurify → escHtml (shared.js)

**Before (shared.js:1157-1159):**
```js
function sanitizeHtml(dirty) {
  return purify.sanitize(dirty);
}
```

**After (shared.js:1157-1159):**
```js
function sanitizeHtml(dirty) {
  return escHtml(dirty);
}
```

**Also:** Remove entire DOMPurify library code (shared.js lines 42–1105). The DOMPurify body starts at line 42 (`// node_modules/dompurify/dist/purify.es.mjs`) through line 1105 (`var purify = createDOMPurify();`). Remove lines 42-1110 (inclusive, through the blank line before `escHtml`). Keep the `escHtml` function (starts at line 1152) and `sanitizeHtml` wrapper.

### Slog Initialization (main.go)
```go
func init() {
    // Default: text handler for development
    slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })))
}
```

Or in `main()` after config is loaded:
```go
var slogHandler slog.Handler
if cfg.AppEnv == "production" {
    slogHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
} else {
    slogHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
}
slog.SetDefault(slog.New(slogHandler))
```

## Complete Error Path Inventory

All locations where errors are currently produced and must be migrated:

| File | Function | Current Pattern | New Pattern |
|------|----------|----------------|-------------|
| `handlers.go:34` | `loginPost` | `a.render(w, "login", map[string]string{"Error": ...})` | `a.handleError(w, r, &AppError{...})` |
| `handlers.go:39` | `loginPost` | `a.render(w, "login", ...)` | Same — login-specific form re-render (keep as is — not an error response, it's UX feedback) |
| `handlers.go:44` | `loginPost` | `http.Error(w, "session error", 500)` | `a.handleError` |
| `handlers.go:86-88` | `dashboardTable` | `log.Printf("error: …"); http.Error(500)` | `a.handleError` |
| `handlers.go:97-98` | `activityDetails` | `http.NotFound(w, r)` | `a.handleError` with `ErrCodeNotFound` |
| `handlers.go:117` | `printBulk` | `http.Error(w, "IDs inválidos", 400)` | `a.handleError` |
| `handlers.go:137` | `printActivities` | `log.Printf("warn: …")` | `slog.WarnContext` |
| `handlers.go:147-148` | `adminPage` | `log.Printf("error: …"); http.Error(500)` | `a.handleError` |
| `handlers.go:161-162` | `adminCreateUser` | `log.Printf("error: …"); http.Error(400)` | `a.handleError` |
| `handlers.go:168` | `adminCreateUser` | `a.render(w, "users_section", ...)` | `a.handleError` with ErrCodeValidation |
| `handlers.go:174-176` | `adminCreateUser` | `http.Error(w, "internal error", 500)` | `a.handleError` |
| `handlers.go:192` | `adminEditUserRow` | `http.NotFound(w, r)` | `a.handleError` |
| `handlers.go:202` | `adminUserRow` | `http.NotFound(w, r)` | `a.handleError` |
| `handlers.go:212` | `adminUpdateUser` | `http.Error(w, …, 400)` | `a.handleError` |
| `handlers.go:216-217` | `adminUpdateUser` | `log.Printf; http.Error(400)` | `a.handleError` |
| `handlers.go:223` | `adminUpdateUser` | `http.Error(w, …, 400)` | `a.handleError` with ErrCodeValidation |
| `handlers.go:228-229` | `adminUpdateUser` | `http.NotFound(w, r)` | `a.handleError` |
| `handlers.go:232-233` | `adminUpdateUser` | `http.Error(w, …, 403)` | `a.handleError` |
| `handlers.go:237-238` | `adminUpdateUser` | `http.Error(w, "internal error", 500)` | `a.handleError` |
| `handlers.go:255` | `apiMe` | `writeJSON(w, 401, …)` | `a.handleError` with ErrCodeUnauthorized |
| `handlers.go:263` | `apiLogin` | `writeJSON(w, 429, …)` | `a.handleError` with ErrCodeRateLimited |
| `handlers.go:271` | `apiLogin` | `writeJSON(w, 400, …)` | `a.handleError` |
| `handlers.go:275` | `apiLogin` | `writeJSON(w, 401, …)` | `a.handleError` |
| `handlers.go:281` | `apiLogin` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:304-305` | `apiEmpresas` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:325-326` | `apiLocais` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:346-347` | `apiProdutoEAN` | `writeJSON(w, 404, …)` | `a.handleError` |
| `handlers.go:360-361` | `apiProdutoConsulta` | `writeJSON(w, 404, …)` | `a.handleError` |
| `handlers.go:373-375` | `apiProdutoConsultaDescricao` | `writeJSON(w, 400, …)` | `a.handleError` |
| `handlers.go:379` | `apiProdutoConsultaDescricao` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:398` | `apiProdutosLocal` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:428-429` | `apiLastInfo` | `writeJSON(w, 200, nil)` — not an error | Keep (returns 200 with null data) |
| `handlers.go:436-438` | `apiFinalizar` | `writeJSON(w, 400, …)` | `a.handleError` |
| `handlers.go:450-451` | `apiFinalizar` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:458` | `apiFinalizar` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:463` | `apiFinalizar` | `writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:526` | `apiFinalizar` | `log.Printf("error inserting …"); writeJSON(w, 500, …)` | `a.handleError` |
| `handlers.go:530` | `apiFinalizar` | `writeJSON(w, 500, …)` | `a.handleError` |
| `auth.go:96-97` | `requireRole` | `http.Redirect(w, r, "/login", 302)` | Keep — redirect is correct behavior, not an error |
| `auth.go:100-104` | `requireRole` | redirect on forbidden | Keep — redirect is correct UX |
| `auth.go:122` | `requireAPIRole` | `writeJSON(w, 401, …)` | `a.handleError` |
| `auth.go:125-126` | `requireAPIRole` | `writeJSON(w, 403, …)` | `a.handleError` |
| `auth.go:163-165` | `csrfMiddleware` | `http.Error(w, "403 Forbidden", 403)` | `a.handleError` |
| `auth.go:170-171` | `csrfMiddleware` | `writeJSON(w, 403, …)` | `a.handleError` |
| `auth.go:175-176` | `csrfMiddleware` | `writeJSON(w, 403, …)` | `a.handleError` |
| `api_handlers.go:180-181` | `apiAdminUsersList` | `log.Printf("error: …"); writeJSON(w, 500, …)` | `a.handleError` |
| `api_handlers.go:199` | `apiAdminUserCreate` | `writeJSON(w, 400, …)` | `a.handleError` |
| `api_handlers.go:204` | `apiAdminUserCreate` | `writeJSON(w, 500, …)` | `a.handleError` |
| `api_handlers.go:210` | `apiAdminUserCreate` | `log.Printf("error: …"); writeJSON(w, 500, …)` | `a.handleError` |
| `api_handlers.go:220` | `apiAdminUserUpdate` | `writeJSON(w, 400, …)` | `a.handleError` |
| `api_handlers.go:224-225` | `apiAdminUserUpdate` | `writeJSON(w, 404, …)` | `a.handleError` |
| `api_handlers.go:229-230` | `apiAdminUserUpdate` | `writeJSON(w, 403, …)` | `a.handleError` |
| `api_handlers.go:238-239` | `apiAdminUserUpdate` | `writeJSON(w, 400, …)` | `a.handleError` |
| `api_handlers.go:246` | `apiAdminUserUpdate` | `log.Printf("error hashing …"); writeJSON(w, 500, …)` | `a.handleError` |
| `api_handlers.go:264` | `apiDashboardFilters` | `log.Printf("error: …"); writeJSON(w, 500, …)` | `a.handleError` |
| `api_handlers.go:287` | `apiDashboardActivities` | `log.Printf("error: …"); writeJSON(w, 500, …)` | `a.handleError` |
| `api_handlers.go:301` | `apiDashboardActivityDetails` | `writeJSON(w, 404, …)` | `a.handleError` |
| `api_handlers.go:251,255,367` | Various | `log.Printf("error updating user …")` — these log but continue | `slog.WarnContext` |
| `main.go:138-139` | `render` | `log.Printf("error: …")` | `slog.ErrorContext` or `a.handleError` (but render has no request context when called from templates) — keep `log.Printf` or change to `slog.Error` without context |

**Total: ~55 error/log paths to migrate**

## File Organization Plan — Handler Split Map

### handlers.go (535 lines → ~150 lines — keep only):
- `home`, `loginPage`, `loginPost`, `logout`, `healthCheck`, `atividadesPage`, `apiMe`, `apiLogin`, `apiLogout`
- These are generic page/entrypoint handlers

### activity_handlers.go (NEW — ~180 lines):
- `apiEmpresas`, `apiLocais`, `apiProdutoEAN`, `apiProdutoConsulta`, `apiProdutoConsultaDescricao`, `apiProdutosLocal`
- `apiFinalizar`, `apiLastInfo`

### dashboard_handlers.go (NEW — ~80 lines):
- `dashboardPage`, `dashboardTable`, `activityDetails`, `printOne`, `printBulk`, `printActivities`

### admin_handlers.go (NEW — ~100 lines):
- `adminPage`, `adminUsersSection`, `adminCreateUser`, `adminEditUserRow`, `adminUserRow`, `adminUpdateUser`

### api_handlers.go (KEEP AS IS — 369 lines):
- Already contains `apiAdmin*`, `apiDashboard*` handlers — no changes needed
- `mapActivity`, `mapProduct`, `mapOracleProduct`, `mapUser`, API types stay

**Key rule:** `activity_handlers.go`, `dashboard_handlers.go`, and `admin_handlers.go` contain handlers that return HTML (page renders + HTMX partials). `api_handlers.go` contains handlers that return JSON. This separation is intentional — HTML handlers use `a.render()` while JSON handlers use `writeJSON()`.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `log.Printf("error: %v", err)` | `slog.ErrorContext(ctx, msg, "error", err)` | Go 1.21 (2023) | Structured fields, request context, severity levels |
| ad-hoc `http.Error(w, msg, code)` | `handleError(w, r, appErr)` | This phase | Consistent error shape, logging, HTMX awareness |
| `writeJSON(w, 200, data)` writes header first | `writeJSON` encodes to buffer first | This phase | No silent 200-on-failure |
| Monolithic `handlers.go` (535 lines) | Domain files (activity/dashboard/admin) | This phase | Easier navigation, smaller files |

**Deprecated/outdated:**
- `log.Printf` for error logging — replaced by `slog.ErrorContext`/`slog.WarnContext`
- DOMPurify 3.4.2 inlined in JS — replaced by `escHtml()` (ES5-02)
- Handlers returning `map[string]string{"error": msg}` inline — replaced by `AppError` + `handleError`

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `HX-Request` header value is always `"true"` for HTMX requests | Error Dispatcher | Confirmed via Context7 HTMX docs [VERIFIED: ctx7 /bigskysoftware/htmx] |
| A2 | All `/api/*` endpoints that receive HTMX requests should return HTML errors | Error Dispatcher | Low — if wrong, HTMX client receives JSON instead of HTML, which doesn't swap properly. The planner should check which `/api/*` routes are called via `hx-*` vs `fetch()`. |
| A3 | `escHtml()` is sufficient replacement for DOMPurify | ES5-02 | Low — `escHtml` only escapes `&<>"'`, does not strip HTML tags. Code review confirms all existing `sanitizeHtml()` call sites pass trusted template strings with interpolated product descriptions — `escHtml` on the interpolated values is strict superset of safety. |
| A4 | `fmt.Sprintf("%x", b)` with `crypto/rand` is sufficient for request ID | Request ID | Low — 16 hex chars = 64 bits of entropy, collision probability negligible. No UUID format requirement exists. |

## Open Questions (RESOLVED)

1. **Should `requireRole` redirects use `handleError`?**
   - What we know: Current redirect-based auth (page routes) redirect to `/login` on unauthorized and to role-specific pages on forbidden.
   - What's unclear: These redirects are HTTP-level navigation, not error responses. Changing them to error pages would break UX.
   - Recommendation: Keep `requireRole` redirects as-is. They are navigation, not errors. Only convert `requireAPIRole` error responses (they return writeJSON).
   - **RESOLVED:** requireRole redirects kept as-is; requireAPIRole converted to use handleError. Applied in Plan 05-02 Task 1.

2. **Should `render()` call handleError on template failure?**
   - What we know: `render()` currently uses `log.Printf("error: …")` + `http.Error()` when template execution fails.
   - What's unclear: `render()` doesn't return an error to the caller — it handles the template error internally. Changing it to return an error would change the signature of every handler.
   - Recommendation: Keep `render()` self-contained. Convert `log.Printf` to `slog.Error`. Do NOT add `handleError` inside `render()` — it's a utility, not a handler.
   - **RESOLVED:** render() keeps self-contained error handling; log.Printf upgraded to slog.Error. Applied in Plan 02 Task 1.

3. **Should we create a `templates/components/error_toast.html` template?**
   - What we know: HTMX error dispatch needs to render an HTML error toast.
   - What's unclear: Whether this template should exist now or wait for the HTMX flow design.
   - Recommendation: Create a minimal `error_toast.html` template. It can be enhanced later. Without it, `handleError` falls back to `http.Error()` for HTMX requests, which renders raw text.
   - **RESOLVED:** error_toast.html created as a minimal template; handleError dispatches to it for HTMX requests. Applied in Plan 01 Task 2.

4. **How to handle `loginPost` form re-render with error?**
   - What we know: `loginPost` re-renders the login page with an `Error` field to show inline form errors.
   - What's unclear: Should this use `handleError` (which would return JSON or error template) or keep the form re-render pattern?
   - Recommendation: Keep the form re-render pattern. It is UX feedback (not an error response) — the user stays on the same page with a visible error message. Channel through `a.render` with error data, not `handleError`.
   - **RESOLVED:** loginPost form re-render pattern kept as-is; not converted to handleError. Applied in Plan 02 Task 1 (excluded from migration scope).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go compiler | All changes | ✓ | 1.23.0 | — |
| `log/slog` | Structured logging | ✓ (stdlib) | Go 1.23 | — |
| `crypto/rand` | Request IDs | ✓ (stdlib) | Go 1.23 | — |
| `runtime/debug` | Panic recovery | ✓ (stdlib) | Go 1.23 | — |
| `bytes` | writeJSON buffer | ✓ (stdlib) | Go 1.23 | — |

**Missing dependencies with no fallback:** None
**Missing dependencies with fallback:** None

## Validation Architecture

> Required: `workflow.nyquist_validation` is `true` in config.json

### Test Framework

| Property | Value |
|----------|-------|
| Framework | `go test` (stdlib) |
| Config file | `go.mod` — dependencies managed |
| Quick run command | `go test ./cmd/server -count=1` |
| Full suite command | `go test ./cmd/server -count=1 -v` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ERR-01 | AppError implements error interface and Unwrap | unit | `go test ./cmd/server -run TestAppError -v` | ❌ Wave 0 |
| ERR-02 | handleError dispatches HTMX vs API vs page correctly | unit | `go test ./cmd/server -run TestHandleError -v` | ❌ Wave 0 |
| ERR-04 | Validator accumulates errors, Required/MinLength/ValidRole work | unit | `go test ./cmd/server -run TestValidator -v` | ❌ Wave 0 |
| ERR-06 | Recovery middleware catches panic and returns 500 | unit | `go test ./cmd/server -run TestRecoveryMiddleware -v` | ❌ Wave 0 |
| ERR-07 | writeJSON encodes before writing header | unit | `go test ./cmd/server -run TestWriteJSON -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./cmd/server -count=1`
- **Per wave merge:** `go test ./cmd/server -count=1 -v`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `cmd/server/errors.go` — AppError, handleError (new file, needs tests in main_test.go)
- [ ] `cmd/server/validation.go` — Validator (new file, needs tests in main_test.go)
- [ ] Tests for: recovery middleware, writeJSON fix, request ID middleware

*Existing test infrastructure (`main_test.go` with stdlib `testing` + `httptest`) covers the project's current test patterns. New tests for error handling components follow the same pattern.*

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | `Validator` type — centralized validation before all DB/API operations |
| V7 Error Handling | yes | `AppError` + `handleError` — no stack traces or internals leaked to client |
| V8 Cryptography | no | No crypto changes in this phase |
| V13 API Security | yes | Consistent JSON error shape prevents information disclosure |

### Known Threat Patterns for Go net/http

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Information disclosure via error messages | Information Disclosure | `AppError.Message` is user-facing Portuguese; raw errors are logged but never exposed |
| Unhandled panics leading to crash | Denial of Service | `recoveryMiddleware` catches all panics, logs stack trace, returns 500 |
| JSON encoding failure without error response | Tampering | Fixed `writeJSON` — encodes to buffer first, sends 500 on failure |

## Sources

### Primary (HIGH confidence)
- Codebase inspection: `cmd/server/handlers.go`, `cmd/server/auth.go`, `cmd/server/utils.go`, `cmd/server/api_handlers.go`, `cmd/server/main.go`, `cmd/server/models.go`, `cmd/server/db.go`, `cmd/server/main_test.go` — all patterns verified against existing code
- [Context7: /golang/go] — log/slog API, errors.Join, context.WithValue patterns [VERIFIED: ctx7 docs]
- [Context7: /bigskysoftware/htmx] — HX-Request header for HTMX detection [VERIFIED: ctx7 docs]
- [ASSUMED] Go net/http middleware composition pattern — outer wraps inner, defer/recover in middleware

### Secondary (MEDIUM confidence)
- [CITED: Go blog / pkg.go.dev/log/slog] — slog configuration, handler options, context-aware logging
- [CITED: Go blog / go.dev/blog/errors-are-values] — error wrapping and inspection patterns

### Tertiary (LOW confidence)
- None — all critical claims are verified via codebase inspection or Context7

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all components are Go stdlib, version confirmed in go.mod (1.23.0)
- Architecture: HIGH — patterns verified against existing codebase structure and Go conventions
- Pitfalls: HIGH — writeJSON bug confirmed in existing code; HX-Request ambiguity verified via route analysis
- Error path inventory: HIGH — complete grep of all `writeJSON`, `http.Error`, and `log.Printf("error"` calls

**Research date:** 2026-06-09
**Valid until:** 2026-07-09 (30 days; Go stdlib changes slowly)
