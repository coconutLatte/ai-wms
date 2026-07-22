package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MigrationFile represents a SQL migration file on disk.
type MigrationFile struct {
	Version  string // e.g. "000001"
	Filename string // e.g. "000001_init_schema.sql"
	Path     string // full path on disk
}

// DiscoverMigrations scans a directory for .sql migration files and returns them
// sorted by version (filename prefix). Files must follow the naming convention
// NNNNNN_description.sql where NNNNNN is a zero-padded sequential version number.
func DiscoverMigrations(dir string) ([]MigrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory %s: %w", dir, err)
	}

	var files []MigrationFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Extract version from filename prefix (e.g. "000001" from "000001_init_schema.sql").
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue // skip files that don't follow the naming convention
		}
		version := parts[0]
		if len(version) != 6 {
			continue
		}

		files = append(files, MigrationFile{
			Version:  version,
			Filename: name,
			Path:     filepath.Join(dir, name),
		})
	}

	// Sort by version (lexical sort works for zero-padded numeric strings).
	sort.Slice(files, func(i, j int) bool {
		return files[i].Version < files[j].Version
	})

	return files, nil
}

// RunMigrationsFromDir discovers and runs all unapplied SQL migration files
// from the given directory. Each migration runs inside a transaction and is
// recorded in the schema_migrations table on success. Already-applied migrations
// are skipped. This is idempotent and safe to call on every startup.
func (db *DB) RunMigrationsFromDir(ctx context.Context, dir string) error {
	repo := NewSchemaMigrationRepo(db)

	// Step 1: Ensure the tracking table exists (idempotent bootstrap).
	if err := repo.EnsureSchemaMigrationsTable(ctx); err != nil {
		return fmt.Errorf("bootstrap schema_migrations: %w", err)
	}

	// Step 2: Discover migration files on disk.
	files, err := DiscoverMigrations(dir)
	if err != nil {
		return fmt.Errorf("discover migrations: %w", err)
	}

	if len(files) == 0 {
		slog.Debug("No migration files found", "dir", dir)
		return nil
	}

	// Step 3: Determine which migrations need to run.
	var pending []MigrationFile
	for _, f := range files {
		applied, err := repo.IsApplied(ctx, f.Version)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", f.Version, err)
		}
		if !applied {
			pending = append(pending, f)
		}
	}

	if len(pending) == 0 {
		slog.Debug("All migrations already applied",
			"total", len(files),
			"latest", files[len(files)-1].Version,
		)
		return nil
	}

	// Step 4: Run pending migrations in order.
	slog.Info("Running pending migrations",
		"pending", len(pending),
		"total", len(files),
	)

	for _, f := range pending {
		if err := db.runMigrationFile(ctx, f); err != nil {
			return fmt.Errorf("migration %s (%s): %w", f.Version, f.Filename, err)
		}
	}

	slog.Info("Migrations complete",
		"applied", len(pending),
		"total", len(files),
	)

	return nil
}

// runMigrationFile reads a single SQL migration file, executes it in a transaction,
// and records it in the schema_migrations table.
func (db *DB) runMigrationFile(ctx context.Context, f MigrationFile) error {
	content, err := os.ReadFile(f.Path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Compute SHA-256 checksum for drift detection.
	checksum := fmt.Sprintf("%x", sha256.Sum256(content))

	// Execute the migration in a transaction so it's atomic: either the SQL
	// AND the tracking record succeed together, or neither does.
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("exec sql: %w", err)
	}

	// Record the migration as applied inside the same transaction.
	_, err = tx.Exec(ctx,
		`INSERT INTO schema_migrations (id, version, filename, checksum, applied_at)
		 VALUES (gen_random_uuid(), $1, $2, NULLIF($3, ''), NOW())
		 ON CONFLICT (version) DO NOTHING`,
		f.Version, f.Filename, checksum,
	)
	if err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	slog.Info("Migration applied",
		"version", f.Version,
		"filename", f.Filename,
	)

	return nil
}
