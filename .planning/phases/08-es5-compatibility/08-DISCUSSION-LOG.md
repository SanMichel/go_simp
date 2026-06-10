# Phase 8: ES5 Compatibility — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-10
**Phase:** 8-ES5 Compatibility
**Areas discussed:** JS rewrite scope, Rewrite approach, Device testing strategy, Page weight optimization

---

## JS Rewrite Scope

| Option | Description | Selected |
|--------|-------------|----------|
| app.js + shared.js + login.js | All three files used by the scanning workflow path | |
| app.js only | Strictest interpretation of scanning workflow JS | |
| Include admin.js part, too | Admin panel partially used by conferentes | |
| **Split entirely** | **Standalone /atividades directory, own login, own JS files** | ✓ |

**User's choice:** Split atividades entirely — standalone folder, own login page
**Notes:** User wants /atividades to have its own templates/atividades/ directory with independent login and scanning pages. The existing login.js, shared.js, and app.js stay untouched for admin/dashboard users. Backend API endpoints are shared.

### Detailed questions

| Question | Answer |
|----------|--------|
| Embedded login or separate route? | Separate route — /atividades/login |
| JS organization? | 4 files: login, scan, consulta, utils |
| shared.js dependency? | Copy utilities into atividades-utils.js. Comment out originals in shared.js with TODO. |
| Template location? | Both templates in templates/atividades/ |
| Go-rendered templates or JS-driven? | Multiple Go-rendered templates |
| Backend route changes? | New handler for /atividades/login. Modify /atividades redirect. |
| Template count? | 2: atividades-login.html + atividades.html |

---

## Rewrite Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Manual rewrite | Write from scratch in ES5. No build step, fully controlled. | |
| One-time Babel pass | Use Babel once to transpile, then split output. | |
| Manual + verification | Manual rewrite plus linter to verify no ES6+ features. | |
| **Write ES5 directly** | **No Babel. No build step. 4 new files written directly in ES5.** | ✓ |

**User's choice:** Write ES5 directly
**Notes:** Considered Babel but decided against it — since these are new files anyway, writing them directly in ES5 avoids adding a dependency.

---

## Device Testing Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Specific device list | Define exact devices/browsers to test | |
| **General guideline** | **Code must work on any browser on warehouse devices** | ✓ |
| Emulators only | Test on BrowserStack/Android emulators | |
| **Real warehouse devices** | **Coordinate with operations for real hardware testing** | ✓ |
| Default 2.x, fallback manually | Start with HTMX 2.x, swap if needed | |
| Start with 1.9.x for atividades | Use proven-compatible version from day one | |
| **Keep current version, test first** | **Use whatever is vendored now. Test. Swap only if needed.** | ✓ |
| Replace global HTMX if needed | All pages use same version | ✓ |
| Iterative on-device testing | Ship, test on real devices, fix issues | ✓ |

**User's choice:** General guideline + real warehouse devices + keep current HTMX + test first
**Notes:** HTMX was briefly considered for removal from atividades, but user decided to start with HTMX and only remove if it proves incompatible on real devices.

---

## Page Weight Optimization

| Option | Description | Selected |
|--------|-------------|----------|
| JS file size target | Focus on total KB transferred | |
| Render performance target | Focus on time to interactive | |
| **Both size and performance** | **Both matter equally** | ✓ |
| Under 50 KB | Hard size target | |
| Under 100 KB | Relaxed target | |
| **No hard target** | **Clean, concise ES5 code** | ✓ |
| Minify JS | Serve .min.js versions | |
| **Skip minification** | **Readability over wire size** | ✓ |
| Simplify template HTML | Reduce DOM complexity | |
| **No CSS changes** | **Existing styles are fine** | ✓ |

**User's choice:** Both size and performance, no hard target, no minification, no CSS changes
**Notes:** Optimization is about writing clean ES5 code, not aggressive size reduction or CSS changes.

---

## the agent's Discretion

- JS file organization within templates/atividades/ (internal function naming, module-like patterns)
- XMLHttpRequest wrapper design for API calls
- Go handler implementation for /atividades/login route
- Template structure details for login and main atividades pages
- ES5 fallback patterns for browser features

## Deferred Ideas

None — discussion stayed within phase scope.
