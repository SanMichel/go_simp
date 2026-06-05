# Testing Patterns

**Analysis Date:** 2026-06-05

## Test Framework

**Runner:**
- Go's built-in `testing` package (Go 1.23.0)
- No test configuration file — no `jest.config.*`, `vitest.config.*`, or `.golangci.yml`
- Config: Not applicable — relies entirely on `go test`

**Assertion Library:**
- Go standard `testing.T` methods only — no `testify`, `assert`, or `require` libraries
- Assertions use: `t.Fatal()`, `t.Fatalf()`, `t.Error()`, `t.Errorf()`

**Run Commands:**
```bash
go test ./cmd/server              # Run all tests
go test ./cmd/server -v           # Verbose output
go test ./cmd/server -count=1     # Disable cache, force re-run
```

## Test File Organization

**Location:**
- Single file: `cmd/server/main_test.go` — all tests co-located in the same `package main`

**Naming:**
- Test functions: `TestCamelCase` — e.g., `TestTemplatesParse`, `TestOracleReadOnlySQLGuard`, `TestLoadDotEnv`, `TestValidRole`, `TestRandomString`, `TestMakeToken`, `TestRateLimiter`, `TestSecurityHeaders`, `TestCSRFMiddlewareSkipsLogin`
- No `_test.go` suffix files other than `main_test.go`
- No testdata directory used (the `.air.toml` references `testdata_dir = "testdata"` but directory doesn't exist)
- No test fixtures directory

**Structure:**
```
cmd/server/
├── main_test.go          # All tests (323 lines, 17 test functions)
├── main.go               # Application code
├── handlers.go
├── api_handlers.go
├── auth.go
├── db.go
├── models.go
├── utils.go
└── templates/            # Embedded templates (not test-related)
```

## Test Structure

**Suite Organization:**
```go
// Standard table-driven test — used for parameterized test cases
func TestOracleReadOnlySQLGuard(t *testing.T) {
    cases := map[string]bool{
        "SELECT * FROM dual":                     true,
        "\n\twith x as (select 1) select * from x": true,
        " INSERT INTO x VALUES (1)":              false,
        // ...
    }
    for query, want := range cases {
        if got := isReadOnlySQL(query); got != want {
            t.Fatalf("isReadOnlySQL(%q)=%v want %v", query, got, want)
        }
    }
}

// Table-driven with struct slice — used for removeSQLComments
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

**Patterns:**
- Setup: Direct struct literal construction of `App`, no test helpers or factory functions
  ```go
  app := &App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!"), SessionTTL: 8 * time.Hour}}
  ```
- Teardown: Implicit — no `t.Cleanup` or deferred teardown used; environment variables manually cleaned up:
  ```go
  defer os.Unsetenv("SESSION_SECRET")
  ```
- Assertion: `t.Fatal()` for hard failures, `t.Error()` for non-fatal assertions
- No subtests (`t.Run`) — every test function is a standalone test

## Mocking

**Framework:** None — no mocking library (no `gomock`, `testify/mock`, `mockgen`, etc.)

**Patterns:**
```go
// No mocking framework. Tests either:
// 1. Test pure functions with no dependencies (most common)
// 2. Construct App with nil/simple fields for state-free tests
app := &App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!")}}

// HTTP testing uses httptest
req := httptest.NewRequest(http.MethodGet, "/", nil)
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)

// No mock for database — DB-dependent functions are either:
// - Not tested (no integration/DB tests)
// - Covered indirectly by being unreachable in tests (nil pg causes segfault)
```

**What to Mock:**
- Not applicable — no mocking is done

**What NOT to Mock:**
- All tests avoid mocking entirely — pure functions are tested; functions with side effects (DB, templates) are mostly untested

## Fixtures and Factories

**Test Data:**
```go
// Inline literal — used for UserRow, Config, etc.
u := UserRow{ID: 1, Username: "test", Role: "gerente"}

// Config literals repeated across tests:
app := &App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!")}}
app := &App{cfg: Config{AppEnv: "production"}}
app := &App{loginLimiter: newRateLimiter()}
```

**Location:**
- No centralized fixtures or factories — all test data is defined inline within each test function
- No `testdata/` directory
- Environment files: temporary `.env` files are created via `os.WriteFile` inside `t.TempDir()`

## Coverage

**Requirements:** None enforced — no `-coverprofile` flag in commands, no CI coverage gate

**Current Coverage Level:**
- Pure utility functions are well-covered: `isReadOnlySQL`, `removeSQLComments`, `validRole`, `firstNonEmpty`, `randomString`, `loadDotEnv`, `newRateLimiter`/`allow`
- HTTP middleware partially covered: `securityHeaders`, `csrfMiddleware`
- Token generation covered: `makeToken`
- **Not tested:** All database functions, all handlers that call database, Oracle reader, migration, seeding, template rendering verification, rate limiter cleanup goroutine

**View Coverage:**
```bash
go test ./cmd/server -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Types

**Unit Tests:**
- 17 test functions, all unit tests
- Scope: isolated pure functions and HTTP middleware with `httptest`
- No integration tests, no database dependency, no Oracle dependency

**Integration Tests:**
- Not used
- No test container setup, no docker-compose test environment

**E2E Tests:**
- Not used
- No Playwright, Cypress, or Selenium

## Common Patterns

**Table-Driven Tests (Pure Functions):**
```go
func TestValidRole(t *testing.T) {
    if !validRole("sysadmin") { t.Error("sysadmin should be valid") }
    if !validRole("gerente")  { t.Error("gerente should be valid") }
    if validRole("admin")     { t.Error("admin should not be valid") }
}
```

**Environment Variable Setup/Teardown:**
```go
func TestLoadDotEnv(t *testing.T) {
    key := "GO_SIMP_TEST_ENV"
    if err := os.Unsetenv(key); err != nil { t.Fatal(err) }

    path := filepath.Join(t.TempDir(), ".env")
    if err := os.WriteFile(path, []byte("GO_SIMP_TEST_ENV=\"ok\"\n"), 0o600); err != nil {
        t.Fatal(err)
    }
    if err := loadDotEnv(path); err != nil { t.Fatal(err) }
    if got := os.Getenv(key); got != "ok" {
        t.Fatalf("env=%q want ok", got)
    }
}
```

**HTTP Middleware Testing:**
```go
func TestSecurityHeaders(t *testing.T) {
    app := &App{cfg: Config{AppEnv: "production"}}
    handler := app.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    h := rec.Header()
    if h.Get("X-Content-Type-Options") != "nosniff" {
        t.Error("missing X-Content-Type-Options: nosniff")
    }
}
```

**CSRF Middleware Testing:**
```go
// Pattern: construct minimal App, wrap handler, send request, check status
func TestCSRFMiddlewareAllowsWithToken(t *testing.T) {
    app := &App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!")}}
    handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    req := httptest.NewRequest(http.MethodPost, "/api/atividades/finalizar", nil)
    req.Header.Set("X-CSRF-Token", "valid-token")
    req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("POST /api with valid CSRF token should be OK, got %d", rec.Code)
    }
}
```

**Empty / Stub Tests:**
```go
// TestContextMissingCancel exists as a signature-only test (no logic)
func TestContextMissingCancel(t *testing.T) {
    // revokeSession with nil pg should not cause test issues
    // This test exists to confirm the function signature works
    // (cannot test with nil pg since it'd segfault — skip actual call)
}
```

## Test Coverage Gaps

**Untested area:** All database query functions
- Files: `cmd/server/db.go`
- Functions: `migrate`, `seedAdmin`, `findUserByUsername`, `findUserByID`, `listUsers`, `listActivities`, `listFilterOptions`, `activityDetailsData`, `findAddressByCode`, `findFullProductByCode`, `OracleReader.QueryContext`, `OracleReader.QueryRowContext`
- Risk: Database logic changes could break without detection

**Untested area:** All page handlers that render templates with data
- Files: `cmd/server/handlers.go`, `cmd/server/api_handlers.go`
- Functions: `loginPost`, `atividadesPage`, `dashboardPage`, `dashboardTable`, `activityDetails`, `adminCreateUser`, `adminUpdateUser`, `apiAdminUserCreate`, `apiAdminUserUpdate`, `apiFinalizar`, `apiDashboardBulkPrint`, etc.
- Risk: Template rendering, form validation, and database interaction paths untested

**Untested area:** Rate limiter background cleanup goroutine
- File: `cmd/server/utils.go:188-200`
- Risk: Goroutine leak or cleanup logic bug

---

*Testing analysis: 2026-06-05*
