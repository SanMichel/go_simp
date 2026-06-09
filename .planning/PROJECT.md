# go-simp

## What This Is

A warehouse activity scanning and product logistics dashboard. Warehouse workers (conferentes) scan and track activities, managers (gerentes) monitor operations, and sysadmins manage users. Migrated from TypeScript to Go for performance and simplicity.

## Core Value

Warehouse workers must be able to scan and register activities reliably without the system getting in their way.

## Current Milestone: v1.1 Simplify & Stabilize

**Goal:** Simplify the codebase, improve maintainability, and assure compatibility with warehouse devices running non-Chrome browsers.

**Target features:**
- Handler decomposition and code organization
- Comprehensive test coverage and standardized error handling
- ES5-compatible frontend for legacy warehouse browsers

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

- [ ] **CODE-01**: Split overgrown handler functions into smaller, focused functions
- [ ] **CODE-02**: Standardize error handling (consistent format, logging, user feedback)
- [ ] **CODE-03**: Unify JSON response patterns across all handlers
- [ ] **CODE-04**: Standardize input validation for form/JSON endpoints
- [ ] **MAINT-01**: Improve file organization within `cmd/server/`
- [ ] **MAINT-02**: Achieve comprehensive test coverage (70%+)
- [ ] **MAINT-03**: Establish consistent middleware chain pattern
- [ ] **COMPAT-01**: Ensure ES5 compatibility for non-Chrome warehouse browsers
- [ ] **COMPAT-02**: Optimize page weight/render for low-end devices

### Out of Scope

- Native mobile apps — web-first; PWAs fill the gap for now
- Offline support — too complex for current scope; revisit in v2
- Architectural split into sub-packages — keep single `main` package, improve within
- Third-party frameworks — stdlib + HTMX stays

## Context

This is an existing Go monolith in `cmd/server/` using the standard library exclusively. The codebase was migrated from a TypeScript/Bun/Drizzle stack; a reference copy lives in `tmp/`. There are no external web frameworks — only `net/http` + `database/sql` with pgx (Postgres) and go-ora (Oracle). Templates use Go `html/template` with HTMX for interactivity.

Warehouse devices run non-Chrome browsers (ES5-only) and have built-in barcode scanners that feed input as keyboard text. The single `main` package has grown large; handler functions need decomposition.

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

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-06-08 after milestone v1.1 initialization*
