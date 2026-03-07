# Wire Dead Code & Update Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire all disconnected handlers/services into the router, delete orphaned code, replace coverage boost tests with meaningful tests, and add v10 migration for sync tables.

**Architecture:** SearchHandler and BrowseHandler already exist with full implementations - just need route registration in main.go. SyncService needs a new SyncHandler (Gin HTTP handler), a v10 database migration for sync_endpoints/sync_sessions/sync_schedules tables, and route registration.

**Tech Stack:** Go 1.24, Gin, SQLite/PostgreSQL dual dialect, existing test helpers

---

### Task 1: Add v10 Migration for Sync Tables

**Files:**
- Modify: `catalog-api/database/migrations_sqlite.go` (add migrateV10SQLite)
- Modify: `catalog-api/database/migrations_postgres.go` (add migrateV10Postgres)
- Modify: `catalog-api/database/migrations.go` (register v10)

**What:** Add sync_endpoints, sync_sessions, sync_schedules tables matching the schema already in test_helper_test.go.

### Task 2: Create SyncHandler

**Files:**
- Create: `catalog-api/handlers/sync_handler.go`

**What:** Gin handler wrapping SyncService methods with JSON request/response, JWT user extraction, and standard error handling.

### Task 3: Wire All Handlers in main.go

**Files:**
- Modify: `catalog-api/main.go`

**What:** Instantiate SearchHandler, BrowseHandler, SyncHandler and register their routes under `/api/v1/`.

### Task 4: Delete Orphaned Code

**Files:**
- Delete: `catalog-api/smb/` (entire directory - 4 files)
- Delete: `catalog-api/pkg/` (empty directory)
- Delete: `catalog-api/old_results/` (stale artifacts)
- Delete: `catalog-api/test-results/` (stale artifacts)

### Task 5: Replace Coverage Boost Tests

**Files:**
- Rewrite: `catalog-api/handlers/coverage_boost_test.go` through `coverage_boost4_test.go`
- Rewrite: Other coverage boost files

**What:** Replace artificial coverage boosters with meaningful tests for newly wired handlers.

### Task 6: Add Tests for SearchHandler

**Files:**
- Create: `catalog-api/handlers/search_test.go`

### Task 7: Add Tests for BrowseHandler

**Files:**
- Create: `catalog-api/handlers/browse_test.go`

### Task 8: Add Tests for SyncHandler

**Files:**
- Create: `catalog-api/handlers/sync_handler_test.go`

### Task 9: Verify Everything Compiles and Tests Pass

**Run:** `cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
