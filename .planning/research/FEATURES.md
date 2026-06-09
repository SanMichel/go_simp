# Feature Landscape

**Domain:** Go single-package warehouse app — code quality + device compatibility features
**Researched:** 2026-06-08

## Table Stakes (Existing, Being Improved)

These features exist but need code quality improvements for the current milestone.

| Feature | Why Expected | Current Status | Complexity of Improvement |
|---------|--------------|---------------|-------------------------|
| Handler decomposition | Overgrown handlers are hard to understand and test | handlers.go (535 lines) + api_handlers.go (369 lines), mixed concerns | Low (mechanical extraction) |
| Consistent error handling | Users see different error formats; devs debug inconsistent logs | 3+ competing patterns (log+http.Error, log+render, writeJSON only) | Low (appError type + handleError adapter) |
| Input validation | Form/JSON endpoints validate inline, no shared validation | Validates inline in each handler, duplicate patterns | Low (centralize in validate.go) |
| File organization | Single main package has all code in 7 files | Growing but still navigable at ~2100 lines total | Low (domain-based file extraction) |
| Test coverage | No handler or integration tests exist | 323 lines testing utilities only, ~15 test functions | Medium (depends on refactoring for testability) |
| Panic recovery | No recovery middleware — any panic crashes the server | No recovery in middleware chain | Very Low (10 lines, stdlib) |
| Error response format | JSON error keys are inconsistent ("error" vs inline strings) | Varies between `map[string]string{"error": msg}` and inline template data | Low (centralize in response.go) |

## Differentiators (New for This Milestone)

| Feature | Value Proposition | Complexity | Notes |
|---------|------------------|------------|-------|
| ES5-compatible frontend | Warehouse workers on non-Chrome browsers can use the system | Medium | JS conversion is mechanical but covers ~3400 lines across 5 files |
| Reduced page weight | Faster load on low-end warehouse devices | Low | Remove inlined DOMPurify (~800 lines), minify, reduce redundant JS |
| Camera-based barcode scanning | Workers can use device camera instead of dedicated scanner | **NOT RESEARCHED** | Depends on browser camera API support (getUserMedia) on legacy devices |
| Standardized middleware chain | Predictable request processing pipeline | Low | Add recovery, document ordering, ensure consistency |
| Comprehensive test coverage (70%+) | Safe refactoring, documented behavior | Medium | Requires handler refactoring first, then test writing |

## Anti-Features

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Splitting `package main` into subpackages | Project constraint. Adds complexity without clear benefit for this application size. | Extract into same-package files organized by domain (admin_handlers.go, etc.) |
| Adding an external web framework (chi, gorilla) | Stdlib works well; Go 1.22 enhanced ServeMux covers routing needs. No dependency churn. | Keep net/http. Only add if specific middleware capability is needed. |
| Client-side SPA framework (React, Vue, Alpine) | HTMX + server-rendered HTML is the existing model and works well for this use case. Adding a JS framework would increase complexity and reduce ES5 compatibility. | Keep HTMX + server-rendered templates. |
| Full DOMPurify library | Inlined version is 800 lines of ES6+ code. The app only needs basic HTML escaping for product descriptions (user data from Oracle DB). | Use existing `escHtml()` function for all sanitization. |

## Feature Dependencies

```
CODE-02 (Error handling) → CODE-03 (JSON response patterns)
    ↓
CODE-01 (Handler decomposition) → CODE-04 (Input validation)
    ↓
MAINT-03 (Middleware chain) → MAINT-02 (Test coverage)
    ↓
COMPAT-01 (ES5) → COMPAT-02 (Page weight) → COMPAT-03 (Camera scanning)
```

- **CODE-02 → CODE-03:** Standardized error types enable unified JSON/HTML response patterns
- **CODE-01 → CODE-04:** Decomposing handlers makes validation extraction natural
- **MAINT-03 → MAINT-02:** Test coverage depends on stable, testable handler interfaces
- **COMPAT-01 → COMPAT-02:** JS conversion and DOMPurify removal both reduce page weight
- **COMPAT-02 → COMPAT-03:** Camera scanning is a JS-level feature that needs ES5 baseline first

## MVP Recommendation

**Priority for this milestone:**

1. **Foundation (Phase 1-2):** errors.go, config.go, middleware.go, domain handler extraction
   - Quick wins with no behavior change risk
   - Enables everything downstream

2. **Service extraction (Phase 3):** queries.go, validate.go, service functions
   - Reduces duplication between page and API handlers
   - Makes business logic testable

3. **Test coverage (Phase 4):** handlers_test.go, middleware_test.go
   - Build safety net before larger changes
   - Target current test functions + new handler tests

4. **ES5 compatibility (Phase 5):** JS conversion or transpilation
   - Separate concern from code quality
   - Needs on-device verification

**Defer:**
- **Camera barcode scanning (COMPAT-03):** Needs dedicated feasibility research on legacy browser getUserMedia support. Defer to a sub-milestone after ES5 baseline is confirmed.

## Sources

- Codebase inspection: handlers.go (535 lines), api_handlers.go (369 lines), main_test.go (323 lines)
- Go testing conventions: https://go.dev/blog/subtests
- Project constraints from .planning/PROJECT.md: single main package, no external frameworks
