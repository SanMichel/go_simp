---
phase: 07-handler-decomposition
plan: 01
subsystem: api
tags: [go, handler-decomposition, service-pattern, pgx, transaction]
requires:
  - phase: 05-error-handling
    provides: AppError type, handleError dispatcher, error code constants
  - phase: 06-testing-infrastructure
    provides: Test helper factories, handler test patterns
provides:
  - finalizeActivity service function (no HTTP types, owns TX lifecycle)
  - FinalizarResult and FlatBundle domain types in models.go
  - Thin apiFinalizar adapter (16 lines)
  - 3 service-level tests for finalizeActivity
affects: Plan 07-02 (remaining 3 handlers follow same extraction pattern)
tech-stack:
  added: []
  patterns:
    - Service function: *App method with context.Context, no HTTP types, returns domain result + *AppError
    - Thin handler: decode JSON, validate fields, call service, encode response (10-20 lines)
    - TX boundary: service owns BeginTx/defer Rollback/Commit, never leaks to handler
key-files:
  created:
    - cmd/server/models.go#FinalizarResult + FlatBundle
    - cmd/server/activity_handlers.go#finalizeActivity
    - cmd/server/main_test.go#TestFinalizeActivity_*
  modified:
    - cmd/server/models.go
    - cmd/server/activity_handlers.go
    - cmd/server/main_test.go
key-decisions:
  - "apiFinalizar keeps two validation checks (JSON decode + required fields) as these are HTTP-decoding concerns; service trusts caller passed validated input"
  - "Service function uses *FinalizarResult return to avoid HTTP types leaking into the business logic boundary"
  - "Result divergence/rupture/replenishment slices initialized with make (not nil) to guarantee JSON [] not null"
patterns-established:
  - "Service function pattern: `func (a *App) serviceName(ctx context.Context, params ...) (*ResultType, error)`"
  - "Thin handler pattern: ~15 lines, decode → validate → call service → encode response"
requirements-completed:
  - HAND-01
  - HAND-03
  - HAND-04
  - HAND-05
duration: 5min
completed: 2026-06-09
---

# Phase 7 Plan 1: Handler Decomposition — apiFinalizar Service Extraction

**Extracted `apiFinalizar` (~101 lines) into a `finalizeActivity` service function (~89 lines) plus a thin handler adapter (~16 lines). Established the service function pattern for downstream Plan 2 replication across remaining 3 handlers.**

## Performance

- **Duration:** 5 min
- **Started:** 2026-06-09T19:44Z
- **Completed:** 2026-06-09T19:49Z
- **Tasks:** 3 / 3
- **Files modified:** 3

## Accomplishments

- `FinalizarResult` and `FlatBundle` types added to models.go — clean domain types with appropriate JSON tags
- `finalizeActivity` service function extracted — owns full TX lifecycle (BeginTx/defer Rollback/Commit), accepts `context.Context`, no HTTP types in signature
- `apiFinalizar` reduced from ~101 lines to ~16 lines — thin adapter that decodes JSON, validates required fields, calls service, and encodes response
- 3 new service-level tests (`TestFinalizeActivity_Success`, `TestFinalizeActivity_MissingFields`, `TestFinalizeActivity_DivergenceDetection`) calling `finalizeActivity` directly
- All 3 existing `TestAPIFinalizar*` handler tests pass unchanged
- Threat model mitigations verified: parameterized SQL preserved, TX ownership contained, empty-slice → `[]` not `null`

## Task Commits

Each task was committed atomically:

1. **Task 1: Add FinalizarResult and FlatBundle types to models.go** - `6d32360` (feat)
2. **Task 2: Extract finalizeActivity service and rewrite apiFinalizar as thin adapter** - `01808a2` (feat)
3. **Task 3: Add service function tests for finalizeActivity** - `e880f2f` (test)

**Plan metadata:** (committed separately with SUMMARY.md)

## Files Created/Modified

- `cmd/server/models.go` - Added `FinalizarResult` (ActivityID, DataFim, Divergences/Ruptures/Replenishments) and `FlatBundle` (with JSON tags matching existing bulk details pattern) after `finalizeReq`
- `cmd/server/activity_handlers.go` - Added `finalizeActivity` service function (~89 lines) with full TX lifecycle; thinned `apiFinalizar` to 16-line adapter
- `cmd/server/main_test.go` - Added 3 `TestFinalizeActivity_*` tests in a new `// ---- Service Function Tests ----` section after the existing finalizar handler tests

## Decisions Made

- **Thin handler keeps input validation**: `apiFinalizar` retains `req.Empresa == 0 || req.Rua == "" || req.SeqLocal == 0` checks because these validate the HTTP request shape — they are HTTP-decoding concerns, not business rules
- **Service uses `*FinalizarResult` return**: Avoids leaking HTTP types (ResponseWriter, etc.) across the service boundary. The handler maps the domain result to JSON
- **`make()` for result slices**: `Divergences`, `Ruptures`, `Replenishments` initialized with `make([]map[string]any, 0)` to guarantee JSON `[]` not `null`, matching original behavior
- **Error messages preserved verbatim**: All Portuguese error strings (`"Erro ao iniciar transação"`, etc.) and `ErrCodeInternal` codes kept identical to original handler

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None — all tests pass, build passes, vet passes.

## Threat Surface Scan

No new threat surface introduced. The extraction is a pure refactor:
- All SQL remains parameterized (no new injection surface)
- Authentication/authorization unchanged (handler was already role-gated at the route level)
- No new endpoints created
- TX lifecycle ownership clarified but behavior unchanged

## Next Phase Readiness

- Service function pattern established: `func (a *App) name(ctx context.Context, params) (*Result, error)`
- Ready for Plan 07-02: extract remaining 3 largest handlers following the same pattern
- `FlatBundle` type extracted (previously inline in api_handlers.go) — available for reuse in downstream bulk-details refactors

## Self-Check: PASSED

- [x] `cmd/server/models.go` exists with FinalizarResult + FlatBundle types
- [x] `cmd/server/activity_handlers.go` exists with finalizeActivity service + thin apiFinalizar
- [x] `cmd/server/main_test.go` exists with 3 TestFinalizeActivity_* functions
- [x] `.planning/phases/07-handler-decomposition/07-01-SUMMARY.md` exists
- [x] Commit 6d32360 (Task 1 — types)
- [x] Commit 01808a2 (Task 2 — service extraction)
- [x] Commit e880f2f (Task 3 — service tests)
- [x] Commit 2a0ce86 (SUMMARY metadata)
- [x] `go vet ./cmd/server` passes
- [x] `go build ./cmd/server` passes
- [x] `go test ./cmd/server` passes (all tests, skipping DB-dependent with TEST_POSTGRES_URL)

---
*Phase: 07-handler-decomposition*
*Completed: 2026-06-09*
