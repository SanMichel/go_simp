# Requirements: go-simp

**Defined:** 2026-06-08
**Core Value:** Warehouse workers must be able to scan and register activities reliably without the system getting in their way.

## v1 Requirements

Requirements for v1.1 Simplify & Stabilize milestone.

### Error Handling

- [ ] **ERR-01**: App defines a custom `AppError` type with Code, Message, HTTPStatus, and wrapped Err fields
- [ ] **ERR-02**: Centralized `handleError()` dispatches to JSON or HTML error responses based on request type (HTMX vs API)
- [ ] **ERR-03**: All existing handlers return errors via the centralized handler instead of inline formatting
- [ ] **ERR-04**: Input validation is standardized via a reusable `Validator` type
- [ ] **ERR-05**: Error logging includes request ID, stack trace, and structured fields
- [ ] **ERR-06**: Panic recovery middleware catches panics before server crash
- [ ] **ERR-07**: `writeJSON` writes header after successful encode (not before), preventing silent 200-on-failure

### Testing

- [ ] **TEST-01**: All existing handlers have table-driven tests using `httptest`
- [ ] **TEST-02**: Auth middleware and session handling have unit tests
- [ ] **TEST-03**: Database query functions have tests (transactional or mocked)
- [ ] **TEST-04**: Error handling (`AppError`, `handleError`) has unit tests
- [ ] **TEST-05**: All existing routes return correct status codes and response shapes under test
- [ ] **TEST-06**: Test coverage reaches 70%+ of the codebase

### Handler Decomposition

- [ ] **HAND-01**: `apiFinalizar` (activity finalization, ~100 lines) is decomposed into handler + service functions
- [ ] **HAND-02**: Next 3 largest handlers by line count are decomposed
- [ ] **HAND-03**: Business logic extracted into service functions that never touch `http.Request`/`http.ResponseWriter`
- [ ] **HAND-04**: Handlers become thin adapters (10-20 lines), delegating to service functions
- [ ] **HAND-05**: No behavior changes during decomposition — test coverage proves equivalence
- [ ] **HAND-06**: Code organized into domain-grouped files (activity_handlers.go, dashboard_handlers.go, admin_handlers.go)

### ES5 Compatibility

- [ ] **ES5-01**: Scanning workflow JS files are rewritten to ES5 (no `const`/`let`, arrow functions, `async`/`await`, `fetch`, template literals)
- [ ] **ES5-02**: Inlined DOMPurify is replaced with existing `escHtml()` for ES5-safe sanitization
- [ ] **ES5-03**: HTMX version is verified compatible with warehouse browsers (fallback to 1.9.x if 2.x fails)
- [ ] **ES5-04**: Page weight/rendering optimized for low-end devices

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Device Compatibility

- **CAM-01**: Server-side barcode decoding endpoint using `gozxing`
- **CAM-02**: Canvas capture JS sends frames to decode endpoint
- **CAM-03**: Manual barcode input fallback
- **ES5-05**: Admin and dashboard JS files rewritten to ES5

### Architecture

- **HAND-07**: Remaining smaller handlers decomposed
- **HAND-08**: Service layer extracted into dedicated files

## Out of Scope

| Feature | Reason |
|---------|--------|
| Native mobile apps | Web-first; PWA scope not yet determined |
| Offline support | Too complex for current scope; revisit in v2 |
| Architectural split into sub-packages | Keep single `main` package; improve within |
| Third-party JS frameworks | stdlib + HTMX stays; no npm/polyfill deps |
| Camera barcode scanning (COMPAT-03) | Deferred to v2; needs real-device HTTPS validation first |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| ERR-01 | — | Pending |
| ERR-02 | — | Pending |
| ERR-03 | — | Pending |
| ERR-04 | — | Pending |
| ERR-05 | — | Pending |
| ERR-06 | — | Pending |
| ERR-07 | — | Pending |
| TEST-01 | — | Pending |
| TEST-02 | — | Pending |
| TEST-03 | — | Pending |
| TEST-04 | — | Pending |
| TEST-05 | — | Pending |
| TEST-06 | — | Pending |
| HAND-01 | — | Pending |
| HAND-02 | — | Pending |
| HAND-03 | — | Pending |
| HAND-04 | — | Pending |
| HAND-05 | — | Pending |
| HAND-06 | — | Pending |
| ES5-01 | — | Pending |
| ES5-02 | — | Pending |
| ES5-03 | — | Pending |
| ES5-04 | — | Pending |

**Coverage:**
- v1 requirements: 23 total
- Mapped to phases: 0
- Unmapped: 23 ⚠️

---
*Requirements defined: 2026-06-08*
*Last updated: 2026-06-08 after initial definition*
