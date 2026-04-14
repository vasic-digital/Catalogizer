# Comprehensive Decoupling Refactoring Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract all reusable Go functionality from catalog-api into 15 independent modules, each fully tested, documented, and pushed to GitHub + GitLab upstreams.

**Architecture:** Each module is a standalone Go library with zero catalog-api dependencies. Modules define minimal interfaces (Logger, MetricsReporter, etc.) and accept them via constructor injection. catalog-api provides thin adapters mapping its concrete types (zap, gin, Prometheus) to these interfaces. All modules use `replace` directives in catalog-api's `go.mod` for local development.

**Tech Stack:** Go 1.24, standard library, testify for assertions, no framework dependencies in modules

**Design document:** `docs/plans/2026-03-06-decoupling-design.md`

---

## Pre-Flight Checklist

Before starting any module work, verify the baseline:

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2  # all tests pass
```

Keep this passing after every module integration. If it breaks, fix before moving on.

---

## Phase 1: High-Impact Existing Modules (8 modules, 48 tasks)

Each module follows the same 6-task pattern:
1. **Populate** — Write generic code + tests in the module
2. **Test** — `go test ./...` passes standalone in the module
3. **Wire** — Add `replace` directive, update catalog-api imports, create adapter
4. **Verify** — Full catalog-api test suite passes
5. **Document** — CLAUDE.md, AGENTS.md, README.md, docs/ tree
6. **Push** — Commit and push module + catalog-api to upstreams

---

### Module 1: Database (vasic-digital/Database)

**Context:** Extract the dual-dialect SQL abstraction (SQLite + PostgreSQL) from `catalog-api/database/`. The Database module already has 7 packages with ~10K lines. We need to add `pkg/dialect` (query rewriting), `pkg/connection` (connection management), and `pkg/helpers` (transaction helpers), plus extend `pkg/migration`.

**Source files to extract from:**
- `catalog-api/database/dialect.go` (152 lines) — DialectType, Dialect struct, all Rewrite* methods
- `catalog-api/database/connection.go` (232 lines) — DB struct, NewConnection, WrapDB, shadowed methods
- `catalog-api/database/tx_helpers.go` (31 lines) — TxInsertReturningID

**Key decoupling challenge:** `connection.go` imports `catalogizer/config` for `DatabaseConfig`. Must define a generic `ConnectionConfig` in the module.

#### Task 1.1: Populate Database Module

**Files:**
- Create: `Database/pkg/dialect/dialect.go`
- Create: `Database/pkg/dialect/dialect_test.go`
- Create: `Database/pkg/connection/connection.go`
- Create: `Database/pkg/connection/connection_test.go`
- Create: `Database/pkg/helpers/helpers.go`
- Create: `Database/pkg/helpers/helpers_test.go`
- Modify: `Database/go.mod` (add dependencies if needed)

**Step 1: Create `Database/pkg/dialect/dialect.go`**

Extract from `catalog-api/database/dialect.go`. This file has zero external dependencies — pure Go. Copy it verbatim but change the package name.

```go
// Database/pkg/dialect/dialect.go
package dialect

import (
	"fmt"
	"regexp"
	"strings"
)

// Type identifies the SQL dialect in use.
type Type string

const (
	SQLite   Type = "sqlite"
	Postgres Type = "postgres"
)

// Dialect provides helpers for cross-database SQL compatibility.
type Dialect struct {
	Type Type
}

// New creates a new Dialect for the given type.
func New(t Type) *Dialect {
	return &Dialect{Type: t}
}

// RewritePlaceholders converts ? placeholders to $1, $2, ... for PostgreSQL.
func (d *Dialect) RewritePlaceholders(query string) string {
	if d.Type != Postgres {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 32)
	n := 0
	inSingleQuote := false
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch == '\'' {
			inSingleQuote = !inSingleQuote
			b.WriteByte(ch)
			continue
		}
		if ch == '?' && !inSingleQuote {
			n++
			fmt.Fprintf(&b, "$%d", n)
		} else {
			b.WriteByte(ch)
		}
	}
	return b.String()
}

// RewriteInsertOrIgnore converts "INSERT OR IGNORE INTO ..." to
// "INSERT INTO ... ON CONFLICT DO NOTHING" for PostgreSQL.
func (d *Dialect) RewriteInsertOrIgnore(query string) string {
	if d.Type != Postgres {
		return query
	}
	upper := strings.ToUpper(query)
	if idx := strings.Index(upper, "INSERT OR IGNORE INTO"); idx != -1 {
		prefix := query[:idx]
		rest := query[idx+len("INSERT OR IGNORE INTO"):]
		return prefix + "INSERT INTO" + rest + " ON CONFLICT DO NOTHING"
	}
	return query
}

// RewriteInsertOrReplace converts "INSERT OR REPLACE INTO ..." to
// PostgreSQL-compatible syntax.
func (d *Dialect) RewriteInsertOrReplace(query string) string {
	if d.Type != Postgres {
		return query
	}
	upper := strings.ToUpper(query)
	if idx := strings.Index(upper, "INSERT OR REPLACE INTO"); idx != -1 {
		prefix := query[:idx]
		rest := query[idx+len("INSERT OR REPLACE INTO"):]
		return prefix + "INSERT INTO" + rest
	}
	return query
}

// AutoIncrement returns the correct auto-increment primary key clause.
func (d *Dialect) AutoIncrement() string {
	if d.Type == Postgres {
		return "SERIAL PRIMARY KEY"
	}
	return "INTEGER PRIMARY KEY AUTOINCREMENT"
}

// TimestampType returns the column type for timestamps.
func (d *Dialect) TimestampType() string {
	if d.Type == Postgres {
		return "TIMESTAMP"
	}
	return "DATETIME"
}

// BooleanDefault returns the default boolean value syntax.
func (d *Dialect) BooleanDefault(val bool) string {
	if d.Type == Postgres {
		if val {
			return "DEFAULT TRUE"
		}
		return "DEFAULT FALSE"
	}
	if val {
		return "DEFAULT 1"
	}
	return "DEFAULT 0"
}

// CurrentTimestamp returns the current timestamp expression.
func (d *Dialect) CurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// IsSQLite returns true if the dialect is SQLite.
func (d *Dialect) IsSQLite() bool {
	return d.Type == SQLite
}

// IsPostgres returns true if the dialect is PostgreSQL.
func (d *Dialect) IsPostgres() bool {
	return d.Type == Postgres
}

// RewriteBooleanLiterals converts "column = 0" to "column = FALSE" and
// "column = 1" to "column = TRUE" for known boolean columns in PostgreSQL.
func (d *Dialect) RewriteBooleanLiterals(query string, boolColumns []string) string {
	if d.Type != Postgres || len(boolColumns) == 0 {
		return query
	}
	pattern := regexp.MustCompile(
		`(?i)\b(` + strings.Join(boolColumns, "|") + `)\s*=\s*([01])\b`)
	return pattern.ReplaceAllStringFunc(query, func(match string) string {
		if strings.HasSuffix(strings.TrimSpace(match), "1") {
			return pattern.ReplaceAllString(match, "${1} = TRUE")
		}
		return pattern.ReplaceAllString(match, "${1} = FALSE")
	})
}

// RewriteAll applies all dialect-specific query transformations.
func (d *Dialect) RewriteAll(query string, boolColumns []string) string {
	query = d.RewritePlaceholders(query)
	if d.IsPostgres() {
		query = d.RewriteInsertOrIgnore(query)
		query = d.RewriteInsertOrReplace(query)
		query = d.RewriteBooleanLiterals(query, boolColumns)
	}
	return query
}
```

**Step 2: Create `Database/pkg/dialect/dialect_test.go`**

```go
// Database/pkg/dialect/dialect_test.go
package dialect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewritePlaceholders_SQLite(t *testing.T) {
	d := New(SQLite)
	assert.Equal(t, "SELECT * FROM t WHERE id = ?", d.RewritePlaceholders("SELECT * FROM t WHERE id = ?"))
}

func TestRewritePlaceholders_Postgres(t *testing.T) {
	d := New(Postgres)
	assert.Equal(t, "SELECT * FROM t WHERE id = $1 AND name = $2",
		d.RewritePlaceholders("SELECT * FROM t WHERE id = ? AND name = ?"))
}

func TestRewritePlaceholders_QuotedStrings(t *testing.T) {
	d := New(Postgres)
	assert.Equal(t, "SELECT * FROM t WHERE name = 'what?' AND id = $1",
		d.RewritePlaceholders("SELECT * FROM t WHERE name = 'what?' AND id = ?"))
}

func TestRewriteInsertOrIgnore_SQLite(t *testing.T) {
	d := New(SQLite)
	q := "INSERT OR IGNORE INTO t (a) VALUES (?)"
	assert.Equal(t, q, d.RewriteInsertOrIgnore(q))
}

func TestRewriteInsertOrIgnore_Postgres(t *testing.T) {
	d := New(Postgres)
	assert.Equal(t, "INSERT INTO t (a) VALUES (?) ON CONFLICT DO NOTHING",
		d.RewriteInsertOrIgnore("INSERT OR IGNORE INTO t (a) VALUES (?)"))
}

func TestRewriteInsertOrReplace_Postgres(t *testing.T) {
	d := New(Postgres)
	assert.Equal(t, "INSERT INTO t (a) VALUES (?)",
		d.RewriteInsertOrReplace("INSERT OR REPLACE INTO t (a) VALUES (?)"))
}

func TestAutoIncrement(t *testing.T) {
	assert.Equal(t, "INTEGER PRIMARY KEY AUTOINCREMENT", New(SQLite).AutoIncrement())
	assert.Equal(t, "SERIAL PRIMARY KEY", New(Postgres).AutoIncrement())
}

func TestTimestampType(t *testing.T) {
	assert.Equal(t, "DATETIME", New(SQLite).TimestampType())
	assert.Equal(t, "TIMESTAMP", New(Postgres).TimestampType())
}

func TestBooleanDefault(t *testing.T) {
	assert.Equal(t, "DEFAULT 1", New(SQLite).BooleanDefault(true))
	assert.Equal(t, "DEFAULT 0", New(SQLite).BooleanDefault(false))
	assert.Equal(t, "DEFAULT TRUE", New(Postgres).BooleanDefault(true))
	assert.Equal(t, "DEFAULT FALSE", New(Postgres).BooleanDefault(false))
}

func TestRewriteBooleanLiterals(t *testing.T) {
	d := New(Postgres)
	cols := []string{"is_active", "deleted"}
	assert.Equal(t, "SELECT * FROM t WHERE is_active = TRUE",
		d.RewriteBooleanLiterals("SELECT * FROM t WHERE is_active = 1", cols))
	assert.Equal(t, "SELECT * FROM t WHERE deleted = FALSE",
		d.RewriteBooleanLiterals("SELECT * FROM t WHERE deleted = 0", cols))
}

func TestRewriteAll(t *testing.T) {
	d := New(Postgres)
	cols := []string{"is_active"}
	result := d.RewriteAll("INSERT OR IGNORE INTO t (is_active) VALUES (?) WHERE is_active = 1", cols)
	assert.Contains(t, result, "$1")
	assert.Contains(t, result, "ON CONFLICT DO NOTHING")
	assert.Contains(t, result, "is_active = TRUE")
}

func TestIsSQLite_IsPostgres(t *testing.T) {
	assert.True(t, New(SQLite).IsSQLite())
	assert.False(t, New(SQLite).IsPostgres())
	assert.True(t, New(Postgres).IsPostgres())
	assert.False(t, New(Postgres).IsSQLite())
}
```

**Step 3: Create `Database/pkg/connection/connection.go`**

Generic connection wrapper with no framework dependencies. Uses `ConnectionConfig` instead of `config.DatabaseConfig`.

```go
// Database/pkg/connection/connection.go
package connection

import (
	"context"
	"database/sql"
	"time"

	"digital.vasic.database/pkg/dialect"
)

// Config holds database connection parameters.
type Config struct {
	Type               string        // "sqlite" or "postgres"
	DSN                string        // Full connection string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
	BusyTimeout        time.Duration
	BooleanColumns     []string      // Known boolean column names for rewriting
}

// DB wraps *sql.DB with dialect-aware query rewriting.
type DB struct {
	*sql.DB
	dialect        *dialect.Dialect
	booleanColumns []string
	busyTimeout    time.Duration
}

// Open creates a new database connection from config.
// The caller is responsible for registering the appropriate SQL driver
// (e.g., importing _ "github.com/lib/pq" or _ "github.com/mattn/go-sqlite3").
func Open(driverName, dsn string, cfg Config) (*DB, error) {
	sqlDB, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConnections > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	dt := dialect.SQLite
	if cfg.Type == "postgres" {
		dt = dialect.Postgres
	}

	return &DB{
		DB:             sqlDB,
		dialect:        dialect.New(dt),
		booleanColumns: cfg.BooleanColumns,
		busyTimeout:    cfg.BusyTimeout,
	}, nil
}

// Wrap wraps a raw *sql.DB with dialect awareness. Used in tests.
func Wrap(sqlDB *sql.DB, dialectType dialect.Type) *DB {
	if sqlDB == nil {
		return nil
	}
	return &DB{
		DB:      sqlDB,
		dialect: dialect.New(dialectType),
	}
}

// Dialect returns the database dialect.
func (db *DB) Dialect() *dialect.Dialect {
	return db.dialect
}

// rewriteQuery applies all dialect-specific transformations.
func (db *DB) rewriteQuery(query string) string {
	return db.dialect.RewriteAll(query, db.booleanColumns)
}

// ExecContext executes a query with dialect rewriting.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, db.rewriteQuery(query), args...)
}

// QueryContext executes a query with dialect rewriting.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, db.rewriteQuery(query), args...)
}

// QueryRowContext executes a single-row query with dialect rewriting.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, db.rewriteQuery(query), args...)
}

// Exec executes a query with dialect rewriting.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(context.Background(), db.rewriteQuery(query), args...)
}

// Query executes a query with dialect rewriting.
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(context.Background(), db.rewriteQuery(query), args...)
}

// QueryRow executes a single-row query with dialect rewriting.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(context.Background(), db.rewriteQuery(query), args...)
}

// InsertReturningID executes an INSERT and returns the new row's ID.
// PostgreSQL: appends "RETURNING id" and uses QueryRow.
// SQLite: uses Exec + LastInsertId.
func (db *DB) InsertReturningID(ctx context.Context, query string, args ...interface{}) (int64, error) {
	query = db.rewriteQuery(query)
	if db.dialect.IsPostgres() {
		query += " RETURNING id"
		var id int64
		err := db.DB.QueryRowContext(ctx, query, args...).Scan(&id)
		return id, err
	}
	result, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// TxInsertReturningID executes an INSERT inside a transaction and returns the ID.
func (db *DB) TxInsertReturningID(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	query = db.rewriteQuery(query)
	if db.dialect.IsPostgres() {
		query += " RETURNING id"
		var id int64
		err := tx.QueryRowContext(ctx, query, args...).Scan(&id)
		return id, err
	}
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// TableExists checks if a table exists in the database.
func (db *DB) TableExists(ctx context.Context, tableName string) (bool, error) {
	if db.dialect.IsPostgres() {
		var exists bool
		err := db.DB.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)",
			tableName).Scan(&exists)
		return exists, err
	}
	var count int
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
		tableName).Scan(&count)
	return count > 0, err
}

// HealthCheck performs a database health check.
func (db *DB) HealthCheck() error {
	ctx, cancel := db.createContext()
	defer cancel()
	return db.PingContext(ctx)
}

// GetStats returns database connection statistics.
func (db *DB) GetStats() sql.DBStats {
	return db.Stats()
}

// DatabaseType returns "postgres" or "sqlite".
func (db *DB) DatabaseType() string {
	if db.dialect.IsPostgres() {
		return "postgres"
	}
	return "sqlite"
}

func (db *DB) createContext() (context.Context, context.CancelFunc) {
	timeout := db.busyTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}
```

**Step 4: Create `Database/pkg/connection/connection_test.go`**

```go
// Database/pkg/connection/connection_test.go
package connection

import (
	"context"
	"database/sql"
	"testing"

	"digital.vasic.database/pkg/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap_NilDB(t *testing.T) {
	assert.Nil(t, Wrap(nil, dialect.SQLite))
}

func TestWrap_SQLite(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := Wrap(sqlDB, dialect.SQLite)
	require.NotNil(t, db)
	assert.True(t, db.Dialect().IsSQLite())
	assert.Equal(t, "sqlite", db.DatabaseType())
}

func TestWrap_Postgres(t *testing.T) {
	// Just test the wrapping, not an actual postgres connection
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := Wrap(sqlDB, dialect.Postgres)
	require.NotNil(t, db)
	assert.True(t, db.Dialect().IsPostgres())
	assert.Equal(t, "postgres", db.DatabaseType())
}

func TestDB_ExecAndQuery_SQLite(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := Wrap(sqlDB, dialect.SQLite)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO test (name) VALUES (?)", "hello")
	require.NoError(t, err)

	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM test WHERE id = ?", 1).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "hello", name)
}

func TestDB_InsertReturningID_SQLite(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := Wrap(sqlDB, dialect.SQLite)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	require.NoError(t, err)

	id, err := db.InsertReturningID(ctx, "INSERT INTO test (name) VALUES (?)", "hello")
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	id, err = db.InsertReturningID(ctx, "INSERT INTO test (name) VALUES (?)", "world")
	require.NoError(t, err)
	assert.Equal(t, int64(2), id)
}

func TestDB_TableExists_SQLite(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := Wrap(sqlDB, dialect.SQLite)
	ctx := context.Background()

	exists, err := db.TableExists(ctx, "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = db.ExecContext(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	exists, err = db.TableExists(ctx, "test")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestDB_HealthCheck(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := Wrap(sqlDB, dialect.SQLite)
	assert.NoError(t, db.HealthCheck())
}

func TestDB_GetStats(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := Wrap(sqlDB, dialect.SQLite)
	stats := db.GetStats()
	assert.Equal(t, 0, stats.InUse)
}
```

**Step 5: Create `Database/pkg/helpers/helpers.go`**

```go
// Database/pkg/helpers/helpers.go
package helpers

import (
	"context"
	"database/sql"
	"fmt"
)

// TxFunc is a function that executes within a transaction.
type TxFunc func(tx *sql.Tx) error

// WithTransaction executes fn within a database transaction.
// If fn returns an error, the transaction is rolled back.
// Otherwise, it is committed.
func WithTransaction(ctx context.Context, db *sql.DB, fn TxFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	return tx.Commit()
}
```

**Step 6: Update `Database/go.mod`**

Ensure go.mod has the SQLite driver for tests:

```
module digital.vasic.database

go 1.24.0

require (
    github.com/stretchr/testify v1.11.1
    github.com/mutecomm/go-sqlcipher v0.0.0-20190227152316-55dbde17881f
)
```

Run: `cd Database && go mod tidy`

#### Task 1.2: Test Database Module

Run: `cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Database && go test ./... -v`

Expected: All tests pass. Fix any compilation errors.

#### Task 1.3: Wire Database into catalog-api

**Files:**
- Modify: `catalog-api/go.mod` (add replace directive)
- Modify: `catalog-api/database/dialect.go` (thin wrapper delegating to module)
- Modify: `catalog-api/database/connection.go` (use module's DB as base)

**Step 1: Add replace directive**

Add to `catalog-api/go.mod` after existing replace directives:

```
replace digital.vasic.database => ../Database
```

And add to require block:

```
digital.vasic.database v0.0.0-00010101000000-000000000000
```

Run: `cd catalog-api && go mod tidy`

**Step 2: Update catalog-api to import from module**

The catalog-api `database/` package should delegate to the module where possible while maintaining its existing API surface. The key changes:

- `database/dialect.go`: Import `digital.vasic.database/pkg/dialect` and create type aliases or thin wrappers
- `database/connection.go`: Use `digital.vasic.database/pkg/connection` internally but keep `config.DatabaseConfig` for the Catalogizer-specific constructor

This is a thin adapter pattern — catalog-api's `database` package becomes a facade over the module.

**Step 3: Run `go mod tidy`**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api && go mod tidy
```

#### Task 1.4: Verify catalog-api Tests

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2
```

Expected: All tests pass. If any fail, the adapter layer needs adjustment.

#### Task 1.5: Document Database Module

**Files to create/update:**
- `Database/CLAUDE.md` — Update with new packages
- `Database/AGENTS.md` — Update with new packages
- `Database/README.md` — Update with usage examples
- `Database/docs/architecture.md` — Architecture diagram
- `Database/docs/api-reference.md` — Update API docs
- `Database/docs/design-patterns.md` — Abstract Factory (dialect), Proxy (rewriting wrapper), Strategy (SQLite vs PostgreSQL)
- `Database/docs/sql-definitions.md` — SQL compatibility reference
- `Database/docs/website/index.md`
- `Database/docs/website/getting-started.md`
- `Database/docs/website/examples.md`
- `Database/docs/website/faq.md`
- `Database/docs/courses/outline.md`
- `Database/docs/courses/lesson-01.md`

#### Task 1.6: Push Database Module

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Database
git add -A && git commit -m "feat: add dialect, connection, and helpers packages extracted from catalog-api"
commit "feat: add dialect, connection, and helpers packages"

cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add Database catalog-api/go.mod catalog-api/go.sum catalog-api/database/
git commit -m "feat(database): wire Database module into catalog-api"
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

---

### Module 2: Concurrency (vasic-digital/Concurrency)

**Context:** Already wired via replace directive. Extend with `pkg/retry` and `pkg/bulkhead` extracted from `catalog-api/internal/recovery/retry.go`.

**Source files:**
- `catalog-api/internal/recovery/retry.go` lines 1-280 — RetryConfig, Retry, ExponentialBackoff, Bulkhead, HealthChecker

**Key decoupling:** Remove `go.uber.org/zap` dependency. Define `Logger` interface in module.

#### Task 2.1: Populate Concurrency Module

**Files:**
- Create: `Concurrency/pkg/retry/retry.go`
- Create: `Concurrency/pkg/retry/retry_test.go`
- Create: `Concurrency/pkg/bulkhead/bulkhead.go`
- Create: `Concurrency/pkg/bulkhead/bulkhead_test.go`

**`Concurrency/pkg/retry/retry.go`** — Generic retry with exponential backoff. Replace `*zap.Logger` with:

```go
// Logger is a minimal logging interface for retry operations.
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}
```

Extract all retry logic from `catalog-api/internal/recovery/retry.go` lines 12-207. Replace `zap.Logger` with `Logger` interface. Replace `zap.Int(...)`, `zap.Error(...)` etc. with simple key-value pairs.

**`Concurrency/pkg/bulkhead/bulkhead.go`** — Extract from lines 209-280 of the same file.

#### Task 2.2: Test Concurrency Module

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Concurrency && go test ./... -v
```

#### Task 2.3: Wire Concurrency (already wired — just update imports)

catalog-api's `internal/recovery/retry.go` should import from `digital.vasic.concurrency/pkg/retry` and adapt the Logger interface.

#### Task 2.4: Verify

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2
```

#### Task 2.5: Document Concurrency Module

Same pattern as Database: update CLAUDE.md, AGENTS.md, README.md, create docs/design-patterns.md (Strategy for backoff, State for circuit breaker, Template Method for retry callbacks).

#### Task 2.6: Push

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Concurrency && commit "feat: add retry and bulkhead packages"
```

---

### Module 3: Observability (vasic-digital/Observability)

**Context:** Extend with `pkg/middleware` (HTTP metrics middleware) extracted from `catalog-api/internal/metrics/`.

**Source files:**
- `catalog-api/internal/metrics/metrics.go` (163 lines) — Prometheus metrics, runtime collector
- `catalog-api/internal/metrics/prometheus.go` — Prometheus handler
- `catalog-api/internal/metrics/health.go` — Health endpoint
- `catalog-api/internal/metrics/middleware.go` — HTTP metrics middleware

**Key decoupling:** Replace `prometheus/client_golang` with generic interfaces. The module should define:

```go
// MetricsReporter records HTTP request metrics.
type MetricsReporter interface {
    ObserveHTTPDuration(method, path, status string, seconds float64)
    IncrHTTPTotal(method, path, status string)
    SetActiveConnections(count float64)
}
```

Prometheus implementation lives in the module's `pkg/metrics/` (already exists). Add `pkg/middleware/` with generic HTTP middleware that accepts `MetricsReporter`.

#### Tasks 3.1–3.6: Same pattern as above

Populate → Test → Wire → Verify → Document → Push

---

### Module 4: Security (vasic-digital/Security)

**Context:** Extract security header middleware and scanning utilities from `catalog-api/internal/middleware/` and `catalog-api/tests/security/`.

**Source files:**
- `catalog-api/internal/middleware/middleware.go` — SecurityHeaders middleware (Gin-specific)
- `catalog-api/tests/security/` — Security test utilities

**Key decoupling:** Convert Gin middleware to `net/http` middleware in module. Add `pkg/headers/` for security header configuration.

```go
// Security/pkg/headers/headers.go
package headers

import "net/http"

// Config defines which security headers to set.
type Config struct {
    ContentSecurityPolicy   string
    XFrameOptions           string
    XContentTypeOptions     string
    StrictTransportSecurity string
    ReferrerPolicy          string
    PermissionsPolicy       string
}

// Middleware returns an http.Handler that sets security headers.
func Middleware(cfg Config) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if cfg.ContentSecurityPolicy != "" {
                w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
            }
            if cfg.XFrameOptions != "" {
                w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
            }
            // ... all other headers
            next.ServeHTTP(w, r)
        })
    }
}
```

#### Tasks 4.1–4.6: Same pattern

---

### Module 5: Middleware (vasic-digital/Middleware)

**Context:** The biggest extraction. Pull auth, caching, rate limiting, validation, compression from `catalog-api/middleware/` and `catalog-api/internal/middleware/`.

**Source files:**
- `catalog-api/middleware/auth.go` — JWT auth middleware
- `catalog-api/middleware/cache_headers.go` — Cache-Control headers
- `catalog-api/middleware/advanced_rate_limiter.go` — Token bucket rate limiter
- `catalog-api/middleware/redis_rate_limiter.go` — Redis-backed rate limiter
- `catalog-api/middleware/input_validation.go` — Input sanitization
- `catalog-api/middleware/request.go` — Request context helpers
- `catalog-api/internal/middleware/compression.go` — Brotli/gzip compression

**New packages to create in module:**
- `pkg/auth/` — Generic JWT authentication middleware (net/http)
- `pkg/cache/` — Cache-Control header middleware
- `pkg/ratelimit/` — Token bucket + sliding window rate limiting
- `pkg/validation/` — Input validation/sanitization
- `pkg/compression/` — Brotli/gzip compression middleware
- `pkg/security/` — Security headers (or delegate to Security module)

**Key decoupling:** All middleware uses `net/http` interfaces. Gin adapter stays in catalog-api.

```go
// Middleware/pkg/auth/auth.go
package auth

import "net/http"

// TokenValidator validates JWT tokens.
type TokenValidator interface {
    ValidateToken(tokenString string) (claims map[string]interface{}, err error)
}

// Middleware returns HTTP middleware that validates JWT tokens.
func Middleware(validator TokenValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            // Validate using TokenValidator interface
            // Set claims in request context
            next.ServeHTTP(w, r)
        })
    }
}
```

#### Tasks 5.1–5.6: Same pattern

---

### Module 6: Media (vasic-digital/Media)

**Context:** Extract media detection pipeline from `catalog-api/internal/media/`.

**Source files:**
- `catalog-api/internal/media/detector/` — Media type detection
- `catalog-api/internal/media/analyzer/` — Media file analysis
- `catalog-api/internal/media/models/` — Media data models
- `catalog-api/internal/media/providers/` — External metadata providers (TMDB, IMDB)
- `catalog-api/internal/media/manager.go` — Media manager orchestrator

**New packages:**
- Extend `pkg/detector/` — File extension/magic byte detection
- Extend `pkg/analyzer/` — Media file metadata extraction
- Add `pkg/models/` — Generic media entity models
- Extend `pkg/provider/` — Provider interface + registry
- Add `pkg/manager/` — Pipeline orchestrator

**Key decoupling:** Replace `catalogizer/models` with generic interfaces. Each provider implements a standard `MetadataProvider` interface.

```go
// Media/pkg/provider/provider.go
package provider

// MediaInfo represents metadata from an external provider.
type MediaInfo struct {
    Title       string
    Year        int
    Description string
    PosterURL   string
    Rating      float64
    Genres      []string
    Source      string // "tmdb", "imdb", etc.
    ExternalID  string
}

// Provider fetches metadata for media.
type Provider interface {
    Name() string
    Search(query string, mediaType string) ([]MediaInfo, error)
    GetByID(externalID string) (*MediaInfo, error)
}

// Registry manages available providers.
type Registry struct {
    providers map[string]Provider
}
```

#### Tasks 6.1–6.6: Same pattern

---

### Module 7: Discovery (vasic-digital/Discovery)

**Context:** Extract resilient SMB connection management from `catalog-api/internal/smb/`.

**Source files:**
- `catalog-api/internal/smb/resilience.go` (801 lines) — ResilientSMBManager, OfflineCache, HealthChecker, connection state machine

**Key decoupling:** Remove `catalogizer/internal/metrics` dependency. Define generic interfaces:

```go
// Discovery/pkg/resilience/resilience.go

// MetricsReporter reports connection health metrics.
type MetricsReporter interface {
    SetSourceHealth(sourceID string, value float64)
}

// Logger for discovery operations.
type Logger interface {
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
    Debug(msg string, keysAndValues ...interface{})
}
```

Extend `pkg/smb/` with resilience patterns. Add `pkg/resilience/` for the generic connection state machine.

#### Tasks 7.1–7.6: Same pattern

---

### Module 8: Streaming (vasic-digital/Streaming)

**Context:** Extract realtime event streaming from `catalog-api/internal/media/realtime/`.

**Source files:**
- `catalog-api/internal/media/realtime/watcher.go` — File system watcher with debouncing
- `catalog-api/internal/media/realtime/enhanced_watcher.go` — Enhanced watcher with event aggregation

**New package:** Add `pkg/realtime/` — Generic file change notification with debouncing and event aggregation.

The module already has `pkg/websocket/`, `pkg/sse/`, `pkg/transport/`. The realtime package connects file system events to these transport mechanisms.

#### Tasks 8.1–8.6: Same pattern

---

## Phase 2: New Standalone Modules (3 modules, 18 tasks)

### Pre-requisite: Create Repositories

Before populating these modules, create the GitHub and GitLab repos:

```bash
# GitHub
gh repo create vasic-digital/Lazy --public --description "Generic reusable Go module: digital.vasic.lazy"
gh repo create vasic-digital/Memory --public --description "Generic reusable Go module: digital.vasic.memory"
gh repo create vasic-digital/Recovery --public --description "Generic reusable Go module: digital.vasic.recovery"

# GitLab
glab project create --group vasic-digital --name lazy --description "digital.vasic.lazy - Reusable Go module"
glab project create --group vasic-digital --name memory --description "digital.vasic.memory - Reusable Go module"
glab project create --group vasic-digital --name recovery --description "digital.vasic.recovery - Reusable Go module"
```

Then create module directories with the standard structure:

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
for mod in Lazy Memory Recovery; do
    mkdir -p "$mod"/{pkg,docs/{website,courses},Upstreams}
    # Initialize go.mod
    cd "$mod"
    go mod init "digital.vasic.$(echo $mod | tr '[:upper:]' '[:lower:]')"
    cd ..
done
```

Create Upstreams scripts for each:

```bash
# Lazy/Upstreams/GitHub.sh
#!/bin/bash
export UPSTREAMABLE_REPOSITORY="git@github.com:vasic-digital/Lazy.git"

# Lazy/Upstreams/GitLab.sh
#!/bin/bash
export UPSTREAMABLE_REPOSITORY="git@gitlab.com:vasic-digital/lazy.git"
```

(Same pattern for Memory and Recovery)

Then: `cd Lazy && install_upstreams` (repeat for Memory, Recovery)

Add as git submodules:

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git submodule add git@github.com:vasic-digital/Lazy.git Lazy
git submodule add git@github.com:vasic-digital/Memory.git Memory
git submodule add git@github.com:vasic-digital/Recovery.git Recovery
```

---

### Module 9: Lazy (vasic-digital/Lazy)

**Source:** `catalog-api/pkg/lazy/lazy.go` (62 lines)

This is a simple generic lazy-loading library using `sync.Once`. Zero external dependencies.

#### Task 9.1: Populate Lazy Module

**Files:**
- Create: `Lazy/pkg/lazy/lazy.go`
- Create: `Lazy/pkg/lazy/lazy_test.go`

Copy `catalog-api/pkg/lazy/lazy.go` verbatim — it already has no external dependencies. Tests from `catalog-api/pkg/lazy/lazy_test.go`.

#### Task 9.2: Test

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/Lazy && go test ./... -v
```

#### Task 9.3: Wire

Add to `catalog-api/go.mod`:
```
replace digital.vasic.lazy => ../Lazy
```

Update all `catalog-api/pkg/lazy` imports to `digital.vasic.lazy/pkg/lazy`.

#### Task 9.4: Verify

Full test suite.

#### Task 9.5: Document

Standard documentation tree + design patterns (Proxy for lazy value, Singleton via sync.Once).

#### Task 9.6: Push

Commit and push Lazy module to both upstreams. Push catalog-api to 6 remotes.

---

### Module 10: Memory (vasic-digital/Memory)

**Source:** `catalog-api/pkg/memory/leak_detector.go` (283 lines)

Memory leak detector using runtime.MemStats. Zero external dependencies.

#### Task 10.1: Populate Memory Module

**Files:**
- Create: `Memory/pkg/memory/leak_detector.go`
- Create: `Memory/pkg/memory/leak_detector_test.go`

Copy from `catalog-api/pkg/memory/`. Already framework-independent.

Add `AlertCallback` observer pattern:

```go
// Memory/pkg/memory/leak_detector.go
// (add to existing MemoryMonitor)

// AlertCallback is called when a potential leak is detected.
type AlertCallback func(LeakReport)
```

#### Tasks 10.2–10.6: Same pattern

---

### Module 11: Recovery (vasic-digital/Recovery)

**Source:** `catalog-api/internal/recovery/` (retry.go + circuit_breaker.go = ~660 lines)

**Key insight:** `circuit_breaker.go` already imports `digital.vasic.concurrency/pkg/breaker`. The Recovery module wraps Concurrency primitives with application-level concerns (named breakers, state change callbacks, health checking).

#### Task 11.1: Populate Recovery Module

**Files:**
- Create: `Recovery/pkg/retry/retry.go` — Same as Concurrency/pkg/retry but with health checking integration
- Create: `Recovery/pkg/breaker/breaker.go` — Named circuit breaker manager with callbacks
- Create: `Recovery/pkg/health/health.go` — Health checker (from retry.go lines 282-404)
- Create: `Recovery/pkg/facade/facade.go` — Unified resilience API

**`Recovery/go.mod`** needs:
```
require digital.vasic.concurrency v0.0.0-00010101000000-000000000000
replace digital.vasic.concurrency => ../Concurrency
```

The Recovery module depends on Concurrency (for the base breaker implementation).

#### Tasks 11.2–11.6: Same pattern

---

## Phase 3: Low-Priority Modules (4 modules, 16 tasks)

### Module 12: Storage (vasic-digital/Storage)

Already has `pkg/local/`, `pkg/object/`, `pkg/provider/`, `pkg/s3/`.

**Task:** Add `pkg/resolver/` — asset resolver that maps logical asset paths to storage backends.

#### Tasks 12.1–12.4: Populate → Test → Wire → Verify

---

### Module 13: Cache (vasic-digital/Cache)

Already wired and has `pkg/cache/`, `pkg/distributed/`, `pkg/memory/`, `pkg/policy/`, `pkg/redis/`.

**Task:** Add `pkg/service/` — service integration wrapper that combines cache policies with service-layer caching patterns.

#### Tasks 13.1–13.4: Populate → Test → Wire → Verify

---

### Module 14: Watcher (vasic-digital/Watcher)

Already has `pkg/debounce/`, `pkg/filter/`, `pkg/handler/`, `pkg/watcher/`.

**Task:** Wire into catalog-api's file watching infrastructure. Replace `fsnotify` direct usage with Watcher module.

#### Tasks 14.1–14.4: Populate → Test → Wire → Verify

---

### Module 15: RateLimiter (vasic-digital/RateLimiter)

Already has `pkg/limiter/`, `pkg/memory/`, `pkg/middleware/`, `pkg/redis/`, `pkg/sliding/`.

**Task:** Wire into catalog-api's rate limiting middleware. Replace `catalog-api/middleware/advanced_rate_limiter.go` with RateLimiter module.

#### Tasks 15.1–15.4: Populate → Test → Wire → Verify

---

## Phase 4: Catalog-API Cleanup (6 tasks)

### Task P4.1: Delete Replaced Code

**Files to delete:**
- `catalog-api/pkg/lazy/` — Replaced by `digital.vasic.lazy`
- `catalog-api/pkg/semaphore/` — Duplicate of `digital.vasic.concurrency/pkg/semaphore`
- `catalog-api/pkg/memory/` — Replaced by `digital.vasic.memory`

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
rm -rf pkg/lazy/ pkg/semaphore/ pkg/memory/
```

### Task P4.2: Replace Remaining Inline Code

Scan catalog-api for any remaining code that should delegate to modules:

```bash
cd catalog-api
grep -r "internal/recovery" --include="*.go" -l
grep -r "internal/metrics" --include="*.go" -l
grep -r "internal/smb" --include="*.go" -l
```

Update all imports to use module packages.

### Task P4.3: Update All Internal Imports

Run `go mod tidy` and verify no compilation errors:

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
go mod tidy
go build ./...
```

### Task P4.4: Fix Upstreams for Containers + Entities

Check that `Containers/` and `Entities/` have proper Upstreams:

```bash
ls Containers/Upstreams/
ls Entities/Upstreams/
```

If missing, create `GitHub.sh` and `GitLab.sh` and run `install_upstreams`.

### Task P4.5: Full Test Suite Verification

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2
```

All tests must pass with zero failures.

### Task P4.6: Commit and Push

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add -A
git commit -m "refactor: catalog-api cleanup — delete replaced packages, update all imports"
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

---

## Phase 5: Documentation & Polish (5 tasks)

### Task P5.1: Architecture Documentation

For ALL 15 modules, ensure `docs/architecture.md` exists with:
- Module purpose and scope
- Package dependency diagram (text-based)
- Design patterns applied with explanations
- Interface contracts

### Task P5.2: Website Content

For ALL 15 modules, create `docs/website/`:
- `index.md` — Landing page content
- `getting-started.md` — Quick start guide
- `examples.md` — Code examples
- `faq.md` — Common questions

### Task P5.3: Course Outlines

For ALL 15 modules, create `docs/courses/`:
- `outline.md` — Video course structure
- `lesson-01.md` through `lesson-N.md` — Lesson plans

### Task P5.4: SQL Definitions

For modules with database operations (Database, Cache, Storage):
- `docs/sql-definitions.md` — Table schemas, migration patterns

### Task P5.5: Challenge Adapters

In catalog-api, create thin challenge adapters that exercise each module:

**File:** `catalog-api/challenges/module_challenges.go`

```go
// Register challenges that verify module functionality
func RegisterModuleChallenges(registry interface{ Register(challenge.Challenge) error }) {
    // One challenge per module verifying basic functionality
    // E.g., "Database dialect rewriting works"
    // E.g., "Concurrency retry succeeds after transient failure"
    // E.g., "Memory leak detector reports heap stats"
}
```

---

## go.mod Final State

After all phases complete, `catalog-api/go.mod` should have these replace directives:

```go
// Existing (10)
replace digital.vasic.challenges => ../Challenges
replace digital.vasic.assets => ../Assets
replace digital.vasic.containers => ../Containers
replace digital.vasic.concurrency => ../Concurrency
replace digital.vasic.config => ../Config
replace digital.vasic.filesystem => ../Filesystem
replace digital.vasic.auth => ../Auth
replace digital.vasic.cache => ../Cache
replace digital.vasic.entities => ../Entities
replace digital.vasic.eventbus => ../EventBus

// New (13)
replace digital.vasic.database => ../Database
replace digital.vasic.middleware => ../Middleware
replace digital.vasic.media => ../Media
replace digital.vasic.discovery => ../Discovery
replace digital.vasic.observability => ../Observability
replace digital.vasic.streaming => ../Streaming
replace digital.vasic.security => ../Security
replace digital.vasic.ratelimiter => ../RateLimiter
replace digital.vasic.watcher => ../Watcher
replace digital.vasic.storage => ../Storage
replace digital.vasic.lazy => ../Lazy
replace digital.vasic.memory => ../Memory
replace digital.vasic.recovery => ../Recovery
```

---

## Safety Rules

1. **Never modify a module and catalog-api in the same step.** Populate/test the module first, then wire it.
2. **Run `go test ./...` after every wire step.** If tests fail, fix before proceeding.
3. **Commit module and catalog-api changes together** so git bisect works.
4. **Resource limits:** Always use `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
5. **One module at a time.** Complete all 6 tasks for module N before starting module N+1.

---

## Progress Tracking

Use TaskCreate/TaskList to track each of the 93 tasks. The naming convention:

- `P1-M1-T1` = Phase 1, Module 1 (Database), Task 1 (Populate)
- `P1-M1-T2` = Phase 1, Module 1 (Database), Task 2 (Test)
- ...
- `P4-T1` = Phase 4, Task 1 (Delete replaced code)
- `P5-T1` = Phase 5, Task 1 (Architecture docs)

Work is resumable at any task boundary. If interrupted, use TaskList to find the last completed task and resume from the next one.

---

## Execution Order Summary

| Order | Module | Key Action |
|-------|--------|------------|
| 1 | Database | Add dialect, connection, helpers packages |
| 2 | Concurrency | Add retry, bulkhead packages |
| 3 | Observability | Add HTTP metrics middleware |
| 4 | Security | Add headers, scanning packages |
| 5 | Middleware | Add auth, cache, ratelimit, validation, compression |
| 6 | Media | Add models, manager, extend detector/analyzer/provider |
| 7 | Discovery | Add resilience package, extend SMB |
| 8 | Streaming | Add realtime package |
| 9 | Lazy | New module — lazy loading |
| 10 | Memory | New module — leak detection |
| 11 | Recovery | New module — circuit breaker + health + facade |
| 12 | Storage | Add resolver package |
| 13 | Cache | Add service wrapper |
| 14 | Watcher | Wire into catalog-api |
| 15 | RateLimiter | Wire into catalog-api |
| 16 | Cleanup | Delete replaced code, fix imports |
| 17 | Docs | Architecture, website, courses, SQL, challenges |
