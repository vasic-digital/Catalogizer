package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"catalogizer/config"

	_ "github.com/lib/pq"
	_ "github.com/mutecomm/go-sqlcipher"
)

// DB represents the database connection with dialect awareness.
type DB struct {
	*sql.DB
	config  *config.DatabaseConfig
	dialect Dialect
}

// NewConnection creates a new database connection.
// It detects the dialect from config.Type ("postgres" or "sqlite")
// and opens the appropriate driver.
func NewConnection(cfg *config.DatabaseConfig) (*DB, error) {
	dbType := cfg.Type
	if dbType == "" {
		dbType = "sqlite"
	}

	var dialect Dialect
	var sqlDB *sql.DB
	var err error

	switch dbType {
	case "postgres":
		dialect = Dialect{Type: DialectPostgres}
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
		sqlDB, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres database: %w", err)
		}

	default: // "sqlite" or empty
		dialect = Dialect{Type: DialectSQLite}
		connStr := cfg.Path + "?_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1"
		if cfg.EnableWAL {
			connStr += "&_wal_autocheckpoint=1000"
		}
		if cfg.CacheSize != 0 {
			connStr += fmt.Sprintf("&_cache_size=%d", cfg.CacheSize)
		}
		sqlDB, err = sql.Open("sqlite3", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}
	}

	// Configure connection pool
	if cfg.MaxOpenConnections > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		DB:      sqlDB,
		config:  cfg,
		dialect: dialect,
	}

	return db, nil
}

// Dialect returns the database dialect.
func (db *DB) Dialect() *Dialect {
	return &db.dialect
}

// --- Shadowed methods that auto-rewrite queries for dialect ---

// rewriteQuery applies all dialect-specific query transformations:
// placeholder rewriting (? → $1), INSERT OR IGNORE, INSERT OR REPLACE.
func (db *DB) rewriteQuery(query string) string {
	query = db.dialect.RewritePlaceholders(query)
	if db.dialect.IsPostgres() {
		query = db.dialect.RewriteInsertOrIgnore(query)
		query = db.dialect.RewriteInsertOrReplace(query)
		query = db.dialect.RewriteBooleanLiterals(query)
	}
	return query
}

// ExecContext executes a query with dialect rewriting.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, db.rewriteQuery(query), args...)
}

// QueryContext executes a query with dialect rewriting.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, db.rewriteQuery(query), args...)
}

// QueryRowContext executes a query returning a single row with dialect rewriting.
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

// QueryRow executes a query returning a single row with dialect rewriting.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(context.Background(), db.rewriteQuery(query), args...)
}

// --- Dialect-aware helpers ---

// InsertReturningID executes an INSERT and returns the new row's ID.
// For PostgreSQL, it appends "RETURNING id" and uses QueryRow.
// For SQLite, it uses Exec + LastInsertId.
func (db *DB) InsertReturningID(ctx context.Context, query string, args ...interface{}) (int64, error) {
	query = db.rewriteQuery(query)

	if db.dialect.IsPostgres() {
		query += " RETURNING id"
		var id int64
		err := db.DB.QueryRowContext(ctx, query, args...).Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	}

	// SQLite path
	result, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// TableExists checks if a table exists in the database.
func (db *DB) TableExists(ctx context.Context, tableName string) (bool, error) {
	var exists bool
	if db.dialect.IsPostgres() {
		err := db.DB.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)",
			tableName).Scan(&exists)
		return exists, err
	}
	// SQLite
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

// DatabaseType returns the configured database type string.
func (db *DB) DatabaseType() string {
	if db.dialect.IsPostgres() {
		return "postgres"
	}
	return "sqlite"
}

// WrapDB wraps a raw *sql.DB in a *DB with the given dialect.
// This is primarily used in tests where sqlmock or other test
// databases provide a raw *sql.DB.
func WrapDB(sqlDB *sql.DB, dialectType DialectType) *DB {
	if sqlDB == nil {
		return nil
	}
	return &DB{
		DB:      sqlDB,
		config:  &config.DatabaseConfig{},
		dialect: Dialect{Type: dialectType},
	}
}

// createContext creates a context with timeout.
func (db *DB) createContext() (context.Context, context.CancelFunc) {
	timeout := time.Duration(db.config.BusyTimeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}
