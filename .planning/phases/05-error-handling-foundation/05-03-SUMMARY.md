---
phase: 05-error-handling-foundation
plan: 03
subsystem: api
tags:
  - error-handling
  - slog
  - migration
  - api

requires:
  - phase: 05-error-handling-foundation
    plan: 01
    provides: AppError type, handleError dispatcher, error code constants
provides:
  - All api_handlers.go error paths use a.handleError with AppError
  - All log.Printf replaced with slog.WarnContext
affects: []

tech-stack:
  added:
    - log/slog (to replace log.Printf in api_handlers.go)
  patterns:
    - Error response pipeline: handler → a.handleError → slog logging + writeJSON

key-files:
  created: []
  modified:
    - cmd/server/api_handlers.go
    - cmd/server/handlers.go (unused import fix)
    - .gitignore (server binary)

key-decisions:
  - "Explicit slog.ErrorContext calls alongside a.handleError are not added as handleError already logs via slog.ErrorContext internally"
  - "Hash error messages normalized to generic 'Erro interno do servidor' to avoid leaking internal details"

requirements-completed:
  - ERR-03
  - ERR-05

duration: 5min
completed: 2026-06-09
---

# Phase 5 Plan 3: Migrate api_handlers.go Error Paths Summary

**All 12+ error paths in api_handlers.go migrated from inline writeJSON error formatting to a.handleError with AppError; all log.Printf calls replaced with slog.WarnContext**

## Performance

- **Duration:** 5 min
- **Started:** 2026-06-09T13:51:43Z
- **Completed:** 2026-06-09T13:56:46Z
- **Tasks:** 1
- **Files modified:** 3

## Accomplishments

- 12 error paths converted from `writeJSON(w, code, map[string]string{"error": ...})` to `a.handleError(w, r, &AppError{...})`
- 3 warn-level `log.Printf` calls (non-fatal, execution continues) replaced with `slog.WarnContext` with structured fields
- 8 error-producing `log.Printf` calls removed — handled by `handleError`'s internal `slog.ErrorContext` logging
- All 8 success `writeJSON(w, http.StatusOK, ...)` calls preserved unchanged
- Import cleanup: removed unused `"log/slog"` from handlers.go (pre-existing from parallel plan 05-02)
- Added `server` binary to `.gitignore`

## Task Commits

Each task was committed atomically:

1. **Task 1: Migrate all error paths in api_handlers.go** - `72d4c8d` (feat)

**Plan metadata:** Pending (orchestrator handles final commit)

## Files Created/Modified

- `cmd/server/api_handlers.go` - All 12 error paths migrated, 3 warn-level log.Printf replaced with slog.WarnContext, `"log"` import removed
- `cmd/server/handlers.go` - Removed unused `"log/slog"` import (Rule 3 - blocking pre-existing issue)
- `.gitignore` - Added `server` binary

## Decisions Made

- **No redundant slog.ErrorContext alongside handleError:** The plan noted explicit `slog.ErrorContext` calls as optional because `handleError` already logs via `slog.ErrorContext` internally. Adding redundant calls would double-log error entries. The acceptance criteria threshold (≥4 slog calls in api_handlers.go) was set expecting explicit calls, but the 3 explicit calls are all `slog.WarnContext` for warn-level non-fatal events. All other error-producing paths log via `handleError`'s internal logging.
- **Hash error messages normalized:** `"Erro ao criar hash"` changed to generic `"Erro interno do servidor"` to avoid leaking internal implementation details about password hashing.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed unused "log/slog" import from handlers.go**
- **Found during:** Task 1 (test execution)
- **Issue:** Pre-existing unused `"log/slog"` import in handlers.go (from parallel Plan 05-02) prevented `go test` from compiling
- **Fix:** Removed the unused import
- **Files modified:** `cmd/server/handlers.go`
- **Verification:** `go build ./cmd/server` and `go test ./cmd/server -count=1` both pass
- **Committed in:** `72d4c8d` (part of task commit)

**2. [Rule 2 - Missing Critical] Added 'server' binary to .gitignore**
- **Found during:** Task 1 (post-commit check for untracked files)
- **Issue:** Go build output binary `server` was not gitignored, appearing as untracked file
- **Fix:** Added `server` to `.gitignore`
- **Files modified:** `.gitignore`
- **Verification:** `git status --short` no longer shows `server` as untracked
- **Committed in:** `72d4c8d` (part of task commit)

---

**Total deviations:** 2 auto-fixed (1 Rule 3 - blocking, 1 Rule 2 - missing critical)
**Impact on plan:** Both auto-fixes necessary for correctness. No scope creep.

## Acceptance Criteria Verification

| Criterion | Result |
|-----------|--------|
| `go build ./cmd/server` succeeds | ✅ PASS |
| `go test ./cmd/server -count=1` passes | ✅ PASS (0.088s) |
| `a.handleError` count >= 12 | ✅ PASS (12) |
| `log.Printf` count == 0 | ✅ PASS (0) |
| `writeJSON(http.StatusOK)` count >= 6 | ✅ PASS (8) |
| `ErrCode` count >= 12 | ✅ PASS (12) |
| `slog.WarnContext|slog.ErrorContext` >= 4 | ⚠️ 3 vs ≥4 (threshold overcount — handleError logs internally, redundant explicit calls deliberately omitted) |

## Issues Encountered

- **Unused import in handlers.go:** Pre-existing `"log/slog"` import in handlers.go from parallel Plan 05-02 caused test build failure. Fixed via Rule 3.
- **Slog acceptance criteria threshold:** Plan expected ≥4 explicit slog calls in api_handlers.go, but only 3 warn-level replacements exist. The remaining error-producing paths log through handleError's internal slog.ErrorContext (in errors.go), making additional explicit calls redundant. This is intentional, not a bug.

## Threat Surface Scan

No new security-relevant surface introduced. All error responses continue to use user-safe Portuguese messages per the existing threat model (T-05-01, T-05-02, T-05-07).

## Next Phase Readiness

- Plan 05-03 complete — api_handlers.go error paths fully migrated
- Plan 05-02 (handlers.go, auth.go migration) runs in parallel
- Ready for full-wave merge and verification

## Self-Check: PASSED

All files verified on disk. Commit `72d4c8d` confirmed in git log. Build and tests pass.

---

*Phase: 05-error-handling-foundation*
*Completed: 2026-06-09*
