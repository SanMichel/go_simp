---
phase: 08
slug: es5-compatibility
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-10
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` (stdlib) |
| **Config file** | none — Go testing convention |
| **Quick run command** | `go test ./cmd/server -run TestAtividades -v -count=1 -timeout 30s` (with TEST_POSTGRES_URL) |
| **Full suite command** | `go test ./cmd/server -count=1 -timeout 60s` (requires TEST_POSTGRES_URL) |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** `go vet ./cmd/server 2>&1` + manual JS review
- **After every plan wave:** `go test ./cmd/server -count=1 -timeout 60s`
- **Before `/gsd-verify-work`:** Full Go suite green + real-device HTMX verified
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 08-01-01 | 01 | 1 | ES5-01 | N/A | N/A | smoke | `go vet ./cmd/server 2>&1` (embed check) | ⬜ W0 | ⬜ pending |
| 08-01-02 | 01 | 1 | ES5-01 | N/A | N/A | unit | `go test ./cmd/server -run TestTemplatesParse -v -count=1` | ✅ | ⬜ pending |
| 08-01-03 | 01 | 1 | ES5-01 | T-08-01 | escHtml on all innerHTML | unit | Manual JS code review | ❌ Not automated | ⬜ pending |
| 08-01-04 | 01 | 1 | ES5-01 | T-08-02 | XHR uses POST not GET for mutations | review | Manual review | ❌ Not automated | ⬜ pending |
| 08-02-01 | 02 | 1 | ES5-03, ES5-04 | T-08-03 | HTMX tested on real device | manual | On-device verification | ❌ Manual only | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `TestAtividadesLoginPage_Success` — new handler test for `/atividades/login` route
- [ ] `TestAtividadesPage_UnauthenticatedRedirect` — verify redirect to `/atividades/login`
- [ ] `TestAtividadesLoginPage_AuthenticatedRedirect` — already-authenticated users skip login

*Testing infrastructure exists (testhelper.go, testApp, testUser). Only new test functions needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| HTMX compatibility on warehouse browsers | ES5-03 | Real devices needed; browser emulators insufficient | 1. Deploy phase to staging 2. Navigate to /atividades on warehouse device 3. Complete scan flow (login → scan → finalizar) 4. If hx-get/hx-post work → pass. If not → swap htmx.min.js to 1.9.x, repeat. |
| JS syntax ES5 compliance | ES5-01 | No JS linter/parser in project | 1. Open each of the 4 new JS files 2. Search for `const`/`let`/`=>`/`async`/`await`/`` ` `` 3. Confirm zero matches |
| Page weight/render optimization | ES5-04 | Subjective / visual | 1. Load page on warehouse device 2. Measure time-to-interactive 3. Ensure clean render without jank/reflow |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
