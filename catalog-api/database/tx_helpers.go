package database

import (
	"context"
	"database/sql"
)

// TxInsertReturningID executes an INSERT inside a transaction and returns the
// new row's ID. It mirrors InsertReturningID but operates on a *sql.Tx.
// For PostgreSQL it appends "RETURNING id" and uses QueryRow; for SQLite it
// uses Exec + LastInsertId.
func (db *DB) TxInsertReturningID(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	query = db.rewriteQuery(query)

	if db.dialect.IsPostgres() {
		query += " RETURNING id"
		var id int64
		err := tx.QueryRowContext(ctx, query, args...).Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	}

	// SQLite path
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
