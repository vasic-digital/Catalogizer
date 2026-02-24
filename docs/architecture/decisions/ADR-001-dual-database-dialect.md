# ADR-001: Dual Database Dialect (SQLite + PostgreSQL)

## Status
Accepted (2026-02-23)

## Context

Catalogizer needs to support two distinct deployment scenarios with fundamentally different database requirements:

1. **Development and single-user deployments** demand zero-configuration startup. Developers should be able to clone the repo, run `go run main.go`, and have a working database without installing or configuring any external services. Single-user setups (e.g., home media server on a Raspberry Pi) should not require PostgreSQL overhead.

2. **Production and multi-user deployments** need a database that supports concurrent connections, ACID compliance under load, horizontal scaling, point-in-time recovery, and the operational maturity of PostgreSQL.

These two requirements are mutually exclusive if only one database engine is supported. Supporting both requires an abstraction that handles the SQL dialect differences between SQLite and PostgreSQL without leaking implementation details into business logic.

Key dialect differences that must be bridged:
- **Placeholders**: SQLite uses `?`, PostgreSQL uses `$1, $2, ...`
- **Upsert syntax**: SQLite uses `INSERT OR IGNORE`, PostgreSQL uses `ON CONFLICT DO NOTHING`
- **Boolean literals**: SQLite stores booleans as `0/1` integers, PostgreSQL uses `TRUE/FALSE` keywords
- **ID retrieval**: SQLite uses `LastInsertId()`, PostgreSQL requires `RETURNING id`
- **Migration syntax**: Column types and constraints differ between engines

## Decision

We implement a transparent dialect abstraction layer in `catalog-api/database/dialect.go` that intercepts all SQL operations and rewrites them to the target dialect. The abstraction operates at the `database.DB` wrapper level, shadowing the standard `*sql.DB` methods (`Exec`, `Query`, `QueryRow`) with versions that apply dialect-specific transformations before executing.

### Architecture

```
Business Logic (handlers, services, repositories)
    |
    | (writes SQL using SQLite-compatible syntax as the "canonical" form)
    v
database.DB (wrapper around *sql.DB)
    |
    | Exec(), Query(), QueryRow() intercept SQL strings
    v
dialect.go transformations:
    - RewritePlaceholders():    ? -> $1, $2, ...  (PostgreSQL only)
    - RewriteInsertOrIgnore():  INSERT OR IGNORE -> ON CONFLICT DO NOTHING
    - BooleanLiterals():        = 0/1 -> = FALSE/TRUE (for known boolean columns)
    v
*sql.DB (standard Go database driver)
```

### Key Implementation Details

- **`DialectType` enum**: `DialectSQLite` and `DialectPostgres` determine which transformations are applied.
- **`database.NewConnection(cfg)`**: Reads `cfg.Database.Type` (`"sqlite"` or `"postgres"`) and creates the appropriate connection with the correct dialect.
- **`InsertReturningID(ctx, query, args...)`**: Cross-dialect ID retrieval. For SQLite, executes normally and calls `LastInsertId()`. For PostgreSQL, appends `RETURNING id` to the query and uses `QueryRow().Scan()`.
- **`TxInsertReturningID(ctx, tx, query, args...)`**: Same as above but within a transaction.
- **`database.WrapDB(sqlDB, DialectSQLite)`**: Creates a `database.DB` wrapper around an existing `*sql.DB` connection. Used in unit tests to wrap in-memory SQLite databases.
- **Migrations**: Separate migration files for each dialect (`migrations_sqlite.go` and `migrations_postgres.go`) because column types, auto-increment syntax, and constraint definitions differ too much to unify.
- **SQLCipher support**: Imported via `github.com/mutecomm/go-sqlcipher` for optional SQLite encryption in sensitive deployments.

### Configuration

Database type is selected via configuration with environment variable override:

```
config.json: { "database": { "type": "sqlite", "path": "./data/catalogizer.db" } }
env override: DATABASE_TYPE=postgres DATABASE_HOST=localhost DATABASE_PORT=5432 ...
```

Default is SQLite with automatic database file creation. PostgreSQL requires explicit configuration of host, port, name, user, password, and SSL mode.

## Consequences

### Positive

- **Zero-config development**: `go run main.go` works immediately with SQLite, no database setup required.
- **Production-grade option**: PostgreSQL available for deployments needing concurrent access, backups, replication, and scaling.
- **Transparent to business logic**: Handlers, services, and repositories write SQL once using the canonical (SQLite) syntax. Dialect differences are handled automatically by the wrapper.
- **Test isolation**: `database.WrapDB()` with in-memory SQLite provides fast, isolated test databases without any external dependencies.
- **Single codebase**: No code duplication between SQLite and PostgreSQL paths beyond the migration files.

### Negative

- **Dual migration maintenance**: Every schema change must be written twice (SQLite and PostgreSQL variants), increasing migration maintenance burden.
- **SQL subset restriction**: Application code must use only SQL constructs that can be reliably transformed between dialects. Advanced PostgreSQL features (e.g., JSON operators, window functions with PostgreSQL-specific syntax) cannot be used directly in shared code.
- **Rewrite overhead**: Every SQL query passes through the rewrite pipeline at runtime, adding a small (microsecond-level) overhead per query. This is negligible compared to actual query execution time.
- **Boolean column awareness**: The `BooleanLiterals()` rewriter must maintain a list of known boolean column names. New boolean columns require updating this list.
