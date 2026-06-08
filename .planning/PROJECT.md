# go-simp

## What This Is

A warehouse activity scanning and product logistics dashboard. Warehouse workers (conferentes) scan and track activities, managers (gerentes) monitor operations, and sysadmins manage users. Migrated from TypeScript to Go for performance and simplicity.

## Core Value

Warehouse workers must be able to scan and register activities reliably without the system getting in their way.

## Requirements

### Validated

- ✓ User authentication (email/password + session tokens) — existing
- ✓ Role-based access control (conferente, gerente, sysadmin) — existing
- ✓ Activity scanning and registration — existing
- ✓ Dashboard with operational metrics — existing
- ✓ Admin panel for user management — existing
- ✓ Postgres persistence with auto-migrations — existing
- ✓ Oracle integration for product/company reads — existing
- ✓ HTMX-driven SPA interface for scanning workflows — existing
- ✓ CSRF protection and security headers — existing

### Active

(Define via /gsd-new-milestone)

### Out of Scope

(Specific exclusions TBD)

## Context

This is an existing Go monolith in `cmd/server/` using the standard library exclusively. The codebase was migrated from a TypeScript/Bun/Drizzle stack; a reference copy lives in `tmp/`. There are no external web frameworks — only `net/http` + `database/sql` with pgx (Postgres) and go-ora (Oracle). Templates use Go `html/template` with HTMX for interactivity.

## Constraints

- **Tech Stack**: Go 1.23 standard library — no external web frameworks
- **Database**: Postgres primary, Oracle read-only (SELECT/WITH only)
- **Roles**: Three-tier RBAC (conferente, gerente, sysadmin)
- **Testing**: No external test framework — standard `testing` package only
- **Architecture**: Single `main` package — no sub-packages or internal modules

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Single `main` package | Keep things simple for a focused app; avoid premature modularization | ✓ Good |
| Standard library only | Minimize dependencies, easy upgrades, no framework churn | ✓ Good |
| Migrate from TypeScript | Performance, simpler deployment (single binary), fewer runtime deps | ✓ Good |
| HTMX over SPA framework | Server-rendered HTML with sprinkles of JS — simpler than full SPA | ✓ Good |

---
*Last updated: 2026-06-08 after initialization*
