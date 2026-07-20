// Package repository defines data access interfaces for the WMS domain.
package repository

import "context"

// TxManager provides atomic transaction boundaries for multi-step operations.
// Services use TxManager to ensure that multiple repository calls within a
// single business operation succeed or fail atomically.
type TxManager interface {
	// WithTx begins a new transaction, calls fn with the transactional context,
	// and commits on success. If fn returns an error, the transaction is rolled
	// back automatically.
	//
	// The ctx passed to fn carries the active transaction; repository
	// implementations that are transaction-aware will automatically use it
	// instead of the connection pool.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
