# go-simp

Single `main` package in `cmd/server/main.go`. No sub-packages, no internal modules.

## Stack

- Go `net/http` + `database/sql` with pgx (Postgres) + go-ora (Oracle, **read-only**)
- HTMX + inline Go templates — templates and CSS are `const` strings in `main.go`, **not separate files**

## Quick start

```bash
cp .env.example .env   # EDIT .env first — POSTGRES_URL is required
go mod tidy
go run ./cmd/server     # or: air (hot reload via .air.toml)
```

On first start, auto-migrates tables and seeds `admin`/`admin` as `sysadmin`.

## Tests

```bash
go test ./cmd/server
```

All 3 tests are in `cmd/server/main_test.go`. No external framework, no integration prerequisites.

## Key gotchas

- `.env` is loaded manually via `loadDotEnv()` — not godotenv
- Oracle connection is guarded by `isReadOnlySQL()` — only `SELECT`/`WITH` pass
- Templates and CSS live inside Go source (two large `const` blocks); editing them requires recompilation
- `tmp/` is a reference copy of the old app — **do not modify**
- `bin/` and `.tmp/` are build artifacts (gitignored)
- Roles: `conferente`, `gerente`, `sysadmin`
