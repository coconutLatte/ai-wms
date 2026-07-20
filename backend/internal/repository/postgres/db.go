// Package postgres implements repository interfaces using PostgreSQL with pgx/v5.
package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ai-wms/ai-wms/backend/pkg/config"
)

// DB holds the PostgreSQL connection pool and provides repository access.
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a new DB with a connection pool to PostgreSQL.
// Pool sizing is driven by the provided config.
func NewDB(ctx context.Context, cfg *config.Config) (*DB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MinConns = cfg.DBMinConns

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connectivity.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("Database connection established",
		slog.Int("max_conns", int(cfg.DBMaxConns)),
		slog.Int("min_conns", int(cfg.DBMinConns)),
	)
	return &DB{Pool: pool}, nil
}

// Close gracefully shuts down the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
	slog.Info("Database connection pool closed")
}

// Ping verifies database connectivity.
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// RunMigrations executes SQL migration files against the database.
// In production, use a proper migration tool (golang-migrate, atlas, etc.).
func (db *DB) RunMigrations(ctx context.Context, migrationsSQL string) error {
	_, err := db.Pool.Exec(ctx, migrationsSQL)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
