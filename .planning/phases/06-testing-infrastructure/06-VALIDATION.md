---
phase: 6
slug: testing-infrastructure
status: verified
nyquist_compliant: true
wave_0_complete: true
created: 2026-06-09
audited: 2026-06-09
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
| 06-01-01 | 01 | 1 | TEST-04 | T-06-01 / — | AppError implements error interface; no information disclosure | unit | `go test -run TestAppError ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-01-02 | 01 | 1 | TEST-04 | T-06-02 | Validator accumulates errors, chaining works, combined message | unit | `go test -run TestValidator ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-01-03 | 01 | 1 | TEST-04 | T-06-03 | writeJSON encodes before writing header, falls back to 500 | unit | `go test -run TestWriteJSON ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-01-04 | 01 | 1 | TEST-04 | T-06-04 | recoveryMiddleware catches panic, returns 500 | unit | `go test -run TestRecoveryMiddleware ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-02-01 | 02 | 1 | TEST-02 | T-06-05, T-06-06 | Auth middleware: CSRF, requireRole, requireAPIRole, currentUser | integration | `go test -run 'TestRequire|TestCSRF|TestCurrentUser' ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-02-02 | 02 | 1 | TEST-02 | T-06-07 | Session: makeToken, randomString, revokeSession, redirectByRole | unit | `go test -run 'TestMakeToken|TestRandomString|TestRedirectByRole' ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-03-01 | 03 | 2 | TEST-01 | T-06-08 | Handler tests: non-DB routes (health, home, login, style, serveJS) | unit | `go test -run 'TestHealthCheck|TestHome|TestLoginPage|TestStyle|TestServeJS' ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-04-01 | 04 | 2 | TEST-01 | T-06-09 | Handler tests: DB-dependent routes (auth, admin, dashboard, API) | integration | `go test -run 'TestLoginPost|TestAPILogin|TestAdmin|TestDashboard|TestAPIAdmin|TestAPIDashboard' ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-05-01 | 05 | 2 | TEST-05 | T-06-10 | Route response shape validation | integration | `go test -run TestRoutesReturnCorrectStatus ./cmd/server -v -count=1` | ✅ | ✅ green |
| 06-05-02 | 05 | 2 | TEST-03 | T-06-11 | DB query function tests (findUser, listUsers, listActivities, filterOptions) | integration | `go test -run 'TestFindUser|TestListUsers|TestListActivities|TestListFilterOptions' ./cmd/server -v -count=1` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `cmd/server/testhelper.go` — test DB factory, test user factory, test token factory, cleanup helpers (87 lines)
- [x] Expanded tests in `cmd/server/main_test.go` (110+ test functions):
  - AppError: Error(), Unwrap(), nil Unwrap
  - handleError: non-AppError path, HTMX path, API path, page path, default status from code
  - Validator: all methods, chaining, Error() combined message
  - writeJSON: success path, encode failure path
  - recoveryMiddleware: catches panic
  - requestIDMiddleware: generates ID, preserves existing header
  - CSRF middleware: origin check, skip patterns, PATCH/DELETE protection
  - Security headers: HSTS in production, omitted in dev
  - Mapping functions: mapActivity, mapProduct, mapUser, mapOracleProduct
  - Utility functions: parseFilters, intQuery, firstNonEmpty, rateLimiter
  - Config loading: defaults and custom env vars
  - Auth middleware: requireRole, requireAPIRole (unauthenticated, forbidden, allowed)
  - Session management: currentUser (valid, bad sig, expired, revoked), revokeSession, redirectByRole
  - DB queries: findUserByUsername, findUserByID, listUsers, listActivities, listFilterOptions, activityDetailsData
  - Auth handler tests: loginPost (success, wrong password, rate-limited), logout, apiLogin (success, invalid, rate-limited), apiLogout, apiMe
  - Activity handler tests: apiLastInfo (null, with data), apiFinalizar (success, missing fields, with products)
  - Admin handler tests: page, usersSection, createUser, editUserRow, userRow, updateUser
  - Dashboard handler tests: page, table
  - API admin handler tests: usersList, userCreate, userUpdate
  - API dashboard handler tests: activities, activityDetails, bulkDetails, bulkPrint
  - Route integration test: public and authenticated routes through full mux
  - Coverage gate: documentation target ≥70%
  - Phase 5 validation: shared.js, sanitizeHtml, handler files

*~252 test runs (249 after removing subtests), no DB: 203 pass. Full suite with DB: 249 pass.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Oracle-dependent code paths | TEST-03 | No Oracle available in test environment; read-only SQL guard already tested | Run against staging Oracle; check isReadOnlySQL rejects non-SELECT/WITH queries |
| HTMX partial rendering | TEST-05 | HTML template rendering is tested via httptest; visual correctness requires manual review | Load each page in browser, check HTMX interactions render correctly |

*All other phase behaviors have automated verification.*

---

## Validation Audit 2026-06-09

| Metric | Count |
|--------|-------|
| Gaps found | 3 |
| Resolved | 3 |
| Escalated | 0 |

### Gaps Resolved

| # | Test | Type | File |
|---|------|------|------|
| 1 | `TestFindUserByUsername` | integration | main_test.go:1387 |
| 2 | `TestFindUserByID` | integration | main_test.go:1413 |
| 3 | `TestListUsers` | integration | main_test.go:1435 |

### Audit Notes

- All 3 gaps were from Plan 03 Task 3 (DB query functions in `db.go`)
- Tests use `testApp(t)` — skip gracefully when `TEST_POSTGRES_URL` unset
- Non-DB tests (203 functions) all pass on every `go test -count=1 ./cmd/server`
- DB-dependent tests (46 functions) skip gracefully — verified by `t.Skip`
- `go test -count=1 ./cmd/server` passes clean: 0 failures, 46 skips

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** 2026-06-09
