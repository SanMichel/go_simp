# Codebase Concerns

**Analysis Date:** 2026-06-05

## Tech Debt

### Single `main` package monolith
- **Issue:** All code lives in a single package `main` in `cmd/server/`. No internal packages, no library separation. This prevents importing any logic from external tools or tests, and forces all symbols to be public within package but inaccessible outside.
- **Files:** `cmd/server/main.go`, `cmd/server/models.go`, `cmd/server/handlers.go`, `cmd/server/auth.go`, `cmd/server/db.go`, `cmd/server/utils.go`, `cmd/server/api_handlers.go`
- **Impact:** Impossible to write integration tests that import application logic. All test code must live inside `package main`. Prevents any reuse or module boundary enforcement.
- **Fix approach:** Extract data access into `internal/db/`, auth into `internal/auth/`, HTTP handlers into `internal/handler/` with clear interface boundaries. Keep `cmd/server/main.go` as thin wiring.

### Massive SPA-like JavaScript files with inline DOMPurify
- **Issue:** `app.js` (1171 lines), `shared.js` (1169 lines with DOMPurify 3.4.2 inlined), `admin.js` and `dashboard.js` (both ~1170 lines) are each monolithic client-side files. DOMPurify is vendored inline into `shared.js` rather than loaded as a separate dependency.
- **Files:** `cmd/server/templates/app.js`, `cmd/server/templates/shared.js`, `cmd/server/templates/admin.js`, `cmd/server/templates/dashboard.js`
- **Impact:** Difficult to maintain, review, or test individual features. Bundling DOMPurify manually means version updates require manual re-vendoring. No module system — all files share global scope.
- **Fix approach:** Use ES modules, split into feature-specific files, serve DOMPurify from a CDN or proper npm install + build step.

### Inline CSS in Go code
- **Issue:** CSS is defined as a `const` string inside `main.go` rather than in a dedicated CSS file embedded via `go:embed`.
- **Files:** `cmd/server/main.go`
- **Impact:** CSS cannot be edited independently during development. Any CSS change requires a Go recompile.
- **Fix approach:** Move CSS into `templates/style.css` (already exists for some styles) and serve it via the existing `go:embed` mechanism used for other static files.

### Single-line HTML templates
- **Issue:** Several templates (`print.html`, `activity_modal.html`, `user_row.html`, `user_edit_row.html`) are written as single-line HTML, making them extremely difficult to read, diff, or debug.
- **Files:** `cmd/server/templates/print.html`, `cmd/server/templates/components/activity_modal.html`, `cmd/server/templates/components/user_row.html`, `cmd/server/templates/components/user_edit_row.html`
- **Impact:** Template syntax errors are hard to spot. Reviewing changes produces unusable diffs.
- **Fix approach:** Format all `.html` templates with proper indentation and line breaks.

### Duplicated Admin/Dashboard JavaScript
- **Issue:** `admin.js` and `dashboard.js` share extensive duplicate code (search dropdowns, print windows, table rendering). This is copy-pasted code from a shared ancestor.
- **Files:** `cmd/server/templates/admin.js` (1169 lines), `cmd/server/templates/dashboard.js` (1169 lines)
- **Impact:** Bug fixes must be applied twice. Feature changes drift apart.
- **Fix approach:** Extract shared UI components into `shared.js` and remove duplicates.

### Oracle schema names hardcoded in queries
- **Issue:** All Oracle queries hardcode the schema name `CONSINCO.` (e.g., `CONSINCO.MRL_PRODUTOEMPRESA`, `CONSINCO.MAP_PRODUTO`).
- **Files:** `cmd/server/db.go` (lines 333-381)
- **Impact:** If the Oracle schema changes, all queries must be updated. No environment-specific configuration.
- **Fix approach:** Make the schema prefix configurable via environment variable or construct queries dynamically.

### Mixed HTMX and custom SPA approach
- **Issue:** The project includes HTMX (`htmx.min.js`) and uses HTMX for dashboard/admin pages, but the main activities flow (`atividades.html` + `app.js`) is a fully custom single-page application with `fetch()` calls and DOM manipulation.
- **Files:** `cmd/server/templates/htmx.min.js`, `cmd/server/templates/app.js`, `cmd/server/templates/atividades.html`
- **Impact:** Two different frontend paradigms in the same application. Confusing for developers. HTMX is included but only used for admin/dashboard — could replace the entire SPA.
- **Fix approach:** Standardize on one approach (HTMX recommended for server-rendered HTML) and remove the unused paradigm.

### `tmp/` directory as dead reference code
- **Issue:** The `tmp/` directory is a read-only reference copy of an old application architecture (TypeScript/Node.js). It is explicitly marked "do not modify" but still part of the repository.
- **Files:** `tmp/` directory (entire tree)
- **Impact:** Bloats repository size with unused code. Developers may accidentally reference or modify it.
- **Fix approach:** Remove `tmp/` directory from version control. Store in documentation if historical reference is needed.

---

## Known Bugs

### JSON decode error silently ignored in admin API handlers
- **Issue:** In `apiAdminUserCreate` (line 123) and `apiAdminUserUpdate` (line 148), the error from `json.NewDecoder(r.Body).Decode(&req)` is ignored. If invalid JSON is sent, `req` remains zero-valued and the handler proceeds with empty/zero defaults.
- **Files:** `cmd/server/api_handlers.go` (lines 123, 148)
- **Trigger:** Send malformed JSON to `POST /api/admin/users` or `PATCH /api/admin/users/{id}`.
- **Symptoms:** User could be created with empty username and empty password (length 0 < 8 triggers validation error, so it would fail, but the behavior is still incorrect).
- **Workaround:** None needed — validation catches most cases by chance, not by design.
- **Fix approach:** Check and return `json.Decode` error immediately: `if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeJSON(w, http.StatusBadRequest, ...); return }`

### `printActivities` updates activities before rendering
- **Issue:** In `handlers.go:135-136`, activities are marked as `impresso=true` *before* the template is rendered. If the template rendering fails (e.g., network drops mid-response), the activity is permanently marked as printed even though the user never received the output.
- **Files:** `cmd/server/handlers.go` (lines 135-137)
- **Trigger:** Print activities when a network interruption occurs during template rendering.
- **Symptoms:** Activities show as printed forever but user never actually printed them.
- **Fix approach:** Move the `UPDATE` after successful rendering, or use a two-phase approach (mark as printed only on confirmation).

### Subdomain CSRF bypass
- **Issue:** The CSRF Origin check at `auth.go:147` only verifies `strings.Contains(origin, "://"+r.Host)`. This can be bypassed by a subdomain of the target — e.g., if `r.Host` is `app.example.com`, an origin of `https://app.example.com.evil.com` would pass the check because it contains `://app.example.com`.
- **Files:** `cmd/server/auth.go` (line 147)
- **Trigger:** A page on a subdomain-attacker domain.
- **Symptoms:** CSRF tokens could be bypassed for logged-in users.
- **Workaround:** The CSRF cookie + header check on `/api/` routes provides a second layer of defense.
- **Fix approach:** Use `url.Parse(origin)` and compare host exactly: `parsedOrigin.Host == r.Host`.

### `POST /logout` lacks CSRF check
- **Issue:** The logout endpoint (`POST /logout`) is excluded from CSRF checks (line 143 condition skips `/login` only — logout IS subject to CSRF). Actually, re-reading: logout IS subject to CSRF. But `/login` is skipped. However, there's no GET-based logout, so this is okay. Verified: the CSRF middleware skips `/login` and `/api/auth/login`. Logout endpoints are properly protected.
- **Files:** `cmd/server/auth.go` (line 143)
- **Clarification:** After review, logout IS CSRF-protected. No bug here. Only login endpoints are whitelisted.

---

## Security Considerations

### CSRF token cookie is not HttpOnly
- **Issue:** The `csrf_token` cookie (`auth.go:129`) is not set with `HttpOnly: true`. JavaScript must read it to include as `X-CSRF-Token` header. This is an inherent design trade-off of the double-submit cookie pattern, but it means any XSS vulnerability can exfiltrate the CSRF token.
- **Files:** `cmd/server/auth.go` (line 129)
- **Current mitigation:** CSP restricts script sources. DOMPurify sanitizes user-rendered HTML.
- **Recommendations:** Consider adopting a cryptographically bound CSRF token pattern (e.g., HMAC-session-bound token in a cookie) instead of the readable double-submit pattern.

### Content Security Policy allows `'unsafe-inline'`
- **Issue:** The CSP header at `utils.go:222` includes `'unsafe-inline'` for both `script-src` and `style-src`. This significantly weakens XSS protection.
- **Files:** `cmd/server/utils.go` (line 222)
- **Current mitigation:** DOMPurify sanitization on the frontend.
- **Recommendations:** Generate a nonce for each request and use `'nonce-...'` instead of `'unsafe-inline'`. This requires server-side nonce injection into templates.

### `randomString` fallback is predictable
- **Issue:** In `utils.go:78`, if `crypto/rand.Read` fails, `randomString` falls back to `strconv.FormatInt(time.Now().UnixNano(), 36)` which is a predictable timestamp.
- **Files:** `cmd/server/utils.go` (lines 75-81)
- **Risk:** If the system entropy pool is exhausted, CSRF tokens become guessable.
- **Current mitigation:** On modern Linux, `rand.Read` only fails in extreme edge cases (entropy pool exhaustion).
- **Recommendations:** If fallback is needed, use a combination of timestamp + PID + a private salt, or panic/fatal instead of silently degrading.

### Default admin credentials created on first run
- **Issue:** `seedAdmin` (`db.go:160`) creates a default `admin`/`admin` sysadmin account on first startup. These credentials are well-known and documented in AGENTS.md.
- **Files:** `cmd/server/db.go` (line 170), `AGENTS.md`
- **Risk:** If deployed to production without changing default credentials, anyone can log in as sysadmin.
- **Current mitigation:** Only seeds if user table is empty.
- **Recommendations:** Force password change on first login, or require setting initial admin password via environment variable.

### Session token has no unique ID / nonce
- **Issue:** The token payload deliberately does not include a nonce (`main_test.go:178-181` asserts "Token should NOT have a nonce field"). Without a unique token ID, individual sessions cannot be revoked — `revokeSession` invalidates *all* tokens for a user by updating `last_token_at`.
- **Files:** `cmd/server/auth.go` (lines 58-73), `cmd/server/main_test.go` (lines 178-181)
- **Risk:** Cannot implement per-session logout, concurrent session limits, or token-specific audit trails.
- **Recommendations:** Add a UUID nonce/jti to each token. Store active tokens in a `sessions` table for per-session revocation.

### Login rate limiting uses IP from `r.RemoteAddr`
- **Issue:** The rate limiter in `handlers.go:29` uses `r.RemoteAddr` to identify clients. Behind a reverse proxy (nginx, load balancer), this will always be the proxy's IP, making rate limiting ineffective.
- **Files:** `cmd/server/handlers.go` (line 29), `cmd/server/utils.go` (lines 204-218)
- **Risk:** Brute-force attacks from a single source behind a proxy are not limited.
- **Recommendations:** Add `X-Forwarded-For` header parsing, with a configurable trust list for proxy IPs.

### Oracle credentials may leak in process listings
- **Issue:** The `.env.example` file itself warns: "The Oracle URL may appear in process listings and logs." The URL is constructed from individual env vars at startup (`utils.go:31-42`).
- **Files:** `cmd/server/utils.go` (lines 31-42), `.env.example` (line 9)
- **Risk:** Connection string (with password) visible in `ps aux` output and potentially in logs.
- **Current mitigation:** Log output in `main.go:47` only says "warning: oracle ping failed" without the URL.
- **Recommendations:** Configure Oracle via named connection string that is never logged.

---

## Performance Bottlenecks

### N+1 Oracle queries per activity detail
- **Issue:** `activityDetailsData` (`db.go:312-351`) iterates over each product verification row and makes a separate Oracle query to fetch description, stock, and sales data. For an activity with 200 products, this generates 200 separate Oracle round-trips.
- **Files:** `cmd/server/db.go` (lines 331-345)
- **Cause:** No batch query — each product's data is fetched individually.
- **Improvement path:** Replace the loop with a single Oracle query using `SEQPRODUTO IN (...)` with all product IDs.

### N+1 activity detail fetches during bulk operations
- **Issue:** `printActivities` (`handlers.go:123-138`), `apiDashboardBulkDetails` (`api_handlers.go:217-241`), and `apiDashboardBulkPrint` (`api_handlers.go:243-251`) call `activityDetailsData` in a loop — once per activity ID. For 10 activities, this executes 10+ Oracle N+1 cascades.
- **Files:** `cmd/server/handlers.go` (lines 129-134), `cmd/server/api_handlers.go` (lines 227-236)
- **Cause:** No batch data fetching. Each activity's details are fetched independently.
- **Improvement path:** Add a `batchActivityDetailsData` function that fetches all activities and their product verifications in bulk queries.

### Dynamic SQL query string per filter combination
- **Issue:** `listActivities` (`db.go:204-284`) builds a SQL query string using `fmt.Sprintf` to insert parameter placeholders. Each unique filter combination produces a different query string, preventing PostgreSQL's query plan cache from being effective.
- **Files:** `cmd/server/db.go` (lines 247-254)
- **Cause:** Parameterized query placeholders are numbered dynamically (`$1`, `$2`, etc.) based on the number of filter values.
- **Improvement path:** Use a fixed query with `COALESCE` or `(column = ANY($N) OR $N IS NULL)` patterns where unused parameters are passed as NULL.

### Redundant filter parsing in dashboard
- **Issue:** `dashboardTable` (`handlers.go:83-92`) calls `parseFilters(r)` twice — once for the activities query (line 84) and once for the render data (line 91).
- **Files:** `cmd/server/handlers.go` (lines 84, 91)
- **Improvement path:** Parse filters once and reuse the result.

### `findFullProductByCode` makes two Oracle round-trips
- **Issue:** `findFullProductByCode` (`db.go:365-382`) first calls `findAddressByCode` (one Oracle query), then makes a second Oracle query for the full product data. These could be merged.
- **Files:** `cmd/server/db.go` (lines 365-381)
- **Improvement path:** Combine into a single query using a CTE or subquery.

---

## Fragile Areas

### `isReadOnlySQL` guard function
- **Why fragile:** This function (`db.go:28-64`) attempts to prevent write operations on the Oracle connection by parsing SQL strings. It has a complex comment-stripping and keyword-matching algorithm that could miss edge cases. Oracle-specific DML like `FLASHBACK`, `PURGE`, `LOCK TABLE` are not in the blocklist. The comment stripper is custom and may have parsing bugs.
- **Files:** `cmd/server/db.go` (lines 28-108)
- **Safe modification:** Always add new DML keywords to the blocklist in `db.go:50-55`. Test with `go test` which includes `TestOracleReadOnlySQLGuard` and `TestRemoveSQLComments`.
- **Test coverage:** Covered by `TestOracleReadOnlySQLGuard` (13 test cases) and `TestRemoveSQLComments` (5 test cases). Missing cases: Oracle-specific keywords, multi-line comments in tricky positions, SQL injection-style bypass attempts.

### Custom session token implementation
- **Why fragile:** The token system (`auth.go:17-73`) is a custom HMAC-signed JSON payload. Unlike standard JWT/PASETO libraries, there's no validation of standard claims (iss, aud), no key rotation support, no token ID for revocation tracking, and strict base64 encoding must match between creation and verification.
- **Files:** `cmd/server/auth.go` (lines 17-73)
- **Safe modification:** Both `makeToken` and `currentUser` must stay in sync — same payload structure, same HMAC key usage, same encoding.
- **Test coverage:** Covered by `TestMakeToken` and implicit paths in CSRF tests. No test for token verification (expired, bad signature, malformed).

### `rateLimiter` goroutine without cleanup
- **Why fragile:** The background cleanup goroutine in `newRateLimiter` (`utils.go:188-199`) runs forever with no shutdown mechanism. In tests that create multiple `rateLimiter` instances, these goroutines leak.
- **Files:** `cmd/server/utils.go` (lines 186-202)
- **Safe modification:** Add a `stop` channel parameter. In production (`main.go`), it runs for the process lifetime so this is a test-only concern.
- **Test coverage:** `TestRateLimiter` creates a rate limiter but doesn't verify cleanup (no stop channel exists).

### `finalizeReq.Empresa` typed as `any`
- **Why fragile:** `finalizeReq.Empresa` in `models.go:128` is typed as `any` (empty interface), then converted to string via `fmt.Sprint(req.Empresa)` in `handlers.go:392`. If the JSON sends a number or object instead of a string, `fmt.Sprint` produces unexpected output.
- **Files:** `cmd/server/models.go` (line 128), `cmd/server/handlers.go` (line 392)
- **Safe modification:** Add explicit JSON type handling — accept both string and number with `json.RawMessage` or use `string` and validate on decode.
- **Test coverage:** Not covered by any test.

---

## Scaling Limits

### Oracle connection pool size
- **Current capacity:** Configurable, default max 5 connections, idle 1.
- **Limit:** The N+1 query pattern in `activityDetailsData` can exhaust the pool under concurrent usage. With 2 concurrent managers viewing different activity details, each with 50+ products, the Oracle pool (default 5) is rapidly exhausted.
- **Scaling path:** Increase `ORACLE_MAX_CONNS`, but the real fix is eliminating N+1 queries.

### Dashboard activity list limit
- **Current capacity:** Hard-coded 50-row limit for non-API dashboard (`handlers.go:78`), 200 for API (`api_handlers.go:190`).
- **Limit:** No pagination mechanism. As the `atividades` table grows, the 200-row limit will become insufficient and the query will slow down.
- **Scaling path:** Add server-side pagination (cursor-based or offset-based) with configurable page size.

---

## Dependencies at Risk

### `go-ora/v2` Oracle driver
- **Risk:** The `go-ora/v2` library (`github.com/sijms/go-ora/v2`) is maintained by a single organization and has a smaller community than `godror`. It may have compatibility issues with newer Oracle versions or less support for edge cases.
- **Impact:** If the driver has a bug with specific Oracle versions or SQL features, the entire Oracle integration breaks.
- **Migration plan:** Wrap all Oracle calls in an interface (`OracleReader`) to make swapping drivers possible. Consider testing against the `godror` driver as a fallback.

### Vendored DOMPurify 3.4.2
- **Risk:** DOMPurify is copied inline into `shared.js`. Security patches to DOMPurify require manual re-vendoring. If a CVE is published for DOMPurify 3.4.2, it must be manually updated.
- **Impact:** XSS sanitization may be compromised until the vendored copy is updated.
- **Migration plan:** Load DOMPurify from a CDN with SRI hash or add a build step that pulls the latest version.

---

## Test Coverage Gaps

### Database layer untested
- **What's not tested:** All database functions — `migrate`, `seedAdmin`, `listActivities`, `listUsers`, `activityDetailsData`, `findUserByUsername`, `findUserByID`, `listFilterOptions`, all Oracle query functions.
- **Files:** `cmd/server/db.go` (entire file)
- **Risk:** Schema migration errors, SQL injection vectors, or incorrect query logic would only be caught in production.
- **Priority:** High

### HTTP handler logic untested with real dependencies
- **What's not tested:** All route handlers — `loginPost`, `adminCreateUser`, `apiFinalizar`, `printActivities`, `dashboardTable`, etc.
- **Files:** `cmd/server/handlers.go` (entire file), `cmd/server/api_handlers.go` (entire file)
- **Risk:** Business logic errors in form validation, role checks, or data processing could ship without detection.
- **Priority:** High

### Oracle read-only guard bypass cases
- **What's not tested:** Edge cases in `isReadOnlySQL` — Oracle-specific syntax like `FLASHBACK TABLE`, multi-statement queries with CTE-based DML, nested comments with DML keywords inside.
- **Files:** `cmd/server/db.go` (lines 28-108)
- **Risk:** A crafted query could bypass the guard and write to the Oracle database.
- **Priority:** High

### Configuration error paths untested
- **What's not tested:** `loadConfig` behavior when `SESSION_SECRET` is missing (log.Fatal), when variables are invalid, when Oracle URL construction fails.
- **Files:** `cmd/server/utils.go` (lines 17-80)
- **Risk:** Configuration errors that cause silent failures or crashes at startup are not caught by tests.
- **Priority:** Medium

---

*Concerns audit: 2026-06-05*
