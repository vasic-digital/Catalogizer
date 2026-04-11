# Module 31 — Database Dialect Rewriting & Migration Parity

**Duration:** 18 minutes
**Prerequisites:** Module 2 (Backend), Module 29 (Module Architecture)
**Learning objectives:**

By the end of this module, you will:
1. Understand why Catalogizer uses a dual-dialect SQL abstraction (SQLite for dev, PostgreSQL for production).
2. Trace a query through the `database.DB` wrapper and see exactly which rewrites happen.
3. Write a new migration that ships for both dialects without duplication.
4. Run the migration parity boot test and interpret its failure modes.

## Segment 1 — Why dual dialect? (0:00 – 2:30)

Catalogizer's developers want a zero-dependency SQLite DB for unit tests and local development. Production operators want PostgreSQL for concurrency, connection pooling, and backup tooling. Duplicating every query across two codebases is a maintenance nightmare, so we use a thin wrapper that rewrites at runtime.

**Show on screen:** `catalog-api/database/dialect.go` — the `DialectType` enum and the three rewrite functions.

**Key talking points:**
- Placeholder rewriting: `?` → `$1, $2, $3, …` for PostgreSQL.
- `INSERT OR IGNORE` → `ON CONFLICT DO NOTHING`.
- Boolean literals `= 0/1` → `= FALSE/TRUE` for known boolean columns.

## Segment 2 — The `database.DB` wrapper (2:30 – 6:30)

**Show on screen:** `catalog-api/database/connection.go` — `database.DB` wraps `*sql.DB` with shadowed `Exec()`, `Query()`, `QueryRow()` that auto-rewrite SQL.

**Demo:**
```go
// Caller writes portable SQL with ? placeholders
rows, err := db.Query(
    "SELECT id, title FROM media_items WHERE type = ? AND active = 1",
    "movie",
)
```

The wrapper transparently converts this to:
```sql
SELECT id, title FROM media_items WHERE type = $1 AND active = TRUE
```

on PostgreSQL. On SQLite, it passes through unchanged.

**Gotcha:** `db.Exec` with a bare `*sql.DB` (not the wrapper) **breaks** on PostgreSQL. Always go through the wrapper.

## Segment 3 — `InsertReturningID` (6:30 – 9:00)

`LastInsertId()` is a SQLite-only feature. PostgreSQL uses `RETURNING id`. The wrapper exposes `InsertReturningID()` and `TxInsertReturningID()` that handle both:

```go
id, err := db.InsertReturningID(ctx,
    "INSERT INTO media_items (title, type) VALUES (?, ?)",
    "title", "movie",
)
```

On SQLite: executes, calls `LastInsertId()`.
On PostgreSQL: appends `RETURNING id` to the query, scans the result.

## Segment 4 — Migration parity (9:00 – 14:00)

**Show on screen:** `catalog-api/database/migrations.go` — the `RunMigrations()` function.

Each migration is a Go function that dispatches on dialect:

```go
func (db *DB) createMediaEntityTables(ctx context.Context) error {
    if db.dialect.IsPostgres() {
        return db.createMediaEntityTablesPostgres(ctx)
    }
    return db.createMediaEntityTablesSQLite(ctx)
}
```

The per-dialect implementations live in `migrations_postgres.go` and `migrations_sqlite.go`. Adding a new migration means:

1. New entry in `RunMigrations()`.
2. Dispatch function in `migrations.go`.
3. Postgres impl in `migrations_postgres.go`.
4. SQLite impl in `migrations_sqlite.go`.
5. Reference `.up.sql` + `.sqlite.up.sql` in `database/migrations/` (for golang-migrate CLI compatibility).

## Segment 5 — The parity boot test (14:00 – 16:30)

**Show on screen:** `catalog-api/database/migrations_parity_test.go::TestRunMigrationsSQLite`.

This test boots a fresh SQLite DB, runs the full migration chain, and asserts the `migrations` table has ≥14 rows afterward. If any dispatch function is missing its SQLite implementation, the boot fails immediately.

**Demo:** add a new migration without the SQLite implementation, run the test, watch it fail. Add the implementation, watch it pass.

## Segment 6 — Real-world gotchas (16:30 – 18:00)

1. **SQLite WAL mode**: explicit `PRAGMA journal_mode=WAL` in `database/connection.go` because go-sqlcipher ignores connection-string pragmas.
2. **Never modify a shipped migration** — create a new version.
3. **Watch out for `CREATE UNIQUE INDEX`** with pre-existing duplicates: dedupe first (see migration v9).

## Exercise

1. Add a new migration `000016_add_user_preferences` that creates a `user_preferences` table with columns `user_id INTEGER`, `key TEXT`, `value TEXT`, `UNIQUE(user_id, key)`.
2. Write both Postgres and SQLite implementations.
3. Run `go test ./database/ -run TestRunMigrationsSQLite` — must pass.
4. Verify the `migrations` table grows by one row.

## Assessment questions

1. What rewrite does the wrapper do for `INSERT OR REPLACE`? (Trick: it does nothing — only `INSERT OR IGNORE` is rewritten.)
2. Why can't you call `sql.Result.LastInsertId()` directly in catalog-api code?
3. What happens if you use raw `*sql.DB.Exec` instead of `database.DB.Exec` on PostgreSQL?
