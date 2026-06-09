---
phase: 6
slug: testing-infrastructure
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-09
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` (stdlib) |
| **Config file** | `go.mod` — dependencies already managed |
| **Quick run command** | `go test -count=1 ./cmd/server` |
| **Full suite command** | `go test -count=1 -v -cover ./cmd/server` |
| **Coverage report** | `go test -coverprofile=c.out -count=1 ./cmd/server && go tool cover -func=c.out` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test -count=1 ./cmd/server`
- **After every plan wave:** Run `go test -cover -count=1 -v ./cmd/server`
- **Before `/gsd-verify-work`:** Full suite green and coverage ≥ 70%
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | TEST-04 | T-06-01 / — | AppError implements error interface; no information disclosure | unit | `go test -run TestAppError ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | TEST-04 | T-06-02 | Validator accumulates errors, chaining works, combined message | unit | `go test -run TestValidator ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-01-03 | 01 | 1 | TEST-04 | T-06-03 | writeJSON encodes before writing header, falls back to 500 | unit | `go test -run TestWriteJSON ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-01-04 | 01 | 1 | TEST-04 | T-06-04 | recoveryMiddleware catches panic, returns 500 | unit | `go test -run TestRecoveryMiddleware ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-02-01 | 02 | 1 | TEST-02 | T-06-05, T-06-06 | Auth middleware: CSRF, requireRole, requireAPIRole, currentUser | integration | `go test -run 'TestRequire|TestCSRF|TestCurrentUser' ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 1 | TEST-02 | T-06-07 | Session: makeToken, randomString, revokeSession | unit | `go test -run 'TestMakeToken|TestRandomString' ./cmd/server -v -count=1` | ✅ (partial) | ⬜ pending |
| 06-03-01 | 03 | 2 | TEST-01 | T-06-08 | Handler tests: non-DB routes (10 routes) | unit | `go test -run TestHandlerNoDB ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-04-01 | 04 | 2 | TEST-01 | T-06-09 | Handler tests: DB-dependent routes (36 routes) | integration | `go test -run TestHandler ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-05-01 | 05 | 2 | TEST-05 | T-06-10 | Route response shape validation | integration | `go test -run TestRoutes ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |
| 06-05-02 | 05 | 2 | TEST-03 | T-06-11 | DB query function tests | integration | `go test -run 'TestFindUser|TestListActivities' ./cmd/server -v -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `cmd/server/testhelper.go` — test DB factory, test user factory, test token factory, cleanup helpers
- [ ] Expanded tests in `cmd/server/main_test.go`:
  - AppError: Error(), Unwrap(), nil Unwrap
  - handleError: non-AppError path, HTMX path, API path, page path, default status from code
  - Validator: all methods, chaining, Error() combined message
  - writeJSON: success path, encode failure path
  - recoveryMiddleware: catches panic
  - requestIDMiddleware: generates ID, preserves existing header
  - CSRF middleware: origin check, non-API POST, PATCH/DELETE
  - Mapping functions: mapActivity, mapProduct, mapUser, mapOracleProduct
  - Utility functions: parseFilters, intQuery
  - Auth: requireRole/requireAPIRole (unauthenticated, forbidden role, allowed role)
  - Handler table-driven tests (36 DB-dependent, 10 non-DB)
  - Route integration test

*Existing `cmd/server/main_test.go` already has 18 tests with 15.3% baseline coverage.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Oracle-dependent code paths | TEST-03 | No Oracle available in test environment; read-only SQL guard already tested | Run against staging Oracle; check isReadOnlySQL rejects non-SELECT/WITH queries |
| HTMX partial rendering | TEST-05 | HTML template rendering is tested via httptest; visual correctness requires manual review | Load each page in browser, check HTMX interactions render correctly |

*All other phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
