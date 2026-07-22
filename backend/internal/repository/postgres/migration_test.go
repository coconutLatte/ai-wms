package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// ── DiscoverMigrations (unit tests — no DB required) ──────────────────────────

func TestDiscoverMigrations_SortsByVersion(t *testing.T) {
	dir := t.TempDir()

	// Create files out of order to verify sorting.
	createFile(t, dir, "000003_third.sql")
	createFile(t, dir, "000001_first.sql")
	createFile(t, dir, "000002_second.sql")
	createFile(t, dir, "not-a-migration.txt") // should be ignored

	files, err := DiscoverMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 migration files, got %d", len(files))
	}

	// Verify sorted order.
	if files[0].Version != "000001" {
		t.Errorf("expected version 000001, got %s", files[0].Version)
	}
	if files[1].Version != "000002" {
		t.Errorf("expected version 000002, got %s", files[1].Version)
	}
	if files[2].Version != "000003" {
		t.Errorf("expected version 000003, got %s", files[2].Version)
	}

	// Verify filenames are preserved.
	if files[0].Filename != "000001_first.sql" {
		t.Errorf("expected filename 000001_first.sql, got %s", files[0].Filename)
	}
}

func TestDiscoverMigrations_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	files, err := DiscoverMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestDiscoverMigrations_NonexistentDirectory(t *testing.T) {
	_, err := DiscoverMigrations("/nonexistent/path/to/migrations")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestDiscoverMigrations_IgnoresNonSQLFiles(t *testing.T) {
	dir := t.TempDir()

	createFile(t, dir, "000001_init.sql")
	createFile(t, dir, "readme.md")
	createFile(t, dir, "notes.txt")
	createFile(t, dir, "migration_backup.sql.bak")

	files, err := DiscoverMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 migration file, got %d", len(files))
	}
	if files[0].Filename != "000001_init.sql" {
		t.Errorf("expected 000001_init.sql, got %s", files[0].Filename)
	}
}

func TestDiscoverMigrations_IgnoresFilesWithoutVersionPrefix(t *testing.T) {
	dir := t.TempDir()

	createFile(t, dir, "not_migration.sql") // no underscore separator
	createFile(t, dir, "001_too_short.sql") // version not 6 chars

	files, err := DiscoverMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

// ── Integration tests (require PostgreSQL) ────────────────────────────────────

func setupMigrationTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	cfg := testConfig()

	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	cleanup := func() {
		db.Pool.Exec(ctx, "DROP TABLE IF EXISTS schema_migrations CASCADE")
		db.Close()
	}

	// Start with a clean state.
	db.Pool.Exec(ctx, "DROP TABLE IF EXISTS schema_migrations CASCADE")

	return db, cleanup
}

func TestSchemaMigrationRepo_EnsureTable(t *testing.T) {
	db, cleanup := setupMigrationTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewSchemaMigrationRepo(db)

	// First call — creates the table.
	if err := repo.EnsureSchemaMigrationsTable(ctx); err != nil {
		t.Fatalf("first EnsureSchemaMigrationsTable failed: %v", err)
	}

	// Second call — idempotent, should not fail.
	if err := repo.EnsureSchemaMigrationsTable(ctx); err != nil {
		t.Fatalf("second EnsureSchemaMigrationsTable failed: %v", err)
	}

	// Verify table exists by inserting and querying.
	applied, err := repo.GetApplied(ctx)
	if err != nil {
		t.Fatalf("GetApplied failed: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("expected 0 applied migrations, got %d", len(applied))
	}
}

func TestSchemaMigrationRepo_RecordAndCheckApplied(t *testing.T) {
	db, cleanup := setupMigrationTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewSchemaMigrationRepo(db)

	if err := repo.EnsureSchemaMigrationsTable(ctx); err != nil {
		t.Fatalf("EnsureSchemaMigrationsTable failed: %v", err)
	}

	// Initially nothing is applied.
	applied, err := repo.IsApplied(ctx, "000001")
	if err != nil {
		t.Fatalf("IsApplied failed: %v", err)
	}
	if applied {
		t.Error("expected migration 000001 to NOT be applied initially")
	}

	// Record a migration.
	m := domain.NewSchemaMigration("000001", "000001_test.sql", "abc123")
	if err := repo.RecordApplied(ctx, m); err != nil {
		t.Fatalf("RecordApplied failed: %v", err)
	}

	// Now it should show as applied.
	applied, err = repo.IsApplied(ctx, "000001")
	if err != nil {
		t.Fatalf("IsApplied failed: %v", err)
	}
	if !applied {
		t.Error("expected migration 000001 to be applied after recording")
	}

	// Another migration should still not be applied.
	applied, err = repo.IsApplied(ctx, "000002")
	if err != nil {
		t.Fatalf("IsApplied failed: %v", err)
	}
	if applied {
		t.Error("expected migration 000002 to NOT be applied")
	}

	// GetApplied should return the recorded migration.
	all, err := repo.GetApplied(ctx)
	if err != nil {
		t.Fatalf("GetApplied failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 applied migration, got %d", len(all))
	}
	if all[0].Version != "000001" {
		t.Errorf("expected version 000001, got %s", all[0].Version)
	}
	if all[0].Filename != "000001_test.sql" {
		t.Errorf("expected filename 000001_test.sql, got %s", all[0].Filename)
	}
	if all[0].Checksum != "abc123" {
		t.Errorf("expected checksum abc123, got %s", all[0].Checksum)
	}
}

func TestSchemaMigrationRepo_DuplicateRecord(t *testing.T) {
	db, cleanup := setupMigrationTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewSchemaMigrationRepo(db)

	if err := repo.EnsureSchemaMigrationsTable(ctx); err != nil {
		t.Fatalf("EnsureSchemaMigrationsTable failed: %v", err)
	}

	// Record the same migration twice (ON CONFLICT DO NOTHING).
	m1 := domain.NewSchemaMigration("000001", "000001_test.sql", "aaa")
	if err := repo.RecordApplied(ctx, m1); err != nil {
		t.Fatalf("first RecordApplied failed: %v", err)
	}

	m2 := domain.NewSchemaMigration("000001", "000001_test.sql", "bbb")
	if err := repo.RecordApplied(ctx, m2); err != nil {
		t.Fatalf("second RecordApplied failed: %v", err)
	}

	// Should still have exactly one record.
	all, err := repo.GetApplied(ctx)
	if err != nil {
		t.Fatalf("GetApplied failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 applied migration, got %d", len(all))
	}
}

func TestRunMigrationsFromDir(t *testing.T) {
	db, cleanup := setupMigrationTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// Create a temp directory with test migration files.
	dir := t.TempDir()
	createFile(t, dir, "000001_create_products.sql",
		`CREATE TABLE IF NOT EXISTS test_products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	createFile(t, dir, "000002_add_products_index.sql",
		`CREATE INDEX IF NOT EXISTS idx_test_products_name ON test_products (name)`)

	// Run migrations.
	if err := db.RunMigrationsFromDir(ctx, dir); err != nil {
		t.Fatalf("first RunMigrationsFromDir failed: %v", err)
	}

	// Verify the table and index exist.
	repo := NewSchemaMigrationRepo(db)
	applied, err := repo.GetApplied(ctx)
	if err != nil {
		t.Fatalf("GetApplied failed: %v", err)
	}
	if len(applied) != 2 {
		t.Fatalf("expected 2 applied migrations, got %d", len(applied))
	}
	if applied[0].Version != "000001" {
		t.Errorf("expected version 000001, got %s", applied[0].Version)
	}
	if applied[1].Version != "000002" {
		t.Errorf("expected version 000002, got %s", applied[1].Version)
	}

	// Run again — should be idempotent (no new migrations applied).
	if err := db.RunMigrationsFromDir(ctx, dir); err != nil {
		t.Fatalf("second RunMigrationsFromDir failed: %v", err)
	}
	applied, err = repo.GetApplied(ctx)
	if err != nil {
		t.Fatalf("GetApplied after rerun failed: %v", err)
	}
	if len(applied) != 2 {
		t.Errorf("expected still 2 applied migrations after rerun, got %d", len(applied))
	}

	// Clean up the test table.
	db.Pool.Exec(ctx, "DROP TABLE IF EXISTS test_products CASCADE")
}

func TestRunMigrationsFromDir_EmptyDir(t *testing.T) {
	db, cleanup := setupMigrationTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	dir := t.TempDir()

	// Running with no migration files should succeed (no-op).
	if err := db.RunMigrationsFromDir(ctx, dir); err != nil {
		t.Fatalf("RunMigrationsFromDir with empty dir failed: %v", err)
	}
}

func TestRunMigrationsFromDir_PartialApply(t *testing.T) {
	db, cleanup := setupMigrationTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewSchemaMigrationRepo(db)
	dir := t.TempDir()

	createFile(t, dir, "000001_first.sql",
		`CREATE TABLE IF NOT EXISTS test_first (id SERIAL PRIMARY KEY)`)
	createFile(t, dir, "000002_second.sql",
		`CREATE TABLE IF NOT EXISTS test_second (id SERIAL PRIMARY KEY)`)
	createFile(t, dir, "000003_third.sql",
		`CREATE TABLE IF NOT EXISTS test_third (id SERIAL PRIMARY KEY)`)

	// Pre-record migration 000001 as already applied (simulating previous run).
	if err := repo.EnsureSchemaMigrationsTable(ctx); err != nil {
		t.Fatalf("EnsureSchemaMigrationsTable failed: %v", err)
	}
	m := domain.NewSchemaMigration("000001", "000001_first.sql", "")
	if err := repo.RecordApplied(ctx, m); err != nil {
		t.Fatalf("pre-record first migration: %v", err)
	}

	// Run migrations — should skip 000001 and apply 000002, 000003.
	if err := db.RunMigrationsFromDir(ctx, dir); err != nil {
		t.Fatalf("RunMigrationsFromDir failed: %v", err)
	}

	// Verify all 3 are now recorded.
	applied, err := repo.GetApplied(ctx)
	if err != nil {
		t.Fatalf("GetApplied failed: %v", err)
	}
	if len(applied) != 3 {
		t.Fatalf("expected 3 applied migrations, got %d", len(applied))
	}
	if applied[0].Version != "000001" {
		t.Errorf("expected version 000001, got %s", applied[0].Version)
	}
	if applied[1].Version != "000002" {
		t.Errorf("expected version 000002, got %s", applied[1].Version)
	}
	if applied[2].Version != "000003" {
		t.Errorf("expected version 000003, got %s", applied[2].Version)
	}

	// Clean up test tables.
	db.Pool.Exec(ctx, "DROP TABLE IF EXISTS test_first, test_second, test_third CASCADE")
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// createFile creates a file in the given directory with optional content.
func createFile(t *testing.T, dir, filename string, content ...string) {
	t.Helper()
	path := filepath.Join(dir, filename)
	var data string
	if len(content) > 0 {
		data = content[0]
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatalf("failed to create test file %s: %v", filename, err)
	}
}
