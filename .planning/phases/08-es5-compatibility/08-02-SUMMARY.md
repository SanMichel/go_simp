---
phase: 08-es5-compatibility
plan: 02
subsystem: api
tags: [go, backend, middleware, routes, tests, atividades]
requires:
  - phase: 08-01
    provides: JS files and templates in templates/atividades/
provides:
  - requireAtividadesRole middleware in auth.go
  - atividadesLoginPage handler in handlers.go
  - Updated go:embed, ParseFS, and routes in main.go
  - 3 handler tests in main_test.go
affects: [08-03, 08-04]

tech-stack:
  added: []
  patterns:
    - requireAtividadesRole middleware mirrors requireRole but with /atividades/login redirect
    - atividadesLoginPage handler checks auth first, renders template or redirects
    - Test pattern: unauthenticated test uses App{tpl: parseTemplates()} (no DB needed)

key-files:
  created: []
  modified:
    - cmd/server/auth.go
    - cmd/server/handlers.go
    - cmd/server/main.go
    - cmd/server/main_test.go

key-decisions:
  - "requireAtividadesRole created as separate middleware (not modification of requireRole) to keep existing /login redirect for admin/dashboard routes unchanged"
  - "atividadesLoginPage uses redirectByRole instead of hardcoded /atividades redirect for consistent role-based behavior"
  - "Unauthenticated tests use parseTemplates() directly without testApp/testDB keeping them fast and dependency-free"

patterns-established:
  - "New middleware pattern: copy requireRole, change redirect URL"
  - "New route pattern: add JS file routes, page routes, and modified middleware reference"
  - "Test pattern for middleware isolation: requireAtividadesRole wrapped around no-op handler"

requirements-completed: [ES5-01, ES5-03, ES5-04]

duration: 1min
completed: 2026-06-10
---

# Phase 8 Plan 02: Backend Integration Summary

**Go backend wiring for atividades frontend: middleware, handler, route registration, and tests**

## Performance

- **Duration:** 1 min
- **Started:** 2026-06-10T19:54:16Z
- **Completed:** 2026-06-10T19:56:03Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Added `requireAtividadesRole` middleware in auth.go — identical to `requireRole` but redirects unauthenticated users to `/atividades/login` instead of `/login`
- Added `atividadesLoginPage` handler in handlers.go — checks auth, redirects by role if authenticated, renders `"atividades-login"` template if not
- Updated main.go: go:embed now includes `templates/atividades/*.html` and `templates/atividades/*.js`; ParseFS includes `"templates/atividades/*.html"`; 4 new JS file routes registered; `/atividades/login` route registered; `/atividades` route switched to `requireAtividadesRole`
- Added 3 handler tests: unauthenticated login page (200/html, no DB), authenticated redirect (302, DB-gated), middleware redirect check (302 + Location header, no DB)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add middleware and handler** - `a5d130b` (feat)
2. **Task 2: Update main.go** - `c4ce0e6` (feat)
3. **Task 3: Add handler tests** - `161489b` (test)

**Plan metadata:** *(committed after SUMMARY.md)*

## Files Created/Modified

- `cmd/server/auth.go` — Added `requireAtividadesRole` middleware (lines after `requireRole`)
- `cmd/server/handlers.go` — Added `atividadesLoginPage` handler (after `atividadesPage`)
- `cmd/server/main.go` — Updated go:embed directive, ParseFS patterns, and route registration (5 new routes, 1 modified)
- `cmd/server/main_test.go` — Added 3 test functions after `TestAtividadesPageAuthenticated`

## Decisions Made

- Created `requireAtividadesRole` as a separate function rather than modifying `requireRole` to accept a redirect parameter — keeps existing `/login` redirect for admin/dashboard routes unchanged and avoids refactoring all call sites
- `atividadesLoginPage` uses `redirectByRole` instead of hardcoded `/atividades` redirect — provides correct role-based routing (conferente→/atividades, gerente→/dashboard, sysadmin→/admin)
- Non-DB tests use `&App{tpl: parseTemplates()}` directly — no test DB dependency, fast execution

## Verification Results

| Check | Result |
|-------|--------|
| `go vet ./cmd/server` | ✅ PASS |
| `requireAtividadesRole` redirects to `/atividades/login` | ✅ PASS (grep + test) |
| `atividadesLoginPage` renders "atividades-login" template | ✅ PASS (test) |
| go:embed includes `templates/atividades/*.html` and `templates/atividades/*.js` | ✅ PASS (grep) |
| ParseFS includes `"templates/atividades/*.html"` | ✅ PASS (grep) |
| 5 new routes (4 JS + login page) | ✅ PASS (grep) |
| `/atividades` uses `requireAtividadesRole` | ✅ PASS (grep) |
| TestAtividadesLoginPage_Unauthenticated | ✅ PASS |
| TestAtividadesLoginPage_AuthenticatedRedirect | ✅ SKIP (no DB — expected) |
| TestAtividadesPage_UnauthenticatedRedirect | ✅ PASS |

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None — all new handler and middleware code is fully wired and tested.

## Threat Flags

No new security-relevant surface introduced beyond what the plan's threat model covers.

## Issues Encountered

None

## Next Phase Readiness

Ready for Plan 08-03 (main atividades SPA template and JS files — the core ES5 rewrite of the scanning workflow).

---

*Phase: 08-es5-compatibility*
*Completed: 2026-06-10*
