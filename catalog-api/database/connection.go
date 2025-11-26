package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"catalogizer/config"
	_ "github.com/mutecomm/go-sqlcipher"
)

// DB represents the database connection
type DB struct {
	*sql.DB
	config *config.DatabaseConfig
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.DatabaseConfig) (*DB, error) {
	// Build connection string with parameters
	connStr := cfg.Path + "?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1"

	if cfg.EnableWAL {
		connStr += "&_wal_autocheckpoint=1000"
	}

	if cfg.CacheSize != 0 {
		connStr += fmt.Sprintf("&_cache_size=%d", cfg.CacheSize)
	}

	// Open database connection
	sqlDB, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		DB:     sqlDB,
		config: cfg,
	}

	return db, nil
}

// HealthCheck performs a database health check
func (db *DB) HealthCheck() error {
	ctx, cancel := db.createContext()
	defer cancel()

	return db.PingContext(ctx)
}

// GetStats returns database connection statistics
func (db *DB) GetStats() sql.DBStats {
	return db.Stats()
}

// createContext creates a context with timeout
func (db *DB) createContext() (context.Context, context.CancelFunc) {
	timeout := time.Duration(db.config.BusyTimeout) * time.Millisecond
	return context.WithTimeout(context.Background(), timeout)
}
