// Package database provides connection management and a dual-dialect abstraction layer
// for SQLite and PostgreSQL databases.
//
// It wraps the standard sql.DB with automatic SQL rewriting to support both dialects
// transparently: placeholder conversion (? to $1,$2,...), INSERT OR IGNORE to ON CONFLICT
// DO NOTHING, and boolean literal translation. The package also manages migrations,
// connection pooling, transaction helpers, and the InsertReturningID pattern for
// cross-dialect compatibility.
package database
