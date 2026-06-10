---
phase: 07
slug: handler-decomposition
status: verified
nyquist_compliant: true
wave_0_complete: true
created: 2026-06-09
verified: 2026-06-10
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none |
| **Quick run command** | `go test ./cmd/server -short -count=1` |
| **Full suite command** | `go test ./cmd/server -count=1` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/server -short -count=1`
- **After every plan wave:** Run `go test ./cmd/server -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| T-01 | 01 | 1 | HAND-01 | T-07-01..05 | Parameterized SQL, TX ownership, empty-slice guard | unit | `go test ./cmd/server -run "Test(APIFinalizar\|FinalizeActivity_)" -count=1` | ✅ | ✅ green |
| T-02 | 01 | 1 | HAND-03 | T-07-03 | No HTTP types in service signature | static | `grep -n 'http\.Request\|http\.ResponseWriter' cmd/server/activity_handlers.go \| grep -v '_test.go' \| grep -v 'func.*App.*(w.*http'` | ✅ | ✅ green |
| T-03 | 01 | 1 | HAND-04 | — | Handler ≤20 lines (was ~101) | manual | `wc -l` inspection | ✅ | ✅ green |
| T-04 | 01 | 1 | HAND-05 | — | 3 handler tests + 3 service tests pass | unit | `go test ./cmd/server -run "TestAPIFinalizar\|TestFinalizeActivity_" -count=1` | ✅ | ✅ green |
| T-05 | 02 | 2 | HAND-02 | T-07-06..11 | Self-edit guard, sysadmin protection, role validation | unit | `go test ./cmd/server -run "Test(UpdateUserAdmin\|BulkActivityDetails\|AdminUpdateUser\|APIAdminUserUpdate\|APIDashboardBulkDetails)" -count=1` | ✅ | ✅ green |
| T-06 | 02 | 2 | HAND-03 | T-07-06..09 | No HTTP types in updateUserAdmin or bulkActivityDetails | static | `grep -n 'http\.Request\|http\.ResponseWriter' cmd/server/admin_handlers.go cmd/server/api_handlers.go \| grep -v '_test.go' \| grep -v 'func.*App.*(w.*http'` | ✅ | ✅ green |
| T-07 | 02 | 2 | HAND-04 | — | All handlers ≤20 lines (adminUpdateUser, apiAdminUserUpdate, apiDashboardBulkDetails) | manual | `wc -l` inspection | ✅ | ✅ green |
| T-08 | 02 | 2 | HAND-05 | T-07-08..11 | 3 handler tests + 6 service tests pass; FlatBundle moved to models.go | unit | `go test ./cmd/server -run "Test(UpdateUserAdmin\|BulkActivityDetails\|AdminUpdateUser\|APIAdminUserUpdate\|APIDashboardBulkDetails)" -count=1` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

> Existing infrastructure covers all phase requirements. No new test framework needed.

- [x] No Wave 0 required — Go testing framework built in, existing tests cover all handlers

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Handler still works end-to-end in browser | HAND-05 | Integration requires running server | `go run ./cmd/server`, log in, verify each decomposed handler still works |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** verified 2026-06-10

---

## Validation Audit 2026-06-10

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
| Total tests | 15 (6 handler + 9 service) |
| Requirements covered | 5/5 |
