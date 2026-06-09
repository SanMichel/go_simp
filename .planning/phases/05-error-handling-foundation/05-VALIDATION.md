---
phase: 5
slug: error-handling-foundation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-09
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — Go standard toolchain |
| **Quick run command** | `go test ./cmd/server -count=1` |
| **Full suite command** | `go test ./cmd/server -count=1 -v` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/server -count=1`
- **After every plan wave:** Run `go test ./cmd/server -count=1 -v`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 5-01-01 | 01 | 1 | ERR-01, ERR-04 | T-05-01 / — | N/A | unit | `go vet ./cmd/server` | ✅ | ⬜ pending |
| 5-01-02 | 01 | 1 | ERR-02, ERR-07 | T-05-02 / — | Error responses must not leak stack traces | unit | `go vet ./cmd/server` | ✅ | ⬜ pending |
| 5-01-03 | 01 | 1 | ERR-05, ERR-06 | T-05-03 / — | Panic recovery prevents crash | unit | `go vet ./cmd/server` | ✅ | ⬜ pending |
| 5-02-01 | 02 | 2 | ERR-03 | T-05-04 / T-05-05 | Consistent JSON error shape prevents info disclosure | manual | `go test ./cmd/server -count=1` | ✅ | ⬜ pending |
| 5-02-02 | 02 | 2 | ERR-05 | — | N/A | manual | `go test ./cmd/server -count=1` | ✅ | ⬜ pending |
| 5-03-01 | 03 | 2 | ERR-03 | T-05-06 | Consistent JSON error shape prevents info disclosure | manual | `go test ./cmd/server -count=1` | ✅ | ⬜ pending |
| 5-04-01 | 04 | 3 | HAND-06 | — | N/A | manual | `go vet ./cmd/server` | ✅ | ⬜ pending |
| 5-04-02 | 04 | 3 | ES5-02 | T-05-09 | escHtml is superset of DOMPurify for template strings | manual | `go vet ./cmd/server` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] Existing infrastructure covers all phase requirements — `go test` and `go vet` are pre-installed.
- [ ] New types (AppError, Validator, recovery middleware) tested inline in `main_test.go` during implementation.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| HTMX error dispatch renders correct partial | ERR-02 | Requires browser or HTMX test harness | Start server, trigger error with HX-Request header, verify response partial |
| ES5-02 DOMPurify removal | ES5-02 | Build-time check | Verify shared.js no longer contains `DOMPurify` string |
| File reorganization | HAND-06 | Structural check | Verify handlers.go split into domain files with correct imports |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
