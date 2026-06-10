# Phase 8: ES5 Compatibility — Context

**Gathered:** 2026-06-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Rewrite the scanning workflow (`/atividades` page) into standalone ES5-compatible JS files, verify HTMX compatibility on warehouse browsers, and optimize page weight for low-end devices. The focus is on the scanning workflow only — admin and dashboard JS are deferred to v2 (ES5-05).

Requires: ES5-01, ES5-03, ES5-04 from REQUIREMENTS.md. ES5-02 (DOMPurify→escHtml) completed in Phase 5.

</domain>

<decisions>
## Implementation Decisions

### JS Rewrite Scope — Standalone Atividades Frontend
- **D-01:** Split `/atividades` into a standalone directory `templates/atividades/` with its own HTML templates, JS files, and login page. The existing `app.js`, `shared.js`, `login.js`, `dashboard.js`, `admin.js` stay untouched.
- **D-02:** New Go route `/atividades/login` renders a separate login template. Unauthenticated users hitting `/atividades` are redirected to `/atividades/login` instead of `/login`.
- **D-03:** Two HTML templates in `templates/atividades/`: `atividades-login.html` (login page) and `atividades.html` (main scanning SPA). Both are multiple Go-rendered templates (not JS-driven screens within a single template).
- **D-04:** Four JS files in `templates/atividades/`: `atividades-login.js` (login form logic), `atividades-scan.js` (scanning workflow), `atividades-consulta.js` (product query), `atividades-utils.js` (shared utilities).
- **D-05:** Copy necessary utilities from `shared.js` (apiCall/XHR equivalent, escHtml, showLoader, formatDate, playBeep) into `atividades-utils.js`. Comment out the copied functions in `shared.js` with a TODO for cleanup in a future phase.
- **D-06:** New Go handler for `/atividades/login`. Modify existing `/atividades` handler to redirect to `/atividades/login` instead of `/login` when unauthenticated.
- **D-07:** Update `go:embed` pattern in `main.go` to include `templates/atividades/*.html` and `templates/atividades/*.js`.
- **D-08:** The existing `login.js` (global login page) stays unchanged — it serves admin/dashboard users who still use the original `/login` route.

### Rewrite Approach — Direct ES5
- **D-09:** Write the 4 new JS files directly in ES5. No Babel, no build step, no transpilation. Manual conversion from the existing modern-JS patterns.
- **D-10:** ES5 constraints for the new files: `var` only (no `const`/`let`), `function(){}` only (no arrow functions), no `async`/`await` (promises or callbacks), no `fetch` (XMLHttpRequest instead), no template literals (string concatenation), no optional chaining (`&&` guards), no `for...of` (index-based loops), no default parameters, no spread operator, no `Array.includes()`/`Array.find()`/`Array.some()` where IE11 would be used (use index-based checks).

### Device Testing Strategy
- **D-11:** General compatibility guideline — code must work on any browser shipped on warehouse devices. No specific device list decided.
- **D-12:** Real warehouse devices are available for testing. Coordinate with operations.
- **D-13:** Keep the current vendored HTMX version (global `htmx.min.js`). Test on real devices. If incompatible, swap the file for 1.9.x.
- **D-14:** The new atividades page reuses the same HTMX-driven pattern (server-rendered HTML partials, hx-get/hx-post attributes). The ES5 rewrite focuses only on JS syntax, not replacing HTMX.
- **D-15:** Testing approach: iterative on-device testing. Ship the new page, test on warehouse devices, fix issues as they surface. If HTMX fails → downgrade to 1.9.x; if that also fails → revisit HTMX removal.

### Page Weight Optimization
- **D-16:** Both JS size AND render performance matter — but no hard size target. The goal is clean, concise ES5 code.
- **D-17:** No minification of the JS files. Readability for maintenance is prioritized over wire size (the files are small enough).
- **D-18:** No CSS changes or optimizations. The existing stylesheets cover the layout needs.

### the agent's Discretion
- JS file organization within `templates/atividades/` (exact exports, internal function naming)
- XMLHttpRequest wrapper design for `atividades-utils.js`
- Go handler implementation details for the new `/atividades/login` route
- Template structure details for `atividades-login.html` and `atividades.html`
- ES5 fallback patterns for browser features that may or may not be available
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §ES5-01, ES5-03, ES5-04 — Core requirements for this phase
- `.planning/REQUIREMENTS.md` §ES5-05 — Deferred: admin/dashboard JS rewrite (v2)

### Existing Code — JS files to replace
- `cmd/server/templates/app.js` — Current scanning SPA (938 lines). Full replacement with split atividades JS files.
- `cmd/server/templates/shared.js` — Shared utilities. Copy apiCall, escHtml, showLoader, formatDate, playBeep into atividades-utils.js.
- `cmd/server/templates/login.js` — Login logic. Replaced by atividades-login.js for the atividades path.

### Existing Code — Go routes to modify
- `cmd/server/main.go` — Route registration. Add `/atividades/login`, update `/atividades` redirect logic.
- `cmd/server/handlers.go` — Existing handler that serves `/atividades`. Modify redirect target.
- `cmd/server/templates/htmx.min.js` — Vendored HTMX. Keep current version, test on devices, swap to 1.9.x if needed.

### Existing Code — Integration points
- `cmd/server/main.go` line `go:embed` — Must add `templates/atividades/*.html` and `templates/atividades/*.js` to the embed pattern.
- `cmd/server/main.go` line serveJS pattern — Follow existing pattern for registering the new JS files as static routes.

### Project Constraints
- `.planning/PROJECT.md` §Constraints — No npm/polyfill deps, single main package, stdlib + HTMX only.
- `.planning/PROJECT.md` §Out of Scope — No third-party JS frameworks, no camera barcode scanning.

### Prior Phase Context
- `.planning/STATE.md` — ES5 rewrite flagged as highest-risk item. HTMX 2.x compatibility unverified on warehouse devices.
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cmd/server/templates/shared.js` — Functions to copy: `apiCall` (convert to XHR), `escHtml`, `showLoader`, `formatDate`, `playBeep`. These are well-isolated and can be ported directly.
- `cmd/server/templates/app.js` — The state management pattern (localStorage-based state, saveState/resetActivityState) can be adapted. The business logic (scan flow, divergence detection, building switch) is sound — only the syntax needs ES5 conversion.
- Cookie-based CSRF token pattern from shared.js `apiCall` — reuse the existing token extraction logic in the new XHR wrapper.

### Established Patterns
- **JS pattern:** var-based variables, function-based code organization, localStorage state persistence, DOM manipulation via innerHTML/document.getElementById.
- **Route pattern:** Go 1.22 ServeMux with method-based routing. New routes follow `mux.HandleFunc("GET /atividades/login", ...)` convention.
- **Template pattern:** Go html/template with define/end blocks, embedded via go:embed. New templates follow the existing naming conventions.

### Integration Points
- New `templates/atividades/` directory needs to be added to the `go:embed` directive in `main.go`.
- New `/atividades/login` route and handler added alongside existing page routes.
- Existing `/atividades` handler modified to redirect to `/atividades/login` when unauthenticated.
- All JS files in `templates/atividades/` need static file routes registered in `main.go` (follow the `app.serveJS("filename.js")` pattern).
- Backend API endpoints (`/api/empresas`, `/api/locais`, `/api/produtos/*`, `/api/atividades/*`, `/api/auth/*`) remain unchanged — shared with admin/dashboard.
</code_context>

<specifics>
## Specific Ideas

- New atividades JS files follow the same localStorage-based state persistence pattern as the current app.js, adapted for ES5.
- The state machine (screens: login → start → scanning → predio-switch → report → start) stays the same.
- Product scan flow (EAN lookup → address validation → divergence detection → building switch) remains unchanged — only the syntax changes.
</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.
</deferred>

---

*Phase: 8-ES5 Compatibility*
*Context gathered: 2026-06-10*
