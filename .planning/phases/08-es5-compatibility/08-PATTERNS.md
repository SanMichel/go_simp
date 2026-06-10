# Phase 8: ES5 Compatibility — Pattern Map

**Mapped:** 2026-06-10
**Files analyzed:** 9 (6 new, 3 modified)
**Analogs found:** 9 / 9

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `templates/atividades/atividades-login.js` | script | request-response | `templates/login.js` | exact |
| `templates/atividades/atividades-scan.js` | script | event-driven | `templates/app.js` (lines 1-417) | exact |
| `templates/atividades/atividades-consulta.js` | script | request-response | `templates/app.js` (lines 794-936) | exact |
| `templates/atividades/atividades-utils.js` | utility | utility | `templates/shared.js` | exact |
| `templates/atividades/atividades-login.html` | template | request-response | `templates/login.html` | exact |
| `templates/atividades/atividades.html` | template | request-response | `templates/atividades.html` | exact |
| `cmd/server/main.go` | config | request-response | `cmd/server/main.go` (existing) | exact |
| `cmd/server/handlers.go` | controller | request-response | `cmd/server/handlers.go` (existing) | exact |
| `cmd/server/templates/shared.js` | utility | utility | `cmd/server/templates/shared.js` (existing) | exact |

## Pattern Assignments

### `templates/atividades/atividades-utils.js` (utility, utility)

**Analog:** `templates/shared.js` (94 lines, fully ES5-compatible except `apiCall` and `formatDate`)

**Imports pattern:** No imports — browser global scope. All functions are top-level `function` declarations.

**Core pattern — ES5-compatible functions to copy verbatim:**

**`escHtml` (shared.js lines 87-91):**
```javascript
function escHtml(unsafe) {
  if (unsafe == null)
    return "";
  return String(unsafe).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}
```

**`sanitizeHtml` (shared.js lines 92-93):**
```javascript
function sanitizeHtml(dirty) {
  return escHtml(dirty);
}
```

**`showLoader` (shared.js lines 42-47):**
```javascript
function showLoader(show) {
  var loader = document.getElementById("loader");
  if (loader) {
    loader.classList.toggle("hidden", !show);
  }
}
```

**`playBeep` (shared.js lines 60-85) — Copy verbatim, already ES5:**
```javascript
function playBeep(type) {
  try {
    var AudioContext = window.AudioContext || window.webkitAudioContext;
    if (!AudioContext) return;
    var ctx = new AudioContext();
    var playTone = function(freq, duration, start) {
      var osc = ctx.createOscillator();
      var gain = ctx.createGain();
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.type = "sine";
      osc.frequency.setValueAtTime(freq, ctx.currentTime + start);
      gain.gain.setValueAtTime(0.1, ctx.currentTime + start);
      osc.start(ctx.currentTime + start);
      osc.stop(ctx.currentTime + start + duration);
    };
    if (type === "success") playTone(880, 0.15, 0);
    else if (type === "warning") { playTone(660, 0.1, 0); playTone(660, 0.1, 0.15); }
    else if (type === "error") playTone(440, 0.4, 0);
  } catch (_e) {}
}
```

**`formatDate` (shared.js lines 48-58) — Rewrite to ES5:**
```javascript
function formatDate(dateStr) {
  if (!dateStr) return "";
  var d = new Date(dateStr);
  if (isNaN(d.getTime())) return "";
  function pad(n) { return n < 10 ? "0" + n : "" + n; }
  return pad(d.getDate()) + "/" + pad(d.getMonth() + 1) + "/" + d.getFullYear();
}
```

**Core pattern — XHR-based apiCall (replaces shared.js `async function apiCall` lines 2-40):**
```javascript
function apiCall(method, url, body, onSuccess, onUnauthorized) {
  var xhr = new XMLHttpRequest();
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
  xhr.open(method, url, true);
  xhr.withCredentials = true;
  xhr.setRequestHeader("Content-Type", "application/json");
  // CSRF token for mutating requests
  if (method === "POST" || method === "PATCH" || method === "DELETE") {
    var csrfMatch = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    if (csrfMatch) {
      xhr.setRequestHeader("X-CSRF-Token", csrfMatch[1]);
    }
  }
  xhr.send(body ? JSON.stringify(body) : null);
}

function apiGet(url, onSuccess, onUnauthorized) {
  apiCall("GET", url, null, onSuccess, onUnauthorized);
}
function apiPost(url, body, onSuccess, onUnauthorized) {
  apiCall("POST", url, body, onSuccess, onUnauthorized);
}
```

**Error handling pattern:** No try/catch needed — XHR callbacks handle errors via HTTP status. Non-JSON responses are wrapped as `{ error: "..." }`.

**Validation pattern:** CSRF token extraction uses `document.cookie.match()` — ES5-compatible regex.

---

### `templates/atividades/atividades-login.js` (script, request-response)

**Analog:** `templates/login.js` (69 lines)

**Imports pattern:** No imports — browser global scope. Depends on functions from `atividades-utils.js` (`apiGet`, `apiPost`, `showLoader`, `escHtml`).

**Core pattern — Login form handling (adapted from login.js lines 27-69):**
```javascript
document.addEventListener("DOMContentLoaded", function() {
  apiGet("/api/auth/me", function(ok, status, data) {
    if (ok && data && data.user && data.user.role) {
      window.location.href = "/atividades";
      return;
    }
  });

  document.getElementById("form-atividades-login").addEventListener("submit", function(e) {
    e.preventDefault();
    showLoader(true);
    var errorEl = document.getElementById("login-error");
    if (errorEl) errorEl.classList.add("hidden");
    var username = document.getElementById("login-username").value;
    var password = document.getElementById("login-password").value;
    apiPost("/api/auth/login", { username: username, password: password }, function(ok, status, data) {
      showLoader(false);
      if (ok && data && data.user) {
        window.location.href = "/atividades";
      } else {
        if (errorEl) {
          errorEl.innerText = data && data.error ? data.error : "Erro ao logar";
          errorEl.classList.remove("hidden");
        }
      }
    }, function() {
      showLoader(false);
      if (errorEl) {
        errorEl.innerText = "Sessão expirada";
        errorEl.classList.remove("hidden");
      }
    });
  });
});
```

**Auth pattern:** Session check on DOMContentLoaded via `GET /api/auth/me`. If valid, redirect to `/atividades`. If not, show login form.

**Error handling pattern:** Show error text in `#login-error` element. `showLoader(false)` on both success and failure paths. Unauthorized callback handles 401 on login.

---

### `templates/atividades/atividades-scan.js` (script, event-driven)

**Analog:** `templates/app.js` lines 1-417 (state, scan flow, predio-switch, finalize)

**Imports pattern:** No imports. Depends on `atividades-utils.js` functions (`apiGet`, `apiPost`, `showLoader`, `sanitizeHtml`, `playBeep`, `formatDate`).

**Core pattern — State management (app.js lines 2-30, already ES5):**
```javascript
var state = {
  screen: localStorage.getItem("simp_screen") || "login",
  user: JSON.parse(localStorage.getItem("simp_user") || "null"),
  token: null,
  atividade: JSON.parse(localStorage.getItem("simp_atividade") || "null"),
  scannedProducts: JSON.parse(localStorage.getItem("simp_scanned") || "[]"),
  expectedProducts: JSON.parse(localStorage.getItem("simp_expected") || "[]"),
  lastScanned: null,
  allowKeyboard: false
};
function saveState() {
  localStorage.setItem("simp_screen", state.screen);
  if (state.user) localStorage.setItem("simp_user", JSON.stringify(state.user));
  else localStorage.removeItem("simp_user");
  if (state.atividade) localStorage.setItem("simp_atividade", JSON.stringify(state.atividade));
  else localStorage.removeItem("simp_atividade");
  localStorage.setItem("simp_scanned", JSON.stringify(state.scannedProducts));
  localStorage.setItem("simp_expected", JSON.stringify(state.expectedProducts));
}
function resetActivityState() {
  state.atividade = null;
  state.scannedProducts = [];
  state.expectedProducts = [];
  saveState();
}
```

**Core pattern — Screen navigation (app.js lines 174-228, ES5 rewrite of `showScreen`):**
```javascript
function showScreen(screenId) {
  if (screenId === "consulta") {
    state.previousScreen = state.screen;
  }
  state.screen = screenId;
  saveState();
  var screens = document.querySelectorAll(".screen");
  for (var i = 0; i < screens.length; i++) {
    screens[i].classList.add("hidden");
  }
  var targetScreen = document.getElementById("screen-" + screenId);
  if (targetScreen) targetScreen.classList.remove("hidden");
  syncKeyboardUI();
  var isAuth = screenId !== "login";
  var headerActions = document.getElementById("header-actions");
  if (headerActions) headerActions.classList.toggle("hidden", !isAuth);
  // ... further screen-specific logic
}
```

**Core pattern — Async function replaced with XHR callback (app.js lines 315-356 `startActivity` → ES5):**
```javascript
function startActivity() {
  var empSelect = document.getElementById("start-empresa");
  var empresa = empSelect.value;
  var seqlocal = document.getElementById("start-local").value;
  var rua = document.getElementById("start-rua").value;
  var predio = document.getElementById("start-predio").value;
  var errEl = document.getElementById("start-error");
  if (!seqlocal) {
    if (errEl) { errEl.innerText = "Selecione um local válido"; errEl.classList.remove("hidden"); }
    return;
  }
  showLoader(true);
  apiGet("/api/produtos/local?empresa=" + encodeURIComponent(empresa) + "&seqlocal=" + encodeURIComponent(seqlocal) + "&rua=" + encodeURIComponent(rua) + "&predio=" + encodeURIComponent(predio), function(ok, status, data) {
    showLoader(false);
    if (ok && Array.isArray(data) && data.length > 0) {
      state.expectedProducts = data;
      state.atividade = { id: 0, empresa: empresa, seqlocal: seqlocal, rua: rua, predio: predio, predios: [predio], currentPredio: predio, status: "aberta", dataInicio: new Date().toISOString() };
      state.scannedProducts = [];
      saveState();
      showScreen("scanning");
    } else {
      if (errEl) { errEl.innerText = "Endereço não encontrado ou sem produtos"; errEl.classList.remove("hidden"); }
    }
  }, function() { showLoader(false); /* onUnauthorized */ });
}
```

**Core pattern — EAN scan handler (app.js lines 619-690, ES5 rewrite):**
```javascript
document.getElementById("form-scan").addEventListener("submit", function(e) {
  e.preventDefault();
  var input = document.getElementById("scan-input");
  var code = input.value.trim();
  if (!code || !state.atividade) return;
  showLoader(true);
  apiGet("/api/produtos/ean/" + encodeURIComponent(code) + "?empresa=" + state.atividade.empresa + "&seqlocal=" + state.atividade.seqlocal, function(ok, status, data) {
    showLoader(false);
    var feedback = document.getElementById("scan-feedback");
    if (!feedback) return;
    if (!ok) {
      playBeep("error");
      feedback.innerHTML = sanitizeHtml('<div style="color: #ef4444; font-weight: bold;">❌ Produto não encontrado</div>');
      input.select();
      return;
    }
    // ... divergence detection logic (same as app.js lines 641-689)
  }, function() { showLoader(false); });
});
```

**Core pattern — Building switch (app.js lines 693-751, ES5 rewrite):**
```javascript
document.getElementById("btn-predio-switch-yes").addEventListener("click", function() {
  if (!state.atividade || !state.lastScanned) return;
  try {
    var newPredio = String(state.lastScanned.predio);
    var predios = state.atividade.predios || [state.atividade.predio];
    var isNewBuilding = predios.indexOf(newPredio) === -1;
    // ... rest of building switch logic
  } catch (e) {
    console.error("Error during building switch:", e);
  }
});
```

**Error handling pattern:** `showLoader(false)` is always called in both success/failure XHR callbacks. `try/catch` for synchronous code blocks. All `innerHTML` assignments use `sanitizeHtml()`.

**`.includes()` → `indexOf` guard (app.js line 195, ES5 rewrite):**
```javascript
// Instead of: ["scanning","divergence","predio-switch","consulta"].includes(screenId)
// Use:
var hideGlobalHeader = screenId === "scanning" || screenId === "divergence" || screenId === "predio-switch" || screenId === "consulta";
```

**`.forEach` → index-based loop (app.js patterns):**
```javascript
// Instead of: arr.forEach(function(x) { ... })
// Use:
for (var i = 0; i < arr.length; i++) {
  var x = arr[i];
  // ...
}
```

**Optional chaining `?.` → `&&` guard (app.js lines 166-167, ES5 rewrite):**
```javascript
// Instead of: document.getElementById("modal-product-detail")?.classList.add("hidden");
// Use:
var el = document.getElementById("modal-product-detail");
if (el) el.classList.add("hidden");
```

**Template literals → string concatenation (app.js line 205, ES5 rewrite):**
```javascript
// Instead of: `${userName.substring(0, 10)}..`
// Use:
userName.substring(0, 10) + ".."
```

---

### `templates/atividades/atividades-consulta.js` (script, request-response)

**Analog:** `templates/app.js` lines 794-936 (product consultation)

**Imports pattern:** No imports. Depends on `atividades-utils.js` functions.

**Core pattern — Consulta mode toggle (app.js lines 795-824, ES5):**
```javascript
var consultaMode = "codigo";
function setConsultaMode(mode) {
  consultaMode = mode;
  var input = document.getElementById("consulta-input");
  var btnCodigo = document.getElementById("btn-consulta-mode-codigo");
  var btnDescricao = document.getElementById("btn-consulta-mode-descricao");
  if (mode === "codigo") {
    input.placeholder = "Escanear ou digitar EAN/DUN";
    input.type = "tel";
    input.inputMode = "none";
    btnCodigo.className = "btn btn-sm btn-primary";
    btnDescricao.className = "btn btn-sm";
  } else {
    input.placeholder = "Digitar descrição do produto";
    input.type = "text";
    input.inputMode = "text";
    btnDescricao.className = "btn btn-sm btn-primary";
    btnCodigo.className = "btn btn-sm";
  }
  // ... hide results, clear input
}
```

**Core pattern — Consulta form submit (app.js lines 830-936, ES5 rewrite):**
```javascript
document.getElementById("form-consulta").addEventListener("submit", function(e) {
  e.preventDefault();
  var input = document.getElementById("consulta-input");
  var code = input.value.trim();
  if (!code) return;
  // ... determine empresa/seqlocal context
  showLoader(true);
  if (consultaMode === "descricao") {
    apiGet("/api/produtos/consulta?q=" + encodeURIComponent(code) + "&empresa=" + empresa + "&seqlocal=" + seqlocal, function(ok, status, data) {
      showLoader(false);
      // ... render results list with sanitizeHtml()
    });
  } else {
    apiGet("/api/produtos/consulta/" + encodeURIComponent(code) + "?empresa=" + empresa + "&seqlocal=" + seqlocal, function(ok, status, data) {
      showLoader(false);
      // ... populate detail view with innerText/sanitizeHtml()
    });
  }
});
```

**Error handling pattern:** Same as scan.js — `showLoader(false)` in both success/failure callbacks. `playBeep("error")` on failure. `alert()` for not-found cases.

---

### `templates/atividades/atividades-login.html` (template, request-response)

**Analog:** `templates/login.html` (43 lines)

**Core pattern — Isolated login template (login.html lines 1-42):**
```html
{{define "atividades-login"}}
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>SIMP — Atividades Login</title>
    <link rel="stylesheet" href="/style.css" />
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
                <div class="form-group">
                    <label class="label">Usuário</label>
                    <input type="text" id="login-username" class="input" required placeholder="Seu usuário">
                </div>
                <div class="form-group">
                    <label class="label">Senha</label>
                    <div class="input-group">
                        <input type="password" id="login-password" class="input" required placeholder="••••••••">
                    </div>
                </div>
                <p id="login-error" class="text-red-500 text-sm text-center hidden"></p>
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

**Key differences from `login.html`:**
- Uses `{{define "atividades-login"}}` (new template name)
- Uses `/style.css` not `/admin.css`
- Uses `id="form-atividades-login"` (not `form-login-page`) to avoid conflict
- Scripts load `/atividades-utils.js` and `/atividades-login.js` (not `/shared.js`, `/login.js`)
- No re-auth modal, no header

---

### `templates/atividades/atividades.html` (template, request-response)

**Analog:** `templates/atividades.html` (334 lines, existing)

**Core pattern — Modified main SPA template (atividades.html adapted):**
- Remove `#screen-login` div (lines 43-65) — login now served by `atividades-login.html`
- Remove `#modal-reauth` (lines 269-288) — re-auth handled by redirect to login
- Keep screens: `screen-start`, `screen-scanning`, `screen-predio-switch`, `screen-consulta`, `screen-report`
- Keep `modal-product-detail` (lines 291-328)
- Script tags at bottom change from:
```html
<script src="/shared.js"></script>
<script src="/app.js"></script>
```
To:
```html
<script src="/atividades-utils.js"></script>
<script src="/atividades-scan.js"></script>
<script src="/atividades-consulta.js"></script>
<script src="/htmx.min.js"></script>
```

---

### `cmd/server/main.go` (config, request-response) — MODIFIED

**Analog:** `cmd/server/main.go` existing lines

**go:embed pattern update (line 215, add subdirectories):**
```go
//go:embed templates/*.html templates/components/*.html templates/*.css templates/*.js templates/atividades/*.html templates/atividades/*.js
var templatesFS embed.FS
```

**ParseFS pattern update (line 212, add templates/atividades/*.html):**
```go
return template.Must(template.New("app").Funcs(funcs).ParseFS(templatesFS, "templates/*.html", "templates/components/*.html", "templates/atividades/*.html"))
```

**Route registration pattern (lines 94-143, add new routes):**
```go
// New routes for atividades login and JS files
mux.HandleFunc("GET /atividades/login", a.atividadesLoginPage)
mux.HandleFunc("GET /atividades-utils.js", a.serveJS("atividades/atividades-utils.js"))
mux.HandleFunc("GET /atividades-login.js", a.serveJS("atividades/atividades-login.js"))
mux.HandleFunc("GET /atividades-scan.js", a.serveJS("atividades/atividades-scan.js"))
mux.HandleFunc("GET /atividades-consulta.js", a.serveJS("atividades/atividades-consulta.js"))
```

**serveJS pattern (lines 171-182) — already supports subdirectory paths:**
```go
func (a *App) serveJS(filename string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
        b, err := templatesFS.ReadFile("templates/" + filename)
        if err != nil {
            slog.Error("failed to read static file", "file", filename, "error", err)
            http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
            return
        }
        _, _ = w.Write(b)
    }
}
```

**requireRole middleware pattern (auth.go lines 86-110) — needs new `requireAtividadesRole`:**
```go
func (a *App) requireAtividadesRole(roles string, next http.HandlerFunc) http.HandlerFunc {
    allowed := map[string]bool{}
    for _, r := range strings.Split(roles, ",") {
        if r != "" { allowed[r] = true }
    }
    return func(w http.ResponseWriter, r *http.Request) {
        u, err := a.currentUser(r)
        if err != nil {
            http.Redirect(w, r, "/atividades/login", http.StatusFound)
            return
        }
        if len(allowed) > 0 && !allowed[u.Role] {
            if u.Role == "conferente" {
                http.Redirect(w, r, "/atividades", http.StatusFound)
            } else {
                http.Redirect(w, r, "/dashboard", http.StatusFound)
            }
            return
        }
        ctx := context.WithValue(r.Context(), ctxUser, u)
        next(w, r.WithContext(ctx))
    }
}
```

**Route registration for `/atividades` with new middleware (main.go line 110):**
```go
// From:
mux.HandleFunc("GET /atividades", a.requireRole("", a.atividadesPage))
// To:
mux.HandleFunc("GET /atividades", a.requireAtividadesRole("", a.atividadesPage))
```

---

### `cmd/server/handlers.go` (controller, request-response) — MODIFIED

**Analog:** `cmd/server/handlers.go` existing handler patterns

**New handler — `atividadesLoginPage` (insert after `atividadesPage` at line 65):**
```go
func (a *App) atividadesLoginPage(w http.ResponseWriter, r *http.Request) {
    u, err := a.currentUser(r)
    if err == nil {
        // Already logged in — redirect to /atividades
        a.redirectByRole(w, r, u.Role)
        return
    }
    a.render(w, "atividades-login", nil)
}
```

**Handler pattern reference — `atividadesPage` (line 65-68, stays same):**
```go
func (a *App) atividadesPage(w http.ResponseWriter, r *http.Request) {
    u := r.Context().Value(ctxUser).(*User)
    a.render(w, "atividades", map[string]any{"User": u})
}
```

**Handler test pattern (main_test.go lines 1677-1691 `TestAtividadesPageAuthenticated`):**
```go
func TestAtividadesLoginPage_Unauthenticated(t *testing.T) {
    app := &App{tpl: parseTemplates()}
    req := httptest.NewRequest(http.MethodGet, "/atividades/login", nil)
    rec := httptest.NewRecorder()
    app.atividadesLoginPage(rec, req)
    if rec.Code != http.StatusOK {
        t.Errorf("atividadesLoginPage should return 200 when unauthenticated, got %d", rec.Code)
    }
    if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
        t.Errorf("content-type = %q, want text/html", ct)
    }
}

func TestAtividadesLoginPage_Authenticated(t *testing.T) {
    app := testApp(t)
    user := testUser(t, app, "test_atividades_login_auth", "conferente", "pass1234")
    req := httptest.NewRequest(http.MethodGet, "/atividades/login", nil)
    ctx := context.WithValue(req.Context(), ctxUser, user)
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()
    app.atividadesLoginPage(rec, req)
    if rec.Code != http.StatusFound {
        t.Errorf("atividadesLoginPage should redirect (302) when authenticated, got %d", rec.Code)
    }
}
```

---

### `cmd/server/templates/shared.js` (utility, utility) — MODIFIED

**Analog:** `templates/shared.js` existing file

**Modification pattern — Add comment above each copied function:**
```javascript
// COPIED to templates/atividades/atividades-utils.js — TODO: remove when admin/dashboard migrate to ES5 (ES5-05)
```
Insert above each of these functions:
- `function apiCall(...)` (line 2)
- `function showLoader(show)` (line 42)
- `function formatDate(dateStr)` (line 48)
- `function playBeep(type)` (line 60) — if not already commented
- `function escHtml(unsafe)` (line 87)
- `function sanitizeHtml(dirty)` (line 92)

---

## Shared Patterns

### ES5 Syntax Rules (All new JS files)
**Source:** RESEARCH.md §ES5 Syntax Constraint Reference
**Apply to:** All 4 new JS files in `templates/atividades/`

| Modern Pattern | ES5 Equivalent |
|---------------|----------------|
| `const/let x` | `var x` |
| `() => {}` | `function() {}` |
| `async/await` + `fetch` | XMLHttpRequest callbacks |
| `` `text ${var}` `` | `"text " + var` |
| `user?.name` | `user && user.name` |
| `arr.includes(x)` | `arr.indexOf(x) !== -1` |
| `arr.find(fn)` | Manual `for` loop |
| `arr.some(fn)` | `for` loop with flag |
| `for...of` | `for (var i = 0; i < arr.length; i++)` |
| `[...arr]` | `arr.slice(0)` |
| `func(x = 5)` | `function(x) { if (x === undefined) x = 5; }` |

### XHR Callback Convention (All 4 new JS files)
**Source:** RESEARCH.md §XHR Wrapper Design
**Apply to:** All API calls

```javascript
// All async operations follow this pattern:
function doSomething() {
  showLoader(true);
  apiGet("/api/endpoint", function(ok, status, data) {
    showLoader(false);
    if (ok) {
      // success — data is parsed JSON
    } else {
      // failure — data has error message
    }
  }, function() {
    // onUnauthorized callback
    showLoader(false);
    window.location.href = "/atividades/login";
  });
}
```

### innerHTML Sanitization (All 4 new JS files)
**Source:** `templates/shared.js` lines 87-93
**Apply to:** All DOM assignments with user-supplied values

```javascript
// Use innerText when no HTML is needed:
el.innerText = userValue;
// Use sanitizeHtml() when innerHTML is unavoidable:
el.innerHTML = sanitizeHtml('<div>' + userValue + '</div>');
```

### DOMContentLoaded Pattern (all 4 new JS files)
**Source:** `templates/login.js` line 41, `templates/app.js` line 443
**Apply to:** All initialization code

```javascript
document.addEventListener("DOMContentLoaded", function() {
  // Initialize state, check session, bind event handlers
  // All var declarations at top of this scope
});
```

### Go Template Naming Convention (both new HTML templates)
**Source:** `templates/login.html` line 1, `templates/atividades.html` line 1
**Apply to:** Both new HTML templates

```html
{{define "atividades-login"}}  <!-- or {{define "atividades"}} -->
<!DOCTYPE html>
<html lang="pt-BR">
<!-- ... -->
</html>
{{end}}
```

### Handler Method Convention (handlers.go)
**Source:** `handlers.go` lines 10-68
**Apply to:** `atividadesLoginPage` handler

```go
func (a *App) atividadesLoginPage(w http.ResponseWriter, r *http.Request) {
    // 1. Check auth condition
    u, err := a.currentUser(r)
    if err == nil {
        // 2. Redirect if condition met
        http.Redirect(w, r, "/atividades", http.StatusFound)
        return
    }
    // 3. Render template
    a.render(w, "atividades-login", nil)
}
```

## No Analog Found

All 9 files have exact matches in the codebase. No files with no analog.

## Metadata

**Analog search scope:** `cmd/server/` (Go), `cmd/server/templates/` (JS/HTML)
**Files scanned:** 12 (main.go, handlers.go, auth.go, errors.go, utils.go, validation.go, main_test.go, app.js, shared.js, login.js, login.html, atividades.html)
**Pattern extraction date:** 2026-06-10
