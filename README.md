# Go SIMP

Go/HTMX rewrite of the `tmp` app. The `tmp` directory is kept unchanged and used only as a reference.

## Run

```bash
cp .env.example .env
go mod tidy
go run ./cmd/server
```

## Development with Hot Reload

For live reloading during development, you can use [Air](https://github.com/air-verse/air):

```bash
air
```

The server listens on `PORT` (`3000` by default).

## Stack

- Go `net/http`
- `database/sql` with `pgx` for local SIMP tables
- `github.com/sijms/go-ora/v2` for Oracle lookups
- HTMX-rendered pages and partials

On first start, the app migrates local tables and seeds `admin` / `admin` as `sysadmin`.
