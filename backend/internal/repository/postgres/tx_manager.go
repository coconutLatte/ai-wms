package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// txKey is an unexported context key type used to store the active pgx transaction.
type txKey struct{}

// TxFromContext extracts the active pgx.Tx from the context, or nil if no
// transaction is active. Repository methods use this to participate in
// transactions transparently.
func TxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey{}).(pgx.Tx)
	return tx
}

// contextWithTx embeds a pgx.Tx into the context for downstream repository calls.
func contextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxManager implements repository.TxManager using a PostgreSQL connection pool.
type TxManager struct {
	db *DB
}

// NewTxManager creates a new TxManager backed by the given DB connection pool.
func NewTxManager(db *DB) *TxManager {
	return &TxManager{db: db}
}

// WithTx begins a PostgreSQL transaction, injects it into the context, and
// calls fn. If fn returns an error, the transaction is rolled back; otherwise
// it is committed.
func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}

	// Ensure the transaction is always cleaned up.
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw after rollback
		}
	}()

	txCtx := contextWithTx(ctx, tx)

	if err = fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}

	return nil
}

// Compile-time interface check.
var _ repository.TxManager = (*TxManager)(nil)
