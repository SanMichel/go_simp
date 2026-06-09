---
phase: 05-error-handling-foundation
plan: 04
subsystem: core
tags: [refactor, security, handlers, dompurify, shared-js]
dependency-graph:
  requires: [05-02]
  provides: [HAND-06, ES5-02]
  affects: [cmd/server, cmd/server/templates]
tech-stack:
  added: []
  patterns: [domain-file-per-handler-group]
key-files:
  created:
    - cmd/server/activity_handlers.go
    - cmd/server/dashboard_handlers.go
    - cmd/server/admin_handlers.go
  modified:
    - cmd/server/handlers.go
    - cmd/server/templates/shared.js
    - .gitignore
key-decisions:
  - "Dropped unused imports from handlers.go (slog, context, database/sql) — Go rejects unused imports, compilation requirement"
  - "Dropped unused imports from activity_handlers.go (strings, log/slog) — same reason"
  - "sanitizeHtml wrapper preserved as API compatibility layer; all 29 callers in app.js, dashboard.js, admin.js unchanged"
metrics:
  duration: 18min
  completed: 2026-06-09
  handlers.go_lines: 116 (was 535)
  shared.js_lines: 94 (was 1159)
  dopurify_lines_removed: 1065
---

# Phase 05 Plan 04: Split handlers + Remove DOMPurify Summary

Split monolithic 535-line handlers.go into domain-grouped files (activity, dashboard, admin) and removed 1065-line vendored DOMPurify library from shared.js in favor of escHtml(). Mechanical refactor — zero behavior change.

## Tasks

### Task 1: Split handlers.go into domain files (HAND-06)

- Created `cmd/server/activity_handlers.go` (247 lines): scanning/activity endpoints — apiEmpresas, apiLocais, apiProdutoEAN, apiProdutoConsulta, apiProdutoConsultaDescricao, apiProdutosLocal, apiLastInfo, apiFinalizar
- Created `cmd/server/dashboard_handlers.go` (77 lines): dashboard page/partial handlers — dashboardPage, dashboardTable, activityDetails, printOne, printBulk, printActivities
- Created `cmd/server/admin_handlers.go` (116 lines): admin CRUD handlers — adminPage, adminUsersSection, adminCreateUser, adminEditUserRow, adminUserRow, adminUpdateUser
- Trimmed `cmd/server/handlers.go` (116 lines, was 535): kept only page/entrypoint handlers — home, loginPage, loginPost, healthCheck, logout, atividadesPage, apiMe, apiLogin, apiLogout
- **Deviation:** Dropped `"log/slog"`, `"context"`, `"database/sql"` from handlers.go imports — Go rejects unused imports. The remaining 9 handlers don't use these packages directly.
- **Deviation:** Dropped `"strings"`, `"log/slog"` from activity_handlers.go imports — same reason.
- **Deviation (Rule 3):** Fixed `.gitignore` pattern `server` → `/server`. The bare `server` pattern glob-matched `cmd/server/*.go`, blocking the new files from being tracked. Changed to `/server` which matches only the root-level binary.

### Task 2: Replace DOMPurify with escHtml (ES5-02)

- Removed 1065 lines of vendored DOMPurify (lines 42–1106 in original shared.js)
- Changed `sanitizeHtml(dirty)` from `return purify.sanitize(dirty)` to `return escHtml(dirty)`
- All 29 existing callers in app.js, dashboard.js, admin.js continue working unchanged
- escHtml function preserved at its original location

## Verification

| Check | Result |
|-------|--------|
| `go build ./cmd/server` | PASS |
| `go test ./cmd/server -count=1` | PASS |
| `wc -l cmd/server/handlers.go` | 116 (≤200 ✓) |
| `grep -c 'func.*apiFinalizar' activity_handlers.go` | 1 ✓ |
| `grep -c 'func.*dashboardPage' dashboard_handlers.go` | 1 ✓ |
| `grep -c 'func.*adminCreateUser' admin_handlers.go` | 1 ✓ |
| `grep -c 'DOMPurify\|purify' shared.js` | 0 ✓ |
| `grep 'sanitizeHtml' shared.js` | `return escHtml(dirty)` ✓ |
| `grep -c 'function escHtml' shared.js` | 1 ✓ |
| `wc -l cmd/server/templates/shared.js` | 94 (≤110 ✓) |

## Deviations from Plan

### Rule 3 — Auto-fix blocking issue: .gitignore pattern blocks new files

- **Found during:** Task 1 (commit staging)
- **Issue:** `.gitignore` contained bare pattern `server` which glob-matches `cmd/server/*.go`, preventing `git add` from seeing the three new handler files
- **Fix:** Changed `server` → `/server` (anchored to root directory, matches only the compiled binary at repo root)
- **Files modified:** `.gitignore`
- **Commit:** dffc1e7

### Import trim deviations

- **handlers.go:** Removed `"log/slog"`, `"context"`, `"database/sql"` from imports — no remaining function in handlers.go uses these. Plan specified keeping them but Go rejects unused imports.
- **activity_handlers.go:** Removed `"strings"`, `"log/slog"` — not used by any moved function. Plan specified including them but unused.
- These are compilation requirements, not design decisions.

## Known Stubs

None.

## Threat Flags

None — mechanical file split and JS-only change. No new trust boundaries, network endpoints, or auth paths introduced.

## Self-Check: PASSED

- [x] handlers.go trimmed to 116 lines (was 535)
- [x] activity_handlers.go exists (247 lines, contains apiFinalizar)
- [x] dashboard_handlers.go exists (77 lines, contains dashboardPage)
- [x] admin_handlers.go exists (116 lines, contains adminCreateUser)
- [x] shared.js DOMPurify-free (94 lines, sanitizeHtml → escHtml)
- [x] Build passes (`go build ./cmd/server`)
- [x] Tests pass (`go test ./cmd/server -count=1`)
- [x] Commit dffc1e7 exists (Task 1 — handlers split)
- [x] Commit 00aaa85 exists (Task 2 — DOMPurify removal)
