---
phase: 05-error-handling-foundation
plan: 01
subsystem: core-errors
tags: [golang, slog, error-handling, validation, middleware, htmx]
requires:
  - phase: 04-auth-patterns
    provides: ctxKey pattern, middleware chain structure
provides:
  - AppError type with Code/Message/HTTPStatus/Err fields and Unwrap()
  - handleError dispatcher routing HTMX/JSON/text error responses
  - Validator type with chainable Required/MinLength/ValidRole/Positive methods
  - requestIDMiddleware generating and propagating request IDs via context
  - recoveryMiddleware catching panics and returning HTTP 500
  - Fixed writeJSON with buffer-first encoding (no silent 200-on-failure)
  - log/slog structured logging replacing log.Printf throughout main.go and utils.go
  - error_toast.html template for HTMX error responses
affects: [06-testing-infrastructure, 07-handler-decomposition, 04-auth-patterns]
tech-stack:
  added: [log/slog (Go stdlib), bytes (Go stdlib), runtime/debug (Go stdlib)]
  patterns:
    - "AppError as canonical error type with errors.As() extraction"
    - "handleError dispatcher with HTMX-first dispatch priority"
    - "Validator chaining with fluent interface"
    - "Middleware chain: recovery → requestID → csrf → securityHeaders → log → mux"
    - "buffer-first JSON encoding to prevent partial 200 responses"
    - "slog structured logging with context-aware *Context methods"
key-files:
  created:
    - cmd/server/errors.go — AppError type, error codes, handleError, requestIDMiddleware, getRequestID
    - cmd/server/validation.go — Validator type with chainable validation methods
    - cmd/server/templates/components/error_toast.html — HTMX error toast partial
  modified:
    - cmd/server/utils.go — fixed writeJSON, added recoveryMiddleware, log middleware to slog
    - cmd/server/main.go — slog init, middleware chain update, all log.Printf→slog replacements
key-decisions:
  - "recoveryMiddleware placed in utils.go (not errors.go) to keep error types separate from middleware"
  - "Middleware order: recovery outermost → requestID → csrf → securityHeaders → log → mux (access log sees final status)"
  - "Non-AppError errors wrapped with ErrCodeInternal (generic message, never leaks raw error text)"
  - "HX-Request checked first (before /api/ prefix) because many /api/ handlers are HTMX targets"
  - "slog.Warn for non-fatal Oracle ping failure instead of log.Printf"
patterns-established:
  - "All future handler code uses a.handleError(w, r, err) instead of inline writeJSON/http.Error"
  - "All future validation uses Validator chainable methods"
  - "All structured logging uses slog with context fields"
requirements-completed: [ERR-01, ERR-02, ERR-04, ERR-05, ERR-06, ERR-07]
duration: 15 min
completed: 2026-06-09
---

# Phase 05: Error Handling Foundation Summary

**AppError type with HTMX-aware handleError dispatcher, Validator chaining, buffer-first writeJSON, panic recovery middleware, request ID tracing, and slog structured logging replacing log.Printf**

## Performance

- **Duration:** 15 min
- **Started:** 2026-06-09T13:45:00Z (approx)
- **Completed:** 2026-06-09T14:00:00Z (approx)
- **Tasks:** 3
- **Files created:** 3
- **Files modified:** 2

## Accomplishments

- Created `AppError` canonical error type with 8 machine-readable error codes, `codeStatus` HTTP mapping, and `handleError` dispatcher that routes HTMX→HTML partial, `/api/*`→JSON, others→`http.Error`
- Implemented `requestIDMiddleware` with `crypto/rand` 64-bit hex ID injection into context and `X-Request-Id` response header
- Built `Validator` type with chainable `Required`, `MinLength`, `ValidRole`, `Positive` methods and `IsValid()/Errors()/Error()` accessors
- Fixed `writeJSON` buffer-first encoding bug (was writing header before encoding — silent 200 on failure)
- Added `recoveryMiddleware` as outermost handler catching all panics with `debug.Stack()` logging
- Migrated all logging from `log.Printf` to `log/slog` structured logging across `utils.go` and `main.go`
- Updated middleware chain to include recovery and request ID: `recovery → requestID → csrf → securityHeaders → log → mux`
- Created `error_toast.html` template for HTMX error responses

## Task Commits

Each task was committed atomically:

1. **Task 1: Create errors.go with AppError, error codes, handleError, request ID middleware** - `778ba13` (feat)
2. **Task 2: Create validation.go and error_toast.html template** - `b3a6523` (feat)
3. **Task 3: Fix writeJSON, add recoveryMiddleware, update log middleware to slog, add slog init** - `8f4f98e` (feat)

_Note: All tasks were type="auto" with no TDD cycle._

## Files Created/Modified

- `cmd/server/errors.go` — AppError, 8 error code constants, codeStatus map, handleError method, ctxRequestID, requestIDMiddleware, getRequestID
- `cmd/server/validation.go` — Validator struct, NewValidator, Required/MinLength/ValidRole/Positive chain methods, IsValid/Errors/Error accessors
- `cmd/server/templates/components/error_toast.html` — `{{define "error_toast"}}` with toast-error div for HTMX OOB swap
- `cmd/server/utils.go` — Fixed writeJSON (buffer-first), added recoveryMiddleware, updated log middleware to slog.Info with request_id
- `cmd/server/main.go` — slog init with JSON/Text handler based on AppEnv, middleware chain updated, all log.Printf replaced with slog equivalents

## Decisions Made

- **Middleware ordering:** recoveryMiddleware is outermost to catch panics from all inner layers; requestIDMiddleware is second so all inner middleware (including log) have access to request ID; log is innermost of the security/csrf/log chain to log final status codes including those set by csrf middleware
- **HTMX-first dispatch:** handleError checks `HX-Request` header before `/api/` prefix because many `/api/*` endpoints are HTMX targets (they need HTML error toasts, not JSON)
- **Non-AppError wrapping:** Any error passed to handleError that is not an AppError is wrapped with `ErrCodeInternal` and generic "Erro interno do servidor" — raw error text never reaches client
- **slog.Warn for Oracle ping:** Non-fatal Oracle ping failure logs as `slog.Warn` instead of `log.Printf` for structured severity

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed unused imports from errors.go**
- **Found during:** Task 1 (errors.go creation)
- **Issue:** Plan listed `bytes`, `os`, `runtime/debug` in errors.go import list, but these are only needed for recoveryMiddleware (placed in utils.go per plan structure). Go would fail to compile with unused imports.
- **Fix:** Omitted unused imports from errors.go. Moved `bytes`, `os`, `runtime/debug` to utils.go imports where recoveryMiddleware lives.
- **Files modified:** cmd/server/errors.go
- **Verification:** `go build ./cmd/server` succeeds
- **Committed in:** 778ba13 (Task 1 commit)

**2. [Rule 2 - Missing Critical] Replaced Oracle ping log.Printf with slog.Warn**
- **Found during:** Task 3 (plan verification step 4)
- **Issue:** Plan's verification step 4 requires no `log.Printf` (except `log.Fatal`) in utils.go or main.go, but `log.Printf("warning: oracle ping failed")` at main.go:51 was not listed for explicit replacement in Task 3
- **Fix:** Replaced with `slog.Warn("oracle ping failed", "error", err)` for structured logging with error field
- **Files modified:** cmd/server/main.go
- **Verification:** `grep -c 'log\.Printf\|log\.Println' cmd/server/utils.go cmd/server/main.go | grep -v 'log\.Fatal'` returns no matches
- **Committed in:** 8f4f98e (Task 3 commit)

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 missing critical)
**Impact on plan:** Both auto-fixes necessary for correctness (Go compilation) and plan verification. No scope creep.

## Issues Encountered

None — all tasks executed as planned with minor auto-fixes for compilation and verification compliance.

## User Setup Required

None — no external service configuration required. All changes use Go standard library.

## Next Phase Readiness

- Error handling foundation complete. Ready for Phase 5 Plan 02 (handler migration to handleError) and Phase 6 (testing infrastructure)
- All handler code after this plan should use `a.handleError(w, r, err)` instead of inline `writeJSON(...)` or `http.Error(...)`
- All input validation after this plan should use `Validator` chainable methods

## Self-Check: PASSED

All 5 committed files verified present on disk:
- `cmd/server/errors.go` — FOUND
- `cmd/server/validation.go` — FOUND
- `cmd/server/templates/components/error_toast.html` — FOUND
- `cmd/server/utils.go` — FOUND
- `cmd/server/main.go` — FOUND

All 4 commits verified in git log:
- `778ba13` — feat(05-01): create errors.go
- `b3a6523` — feat(05-01): create validation.go and error_toast.html
- `8f4f98e` — feat(05-01): fix writeJSON, add recoveryMiddleware, migrate to slog
- `b1c068b` — docs(05-01): complete error handling foundation plan

---

*Phase: 05-error-handling-foundation*
*Completed: 2026-06-09*
