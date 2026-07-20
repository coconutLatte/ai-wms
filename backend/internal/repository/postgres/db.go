// Package postgres implements repository interfaces using PostgreSQL with pgx/v5.
package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB holds the PostgreSQL connection pool and provides repository access.
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a new DB with a connection pool to PostgreSQL.
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Printf("[postgres] Connected (max_conns=%d, min_conns=%d)", config.MaxConns, config.MinConns)
	return &DB{Pool: pool}, nil
}

// Close gracefully shuts down the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
	log.Println("[postgres] Connection pool closed")
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
