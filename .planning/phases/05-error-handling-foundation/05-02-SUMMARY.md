---
phase: 05-error-handling-foundation
plan: 02
subsystem: api
tags: [error-handling, apperror, handleerror, slog, structured-logging]
requires:
  - phase: 05-error-handling-foundation
    provides: 05-01 AppError type, error code constants, handleError dispatcher, error_toast template
provides:
  - All 35 error paths in handlers.go use a.handleError with AppError
  - All 5 error paths in auth.go use a.handleError with AppError
  - log.Printf replaced by slog.ErrorContext/slog.WarnContext throughout handlers.go
  - Consistent error response pipeline: AppError → handleError → structured log + response
affects: [05-03 api_handlers.go migration]
tech-stack:
  added: [log/slog for structured logging in error paths]
  patterns:
    - "Inline http.Error(http.NotFound/writeJSON replaced by a.handleError(w, r, &AppError{...})"
    - "log.Printf(\"error/warn:...\") replaced by slog.ErrorContext/slog.WarnContext with structured fields"
key-files:
  created: []
  modified:
    - cmd/server/handlers.go - 35 error paths migrated from inline formatting to a.handleError
    - cmd/server/auth.go - 5 error paths (requireAPIRole, csrfMiddleware) migrated to a.handleError
key-decisions:
  - "requireRole redirects kept as-is (navigation, not errors)"
  - "loginPost form re-renders kept as-is (UX feedback, not error responses)"
  - "log.Printf(\"error:...\") in log-and-continue patterns (adminUpdateUser UPDATE errors) replaced with slog.ErrorContext"
  - "adminCreateUser INSERT error kept as form re-render (UX feedback) with slog.ErrorContext logging"
  - "apiLastInfo writeJSON(nil, 200) kept as-is (intentional empty response pattern)"
  - "dashboardPage/printActivities silent error-discard patterns (activities, _ :=) kept as-is"
patterns-established:
  - "All error paths in core handlers flow through AppError → handleError → structured log → response"
  - "Form re-renders with inline error messages are distinct from API error responses and stay as-is"
requirements-completed:
  - ERR-03
  - ERR-05
duration: 8min
completed: 2026-06-09
---

# Phase 05: Error Handling Foundation — Plan 02 Summary

**All error-producing paths in handlers.go and auth.go migrated to centralized a.handleError dispatcher with AppError types and structured slog logging**

## Performance

- **Duration:** 8 min
- **Started:** 2026-06-09T10:51:00-03:00
- **Completed:** 2026-06-09T10:59:21-03:00
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- All 35 error paths in handlers.go converted from inline `http.Error()`, `http.NotFound()`, and `writeJSON(w, status, map[string]string{"error":...})` to `a.handleError(w, r, &AppError{...})` with appropriate error code constants
- All 5 error paths in auth.go (`requireAPIRole` unauthorized/forbidden, `csrfMiddleware` origin mismatch/CSRF missing/CSRF invalid) migrated to `a.handleError`
- All `log.Printf("error:...")` calls in handlers.go replaced with `slog.ErrorContext` (structured fields, request context)
- All `log.Printf("warn:...")` calls in handlers.go replaced with `slog.WarnContext`
- `requireRole` redirects preserved unchanged (navigation semantics)
- `loginPost` form re-renders preserved unchanged (UX feedback pattern)

## Task Commits

Each task was committed atomically:

1. **Task 1: Migrate all error paths in handlers.go** - `e9ec8b1` (feat)
2. **Task 2: Migrate auth.go error paths to handleError** - `e9ec8b1` (feat, same commit — both files migrated in batch)

**Plan metadata:** Pending (docs: complete plan)

## Files Created/Modified

- `cmd/server/handlers.go` - 35 error paths migrated to `a.handleError`; `log.Printf` replaced with `slog.ErrorContext`/`slog.WarnContext`
- `cmd/server/auth.go` - 5 error paths (`requireAPIRole` × 2, `csrfMiddleware` × 3) migrated to `a.handleError`

## Decisions Made

- **requireRole redirects preserved:** The plan explicitly excludes redirect-based auth from the migration — these are HTTP navigation, not error responses. Only `requireAPIRole` (which returns JSON) was converted.
- **loginPost form re-renders preserved:** Login form validation errors re-render the login page with an inline message. This is UX feedback, not an error response — kept as-is.
- **log-and-continue patterns:** Silent error discard patterns (`activities, _ := a.listActivities(...)`) remain. The log.Printf inside error-only branches (`adminUpdateUser` UPDATE failures) were upgraded to `slog.ErrorContext`.
- **adminCreateUser INSERT error:** This is a form re-render with error state (UX feedback), not a response error — kept as form re-render with `slog.ErrorContext` logging.

## Deviations from Plan

None - plan executed exactly as written.

### Acceptance Criteria Verification

All acceptance criteria met:

| Check | Expected | Actual | Status |
|-------|----------|--------|--------|
| `go build ./cmd/server` succeeds | yes | yes | PASS |
| `a.handleError` in handlers.go | >= 30 | 35 | PASS |
| `writeJSON(w, map[string]string{"error":...})` in handlers.go | <= 5 | 0 | PASS |
| `http.Error(w` in handlers.go | 0 | 0 | PASS |
| `http.NotFound(w` in handlers.go | 0 | 0 | PASS |
| `log.Printf` in handlers.go | <= 2 | 0 | PASS |
| `ErrCode` in handlers.go | >= 30 | 35 | PASS |
| `a.handleError` in auth.go | 5 | 5 | PASS |
| `writeJSON...map[string]string` in auth.go | 0 | 0 | PASS |
| `http.Redirect` in auth.go | >= 4 | 4 | PASS |
| CSRF/access messages preserved | > 0 | 3 | PASS |
| `go test ./cmd/server -count=1` passes | yes | yes | PASS |

## Issues Encountered

None - all error paths migrated cleanly, build and tests pass.

## Next Phase Readiness

Ready for Plan 05-03: Migrate api_handlers.go error paths to a.handleError and split dashboard_handlers.go / admin_handlers.go.

## Self-Check: PASSED

- ✅ `go build ./cmd/server` succeeds
- ✅ `go test ./cmd/server -count=1` passes
- ✅ `a.handleError` in handlers.go: 35 (>= 30)
- ✅ `http.Error(w` in handlers.go: 0
- ✅ `http.NotFound(w` in handlers.go: 0
- ✅ `log.Printf` in handlers.go: 0 (<= 2)
- ✅ `ErrCode` in handlers.go: 35 (>= 30)
- ✅ `a.handleError` in auth.go: 5
- ✅ `writeJSON...map[string]string` in auth.go: 0
- ✅ `http.Redirect` in auth.go: 4 (>= 4)
- ✅ CSRF/Access denied messages preserved: 3 (> 0)
- ✅ Commit e9ec8b1 exists

---

*Phase: 05-error-handling-foundation*
*Completed: 2026-06-09*
