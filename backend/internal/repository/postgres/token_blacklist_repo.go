package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// TokenBlacklistRepo implements repository.TokenBlacklistRepository using PostgreSQL.
type TokenBlacklistRepo struct {
	db *DB
}

// NewTokenBlacklistRepo creates a new TokenBlacklistRepo.
func NewTokenBlacklistRepo(db *DB) *TokenBlacklistRepo {
	return &TokenBlacklistRepo{db: db}
}

// Add inserts a JTI into the blacklist.
func (r *TokenBlacklistRepo) Add(ctx context.Context, entry *domain.TokenBlacklistEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	entry.CreatedAt = time.Now()

	const query = `
		INSERT INTO token_blacklist (id, jti, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.exec(ctx, query,
		entry.ID, entry.JTI, entry.UserID, entry.ExpiresAt, entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("add to token blacklist: %w", err)
	}
	return nil
}

// IsBlacklisted checks whether a JTI has been revoked.
func (r *TokenBlacklistRepo) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE jti = $1)`

	var exists bool
	err := r.queryRow(ctx, query, jti).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check token blacklist: %w", err)
	}
	return exists, nil
}

// DeleteExpired removes entries whose expires_at has passed.
func (r *TokenBlacklistRepo) DeleteExpired(ctx context.Context) (int64, error) {
	const query = `DELETE FROM token_blacklist WHERE expires_at < $1`

	tag, err := r.exec(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("delete expired tokens: %w", err)
	}
	return tag, nil
}

// ── Transaction-aware dispatch helpers ─────────────────────

// exec dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *TokenBlacklistRepo) exec(ctx context.Context, sql string, args ...any) (int64, error) {
	if tx := TxFromContext(ctx); tx != nil {
		tag, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return 0, err
		}
		return tag.RowsAffected(), nil
	}
	tag, err := r.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// queryRow dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *TokenBlacklistRepo) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}
	return r.db.Pool.QueryRow(ctx, sql, args...)
}
