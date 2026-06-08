# Testing Patterns

**Analysis Date:** 2026-06-08

## Test Framework

**Runner:** Go standard `testing` package (no external test runner).

**Assertion Library:** Go standard library only — `t.Fatal()`, `t.Fatalf()`, `t.Error()`, `t.Errorf()`. No testify, no assert, no require.

**HTTP Testing:** `net/http/httptest` package — `httptest.NewRequest()` and `httptest.NewRecorder()`.

**Run Commands:**
```bash
go test ./cmd/server              # Run all tests
go test -v ./cmd/server            # Verbose output
go test -run TestMakeToken ./cmd/server  # Run single test
go test -count=1 ./cmd/server      # Bypass cache
```

## Test File Organization

**Location:** Single test file at `cmd/server/main_test.go`. Same `package main` as source code (white-box testing).

**Naming:** One file per project: `main_test.go`. Test functions follow `TestCamelCase` convention.

**Structure:**
```
cmd/server/
├── main_test.go    # 323 lines, ~19 test functions
├── main.go
├── models.go
├── handlers.go
├── api_handlers.go
├── auth.go
├── db.go
└── utils.go
```

## Test Structure

**Suite Organization:** All tests are flat top-level functions. No `TestMain`, no subtests with `t.Run()`, no suite structs.

**Patterns:**

**Table-driven tests** for pure functions:
```go
// From main_test.go:21-49
func TestOracleReadOnlySQLGuard(t *testing.T) {
    cases := map[string]bool{
        "SELECT * FROM dual": true,
        " INSERT INTO x VALUES (1)": false,
        "": false,
        // ...
    }
    for query, want := range cases {
        if got := isReadOnlySQL(query); got != want {
            t.Fatalf("isReadOnlySQL(%q)=%v want %v", query, got, want)
        }
    }
}
```

**Table-driven with struct** for more complex cases:
```go
// From main_test.go:103-122
func TestRemoveSQLComments(t *testing.T) {
    cases := []struct {
        input    string
        expected string
    }{
        {"SELECT 1 FROM dual", "SELECT 1 FROM dual"},
        // ...
    }
    for _, c := range cases {
        got := removeSQLComments(c.input)
        // ...
    }
}
```

**Setup patterns:**
- Temporary files via `t.TempDir()` + `os.WriteFile()` (`TestLoadDotEnv`, `TestLoadDotEnvCRLF`)
- Environment variable setup/teardown with `os.Setenv`/`os.Unsetenv` (`TestLoadConfigRequiresSecret`, `TestConfigHasSessionTTL`)
- App struct construction inline with minimal config: `&App{cfg: Config{SessionSecret: ..., SessionTTL: ...}}` (`TestMakeToken`, `TestCSRFMiddleware*`, `TestSecurityHeaders`)

**Teardown patterns:** `defer os.Unsetenv(key)` after `os.Setenv` to restore environment.

## Mocking

**Framework:** None. No mock objects, no mock generation, no interface mocking.

**Approach:** 
- Pure functions are tested directly (no mocking needed)
- App methods that require database connections are NOT tested (no DB mocks, no test containers, no in-memory DB)
- Middleware and handlers that can run without DB dependencies are tested by constructing `App` with `nil` or minimal config
- The test `TestHealthCheckResponse` tests `healthCheck` with `loginLimiter` only (no pg, no ora)

**What is NOT mocked (and so is untested):**
- All Postgres queries (`listActivities`, `findUserByUsername`, `findUserByID`, etc.)
- All Oracle queries (`findAddressByCode`, `findProductsByDescription`, etc.)
- All page handlers that call DB methods
- All API handlers that call DB methods (except `healthCheck`)

## Fixtures and Factories

**No fixture files or factory functions.** Test data is created inline:
- `&App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!"), SessionTTL: 8 * time.Hour}}` — used in 5 different tests
- `UserRow{ID: 1, Username: "test", Role: "gerente"}` — inline struct literal
- `map[string]bool{"SELECT * FROM dual": true, ...}` — inline map literal

**Shared string constant:** The test session secret `"this-is-a-32-char-secret-for-testing!"` is duplicated in multiple tests (`TestMakeToken`, `TestCSRFMiddlewareSkipsLogin`, `TestCSRFMiddlewareBlocksAPIWithoutToken`, `TestCSRFMiddlewareAllowsWithToken`, `TestLoadConfigRequiresSecret`, `TestConfigHasSessionTTL`). Not defined as a package-level constant.

## Coverage

**Requirements:** None enforced. No `-cover` flag used in any documented command. No coverage threshold or CI gating.

**View Coverage:**
```bash
go test -cover ./cmd/server
go test -coverprofile=coverage.out ./cmd/server && go tool cover -html=coverage.out
```

**Estimated coverage gaps:**
- Zero coverage for all database access methods (`listActivities`, `findUserByUsername`, `findUserByID`, `listUsers`, `listFilterOptions`, `activityDetailsData`, `findAddressByCode`, `findProductsByDescription`, `findFullProductByCode`)
- Zero coverage for all page handlers (`home`, `loginPage`, `loginPost`, `dashboardPage`, `atividadesPage`, `adminPage`, `adminCreateUser`, `adminUpdateUser`, etc.)
- Zero coverage for API handlers (`apiEmpresas`, `apiLocais`, `apiProdutoEAN`, `apiFinalizar`, `apiAdminUsersList`, `apiDashboardActivities`, etc.)
- Zero coverage for Oracle methods (`OracleReader.QueryContext`, `OracleReader.QueryRowContext`)

**Areas with coverage:**
- Template parsing — `TestTemplatesParse`
- SQL read-only guard — `TestOracleReadOnlySQLGuard` (17 test cases)
- `.env` loading — `TestLoadDotEnv`, `TestLoadDotEnvCRLF`
- Role validation — `TestValidRole` (6 cases)
- SQL comment removal — `TestRemoveSQLComments` (5 cases)
- Random string generation — `TestRandomString`
- Utility functions — `TestFirstNonEmpty`
- Token creation — `TestMakeToken`
- Rate limiter — `TestRateLimiter` (success path + block + different IP)
- Security headers — `TestSecurityHeaders` (5 header assertions)
- UserRow no password leak — `TestUserRowNoPassword`
- Config loading — `TestLoadConfigRequiresSecret`, `TestConfigHasSessionTTL`
- CSRF middleware — `TestCSRFMiddlewareSkipsLogin`, `TestCSRFMiddlewareBlocksAPIWithoutToken`, `TestCSRFMiddlewareAllowsWithToken`
- Health check — `TestHealthCheckResponse`
- Context cancel check — `TestContextMissingCancel` (no-op assertion)

## Test Types

**Unit Tests:** All tests are unit tests in the traditional sense. No integration database tests. No mock-based tests.

**Integration Tests:** None. No database-backed tests, no test containers, no external service integration tests.

**E2E Tests:** None.

**Regression Tests:** `TestUserRowNoPassword` verifies the `UserRow` struct doesn't contain a password field — a security regression guard.

## Common Patterns

**Simple assertion pattern:**
```go
if got := firstNonEmpty("a", "b"); got != "a" {
    t.Fatalf("firstNonEmpty(a,b)=%q want a", got)
}
```

**HTTP handler testing pattern:**
```go
app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-long!")}}
handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}))
req := httptest.NewRequest(http.MethodPost, "/api/atividades/finalizar", nil)
req.Header.Set("X-CSRF-Token", "valid-token")
req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)
if rec.Code != http.StatusOK {
    t.Fatalf("expected 200, got %d", rec.Code)
}
```

**Environment variable test pattern:**
```go
key := "GO_SIMP_TEST_ENV"
if err := os.Unsetenv(key); err != nil {
    t.Fatal(err)
}
path := filepath.Join(t.TempDir(), ".env")
if err := os.WriteFile(path, []byte("GO_SIMP_TEST_ENV=\"ok\"\n"), 0o600); err != nil {
    t.Fatal(err)
}
if err := loadDotEnv(path); err != nil {
    t.Fatal(err)
}
if got := os.Getenv(key); got != "ok" {
    t.Fatalf("env=%q want ok", got)
}
```

**Edge case / negative testing patterns:**
- `TestOracleReadOnlySQLGuard` — tests empty string, DML statements, multi-statement, comments
- `TestRateLimiter` — tests block after 5 attempts, different IP still allowed
- `TestCSRFMiddlewareBlocksAPIWithoutToken` — tests missing CSRF token returns 403
- `TestValidRole` — tests invalid and empty roles

## Notable Observations

1. **No external test dependencies** — `go test` works without any additional tools, databases, or services. This makes tests fast and portable.

2. **No subtests** — `t.Run()` is never used. Each test case in table-driven tests uses `t.Fatalf()` which terminates the entire function on first failure, so only the first failing case is reported.

3. **No parallel tests** — `t.Parallel()` is never used. Tests run sequentially.

4. **No benchmarks or examples** — No `BenchmarkXxx` or `ExampleXxx` functions.

5. **Single test file** — All 19 test functions (~323 lines) are in one file. This is manageable for the current codebase size but may become unwieldy.

6. **Test isolation concern** — `TestLoadConfigRequiresSecret` and `TestConfigHasSessionTTL` both set and unset `SESSION_SECRET` and `POSTGRES_URL`. Running with `-count=1` is recommended since test ordering could matter if env vars leak.

---

*Testing analysis: 2026-06-08*
