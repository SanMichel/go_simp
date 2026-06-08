# Roadmap: go-simp

## Overview

Warehouse activity scanning and logistics dashboard. Existing codebase migrated from TypeScript to Go.

## Phases

- [x] **Phase 1: Foundation** — Go migration, core architecture, auth, DB setup
- [x] **Phase 2: Activity Scanning** — HTMX-driven scanning SPA, activity registration
- [x] **Phase 3: Dashboard** — Operational metrics and monitoring dashboard
- [x] **Phase 4: Admin** — User management, RBAC administration

(TBD via /gsd-new-milestone)

## Phase Details

### Phase 1: Foundation
**Goal**: Go server with auth, database, and basic structure
**Depends on**: Nothing
**Success Criteria**:
  1. Server starts and serves HTTP
  2. Users can log in with email/password
  3. Postgres auto-migrates on startup
**Plans**: N/A (existing)

### Phase 2: Activity Scanning
**Goal**: Warehouse scanning workflows
**Depends on**: Phase 1
**Success Criteria**:
  1. Conferentes can scan and register activities
  2. Activity data persisted to Postgres
  3. HTMX partials render activity tables and modals
**Plans**: N/A (existing)

### Phase 3: Dashboard
**Goal**: Operational metrics dashboard
**Depends on**: Phase 2
**Success Criteria**:
  1. Gerentes can view real-time operational data
  2. Dashboard SPA loads and updates properly
**Plans**: N/A (existing)

### Phase 4: Admin
**Goal**: User management panel
**Depends on**: Phase 1
**Success Criteria**:
  1. Sysadmins can create, edit, delete users
  2. Role assignment works correctly
  3. Admin SPA with HTMX dynamic rows
**Plans**: N/A (existing)

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | — | Complete | 2026-06-08 |
| 2. Activity Scanning | — | Complete | 2026-06-08 |
| 3. Dashboard | — | Complete | 2026-06-08 |
| 4. Admin | — | Complete | 2026-06-08 |
