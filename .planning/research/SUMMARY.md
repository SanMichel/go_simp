# Research Summary: go-simp v1.1 — Code Quality & Device Compatibility

**Domain:** Warehouse Activity Scanning Dashboard (Go + HTMX)
**Researched:** 2026-06-08
**Overall confidence:** HIGH

## Executive Summary

This milestone adds zero new user-facing features. All work is **infrastructure improvement**: handler decomposition, error handling standardization, comprehensive testing, ES5 browser compatibility, and camera barcode scanning for legacy warehouse devices.

The existing codebase has solid foundations (Go 1.23 stdlib, HTMX frontend, Postgres + Oracle backends, CSRF protection, RBAC). The two biggest problems are: (1) handler functions are overgrown (e.g., `apiFinalizar` at 101 lines mixing HTTP, business logic, and DB), and (2) ALL JavaScript uses ES6+ syntax (`async/await`, `fetch`, arrow functions, template literals) which won't run on the non-Chrome warehouse browsers.

Research confirms that both problems have well-established solutions within the single-`main`-package constraint. Handler decomposition follows the "thin adapter" pattern — extract business logic into service functions that never touch `http.Request`, leaving handlers as thin HTTP translators. Error handling standardizes via a custom `AppError` type with a centralized `handleError()` dispatcher that routes to JSON or HTML based on request type (HTMX vs API).

ES5 compatibility requires a complete rewrite of all JS files, replacing `fetch` with `XMLHttpRequest`, `async/await` with callbacks, arrow functions with `function(){}`, and template literals with concatenation. Camera barcode scanning should default to **server-side decoding** (capture frame → POST to Go endpoint → decode with `gozxing`) for maximum compatibility, with client-side scanning as a progressive enhancement.

## Key Findings

**Stack:** No new external frameworks. Add `gozxing` for server-side barcode decoding. Keep HTMX but verify 2.x ES5 compat (fallback to 1.x if needed). No npm/polyfill dependencies — write ES5 directly.

**Architecture:** Extract business logic into `services.go` (no HTTP types), validation into `validation.go` (`Validator` type), errors into `errors.go` (`AppError` + `handleError()`). Handlers become 10-20 line thin adapters. File organization groups by domain (activity_handlers.go, dashboard_handlers.go, admin_handlers.go).

**Critical pitfall:** ES6+ features leaking into ES5 rewrite. CI linting + PR checklist + real device testing are essential. Second critical pitfall: trying to decompose all handlers at once — do one per PR with tests first.

## Implications for Roadmap

Based on research, suggested phase structure:

1. **Foundation: Error handling + validation + file organization**
   - Addresses: CODE-02, CODE-03, CODE-04, MAINT-01
   - Creates `errors.go`, `validation.go`, groups files by domain
   - Avoids: Adding new code with inconsistent error patterns
   - Risk: LOW — purely additive, no behavior change

2. **Testing infrastructure**
   - Addresses: MAINT-02
   - Table-driven handler tests, middleware tests, service tests
   - Avoids: Decomposing handlers without a safety net
   - Risk: LOW — extends existing test patterns

3. **Handler decomposition**
   - Addresses: CODE-01
   - One handler at a time, starting with simple GET handlers
   - Avoids: Breaking existing behavior, massive PRs
   - Risk: MEDIUM — must not change behavior during refactor

4. **ES5 compatibility**
   - Addresses: COMPAT-01, COMPAT-02
   - Rewrite all JS to ES5, verify HTMX version, optimize page weight
   - Avoids: Browser breakage on warehouse devices
   - Risk: HIGH — largest scope, manual rewrite, testing dependency

5. **Camera barcode scanning**
   - Addresses: COMPAT-03
   - Add Go endpoint + ES5 capture JS + manual fallback
   - Avoids: Client-side library bloat
   - Risk: MEDIUM — new feature, depends on ES5 rewrite

**Phase ordering rationale:**
Phases 1-3 can be done in parallel with Phase 4 (different files). However, doing error handling first means decomposed handlers will use the new error patterns from the start, avoiding a second pass. Testing before decomposition means each extract is verified. ES5 rewrite is the riskiest item and benefits from the safety net of tests. Camera scanning depends on having ES5 secure JS patterns established first.

**Research flags for phases:**
- Phase 2: Deeper research needed on mocking strategy within single `main` package (function variable assignment vs interface + struct embedding)
- Phase 4: Must test HTMX 2.x on actual warehouse devices — can't guarantee compatibility from desktop research alone
- Phase 5: Camera API availability on warehouse browsers needs real-device validation; fallback to manual input is critical

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Handler decomposition patterns | HIGH | Well-documented "thin handler/adapter" pattern in Go stdlib |
| Error handling standardization | HIGH | `AppError` + sentinel errors + centralized handler is standard Go practice |
| Go web testing with httptest | HIGH | Existing tests already use this pattern, well-documented |
| ES5 browser compatibility | MEDIUM | ECMAScript 5 spec is frozen and documented; risk is in execution (rewriting correctly + HTMX compat) |
| Camera barcode scanning | HIGH | gozxing is well-maintained; capture-to-canvas-to-server is a proven pattern |
| File organization in single main package | HIGH | Well-established convention in Go community |

## Gaps to Address

- **HTMX version ES5 compatibility:** Need to verify HTMX 2.0.4 works on actual warehouse browsers. If not, HTMX 1.9.x is the fallback. Desktop research can't confirm this — requires device testing.
- **Camera API availability on specific warehouse devices:** The exact make/model/browser of warehouse devices is not documented. Assume `getUserMedia` with webkit prefix + XMLHttpRequest is the floor, bare manual input is the basement.
- **Mocking strategy details:** The best approach for mocking within a single `main` package needs decision — function variable assignment is simplest but package-level state, interface + struct embedding is cleaner but more code.
- **gozxing image format support:** Need to verify what input formats gozxing accepts (JPEG? PNG? raw pixel data?) and adapt the canvas capture accordingly.
