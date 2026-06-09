---
phase: 07
slug: handler-decomposition
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-09
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
| TBD | 01 | 1 | HAND-01 | — | N/A | unit | `go test ./cmd/server -run ... -count=1` | ✅ | ⬜ pending |
| TBD | 02 | 1 | HAND-02 | — | N/A | unit | `go test ./cmd/server -run ... -count=1` | ✅ | ⬜ pending |
| TBD | 03 | 2 | HAND-03 | — | N/A | unit | `go test ./cmd/server -run ... -count=1` | ✅ | ⬜ pending |
| TBD | 04 | 2 | HAND-04 | — | N/A | unit | `go test ./cmd/server -run ... -count=1` | ✅ | ⬜ pending |
| TBD | 05 | 3 | HAND-05 | — | N/A | integration | `go test ./cmd/server -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

> Existing infrastructure covers all phase requirements. No new test framework needed.

- [ ] No Wave 0 required — Go testing framework built in, existing tests cover all handlers

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Handler still works end-to-end in browser | HAND-05 | Integration requires running server | `go run ./cmd/server`, log in, verify each decomposed handler still works |

*If none: "All phase behaviors have automated verification."*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
