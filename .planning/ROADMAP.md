# Roadmap: go-simp

## Overview

Warehouse activity scanning and logistics dashboard. Existing codebase migrated from TypeScript to Go.

## Phases

- [x] **Phase 1: Foundation** — Go migration, core architecture, auth, DB setup
- [x] **Phase 2: Activity Scanning** — HTMX-driven scanning SPA, activity registration
- [x] **Phase 3: Dashboard** — Operational metrics and monitoring dashboard
- [x] **Phase 4: Admin** — User management, RBAC administration

- [x] **Phase 5: Error Handling Foundation** — Custom AppError, centralized error dispatch, validation, file reorg
- [ ] **Phase 6: Testing Infrastructure** — Table-driven handler tests, auth/db/error tests, 70%+ coverage
- [ ] **Phase 7: Handler Decomposition** — Decompose overgrown handlers into thin adapters + service layer
- [ ] **Phase 8: ES5 Compatibility** — ES5 JS rewrite, HTMX compat verification, page weight optimization

## Phase Details

### Phase 1: Foundation

**Goal**: Go server with auth, database, and basic structure
**Depends on**: Nothing
**Success Criteria**:

  1. Server starts and serves HTTP
  2. Users can log in with email/password
  3. Postgres auto-migrates on startup

**Plans**: N/A (existing)

### Phase 2: Activity Scanning

**Goal**: Warehouse scanning workflows
**Depends on**: Phase 1
**Success Criteria**:

  1. Conferentes can scan and register activities
  2. Activity data persisted to Postgres
  3. HTMX partials render activity tables and modals

**Plans**: N/A (existing)

### Phase 3: Dashboard

**Goal**: Operational metrics dashboard
**Depends on**: Phase 2
**Success Criteria**:

  1. Gerentes can view real-time operational data
  2. Dashboard SPA loads and updates properly

**Plans**: N/A (existing)

### Phase 4: Admin

**Goal**: User management panel
**Depends on**: Phase 1
**Success Criteria**:

  1. Sysadmins can create, edit, delete users
  2. Role assignment works correctly
  3. Admin SPA with HTMX dynamic rows

**Plans**: N/A (existing)

### Phase 5: Error Handling Foundation

**Goal**: Standardized error handling, input validation, and code organization provide a consistent foundation for all milestone work
**Depends on**: Nothing (first phase of v1.1)
**Requirements**: ERR-01, ERR-02, ERR-03, ERR-04, ERR-05, ERR-06, ERR-07, HAND-06, ES5-02
**Success Criteria** (what must be TRUE):

   1. All handlers deliver errors through the centralized `handleError()` dispatcher instead of inline formatting
   2. Input validation across all form/JSON endpoints uses the standardized `Validator` type
   3. Error logs include request ID and structured fields; panic recovery middleware prevents server crashes
   4. `writeJSON` encodes before writing HTTP headers (no silent 200-on-failure responses)
   5. Code is organized into domain-grouped files (errors.go, validation.go, activity_handlers.go, dashboard_handlers.go, admin_handlers.go)

**Plans**: 4 plans in 3 wavesPlans:
**Wave 1**

- [x] 05-01-PLAN.md — Foundation: AppError, handleError, Validator, writeJSON fix, slog, recovery middleware

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 05-02-PLAN.md — handlers.go + auth.go error path migration
- [x] 05-03-PLAN.md — api_handlers.go error path migration

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 05-04-PLAN.md — File reorganization (HAND-06) + ES5-02 (DOMPurify→escHtml)

### Phase 6: Testing Infrastructure

**Goal**: Comprehensive test coverage provides a safety net for refactoring and future changes
**Depends on**: Phase 5
**Requirements**: TEST-01, TEST-02, TEST-03, TEST-04, TEST-05, TEST-06
**Success Criteria** (what must be TRUE):

  1. All existing handlers have table-driven tests using `httptest` covering success and error paths
  2. Auth middleware and session management have dedicated unit tests
  3. Database query functions have tests (transactional or mocked)
  4. Error handling types (`AppError`, `handleError`) have dedicated unit tests
  5. All existing routes return correct status codes and response shapes; overall coverage ≥ 70%

**Plans**: 4 plans in 4 waves
Plans:
**Wave 1** *(no dependencies)*
- [ ] 06-01-PLAN.md — Foundation: testhelper.go + error handling + validator + middleware tests (no DB, always runs)

**Wave 2** *(blocked on Wave 1 completion)*
- [ ] 06-02-PLAN.md — CSRF/mapping/utility + non-DB handler tests (no DB, always runs)

**Wave 3** *(blocked on Wave 2 completion)*
- [ ] 06-03-PLAN.md — Auth integration + DB query tests (TEST_POSTGRES_URL gated)

**Wave 4** *(blocked on Wave 3 completion)*
- [ ] 06-04-PLAN.md — DB handler tests + route integration + coverage gate (TEST_POSTGRES_URL gated)

### Phase 7: Handler Decomposition

**Goal**: Overgrown handlers are decomposed into thin HTTP adapters plus testable service functions
**Depends on**: Phase 6
**Requirements**: HAND-01, HAND-02, HAND-03, HAND-04, HAND-05
**Success Criteria** (what must be TRUE):

   1. `apiFinalizar` (~100 lines) is decomposed into a thin handler + service functions
   2. The next 3 largest handlers by line count are similarly decomposed
   3. Extracted service functions never touch `http.Request` or `http.ResponseWriter`
   4. Handlers are thin adapters (10-20 lines), delegating to service functions
   5. All existing tests pass before and after decomposition (no behavior changes)

**Plans**: 2 plans in 2 waves
Plans:
**Wave 1** *(no dependencies)*
- [ ] 07-01-PLAN.md — apiFinalizar decomposition: FinalizarResult/FlatBundle types, finalizeActivity service, thin handler, service tests (HAND-01)

**Wave 2** *(blocked on Wave 1 completion)*
- [ ] 07-02-PLAN.md — Remaining 3 handlers: updateUserAdmin shared service, adminUpdateUser/apiAdminUserUpdate thin adapters, bulkActivityDetails service, apiDashboardBulkDetails thin adapter, service tests (HAND-02)

### Phase 8: ES5 Compatibility

**Goal**: Scanning workflow frontend runs reliably on non-Chrome warehouse browsers
**Depends on**: Phase 5
**Requirements**: ES5-01, ES5-03, ES5-04
**Success Criteria** (what must be TRUE):

  1. Scanning workflow JS files use only ES5 syntax (no `const`/`let`, arrow functions, `async`/`await`, `fetch`, or template literals)
  2. HTMX version works on warehouse browsers (verified via device testing; falls back to 1.9.x if 2.x fails)
  3. Page weight and rendering is optimized for low-end devices (reduced payload, minimal reflows)

**Plans**: TBD
**UI hint**: yes

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | — | Complete | 2026-06-08 |
| 2. Activity Scanning | — | Complete | 2026-06-08 |
| 3. Dashboard | — | Complete | 2026-06-08 |
| 4. Admin | — | Complete | 2026-06-08 |
| 5. Error Handling Foundation | 4/4 | Complete | 2026-06-09 |
| 6. Testing Infrastructure | 0/4 | Planning | - |
| 7. Handler Decomposition | 0/2 | Planning | - |
| 8. ES5 Compatibility | 0/0 | Planning | - |
