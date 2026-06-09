# Pitfalls Research

**Domain:** Go stdlib web app with HTMX — adding code quality, testing, ES5 compat, and camera scanning to an existing warehouse application
**Researched:** 2026-06-08
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: ES6+ JavaScript breaks completely on ES5-only warehouse browsers

**What goes wrong:**
The entire frontend (app.js, dashboard.js, admin.js, shared.js) stops working on legacy warehouse browsers. Users see blank screens or cryptic syntax errors. Scanning operations halt.

**Why it happens:**
The current JS files are transpiled from TypeScript but retain extensive ES6+ features that are **not supported** in ES5:
- Arrow functions (`() => {}`), `const`/`let` — used everywhere, all files
- Template literals (`` `${var}` ``) — hundreds of interpolations in app.js
- Optional chaining (`?.`) — dozens: `document.getElementById()?.classList`, `state.user?.username`
- `async`/`await` — every API call function
- `for...of`, `Set`, `Array.from()`, `NodeList.forEach()`, default parameters, destructuring
- `Array.includes()`, `.find()`, `.some()`, `.filter()`, `.map()`, `Object.entries()`

These will throw `SyntaxError` before executing a single line.

**How to avoid:**
1. Add a Babel transpilation step with `@babel/preset-env` targeting the specific device browsers.
2. Use `es-check` programmatically to validate output JS files against ES5 after transpilation.
3. DOMPurify 3.4.2 (inlined in shared.js) uses ES6+ — **must** go through Babel too, or be replaced with the existing `escHtml()` which is ES5-compatible.
4. HTMX 2.x bundle must be ES5-checked — if not, pin to HTMX 1.9.x.

**Warning signs:**
- `es-check es5 dist/*.js` fails → untranspiled code
- Target device shows blank screen, `SyntaxError: expected ;` or `SyntaxError: unexpected token =>`
- Source maps show original arrow functions → Babel not running

**Phase to address:**
COMPAT-01 (ES5 compatibility)

---

### Pitfall 2: Camera scanning requires HTTPS — fails silently over HTTP

**What goes wrong:**
Camera scanning works when tested locally on `localhost` but silently fails on warehouse devices served over HTTP. `navigator.mediaDevices` is `undefined`. Users can't scan barcodes. No error is surfaced.

**Why it happens:**
`MediaDevices.getUserMedia()` **requires a secure context (HTTPS)**. In insecure contexts, `navigator.mediaDevices` does not exist. Localhost is treated as secure by browsers, so testing locally never catches this. Additionally, legacy Android WebViews may only support the deprecated `navigator.getUserMedia()` with vendor prefixes (`webkitGetUserMedia`, `mozGetUserMedia`).

**How to avoid:**
1. Deploy behind HTTPS (TLS termination at reverse proxy). Non-negotiable for camera access.
2. Implement progressive enhancement with legacy fallback:
```javascript
async function getCameraStream(constraints) {
    try {
        return await navigator.mediaDevices.getUserMedia(constraints);
    } catch (e) {
        return new Promise((resolve, reject) => {
            const legacy = navigator.getUserMedia
                || navigator.webkitGetUserMedia
                || navigator.mozGetUserMedia;
            if (!legacy) { reject(new Error('Camera not supported')); return; }
            legacy.call(navigator, constraints, resolve, reject);
        });
    }
}
```
3. Detect camera support gracefully — show manual barcode entry as primary fallback.
4. Add `media-src 'self' blob:` to Content-Security-Policy.

**Warning signs:**
- Camera button does nothing, no console errors → likely insecure context
- `navigator.mediaDevices` is `undefined`
- Camera works on developer laptop but not on warehouse devices
- `navigator.getUserMedia` exists but is undefined after WebView update

**Phase to address:**
COMPAT-03 (Camera scanning) — HTTPS is a prerequisite

---

### Pitfall 3: Handler decomposition breaks implicit ResponseWriter contracts

**What goes wrong:**
Splitting large handler functions into smaller helpers changes HTTP response behavior. Headers written twice, status codes set after body starts writing, responses silently swallowed. App returns 200 OK with empty body instead of proper errors.

**Why it happens:**
In Go's `net/http`, once `w.WriteHeader()` or `w.Write()` is called, the response status and headers are locked. When refactoring handlers:
1. A helper calls `http.Error(w, ...)` which internally calls `w.WriteHeader()` — caller then tries to write more headers, silently ignored.
2. `logWriter` wrapper (utils.go:139) intercepts `WriteHeader` — helpers bypassing it (using raw `http.Error()`) won't have status logged.
3. Early returns in decomposed functions may skip critical cleanup (`defer rows.Close()`).

**How to avoid:**
1. Establish a strict pattern: **one function writes the response**. Helper functions return data/errors, never write to ResponseWriter.
2. Create standard response helpers and use everywhere:
```go
func (a *App) respondError(w http.ResponseWriter, r *http.Request, status int, msg string) {
    if strings.HasPrefix(r.URL.Path, "/api/") {
        writeJSON(w, status, map[string]string{"error": msg})
    } else {
        http.Error(w, msg, status)
    }
}
```
3. Add tests that wrap handlers in `httptest.ResponseRecorder` and verify single header write.
4. Never return `(int, error)` from a helper that partially writes a response.

**Warning signs:**
- `superfluous response.WriteHeader call` in logs (Go 1.23 logs this automatically)
- `logWriter.status` always 200 even for error responses
- API responses return truncated body

**Phase to address:**
CODE-01 (Handler decomposition) — enforce response ownership rules before splitting

---

### Pitfall 4: Service extraction breaks transaction integrity in apiFinalizar

**What goes wrong:**
Extracting the `apiFinalizar` transaction logic (handlers.go:434-534) into a separate `finalizeActivity()` service function introduces bugs in transaction management. Connections leak, partial commits happen, phantom rows appear.

**Why it happens:**
The current code manually manages a Postgres transaction with `BeginTx`, `defer tx.Rollback()`, and `tx.Commit()` spanning ~100 lines with multiple error return points. Three specific traps:
1. **`defer tx.Rollback()` on a committed tx is safe** (Go docs confirm it's a no-op) — but if the helper returns early before reaching `tx.Commit()`, the deferred rollback runs and the caller never knows the transaction was abandoned.
2. **The Oracle read-only guard** (`isReadOnlySQL`) blocks write queries — if a service function uses `ora` pool for a write, it fails silently. The `QueryRowContext` guard (db.go:302-304) returns a dummy row instead of an error, masking the problem.
3. **Context cancellation** mid-transaction (timeout, user navigates away) causes `tx.ExecContext()` to fail, but the deferred rollback may not propagate the cancellation error correctly.

**How to avoid:**
1. Extract with a clean function signature: `func (a *App) finalizeActivity(ctx context.Context, req finalizeReq, userID int) error`
2. The extracted function owns the full transaction lifecycle — `BeginTx` in, nothing but `Commit` or `Rollback` out.
3. Do NOT return an open transaction from the service function. Transaction must be fully resolved inside.
4. Write tests that verify transaction behavior: success path, partial DB failure (`context.Canceled` injected via context), and duplicate submission (same `codigo_barras` twice).
5. Keep `defer tx.Rollback()` — it's safe after `tx.Commit()` returns nil.

**Warning signs:**
- Integration test shows missing `ProductVerification` rows after successful finalization
- Increased Postgres connection pool exhaustion
- `pq: current transaction is aborted, commands ignored until end of transaction block`

**Phase to address:**
CODE-03 (Service extraction) — this is the highest-risk refactoring in the milestone

---

### Pitfall 5: WriteJSON header-then-body ordering

**What goes wrong:**
When `writeJSON` fails to encode the response body (e.g., circular reference, type assertion panic), the HTTP status has already been written as 200. The client receives 200 OK with an empty or partial body. Error responses are silently lost.

**Why it happens:**
Current `writeJSON` (utils.go:165-174):
```go
func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data) // can return error, but it's ignored
}
```
The `WriteHeader(status)` happens **before** `Encode()`. If encoding fails, the status is already locked. Additionally, `Encode()` can panic if `data` contains unsupported types (channels, functions, complex numbers).

**How to avoid:**
1. Buffer the JSON first, then write:
```go
func writeJSON(w http.ResponseWriter, status int, data any) {
    buf, err := json.Marshal(data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(buf)
}
```
2. Never pass `any` where interface could be typed (caught more encoding errors at compile time).

**Warning signs:**
- Client receives 200 OK with empty body
- `http: superfluous response.WriteHeader call` when someone adds error handling after WriteHeader
- API responses are valid JSON that bears no relation to the data (e.g., `null`)

**Phase to address:**
CODE-02 (Error handling) — fix writeJSON before other error handling changes, as all API handlers depend on it

---

### Pitfall 6: CSRF middleware breaks HTMX requests

**What goes wrong:**
HX-* headers from HTMX requests get a 403 Forbidden or 302 redirect instead of proper responses. The CSRF middleware rejects HTMX form submissions. HTMX then renders the redirected HTML response fragment inline, corrupting the page.

**Why it happens:**
1. HTMX sends `HX-Request: true` header on every AJAX request. The CSRF middleware skips validation only for exact path `/login` (auth.go:27-31). All other paths require CSRF token validation.
2. The `requireRole` middleware returns a 302 redirect on authorization failure (auth.go:46-50). HTMX follows the redirect silently — the user sees the login page rendered inside their current container. The page is corrupted.
3. Session cookie validation happens per-request. If the session DB query fails (connection pool issue, Oracle not available), the middleware returns 403 before the handler runs.

**How to avoid:**
1. Keep CSRF skip path exact: `r.URL.Path == "/login"` — not `strings.HasPrefix`, not regex.
2. Add a CSRF token helper for HTMX that doesn't require a full page render:
```go
// Return CSRF token for HTMX-initiated requests
func (a *App) handleCSRFToken(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := ctx.Value(ctxUser).(int)
    token, err := a.generateCSRFToken(ctx, userID)
    if err != nil { http.Error(w, "CSRF error", 500); return }
    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(token))
}
```
3. For `requireRole`, return JSON error for API paths and HTMX-triggered requests instead of redirect. HTMX can handle 401/403 statuses.
4. Test middleware order invariants — recovery → CSRF → auth → log → handler.

**Warning signs:**
- HTMX requests return HTML fragments that look like the login page
- Console shows 302 redirects on HTMX POST requests
- "CSRF token mismatch" intermittent — possibly session DB connectivity
- Trailing slash on `/login` bypasses CSRF check

**Phase to address:**
CODE-02 (Error handling) — fix middleware error responses for HTMX

---

### Pitfall 7: Template FuncMap and go:embed break silently

**What goes wrong:**
Adding or renaming template files breaks `go:embed` patterns at compile time. Adding template functions to `FuncMap` without registering them causes runtime panics. Both happen on different pipelines (compile vs runtime) making debugging harder.

**Why it happens:**
1. `go:embed` directive matches `cmd/server/templates/*` (glob). Moving templates to subdirectories to organize them removes them from the embed — only panics at runtime when `ExecuteTemplate` can't find the file.
2. `template.FuncMap` entries are registered once at startup. If a function is added to the map but the handler still calls `$.SomeFunction`, the template panics silently — `html/template` returns an error from `ExecuteTemplate` that most handlers ignore.

**How to avoid:**
1. Template organization: keep all templates in flat `templates/` directory. Use naming convention with prefixes for organization: `admin_users_list`, `admin_users_form`.
2. Register all template functions in a single `funcMap` variable with a test that validates coverage:
```go
// In main_test.go
func TestFuncMapContract(t *testing.T) {
    // Read all template files, parse with funcMap, verify no missing functions
}
```
3. Every handler that calls `ExecuteTemplate` must handle the error:
```go
func (a *App) render(w http.ResponseWriter, name string, data any) {
    if err := a.tmpl.ExecuteTemplate(w, name, data); err != nil {
        log.Printf("template error: %v", err)
        http.Error(w, "template error", 500)
    }
}
```

**Warning signs:**
- Runtime panic: `function "x" not defined` on page render
- `go build` succeeds but server panics on first request
- Template renders blank page with no error logged (error ignored in handler)

**Phase to address:**
CODE-01 (Handler decomposition) — add template registration tests alongside extraction

---

### Pitfall 8: go:embed breaks when files move across directories

**What goes wrong:**
The `go:embed` directive uses a glob pattern (`cmd/server/templates/*`). If files are moved to subdirectories during handler decomposition (e.g., `templates/admin/dashboard.html`), they're no longer matched. `go build` succeeds but the server panics at runtime.

**Why it happens:**
`go:embed` patterns are relative to the directory containing the `go` file with the directive. The pattern `templates/*` matches files directly in `templates/` but not in subdirectories. Go does not warn about zero matches — it just embeds what it finds.

**How to avoid:**
1. Keep all templates flat in `templates/`.
2. Use naming conventions for organization (e.g., `admin_dashboard.html`, `print_labels.html`).
3. If subdirectories are absolutely necessary, update the embed glob: `templates/**/*` or add multiple directives.
4. Add a test:
```go
func TestEmbeddedTemplates(t *testing.T) {
    entries, err := templateFiles.ReadDir("templates")
    if err != nil { t.Fatal(err) }
    // Parse all template entries, should succeed
}
```

**Phase to address:**
CODE-01 (Handler decomposition) — verify embed patterns before any file moves

---

## Moderate Pitfalls

### Pitfall 9: File extraction without import verification

**What goes wrong:**
Extracting handlers, auth, or DB code into separate files leaves behind stale import references in main.go. Go compiler catches this but the error messages can be confusing when multiple files reference the same types.

**How to avoid:**
1. Incremental extraction: create new file, copy the code, `go build` to verify, only then delete from original.
2. Keep all files in `package main` — no sub-packages, no import cycles.
3. After each extraction, run `go vet ./cmd/server/` to check for suspicious patterns.

### Pitfall 10: Error adapter pattern breaks HTMX error handling

**What goes wrong:**
The `handleError` adapter (wrapping `func(w, r) error`) changes how errors surface. HTMX receives JSON error bodies for API calls but HTML for page requests — mixing them causes HTMX to render error payloads as page fragments.

**How to avoid:**
1. Make the error adapter aware of content type:
```go
func (a *App) handleError(fn func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := fn(w, r); err != nil {
            if strings.HasPrefix(r.URL.Path, "/api/") || r.Header.Get("HX-Request") == "true" {
                writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            } else {
                http.Error(w, "Erro interno", http.StatusInternalServerError)
            }
        }
    }
}
```
2. Use the adapter selectively — not every handler needs it.

### Pitfall 11: Table-driven tests that only hit happy paths

**What goes wrong:**
Adding table-driven tests that only test the success case gives a false sense of coverage. Test passes, confidence increases, but all the edge cases are still untested.

**How to avoid:**
For every handler/function being tested, include at minimum:
- ✅ Success case (known valid input → expected output)
- ❌ Bad input (malformed JSON, missing fields → 400)
- ❌ Unauthorized (no session → 401/302 via middleware)
- ❌ Not found (nonexistent ID → 404)
- ❌ Internal error (mock DB that returns error → 500)

The table-driven Go subtest pattern (`t.Run`) makes this easy to add cases.

### Pitfall 12: N+1 query pattern in activityDetailsData

**What goes wrong:**
The `activityDetailsData` function (db.go) runs an initial query to get activities, then loops through them executing a query per activity. At scale (100+ activities per shift), this generates 101+ DB round trips. The page load time degrades linearly.

**How to avoid:**
1. Rewrite as a single JOIN query that fetches all data at once.
2. If the extraction is deferred, add a TODO comment in the code and a test that measures query count by wrapping the DB driver.
3. Document this as a known performance issue — don't fix now, but don't make it worse by adding more queries to the loop.

### Pitfall 13: Oracle read-only guard blocks MERGE and CALL statements

**What goes wrong:**
The `isReadOnlySQL` function (db.go:270-276) checks if a statement starts with `SELECT` or `WITH`. Oracle uses `MERGE` (upsert) and `CALL` (stored procedures) — these are blocked despite being read-write in practice.

**How to avoid:**
1. The guard is intentional — Oracle connection is documented as read-only in AGENTS.md.
2. If a future feature needs Oracle writes, the guard must be explicitly updated.
3. For now, document this limitation in the `.env.example` and code comments.

---

## Minor Pitfalls

### Pitfall 14: Centralized error handler suppresses different error types

**What goes wrong:**
A single error handler that always returns `{"error": msg}` loses nuance — validation errors should return field-level information (e.g., `{"field": "codigo_barras", "error": "campo obrigatório"}`), not flat error strings.

**How to avoid:**
1. Define error types that carry extra context:
```go
type ValidationError struct {
    Field  string `json:"field"`
    Detail string `json:"error"`
}
type AppError struct {
    Status  int    `json:"-"`
    Message string `json:"error"`
    Err     error  `json:"-"`
}
```
2. The error adapter checks the type: `ValidationError` → 400 with field detail, `AppError` → appropriate status, `error` → 500 generic.

### Pitfall 15: LogWriter overwrites WriteHeader on concurrent requests

**What goes wrong:**
The `logWriter` struct (utils.go:139-146) wraps `ResponseWriter` to capture status and size. If the same `logWriter` instance is reused across requests (it isn't — it's created per-request in the middleware chain), concurrent writes would corrupt. Not currently an issue, but worth understanding.

**What to do:** The current pattern (`logWriter` created fresh in middleware handler) is correct. Do not extract `logWriter` to a shared/global instance.

### Pitfall 16: HTMX request headers bypassed by the route split

**What goes wrong:**
When routes are split across files (`routes_web.go`, `routes_api.go`), setting HTMX-specific headers (HX-Redirect, HX-Refresh) may be missed by handlers in the other file.

**How to avoid:**
1. Document which HTMX response headers each handler group uses.
2. Create a `hxRedirect(w, url)` helper and use it consistently.
3. Keep all route registrations in a single `routes()` function even if handler implementations are split.

### Pitfall 17: `apiCallError` fallback catches too broadly

**What goes wrong:**
The `apiCallError` function (shared.js:1255) shows the full response text as an error to the user. If a server error returns stack traces or SQL queries (not currently, but regression), these leak to the user.

**How to avoid:**
Frontend: show generic "Erro ao conectar" and log the raw text to console. Backend: ensure error responses never contain stack traces or internal details.

---

## Integration Gotchas

| Gotcha | Where | What happens | Prevention |
|--------|-------|-------------|------------|
| HTMX + CSRF | Any form POST | 403 Forbidden | Verify CSRF token in HX request headers |
| Session expired + HTMX | Any page after session timeout | Login page rendered inside container | Return 401 instead of 302 for HX-Request |
| Oracle idle connection timeout | DB pool | Latent connections dropped, first query after idle fails | Set `SetConnMaxLifetime` on Oracle pool |
| Go 1.23+ TLS handshake changes | HTTPS deployment | Legacy warehouse devices on old TLS reject connection | Check min TLS version config on reverse proxy |
| Babel + template literals | ES5 transpilation | Tagged template literals from DOMPurify don't transpile cleanly | Audit Babel `@babel/polyfill` needs |
| `hx-vals` with special chars | HTMX requests | URL encoding mismatch causes 400 | Escape JSON values passed to hx-vals |

---

## Performance Traps

| Trap | Why it hurts | When it matters |
|------|-------------|-----------------|
| Query-per-activity in `activityDetailsData` | N+1 DB round trips | 50+ activities per warehouse round |
| Unbounded `SELECT *` from Oracle | Full table scan on potentially huge tables | Production warehouse with 100K+ records |
| No pagination on dashboard/activity list | Browser DOM renders all results | 100+ activities visible at once |
| CSS in Go const (JS at ~30KB) | No CSS caching, served as Go string | Every page load repasses the CSS |
| DOMPurify ~800 lines inlined (ES6) | Unused overhead, blocks validators | Every page load |

---

## Security Mistakes

| Mistake | Impact | Why it sneaks in |
|---------|--------|------------------|
| CSRF token not refreshed on login | Stale token after login still valid | Token generated once per session, no refresh hook |
| CSP with `'unsafe-inline'` | XSS not mitigated | Required for inline scripts/styles in HTML templates |
| Error messages leak Oracle/Postgres details | Enumeration attack vector | Generic error handler catches untyped errors |
| Session cookie without Secure flag (HTTP) | Session hijacking over network | Dev runs on HTTP, flag only set in HTTPS handler |
| `escHtml()` replacing `sanitizeHtml()` | XSS if user content contains HTML | Product descriptions from Oracle are plain text (safe) |
| Camera stream not released on component unmount | Camera LED stays on, resource leak | Missing `stream.getTracks().forEach(t => t.stop())` |
| Babel transpilation introduces `regeneratorRuntime` | Polyfill-based XSS via code injection | Old polyfills may have CVEs |

---

## UX Pitfalls

| Pitfall | Happens when | Consequence |
|---------|-------------|-------------|
| Scanner locked to first input | Focus-locking logic in app.js | User can't type in any other field |
| No loading state during camera init | getUserMedia takes 1-3s on slow devices | User thinks scan is broken, taps again, opens multiple streams |
| Beep on scan but muted on warehouse device | AudioContext requires user gesture | User gets no feedback, scans repeatedly creating duplicates |
| `aria-live` clear during HX-swap | HTMX replaces entire container | Screen reader loses context, announces wrong region |
| Error in Portuguese, stack in English | Mixed language UX | Inconsistent for PT-BR users |

---

## "Looks Done But Isn't" Checklist

Check these before shipping any phase:

**CODE phases:**
- [ ] `go vet ./cmd/server/` passes with zero warnings
- [ ] Every new handler has a corresponding test in `main_test.go`
- [ ] No handler writes to `w` after calling an external helper (one-function-response rule)
- [ ] All `writeJSON` calls use the buffered version
- [ ] No new global variables or `init()` functions
- [ ] Template tests pass (all FuncMap entries used by templates)

**MAINT phases:**
- [ ] `TestLoadDotEnv` and `TestLoadDotEnvCRLF` updated if env parsing changed
- [ ] No commented-out code in tests — either it's testing something or it's deleted
- [ ] Table-driven test cases include error modes, not just happy paths

**COMPAT phases:**
- [ ] `es-check es5 dist/*.js` passes
- [ ] Camera tested on actual warehouse device (not just `localhost`)
- [ ] HTMX behavior tested without JavaScript (progressive enhancement)
- [ ] CSP updated for camera blob URLs
- [ ] `playBeep` fallback for browsers without AudioContext

---

## Recovery Strategies

If a pitfall materializes in production:

| Pitfall | Immediate fix | Root-cause fix |
|---------|--------------|---------------|
| ES6 breakage | Revert JS files to previous version, deploy | Set up Babel build step, test on device |
| Camera fails over HTTP | Deploy TLS cert (Let's Encrypt) | Add graceful fallback with manual entry |
| Transaction bug | Rollback the git commit, deploy previous version | Write integration test before reintroducing |
| CSRF + HTMX breakage | Add HX-Request check to middleware | Test middleware with HTMX before shipping |
| DB connection leak | Restart server (frees connections) | Fix defer/rollback pattern in service functions |
| Broken | Rewrite webhook for error response | Update `funcMap` test to catch |

---

## Pitfall-to-Phase Mapping

| Phase | Primary Pitfalls | Secondary Pitfalls | Research Depth Needed |
|-------|-----------------|-------------------|-----------------------|
| **CODE-01** Handler decomposition | #3 (ResponseWriter), #7 (Templates), #8 (go:embed) | #9 (Imports), #16 (Route split) | LOW — Go stdlib well-known patterns |
| **CODE-02** Error handling | #5 (writeJSON), #6 (CSRF+HTMX), #10 (Error adapter) | #14 (Error types), #17 (Error leak) | LOW — choose pattern, implement consistently |
| **CODE-03** Service extraction | #4 (Transaction integrity) | #12 (N+1 queries) | MEDIUM — needs integration test DB setup |
| **CODE-04** Test coverage | #11 (Happy-path tests) | #9 (Import refs) | LOW — Go testing conventions are well-documented |
| **MAINT-01/02/03** Maintenance | #9 (Import cycles), #13 (Oracle guard) | #15 (LogWriter) | LOW — existing patterns to preserve |
| **COMPAT-01** ES5 conversion | #1 (ES6 breaks), #2 (DOMPurify) | #10 (HTMX version) | HIGH — needs Babel setup, device testing |
| **COMPAT-02** Device testing | #1 (ES6), #2 (DOMPurify) | — | MEDIUM — test plan needed |
| **COMPAT-03** Camera scanning | #2 (HTTPS required), legacy API | Stream release | HIGH — needs secure context + device testing |

---

## Sources

- **Codebase analysis:** Direct inspection of all Go files in `cmd/server/` (main.go, handlers.go, api_handlers.go, auth.go, db.go, utils.go, models.go, main_test.go) and all JS files (app.js, dashboard.js, admin.js, shared.js, login.js, htmx.min.js). ES6+ features confirmed by pattern search across all JS files. [MEDIUM/HIGH confidence — direct codebase audit]
- **`mediaDevices.getUserMedia()` HTTPS requirement:** MDN docs — `navigator.mediaDevices` is `undefined` in insecure contexts. https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getUserMedia#security [HIGH confidence — MDN spec]
- **Legacy getUserMedia deprecation:** MDN — `navigator.getUserMedia()` deprecated but still supported in old WebView. Polyfill strategy documented. https://developer.mozilla.org/en-US/docs/Web/API/Navigator/getUserMedia [HIGH confidence]
- **Go `database/sql` transaction semantics:** Go docs — `Tx.Rollback()` is a no-op after `Tx.Commit()` succeeds. https://pkg.go.dev/database/sql#Tx.Rollback [HIGH confidence]
- **`http.ResponseWriter` header locking:** Go docs — `WriteHeader` locks headers; subsequent calls are logged by Go 1.23+. https://pkg.go.dev/net/http#ResponseWriter [HIGH confidence]
- **Babel `@babel/preset-env` for ES5:** Babel docs — preset-env with `targets` automatically determines needed transforms. https://babeljs.io/docs/babel-preset-env [HIGH confidence]
- **DOMPurify 3.x ES6 requirements:** DOMPurify targets modern browsers; source uses arrow functions, `const`, template literals, `Array.from()`, `Set`, `for...of`. [MEDIUM confidence — code inspection, not explicitly documented]
- **HTMX 2.x browser support:** HTMX 2.x dropped IE11 support. ES5-specific bundle not available. https://htmx.org/posts/2024-htmx-2-0/ [HIGH confidence]
- **`es-check` validation tool:** NPM package for checking JS syntax level. https://www.npmjs.com/package/es-check [MEDIUM confidence — docs may have changed]
- **Go middleware patterns:** Go blog — middleware wrapping pattern for `http.Handler`. https://go.dev/blog/context [HIGH confidence]