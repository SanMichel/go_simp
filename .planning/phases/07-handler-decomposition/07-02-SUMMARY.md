---
phase: 07-handler-decomposition
plan: 02
subsystem: api
tags: [go, handler-decomposition, service-pattern, shared-service, refactor]
requires:
  - phase: 05-error-handling
    provides: AppError type, handleError dispatcher, error code constants
  - phase: 06-testing-infrastructure
    provides: Test helper factories, handler test patterns
  - phase: 07-handler-decomposition
    plan: 01
    provides: Service pattern, FlatBundle type in models.go
provides:
  - updateUserAdmin shared service function (no HTTP types)
  - bulkActivityDetails service function (loop + mapping, reused FlatBundle)
  - 3 thin handler adapters (adminUpdateUser, apiAdminUserUpdate, apiDashboardBulkDetails)
  - 6 service-level tests for updateUserAdmin and bulkActivityDetails
affects: Phase 08 (ES5 compatibility — uses handler APIs unchanged)
tech-stack:
  added: []
  patterns:
    - Shared service across HTML and JSON handlers: updateUserAdmin used by both adminUpdateUser (form) and apiAdminUserUpdate (JSON)
    - Loop extraction: bulkActivityDetails owns the ID-loop + mapping logic, handler is pure delegation
key-files:
  created: []
  modified:
    - cmd/server/admin_handlers.go
    - cmd/server/api_handlers.go
    - cmd/server/main_test.go
key-decisions:
  - "updateUserAdmin returns *UserRow (not *User) to hide PasswordHash from template rendering; converts internally from findUserByID result"
  - "Slog.ErrorContext (not WarnContext) used in service DB update error logging, matching adminUpdateUser original; apiAdminUserUpdate was slightly inconsistent using WarnContext"
  - "apiAdminUserUpdate error code for invalid role changed from ErrCodeValidation to ErrCodeBadRequest (service uses ErrCodeBadRequest, matching adminUpdateUser original — both handlers now identical)"
patterns-established:
  - "Shared service across form + JSON handlers: service owns all permission logic (self-edit guard, role validation, sysadmin protection), handler only extracts request params"
  - "Loop extraction: bulkActivityDetails owns the error-silent-skip pattern for failed activityDetailsData calls, matching original behavior"
requirements-completed:
  - HAND-02
  - HAND-03
  - HAND-04
  - HAND-05
duration: 11 min
completed: 2026-06-09
---

# Phase 7 Plan 2: Handler Decomposition — updateUserAdmin + bulkActivityDetails Service Extraction

**Extracted `adminUpdateUser` (43 lines) and `apiAdminUserUpdate` (41 lines) into a shared `updateUserAdmin` service, and `apiDashboardBulkDetails` (35 lines) into `bulkActivityDetails` service. Both handlers become ~15 line thin adapters. All 3 remaining large handlers decomposed — completes all HAND targets for Phase 7.**

## Performance

- **Duration:** 11 min
- **Started:** 2026-06-09T16:44Z
- **Completed:** 2026-06-09T16:55Z
- **Tasks:** 3 / 3
- **Files modified:** 3

## Accomplishments

- `updateUserAdmin` shared service function created — used by both `adminUpdateUser` (HTML form) and `apiAdminUserUpdate` (JSON API), eliminating duplicate permission logic
- `adminUpdateUser` reduced from 43 to 14 body lines — pure form parsing + service delegation + template rendering
- `apiAdminUserUpdate` reduced from 41 to 12 body lines — pure JSON decode + service delegation + response
- `bulkActivityDetails` service function extracted — owns the ID-loop, mapActivity/mapProduct calls, FlatBundle construction, and nil-to-empty-slice guard
- `apiDashboardBulkDetails` reduced from 35 to 6 body lines — pure JSON decode + service call + response
- Inline `FlatBundle` type removed from api_handlers.go (now only in models.go, via Plan 1)
- 6 new service-level tests call service functions directly (not via HTTP)

## Task Commits

Each task was committed atomically:

1. **Task 1: Extract updateUserAdmin service, rewrite adminUpdateUser and apiAdminUserUpdate** - `8ea4a4d` (feat)
2. **Task 2: Extract bulkActivityDetails service and rewrite apiDashboardBulkDetails** - `edb8f96` (feat)
3. **Task 3: Add service function tests for updateUserAdmin and bulkActivityDetails** - `fe69229` (test)

## Files Created/Modified

- `cmd/server/admin_handlers.go` - Added `updateUserAdmin` service function (~40 lines) before `adminUpdateUser`; thinned `adminUpdateUser` to 14 body lines (form parse → call service → render); added `"context"` import
- `cmd/server/api_handlers.go` - Added `bulkActivityDetails` service function (~25 lines); thinned `apiAdminUserUpdate` to 12 body lines and `apiDashboardBulkDetails` to 6 body lines; removed inline `FlatBundle` type; added `"context"` import
- `cmd/server/main_test.go` - Added 6 `TestUpdateUserAdmin_*` and `TestBulkActivityDetails_*` tests in new `// ---- Service Function Tests ----` section

## Decisions Made

- **`updateUserAdmin` returns `*UserRow` not `*User`**: Hides `PasswordHash` from template rendering. The function converts `*User` to `*UserRow` after the DB update re-fetch.
- **Error code normalization**: `apiAdminUserUpdate` used `ErrCodeValidation` for invalid role, but `adminUpdateUser` used `ErrCodeBadRequest`. The shared service uses `ErrCodeBadRequest` (matching the HTML handler's original behavior) — both handlers now respond identically to invalid roles.
- **Log level consistency**: `adminUpdateUser` used `slog.ErrorContext` for DB update errors; `apiAdminUserUpdate` used `slog.WarnContext`. The shared service uses `slog.ErrorContext` — matching the more conservative (higher severity) logging pattern.
- **`bulkActivityDetails` skips errors silently**: Matching original `apiDashboardBulkDetails` behavior, errors from `activityDetailsData` per ID result in that ID being silently skipped. Error details are never exposed to the client.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `findUserByID` returns `*User`, not `*UserRow`. The plan specified `(*UserRow, error)` as the return type. Fixed by converting from `*User` to `*UserRow` after the re-fetch query. No behavior change — the template renders the same subset of fields.

## Threat Surface Scan

No new threat surface introduced. The extraction is a pure refactor:
- All SQL remains parameterized (no new injection surface)
- Permission checks (self-edit guard, sysadmin protection, validRole) preserved identically in the service
- Error messages preserved verbatim in Portuguese
- `updateUserAdmin` adds defense-in-depth permission checks (original handlers were already route-gated by `requireRole("sysadmin")`)

## Next Phase Readiness

- All 4 Phase 7 handlers decomposed: `apiFinalizar` (Plan 1), `adminUpdateUser`, `apiAdminUserUpdate`, `apiDashboardBulkDetails` (Plan 2)
- Total removed: ~220 lines of handler code → ~80 lines of thin adapters + ~65 lines of service + ~65 lines of service tests
- Ready for Phase 8: ES5 Compatibility

## Self-Check: PASSED

- [x] `go vet ./cmd/server` passes
- [x] `go build ./cmd/server` passes
- [x] `cmd/server/admin_handlers.go` has `updateUserAdmin` service + thin `adminUpdateUser`
- [x] `cmd/server/api_handlers.go` has `bulkActivityDetails` service + thin handlers
- [x] `cmd/server/main_test.go` has 6 new service tests
- [x] `grep -c 'type FlatBundle struct' cmd/server/api_handlers.go` returns 0
- [x] No HTTP types in service functions
- [x] Commit 8ea4a4d (Task 1 — service extraction)
- [x] Commit edb8f96 (Task 2 — service extraction)
- [x] Commit fe69229 (Task 3 — service tests)

---
*Phase: 07-handler-decomposition*
*Completed: 2026-06-09*
