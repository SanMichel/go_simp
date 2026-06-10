# Phase 7: Handler Decomposition — Research

**Researched:** 2026-06-09
**Domain:** Go handler refactoring — extracting service functions from overgrown handlers
**Confidence:** HIGH

## Summary

This phase decomposes the 4 largest HTTP handlers in `cmd/server/` into thin adapters (10-20 lines) backed by testable service functions. The mandatory target is `apiFinalizar` (101 lines, HAND-01); the next 3 largest are `adminUpdateUser` (43 lines), `apiAdminUserUpdate` (41 lines), and `apiDashboardBulkDetails` (35 lines). Business logic moves into `*App` methods that never receive `http.Request` or `http.ResponseWriter`. The existing test suite (2357 lines, 124+ test functions) serves as the equivalence gate — after each decomposition, all existing tests must pass unchanged.

**Primary recommendation:** Extract pure-data service methods on `*App` that accept `context.Context` + domain parameters + return domain objects/errors. Handlers become thin adapters that decode HTTP input, call service, encode HTTP output. Decompose one handler at a time, run full test suite between each.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| HAND-01 | `apiFinalizar` decomposed into handler + service functions | Handler is 101 lines (activity_handlers.go:147-247). Clear transaction boundary. Extract `finalizeActivity()` service function. |
| HAND-02 | Next 3 largest handlers by line count decomposed | Identified: `adminUpdateUser` (43 lines), `apiAdminUserUpdate` (41 lines), `apiDashboardBulkDetails` (35 lines). |
| HAND-03 | Business logic extracted into service functions that never touch HTTP types | Service functions use `context.Context` + domain params. `*App` receiver OK because App has no HTTP fields. |
| HAND-04 | Handlers become thin adapters (10-20 lines) | After extraction, each target handler is ~12-18 lines: decode, validate, call service, encode response. |
| HAND-05 | No behavior changes — test coverage proves equivalence | 124+ tests exist. Run full suite after each decomposition. Add unit tests for new service functions. |
</phase_requirements>

## Handler Size Analysis

### Full Handler Census (sorted by line count)

| Rank | Handler | File | Lines | Function Body Lines | Target? |
|------|---------|------|-------|-------------------|---------|
| 1 | `apiFinalizar` | activity_handlers.go | 147-247 | **101** | ✅ HAND-01 mandatory |
| 2 | `adminUpdateUser` | admin_handlers.go | 74-116 | **43** | ✅ HAND-02 |
| 3 | `apiAdminUserUpdate` | api_handlers.go | 214-256 | **41** | ✅ HAND-02 |
| 4 | `apiDashboardBulkDetails` | api_handlers.go | 316-352 | **35** | ✅ HAND-02 |
| 5 | `apiLogin` | handlers.go | 79-108 | 28 | — |
| 6 | `apiProdutosLocal` | activity_handlers.go | 101-129 | 27 | — |
| 7 | `adminCreateUser` | admin_handlers.go | 27-53 | 25 | — |
| 8 | `loginPost` | handlers.go | 22-45 | 22 | — |
| 9 | `apiAdminUserCreate` | api_handlers.go | 190-213 | 22 | — |
| 10 | `apiDashboardActivityDetails` | api_handlers.go | 292-315 | 22 | — |
| 11 | `printActivities` | dashboard_handlers.go | 56-77 | 22 | — |
| 12-22 | Remaining 11 handlers | various | various | 8-20 | — |

### Top 4 Candidates Detail

| Handler | Lines | File | Complexity | Business Logic Summary |
|---------|-------|------|------------|----------------------|
| `apiFinalizar` | 101 | activity_handlers.go:147-247 | HIGH — transaction, product comparison, divergence/rupture/replenishment classification | Parse request → Begin TX → Insert activity → Insert addresses → Compare read vs expected products → Classify divergences/ruptures/replenishments → Insert verification records → Commit → Return summary |
| `adminUpdateUser` | 43 | admin_handlers.go:74-116 | MEDIUM — permission checks, conditional password update | Parse form → Validate role → Self-edit guard → Sysadmin protection → Update with/without password → Render template |
| `apiAdminUserUpdate` | 41 | api_handlers.go:214-256 | MEDIUM — same logic as adminUpdateUser but JSON API | Parse JSON → Permission checks → Validate role → Update with/without password → Return JSON |
| `apiDashboardBulkDetails` | 35 | api_handlers.go:316-352 | LOW — loop with existing service calls | Parse IDs → Loop: call activityDetailsData → Map to API types → Return JSON |

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Input validation | Handler | Service | Parse + validate request shape in handler; validate business rules in service |
| Transaction management | Service | — | Service owns TX begin/commit/rollback so handler stays HTTP-focused |
| Product comparison logic | Service | — | Pure business logic — no HTTP concerns |
| Permission checks | Service | — | `adminUpdateUser` and `apiAdminUserUpdate` share same permission rules |
| Response encoding | Handler | — | Handler knows response format (JSON vs HTML template) |
| Bulk data aggregation | Service | — | Loop-with-details belongs in service, not handler |
| Session/auth checks | Middleware | — | Already handled by `requireRole` / `requireAPIRole` before handler runs |

## Standard Stack

### Core (exists, no new dependencies)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| net/http | stdlib (Go 1.23) | HTTP handlers | Project constraint — no external web frameworks |
| database/sql + pgx | stdlib + pgx v5 | Postgres access | Existing — all DB access goes through `*App.pg` |
| testing | stdlib | Test equivalence verification | Existing — 124+ tests already in place |

### Supporting

| Pattern | Purpose | When to Use |
|---------|---------|-------------|
| `*App` methods as service functions | Business logic container | Service needs DB access via `a.pg` or `a.ora` |
| Standalone functions | Pure computation | Logic that only depends on parameters (no DB) |
| Domain types in models.go | Service return types | Return rich result objects instead of raw maps |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Methods on `*App` | Sub-package `service/` | Project constraint: single `main` package. Sub-package creates import cycle risk. |
| DI framework | Manual `*App` wiring | Overkill for this project size. `App` struct is already the DI container. |
| `interface`-based services | Concrete `*App` methods | No consumer needs abstraction — single implementation. Interfaces add indirection with zero benefit. |

## Package Legitimacy Audit

> No external packages are introduced in this phase. The decomposition uses only the existing Go standard library (`net/http`, `database/sql`, `context`, `testing`) and existing project dependencies. No new packages to install or verify.

## Decomposition Strategy

### Strategy: Extract-Service-First (safe refactoring pattern)

The core pattern for each target handler:

```
1. CREATE service function in the same file (or domain file)
   - Signature: func (a *App) serviceName(ctx context.Context, params...) (result, error)
   - Never receives http.Request or http.ResponseWriter
   - Returns domain types + error (not HTTP responses)

2. WRITE unit tests for the new service function
   - Table-driven tests covering success + each error path
   - Use existing testApp(), testUser(), testDB() helpers

3. REWRITE handler to call service function
   - Handler: decode HTTP input → validate → call service → encode HTTP response
   - Target: 10-20 lines per handler after extraction

4. VERIFY equivalence
   - Run full test suite: go test ./cmd/server
   - All existing handler tests must pass unchanged
```

### Handler 1: `apiFinalizar` (101 lines → ~15 lines)

**Current (activity_handlers.go:147-247):**
```go
func (a *App) apiFinalizar(w http.ResponseWriter, r *http.Request, u *User) {
    // 1. Decode JSON (5 lines)
    // 2. Validate fields (8 lines)
    // 3. Default predio (3 lines)
    // 4. Begin TX (6 lines)
    // 5. Insert atividade (8 lines)
    // 6. Insert enderecos (7 lines)
    // 7. Build read/expected maps (16 lines)
    // 8. Classify divergences/ruptures/replenishments (38 lines)
    // 9. Insert verificacao records (6 lines)
    // 10. Commit (4 lines)
    // 11. Return JSON response (1 line)
}
```

**Proposed service function:**
```go
type FinalizarResult struct {
    ActivityID    int
    DataFim       time.Time
    Divergences   []map[string]any
    Ruptures      []map[string]any
    Replenishments []map[string]any
}

func (a *App) finalizeActivity(ctx context.Context, req finalizeReq, userID int) (*FinalizarResult, error) {
    // Owns TX begin/commit/rollback
    // Owns all INSERT logic
    // Owns product comparison + classification
    // Returns result + error
}
```

**After handler (~15 lines):**
```go
func (a *App) apiFinalizar(w http.ResponseWriter, r *http.Request, u *User) {
    var req finalizeReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        a.handleError(w, r, &AppError{Code: ..., Message: "JSON inválido", ...})
        return
    }
    result, err := a.finalizeActivity(r.Context(), req, u.ID)
    if err != nil {
        a.handleError(w, r, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{
        "success": true,
        "atividadeId": result.ActivityID,
        "dataFim": result.DataFim,
        "divergences": result.Divergences,
        "ruptures": result.Ruptures,
        "replenishments": result.Replenishments,
    })
}
```

### Handler 2: `adminUpdateUser` (43 lines → ~15 lines)

**Current (admin_handlers.go:74-116):**
- Gets current user from context
- Parses form, validates role
- Permission check (self-edit guard, sysadmin protection)
- Conditional DB update (with/without password)
- Renders template

**Proposed service function:**
```go
func (a *App) updateUserAdmin(ctx context.Context, currentUserID, targetID int, role, password string) (*UserRow, error) {
    // Permission checks
    // bcrypt hash (if password provided)
    // DB update
    // Return updated UserRow + error
}
```

**After handler (~15 lines):**
```go
func (a *App) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
    currentUser := r.Context().Value(ctxUser).(*User)
    id, _ := strconv.Atoi(r.PathValue("id"))
    r.ParseForm()
    role := r.FormValue("role")
    password := r.FormValue("password")

    u, err := a.updateUserAdmin(r.Context(), currentUser.ID, id, role, password)
    if err != nil {
        a.handleError(w, r, err)
        return
    }
    a.render(w, "user_row", map[string]any{"RowUser": u})
}
```

### Handler 3: `apiAdminUserUpdate` (41 lines → ~15 lines)

**Current (api_handlers.go:214-256):**
- Same permission/business logic as `adminUpdateUser` but with JSON I/O

**Key insight:** Reuse the same `updateUserAdmin` service function. The handler only differs in I/O format.

**After handler (~15 lines):**
```go
func (a *App) apiAdminUserUpdate(w http.ResponseWriter, r *http.Request, u *User) {
    id, _ := strconv.Atoi(r.PathValue("id"))
    var req struct { Role string; Password string }
    json.NewDecoder(r.Body).Decode(&req)

    _, err := a.updateUserAdmin(r.Context(), u.ID, id, req.Role, req.Password)
    if err != nil {
        a.handleError(w, r, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}
```

### Handler 4: `apiDashboardBulkDetails` (35 lines → ~15 lines)

**Current (api_handlers.go:316-352):**
- Parses JSON with IDs
- Loops: calls `activityDetailsData` for each ID
- Maps to FlatBundle
- Returns JSON

**Proposed service function:**
```go
func (a *App) bulkActivityDetails(ctx context.Context, ids []int) []FlatBundle {
    // Loop, call activityDetailsData, map to FlatBundle
    // Return slice (nil → empty slice for JSON "[]")
}
```

**After handler (~12 lines):**
```go
func (a *App) apiDashboardBulkDetails(w http.ResponseWriter, r *http.Request, u *User) {
    var req struct { Ids []int }
    json.NewDecoder(r.Body).Decode(&req)
    bundles := a.bulkActivityDetails(r.Context(), req.Ids)
    writeJSON(w, http.StatusOK, bundles)
}
```

## Service Layer Pattern

### Recommended Signature Convention

```go
// Service functions on *App (receiver OK because App has no HTTP fields)
// Signature: func (a *App) VerbNoun(ctx context.Context, domainParams...) (DomainResult, error)

// Pure computation (no DB access) as standalone functions
// Signature: func classifyProduct(read ReadProduct, expected ExpectedProduct) ProductClassification
```

### Key Rules

1. **No `http.Request` or `http.ResponseWriter`** in any service function signature. This is the hard boundary for HAND-03 compliance.

2. **`*App` receiver is permitted** because `App` struct (models.go:22-28) contains only `cfg`, `pg`, `ora`, `tpl`, `loginLimiter` — no HTTP types.

3. **Return domain types, not HTTP responses.** Return `(*FinalizarResult, error)` not `map[string]any`. The handler maps domain types to the response format.

4. **Error handling: return `*AppError` or wrapped error.** Service functions can return `*AppError` directly or wrap errors that `handleError` translates. The handler calls `handleError` with whatever the service returns.

5. **Context propagation.** Always accept `context.Context` as first argument. Pass it to DB queries. Never store context in the service struct.

### Domain File Organization

The extracted service functions should be placed in their **domain-aligned file**:

| File | Handler | Service Function |
|------|---------|-----------------|
| `activity_handlers.go` | `apiFinalizar` | `finalizeActivity` |
| `admin_handlers.go` | `adminUpdateUser` | `updateUserAdmin` |
| `api_handlers.go` | `apiAdminUserUpdate` | (reuses `updateUserAdmin`) |
| `api_handlers.go` | `apiDashboardBulkDetails` | `bulkActivityDetails` |

This keeps related functions together and follows the existing domain grouping (HAND-06).

## Thin Adapter Pattern

### What a 10-20 Line Handler Looks Like

```go
func (a *App) handlerName(w http.ResponseWriter, r *http.Request, u *User) {
    // 1. Parse request (2-5 lines) — decode JSON or parse form
    // 2. Basic validation (1-3 lines) — check required fields exist
    // 3. Call service (1 line) — one service call
    // 4. Handle error (2-4 lines) — if err != nil { handleError }
    // 5. Encode response (1-3 lines) — writeJSON or a.render
}
```

**Examples after decomposition:**

```go
// apiFinalizar — ~15 lines
func (a *App) apiFinalizar(w http.ResponseWriter, r *http.Request, u *User) {
    var req finalizeReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "JSON inválido", HTTPStatus: http.StatusBadRequest})
        return
    }
    result, err := a.finalizeActivity(r.Context(), req, u.ID)
    if err != nil {
        a.handleError(w, r, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{
        "success": true, "atividadeId": result.ActivityID,
        "dataFim": result.DataFim, "divergences": result.Divergences,
        "ruptures": result.Ruptures, "replenishments": result.Replenishments,
    })
}

// apiDashboardBulkDetails — ~12 lines
func (a *App) apiDashboardBulkDetails(w http.ResponseWriter, r *http.Request, u *User) {
    var req struct { Ids []int }
    json.NewDecoder(r.Body).Decode(&req)
    bundles := a.bulkActivityDetails(r.Context(), req.Ids)
    writeJSON(w, http.StatusOK, bundles)
}
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Service layer abstraction | Sub-package, interfaces, DI framework | Methods on `*App` | Single `main` package constraint. No consumers need abstraction. |
| Code generation for handler/service binding | Codegen tool | Manual extraction | Only 4 handlers to decompose. Codegen adds complexity for one-shot work. |
| Separate error type system for services | New service-layer error type | Existing `*AppError` | `*AppError` already carries code + message + HTTP status. Services return it directly. |
| Testing framework | Custom test harness | Existing `testApp()`, `httptest` | Phase 6 already established test infrastructure. Reuse it. |

**Key insight:** The project is small enough (single `main` package, ~15 handler files) that over-engineering the service layer would create more maintenance burden than the original fat handlers. The goal is pragmatic decomposition, not architectural purity.

## Test Strategy

### Equivalence Verification Protocol

The existing test suite is the gate for HAND-05 compliance. Protocol per handler:

1. **Before any changes:** `go test ./cmd/server` — must pass green.

2. **Create service function tests** — new table-driven tests that exercise the extracted business logic directly:
   - Test each error path (invalid input, permission denied, DB error)
   - Test each success path (happy path, edge cases)
   - These are new tests that add coverage, not part of the equivalence gate

3. **Refactor handler** — extract inline logic → service call.

4. **Run existing tests:** `go test ./cmd/server` — must still pass green.
   - `TestAPIFinalizarSuccess` (line 1819)
   - `TestAPIFinalizarMissingFields` (line 1848)
   - `TestAPIFinalizarWithReadProducts` (line 1861)
   - `TestAdminUpdateUser` (line 1962)
   - `TestAPIAdminUserUpdate` (line 2074)
   - `TestAPIDashboardBulkDetails` (line 2154)

5. **If tests fail:** Revert handler, fix service, re-run. Do not modify existing tests — they define the contract.

### Service Function Test Pattern

```go
func TestFinalizeActivity_Success(t *testing.T) {
    app := testApp(t)
    user := testUser(t, app, "test_finalize_svc", "conferente", "pass1234")
    ctx := context.Background()
    req := finalizeReq{Empresa: 1, SeqLocal: 1, Rua: "RUA A", Predio: []string{"PREDIO 1"}}

    result, err := app.finalizeActivity(ctx, req, user.ID)
    if err != nil {
        t.Fatalf("finalizeActivity: %v", err)
    }
    if result.ActivityID <= 0 {
        t.Error("expected positive activity ID")
    }
    // Verify DB state
    var count int
    app.pg.QueryRowContext(ctx, `SELECT COUNT(*) FROM atividades WHERE id=$1`, result.ActivityID).Scan(&count)
    if count != 1 {
        t.Errorf("expected 1 activity, got %d", count)
    }
}

func TestFinalizeActivity_MissingFields(t *testing.T) {
    app := testApp(t)
    user := testUser(t, app, "test_finalize_miss", "conferente", "pass1234")
    req := finalizeReq{SeqLocal: 1, Rua: "RUA A"} // Empresa = 0 → invalid
    _, err := app.finalizeActivity(context.Background(), req, user.ID)
    if err == nil {
        t.Fatal("expected error for missing Empresa")
    }
}
```

### Coverage Target

| Area | Before Phase 7 | After Phase 7 | Measurement |
|------|---------------|---------------|-------------|
| Handler tests | ~30 handler tests | Same (equivalence) | `go test -run Test.*Handler -v ./cmd/server` |
| Service function tests | 0 | ~8-12 new tests | New test functions |
| Total test functions | 124+ | 132-136+ | `go test -count=1 ./cmd/server` |
| Line coverage | ~58% (Phase 6 baseline) | ~62-65% | `go test -cover -count=1 ./cmd/server` |

## Dependency Analysis

### What Each Handler Depends On

| Handler | App Dependencies | Request Data | Service Needs |
|---------|-----------------|--------------|---------------|
| `apiFinalizar` | `a.pg` (Postgres), `a.handleError`, `writeJSON` | JSON body, `*User` from context | `context.Context`, `finalizeReq`, `userID` |
| `adminUpdateUser` | `a.pg`, `a.handleError`, `a.render`, `a.listUsers` | Form body, path param `{id}`, `*User` from context | `context.Context`, `currentUserID`, `targetID`, `role`, `password` |
| `apiAdminUserUpdate` | `a.pg`, `a.handleError`, `writeJSON` | JSON body, path param `{id}`, `*User` | Same as `adminUpdateUser` — REUSE possible |
| `apiDashboardBulkDetails` | `a.pg`, `a.ora`, `a.handleError`, `writeJSON`, `activityDetailsData`, `mapActivity`, `mapProduct` | JSON body with IDs | `context.Context`, `ids []int` |

### Dependency Injection Pattern

Service functions get dependencies through `*App` receiver:

```go
// Handler has access to everything
func (a *App) apiFinalizar(w http.ResponseWriter, r *http.Request, u *User) {
    result, err := a.finalizeActivity(r.Context(), req, u.ID)
    // ...
}

// Service uses a.pg (Postgres) via receiver
func (a *App) finalizeActivity(ctx context.Context, req finalizeReq, userID int) (*FinalizarResult, error) {
    tx, err := a.pg.BeginTx(ctx, nil)
    // ...
}
```

No structural changes to `App` or its initialization needed.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| **Accidental behavior change** | MEDIUM | HIGH — could corrupt data or return wrong response | Test equivalence gate (HAND-05). Run full suite after each decomposition. |
| **Transaction boundary leak** | LOW | MEDIUM — service returns before rollback | Verify `defer tx.Rollback()` is inside service, not handler. |
| **Error message change** | MEDIUM | LOW — user-facing Portuguese string changes | Existing tests assert specific error strings. Tests catch this. |
| **Shadowing bug** (e.g., `err` reuse) | LOW | MEDIUM — silent wrong behavior | Go vet catches shadowed variables. Run `go vet ./cmd/server`. |
| **Empty slice vs nil in response** | MEDIUM | LOW — JSON `null` vs `[]` | Existing tests check response body. Service returns empty slice explicitly (as `apiDashboardBulkDetails` already does for bundles). |
| **Test data cleanup interference** | LOW | MEDIUM — tests leave DB state | `cleanupTestData` runs in `t.Cleanup`. Each test gets fresh state. |
| **Cross-file import missing** | LOW | LOW — compile error, caught immediately | Go compiler catches missing types between files. |

### Rollback Plan

If a decomposition breaks tests:
1. `git checkout -- cmd/server/<file>` to restore the handler
2. Fix the service function
3. Re-run tests
4. Retry the handler rewrite

## Phase Boundary

### IN SCOPE

| Item | Rationale |
|------|-----------|
| Decompose `apiFinalizar` (101 lines) → service + thin handler | HAND-01 requirement |
| Decompose `adminUpdateUser` (43 lines) → service + thin handler | HAND-02 requirement |
| Decompose `apiAdminUserUpdate` (41 lines) → service + thin handler | HAND-02 requirement |
| Decompose `apiDashboardBulkDetails` (35 lines) → service + thin handler | HAND-02 requirement |
| Extract business logic into `*App` methods with no HTTP types | HAND-03 requirement |
| Reduce handlers to 10-20 lines each | HAND-04 requirement |
| Write service function unit tests | New tests for extracted logic |
| Run full test suite after each decomposition | HAND-05 — prove equivalence |
| Reuse shared service function where possible (`updateUserAdmin` reused by both admin handlers) | Code reuse opportunity |

### OUT OF SCOPE

| Item | Rationale |
|------|-----------|
| Decompose remaining handlers (#5-22) | Deferred to v2 (HAND-07 in REQUIREMENTS.md) |
| Create sub-packages (`service/`, `handler/`) | Out of scope per project constraints — single `main` package |
| Introduce DI framework or interfaces | Unnecessary — `*App` receiver provides sufficient decoupling |
| Change handler signatures | Must maintain same signatures for route registration |
| Add external dependencies | Not needed — all refactoring uses stdlib + existing deps |
| Rewrite tests or change existing test assertions | Tests are the equivalence gate — they must not change |
| Refactor non-handler code | DB functions, utils, templates are already clean |
| Add coverage for non-decomposed handlers | Phase 6 owns test coverage improvements |
| Add integration tests for Oracle-dependent code | Oracle not available in test environment |

## Common Pitfalls

### Pitfall 1: Transaction Boundary Confusion
**What goes wrong:** `apiFinalizar` owns TX begin/commit/rollback. If the service function starts its own TX but the handler has already started one (or vice versa), nesting issues arise.
**How to avoid:** The service function fully owns the TX lifecycle. The handler does not touch `a.pg.BeginTx` or `tx.Commit`/`tx.Rollback`. After extraction, all TX logic lives in `finalizeActivity`.
**Warning signs:** Handler code still references `a.pg.BeginTx` after extraction.

### Pitfall 2: Error Message Drift
**What goes wrong:** During extraction, error messages change slightly. Existing Portuguese-language assertions (`"Campos obrigatórios ausentes"`, `"Erro ao salvar atividade"`) fail.
**How to avoid:** Preserve exact error messages. Service functions should return `*AppError` with the same `Message` strings as the original handler.
**Warning signs:** Tests fail with `got "..." want "..."` on error message assertions.

### Pitfall 3: Empty Slice vs Nil Slice
**What goes wrong:** Service returns `nil` for divergences/ruptures/replenishments, but the original handler always populated these slices (even if empty). JSON response changes from `[]` to `null`.
**How to avoid:** Initialize all result slices with `make([]map[string]any, 0)` in the service function. Same for bundles in `apiDashboardBulkDetails`.
**Warning signs:** Test assertions use `len()` which passes for both nil and empty.

### Pitfall 4: Reusing Service Functions Incorrectly
**What goes wrong:** `adminUpdateUser` (HTML) and `apiAdminUserUpdate` (JSON) share `updateUserAdmin`. The HTML handler needs `a.listUsers` for the response template. If the service function tries to render templates, it violates HAND-03.
**How to avoid:** Service function returns domain data. Handlers decide how to render. `updateUserAdmin` returns `(*UserRow, error)` — HTML handler calls `a.listUsers` separately if needed.
**Warning signs:** Service function calls `a.render` or writes to `w`.

### Pitfall 5: Breaking Equivalence by Modifying Tests
**What goes wrong:** A test fails after decomposition, so the developer changes the test assertion instead of fixing the service.
**How to avoid:** Tests are the spec. Never modify existing tests during decomposition. If a test fails, the service is wrong, not the test.
**Warning signs:** Commit message says "fix test" without "fix service".

## Validation Architecture

> nyquist_validation is enabled (config.json: `"nyquist_validation": true`). Include this section.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | `testing` (Go stdlib) |
| Config file | none — Go testing convention |
| Quick run command | `go test ./cmd/server -count=1 -timeout 30s 2>&1 | tail -20` |
| Full suite command | `go test ./cmd/server -count=1 -v -timeout 60s 2>&1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| HAND-01 | `apiFinalizar` decomposed — existing tests pass | verification | `go test ./cmd/server -run TestAPIFinalizar -count=1` | ✅ `main_test.go` (3 tests) |
| HAND-02 | 3 largest handlers decomposed — existing tests pass | verification | `go test ./cmd/server -run "TestAdminUpdateUser|TestAPIAdminUserUpdate|TestAPIDashboardBulkDetails" -count=1` | ✅ `main_test.go` |
| HAND-03 | Service functions never touch HTTP types | static analysis | `grep -n 'http\.Request\|http\.ResponseWriter' cmd/server/*.go \| grep -v '_test.go' \| grep 'func.*App.*('` | N/A — grep check |
| HAND-04 | Handlers are 10-20 lines | static analysis | `awk '/^func.*App.*\(w http/{f=1;c=0} f{c++} /^}$/{if(f) print c-3,"lines"; f=0}' cmd/server/handlers.go cmd/server/activity_handlers.go` | N/A — manual check |
| HAND-05 | No behavior changes — test equivalence | CI gate | Full suite — all tests must pass with zero modifications | ✅ `main_test.go` |

### New Tests to Add

| Test | What It Covers | File |
|------|---------------|------|
| `TestFinalizeActivity_Success` | Happy path: activity created, products processed | `main_test.go` |
| `TestFinalizeActivity_MissingFields` | Validation: missing empresa/rua/seqlocal | `main_test.go` |
| `TestFinalizeActivity_DivergenceDetection` | Business logic: product status classification | `main_test.go` |
| `TestUpdateUserAdmin_Success` | Permission checks + DB update | `main_test.go` |
| `TestUpdateUserAdmin_SelfEditBlocked` | Permission: can't edit own user | `main_test.go` |
| `TestUpdateUserAdmin_SysadminProtection` | Permission: non-sysadmin can't edit sysadmin | `main_test.go` |
| `TestUpdateUserAdmin_WithPassword` | bcrypt + update with password change | `main_test.go` |
| `TestBulkActivityDetails_Success` | Bulk fetch with multiple IDs | `main_test.go` |
| `TestBulkActivityDetails_EmptyIDs` | Edge case: empty ID list | `main_test.go` |
| `TestBulkActivityDetails_PartialFail` | Edge case: some IDs don't exist | `main_test.go` |

### Sampling Rate

- **Per task commit:** `go test ./cmd/server -count=1 -timeout 30s 2>&1 | tail -5`
- **Per wave merge:** `go test ./cmd/server -count=1 -v -timeout 60s 2>&1`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] New test functions for `finalizeActivity`, `updateUserAdmin`, `bulkActivityDetails` — must be written
- [ ] No additional framework install needed — existing `testing` + `testhelper.go` infrastructure covers all

## Security Domain

> `security_enforcement` is enabled (absent from config.json — treat as enabled per instructions).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Handled by existing `requireRole`/`requireAPIRole` middleware — not modified in this phase |
| V3 Session Management | No | Session tokens via `currentUser` — not modified |
| V4 Access Control | Yes | Permission checks moved to service functions — must preserve exact same guards |
| V5 Input Validation | Yes | Move validation to service functions — use existing `Validator` type |
| V6 Cryptography | No | bcrypt already in handler code — moves with the permission logic to service |

### Known Threat Patterns for Go net/http + Postgres

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SQL injection via handler params | Tampering | Parameterized queries via pgx `$1`, `$2` — preserved in extraction |
| Permission escalation via request manipulation | Elevation of Privilege | Service function repeats permission checks from handler — same logic, same safety |
| Mass assignment via extra JSON fields | Tampering | Struct-based decoding ignores unknown fields — brittle, but same risk as existed |

### Security Equivalence Check

After decomposition, verify for each handler:
1. The permission guards that existed in the handler body are still present in the service function
2. The same fields are validated (no relaxation of validation)
3. The same DB queries execute with the same parameters

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The existing test suite (124+ tests) provides sufficient coverage to detect behavioral regressions | Test Strategy | Low — covers the 4 target handlers directly. If missing coverage exists, equivalence gate still catches most regressions. |
| A2 | `adminUpdateUser` and `apiAdminUserUpdate` can share the same `updateUserAdmin` service function | Decomposition Strategy | Low — both handlers validate same rules (self-edit, sysadmin, role). If response needs differ, can add a parameter or keep separate. |
| A3 | The `apiDashboardBulkDetails` existing inline type `FlatBundle` can be extracted to models.go | Dependency Analysis | LOW — this is a minor organizational change. If omitted, the type stays in api_handlers.go. |

**If this table is empty:** N/A — table populated with LOW risk items.

## Open Questions

1. **Should `updateUserAdmin` return `*UserRow` for HTML rendering or just `error` for JSON?**
   - What we know: HTML handler needs `*UserRow` to render template; JSON handler only needs `error`
   - Recommendation: Return `(*UserRow, error)`. JSON handler ignores the row return. Single unified service.

2. **Should `bulkActivityDetails` be a standalone function or `*App` method?**
   - What we know: It calls `a.activityDetailsData` (an `*App` method), so it needs access to `*App`.
   - Recommendation: `*App` method for now. If future phases extract DB layer, this changes.

3. **What about `apiDashboardBulkPrint` (12 lines)?**
   - What we know: It's already small but does a DB write inline.
   - Recommendation: Out of scope for this phase (HAND-02 targets the 3 after apiFinalizar). Consider in v2.

## Environment Availability

> This phase has no external dependencies. All work is code refactoring within existing Go files.

**Step 2.6: SKIPPED** (no external dependencies — code-only refactoring within existing Go stdlib + pgx/go-ora ecosystem)

## Sources

### Primary (HIGH confidence)

- Project codebase: `cmd/server/` — all handler source files read and analyzed
- Project REQUIREMENTS.md — HAND-01 through HAND-05 requirement definitions
- PROJECT.md — single `main` package constraint, no external frameworks

### Secondary (MEDIUM confidence)

- `main_test.go` (2357 lines) — verified existing test coverage for all 4 target handlers
- `testhelper.go` — verified test infrastructure (`testApp`, `testUser`, `testToken`, `cleanupTestData`)

### Tertiary (LOW confidence)

- None — all findings verified against actual codebase

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified by reading all Go source files and AGENTS.md
- Architecture: HIGH — extracted from actual handler code analysis
- Pitfalls: HIGH — derived from Go refactoring patterns and specific project constraints
- Test strategy: HIGH — verified existing test infrastructure in main_test.go and testhelper.go

**Research date:** 2026-06-09
**Valid until:** Stable — Go 1.23 stdlib doesn't change
