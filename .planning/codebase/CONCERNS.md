# Codebase Concerns

**Analysis Date:** 2026-06-08

## Tech Debt

### Monolithic Package Structure
- Issue: All source files live in a single `package main` under `cmd/server/` with no architectural layering, interfaces, or sub-packages. This prevents unit testing in isolation, forces a monolithic binary, and makes dependency injection impossible to verify at compile time.
- Files: all `.go` files under `cmd/server/`
- Impact: Cannot test components in isolation. Any change to `App` struct or its methods requires recompiling everything. No clear dependency direction between "layers."
- Fix approach: Extract into `internal/db/`, `internal/auth/`, `internal/handlers/`, `internal/models/` packages with explicit interfaces for `UserStore`, `ActivityStore`, `OracleReader`.

### Dynamic SQL Query Construction in listActivities
- Issue: `listActivities` (`cmd/server/db.go:210-298`) builds SQL queries using `fmt.Sprintf` with string concatenation for WHERE clauses, ORDER BY, and LIMIT. Despite using parameterized args, the query structure itself is constructed via string operations, making it fragile.
- Files: `cmd/server/db.go:210-298`
- Impact: 75-line function with multiple moving parts. Adding new filter fields or changing sort columns requires modifying the string manipulation logic, not just adding to a struct.
- Fix approach: Use a query builder (e.g., `sq` or `goqu`) or define filter-specific query templates.

### Duplicated User CRUD Logic
- Issue: User creation is implemented twice: once in `adminCreateUser` (`cmd/server/handlers.go:159-186`, form-based) and once in `apiAdminUserCreate` (`cmd/server/api_handlers.go:191-214`, JSON-based). The same duplication exists for user update: `adminUpdateUser` (`cmd/server/handlers.go:208-251`) and `apiAdminUserUpdate` (`cmd/server/api_handlers.go:216-258`).
- Files: `cmd/server/handlers.go:159-186`, `cmd/server/api_handlers.go:191-214`, `cmd/server/handlers.go:208-251`, `cmd/server/api_handlers.go:216-258`
- Impact: Bug fixes or validation changes must be applied in 2+ places. The form handler and API handler have slightly different error handling behaviors.
- Fix approach: Extract a shared `createUser` / `updateUser` method on `App` that both handler types call.

### Massive JavaScript Files with Embedded Libraries
- Issue: The entire DOMPurify library (~1400 lines) is vendored directly into `shared.js`. `app.js` (938 lines), `dashboard.js` (1103 lines), and `admin.js` (1345 lines) are large monolithic files with no module system.
- Files: `cmd/server/templates/shared.js`, `cmd/server/templates/app.js`, `cmd/server/templates/dashboard.js`, `cmd/server/templates/admin.js`
- Impact: No tree-shaking, no code splitting, no module boundaries. Every page loads the full DOMPurify library even if not needed. Hard to maintain.
- Fix approach: Serve DOMPurify from CDN or npm-bundled file. Split JS into logical modules using ES modules (type="module").

### print.html as Single-Line Template
- Issue: `cmd/server/templates/print.html` is a single line of minified HTML with no whitespace, making it extremely difficult to read, debug, or maintain.
- Files: `cmd/server/templates/print.html`
- Impact: Any template change requires parsing the minified single line. Template syntax errors are hard to spot.
- Fix approach: Format with proper indentation. The template is small enough to remain readable.

### Full-page Login Screen Redundancy
- Issue: There are two separate full login implementations: the server-rendered `/login` page (`cmd/server/templates/login.html`) and the SPA-style login in `app.js` (atividades.html) and dashboard.js (admin.html). Each has its own login form, error handling, and redirect logic.
- Files: `cmd/server/templates/login.html`, `cmd/server/templates/atividades.html`, `cmd/server/templates/admin.html`, `cmd/server/templates/login.js`, `cmd/server/templates/dashboard.js:1001-1029`, `cmd/server/templates/app.js:538-567`
- Impact: Login flow behavior differs between pages. The `admin.html` and `dashboard.html` both have embedded login screens in addition to the dedicated `/login` page.
- Fix approach: Consolidate to a single login mechanism.

### Missing Cache-Control Headers for Static Assets
- Issue: The `style()`, `adminStyle()`, and `serveJS()` functions in `cmd/server/main.go:142-173` serve CSS and JS files without any cache-control headers. Every page load re-fetches all assets.
- Files: `cmd/server/main.go:142-173`
- Impact: Increased server load and slower page loads. Static assets are embedded in the binary and never change during runtime.
- Fix approach: Add `Cache-Control: public, max-age=3600` headers and consider ETag support.

## Known Bugs

### adminUpdateUser Always Revokes Sessions on Role Change
- Issue: `adminUpdateUser` (`cmd/server/handlers.go:241,245`) always sets `last_token_at=now()` when updating a user, even if only the role changed (no password change). This revokes all existing sessions for that user.
- Files: `cmd/server/handlers.go:241,245`
- Trigger: A sysadmin changes a user's role via the admin panel.
- Workaround: None — role change always forces re-login.
- Fix: Only set `last_token_at` when the password actually changes.

### findFullProductByCode Makes Two Oracle Round Trips
- Issue: `findFullProductByCode` (`cmd/server/db.go:422-443`) first calls `findAddressByCode` to get the `SEQPRODUTO`, then immediately calls a second full query with the same `SEQPRODUTO`. This is an N+1 pattern on a single-product lookup.
- Files: `cmd/server/db.go:422-443`
- Impact: Every product detail lookup makes 2 Oracle queries when 1 would suffice. Oracle is already the bottleneck.
- Fix: Merge the two queries using a JOIN or subquery to get all fields in one round trip.

### adminPage User Stats May Be Incorrect
- Issue: The `adminPage` handler (`cmd/server/handlers.go:143-152`) loads users via `listUsers` which returns only `UserRow` (no password hash), but the admin template stats are computed client-side from the user list. The admin page only shows total/gerente/conferente counts, which may not reflect actual `sysadmin` count.
- Files: `cmd/server/handlers.go:143-152`, `cmd/server/templates/admin.html:90-112`
- Impact: Missing sysadmin counter makes admin stats incomplete.
- Fix: Add sysadmin count to the server-rendered stats or compute all counts server-side.

### printActivities Updates impresso Outside Transaction
- Issue: `printActivities` (`cmd/server/handlers.go:135-139`) sets `impresso=true` after loading details with a separate `UPDATE` per activity, outside any transaction.
- Files: `cmd/server/handlers.go:135-139`
- Impact: If the render at line 140 fails, activities are marked as printed but never rendered. This is a data integrity issue.
- Fix: Move the update into the same transaction used for loading, or only update after successful render.

### Oracle QueryRowContext Returns Dummy Row on Non-SELECT
- Issue: `QueryRowContext` on `OracleReader` (`cmd/server/db.go:22-27`) returns `o.db.QueryRowContext(ctx, "SELECT 1 FROM dual WHERE 1=0")` when the query is not read-only, which will return `sql.ErrNoRows` instead of clearly signaling the "read-only violation" error.
- Files: `cmd/server/db.go:22-27`
- Trigger: Any non-SELECT/WITH query attempted on the Oracle connection.
- Impact: The caller gets `sql.ErrNoRows` which is indistinguishable from "no data found" — a misleading error.
- Fix: Return `errors.New("oracle connection is read-only")` or create a sentinel error instead of executing a dummy query.

### Query Parameters in Oracle Use Positional Instead of Named
- Issue: Oracle queries in `db.go` use positional placeholders (`:1`, `:2`, `:3`) which make it easy to mismatch arguments. Example `findAddressByCode` at `db.go:379` uses `seqlocal, empresa, codigo` but the query has `:1, :2, :3` in a different order than one might expect.
- Files: `cmd/server/db.go:346,379,395,403,429,437`
- Impact: Parameter ordering is fragile. Adding/removing parameters requires careful counting and renumbering.
- Fix: Use named parameters or document the parameter mapping clearly.

### listActivities Sort Map Has Silent Fallback
- Issue: The `sortMap` in `listActivities` (`cmd/server/db.go:243`) silently falls back to `"a.data_fim"` if the sort column is not found. This means a client sending an invalid sort column gets no error or feedback.
- Files: `cmd/server/db.go:243-247`
- Impact: Hard to debug API client issues. Invalid sort parameters are silently swallowed.
- Fix: Return an error for invalid sort columns instead of silently falling back.

## Security Considerations

### Weak CSRF Protection for Non-API Form Posts
- Issue: The CSRF middleware (`cmd/server/auth.go:156-184`) only validates the `X-CSRF-Token` header for `/api/` routes. For form-based POST/PATCH/DELETE requests (non-API), it only checks the `Origin` header — which can be spoofed or may be absent entirely. There is no CSRF token validation for form submissions.
- Files: `cmd/server/auth.go:156-184`
- Risk: Cross-site request forgery on form endpoints like `/admin/users` (create user), `/admin/users/{id}` (update user), `/logout`.
- Current mitigation: `SameSite=StrictMode` on session cookie and `Origin` header check, but this is not sufficient against all attack vectors.
- Recommendations: Include a hidden CSRF token field in forms and validate it server-side. The `csrf_token` cookie is set but never validated against form POSTs.

### Admin Password Printed to Logs on First Run
- Issue: `seedAdmin` (`cmd/server/db.go:179`) logs the initial admin password to stdout: `log.Printf("⚠️  FIRST RUN — admin password: %s  (change immediately)", pass)`. This is visible in server logs, cloud logging services, and process output.
- Files: `cmd/server/db.go:179`
- Risk: The initial admin password is exposed to anyone who can access server logs.
- Current mitigation: Warning message tells user to change immediately.
- Recommendations: Write the password to stderr or a one-time file instead of server logs. Or force password change on first login.

### Session Token Lacks Nonce
- Issue: `makeToken` (`cmd/server/auth.go:62-77`) creates session tokens containing only `id`, `exp`, and `iat`. There is no random nonce, so the same user logging in twice at the same second gets identical tokens (same `id`, `exp`, `iat`).
- Files: `cmd/server/auth.go:62-77`
- Risk: Token replay is slightly easier without a unique identifier per session.
- Recommendations: Add a random nonce to the token payload so each session gets a unique token.

### No Rate Limiting on Admin User Creation
- Issue: The rate limiter (`cmd/server/utils.go:188-225`) is only applied to login endpoints. User creation and update endpoints (`POST /admin/users`, `POST /api/admin/users`) have no rate limiting, allowing an attacker to brute-force or flood user creation.
- Files: `cmd/server/handlers.go:159-186`, `cmd/server/api_handlers.go:191-214`
- Risk: Unauthenticated or compromised sysadmin session can create unlimited users.
- Recommendations: Apply rate limiting to user creation endpoints and consider IP-based blocking for repeated attempts.

### No Request Body Size Limits
- Issue: No endpoint uses `http.MaxBytesReader` to limit request body sizes. The `apiFinalizar` endpoint (`cmd/server/handlers.go:434`) accepts a JSON body that can contain arrays of products without size limits.
- Files: All POST/PATCH endpoints
- Risk: Memory exhaustion attack via large request bodies.
- Recommendations: Apply `http.MaxBytesReader` on all mutating endpoints.

### loginLimiter Uses r.RemoteAddr Which Is Unreliable Behind Proxies
- Issue: `loginPost` (`cmd/server/handlers.go:29`) and `apiLogin` (`cmd/server/handlers.go:262`) use `r.RemoteAddr` for rate limiting. Behind a reverse proxy, this will be the proxy's IP, not the client's IP.
- Files: `cmd/server/handlers.go:29,262`
- Risk: All clients behind the same proxy share a single rate limit bucket, effectively disabling the rate limiter in production.
- Recommendations: Check `X-Forwarded-For` or `X-Real-IP` headers when behind a proxy.

### Missing HSTS in Development Mode
- Issue: `securityHeaders` (`cmd/server/utils.go:233-234`) only sets HSTS in production mode. While acceptable, the CSP header allows `'unsafe-inline'` for both scripts and styles, which weakens XSS protection.
- Files: `cmd/server/utils.go:227-237`
- Risk: CSP `'unsafe-inline'` weakens the XSS protection significantly. Combined with the template rendering approach, stored XSS in product descriptions could execute scripts.
- Recommendations: Restrict CSP to nonces or hashes for inline scripts if possible.

### Cookie Secure Flag in Non-TLS Environments
- Issue: Cookies are set with `Secure: true` (`cmd/server/handlers.go:47`), which means they won't work over plain HTTP during development without TLS termination. The `AppEnv` check is not used to conditionally set Secure.
- Files: `cmd/server/handlers.go:47,66,284,295`, `cmd/server/auth.go:145,149`
- Risk: May cause confusing login failures in local development without HTTPS termination.
- Recommendations: Conditionally set `Secure` based on `AppEnv == "production"`.

## Performance Bottlenecks

### N+1 Oracle Queries in activityDetailsData
- Issue: `activityDetailsData` (`cmd/server/db.go:325-370`) calls `a.ora.QueryRowContext` for EACH product verification row, plus a `a.pg.QueryRowContext` for reincidencia count. If an activity has 500 products, this triggers 500 Oracle queries + 500 Postgres queries.
- Files: `cmd/server/db.go:325-370`
- Cause: Oracle lookup for description/stock/MDV and reincidencia count are per-row.
- Improvement path: Batch the Oracle queries using `WHERE seqproduto IN (...)`, and batch the reincidencia counts using a single `GROUP BY` query.

### Print Activities Loads Full Details Per Activity Sequentially
- Issue: `printActivities` (`cmd/server/handlers.go:123-141`) and `apiDashboardBulkDetails` (`cmd/server/api_handlers.go:321-356`) iterate over IDs and call `activityDetailsData` sequentially. Each call triggers the N+1 pattern above.
- Files: `cmd/server/handlers.go:123-141`, `cmd/server/api_handlers.go:321-356`
- Cause: Sequential per-activity loading with no parallelism.
- Improvement path: Load all activities' product data with a single query (`WHERE atividade_id = ANY($1)`) and batch the Oracle lookups.

### listActivities Uses Double Subquery Pattern
- Issue: `listActivities` (`cmd/server/db.go:254-267`) uses a subquery pattern where the inner query applies `ORDER BY` and `LIMIT`, and the outer query re-joins to get all rows. The inner sort must rewrite column aliases (line 253). With large datasets (100K+ activities), this can be slow.
- Files: `cmd/server/db.go:254-267`
- Cause: Need to support distinct activities with joined address rows while paginating.
- Improvement path: Consider `DISTINCT ON` or lateral joins. Evaluate query plan with `EXPLAIN ANALYZE`.

### No Pagination in Dashboard API
- Issue: `listActivities` is called with hardcoded limits of 50 (`handlers.go:78`), 50 (`handlers.go:84`), or 200 (`api_handlers.go:284`). There is no cursor or offset-based pagination. As the activities table grows, these queries will become slower.
- Files: `cmd/server/handlers.go:78,84`, `cmd/server/api_handlers.go:284`
- Cause: Simple KISS approach, no pagination requirements from the UI yet.
- Improvement path: Implement cursor-based pagination using `data_fim` + `id` as composite cursor.

### Oracle SQL Functions Called in SELECT Columns
- Issue: Queries in `findProductsByDescription` and `findFullProductByCode` call `CONSINCO.fBuscaPrecoAtualPdv(...)` and scalar subqueries with `LISTAGG` and `MAX` in the SELECT list. These functions execute per row and may be expensive on Oracle.
- Files: `cmd/server/db.go:396-399,429-433`
- Cause: Need to fetch computed pricing data from Oracle.
- Improvement path: Verify with Oracle DBA whether these functions are performant. Consider caching frequently accessed product data in Postgres.

## Fragile Areas

### isReadOnlySQL Security Gateway
- Issue: `isReadOnlySQL` (`cmd/server/db.go:29-65`) is the only guard preventing write operations to Oracle. It uses string matching with many edge cases (comments, string literals, nested statements). A bypass would allow data corruption in the source database.
- Files: `cmd/server/db.go:29-65`
- Why fragile: The function must correctly identify all possible SQL statement types. Comment stripping (`removeSQLComments`) is a custom parser that may miss edge cases. SQL injection in the query arguments could potentially bypass the guard if combined with certain query structures.
- Test coverage: There is a test (`main_test.go:21-49`) with 18 test cases, but edge cases like nested comments, multi-line strings, and exotic Oracle syntax are not covered.
- Safe modification: Always add test cases before modifying. Never widen allowed statements without review.

### removeSQLComments Custom Parser
- Issue: `removeSQLComments` (`cmd/server/db.go:67-109`) is a character-by-character SQL comment/string literal parser with non-standard escape handling (it handles `\` as escape inside quoted strings, which is not standard SQL).
- Files: `cmd/server/db.go:67-109`
- Why fragile: The parser handles `--` comments, `/* */` block comments, single/double quoted strings with `\` escape — but Oracle SQL uses `''` for string escapes, not `\`. This discrepancy could cause incorrect parsing.
- Test coverage: Only 5 test cases in `main_test.go:103-122`.
- Safe modification: Add comprehensive test cases covering Oracle-specific syntax before changing.

### Rate Limiter Is In-Memory, Single-Instance
- Issue: `rateLimiter` (`cmd/server/utils.go:188-225`) stores state in a `map[string]*rateEntry` protected by a `sync.Mutex`. This does not scale to multiple server instances and is lost on restart.
- Files: `cmd/server/utils.go:188-225`
- Why fragile: If the app runs behind a load balancer with multiple instances, each instance has its own rate limit state, effectively multiplying the allowed attempts by the instance count.
- Safe modification: Extract the rate limiter behind an interface so it can be replaced with a Redis-backed implementation.

### All Go Code Is Package main — No Encapsulation
- Issue: Every `.go` file declares `package main`. There are no interfaces, no exported types, and no package boundaries. Every function can access every other function and all global state.
- Files: All `.go` files in `cmd/server/`
- Why fragile: Any function can call any other function. There's no compile-time check enforcing layering. Introducing an interface requires refactoring all callers.
- Safe modification: Extract one package at a time, starting with `db.go` into `internal/db/`.

## Error Handling Gaps

### Scan Errors Silently Discarded
- Issue: Multiple row scan operations (e.g., `cmd/server/db.go:204`, `cmd/server/db.go:310`, `cmd/server/db.go:339`, `cmd/server/db.go:412`) use `if rows.Scan(...) == nil` — silently skipping rows that fail to scan, with no error logging.
- Files: `cmd/server/db.go:204,310,339,412`, `cmd/server/handlers.go:312,331`
- Impact: Data corruption goes unnoticed. A row with an incompatible type (e.g., NULL into non-NULL field) is silently dropped.
- Fix: Always log scan errors, even if the row is skipped.

### listFilterOptions Silently Returns Empty on Error
- Issue: `listFilterOptions` (`cmd/server/db.go:300-324`) has a nested `read` function that returns `nil` on query error (line 303-304). The final `FilterOptions` struct may have empty slices without any error indication.
- Files: `cmd/server/db.go:300-324`
- Impact: The dashboard may show no filter options when the database has a transient error, confusing users.
- Fix: Return the error from `listFilterOptions` and handle it in callers.

### apiAdminUserCreate Ignores JSON Decode Error
- Issue: `apiAdminUserCreate` (`cmd/server/api_handlers.go:197`) calls `json.NewDecoder(r.Body).Decode(&req)` without checking the return value. If the body is malformed, `req` remains zero-valued, and the function proceeds to validation with empty fields.
- Files: `cmd/server/api_handlers.go:197`
- Impact: Malformed JSON requests get a "Dados inválidos" error (line 198) instead of a proper "JSON inválido" message.
- Fix: Check the decode error and return a proper JSON error response.

### No Request Context Timeouts for POST Endpoints
- Issue: API query handlers for Oracle (`apiEmpresas`, `apiLocais`, etc.) use `context.WithTimeout`. However, mutation endpoints (`apiFinalizar`, `adminCreateUser`, etc.) use `r.Context()` directly without a timeout.
- Files: `cmd/server/handlers.go:434-535`, `cmd/server/handlers.go:159-186`, `cmd/server/api_handlers.go:191-214`
- Impact: A slow database write could hang the request indefinitely (until the HTTP server's `WriteTimeout` of 30s kicks in).
- Fix: Apply `context.WithTimeout` to mutation endpoint contexts.

## Duplicate Code

### User Creation Logic Duplicated
- Locations: `cmd/server/handlers.go:159-186` (form handler), `cmd/server/api_handlers.go:191-214` (API handler)
- Lines: ~28 lines each, ~56 total duplicated
- Variance: Form handler uses `r.FormValue()`, API handler uses JSON decode. Validation logic is identical (`username == ""`, `len(password) < 8`, `validRole`).

### User Update Logic Duplicated
- Locations: `cmd/server/handlers.go:208-251` (form handler), `cmd/server/api_handlers.go:216-258` (API handler)
- Lines: ~44 lines each, ~88 total duplicated
- Variance: The form handler renders HTML templates on error, the API handler returns JSON. Core logic (role update, password update with `last_token_at`) is duplicated.

### Oracle Product Query Duplication
- Locations: `cmd/server/db.go:394-404` (by description) and `cmd/server/db.go:428-438` (by code)
- Lines: Both query the same columns from the same tables with the same subqueries (`LISTAGG`, `MAX(marca)`, `fBuscaPrecoAtualPdv`)
- Variance: The WHERE clause differs (LIKE pattern vs. exact SEQPRODUTO match), but the SELECT list is identical.
- Fix: Extract a shared helper function that builds the common SELECT part.

## Configuration Risks

### SessionSecret Minimum Length Not Enforced at Build Time
- Issue: `loadConfig` (`cmd/server/utils.go:23-25`) requires `SESSION_SECRET` to be at least 32 characters, enforced at startup. However, the validation is via `log.Fatal` — if someone sets it to exactly 32 characters of all lowercase letters, HMAC-SHA256 still works but entropy is low.
- Files: `cmd/server/utils.go:19-25`
- Risk: Weak session secrets reduce the effectiveness of token HMAC signing.
- Recommendations: Document minimum entropy requirements. Consider checking character diversity.

### Two Oracle URL Configuration Paths
- Issue: Oracle connection can be configured via direct URL (`ORACLE_URL`) or via individual components (`ORACLE_HOST`, `ORACLE_PORT`, etc.). Both paths exist in `loadConfig` (`cmd/server/utils.go:31-42`).
- Files: `cmd/server/utils.go:31-42`
- Risk: Confusing behavior if both are set — `ORACLE_URL` wins, but other vars are silently ignored.
- Recommendations: Log a warning if both are set, and clarify precedence in documentation.

## Missing Test Coverage

### No Database Integration Tests
- Issue: All tests in `main_test.go` are unit tests that don't require a database. There are no tests for `listActivities`, `activityDetailsData`, `apiFinalizar`, or any database query function.
- Files: `cmd/server/main_test.go`
- Risk: Database query logic (especially the complex `listActivities` query construction) has no automated test coverage. Refactoring the SQL is high-risk.

### No Integration Tests for API Handlers
- Issue: API handlers like `apiFinalizar`, `apiEmpresas`, `apiProdutoEAN` have no test coverage at all. Request parsing, validation, and response formatting are untested.
- Risk: Adding new fields or changing validation rules requires manual testing.

### No Session/Auth Integration Tests
- Issue: The auth middleware (`requireRole`, `requireAPIRole`, `csrfMiddleware`) is tested only with mock `App` structs that have no database. Session validation (`currentUser`) is not tested.
- Risk: Auth bypass or session handling bugs would not be caught by tests.

---

*Concerns audit: 2026-06-08*
