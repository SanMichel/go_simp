# Technology Stack

**Project:** go-simp v1.1 Simplify & Stabilize
**Researched:** 2026-06-08

## Recommended Stack

### Core Framework
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.23+ | Application runtime | Existing. Single binary, fast compilation, excellent stdlib. Keep. |
| net/http | 1.22+ (stdlib) | HTTP server + routing | Existing. Go 1.22 enhanced ServeMux with method patterns and path variables. No framework needed. |
| database/sql | stdlib | Database abstraction | Existing. Standard Go interface, works with pgx and go-ora drivers. |
| html/template | stdlib | Server-side rendering | Existing. Embedded templates via go:embed, FuncMap for helpers. |

### Database
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| PostgreSQL (pgx) | v5 (stdlib driver) | Primary database | Existing. Auto-migrations on startup. Connection pooling configurable. |
| Oracle (go-ora) | v2 | Read-only product queries | Existing. Wrapped in OracleReader with SQL injection guard. |
| sql.Null* types | stdlib | Nullable column mapping | Existing. Required for Oracle queries with NULL values. |

### Middleware (stdlib, no external frameworks)
| Component | Implementation | Purpose |
|-----------|---------------|---------|
| CSRF Protection | Double-submit cookie pattern | Existing in auth.go. Cookie + header comparison for APIs, Origin header check for forms. |
| Security Headers | Response header middleware | Existing in utils.go. CSP, nosniff, X-Frame-Options, HSTS, Referrer-Policy. |
| Logging | ResponseWriter wrapper | Existing in utils.go. Method, path, status code, duration. |
| Rate Limiter | In-memory map with periodic cleanup | Existing in utils.go. Per-IP, 5 attempts/minute window. |
| Recovery (NEW) | defer/recover middleware | **Proposed.** Catches panics, logs stack trace, returns 500. Zero overhead when no panic. |

### Supporting Libraries
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| golang.org/x/crypto/bcrypt | latest | Password hashing | Existing. User creation and login. |
| github.com/jackc/pgx/v5/stdlib | latest | Postgres stdlib driver | Existing. Used via database/sql. |
| github.com/sijms/go-ora/v2 | latest | Oracle stdlib driver | Existing. Used via database/sql. Read-only enforcement wrapper. |

### Frontend
| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| Go html/template | stdlib | HTML templates | Embedded via go:embed. Separate .html files in templates/. |
| HTMX | ~2.x (current) / 1.x (ES5 compat) | AJAX + HTML swapping | **NEEDS VERIFICATION** — check if 1.x targets ES5. 2.0.4 uses ES6+. |
| Vanilla JS | ES5 (after conversion) | Client interactivity | Currently ES6+. **NEEDS CONVERSION** for warehouse browsers. |
| DOMPurify | Inlined 3.4.2 (REMOVE) | HTML sanitization | **INCOMPATIBLE** — full inlined library is ES6+. Replace with escHtml. |

### Testing
| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| testing | stdlib | Unit/integration tests | Existing. No external test framework. |
| net/http/httptest | stdlib | HTTP handler testing | Existing and sufficient. Use NewRequest + NewRecorder. |
| testing.T.Run | stdlib | Table-driven subtests | Idiomatic Go pattern. Each test case is self-documenting. |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Web framework | net/http (stdlib) | chi, gorilla/mux, httprouter | Project constraint: no external web frameworks. Stdlib works well for this size. |
| Error handling | appError + handleError adapter | upspin.io/errors, pkg/errors | Project constraint: minimize dependencies. Simple adapter pattern is sufficient. |
| Test framework | testing (stdlib) | testify, ginkgo | Project constraint: no external test frameworks. Table-driven subtests are idiomatic and sufficient. |
| JS transpiler | Babel CLI (build step) | esbuild, swc | If ES5 transpilation is needed, Babel is the most mature option. Could also pre-transpile and commit. |
| HTML sanitizer | escHtml (existing) | DOMPurify, bluemonday | App renders no user-generated HTML. Basic escaping is sufficient. |

## Installation

No stack changes require new Go dependencies. The existing `go.mod` already has all runtime packages.

```bash
# If Babel is added for JS transpilation:
npm install -D @babel/cli @babel/preset-env

# Or pre-transpile JS and commit the output:
npx babel templates/*.js --out-dir templates/es5/ --presets=@babel/preset-env
```

## Sources

- Go 1.22 ServeMux patterns: https://go.dev/blog/routing-enhancements
- Go httptest: https://pkg.go.dev/net/http/httptest
- Go testing subtests: https://go.dev/blog/subtests
- HTMX browser support: https://htmx.org/docs/#browser-support (v2.x requires ES6)
- DOMPurify source: verified inlined version in shared.js uses ES6+ syntax (const, let, arrow functions, Object.entries, for...of, spread operator)
- ES5 test reference: https://caniuse.com/es5
