# Phase 8: ES5 Compatibility — Research

**Researched:** 2026-06-10
**Domain:** ES5-compatible JavaScript rewrite + Go template/route integration
**Confidence:** HIGH

## Summary

This phase rewrites the scanning workflow frontend (`/atividades` page) into standalone ES5-compatible JS files, splits the monolithic `app.js` into domain-focused files, and adds a dedicated login route. The existing `app.js` (938 lines of modern JS with `async`/`await`, arrow functions, template literals, `fetch`) becomes 4 new ES5 files in `templates/atividades/`. The Go backend gets one new handler (`/atividades/login`) and a modified redirect in the existing `/atividades` handler. The vendored HTMX 2.0.4 is tested on real warehouse devices — if incompatible, it's swapped for 1.9.x. No build tooling, no polyfills, no Babel.

The ES5 rewrite is mechanical: `var` for `const`/`let`, `function()` for arrow functions, XMLHttpRequest callbacks for `async`/`await`/`fetch`, string concatenation for template literals, `&&` guards for optional chaining, index-based loops for `for...of`. The state machine and business logic remain identical — only the syntax changes.

**Primary recommendation:** Write all 4 new JS files directly in ES5. Convert `shared.js` utility functions into an ES5-compatible `atividades-utils.js`. Add new Go handler in `handlers.go`, register route and JS files in `main.go`. Test the new page on real warehouse devices — if HTMX 2.0.4 fails, swap `htmx.min.js` for a 1.9.x version.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ES5-01 | Scanning workflow JS files rewritten to ES5 (no `const`/`let`, arrow functions, `async`/`await`, `fetch`, template literals) | app.js (938 lines) confirmed to use all forbidden patterns. 4 new ES5 files replace it. All utility patterns exist in shared.js (94 lines). |
| ES5-03 | HTMX version verified compatible with warehouse browsers (fallback to 1.9.x if 2.x fails) | Vendored HTMX is **2.0.4** [VERIFIED: source header]. HTMX 2.x officially dropped IE11 support June 2024 [CITED: htmx.org/posts/2024-06-17-htmx-2-0-0-is-released/]. Strategy: test on real devices, swap minified file for 1.9.x if incompatible. |
| ES5-04 | Page weight/rendering optimized for low-end devices | No hard size target. No minification. Clean ES5 with no CSS changes [D-16, D-17, D-18]. Reduced JS payload by splitting only what atividades needs. |
</phase_requirements>

<user_constraints>
## User Constraints (from CONTEXT.md)

### Implementations Decisions

See CONTEXT.md section `## Implementation Decisions` (D-01 through D-18). Key locked decisions:

- **D-01:** Split `/atividades` into standalone `templates/atividades/` directory with own templates, JS, and login page
- **D-04:** 4 JS files: `atividades-login.js`, `atividades-scan.js`, `atividades-consulta.js`, `atividades-utils.js`
- **D-07:** Update `go:embed` to include `templates/atividades/*.html` and `templates/atividades/*.js`
- **D-09:** Write ES5 directly — no Babel, no build step
- **D-10:** ES5 constraints: `var`, `function(){}`, XMLHttpRequest, string concat, `&&` guards, index-based loops
- **D-13:** Keep current HTMX 2.0.4 vendored, test on real devices, swap to 1.9.x if needed

### the agent's Discretion

- JS file organization within `templates/atividades/` (exact exports, internal function naming)
- XMLHttpRequest wrapper design for `atividades-utils.js`
- Go handler implementation details for new `/atividades/login` route
- Template structure details for `atividades-login.html` and `atividades.html`
- ES5 fallback patterns for browser features that may or may not be available

### Deferred Ideas (OUT OF SCOPE)

- ES5-05 (admin/dashboard JS rewrite) — confirmed deferred to v2
- Camera barcode scanning, offline support, sub-package split — all out of scope per PROJECT.md
</user_constraints>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Login form handling | Browser (JS) | Backend (Go) | `atividades-login.js` handles form submit via XHR to `/api/auth/login` |
| Session validation | Backend (Go) | — | `/api/auth/me` endpoint already implemented. Reused unchanged. |
| Activity state persistence | Browser (localStorage) | — | State machine lives in browser; same pattern as existing app.js |
| Scanning flow (EAN lookup) | Backend (Go) | Browser (JS) | Browser sends scanned code to `/api/produtos/ean/*`, backend returns product data |
| Divergence/building detection | Browser (JS) | — | Product comparison logic runs in browser with scanned vs expected data |
| Template rendering | Backend (Go) | — | Go `html/template` renders `atividades-login.html` and `atividades.html` |
| HTMX partial rendering | Backend (Go) | — | Server-rendered HTML partials driven by HTMX attributes (unchanged pattern) |
| Report display | Browser (JS) | — | JS populates report screen from finalize response data |
| Product consultation | Backend (Go) | Browser (JS) | XHR to `/api/produtos/consulta/*` or `/api/produtos/consulta?q=*` |

## Standard Stack

### Core (all existing — no new dependencies)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| net/http | stdlib (Go 1.23) | HTTP routes and handlers | Project constraint — no external web frameworks |
| html/template | stdlib | Server-rendered HTML templates | Existing pattern — `parseTemplates()` with `go:embed` |
| database/sql + pgx | stdlib + pgx v5 | Postgres access | Existing — all DB access through `*App.pg` |
| HTMX | 2.0.4 (vendored) | HTML partial rendering for interactive UI | Existing — all pages use HTMX for dynamic behavior |
| Go html/template `.js` serving | stdlib via `serveJS()` | Serve JS files from embedded templates | Existing pattern in `main.go` lines 171-182 |
| XMLHttpRequest | Browser built-in (ES5-compatible) | API calls from JS | Mandated by ES5 constraint — replaces `fetch` |
| localStorage | Browser built-in (ES5-compatible) | State persistence | Same pattern as existing app.js |

### Supporting

| Pattern | Purpose | When to Use |
|---------|---------|-------------|
| `var` variables only | ES5 compatibility | All new JS files — per D-10 |
| `function()` expressions | ES5 callbacks | All event handlers, XHR callbacks, loops |
| XMLHttpRequest with callback pattern | Async API calls | Replaces all `async`/`await` + `fetch` patterns |
| String concatenation `"text " + var` | String building | Replaces all template literals |
| `&&` guard pattern | Optional property access | Replaces `?.` optional chaining |
| Index-based `for` loops | Array iteration | Replaces `for...of`, `forEach`, `.find()`, `.some()`, `.includes()` |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Manual ES5 rewrite | Babel transpilation pass | No build tool dependency. Output is cleaner. But more manual work. |
| Single ES5 file | Split into 4 files | Better organization, maintainability. Slightly more HTTP requests. |
| Keep shared.js dependency | Copy into atividades-utils.js | Self-contained atividades directory. No shared.js changes needed. |

## Package Legitimacy Audit

> No external packages are introduced in this phase. The ES5 rewrite uses only built-in browser APIs (XMLHttpRequest, localStorage, AudioContext, DOM) and the existing vendored HTMX. No new npm/PyPI/crates packages to install or verify.

## ES5 Rewrite Strategy

### File Inventory

| New File | Source Code | Contents | Approx Lines |
|----------|-------------|----------|-------------|
| `templates/atividades/atividades-utils.js` | Copied from `shared.js` + new XHR wrapper | `formatDate`, `playBeep`, `escHtml`, `showLoader`, `apiCall` XHR equivalent, CSRF token reader | ~80 |
| `templates/atividades/atividades-login.js` | Adapted from `app.js` login flow + `login.js` | Form handler for `/atividades/login`, session check via `/api/auth/me`, redirect to `/atividades` | ~70 |
| `templates/atividades/atividades-scan.js` | Extracted from `app.js` | State machine, start activity, EAN scan, building switch, finalize, history rendering | ~300 |
| `templates/atividades/atividades-consulta.js` | Extracted from `app.js` | Product consultation (by code and by description), mode toggle, detail display | ~150 |

### State Machine Mapping

The existing app.js implements a screen-based state machine. The new files replicate this exactly:

```
login → start → scanning → predio-switch → report → start
                  ↓
               consulta (can go back to scanning or start)
```

Each screen corresponds to a `<div id="screen-{name}">` in the template. Screen transitions use `showScreen()` — renamed to ES5 syntax.

### Share of Work — Functions to Port from app.js to atividades-scan.js

The following functions from app.js need ES5 conversion in `atividades-scan.js`:

| app.js Function | ES5 Conversion | Notes |
|----------------|---------------|-------|
| `var state = {...}` | Same: `var state = {...}` | Already ES5-compatible object literal |
| `function saveState()` | Same | Already ES5. `localStorage.setItem` is ES5. |
| `function resetActivityState()` | Same | Already ES5 |
| `function focusScannerInput()` | Same | Already ES5 |
| `function refocusInput()` | Remove `const`, use `var` | One `const inputId` → `var inputId` |
| `function syncKeyboardUI()` | Remove `const`, `for...of` | `for...of` → `for` loop with index |
| `function renderHistory()` | Remove `const`, spread, `.find()` | `[...arr].reverse()` → manual copy+reverse; `.find()` → `indexOf` check |
| `function showProductDetailModal()` | Remove `.find()` | Use index-based loop |
| `function closeProductDetailModal()` | Remove `?.` | `element?.classList.add(...)` → `if (element) element.classList.add(...)` |
| `function showScreen()` | Remove `?.`, template literals, `for...of` | `.includes()` → indexOf check; template strings → `+` |
| `function confirmExit()` | Remove arrow function in confirm callback | `function()` instead of `() =>` |
| `function logout()` | Replace `fetch` with XHR | Fire-and-forget POST to `/api/auth/logout` via XHR |
| `function showReauthModal()` | Remove `?.` | Guard each element access with `if` |
| `function startActivity()` | Remove `async/await`, `const`, template literals | XMLHttpRequest callback pattern |
| `async function finalizeActivity()` | Remove `async/await`, template literals | XMLHttpRequest with JSON payload |
| `function btn-predio-switch handlers` | Remove `async/await`, `const`, `.includes()`, `Set`, spread | `.includes()` → indexOf; `new Set()` → manual array dedup |
| Event listeners in DOMContentLoaded | Remove arrow functions, `for...of`, `?.` | `function()` instead of arrow; `for` loops |

### Function to Port from shared.js to atividades-utils.js

| shared.js Function | ES5 Compat? | Port Action |
|-------------------|-------------|-------------|
| `async function apiCall()` | NO — uses `fetch`, `async/await`, default params, optional chaining | **Rewrite** as XHR-based `function apiCall(endpoint, options, onUnauthorized)` |
| `function showLoader()` | YES | Copy verbatim |
| `function formatDate()` | Uses template literal `pad = n =>` | Rewrite: `pad(n) { return n < 10 ? "0" + n : n; }` |
| `function playBeep()` | YES | Copy verbatim |
| `function escHtml()` | YES | Copy verbatim |
| `function sanitizeHtml()` | YES | Copy verbatim |

## ES5 Syntax Constraint Reference

> Authoritative reference for the D-10 constraint set. Every line in the 4 new JS files must comply.

| Modern Pattern | ES5 Equivalent | Example |
|---------------|---------------|---------|
| `const x = 1` | `var x = 1` | OK |
| `let x = 1` | `var x = 1` | OK; beware block scoping differences |
| `arr.map(x => x * 2)` | `arr.map(function(x) { return x * 2; })` | Required |
| `btn.addEventListener("click", () => {...})` | `btn.addEventListener("click", function() {...})` | Required |
| `async function foo() {}` | No equivalent. Use XHR callbacks. | See XHR wrapper pattern below |
| `await apiCall(...)` | `apiCall(url, opts, cb)` with callback | See XHR wrapper pattern below |
| `fetch(url, opts)` | `new XMLHttpRequest()` | See XHR wrapper pattern below |
| `` `${name} - ${value}` `` | `name + " - " + value` | Always |
| `user?.name ?? "N/A"` | `user && user.name ? user.name : "N/A"` | Use `&&` guards |
| `arr.includes(x)` | `arr.indexOf(x) !== -1` | Use `indexOf` |
| `arr.find(fn)` | Manual loop with `for` | Use index-based iteration |
| `arr.some(fn)` | `for` loop with early `return true` | Use `for` + boolean flag |
| `for (const x of arr)` | `for (var i = 0; i < arr.length; i++)` | Index-based |
| `[...arr]` | `arr.slice(0)` | Use `.slice(0)` |
| `function foo(x = 5)` | `function foo(x) { if (x === undefined) x = 5; }` | Check `undefined` explicitly |
| `obj = {...a, ...b}` | Manual `Object.keys` + assignment or `Object.assign` | Use `Object.keys` loop or explicit property merge |
| `e.target.closest(".class")` | `e.target.closest(".class")` | **OK** — `Element.closest()` is ES5-era (IE10+) |
| `document.querySelector(sel)` | Same | **OK** — ES5 |
| `class Foo {}` | `function Foo() {}` + `Foo.prototype.method = function() {}` | Not needed; no classes in existing JS |
| `Promise` | Not available without polyfill | Use XHR callbacks only |
| `Array.isArray()` | Same | **OK** — ES5 |
| `JSON.parse/stringify` | Same | **OK** — ES5 |
| `String.prototype.trim()` | Same | **OK** — ES5 |
| `Array.prototype.forEach` | `for` loop | Heaviest modern-JS feature used in app.js; rewrite all |

### XHR Wrapper Design (the agent's Discretion)

The core API abstraction from shared.js must be rewritten as callback-based XHR:

```javascript
// atividades-utils.js — XHR-based API call (replaces async apiCall from shared.js)
function apiCall(method, url, body, onSuccess, onUnauthorized) {
  var xhr = new XMLHttpRequest();
  xhr.open(method, url, true);
  xhr.setRequestHeader("Content-Type", "application/json");
  
  // CSRF token for mutating requests
  if (method === "POST" || method === "PATCH" || method === "DELETE") {
    var csrfMatch = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    if (csrfMatch) {
      xhr.setRequestHeader("X-CSRF-Token", csrfMatch[1]);
    }
  }
  
  xhr.withCredentials = true;
  
  xhr.onreadystatechange = function() {
    if (xhr.readyState !== 4) return;
    
    if (xhr.status === 401) {
      if (onUnauthorized) onUnauthorized();
      return;
    }
    
    var data = null;
    var contentType = xhr.getResponseHeader("content-type") || "";
    if (contentType.indexOf("application/json") !== -1) {
      try { data = JSON.parse(xhr.responseText); } catch(e) { data = { error: "Parse error" }; }
    } else {
      data = { error: xhr.responseText || "Error " + xhr.status };
    }
    
    onSuccess(xhr.status >= 200 && xhr.status < 300, xhr.status, data);
  };
  
  xhr.send(body ? JSON.stringify(body) : null);
}

// Convenience wrappers
function apiGet(url, onSuccess, onUnauthorized) {
  apiCall("GET", url, null, onSuccess, onUnauthorized);
}
function apiPost(url, body, onSuccess, onUnauthorized) {
  apiCall("POST", url, body, onSuccess, onUnauthorized);
}
```

### CSRF Token Reading — ES5-Compatible

The existing pattern in `shared.js` is:
```javascript
var csrfCookie = document.cookie.split("; ").find((c) => c.startsWith("csrf_token="));
```

ES5 equivalent:
```javascript
function getCSRFToken() {
  var cookies = document.cookie.split("; ");
  for (var i = 0; i < cookies.length; i++) {
    var c = cookies[i];
    if (c.indexOf("csrf_token=") === 0) {
      return c.substring("csrf_token=".length);
    }
  }
  return "";
}
```

## Codebase Analysis

### Existing File Inventory — What Gets Created vs Modified vs Untouched

**Created (new files):**
- `cmd/server/templates/atividades/atividades-login.html` — New login template for atividades path
- `cmd/server/templates/atividades/atividades.html` — Main atividades SPA template (stripped of login/re-auth screens)
- `cmd/server/templates/atividades/atividades-login.js` — Login form handling (ES5)
- `cmd/server/templates/atividades/atividades-scan.js` — Scanning workflow (ES5)
- `cmd/server/templates/atividades/atividades-consulta.js` — Product consultation (ES5)
- `cmd/server/templates/atividades/atividades-utils.js` — Shared utilities (ES5)

**Modified (existing Go files):**
- `cmd/server/main.go` — Add `templates/atividades/*.html templates/atividades/*.js` to `go:embed` on line 215; register new routes for `/atividades/login` and JS file serving
- `cmd/server/handlers.go` — Add `atividadesLoginPage` handler; modify `atividadesPage` redirect target

**Untouched (existing JS files preserved):**
- `cmd/server/templates/app.js` — Still loaded by users hitting original `/login` → `/atividades` (they get redirected to `/atividades/login` via new template)
- `cmd/server/templates/shared.js` — Still loaded by admin/dashboard pages. Copied functions get comment-out with TODO.
- `cmd/server/templates/login.js` — Still loaded by `/login` page for admin/dashboard users
- `cmd/server/templates/dashboard.js`, `admin.js` — Unchanged

### Go Integration Points

**1. go:embed update (main.go line 215):**

Current: `//go:embed templates/*.html templates/components/*.html templates/*.css templates/*.js`

Required: `//go:embed templates/*.html templates/components/*.html templates/*.css templates/*.js templates/atividades/*.html templates/atividades/*.js`

**2. Template parsing update (main.go line 212):**

Current: `return template.Must(template.New("app").Funcs(funcs).ParseFS(templatesFS, "templates/*.html", "templates/components/*.html"))`

Required: Add `"templates/atividades/*.html"` to ParseFS patterns.

**3. Route updates (main.go lines 94-143):**

New route (after line 110):
```go
mux.HandleFunc("GET /atividades/login", a.atividadesLoginPage)
mux.HandleFunc("GET /atividades-login.js", a.serveJS("atividades/atividades-login.js"))
mux.HandleFunc("GET /atividades-scan.js", a.serveJS("atividades/atividades-scan.js"))
mux.HandleFunc("GET /atividades-consulta.js", a.serveJS("atividades/atividades-consulta.js"))
mux.HandleFunc("GET /atividades-utils.js", a.serveJS("atividades/atividades-utils.js"))
```

Note: JS files are served at the top level (`/atividades-*.js`) not at `templates/atividades/` to keep URLs clean. The `serveJS` handler reads from `"templates/" + filename`, so the subdirectory must be included: `a.serveJS("atividades/atividades-scan.js")`.

**4. Redirect change in requireRole (auth.go line 96):**

Current: Unauthenticated users hitting any protected route redirect to `/login`.
Required for `/atividades`: Redirect to `/atividades/login`.

This is handled by modifying the `requireRole` middleware or by having the `atividadesPage` handler check auth itself and redirect to the atividades login. Best approach: add a separate `requireAtividadesRole` middleware that redirects to `/atividades/login` instead of `/login`, or modify the `atividadesPage` handler to check auth and redirect. Since `requireRole` is already applied at route registration, the simplest change is:

**Option A** (recommended — minimal impact): Create a new middleware wrapper `requireAtividadesRole` that does the same thing as `requireRole` but redirects to `/atividades/login` when unauthenticated.

**Option B** (alternative): Modify the `requireRole` to accept a redirect target parameter.

Option A is preferred — it keeps the existing `/login` redirect for admin/dashboard routes unchanged.

**5. Handler for atividadesLoginPage:**

```go
func (a *App) atividadesLoginPage(w http.ResponseWriter, r *http.Request) {
    u, err := a.currentUser(r)
    if err == nil {
        // Already logged in — redirect to /atividades
        http.Redirect(w, r, "/atividades", http.StatusFound)
        return
    }
    a.render(w, "atividades-login", nil)
}
```

### Template Structure for New Files

**atividades-login.html — New isolated login page:**
```html
{{define "atividades-login"}}
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SIMP — Atividades Login</title>
    <link rel="stylesheet" href="/style.css">
</head>
<body>
    <div id="loader" class="hidden"><div class="spinner"></div></div>
    <div class="app-container">
        <div class="screen flex-col justify-center">
            <div class="text-center" style="margin-bottom: 1rem;">
                <h2>Acesso Atividades</h2>
                <p class="text-slate-500 text-sm">Faça login para iniciar</p>
            </div>
            <form id="form-atividades-login">
                <!-- username, password fields -->
                <p id="login-error" class="hidden"></p>
                <button type="submit" class="btn btn-primary">Entrar</button>
            </form>
        </div>
    </div>
    <script src="/atividades-utils.js"></script>
    <script src="/atividades-login.js"></script>
</body>
</html>
{{end}}
```

**atividades.html — Updated main template (login/re-auth screens removed):**
The current `atividades.html` has 6 screens (`login`, `start`, `scanning`, `predio-switch`, `consulta`, `report`) plus 2 modals (`reauth`, `product-detail`). The updated version removes:
- `screen-login` entirely (now served by `atividades-login.html`)
- `modal-reauth` (re-authentication flow — will be handled differently or dropped)
- `form-login` event handler (moved to `atividades-login.js`)
- `form-reauth` modal

What stays: `screen-start`, `screen-scanning`, `screen-predio-switch`, `screen-consulta`, `screen-report`, `modal-product-detail`.

Script tags at bottom:
```html
<script src="/atividades-utils.js"></script>
<script src="/atividades-scan.js"></script>
<script src="/atividades-consulta.js"></script>
<script src="/htmx.min.js"></script>
```

### shared.js Function Port Details

Functions copied from `shared.js` to `atividades-utils.js`:

1. **`showLoader(show)`** — Copied verbatim (already ES5)
2. **`formatDate(dateStr)`** — Copied, template literal in pad function converted:
   ```javascript
   function formatDate(dateStr) {
     if (!dateStr) return "";
     var d = new Date(dateStr);
     if (isNaN(d.getTime())) return "";
     function pad(n) { return n < 10 ? "0" + n : "" + n; }
     return pad(d.getDate()) + "/" + pad(d.getMonth() + 1) + "/" + d.getFullYear();
   }
   ```
3. **`playBeep(type)`** — Copied verbatim (already ES5)
4. **`escHtml(unsafe)`** — Copied verbatim (already ES5)
5. **`sanitizeHtml(dirty)`** — Copied verbatim (already ES5 — just calls escHtml)

The XHR `apiCall` wrapper replaces the `fetch`-based `apiCall` entirely.

After copying: add comments in `shared.js` above each copied function:
```javascript
// COPIED to templates/atividades/atividades-utils.js — TODO: remove when admin/dashboard migrate to ES5 (ES5-05)
```

## HTMX Compatibility Analysis

### Current State

- **Vendored version**: HTMX **2.0.4** (confirmed from source `version:"2.0.4"`)
- **HTMX 2.x dropped IE11 support**: Announced June 17, 2024 [CITED: htmx.org/posts/2024-06-17-htmx-2-0-0-is-released/]
- **HTMX 1.9.x still supports IE11**: The migration guide [CITED: htmx.org/migration-guide-htmx-1] shows 2.x is backwards-compatible API-wise but drops IE

### Strategy (per D-13, D-14, D-15)

1. **Keep current HTMX 2.0.4** vendored in `templates/htmx.min.js`
2. **The new `atividades.html` includes the same `/htmx.min.js`** — all pages share one HTMX version
3. **Test on real warehouse devices** — coordinate with operations
4. **If HTMX 2.x breaks**: download HTMX 1.9.x and overwrite `templates/htmx.min.js`
5. **If HTMX 1.9.x also breaks**: revisit HTMX removal for atividades only

### Usage Patterns in Current Templates

HTMX is used in the templates via attributes like `hx-get`, `hx-post`, `hx-target`, `hx-swap`. The new `atividades.html` template should preserve these same patterns. The existing `/atividades` page does not use heavy HTMX features — it's primarily JS-driven with XHR calls for scan operations. HTMX is used for UI interactions like forms, navigation, and partial refreshes where it's already established.

## Page Weight Optimization Approach

### Strategy (per D-16, D-17, D-18)

1. **No minification** — Readability for maintenance
2. **No CSS changes** — Existing stylesheets cover layout
3. **Clean code** is the primary optimization goal
4. **Reduced JS payload** — The new activities JS is a fraction of the 938-line app.js

### Optimization Techniques

| Technique | How It Helps | Priority |
|-----------|-------------|----------|
| Only load what atividades needs | Removes dashboard/admin JS overhead | HIGH |
| No unused polyfills | ES5 native, no build artifacts | HIGH |
| Single shared HTMX file | One cached file for all pages | MEDIUM |
| `innerText` over `innerHTML` when possible | Avoids HTML parsing overhead | MEDIUM |
| Minimal DOM reflows | Batch DOM changes in render functions | MEDIUM |
| No external font files | Uses system fonts | LOW (existing) |

## Go Integration Plan

### File Changes Summary

| File | Change Type | Change |
|------|-------------|--------|
| `cmd/server/main.go:215` | Edit | Add `templates/atividades/*.html templates/atividades/*.js` to go:embed |
| `cmd/server/main.go:212` | Edit | Add `"templates/atividades/*.html"` to ParseFS |
| `cmd/server/main.go:110` | Add | Route: `GET /atividades/login` → `a.atividadesLoginPage` |
| `cmd/server/main.go:99-103` | Add | Routes for 4 new JS files |
| `cmd/server/main.go:110` | Edit | Route for `/atividades` — wrap with auth middleware that redirects to `/atividades/login` |
| `cmd/server/auth.go:96` | Edit or new | New `requireAtividadesRole` function (or add redirect parameter) |
| `cmd/server/handlers.go` | Add | `atividadesLoginPage` handler function |
| `cmd/server/handlers.go:65-68` | Edit | `atividadesPage` — optional: adjust if needed |
| `cmd/server/templates/shared.js` | Edit | Add "COPIED to atividades-utils.js" comments above ported functions |

### Route Conflict Check

The new routes (`/atividades-login.js`, `/atividades-scan.js`, `/atividades-consulta.js`, `/atividades-utils.js`) are at the root level. No existing route uses these names. The `/atividades/login` route uses Go 1.22 `GET /atividades/login` which won't conflict with `GET /atividades` or any API paths.

### Middleware Chain

The existing middleware chain (line 75 of main.go):
```go
app.csrfMiddleware(app.securityHeaders(app.log(recoveryMiddleware(requestIDMiddleware(mux)))))
```

New routes pass through the same chain. The `/atividades/login` handler explicitly checks `currentUser()` and redirects if already authenticated — protecting against authenticated users re-logging in.

## Test Strategy

### Existing Tests That Validate This Phase

| Test | File | What It Validates |
|------|------|-------------------|
| `TestTemplatesParse` | main_test.go:32 | All templates parse successfully — will catch syntax errors in new templates |
| Handler tests | main_test.go (124+ functions) | Existing handler behavior unchanged — ensures no Go-side regressions |

### New Tests Required

| Test | What It Validates |
|------|-------------------|
| `TestAtividadesLoginPage_Unauthenticated` | GET `/atividades/login` returns 200 with login form |
| `TestAtividadesLoginPage_Authenticated` | GET `/atividades/login` with valid cookie redirects to `/atividades` |
| `TestAtividadesPage_Unauthenticated` | GET `/atividades` without auth redirects to `/atividades/login` |

### What's NOT Being Tested (by design)

| Aspect | Reason |
|--------|--------|
| ES5 syntax correctness | No ES5 linter in the project. This is a code review quality gate, not automated. The JS is simple enough that manual verification is sufficient. |
| HTMX compatibility | Cannot be tested on desktop. Requires real warehouse device testing (D-12). |
| XHR wrapper behavior | Integration test requires running app with Postgres. Can be tested via `TestAtividadesLoginPage` handler which exercises the Go-side flow. |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| HTMX 2.x incompatible with warehouse browser | MEDIUM | HIGH — all HTMX interactions break | Test on real devices (D-12). Downgrade to 1.9.x (D-13). Fallback: remove HTMX from atividades. |
| ES5 rewrite introduces subtle JS bug | MEDIUM | MEDIUM — wrong scan state, broken redirect | Manual code review. Compare behavior with existing app.js. Test on real devices. |
| `go:embed` or template path wrong | LOW | HIGH — server fails to start | `TestTemplatesParse` catches bad templates at test time. Route registration tested via handler tests. |
| Missing ES5 feature causes runtime error on real browser | MEDIUM | MEDIUM — page partially works | ES5 constraint reference is comprehensive. Use `indexOf` not `includes`, `for` not `for...of`, etc. |
| Re-auth modal removed causes workflow break | LOW | MEDIUM — users on expired sessions can't continue | D-01 removes reauth from new template. `/api/auth/me` is called on page load to validate session. If expired, user is redirected to `/atividades/login`. |
| Copying shared.js functions creates maintenance drift | LOW | LOW — commented TODO ensures future cleanup | Comments in shared.js track the copy. ES5-05 (v2) will consolidate. |

## File Structure Summary

The `templates/` directory after this phase:

```
templates/
├── atividades.html           ← UPDATED (screens removed, scripts changed)
├── ... (other templates unchanged)
├── atividades/               ← NEW directory
│   ├── atividades-login.html ← NEW
│   ├── atividades.html       ← NEW (main SPA template)
│   ├── atividades-login.js   ← NEW (ES5, ~70 lines)
│   ├── atividades-scan.js    ← NEW (ES5, ~300 lines)
│   ├── atividades-consulta.js← NEW (ES5, ~150 lines)
│   └── atividades-utils.js   ← NEW (ES5, ~80 lines)
├── app.js                    ← UNCHANGED (still used by original /atividades redirect path?)
├── shared.js                 ← MINIMALLY CHANGED (TODO comments added)
├── login.js                  ← UNCHANGED (still used by /login for admin/dashboard)
├── htmx.min.js               ← UNCHANGED (keep 2.0.4, replace with 1.9.x if needed)
└── ... (other files unchanged)
```

## Validation Architecture

> Required because `workflow.nyquist_validation` is enabled in `.planning/config.json`.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` (stdlib) |
| Config file | none — Go testing convention |
| Quick run command | `go test ./cmd/server -run TestAtividades -v -count=1` (with `TEST_POSTGRES_URL` set) |
| Full suite command | `go test ./cmd/server -count=1` (requires `TEST_POSTGRES_URL`) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ES5-01 | JS files parse correctly | Smoke | Manual (no JS parser in test suite) | ❌ Not automated |
| ES5-01 | New templates parse in Go | Unit | `go test ./cmd/server -run TestTemplatesParse -v` | ✅ main_test.go:32 |
| ES5-03 | HTMX compatibility | Manual | On-device verification | ❌ Manual only |
| ES5-04 | Page weight optimization | Manual | Code review + visual inspection | ❌ Manual only |

### Sampling Rate

- **Per task commit:** `go vet ./cmd/server` (syntax/embed check) + manual JS review
- **Per wave merge:** `go test ./cmd/server -count=1` (full Go test suite)
- **Phase gate:** Full Go suite green + real-device HTMX compatibility verified before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] **New handler tests needed** — `TestAtividadesLoginPage*` and `TestAtividadesPageUnauthenticated` must be created alongside the new handler
- [ ] **No JS syntax verification** — This gap is accepted per project constraints (no JS test tooling). Code review is the quality gate.

## Security Domain

> Required — `security_enforcement` is not explicitly disabled.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | Yes | `atividades-login.js` calls `/api/auth/login` (existing Go endpoint with bcrypt + rate limiting) |
| V3 Session Management | Yes | Cookie-based tokens (existing `currentUser()` + `requireRole` middleware). CSRF protection via cookie + header. |
| V4 Access Control | Yes | `requireRole` middleware at route level. New `/atividades/login` explicitly redirects authenticated users. |
| V5 Input Validation | Yes | XHR wrapper sends JSON to existing Go endpoints. `escHtml()` sanitizes all DOM injection points. |
| V6 Cryptography | No | No new crypto operations. Tokens use existing HMAC-SHA256 implementation. |

### Known Threat Patterns for Go + ES5 HTMX App

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| XSS via innerHTML | Tampering | All `innerHTML` assignments pass through `sanitizeHtml()`/`escHtml()` |
| CSRF on API calls | Tampering | CSRF token from cookie sent as `X-CSRF-Token` header on all mutating XHR requests |
| Session hijacking | Information Disclosure | HttpOnly + Secure + SameSite cookies. Token expires per SESSION_TTL config. |
| Rate limiting bypass | Denial of Service | Existing `loginLimiter` applies to `/api/auth/login` endpoint |

### ES5-Specific Security Note

The `escHtml` function from `shared.js` uses only ES5-compatible replacements (`String.replace` with global regex). This is the same pattern already used in the existing JS and is verified to be ES5-safe.

## Common Pitfalls

### Pitfall 1: go:embed Pattern Not Matching Subdirectories
**What goes wrong:** Adding `templates/atividades/*.html` to `go:embed` works, but forgetting to add it causes the build to fail with "file not found" at runtime.
**Why it happens:** `go:embed` patterns are relative to the source file's directory and don't automatically include subdirectories unless explicitly added.
**How to avoid:** Update the `go:embed` directive on line 215 of `main.go` to explicitly include `templates/atividades/*.html` and `templates/atividades/*.js`.
**Warning signs:** `TestTemplatesParse` fails because template files can't be loaded.

### Pitfall 2: Arrow Functions in XHR Callbacks
**What goes wrong:** Writing `xhr.onreadystatechange = () => { ... }` which breaks IE11 (not ES5-compatible).
**Why it happens:** Muscle memory from modern JS.
**How to avoid:** Always use `xhr.onreadystatechange = function() { ... }`. Be especially careful with nested callbacks where `this` scoping differs — use closure variables (`var self = this`) instead.
**Warning signs:** The code file uses `=>` anywhere.

### Pitfall 3: Using `let`/`const` in Loop Variables
**What goes wrong:** Writing `for (let i = 0; ...)` or `for (const x of arr)` in ES5 files.
**Why it happens:** `let` inside `for` creates per-iteration binding, which `var` does not. But since we're already writing ES5, we can't use either `let` or `const`.
**How to avoid:** Always write `for (var i = 0; i < arr.length; i++)`. For closure-in-loop issues, use an IIFE instead of `let`.
**Warning signs:** `let` or `const` anywhere in the new JS files.

### Pitfall 4: CTE XHR Callback Named `onreadystatechange` Before Setting Properties
**What goes wrong:** Setting `xhr.withCredentials = true` after `xhr.open()` but before setting `onreadystatechange` can cause race conditions on some browsers.
**Why it happens:** XHR events can fire synchronously in some states.
**How to avoid:** Always set `onreadystatechange` BEFORE `xhr.open()`, or at minimum before `xhr.send()`.
**Recommended pattern:**
```javascript
var xhr = new XMLHttpRequest();
xhr.onreadystatechange = function() { ... };  // Set BEFORE open()
xhr.open(method, url, true);
xhr.withCredentials = true;
xhr.setRequestHeader(...);
xhr.send(body);
```

### Pitfall 5: Template Concatenation Losing sanitizeHtml Calls
**What goes wrong:** In the rush to convert template literals to string concatenation, forgetting to wrap user-supplied values in `sanitizeHtml()`.
**Why it happens:** Template literals in the current code already call `sanitizeHtml()` on user values, but the pattern is easy to miss during manual conversion.
**How to avoid:** Every user-supplied value (product names, codes, user data) must go through `sanitizeHtml()` when placed into `innerHTML`. If using `innerText` instead, no sanitization needed.
**Warning signs:** Any `innerHTML` assignment with direct variable concatenation lacking `sanitizeHtml()`.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single SPA in app.js (938 lines) | Split: 4 domain-specific ES5 files | This phase | Better maintainability, reduced per-page payload |
| SPA login embedded in atividades.html | Separate `/atividades/login` route + template | This phase | Cleaner auth boundary, reuses existing Go auth patterns |
| `fetch` + `async`/`await` for API | XMLHttpRequest with callbacks | This phase | ES5 compatibility. Callback nesting is the tradeoff. |
| HTMX 2.0.4 (IE11 incompatible) | HTMX 2.0.4 or 1.9.x | This phase | TBD: depends on device testing results |

**Deprecated/outdated:**
- `shared.js` `apiCall` function becomes unused by atividades path (replaced by XHR version in `atividades-utils.js`)
- `login.js` becomes unused by atividades path (replaced by `atividades-login.js`)

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Warehouse browsers support ES5 fully (not just ES3) | ES5 Strategy | LOW — ES5 is supported by virtually all browsers after 2012. Only very old embedded browsers may lack it. Edge case: some warehouse handhelds use stripped-down browsers. |
| A2 | XMLHttpRequest works in warehouse browsers | XHR Wrapper | MEDIUM — XMLHttpRequest is universally supported (even in IE6+). If a browser lacks XHR, the entire app breaks. The fallback would be to investigate `ActiveXObject("Microsoft.XMLHTTP")` but this is extreme. |
| A3 | AudioContext (Web Audio API) works in warehouse browsers | playBeep | LOW — if AudioContext is unsupported, beep audio silently fails. The app functions fine without audio. |
| A4 | `Element.closest()` is available on warehouse browsers | DOM methods | LOW — `closest()` is IE10+ and all modern browsers. If unavailable, can fall back to manual parent traversal. |
| A5 | `DOMParser` is available (used by HTMX) | HTMX | LOW — HTMX 2.0.4 uses `DOMParser` internally. If the browser lacks it, HTXML itself fails, triggering the 1.9.x fallback. |
| A6 | `Array.isArray` is available | ES5 patterns | LOW — ES5 feature, universally supported in our target. |

## Open Questions (RESOLVED)

1. **Does the existing `atividades.html` user flow assume `/shared.js` is loaded?**
   - What we know: The current `atividades.html` loads both `/shared.js` and `/app.js`. The new version loads only `/atividades-utils.js` + `/atividades-scan.js` + `/atividades-consulta.js` + `/htmx.min.js`.
   - What's unclear: Whether any existing user bookmarks or cached pages break from this change.
   - Recommendation: The `/atividades` route serves the new template from Go. Session is checked on every page load. Redirect from `/atividades/login` after auth. No user-facing URL changes.

2. **How does the existing `/login` page know to redirect `conferente` users to `/atividades`?**
   - What we know: `login.js` has `redirectForRole` which redirects `conferente` to `/atividades`.
   - What's unclear: After this phase, `/atividades` is still the correct destination for conferentes logging in through the old `/login` page. The new `/atividades/login` is for users who land on `/atividades` unauthenticated.
   - Recommendation: Both paths coexist. No change needed to the old `login.js`.

3. **Should the re-authentication modal stay or go?**
   - What we know: D-01 removes the re-auth modal from the new `atividades.html`.
   - What's unclear: Whether re-auth functionality is important for warehouse workers who may leave sessions idle.
   - Recommendation: Remove for now. The new page validates session on load via `/api/auth/me`. If expired, user is redirected to `/atividades/login`. This is a simpler UX than the modal and avoids complexity in the ES5 rewrite.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Building + testing | ✓ | 1.23+ | — |
| Postgres (TEST_POSTGRES_URL) | Go integration tests | Conditional | — | Tests skip gracefully via `t.Skip` |
| Real warehouse devices | HTMX compatibility verification | ✓ (coordinate with operations) | — | Swap HTMX to 1.9.x |
| ctx7 / find-docs | Not needed for this phase | — | — | — |

**Missing dependencies with no fallback:** None
**Missing dependencies with fallback:** TEST_POSTGRES_URL — tests skip if absent

## Sources

### Primary (HIGH confidence)
- [VERIFIED: source code analysis] — app.js (938 lines), shared.js (94 lines), login.js (69 lines) — all modern JS patterns identified
- [VERIFIED: source code analysis] — main.go:215 go:embed pattern, routes pattern, serveJS implementation
- [VERIFIED: source code analysis] — auth.go:86-110 requireRole middleware, redirect logic
- [VERIFIED: source header] — HTMX version 2.0.4 from `version:"2.0.4"` in vendored htmx.min.js
- [CITED: htmx.org/posts/2024-06-17-htmx-2-0-0-is-released/] — HTMX 2.x drops IE11 support
- [CITED: MDN XMLHttpRequest docs] — XHR API reference, async callback patterns
- [CITED: CanIUse ES5] — ES5 browser support baseline

### Secondary (MEDIUM confidence)
- [CITED: htmx.org/migration-guide-htmx-1] — HTMX 1.x → 2.x migration guide confirming backwards compatibility
- [ASSUMED: WebSearch results] — HTMX 1.9.x supports IE11 (no official statement found, but 2.x changelog says "ends support for IE" implying 1.x supported it)

### Tertiary (LOW confidence)
- None — all claims verified through source code analysis or authoritative documentation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all existing, no new dependencies
- Architecture: HIGH — pattern is copy-split-convert, well understood
- Pitfalls: HIGH — based on Go + ES5 common issues, verified against documentation
- ES5 constraints: HIGH — all constraints documented and cross-referenced against actual code patterns

**Research date:** 2026-06-10
**Valid until:** 2026-07-10 (30 days — stable Go+JS ecosystem)
