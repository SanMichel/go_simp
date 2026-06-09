# Architecture Patterns

**Domain:** Go stdlib warehouse app — ES5 JS frontend, camera barcode scanning, test organization, error handling
**Researched:** 2026-06-08

## Current Architecture (unchanged)

```
cmd/server/
├── main.go           ← Entrypoint, routes, go:embed, render
├── handlers.go       ← HTTP handler functions
├── auth.go           ← Auth, sessions, RBAC middleware
├── db.go             ← Postgres + Oracle queries, migrations
├── models.go         ← Data structures, App/Config structs
├── utils.go          ← Helpers (JSON, rate limiter, etc.)
├── main_test.go      ← All tests (single file)
└── templates/        ← Embedded templates, CSS, JS
    ├── *.html
    ├── components/*.html
    ├── *.css
    ├── *.js           ← 6 JS files (ES6+, compiled from TS)
    └── htmx.min.js    ← HTMX v2.0.4
```

## Recommended Architecture Changes

### 1. JS Build Pipeline (NEW)

```
project root/
├── package.json                       ← npm config
├── babel.config.json                   ← ES5 transpilation targets
├── cmd/server/
│   ├── templates/                      ← SOURCE JS (write here, edit here)
│   │   ├── app.js
│   │   ├── shared.js
│   │   ├── dashboard.js
│   │   ├── admin.js
│   │   ├── login.js
│   │   ├── scanner.js                  ← NEW: Quagga2 wrapper
│   │   ├── htmx.min.js                 ← REPLACED with v1.9.12
│   │   └── quagga.min.js               ← NEW: self-hosted Quagga2
│   ├── templates-dist/                 ← BUILT JS (Babel output, gitignored)
│   │   ├── app.js                      ← ES5-transpiled
│   │   ├── shared.js                   ← ES5-transpiled
│   │   └── ...                         ← All other JS files transpiled
│   └── main.go                         ← go:embed updated to embed "templates-dist/*.js"
└── Makefile                             ← NEW: build-js target
```

**Build flow:**
```
npm run build-js
  → Babel reads cmd/server/templates/*.js (excluding htmx/quagga)
  → Outputs ES5 version to cmd/server/templates-dist/
  → htmx.min.js + quagga.min.js copied as-is (vendored from node_modules)

go build ./cmd/server
  → Embeds templates-dist/*.js (ES5-transpiled)
  → Embeds templates/*.html, templates/*.css (unchanged)
```

**Key design decisions:**
- Source JS stays in `templates/` — all editing happens here
- Built JS goes to `templates-dist/` — never edited by hand
- Only `.js` files go through Babel; `.html`, `.css`, `.min.js` files pass through unchanged
- `go:embed` directive updates to read from `templates-dist/` for JS, `templates/` for everything else
- `templates-dist/` is gitignored; CI builds it from source

### 2. Quagga2 Integration (NEW component within existing pattern)

```
Browser → existing serveJS("quagga.min.js") → serves Quagga2 runtime
Browser → existing serveJS("scanner.js")     → serves scanner wrapper

scanner.js pattern:
  function initScanner(options) {
    if (!navigator.mediaDevices?.getUserMedia) return false;  // feature detect
    Quagga.init({...});
    Quagga.onDetected((data) => {
      document.getElementById("scan-input").value = data.codeResult.code;
      document.getElementById("form-scan").requestSubmit();    // HTMX integration
    });
    Quagga.start();
    return true;
  }
```

- No Go-side changes needed for scanner integration
- Camera feed renders in existing DOM via Quagga's built-in renderer
- Scan results flow into existing `scan-input` field → existing HTMX form submission
- Fallback: `initScanner()` returns `false` → existing manual input flow continues unchanged

### 3. Error Handling Middleware (within existing middleware chain)

```
Current middleware chain (main.go:66):
  app.csrfMiddleware(app.securityHeaders(app.log(mux)))

Proposed middleware chain:
  app.csrfMiddleware(app.securityHeaders(app.recover(app.log(app.errorHandler(mux)))))
                                    ^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^
                                    NEW: panic recovery  NEW: standard error mapping

errorHandler middleware:
  - Wraps each handler response
  - Intercepts known error types → maps to HTTP status + JSON error body
  - Type switch: NotFoundError → 404, ValidationError → 422, etc.
  - Uses slog to log errors with context (method, path, duration)

recover middleware:
  - defer/recover with debug.Stack()
  - Logs panic with slog.Error, stack trace
  - Returns 500 with standard error envelope
```

### 4. Test File Organization (within single main package)

```
cmd/server/
├── main_test.go           ← Existing + new tests (still single file)
├── testdata/              ← NEW: test fixtures (not a package)
│   ├── valid_session.json
│   └── sample_activity.json
```

Tests stay in a single `main_test.go` file (package `main`). Use `t.Run` subtests for organization:

```go
func TestHandlers(t *testing.T) {
    t.Run("health check returns 200", func(t *testing.T) { ... })
    t.Run("login with valid credentials", func(t *testing.T) { ... })
    t.Run("login with wrong password returns 401", func(t *testing.T) { ... })
}
```

## Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| Babel pipeline | Transpile JS to ES5 | `cmd/server/templates/` (input), `cmd/server/templates-dist/` (output) |
| Quagga2 (browser) | Camera access, barcode image processing, decoding | Camera hardware, scanner.js wrapper |
| scanner.js | Quagga2 init, feature detection, scan result → input field | Quagga2, DOM (scan-input), existing form-scan flow |
| Error middleware | Catch panics, map errors to HTTP status+body, structured logging | slog.Logger, http.ResponseWriter, handler chain |
| Test suite | Verify handler behavior, utility functions, middleware | httptest.ResponseRecorder, testdata/ fixtures |

## Data Flow for Barcode Scanning

```
User clicks "Scan with Camera"
  → scanner.js checks navigator.mediaDevices.getUserMedia
    → FAIL: Show message "Camera not available", fall back to manual input
    → OK: Quagga2 activates camera, renders viewfinder in DOM
      → User points camera at barcode
        → Quagga2 detects + decodes frame
        → onDetected callback fires
          → Sets scan-input value
          → Triggers form-scan submit (same as manual entry)
          → Existing API call to /api/produtos/ean/{code}
          → HTMX updates UI with result
```

## Patterns to Follow

### Pattern 1: Sentinel Error Mapping

```go
// models.go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrValidation   = errors.New("validation error")
    ErrConflict     = errors.New("conflict")
)

// errors.go (NEW or in utils.go)
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"error"`
    Err     error  `json:"-"`
    Status  int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NotFound(msg string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: msg, Status: 404}
}
func Unauthorized(msg string) *AppError {
    return &AppError{Code: "UNAUTHORIZED", Message: msg, Status: 401}
}
func Validation(msg string) *AppError {
    return &AppError{Code: "VALIDATION", Message: msg, Status: 422}
}
```

### Pattern 2: Error Wrapping with Context

```go
// Instead of:
user, err := db.QueryUser(id)
if err != nil {
    return nil, err  // BARE: no context
}

// Do:
user, err := db.QueryUser(id)
if err != nil {
    return nil, fmt.Errorf("get user %s: %w", id, err)  // WRAPPED: context + chain
}
```

### Pattern 3: slog Structured Logging

```go
// Instead of:
log.Printf("error: %v", err)

// Do:
slog.Error("failed to process scan",
    "err", err,
    "user_id", userID,
    "ean", scannedCode,
    "trace_id", traceID,
)
```

### Pattern 4: Table-Driven Tests with httptest

```go
func TestHealthCheck(t *testing.T) {
    app := &App{loginLimiter: newRateLimiter()}
    req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
    rec := httptest.NewRecorder()
    app.healthCheck(rec, req)

    assert(t, rec.Code == http.StatusOK, "expected 200, got %d", rec.Code)
    assert(t, strings.Contains(rec.Body.String(), `"status":"ok"`), "unexpected body: %s", rec.Body.String())
}

func assert(t *testing.T, ok bool, msg string, args ...any) {
    t.Helper()
    if !ok {
        t.Fatalf(msg, args...)
    }
}
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: polyfill.io CDN
**What:** Loading polyfills from the compromised CDN.
**Why bad:** Known supply chain attack. Serves credential-stealing JS.
**Instead:** Self-host `core-js` modules via Babel's `useBuiltIns: "usage"`.

### Anti-Pattern 2: Log and return
**What:** Logging an error AND returning it to the caller.
**Why bad:** Double-handling — error gets logged at every stack frame, produces noise.
**Instead:** Log at the top level (middleware/handler), wrap-and-return at lower levels.

### Anti-Pattern 3: External test framework
**What:** Adding Testify or similar as a Go dependency.
**Why bad:** Violates project constraint. Adds dependency maintenance.
**Instead:** Stdlib `testing` + helper functions. `gotestsum` is a CLI tool, not a framework.

### Anti-Pattern 4: Bypassing Babel for individual JS files
**What:** Only transpiling SOME files or doing ad-hoc manual ES5 rewrites.
**Why bad:** Inconsistent output, missed ES6 features, maintenance burden.
**Instead:** All custom JS files run through Babel. HTMX v1 and Quagga2 are already compatible.

## Scalability Considerations

| Concern | Current | After changes |
|---------|---------|---------------|
| JS file count | 6 custom + 1 vendor | 6 custom + 2 vendor (htmx + quagga) |
| Build time | None | +1-2 seconds for Babel |
| Go binary size | ~20MB | +~500KB (Quagga2), +~100KB (HTMX v1 is same size as v2) |
| Test count | 15 tests, 323 lines | Target: 70%+ coverage |
| Error log quality | Unstructured, no context | Structured with slog, request IDs, trace context |

## Sources

- Go error handling patterns: bugsly.dev (March 2025), bytesizego.com, JetBrains Go Guide — HIGH confidence
- `log/slog` usage: Go 1.21 release notes — HIGH confidence
- HTMX v1↔v2 migration: htmx.org docs — HIGH confidence
- Quagga2 browser API requirements: GitHub README — MEDIUM confidence (no explicit ES5 verification)
- Babel + core-js integration: core-js.io/docs — HIGH confidence
- polyfill.io compromise: Cloudflare blog, Checkmarx, Qualys, Lokker — HIGH confidence (multiple independent confirmations)
