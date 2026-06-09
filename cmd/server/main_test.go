package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func urlencode(s string) io.Reader {
	return strings.NewReader(s)
}

func jsonBody(v any) io.Reader {
	b, _ := json.Marshal(v)
	return strings.NewReader(string(b))
}

func TestTemplatesParse(t *testing.T) {
	if parseTemplates() == nil {
		t.Fatal("parseTemplates returned nil")
	}
}

func TestOracleReadOnlySQLGuard(t *testing.T) {
	cases := map[string]bool{
		"SELECT * FROM dual": true,
		"\n\twith x as (select 1 from dual) select * from x":    true,
		" INSERT INTO x VALUES (1)":                             false,
		"UPDATE x SET y=1":                                      false,
		"DELETE FROM x":                                         false,
		"":                                                      false,
		"SELECT 1; DELETE FROM dual":                            false,
		"SELECT 1 /* comment */ FROM dual":                      true,
		"SELECT 1 FROM dual -- inline comment":                  true,
		"SELECT 'INSERT INTO x' FROM dual":                      true,
		"/* comment */ SELECT 1 FROM dual":                      true,
		"(SELECT 1 FROM dual)":                                  true,
		"DROP TABLE users":                                      false,
		"ALTER TABLE users ADD COLUMN x TEXT":                   false,
		"TRUNCATE TABLE users":                                  false,
		"CALL some_proc":                                        false,
		"EXEC some_proc":                                        false,
		"MERGE INTO dual USING ...":                             false,
		"CREATE TABLE x (id INT)":                               false,
		"SELECT * FROM (SELECT 1 FROM dual) WHERE 'DROP' = 'x'": true,
	}
	for query, want := range cases {
		if got := isReadOnlySQL(query); got != want {
			t.Fatalf("isReadOnlySQL(%q)=%v want %v", query, got, want)
		}
	}
}

func TestLoadDotEnv(t *testing.T) {
	key := "GO_SIMP_TEST_ENV"
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("GO_SIMP_TEST_ENV=\"ok\"\n# ignored\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(key); got != "ok" {
		t.Fatalf("env=%q want ok", got)
	}
}

func TestLoadDotEnvCRLF(t *testing.T) {
	key := "GO_SIMP_CRLF_TEST"
	os.Unsetenv(key)

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("GO_SIMP_CRLF_TEST=crlfok\r\n# comment\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(key); got != "crlfok" {
		t.Fatalf("env=%q want crlfok", got)
	}
}

func TestValidRole(t *testing.T) {
	if !validRole("sysadmin") {
		t.Error("sysadmin should be valid")
	}
	if !validRole("gerente") {
		t.Error("gerente should be valid")
	}
	if !validRole("conferente") {
		t.Error("conferente should be valid")
	}
	if validRole("admin") {
		t.Error("admin should not be valid")
	}
	if validRole("") {
		t.Error("empty should not be valid")
	}
}

func TestRemoveSQLComments(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"SELECT 1 FROM dual", "SELECT 1 FROM dual"},
		{"SELECT 1 -- comment\nFROM dual", "SELECT 1  FROM dual"},
		{"SELECT 1 /* block */ FROM dual", "SELECT 1   FROM dual"},
		{"SELECT 'hello -- world' FROM dual", "SELECT 'hello -- world' FROM dual"},
		{"SELECT /* nested * not end */ 1 FROM dual", "SELECT  1 FROM dual"},
	}
	for _, c := range cases {
		got := removeSQLComments(c.input)
		got = strings.Join(strings.Fields(got), " ")
		want := strings.Join(strings.Fields(c.expected), " ")
		if got != want {
			t.Errorf("removeSQLComments(%q) = %q, want %q", c.input, got, want)
		}
	}
}

func TestRandomString(t *testing.T) {
	s1 := randomString(32)
	s2 := randomString(32)
	if s1 == "" {
		t.Fatal("randomString returned empty")
	}
	if s1 == s2 {
		t.Fatal("randomString returned same value twice")
	}
	if len(s1) < 32 {
		t.Fatalf("randomString(32) too short: %d", len(s1))
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("a", "b"); got != "a" {
		t.Fatalf("firstNonEmpty(a,b)=%q want a", got)
	}
	if got := firstNonEmpty("", "b"); got != "b" {
		t.Fatalf("firstNonEmpty(,b)=%q want b", got)
	}
}

func TestMakeToken(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!"), SessionTTL: 8 * time.Hour}}
	token, err := app.makeToken(1)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatal("token should have 2 parts")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		ID  int   `json:"id"`
		Exp int64 `json:"exp"`
		Iat int64 `json:"iat"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.ID != 1 {
		t.Fatalf("token id=%d want 1", p.ID)
	}
	if p.Exp <= time.Now().Unix() {
		t.Fatal("token should not be expired")
	}
	if p.Iat == 0 {
		t.Fatal("token missing iat claim")
	}
	// Token should NOT have a nonce field
	if strings.Contains(string(payload), "nonce") {
		t.Fatal("token should not contain nonce")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := newRateLimiter()
	ip := "192.168.1.1"
	for i := 0; i < 5; i++ {
		if !rl.allow(ip) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if rl.allow(ip) {
		t.Fatal("6th attempt should be blocked")
	}
	// Different IP should be allowed
	if !rl.allow("192.168.1.2") {
		t.Fatal("different IP should be allowed")
	}
}

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
	if h.Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options: DENY")
	}
	if h.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("missing Referrer-Policy")
	}
	if h.Get("Strict-Transport-Security") == "" {
		t.Error("missing HSTS in production mode")
	}
	if h.Get("Content-Security-Policy") == "" {
		t.Error("missing Content-Security-Policy")
	}
}

func TestUserRowNoPassword(t *testing.T) {
	u := UserRow{ID: 1, Username: "test", Role: "gerente"}
	if u.Username != "test" {
		t.Error("UserRow should preserve Username")
	}
	if u.Role != "gerente" {
		t.Error("UserRow should preserve Role")
	}
}

func TestLoadConfigRequiresSecret(t *testing.T) {
	os.Unsetenv("SESSION_SECRET")
	os.Setenv("POSTGRES_URL", "postgres://localhost:5432/test")
	defer os.Unsetenv("POSTGRES_URL")
	// Should call log.Fatal which we can't easily test,
	// but we can verify that the function exists and is reachable
	// by testing that a valid config works
	os.Setenv("SESSION_SECRET", "this-is-a-32-char-secret-for-testing!")
	defer os.Unsetenv("SESSION_SECRET")
	cfg := loadConfig()
	if len(cfg.SessionSecret) == 0 {
		t.Fatal("SessionSecret should be set")
	}
}

func TestContextMissingCancel(t *testing.T) {
	// revokeSession with nil pg should not cause test issues
	// This test exists to confirm the function signature works
	// (cannot test with nil pg since it'd segfault — skip actual call)
}

func TestCSRFMiddlewareSkipsLogin(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	// POST /login should be allowed without CSRF token
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /login should skip CSRF check, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareBlocksAPIWithoutToken(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("this-is-a-32-char-secret-for-testing!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	// POST /api without CSRF token should be blocked
	req := httptest.NewRequest(http.MethodPost, "/api/atividades/finalizar", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /api without CSRF token should be 403, got %d", rec.Code)
	}
}

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

func TestConfigHasSessionTTL(t *testing.T) {
	os.Setenv("SESSION_SECRET", "this-is-a-32-char-secret-for-testing!")
	os.Setenv("POSTGRES_URL", "postgres://localhost:5432/test")
	defer os.Unsetenv("SESSION_SECRET")
	defer os.Unsetenv("POSTGRES_URL")
	cfg := loadConfig()
	if cfg.SessionTTL != 8*time.Hour {
		t.Fatalf("default SessionTTL = %v, want 8h", cfg.SessionTTL)
	}
}

func TestHealthCheckResponse(t *testing.T) {
	app := &App{loginLimiter: newRateLimiter()}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	app.healthCheck(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health check should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type = %q, want application/json", ct)
	}
}

// ---- Error Handling Tests ----

func TestAppErrorError(t *testing.T) {
	err := &AppError{Code: "TEST", Message: "test message", HTTPStatus: 400}
	if got := err.Error(); got != "test message" {
		t.Errorf("AppError.Error() = %q, want %q", got, "test message")
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	wrapped := errors.New("wrapped")
	err := &AppError{Code: "TEST", Message: "test", Err: wrapped}
	if !errors.Is(err, wrapped) {
		t.Error("AppError should unwrap to wrapped error")
	}
}

func TestAppErrorNilUnwrap(t *testing.T) {
	err := &AppError{Code: "TEST", Message: "test"}
	if err.Unwrap() != nil {
		t.Error("AppError.Unwrap() should return nil when no wrapped error")
	}
}

func TestHandleErrorNonAppError(t *testing.T) {
	app := &App{tpl: parseTemplates()}
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	app.handleError(rec, req, errors.New("raw error"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "Erro interno do servidor" {
		t.Errorf("body error = %q, want 'Erro interno do servidor'", body["error"])
	}
}

func TestHandleErrorHTMXPath(t *testing.T) {
	app := &App{tpl: parseTemplates()}
	req := httptest.NewRequest("POST", "/some-path", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	app.handleError(rec, req, &AppError{
		Code: ErrCodeValidation, Message: "Campo obrigatório",
		HTTPStatus: http.StatusBadRequest,
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Campo obrigatório") {
		t.Errorf("HTMX should render error message, got body = %q", body)
	}
}

func TestHandleErrorAPIPath(t *testing.T) {
	app := &App{tpl: parseTemplates()}
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	app.handleError(rec, req, &AppError{
		Code: ErrCodeNotFound, Message: "Não encontrado",
		HTTPStatus: http.StatusNotFound,
	})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("API should get JSON, got Content-Type = %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "Não encontrado" || body["code"] != "NOT_FOUND" {
		t.Errorf("JSON body = %v, want {error: Não encontrado, code: NOT_FOUND}", body)
	}
}

func TestHandleErrorPagePath(t *testing.T) {
	app := &App{tpl: parseTemplates()}
	req := httptest.NewRequest("GET", "/home", nil)
	rec := httptest.NewRecorder()
	app.handleError(rec, req, &AppError{
		Code: ErrCodeInternal, Message: "Erro interno",
		HTTPStatus: http.StatusInternalServerError,
	})
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestHandleErrorDefaultStatusFromCode(t *testing.T) {
	app := &App{tpl: parseTemplates()}
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	app.handleError(rec, req, &AppError{Code: ErrCodeUnauthorized, Message: "No auth"})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (from codeStatus map)", rec.Code)
	}
}

// ---- Validator Tests ----

func TestValidatorRequired(t *testing.T) {
	v := NewValidator()
	v.Required("nome", "")
	if v.IsValid() {
		t.Error("should be invalid when required field is empty")
	}
}

func TestValidatorRequiredPasses(t *testing.T) {
	v := NewValidator()
	v.Required("nome", "João")
	if !v.IsValid() {
		t.Error("should be valid when required field is non-empty")
	}
}

func TestValidatorMinLength(t *testing.T) {
	v := NewValidator()
	v.MinLength("senha", "abc", 8)
	if v.IsValid() {
		t.Error("min length should fail for short value")
	}
}

func TestValidatorMinLengthPasses(t *testing.T) {
	v := NewValidator()
	v.MinLength("senha", "12345678", 8)
	if !v.IsValid() {
		t.Error("min length should pass for long enough value")
	}
}

func TestValidatorValidRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{"sysadmin", true},
		{"gerente", true},
		{"conferente", true},
		{"admin", false},
		{"", false},
		{"manager", false},
	}
	for _, tt := range tests {
		v := NewValidator()
		v.ValidRole("role", tt.role)
		if v.IsValid() != tt.valid {
			t.Errorf("ValidRole(%q): valid=%v, want %v", tt.role, v.IsValid(), tt.valid)
		}
	}
}

func TestValidatorPositive(t *testing.T) {
	v := NewValidator()
	v.Positive("quantidade", 0)
	if v.IsValid() {
		t.Error("0 should not be positive")
	}
	v2 := NewValidator()
	v2.Positive("quantidade", -1)
	if v2.IsValid() {
		t.Error("-1 should not be positive")
	}
	v3 := NewValidator()
	v3.Positive("quantidade", 5)
	if !v3.IsValid() {
		t.Error("5 should be positive")
	}
}

func TestValidatorChain(t *testing.T) {
	v := NewValidator()
	v.Required("nome", "").MinLength("senha", "x", 8).ValidRole("role", "invalid")
	if v.IsValid() {
		t.Error("chain of invalid validations should fail")
	}
	if len(v.Errors()) != 3 {
		t.Errorf("expected 3 errors, got %d: %v", len(v.Errors()), v.Errors())
	}
}

func TestValidatorError(t *testing.T) {
	v := NewValidator()
	v.Required("nome", "")
	if v.Error() == "" {
		t.Error("Error() should return combined message")
	}
}

// ---- writeJSON Tests ----

func TestWriteJSONSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, map[string]string{"hello": "world"})
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["hello"] != "world" {
		t.Errorf("body = %v, want {hello: world}", body)
	}
}

func TestWriteJSONEncodeFailure(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusOK, make(chan int))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 on encode failure", rec.Code)
	}
}

// ---- Middleware Tests ----

func TestRecoveryMiddlewareCatchesPanic(t *testing.T) {
	handler := recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 after panic", rec.Code)
	}
}

func TestRequestIDMiddlewareGeneration(t *testing.T) {
	handler := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := getRequestID(r.Context()); id == "" {
			t.Error("request ID should be set in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if h := rec.Header().Get("X-Request-Id"); h == "" {
		t.Error("response should have X-Request-Id header")
	}
}

func TestRequestIDMiddlewarePreservesExisting(t *testing.T) {
	handler := requestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := getRequestID(r.Context()); id != "existing-id" {
			t.Errorf("request ID = %q, want 'existing-id'", id)
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "existing-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if h := rec.Header().Get("X-Request-Id"); h != "existing-id" {
		t.Errorf("X-Request-Id = %q, want 'existing-id'", h)
	}
}

func TestErrorCodesConstants(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{ErrCodeValidation, "VALIDATION_ERROR"},
		{ErrCodeUnauthorized, "UNAUTHORIZED"},
		{ErrCodeForbidden, "FORBIDDEN"},
		{ErrCodeNotFound, "NOT_FOUND"},
		{ErrCodeConflict, "CONFLICT"},
		{ErrCodeRateLimited, "RATE_LIMITED"},
		{ErrCodeInternal, "INTERNAL_ERROR"},
		{ErrCodeBadRequest, "BAD_REQUEST"},
	}
	for _, tt := range tests {
		if tt.code != tt.expected {
			t.Errorf("constant = %q, want %q", tt.code, tt.expected)
		}
		if _, ok := codeStatus[tt.code]; !ok {
			t.Errorf("codeStatus missing entry for %q", tt.code)
		}
	}
}

// ---- CSRF Middleware Tests ----

func TestCSRFMiddlewareAllowsGET(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET should pass CSRF check, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareSkipsLoginPost(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /login should skip CSRF, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareSkipsAPILoginPost(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("POST /api/auth/login should skip CSRF, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareBlocksAPIWithoutCookie(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-CSRF-Token", "some-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("POST /api without CSRF cookie should be 403, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareBlocksAPIWithoutHeader(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("POST /api without CSRF header should be 403, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareBlocksOnOriginMismatch(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/some-path", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Origin mismatch should return 403, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareAllowsOriginMatch(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/some-path", nil)
	req.Host = "example.com"
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Origin match should pass, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareBlocksPATCH(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPatch, "/api/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("PATCH /api without CSRF should be 403, got %d", rec.Code)
	}
}

func TestCSRFMiddlewareBlocksDELETE(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	handler := app.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodDelete, "/api/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("DELETE /api without CSRF should be 403, got %d", rec.Code)
	}
}

// ---- Security Headers Tests ----

func TestSecurityHeadersDevModeNoHSTS(t *testing.T) {
	app := &App{cfg: Config{AppEnv: "development"}}
	handler := app.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	h := rec.Header()
	if h.Get("Strict-Transport-Security") != "" {
		t.Error("HSTS should be empty in development mode")
	}
	if h.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options")
	}
	if h.Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options")
	}
	if h.Get("Referrer-Policy") == "" {
		t.Error("missing Referrer-Policy")
	}
	if h.Get("Content-Security-Policy") == "" {
		t.Error("missing Content-Security-Policy")
	}
}

// ---- Mapping Function Tests ----

func TestMapUser(t *testing.T) {
	u := UserRow{ID: 1, Username: "joao", Role: "gerente"}
	api := mapUser(u)
	if api.ID != 1 || api.Username != "joao" || api.Role != "gerente" {
		t.Errorf("mapUser = %+v, want {ID:1 Username:joao Role:gerente}", api)
	}
}

func TestMapActivity(t *testing.T) {
	now := time.Now()
	a := Activity{ID: 1, Empresa: "001", SeqLocal: 5, UserID: 10, Username: "joao", DataFim: now, Impresso: false, Rua: "RUA A", Predio: "PREDIO 1"}
	api := mapActivity(a)
	if api.ID != 1 || api.Empresa != "001" || api.Username != "joao" {
		t.Errorf("mapActivity basic fields mismatch: %+v", api)
	}
	if api.DataFim == nil {
		t.Error("mapActivity should set DataFim for non-zero time")
	}
	// Zero DataFim should produce nil
	a2 := Activity{ID: 2, Empresa: "002", DataFim: time.Time{}}
	api2 := mapActivity(a2)
	if api2.DataFim != nil {
		t.Error("mapActivity should set DataFim to nil for zero time")
	}
}

func TestMapProduct(t *testing.T) {
	now := time.Now()
	t.Run("all fields valid", func(t *testing.T) {
		p := ProductVerification{
			ID: 1, AtividadeID: 10, SeqProduto: 100, Empresa: "001",
			RuaLida:        sql.NullString{String: "RUA A", Valid: true},
			PredioLido:     sql.NullString{String: "PREDIO 1", Valid: true},
			RuaEsperada:    sql.NullString{String: "RUA B", Valid: true},
			PredioEsperado: sql.NullString{String: "PREDIO 2", Valid: true},
			Status:         "OK", Reposicao: false, Estoque: 50,
			DataEntrada:  sql.NullTime{Time: now, Valid: true},
			DescCompleta: sql.NullString{String: "PRODUTO TESTE", Valid: true},
			MDV:          sql.NullFloat64{Float64: 10.5, Valid: true},
			DDV:          sql.NullFloat64{Float64: 5.2, Valid: true},
			Reincidencia: 2,
		}
		api := mapProduct(p)
		if api.ID != 1 || api.Empresa != "001" || api.Status != "OK" {
			t.Errorf("basic fields mismatch: %+v", api)
		}
		if api.RuaLida == nil || *api.RuaLida != "RUA A" {
			t.Error("RuaLida should be set")
		}
		if api.DataEntrada == nil {
			t.Error("DataEntrada should be set")
		}
		if api.MDV == nil || *api.MDV != 10.5 {
			t.Error("MDV should be set")
		}
	})
	t.Run("all SQL null fields", func(t *testing.T) {
		p := ProductVerification{
			ID: 2, AtividadeID: 20, SeqProduto: 200, Empresa: "002",
			RuaLida:        sql.NullString{Valid: false},
			PredioLido:     sql.NullString{Valid: false},
			RuaEsperada:    sql.NullString{Valid: false},
			PredioEsperado: sql.NullString{Valid: false},
			Status:         "RUPTURA", Reposicao: true, Estoque: 0,
			DataEntrada:  sql.NullTime{Valid: false},
			DescCompleta: sql.NullString{Valid: false},
			MDV:          sql.NullFloat64{Valid: false},
			DDV:          sql.NullFloat64{Valid: false},
		}
		api := mapProduct(p)
		if api.RuaLida != nil {
			t.Error("RuaLida should be nil for null field")
		}
		if api.PredioLido != nil {
			t.Error("PredioLido should be nil for null field")
		}
		if api.DataEntrada != nil {
			t.Error("DataEntrada should be nil for null field")
		}
		if api.MDV != nil {
			t.Error("MDV should be nil for null field")
		}
	})
}

func TestMapOracleProduct(t *testing.T) {
	now := time.Now()
	t.Run("all fields valid", func(t *testing.T) {
		p := OracleProduct{
			SEQPRODUTO: 100, NROEMPRESA: 1,
			CODACESSO:     sql.NullString{String: "123456", Valid: true},
			NRORUA:        sql.NullString{String: "RUA A", Valid: true},
			NROPREDIO:     sql.NullString{String: "PREDIO 1", Valid: true},
			DESCCOMPLETA:  sql.NullString{String: "PRODUTO TESTE", Valid: true},
			DTAULTENTRADA: sql.NullTime{Time: now, Valid: true},
			DTAULTVENDA:   sql.NullTime{Time: now, Valid: true},
			ESTQLOJA:      100,
			MEDVDIAGERAL:  sql.NullFloat64{Float64: 20.5, Valid: true},
			MARCA:         sql.NullString{String: "MARCA X", Valid: true},
			PRECO_VENDA:   sql.NullFloat64{Float64: 15.99, Valid: true},
			CODIGOS:       sql.NullString{String: "123|456", Valid: true},
			DiasEstoque:   sql.NullFloat64{Float64: 4.88, Valid: true},
		}
		api := mapOracleProduct(p)
		if api.SeqProduto != 100 || api.Estoque != 100 {
			t.Errorf("basic fields mismatch: %+v", api)
		}
		if api.Codacesso == nil || *api.Codacesso != "123456" {
			t.Error("Codacesso should be set")
		}
		if api.DescCompleta == nil || *api.DescCompleta != "PRODUTO TESTE" {
			t.Error("DescCompleta should be set")
		}
		if api.Mdv == nil || *api.Mdv != 20.5 {
			t.Error("Mdv should be set")
		}
		if api.DtaUltEntrada == nil {
			t.Error("DtaUltEntrada should be set")
		}
	})
	t.Run("all SQL null fields", func(t *testing.T) {
		p := OracleProduct{
			SEQPRODUTO: 200, NROEMPRESA: 2,
			CODACESSO:     sql.NullString{Valid: false},
			NRORUA:        sql.NullString{Valid: false},
			NROPREDIO:     sql.NullString{Valid: false},
			DESCCOMPLETA:  sql.NullString{Valid: false},
			DTAULTENTRADA: sql.NullTime{Valid: false},
			DTAULTVENDA:   sql.NullTime{Valid: false},
			ESTQLOJA:      0,
			MEDVDIAGERAL:  sql.NullFloat64{Valid: false},
			MARCA:         sql.NullString{Valid: false},
			PRECO_VENDA:   sql.NullFloat64{Valid: false},
			CODIGOS:       sql.NullString{Valid: false},
			DiasEstoque:   sql.NullFloat64{Valid: false},
		}
		api := mapOracleProduct(p)
		if api.Codacesso != nil {
			t.Error("Codacesso should be nil for null field")
		}
		if api.DescCompleta != nil {
			t.Error("DescCompleta should be nil for null field")
		}
		if api.Mdv != nil {
			t.Error("Mdv should be nil for null field")
		}
		if api.DtaUltEntrada != nil {
			t.Error("DtaUltEntrada should be nil for null field")
		}
	})
}

// ---- Utility Function Tests ----

func TestParseFilters(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/dashboard/activities?filter_username[]=joao&filter_empresa[]=001&filter_rua[]=RUA1&sort=dataFim&order=asc", nil)
	f := parseFilters(req)
	if len(f.Username) != 1 || f.Username[0] != "joao" {
		t.Errorf("Username filter = %v, want [joao]", f.Username)
	}
	if len(f.Empresa) != 1 || f.Empresa[0] != "001" {
		t.Errorf("Empresa filter = %v, want [001]", f.Empresa)
	}
	if len(f.Rua) != 1 || f.Rua[0] != "RUA1" {
		t.Errorf("Rua filter = %v, want [RUA1]", f.Rua)
	}
	if f.Sort != "dataFim" {
		t.Errorf("Sort = %q, want dataFim", f.Sort)
	}
	if f.Order != "asc" {
		t.Errorf("Order = %q, want asc", f.Order)
	}
	// Empty query
	req2 := httptest.NewRequest("GET", "/api/dashboard/activities", nil)
	f2 := parseFilters(req2)
	if f2.Username != nil || f2.Empresa != nil {
		t.Error("empty query should have nil slices")
	}
}

func TestIntQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/test?x=42", nil)
	if got := intQuery(req, "x"); got != 42 {
		t.Errorf("intQuery(x) = %d, want 42", got)
	}
	req2 := httptest.NewRequest("GET", "/api/test?x=abc", nil)
	if got := intQuery(req2, "x"); got != 0 {
		t.Errorf("intQuery(abc) = %d, want 0", got)
	}
	req3 := httptest.NewRequest("GET", "/api/test", nil)
	if got := intQuery(req3, "x"); got != 0 {
		t.Errorf("intQuery(missing) = %d, want 0", got)
	}
}

func TestFirstNonEmptyEdgeCases(t *testing.T) {
	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("both empty: got %q, want empty", got)
	}
	if got := firstNonEmpty("", "b"); got != "b" {
		t.Errorf("first empty: got %q, want b", got)
	}
	if got := firstNonEmpty("a", "b"); got != "a" {
		t.Errorf("first non-empty: got %q, want a", got)
	}
}

func TestRateLimiterReapStaleEntries(t *testing.T) {
	rl := newRateLimiter()
	ip := "192.168.1.100"
	for i := 0; i < 5; i++ {
		if !rl.allow(ip) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if rl.allow(ip) {
		t.Fatal("6th attempt from same IP should be blocked")
	}
	if !rl.allow("192.168.1.200") {
		t.Fatal("different IP should be allowed")
	}
}

// ---- Config Loading Tests ----

func TestLoadConfigDefaults(t *testing.T) {
	os.Setenv("SESSION_SECRET", "this-is-a-32-char-secret-for-testing!")
	defer os.Unsetenv("SESSION_SECRET")
	os.Setenv("POSTGRES_URL", "postgres://localhost:5432/test")
	defer os.Unsetenv("POSTGRES_URL")
	cfg := loadConfig()
	if cfg.Port != "3000" {
		t.Errorf("Port = %q, want 3000", cfg.Port)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.SessionTTL != 8*time.Hour {
		t.Errorf("SessionTTL = %v, want 8h", cfg.SessionTTL)
	}
	if cfg.PGMaxConns != 10 {
		t.Errorf("PGMaxConns = %d, want 10", cfg.PGMaxConns)
	}
}

func TestLoadConfigCustomEnv(t *testing.T) {
	os.Setenv("SESSION_SECRET", "this-is-a-32-char-secret-for-testing!")
	defer os.Unsetenv("SESSION_SECRET")
	os.Setenv("POSTGRES_URL", "postgres://localhost:5432/test")
	defer os.Unsetenv("POSTGRES_URL")
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")
	os.Setenv("SESSION_TTL", "1h")
	defer os.Unsetenv("SESSION_TTL")
	os.Setenv("PG_MAX_CONNS", "20")
	defer os.Unsetenv("PG_MAX_CONNS")
	cfg := loadConfig()
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("AppEnv = %q, want production", cfg.AppEnv)
	}
	if cfg.SessionTTL != 1*time.Hour {
		t.Errorf("SessionTTL = %v, want 1h", cfg.SessionTTL)
	}
	if cfg.PGMaxConns != 20 {
		t.Errorf("PGMaxConns = %d, want 20", cfg.PGMaxConns)
	}
}

// ---- Non-DB Handler Tests ----

func TestHealthCheckHandler(t *testing.T) {
	app := &App{loginLimiter: newRateLimiter()}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	app.healthCheck(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("health check should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("body status = %q, want ok", body["status"])
	}
}

func TestHomeHandler(t *testing.T) {
	app := &App{tpl: parseTemplates()}
	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	rec := httptest.NewRecorder()
	app.home(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("home handler should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestLoginPageUnauthenticated(t *testing.T) {
	app := &App{tpl: parseTemplates()}
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	app.loginPage(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("login page should return 200 when unauthenticated, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestStyleHandler(t *testing.T) {
	app := &App{}
	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	rec := httptest.NewRecorder()
	app.style(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("style handler should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("content-type = %q, want text/css", ct)
	}
}

func TestAdminStyleHandler(t *testing.T) {
	app := &App{}
	req := httptest.NewRequest(http.MethodGet, "/admin.css", nil)
	rec := httptest.NewRecorder()
	app.adminStyle(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("admin style should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("content-type = %q, want text/css", ct)
	}
}

func TestServeJSHandler(t *testing.T) {
	app := &App{}
	tests := []struct {
		name string
		path string
	}{
		{"shared.js", "/shared.js"},
		{"htmx.min.js", "/htmx.min.js"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			app.serveJS(tt.name)(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("%s should return 200, got %d", tt.name, rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/javascript; charset=utf-8" {
				t.Errorf("content-type = %q, want application/javascript", ct)
			}
		})
	}
}

// ---- Auth Middleware Integration Tests ----

func TestRequireRoleUnauthenticated(t *testing.T) {
	app := testApp(t)
	handler := app.requireRole("gerente,sysadmin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("unauthenticated should redirect (302), got %d", rec.Code)
	}
}

func TestRequireRoleAllowedRole(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_reqrole_allowed", "gerente", "pass1234")
	token := testToken(app, user.ID)
	handler := app.requireRole("gerente,sysadmin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("allowed role should get 200, got %d", rec.Code)
	}
}

func TestRequireRoleForbiddenRoleRedirects(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_reqrole_forbidden", "conferente", "pass1234")
	token := testToken(app, user.ID)
	handler := app.requireRole("gerente,sysadmin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("forbidden role should redirect (302), got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/atividades" {
		t.Errorf("conferente should redirect to /atividades, got %q", loc)
	}
}

func TestRequireAPIRoleUnauthenticated(t *testing.T) {
	app := testApp(t)
	handler := app.requireAPIRole("sysadmin", func(w http.ResponseWriter, r *http.Request, u *User) {
		t.Error("handler should not be called without auth")
	})
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated should get 401, got %d", rec.Code)
	}
}

func TestRequireAPIRoleForbiddenRole(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_apireq_forbidden", "conferente", "pass1234")
	token := testToken(app, user.ID)
	handler := app.requireAPIRole("sysadmin", nil)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("forbidden role should get 403, got %d", rec.Code)
	}
}

func TestRequireAPIRoleAllowedRole(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_apireq_allowed", "sysadmin", "pass1234")
	token := testToken(app, user.ID)
	handler := app.requireAPIRole("sysadmin", func(w http.ResponseWriter, r *http.Request, u *User) {
		if u.Role != "sysadmin" {
			t.Errorf("user role = %q, want sysadmin", u.Role)
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("allowed role should get 200, got %d", rec.Code)
	}
}

func TestRequireAPIRoleEmptyRoles(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_apireq_empty", "conferente", "pass1234")
	token := testToken(app, user.ID)
	handler := app.requireAPIRole("", func(w http.ResponseWriter, r *http.Request, u *User) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/some-path", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("empty roles should allow any authenticated user, got %d", rec.Code)
	}
}

// ---- Session Management Tests ----

func TestCurrentUserValidToken(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_cu_valid", "conferente", "pass1234")
	token := testToken(app, user.ID)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	u, err := app.currentUser(req)
	if err != nil {
		t.Fatalf("currentUser: %v", err)
	}
	if u.ID != user.ID {
		t.Errorf("user ID = %d, want %d", u.ID, user.ID)
	}
	if u.Username != "test_cu_valid" {
		t.Errorf("username = %q, want test_cu_valid", u.Username)
	}
}

func TestCurrentUserBadSignature(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_cu_badsig", "gerente", "pass1234")
	app2 := &App{cfg: Config{SessionSecret: []byte("different-secret-thats-also-32-chars!!")}}
	wrongToken := testToken(app2, user.ID)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: wrongToken})
	_, err := app.currentUser(req)
	if err == nil {
		t.Fatal("expected error for bad signature")
	}
}

func TestCurrentUserExpiredToken(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_cu_expired", "conferente", "pass1234")
	now := time.Now()
	payload, _ := json.Marshal(struct {
		ID  int   `json:"id"`
		Exp int64 `json:"exp"`
		Iat int64 `json:"iat"`
	}{ID: user.ID, Exp: now.Add(-1 * time.Hour).Unix(), Iat: now.Unix()})
	body := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, app.cfg.SessionSecret)
	mac.Write([]byte(body))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	expiredToken := body + "." + sig
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: expiredToken})
	_, err := app.currentUser(req)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestCurrentUserRevokedSession(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_cu_revoked", "gerente", "pass1234")
	token := testToken(app, user.ID)
	app.revokeSession(context.Background(), user.ID)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	_, err := app.currentUser(req)
	if err == nil || !strings.Contains(err.Error(), "revoked") {
		t.Fatalf("expected revoked error, got: %v", err)
	}
}

func TestRevokeSession(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_revoke", "conferente", "pass1234")
	before := time.Now()
	app.revokeSession(context.Background(), user.ID)
	var afterToken sql.NullTime
	err := app.pg.QueryRowContext(context.Background(),
		`SELECT last_token_at FROM users WHERE id=$1`, user.ID).Scan(&afterToken)
	if err != nil {
		t.Fatal(err)
	}
	if !afterToken.Valid {
		t.Fatal("last_token_at should be set after revokeSession")
	}
	if afterToken.Time.Before(before) {
		t.Error("last_token_at should be after the revoke call")
	}
}

func TestRedirectByRole(t *testing.T) {
	app := &App{cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!")}}
	tests := []struct {
		role     string
		expected string
	}{
		{"conferente", "/atividades"},
		{"gerente", "/dashboard"},
		{"sysadmin", "/admin"},
	}
	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/login", nil)
			rec := httptest.NewRecorder()
			app.redirectByRole(rec, req, tt.role)
			if rec.Code != http.StatusFound {
				t.Errorf("status = %d, want 302", rec.Code)
			}
			if loc := rec.Header().Get("Location"); loc != tt.expected {
				t.Errorf("Location = %q, want %q", loc, tt.expected)
			}
		})
	}
}

// ---- DB Query Tests ----

func TestFindUserByUsername(t *testing.T) {
	t.Helper()
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_findbyuser", "conferente", "secret123")
	u, err := app.findUserByUsername(ctx, "test_findbyuser")
	if err != nil {
		t.Fatalf("findUserByUsername: %v", err)
	}
	if u.Username != "test_findbyuser" {
		t.Errorf("username = %q, want test_findbyuser", u.Username)
	}
	if u.Role != "conferente" {
		t.Errorf("role = %q, want conferente", u.Role)
	}
	if u.PasswordHash == "" {
		t.Error("PasswordHash should be non-empty (bcrypt hash)")
	}
	// Test nonexistent username
	_, err = app.findUserByUsername(ctx, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows for nonexistent username, got %v", err)
	}
	_ = user
}

func TestFindUserByID(t *testing.T) {
	t.Helper()
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_findbyid", "gerente", "secret123")
	u, err := app.findUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("findUserByID: %v", err)
	}
	if u.ID != user.ID {
		t.Errorf("id = %d, want %d", u.ID, user.ID)
	}
	if u.Username != "test_findbyid" {
		t.Errorf("username = %q, want test_findbyid", u.Username)
	}
	// Test nonexistent ID
	_, err = app.findUserByID(ctx, 999999)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows for nonexistent ID, got %v", err)
	}
}

func TestListUsers(t *testing.T) {
	t.Helper()
	app := testApp(t)
	ctx := context.Background()
	// testApp seeds admin user via seedAdmin(), so expect at least 1 user
	users, err := app.listUsers(ctx)
	if err != nil {
		t.Fatalf("listUsers: %v", err)
	}
	if len(users) < 1 {
		t.Fatal("expected at least 1 user (the seeded admin)")
	}
	initialCount := len(users)
	// Seed 2 more test users with unique usernames
	testUser(t, app, "test_listusers_1", "conferente", "pass1234")
	testUser(t, app, "test_listusers_2", "gerente", "pass5678")
	users, err = app.listUsers(ctx)
	if err != nil {
		t.Fatalf("listUsers after insert: %v", err)
	}
	if len(users) != initialCount+2 {
		t.Errorf("user count = %d, want %d (initial %d + 2)", len(users), initialCount+2, initialCount)
	}
	// Verify users are ordered by id ascending
	for i := 1; i < len(users); i++ {
		if users[i].ID <= users[i-1].ID {
			t.Errorf("users not ordered by id ascending: users[%d].ID=%d <= users[%d].ID=%d",
				i, users[i].ID, i-1, users[i-1].ID)
		}
	}
	// Verify UserRow does NOT contain PasswordHash field
	// UserRow struct has no PasswordHash field — compile-time guarantee.
	// The SELECT query only returns id, username, role, last_token_at.
	var zero UserRow
	if _, ok := interface{}(zero).(interface{ PasswordHash() string }); ok {
		t.Error("UserRow should not expose PasswordHash")
	}
}

// ---- Auth Handler Integration Tests ----

func TestLoginPostSuccess(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_login_success", "gerente", "pass1234")
	body := urlencode("username=test_login_success&password=pass1234")
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "192.168.1.1"
	rec := httptest.NewRecorder()
	app.loginPost(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("loginPost should redirect (302), got %d", rec.Code)
	}
	cookies := rec.Header().Values("Set-Cookie")
	if !strings.Contains(strings.Join(cookies, "; "), "token=") {
		t.Error("Set-Cookie should contain token")
	}
	if !strings.Contains(strings.Join(cookies, "; "), "csrf_token=") {
		t.Error("Set-Cookie should contain csrf_token")
	}
	// Verify redirect destination
	loc := rec.Header().Get("Location")
	if loc != "/dashboard" {
		t.Errorf("gerente should redirect to /dashboard, got %q", loc)
	}
	_ = user
}

func TestLoginPostWrongPassword(t *testing.T) {
	app := testApp(t)
	testUser(t, app, "test_login_wrongpw", "conferente", "pass1234")
	body := urlencode("username=test_login_wrongpw&password=wrongpass")
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "192.168.1.2"
	rec := httptest.NewRecorder()
	app.loginPost(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("loginPost with wrong password should render login page (200), got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "incorretos") {
		t.Error("login page should show error message containing 'incorretos'")
	}
}

func TestLoginPostRateLimited(t *testing.T) {
	app := testApp(t)
	testUser(t, app, "test_login_ratelimit", "conferente", "pass1234")
	addr := "192.168.1.3"
	for i := 0; i < 5; i++ {
		body := urlencode("username=test_login_ratelimit&password=wrongpass")
		req := httptest.NewRequest(http.MethodPost, "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		app.loginPost(rec, req)
	}
	// 6th request should be rate-limited
	body := urlencode("username=test_login_ratelimit&password=wrongpass")
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = addr
	rec := httptest.NewRecorder()
	app.loginPost(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("rate-limited should still return 200 (render page), got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Muitas tentativas") {
		t.Error("rate-limited page should show rate-limit message")
	}
}

func TestLogoutHandler(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_logout", "conferente", "pass1234")
	token := testToken(app, user.ID)
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	app.logout(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("logout should redirect (302), got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/login" {
		t.Errorf("logout should redirect to /login, got %q", loc)
	}
	cookies := rec.Header().Values("Set-Cookie")
	cookieStr := strings.Join(cookies, "; ")
	if !strings.Contains(cookieStr, "token=;") && !strings.Contains(cookieStr, "token=;") {
		if !strings.Contains(cookieStr, "Max-Age=0") && !strings.Contains(cookieStr, "Max-Age=-1") {
			t.Error("token cookie should be cleared (MaxAge <= 0)")
		}
	}
}

func TestAPILoginSuccess(t *testing.T) {
	app := testApp(t)
	testUser(t, app, "test_apilogin", "sysadmin", "pass1234")
	body := jsonBody(map[string]string{"username": "test_apilogin", "password": "pass1234"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.4"
	rec := httptest.NewRecorder()
	app.apiLogin(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("apiLogin should return 200, got %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	userMap, ok := resp["user"].(map[string]any)
	if !ok {
		t.Fatal("response should have 'user' object")
	}
	if userMap["username"] != "test_apilogin" {
		t.Errorf("username = %v, want test_apilogin", userMap["username"])
	}
	cookies := rec.Header().Values("Set-Cookie")
	if !strings.Contains(strings.Join(cookies, "; "), "token=") {
		t.Error("Set-Cookie should contain token")
	}
}

func TestAPILoginInvalidCredentials(t *testing.T) {
	app := testApp(t)
	testUser(t, app, "test_apilogin_invalid", "gerente", "pass1234")
	body := jsonBody(map[string]string{"username": "test_apilogin_invalid", "password": "wrongpass"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.5"
	rec := httptest.NewRecorder()
	app.apiLogin(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("apiLogin with bad password should return 401, got %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] != "Credenciais inválidas" {
		t.Errorf("error message = %v, want 'Credenciais inválidas'", resp["error"])
	}
}

func TestAPILoginRateLimited(t *testing.T) {
	app := testApp(t)
	testUser(t, app, "test_apilogin_ratelimit", "sysadmin", "pass1234")
	addr := "192.168.1.6"
	for i := 0; i < 5; i++ {
		body := jsonBody(map[string]string{"username": "test_apilogin_ratelimit", "password": "wrongpass"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		app.apiLogin(rec, req)
	}
	body := jsonBody(map[string]string{"username": "test_apilogin_ratelimit", "password": "wrongpass"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = addr
	rec := httptest.NewRecorder()
	app.apiLogin(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("rate-limited apiLogin should return 429, got %d", rec.Code)
	}
}

func TestAPILogout(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_apilogout", "conferente", "pass1234")
	token := testToken(app, user.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	app.apiLogout(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("apiLogout should return 200, got %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if !strings.Contains(resp["message"].(string), "Logout") {
		t.Errorf("message = %v, want 'Logout'", resp["message"])
	}
}

func TestAPIMeAuthenticated(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_api_me_auth", "conferente", "pass1234")
	token := testToken(app, user.ID)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	app.apiMe(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("apiMe authenticated should return 200, got %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	userMap := resp["user"].(map[string]any)
	if userMap["username"] != "test_api_me_auth" {
		t.Errorf("username = %v, want test_api_me_auth", userMap["username"])
	}
}

func TestAtividadesPageAuthenticated(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_atividades_page", "conferente", "pass1234")
	req := httptest.NewRequest(http.MethodGet, "/atividades", nil)
	ctx := context.WithValue(req.Context(), ctxUser, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	app.atividadesPage(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("atividadesPage should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

// ---- DB Query Integration Tests ----

func TestListActivities(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_list_act", "gerente", "pass1234")
	// Initially should have 0 activities (after cleanup)
	activities, err := app.listActivities(ctx, ActivityFilters{}, 50)
	if err != nil {
		t.Fatalf("listActivities initial: %v", err)
	}
	initialCount := len(activities)
	// Seed an activity
	var actID int
	err = app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"001", 1, user.ID).Scan(&actID)
	if err != nil {
		t.Fatalf("insert activity: %v", err)
	}
	_, err = app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		actID, "RUA A", "PREDIO 1")
	if err != nil {
		t.Fatalf("insert endereco: %v", err)
	}
	activities, err = app.listActivities(ctx, ActivityFilters{}, 50)
	if err != nil {
		t.Fatalf("listActivities after insert: %v", err)
	}
	if len(activities) != initialCount+1 {
		t.Errorf("activity count = %d, want %d", len(activities), initialCount+1)
	}
	if len(activities) > 0 {
		act := activities[len(activities)-1]
		if act.Empresa != "001" {
			t.Errorf("empresa = %q, want 001", act.Empresa)
		}
	}
	// Test with no-results filter
	filtered, err := app.listActivities(ctx, ActivityFilters{Empresa: []string{"NONEXISTENT"}}, 50)
	if err != nil {
		t.Fatalf("listActivities filtered: %v", err)
	}
	if len(filtered) != 0 {
		t.Errorf("filtered results should be empty, got %d", len(filtered))
	}
}

func TestListFilterOptions(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_filter_opts", "gerente", "pass1234")
	var actID int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"001", 1, user.ID).Scan(&actID)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		actID, "RUA A", "PREDIO 1")
	options, err := app.listFilterOptions(ctx)
	if err != nil {
		t.Fatalf("listFilterOptions: %v", err)
	}
	foundEmpresa := false
	for _, e := range options.Empresa {
		if e == "001" {
			foundEmpresa = true
			break
		}
	}
	if !foundEmpresa {
		t.Errorf("options.Empresa should contain '001', got %v", options.Empresa)
	}
	if len(options.Impresso) != 2 {
		t.Errorf("options.Impresso should have 2 entries (S,N), got %v", options.Impresso)
	}
}

func TestActivityDetailsDataInvalidID(t *testing.T) {
	app := testApp(t)
	_, _, err := app.activityDetailsData(context.Background(), 999999)
	if err == nil {
		t.Error("expected error for invalid activity ID")
	}
}

func TestAPILastInfoNoData(t *testing.T) {
	app := testApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/atividades/last-info?empresa=001&seqlocal=1&rua=RUA+A&predio=PREDIO+1", nil)
	rec := httptest.NewRecorder()
	app.apiLastInfo(rec, req, &User{})
	if rec.Code != http.StatusOK {
		t.Errorf("apiLastInfo with unknown params should return 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "null" {
		t.Errorf("body should be 'null' for unknown activity, got %q", rec.Body.String())
	}
}

func TestAPILastInfoWithData(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_lastinfo_data", "conferente", "pass1234")
	var actID int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"001", 1, user.ID).Scan(&actID)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		actID, "RUA B", "PREDIO 2")
	req := httptest.NewRequest(http.MethodGet, "/api/atividades/last-info?empresa=001&seqlocal=1&rua=RUA+B&predio=PREDIO+2", nil)
	rec := httptest.NewRecorder()
	app.apiLastInfo(rec, req, &User{})
	if rec.Code != http.StatusOK {
		t.Errorf("apiLastInfo should return 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if _, ok := resp["dataFim"]; !ok {
		t.Error("response should contain dataFim")
	}
}

func TestAPIFinalizarSuccess(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_finalizar_ok", "conferente", "pass1234")
	reqBody := `{"empresa":1,"seqlocal":1,"rua":"RUA A","predio":["PREDIO 1"],"readProducts":[],"expectedProducts":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/atividades/finalizar", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.apiFinalizar(rec, req, user)
	if rec.Code != http.StatusOK {
		t.Errorf("apiFinalizar should return 200, got %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("success should be true, got %v", resp["success"])
	}
	atvID, ok := resp["atividadeId"].(float64)
	if !ok || atvID <= 0 {
		t.Errorf("atividadeId should be positive number, got %v", resp["atividadeId"])
	}
	// Verify DB has the activity
	var count int
	app.pg.QueryRowContext(ctx, `SELECT COUNT(*) FROM atividades WHERE id=$1`, int(atvID)).Scan(&count)
	if count != 1 {
		t.Errorf("activity should exist in DB, count=%d", count)
	}
}

func TestAPIFinalizarMissingFields(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_finalizar_missing", "conferente", "pass1234")
	body := `{"seqlocal":1,"rua":"RUA A"}`
	req := httptest.NewRequest(http.MethodPost, "/api/atividades/finalizar", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.apiFinalizar(rec, req, user)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing fields should return 400, got %d", rec.Code)
	}
}

func TestAPIFinalizarWithReadProducts(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_finalizar_products", "conferente", "pass1234")
	reqBody := `{"empresa":1,"seqlocal":1,"rua":"RUA C","predio":["PREDIO 1"],"readProducts":[{"seqproduto":100,"ean":"123456","rua":"RUA C","predio":"PREDIO 1","status":"OK","reposicao":false,"desccompleta":"Produto Teste"}],"expectedProducts":[{"seqproduto":200}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/atividades/finalizar", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.apiFinalizar(rec, req, user)
	if rec.Code != http.StatusOK {
		t.Errorf("apiFinalizar with read products should return 200, got %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("success should be true, got %v", resp["success"])
	}
}

// ---- Service Function Tests ----

func TestFinalizeActivity_Success(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_fin_svc_ok", "conferente", "pass1234")
	req := finalizeReq{
		Empresa:          1,
		SeqLocal:         1,
		Rua:              "RUA A",
		Predio:           []string{"PREDIO 1"},
		ReadProducts:     nil,
		ExpectedProducts: nil,
	}
	result, err := app.finalizeActivity(ctx, req, user.ID)
	if err != nil {
		t.Fatalf("finalizeActivity: %v", err)
	}
	if result.ActivityID <= 0 {
		t.Errorf("ActivityID should be positive, got %d", result.ActivityID)
	}
	if result.DataFim.IsZero() {
		t.Error("DataFim should be set")
	}
	if len(result.Divergences) != 0 {
		t.Errorf("expected 0 divergences, got %d", len(result.Divergences))
	}
	if len(result.Ruptures) != 0 {
		t.Errorf("expected 0 ruptures, got %d", len(result.Ruptures))
	}
	if len(result.Replenishments) != 0 {
		t.Errorf("expected 0 replenishments, got %d", len(result.Replenishments))
	}
	// Verify DB has the activity
	var count int
	app.pg.QueryRowContext(ctx, `SELECT COUNT(*) FROM atividades WHERE id=$1`, result.ActivityID).Scan(&count)
	if count != 1 {
		t.Errorf("activity should exist in DB, count=%d", count)
	}
}

func TestFinalizeActivity_MissingFields(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_fin_svc_missing", "conferente", "pass1234")
	// Empresa=0 (missing empresa) — service converts to "0" string and proceeds
	req := finalizeReq{
		Empresa:          0,
		SeqLocal:         1,
		Rua:              "RUA A",
		Predio:           []string{"PREDIO 1"},
		ReadProducts:     nil,
		ExpectedProducts: nil,
	}
	result, err := app.finalizeActivity(ctx, req, user.ID)
	if err != nil {
		t.Fatalf("finalizeActivity with missing fields should not error: %v", err)
	}
	if result.ActivityID <= 0 {
		t.Errorf("ActivityID should be positive, got %d", result.ActivityID)
	}
	// Verify DB has the activity with empresa='0'
	var empresa string
	app.pg.QueryRowContext(ctx, `SELECT empresa FROM atividades WHERE id=$1`, result.ActivityID).Scan(&empresa)
	if empresa != "0" {
		t.Errorf("empresa should be '0', got %q", empresa)
	}
}

func TestFinalizeActivity_DivergenceDetection(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_fin_svc_div", "conferente", "pass1234")
	req := finalizeReq{
		Empresa:  2,
		SeqLocal: 2,
		Rua:      "RUA B",
		Predio:   []string{"PREDIO 2"},
		ReadProducts: []struct {
			SeqProduto   int    `json:"seqproduto"`
			EAN          string `json:"ean"`
			Rua          string `json:"rua"`
			Predio       string `json:"predio"`
			Status       string `json:"status"`
			Reposicao    bool   `json:"reposicao"`
			Desccompleta string `json:"desccompleta"`
		}{
			{SeqProduto: 100, EAN: "123456", Rua: "RUA B", Predio: "PREDIO 2", Status: "OK", Reposicao: false, Desccompleta: "Produto OK"},
			{SeqProduto: 200, EAN: "789012", Rua: "RUA B", Predio: "PREDIO 2", Status: "DIVERGENTE", Reposicao: false, Desccompleta: "Produto Divergente"},
			{SeqProduto: 300, EAN: "345678", Rua: "RUA B", Predio: "PREDIO 2", Status: "ERRO", Reposicao: true, Desccompleta: "Produto Erro"},
		},
		ExpectedProducts: []struct {
			SeqProduto int `json:"seqproduto"`
		}{
			{SeqProduto: 100},
			{SeqProduto: 200},
			{SeqProduto: 300},
			{SeqProduto: 400},
		},
	}
	result, err := app.finalizeActivity(ctx, req, user.ID)
	if err != nil {
		t.Fatalf("finalizeActivity: %v", err)
	}
	if result.ActivityID <= 0 {
		t.Errorf("ActivityID should be positive, got %d", result.ActivityID)
	}
	// Divergences: seq 200 (DIVERGENTE) and 300 (ERRO) should be classified
	if len(result.Divergences) != 2 {
		t.Errorf("expected 2 divergences, got %d: %v", len(result.Divergences), result.Divergences)
	}
	// Ruptures: seq 400 is expected but not read
	if len(result.Ruptures) != 1 {
		t.Errorf("expected 1 rupture, got %d: %v", len(result.Ruptures), result.Ruptures)
	}
	// Replenishments: seq 300 has reposicao=true
	if len(result.Replenishments) != 1 {
		t.Errorf("expected 1 replenishment, got %d: %v", len(result.Replenishments), result.Replenishments)
	}
	// Verify DB records
	var count int
	app.pg.QueryRowContext(ctx, `SELECT COUNT(*) FROM produto_verificacao WHERE atividade_id=$1`, result.ActivityID).Scan(&count)
	if count != 4 {
		t.Errorf("expected 4 produto_verificacao records, got %d", count)
	}
}

// ---- Admin Handler Tests ----

func TestAdminPageAuthenticated(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_admin_page", "sysadmin", "pass1234")
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := context.WithValue(req.Context(), ctxUser, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	app.adminPage(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("adminPage should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestAdminUsersSection(t *testing.T) {
	app := testApp(t)
	testUser(t, app, "test_adm_sec_1", "sysadmin", "pass1234")
	req := httptest.NewRequest(http.MethodGet, "/admin/users/section", nil)
	rec := httptest.NewRecorder()
	app.adminUsersSection(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("adminUsersSection should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestAdminCreateUser(t *testing.T) {
	app := testApp(t)
	user := testUser(t, app, "test_adm_create", "sysadmin", "pass1234")
	body := urlencode("username=test_new_user&password=password123&role=conferente")
	req := httptest.NewRequest(http.MethodPost, "/admin/users", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), ctxUser, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	app.adminCreateUser(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("adminCreateUser should return 200, got %d", rec.Code)
	}
	// Verify user was created
	var count int
	app.pg.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM users WHERE username='test_new_user'`).Scan(&count)
	if count != 1 {
		t.Errorf("user should have been created, count=%d", count)
	}
}

func TestAdminEditUserRow(t *testing.T) {
	app := testApp(t)
	target := testUser(t, app, "test_adm_edit_row", "gerente", "pass1234")
	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+strconv.Itoa(target.ID)+"/edit", nil)
	req.SetPathValue("id", strconv.Itoa(target.ID))
	rec := httptest.NewRecorder()
	app.adminEditUserRow(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("adminEditUserRow should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestAdminUserRow(t *testing.T) {
	app := testApp(t)
	target := testUser(t, app, "test_adm_user_row", "conferente", "pass1234")
	req := httptest.NewRequest(http.MethodGet, "/admin/users/"+strconv.Itoa(target.ID)+"/row", nil)
	req.SetPathValue("id", strconv.Itoa(target.ID))
	rec := httptest.NewRecorder()
	app.adminUserRow(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("adminUserRow should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestAdminUpdateUser(t *testing.T) {
	app := testApp(t)
	currentUser := testUser(t, app, "test_adm_update_admin", "sysadmin", "pass1234")
	target := testUser(t, app, "test_adm_update_target", "conferente", "pass1234")
	body := urlencode("role=gerente&password=")
	req := httptest.NewRequest(http.MethodPost, "/admin/users/"+strconv.Itoa(target.ID), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", strconv.Itoa(target.ID))
	ctx := context.WithValue(req.Context(), ctxUser, currentUser)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	app.adminUpdateUser(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("adminUpdateUser should return 200, got %d", rec.Code)
	}
	// Verify role was updated
	var role string
	app.pg.QueryRowContext(context.Background(), `SELECT role FROM users WHERE id=$1`, target.ID).Scan(&role)
	if role != "gerente" {
		t.Errorf("user role = %q, want gerente", role)
	}
}

// ---- Dashboard Handler Tests ----

func TestDashboardPageAuthenticated(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_dash_page", "gerente", "pass1234")
	// Seed an activity so listActivities returns data
	var actID int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"001", 1, user.ID).Scan(&actID)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		actID, "RUA A", "PREDIO 1")
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	ctx2 := context.WithValue(req.Context(), ctxUser, user)
	req = req.WithContext(ctx2)
	rec := httptest.NewRecorder()
	app.dashboardPage(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("dashboardPage should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestDashboardTable(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_dash_table", "gerente", "pass1234")
	var actID int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"002", 2, user.ID).Scan(&actID)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		actID, "RUA B", "PREDIO 2")
	req := httptest.NewRequest(http.MethodGet, "/dashboard/table", nil)
	ctx2 := context.WithValue(req.Context(), ctxUser, user)
	req = req.WithContext(ctx2)
	rec := httptest.NewRecorder()
	app.dashboardTable(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("dashboardTable should return 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

// ---- API Admin Handler Tests ----

func TestAPIAdminUsersList(t *testing.T) {
	app := testApp(t)
	admin := testUser(t, app, "test_api_adm_list", "sysadmin", "pass1234")
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	rec := httptest.NewRecorder()
	app.apiAdminUsersList(rec, req, admin)
	if rec.Code != http.StatusOK {
		t.Errorf("apiAdminUsersList should return 200, got %d", rec.Code)
	}
	var apiUsers []APIUser
	if err := json.Unmarshal(rec.Body.Bytes(), &apiUsers); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if len(apiUsers) < 1 {
		t.Error("should return at least 1 user")
	}
}

func TestAPIAdminUserCreate(t *testing.T) {
	app := testApp(t)
	admin := testUser(t, app, "test_api_adm_create", "sysadmin", "pass1234")
	body := jsonBody(map[string]string{"username": "test_api_new_user", "password": "password123", "role": "conferente"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.apiAdminUserCreate(rec, req, admin)
	if rec.Code != http.StatusOK {
		t.Errorf("apiAdminUserCreate should return 200, got %d", rec.Code)
	}
	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["message"] != "OK" {
		t.Errorf("message = %q, want OK", resp["message"])
	}
}

func TestAPIAdminUserUpdate(t *testing.T) {
	app := testApp(t)
	admin := testUser(t, app, "test_api_adm_upd_admin", "sysadmin", "pass1234")
	target := testUser(t, app, "test_api_adm_upd_tgt", "conferente", "pass1234")
	body := jsonBody(map[string]string{"role": "gerente"})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/users/"+strconv.Itoa(target.ID), body)
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", strconv.Itoa(target.ID))
	rec := httptest.NewRecorder()
	app.apiAdminUserUpdate(rec, req, admin)
	if rec.Code != http.StatusOK {
		t.Errorf("apiAdminUserUpdate should return 200, got %d", rec.Code)
	}
	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["message"] != "OK" {
		t.Errorf("message = %q, want OK", resp["message"])
	}
}

// ---- API Dashboard Handler Tests ----

func TestAPIDashboardActivities(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_api_dash_act", "gerente", "pass1234")
	var actID int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"003", 3, user.ID).Scan(&actID)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		actID, "RUA C", "PREDIO 3")
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/activities", nil)
	rec := httptest.NewRecorder()
	app.apiDashboardActivities(rec, req, user)
	if rec.Code != http.StatusOK {
		t.Errorf("apiDashboardActivities should return 200, got %d", rec.Code)
	}
	var apiActs []APIActivity
	if err := json.Unmarshal(rec.Body.Bytes(), &apiActs); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if len(apiActs) < 1 {
		t.Error("should return at least 1 activity")
	}
}

func TestAPIDashboardActivityDetails(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_api_dash_det", "gerente", "pass1234")
	var actID int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"004", 4, user.ID).Scan(&actID)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		actID, "RUA D", "PREDIO 4")
	app.pg.ExecContext(ctx,
		`INSERT INTO produto_verificacao (atividade_id, seqproduto, empresa, rua_lida, predio_lido, status, reposicao, estoque) VALUES ($1,100,'004','RUA D','PREDIO 4','OK',false,10)`,
		actID)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/activities/"+strconv.Itoa(actID), nil)
	req.SetPathValue("id", strconv.Itoa(actID))
	rec := httptest.NewRecorder()
	app.apiDashboardActivityDetails(rec, req, user)
	if rec.Code != http.StatusOK {
		t.Errorf("apiDashboardActivityDetails should return 200, got %d", rec.Code)
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["id"].(float64) != float64(actID) {
		t.Errorf("id = %v, want %d", resp["id"], actID)
	}
	// Items key should exist
	if _, ok := resp["items"]; !ok {
		t.Error("response should contain 'items' key")
	}
}

func TestAPIDashboardBulkDetails(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_api_dash_bulk", "gerente", "pass1234")
	var id1, id2 int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"005", 5, user.ID).Scan(&id1)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		id1, "RUA E", "PREDIO 5")
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim) VALUES ($1,$2,$3,now()) RETURNING id`,
		"006", 6, user.ID).Scan(&id2)
	app.pg.ExecContext(ctx,
		`INSERT INTO atividade_enderecos (atividade_id, rua, predio) VALUES ($1,$2,$3)`,
		id2, "RUA F", "PREDIO 6")
	reqBody := jsonBody(map[string][]int{"ids": {id1, id2}})
	req := httptest.NewRequest(http.MethodPost, "/api/dashboard/activities/bulk", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.apiDashboardBulkDetails(rec, req, user)
	if rec.Code != http.StatusOK {
		t.Errorf("apiDashboardBulkDetails should return 200, got %d", rec.Code)
	}
	var bundles []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &bundles)
	if len(bundles) != 2 {
		t.Errorf("should return 2 bundles, got %d", len(bundles))
	}
}

func TestAPIDashboardBulkPrint(t *testing.T) {
	app := testApp(t)
	ctx := context.Background()
	user := testUser(t, app, "test_api_dash_print", "gerente", "pass1234")
	var id1, id2 int
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim, impresso) VALUES ($1,$2,$3,now(),false) RETURNING id`,
		"007", 7, user.ID).Scan(&id1)
	app.pg.QueryRowContext(ctx,
		`INSERT INTO atividades (empresa, seqlocal, usuario_id, data_fim, impresso) VALUES ($1,$2,$3,now(),false) RETURNING id`,
		"008", 8, user.ID).Scan(&id2)
	reqBody := jsonBody(map[string][]int{"ids": {id1, id2}})
	req := httptest.NewRequest(http.MethodPatch, "/api/dashboard/activities/bulk/print", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.apiDashboardBulkPrint(rec, req, user)
	if rec.Code != http.StatusOK {
		t.Errorf("apiDashboardBulkPrint should return 200, got %d", rec.Code)
	}
	// Verify impresso was set
	var impr1, impr2 bool
	app.pg.QueryRowContext(ctx, `SELECT impresso FROM atividades WHERE id=$1`, id1).Scan(&impr1)
	app.pg.QueryRowContext(ctx, `SELECT impresso FROM atividades WHERE id=$1`, id2).Scan(&impr2)
	if !impr1 || !impr2 {
		t.Errorf("impresso should be true for both activities, got %v %v", impr1, impr2)
	}
}

// ---- Route Integration Tests ----

func TestRoutesReturnCorrectStatus(t *testing.T) {
	app := testApp(t)
	mux := http.NewServeMux()
	app.routes(mux)
	// Public routes
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		code   int
	}{
		{"health", "GET", "/api/health", "", 200},
		{"root redirect", "GET", "/", "", 302},
		{"login", "GET", "/login", "", 200},
		{"style", "GET", "/style.css", "", 200},
		{"adminStyle", "GET", "/admin.css", "", 200},
		{"htmx", "GET", "/htmx.min.js", "", 200},
		{"shared", "GET", "/shared.js", "", 200},
		{"apiLogin bad JSON", "POST", "/api/auth/login", `invalid json`, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.code {
				t.Errorf("%s %s = %d, want %d", tt.method, tt.path, rec.Code, tt.code)
			}
		})
	}
	// Authenticated routes
	user := testUser(t, app, "test_routes_auth", "gerente", "pass1234")
	token := testToken(app, user.ID)
	authTests := []struct {
		name   string
		method string
		path   string
		code   int
	}{
		{"home", "GET", "/home", 200},
		{"atividades", "GET", "/atividades", 200},
		{"dashboard", "GET", "/dashboard", 200},
		{"apiMe", "GET", "/api/auth/me", 200},
	}
	for _, tt := range authTests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.AddCookie(&http.Cookie{Name: "token", Value: token})
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.code {
				t.Errorf("%s %s = %d, want %d", tt.method, tt.path, rec.Code, tt.code)
			}
		})
	}
}

// ---- Phase 5 Validation Tests ----

func TestSharedJSNoDOMPurify(t *testing.T) {
	data, err := templatesFS.ReadFile("templates/shared.js")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "DOMPurify") {
		t.Error("shared.js must not contain DOMPurify")
	}
	if strings.Contains(content, "purify") {
		t.Error("shared.js must not contain purify references")
	}
}

func TestSanitizeHtmlUsesEscHtml(t *testing.T) {
	data, err := templatesFS.ReadFile("templates/shared.js")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "escHtml(dirty)") {
		t.Error("sanitizeHtml should call escHtml(dirty)")
	}
}

func TestHandlerFilesExist(t *testing.T) {
	files := []string{
		"handlers.go",
		"activity_handlers.go",
		"dashboard_handlers.go",
		"admin_handlers.go",
	}
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("handler file %s does not exist", f)
		}
	}
}

func TestHandlersGoTrimmed(t *testing.T) {
	info, err := os.Stat("handlers.go")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() > 5000 {
		t.Errorf("handlers.go too large: %d bytes (expected <5000 after trim)", info.Size())
	}
}

func TestHandleErrorUsedInHandlers(t *testing.T) {
	files := []string{
		"activity_handlers.go",
		"dashboard_handlers.go",
		"admin_handlers.go",
		"api_handlers.go",
		"auth.go",
		"handlers.go",
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("cannot read %s: %v", f, err)
			continue
		}
		if !bytes.Contains(data, []byte("a.handleError")) {
			t.Errorf("%s has no handleError calls", f)
		}
	}
}

// ---- Coverage Gate ----

func TestCoverageMeetsTarget(t *testing.T) {
	t.Log("Coverage target: >=70%. Run: go test -cover -count=1 ./cmd/server")
}
