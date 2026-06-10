---
phase: 08-es5-compatibility
plan: 01
subsystem: ui
tags: [es5, xhr, scanning, warehouse, templates, csrf]

# Dependency graph
requires:
  - phase: 07-handler-decomposition
    provides: domain handler files (activity_handlers.go, api_handlers.go, etc.)

provides:
  - Standalone atividades frontend in cmd/server/templates/atividades/
  - ES5-compatible JS utilities (XHR wrapper, sanitize, date, audio)
  - ES5 login handler (session check, form submit, redirect)
  - ES5 scanning state machine (load, scan, building switch, finalize, report)
  - ES5 product consultation (code search, description search)
  - COPIED traceability comments in shared.js for future cleanup

affects:
  - 08-02 (Go route/handler updates for /atividades/login and HTMX)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - ES5 XHR-based API calls (XMLHttpRequest, onreadystatechange before open)
    - sanitizeHtml wrapper for all innerHTML assignments (XSS mitigation)
    - CSRF token extraction from cookie for mutating requests
    - onUnauthorized callback redirecting to /atividades/login
    - state management via localStorage with saveState/resetActivityState

key-files:
  created:
    - cmd/server/templates/atividades/atividades-login.html
    - cmd/server/templates/atividades/atividades.html
    - cmd/server/templates/atividades/atividades-utils.js
    - cmd/server/templates/atividades/atividades-login.js
    - cmd/server/templates/atividades/atividades-scan.js
    - cmd/server/templates/atividades/atividades-consulta.js
  modified:
    - cmd/server/templates/shared.js

key-decisions:
  - "D-01: atividades split into standalone templates/atividades/ directory"
  - "D-03: Two HTML templates (atividades-login.html, atividades.html)"
  - "D-04: Four JS files (login, utils, scan, consulta)"
  - "D-05: Utilities copied from shared.js into atividades-utils.js with COPIED comments"
  - "D-08: Existing login.js left unchanged"
  - "D-09: All 4 JS files written directly in ES5 — no Babel, no build step"
  - "D-10: ES5 constraints enforced: var, function(){}, XHR, string concat, index-based loops"
  - "D-16: No hard size target — clean ES5 is the goal"
  - "D-17: No minification — readability prioritized over wire size"
  - "D-18: No CSS changes — existing stylesheets reused"

patterns-established:
  - "ES5 XHR apiCall wrapper: onreadystatechange before open, withCredentials=true, CSRF via cookie extraction"
  - "sanitizeHtml wrapper for all innerHTML assignments (XSS mitigation per threat model T-08-01)"
  - "onUnauthorized callback pattern: all XHR calls redirect to /atividades/login on 401"
  - "state persistence: localStorage with saveState/resetActivityState"

requirements-completed: [ES5-01]

# Metrics
duration: 3min
completed: 2026-06-10
---

# Phase 08 Plan 01: Standalone Atividades Frontend Summary

**ES5-compatible standalone atividades frontend: 2 HTML templates + 4 XHR-based JS files (1000+ lines), ported from monolithic app.js/shared.js to templates/atividades/ without any ES6 features**

## Performance

- **Duration:** 3 min
- **Started:** 2026-06-10T16:48:00-03:00
- **Completed:** 2026-06-10T16:51:35-03:00
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments

- Created `templates/atividades/` directory with 2 HTML templates: standalone login page (`atividades-login.html`) and main SPA (`atividades.html`) with 5 screens (start, scanning, predio-switch, consulta, report) and product-detail modal
- Created `atividades-utils.js` with ES5 XHR wrapper (`apiCall`/`apiGet`/`apiPost`) including CSRF token extraction, plus ported utilities (`showLoader`, `formatDate`, `playBeep`, `escHtml`, `sanitizeHtml`)
- Created `atividades-login.js` with session check on load and XHR-based login form submission to `/api/auth/login`
- Created `atividades-scan.js` (850 lines) — full ES5 port of the scanning workflow state machine: state management, screen navigation, EAN scan flow with divergence detection, building switch with expected product dedup, finalize with report rendering
- Created `atividades-consulta.js` — ES5 product consultation with code search (`/api/produtos/consulta/:ean`) and description search (`/api/produtos/consulta?q=`)
- All 4 JS files are pure ES5: zero `const`, `let`, `=>`, `` ` ``, `async`/`await`, `fetch`, or `?.`
- All innerHTML assignments use `sanitizeHtml()` wrapper per threat model T-08-01 (22+9=31 uses across scan.js + consulta.js)
- Added 6 `COPIED` traceability comments in `shared.js` marking functions ported to `atividades-utils.js`

## Task Commits

Each task was committed atomically:

1. **Task 1: Create HTML templates** — `eb2e8bf` (feat)
2. **Task 2: Create utils.js and login.js** — `e9c7f3a` (feat)
3. **Task 3: Create scan.js, consulta.js, update shared.js** — `84a3418` (feat)

## Files Created/Modified

- `cmd/server/templates/atividades/atividades-login.html` — Standalone login template `{{define "atividades-login"}}`, form id="form-atividades-login", scripts for utils+login
- `cmd/server/templates/atividades/atividades.html` — Main SPA template `{{define "atividades"}}`, 5 screens, product modal, 4 scripts (utils, scan, consulta, htmx)
- `cmd/server/templates/atividades/atividades-utils.js` — 8 ES5 functions: apiCall, apiGet, apiPost, showLoader, formatDate, playBeep, escHtml, sanitizeHtml
- `cmd/server/templates/atividades/atividades-login.js` — Session check + login form handler via XHR
- `cmd/server/templates/atividades/atividades-scan.js` — Full scanning state machine (850 lines ES5)
- `cmd/server/templates/atividades/atividades-consulta.js` — Product consultation (code + description search)
- `cmd/server/templates/shared.js` — 6 COPIED comments added for traceability (no functional changes)

## Decisions Made

None — plan executed exactly as written. All design decisions (D-01 through D-18) were established in the research/patterns phase and followed during execution.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## Security Compliance

All threat model mitigations validated:
- **T-08-01 (XSS):** Every innerHTML assignment in scan.js (22 uses) and consulta.js (9 uses) wraps content in `sanitizeHtml()`. Login.js uses `innerText` exclusively.
- **T-08-02 (CSRF):** XHR apiCall reads `csrf_token` cookie and sets `X-CSRF-Token` header on all POST/PATCH/DELETE requests.
- **T-08-03 (Info Disclosure):** Session token is HttpOnly (not accessible from JS). XHR uses `withCredentials=true` for cookie-based auth.
- **T-08-04 (Repudiation):** Every API call has onSuccess and onUnauthorized callbacks. `showLoader(false)` called in both success and failure paths.

## Stub Tracking

No stubs found — all JS files contain complete implementations with real API calls and data flow.

## Threat Flags

None — all security-relevant surface was planned and mitigated.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

Plan 08-02 (Go handler changes) can proceed: it needs to add `/atividades/login` route in `handlers.go`, register template parses and routes in `main.go`, vendor `htmx.min.js`, and update tests. The frontend files created here are ready and self-contained.

## Self-Check: PASSED

- [x] All 6 files exist
- [x] All 3 commits verified in git log
- [x] SUMMARY.md content verified
- [x] All success criteria met

---
*Phase: 08-es5-compatibility*
*Completed: 2026-06-10*
