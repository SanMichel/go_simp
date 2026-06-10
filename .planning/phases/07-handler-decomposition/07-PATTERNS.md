# Phase 7: Handler Decomposition — Pattern Map

**Mapped:** 2026-06-09
**Files analyzed:** 5 new/modified files
**Analogs found:** 5 / 5

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `cmd/server/activity_handlers.go` | handler + service | transaction (request-response) | `cmd/server/activity_handlers.go` (self — modify in place) | exact |
| `cmd/server/admin_handlers.go` | handler + service | CRUD (request-response) | `cmd/server/admin_handlers.go` (self — modify in place) | exact |
| `cmd/server/api_handlers.go` | handler + service | CRUD (request-response) | `cmd/server/api_handlers.go` (self — modify in place) | exact |
| `cmd/server/main_test.go` | test | transaction + CRUD | `cmd/server/main_test.go` (same file — add new tests) | exact |
| `cmd/server/models.go` | model | data types | `cmd/server/models.go` (same file — add domain types) | exact |

## Pattern Assignments

### `cmd/server/activity_handlers.go` — `apiFinalizar` → `finalizeActivity` (HAND-01)

**Analog (current handler — what to extract FROM):** `cmd/server/activity_handlers.go:147-247`

**Current handler imports** (activity_handlers.go:3-10):
```go
import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "time"
)
```

**Current handler signature + transaction pattern** (activity_handlers.go:147-247) — the existing 101-line handler that owns TX lifecycle:
```go
func (a *App) apiFinalizar(w http.ResponseWriter, r *http.Request, u *User) {
    var req finalizeReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "JSON inválido", HTTPStatus: http.StatusBadRequest})
        return
    }
    if req.Empresa == 0 || req.Rua == "" || req.SeqLocal == 0 {
        a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "Campos obrigatórios ausentes", HTTPStatus: http.StatusBadRequest})
        return
    }
    // ... 80+ lines of business logic (product comparison, classification, TX) ...
    tx, err := a.pg.BeginTx(r.Context(), nil)
    if err != nil { /*...*/ }
    defer tx.Rollback()
    // ... INSERT atividades, INSERT enderecos, INSERT produto_verificacao, COMMIT ...
}
```

**Target thin-adapter pattern** — RESEARCH.md lines 157-179:
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
        "success": true, "atividadeId": result.ActivityID,
        "dataFim": result.DataFim, "divergences": result.Divergences,
        "ruptures": result.Ruptures, "replenishments": result.Replenishments,
    })
}
```

**Best-existing thin-handler analog** — `apiProdutoEAN` (activity_handlers.go:51-63) — decode → service → handle error → writeJSON:
```go
func (a *App) apiProdutoEAN(w http.ResponseWriter, r *http.Request, u *User) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    codigo := r.PathValue("codigo")
    empresa, _ := strconv.Atoi(r.URL.Query().Get("empresa"))
    seqlocal, _ := strconv.Atoi(r.URL.Query().Get("seqlocal"))
    p, err := a.findAddressByCode(ctx, codigo, empresa, seqlocal)
    if err != nil {
        a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Produto não encontrado", HTTPStatus: http.StatusNotFound, Err: err})
        return
    }
    writeJSON(w, http.StatusOK, mapOracleProduct(p))
}
```

**Best-existing thin-JSON-handler analog** — `apiDashboardActivityDetails` (api_handlers.go:292-314):
```go
func (a *App) apiDashboardActivityDetails(w http.ResponseWriter, r *http.Request, u *User) {
    id, _ := strconv.Atoi(r.PathValue("id"))
    act, items, err := a.activityDetailsData(r.Context(), id)
    if err != nil {
        a.handleError(w, r, &AppError{Code: ErrCodeNotFound, Message: "Atividade não encontrada", HTTPStatus: http.StatusNotFound})
        return
    }
    apiAct := mapActivity(act)
    apiItems := make([]APIProductVerification, len(items))
    for i, item := range items {
        apiItems[i] = mapProduct(item)
    }
    writeJSON(w, http.StatusOK, map[string]any{
        "id": apiAct.ID, "empresa": apiAct.Empresa, "username": apiAct.Username,
        "dataFim": apiAct.DataFim, "rua": apiAct.Rua, "predio": apiAct.Predio,
        "impresso": apiAct.Impresso, "items": apiItems,
    })
}
```

**Existing service-function analog** — `activityDetailsData` (db.go:325-370) — *App method, returns domain types, uses context:
```go
func (a *App) activityDetailsData(ctx context.Context, id int) (Activity, []ProductVerification, error) {
    acts, err := a.listActivities(ctx, ActivityFilters{ID: []string{strconv.Itoa(id)}}, 1)
    if err != nil || len(acts) == 0 {
        return Activity{}, nil, sql.ErrNoRows
    }
    rows, err := a.pg.QueryContext(ctx, `SELECT ...`, id)
    if err != nil {
        return acts[0], nil, nil
    }
    defer rows.Close()
    var items []ProductVerification
    for rows.Next() {
        var p ProductVerification
        // ... scan and Oracle enrichment ...
        items = append(items, p)
    }
    return acts[0], items, nil
}
```

**Error handling pattern** — `handleError` (errors.go:51-107) — every handler calls this:
```go
func (a *App) handleError(w http.ResponseWriter, r *http.Request, err error) {
    var appErr *AppError
    if !errors.As(err, &appErr) {
        appErr = &AppError{
            Code: ErrCodeInternal, Message: "Erro interno do servidor",
            HTTPStatus: http.StatusInternalServerError, Err: err,
        }
    }
    // ... HTMX vs JSON vs plain text dispatching ...
}
```

---

### `cmd/server/admin_handlers.go` — `adminUpdateUser` → `updateUserAdmin` (HAND-02)

**Analog (current handler — what to extract FROM):** `admin_handlers.go:74-116`

**Current handler** (admin_handlers.go:74-116):
```go
func (a *App) adminUpdateUser(w http.ResponseWriter, r *http.Request) {
    currentUser := r.Context().Value(ctxUser).(*User)
    id, _ := strconv.Atoi(r.PathValue("id"))
    if id == currentUser.ID {
        a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "Não é possível editar o próprio usuário", HTTPStatus: http.StatusBadRequest})
        return
    }
    if err := r.ParseForm(); err != nil { /*...*/ }
    role := r.FormValue("role")
    password := r.FormValue("password")
    if !validRole(role) { /*...*/ }
    target, err := a.findUserByID(r.Context(), id)
    if err != nil { /*...*/ }
    if target.Role == "sysadmin" && currentUser.Role != "sysadmin" {
        a.handleError(w, r, &AppError{Code: ErrCodeForbidden, Message: "Sem permissão", HTTPStatus: http.StatusForbidden})
        return
    }
    if password != "" {
        hash, _ := bcrypt.GenerateFromPassword(...)
        a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1, password=$2, last_token_at=now() WHERE id=$3`, ...)
    } else {
        a.pg.ExecContext(r.Context(), `UPDATE users SET role=$1, last_token_at=now() WHERE id=$2`, ...)
    }
    u, _ := a.findUserByID(r.Context(), id)
    a.render(w, "user_row", map[string]any{"RowUser": u})
}
```

**Target thin-adapter pattern** — RESEARCH.md lines 200-216:
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

**Existing handler analog for form processing** — `adminCreateUser` (admin_handlers.go:27-53):
```go
func (a *App) adminCreateUser(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        a.handleError(w, r, &AppError{Code: ErrCodeBadRequest, Message: "Erro ao processar formulário", HTTPStatus: http.StatusBadRequest, Err: err})
        return
    }
    username := strings.TrimSpace(r.FormValue("username"))
    password := r.FormValue("password")
    role := r.FormValue("role")
    // ... validation ... bcrypt ... DB insert ... render ...
}
```

**Service function signature** (RESEARCH.md lines 192-198):
```go
func (a *App) updateUserAdmin(ctx context.Context, currentUserID, targetID int, role, password string) (*UserRow, error) {
    // Permission checks (self-edit guard, sysadmin protection)
    // bcrypt hash (if password provided)
    // DB update
    // Return updated UserRow + error
}
```

---

### `cmd/server/api_handlers.go` — `apiAdminUserUpdate` → reuses `updateUserAdmin` (HAND-02)

**Analog (current handler):** `api_handlers.go:214-255`

**Current handler** (api_handlers.go:214-255):
```go
func (a *App) apiAdminUserUpdate(w http.ResponseWriter, r *http.Request, u *User) {
    id, _ := strconv.Atoi(r.PathValue("id"))
    if id == u.ID { /* self-edit guard */ }
    target, err := a.findUserByID(r.Context(), id)
    if err != nil { /*...*/ }
    if target.Role == "sysadmin" && u.Role != "sysadmin" { /* sysadmin protection */ }
    var req struct { Role string; Password string }
    json.NewDecoder(r.Body).Decode(&req)
    if !validRole(req.Role) { /*...*/ }
    if req.Password != "" { /* bcrypt + update with password */ } else { /* update without password */ }
    writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}
```

**Target thin-adapter pattern** — RESEARCH.md lines 225-238:
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

**Existing JSON handler analog for thin-adapter** — `apiAdminUserCreate` (api_handlers.go:190-212):
```go
func (a *App) apiAdminUserCreate(w http.ResponseWriter, r *http.Request, u *User) {
    var req struct { Username string; Password string; Role string }
    json.NewDecoder(r.Body).Decode(&req)
    if req.Username == "" || len(req.Password) < 8 || !validRole(req.Role) {
        a.handleError(w, r, &AppError{Code: ErrCodeValidation, Message: "Dados inválidos", HTTPStatus: http.StatusBadRequest})
        return
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil { /*...*/ }
    _, err = a.pg.ExecContext(r.Context(), `INSERT INTO users ...`)
    if err != nil { /*...*/ }
    writeJSON(w, http.StatusOK, map[string]string{"message": "OK"})
}
```

---

### `cmd/server/api_handlers.go` — `apiDashboardBulkDetails` → `bulkActivityDetails` (HAND-02)

**Analog (current handler):** `api_handlers.go:316-351`

**Current handler** (api_handlers.go:316-351):
```go
func (a *App) apiDashboardBulkDetails(w http.ResponseWriter, r *http.Request, u *User) {
    var req struct { Ids []int `json:"ids"` }
    json.NewDecoder(r.Body).Decode(&req)
    type FlatBundle struct { /*...*/ }
    var bundles []FlatBundle
    for _, id := range req.Ids {
        act, items, err := a.activityDetailsData(r.Context(), id)
        if err == nil {
            apiAct := mapActivity(act)
            apiItems := make([]APIProductVerification, len(items))
            for i, item := range items { apiItems[i] = mapProduct(item) }
            bundles = append(bundles, FlatBundle{/*...*/})
        }
    }
    if bundles == nil { bundles = []FlatBundle{} }
    writeJSON(w, http.StatusOK, bundles)
}
```

**Target thin-adapter pattern** — RESEARCH.md lines 257-265:
```go
func (a *App) apiDashboardBulkDetails(w http.ResponseWriter, r *http.Request, u *User) {
    var req struct { Ids []int `json:"ids"` }
    json.NewDecoder(r.Body).Decode(&req)
    bundles := a.bulkActivityDetails(r.Context(), req.Ids)
    writeJSON(w, http.StatusOK, bundles)
}
```

**Service function signature** (RESEARCH.md lines 249-254):
```go
func (a *App) bulkActivityDetails(ctx context.Context, ids []int) []FlatBundle {
    // Loop, call activityDetailsData, map to FlatBundle
    // Return slice (nil → empty slice for JSON "[]")
}
```

**Current `FlatBundle` inline type** to extract to models.go (api_handlers.go:321-330):
```go
type FlatBundle struct {
    ID       int                      `json:"id"`
    Empresa  string                   `json:"empresa"`
    Username string                   `json:"username"`
    DataFim  *time.Time               `json:"dataFim"`
    Rua      string                   `json:"rua"`
    Predio   string                   `json:"predio"`
    Impresso bool                     `json:"impresso"`
    Items    []APIProductVerification `json:"items"`
}
```

---

### `cmd/server/main_test.go` — New service function tests

**Existing test pattern — handler test using `httptest`** `TestAPIFinalizarSuccess` (main_test.go:1819-1846):
```go
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
    // Verify DB state
    var count int
    app.pg.QueryRowContext(ctx, `SELECT COUNT(*) FROM atividades WHERE id=$1`, int(atvID)).Scan(&count)
    if count != 1 {
        t.Errorf("activity should exist in DB, count=%d", count)
    }
}
```

**Target service function test pattern** (RESEARCH.md lines 387-417):
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

**Existing test helpers** (testhelper.go:34-56):
```go
func testApp(t *testing.T) *App {
    t.Helper()
    pg := testDB(t)
    app := &App{
        cfg: Config{SessionSecret: []byte("test-secret-32-chars-minimum-length!"), SessionTTL: 8 * time.Hour, AppEnv: "test"},
        pg: pg, tpl: parseTemplates(), loginLimiter: newRateLimiter(),
    }
    // auto-migrate + seed admin
    t.Cleanup(func() { cleanupTestData(t, pg) })
    return app
}
```

**jsonBody helper** (main_test.go:27-30):
```go
func jsonBody(v any) io.Reader {
    b, _ := json.Marshal(v)
    return strings.NewReader(string(b))
}
```

**cleanupTestData** (testhelper.go:90-96):
```go
func cleanupTestData(t *testing.T, pg *sql.DB) {
    ctx := context.Background()
    _, _ = pg.ExecContext(ctx, `DELETE FROM produto_verificacao`)
    _, _ = pg.ExecContext(ctx, `DELETE FROM atividade_enderecos`)
    _, _ = pg.ExecContext(ctx, `DELETE FROM atividades`)
    _, _ = pg.ExecContext(ctx, `DELETE FROM users WHERE username LIKE 'test_%'`)
}
```

---

### `cmd/server/models.go` — Add `FinalizarResult` domain type

**Target type** — RESEARCH.md lines 141-148:
```go
type FinalizarResult struct {
    ActivityID     int
    DataFim        time.Time
    Divergences    []map[string]any
    Ruptures       []map[string]any
    Replenishments []map[string]any
}
```

**Existing domain type pattern** (models.go:49-60):
```go
type Activity struct {
    ID       int
    Empresa  string
    SeqLocal int
    UserID   int
    Username string
    DataFim  time.Time
    Impresso bool
    Rua      string
    Predio   string
    Predios  []string
}
```

**Imports pattern** (models.go:3-7):
```go
import (
    "database/sql"
    "html/template"
    "time"
)
```

## Shared Patterns

### Error Handling
**Source:** `cmd/server/errors.go:51-107` (`handleError`)
**Apply to:** All handler files (unchanged — already used)
```go
func (a *App) handleError(w http.ResponseWriter, r *http.Request, err error) {
    var appErr *AppError
    if !errors.As(err, &appErr) {
        appErr = &AppError{Code: ErrCodeInternal, Message: "Erro interno do servidor", HTTPStatus: http.StatusInternalServerError, Err: err}
    }
    if appErr.HTTPStatus == 0 {
        if s, ok := codeStatus[appErr.Code]; ok { appErr.HTTPStatus = s }
    }
    // ... HTMX vs JSON vs plain text dispatch ...
}
```

### JSON Response Encoding
**Source:** `cmd/server/utils.go:205-216` (`writeJSON`)
**Apply to:** All API handlers
```go
func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    buf := new(bytes.Buffer)
    if err := json.NewEncoder(buf).Encode(data); err != nil {
        slog.Error("writeJSON encode failed", "error", err, "status", status)
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"error":"internal error"}`))
        return
    }
    w.WriteHeader(status)
    _, _ = w.Write(buf.Bytes())
}
```

### Error Code Constants
**Source:** `cmd/server/errors.go:14-23`
**Apply to:** All service functions that return errors
```go
const (
    ErrCodeValidation   = "VALIDATION_ERROR"
    ErrCodeUnauthorized = "UNAUTHORIZED"
    ErrCodeForbidden    = "FORBIDDEN"
    ErrCodeNotFound     = "NOT_FOUND"
    ErrCodeConflict     = "CONFLICT"
    ErrCodeRateLimited  = "RATE_LIMITED"
    ErrCodeInternal     = "INTERNAL_ERROR"
    ErrCodeBadRequest   = "BAD_REQUEST"
)
```

### AppError Type
**Source:** `cmd/server/errors.go:25-35`
**Apply to:** Service functions return this type directly
```go
type AppError struct {
    Code       string // machine-readable code
    Message    string // user-facing message in Portuguese
    HTTPStatus int    // HTTP status code
    Err        error  // wrapped original error (optional)
}
func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }
```

### Auth Middleware (already in place — handlers receive `*User`)
**Source:** `cmd/server/auth.go:112-131` (`requireAPIRole`)
**Apply to:** All API handlers — `*User` is injected by middleware, service functions never see it
```go
func (a *App) requireAPIRole(roles string, next func(http.ResponseWriter, *http.Request, *User)) http.HandlerFunc {
    allowed := map[string]bool{}
    for _, r := range strings.Split(roles, ",") { if r != "" { allowed[r] = true } }
    return func(w http.ResponseWriter, r *http.Request) {
        u, err := a.currentUser(r)
        if err != nil {
            a.handleError(w, r, &AppError{Code: ErrCodeUnauthorized, Message: "Não autorizado", HTTPStatus: http.StatusUnauthorized})
            return
        }
        if len(allowed) > 0 && !allowed[u.Role] {
            a.handleError(w, r, &AppError{Code: ErrCodeForbidden, Message: "Sem permissão", HTTPStatus: http.StatusForbidden})
            return
        }
        next(w, r, u)
    }
}
```

### Template Rendering (HTML handlers)
**Source:** `cmd/server/main.go:144-150`
**Apply to:** Admin HTML handlers
```go
func (a *App) render(w http.ResponseWriter, name string, data any) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := a.tpl.ExecuteTemplate(w, name, data); err != nil {
        slog.Error("template render failed", "template", name, "error", err)
        http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
    }
}
```

### Existing Service Function Patterns (for new service functions to follow)

**Return domain types, not HTTP types** — `activityDetailsData` (db.go:325-370):
```go
func (a *App) activityDetailsData(ctx context.Context, id int) (Activity, []ProductVerification, error) {
    // ... DB queries that return domain types, never http.Request/ResponseWriter ...
    return acts[0], items, nil
}
```

**Transaction pattern** — currently in `apiFinalizar` (activity_handlers.go:161-166). After extraction, this pattern moves into `finalizeActivity` service function:
```go
tx, err := a.pg.BeginTx(ctx, nil)
if err != nil { return nil, &AppError{...} }
defer tx.Rollback()
// ... business logic ...
if err := tx.Commit(); err != nil {
    return nil, &AppError{...}
}
```

### Validation (Validator type)
**Source:** `cmd/server/validation.go:8-63`
**Apply to:** Service function input validation
```go
v := NewValidator()
v.Required("empresa", strconv.Itoa(req.Empresa))
v.Positive("seqlocal", req.SeqLocal)
v.Required("rua", req.Rua)
if !v.IsValid() {
    return nil, &AppError{Code: ErrCodeValidation, Message: v.Error(), HTTPStatus: http.StatusBadRequest}
}
```

## Service Function Signature Conventions

| Convention | Rule | Source |
|------------|------|--------|
| Context first param | `func (a *App) svcName(ctx context.Context, ...) (..., error)` | RESEARCH.md ~$270 |
| No HTTP types | Never `http.Request` or `http.ResponseWriter` in signature | HAND-03 requirement |
| *App receiver OK | `a.pg`, `a.ora` accessible via receiver | App has no HTTP fields |
| Domain params only | e.g., `(ctx, req finalizeReq, userID int)` not parsed from HTTP | RESEARCH.md ~$280 |
| Return `*AppError` | Error type known to `handleError` dispatcher | errors.go:25-35 |
| TX ownership | Service owns BeginTx/Commit/Rollback | Pitfall 1 (RESEARCH.md ~$514) |

## Handler Signature Conventions

| Handler Type | Signature | Route Registration |
|---|---|---|
| Admin (HTML) | `func (a *App) handlerName(w http.ResponseWriter, r *http.Request)` | `mux.HandleFunc("POST /admin/users/{id}", a.requireRole("sysadmin", a.adminUpdateUser))` |
| Page (HTML) | `func (a *App) handlerName(w http.ResponseWriter, r *http.Request)` | `mux.HandleFunc("GET /admin", a.requireRole("sysadmin", a.adminPage))` |
| API (JSON) | `func (a *App) handlerName(w http.ResponseWriter, r *http.Request, u *User)` | `mux.HandleFunc("POST /api/...", a.requireAPIRole("...", a.handler))` |

## No Analog Found

All files have exact analogs (the existing handler files themselves). No new files are being created — all changes are in-place modifications.

## Metadata

**Analog search scope:** `cmd/server/` (all 15 files)
**Files scanned:** 12 source files + 1 test file
**Pattern extraction date:** 2026-06-09
**Single main package constraint:** All files are in `package main` — no sub-packages or internal modules.
