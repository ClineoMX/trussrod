package database

import (
	"context"
)

// Rows is an interface for iterating over query result rows.
type Rows interface {
	// Next prepares the next result row for reading with Scan.
	// It returns true on success, or false if there is no next result row or an error occurred.
	Next() bool
	// Scan copies the columns from the current row into dest.
	Scan(dest ...any) error
	// Close closes the rows, preventing further enumeration.
	Close()
	// Err returns any error that occurred during iteration.
	Err() error
}

// Tx is an interface for a database transaction.
type Tx interface {
	// Commit commits the transaction.
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error
	// Query executes a query that returns rows.
	Query(ctx context.Context, sql string, args ...any) (Rows, error)

	// QueryRow executes a query that returns a single row.
	QueryRow(ctx context.Context, sql string, args ...any) Row

	// Exec executes a query that doesn't return rows.
	Exec(ctx context.Context, sql string, args ...any) (Result, error)
}

// Row is an interface for scanning a single query result row.
type Row interface {
	// Scan copies the columns from the row into dest.
	Scan(dest ...any) error
}

// Result represents the result of an Exec command.
type Result interface {
	// RowsAffected returns the number of rows affected.
	RowsAffected() int64
}

// DB is an interface that abstracts database operations.
// This allows for easier testing and potential database driver swaps.
type DB interface {
	// Query executes a query that returns rows.
	Query(ctx context.Context, sql string, args ...any) (Rows, error)

	// QueryRow executes a query that returns a single row.
	QueryRow(ctx context.Context, sql string, args ...any) Row

	// Exec executes a query that doesn't return rows.
	Exec(ctx context.Context, sql string, args ...any) (Result, error)

	// BeginTx begins a new transaction.
	BeginTx(ctx context.Context, opts any) (Tx, error)

	// Ping checks if the database is reachable.
	Ping(ctx context.Context) error
}
