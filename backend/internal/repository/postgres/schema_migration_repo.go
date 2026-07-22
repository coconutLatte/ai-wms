package postgres

import (
	"context"
	"fmt"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// SchemaMigrationRepo implements repository.MigrationRepository using PostgreSQL.
type SchemaMigrationRepo struct {
	db *DB
}

// NewSchemaMigrationRepo creates a new SchemaMigrationRepo.
func NewSchemaMigrationRepo(db *DB) *SchemaMigrationRepo {
	return &SchemaMigrationRepo{db: db}
}

// GetApplied returns all migrations that have been applied, ordered by version.
func (r *SchemaMigrationRepo) GetApplied(ctx context.Context) ([]*domain.SchemaMigration, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, version, filename, COALESCE(checksum, ''), applied_at
		 FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []*domain.SchemaMigration
	for rows.Next() {
		m := &domain.SchemaMigration{}
		if err := rows.Scan(&m.ID, &m.Version, &m.Filename, &m.Checksum, &m.AppliedAt); err != nil {
			return nil, fmt.Errorf("scan migration row: %w", err)
		}
		migrations = append(migrations, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate migration rows: %w", err)
	}
	return migrations, nil
}

// IsApplied checks whether a specific migration version has been applied.
func (r *SchemaMigrationRepo) IsApplied(ctx context.Context, version string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check migration applied: %w", err)
	}
	return exists, nil
}

// RecordApplied records that a migration has been successfully applied.
func (r *SchemaMigrationRepo) RecordApplied(ctx context.Context, m *domain.SchemaMigration) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO schema_migrations (id, version, filename, checksum, applied_at)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		 ON CONFLICT (version) DO NOTHING`,
		m.ID, m.Version, m.Filename, m.Checksum, m.AppliedAt,
	)
	if err != nil {
		return fmt.Errorf("record migration: %w", err)
	}
	return nil
}

// EnsureSchemaMigrationsTable creates the schema_migrations table if it does not exist.
// This is the bootstrap step that runs before any migration tracking is possible.
// It must be idempotent — safe to call on every startup.
func (r *SchemaMigrationRepo) EnsureSchemaMigrationsTable(ctx context.Context) error {
	_, err := r.db.Pool.Exec(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			version     VARCHAR(50)   NOT NULL UNIQUE,
			filename    VARCHAR(200)  NOT NULL,
			checksum    VARCHAR(128),
			applied_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	// Create the index if it doesn't exist (table-level CREATE IF NOT EXISTS
	// handles the idempotency; index may have been dropped manually).
	_, err = r.db.Pool.Exec(ctx,
		`CREATE INDEX IF NOT EXISTS idx_schema_migrations_version
		 ON schema_migrations (version)`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations index: %w", err)
	}

	return nil
}

// Compile-time check that SchemaMigrationRepo satisfies the interface from the
// repository package. We reference the interface type indirectly through a
// variable declaration to avoid circular imports (the tests wire it explicitly).
var _ interface {
	GetApplied(ctx context.Context) ([]*domain.SchemaMigration, error)
	IsApplied(ctx context.Context, version string) (bool, error)
	RecordApplied(ctx context.Context, m *domain.SchemaMigration) error
} = (*SchemaMigrationRepo)(nil)

// GetDB returns the underlying DB for use by the migration runner.
func (r *SchemaMigrationRepo) GetDB() *DB {
	return r.db
}
