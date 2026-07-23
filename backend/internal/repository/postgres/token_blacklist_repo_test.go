package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// setupTokenBlacklistTestDB creates a test database and cleans up token blacklist test data.
func setupTokenBlacklistTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	cfg := testConfig()

	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	// Clean up previous test data
	db.Pool.Exec(ctx, "DELETE FROM token_blacklist WHERE jti LIKE 'TEST-%'")

	cleanup := func() {
		db.Pool.Exec(ctx, "DELETE FROM token_blacklist WHERE jti LIKE 'TEST-%'")
		db.Close()
	}

	return db, cleanup
}

func TestTokenBlacklistRepo_AddAndCheck(t *testing.T) {
	db, cleanup := setupTokenBlacklistTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewTokenBlacklistRepo(db)

	entry := &domain.TokenBlacklistEntry{
		JTI:       "TEST-JTI-" + uuid.New().String(),
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Add
	if err := repo.Add(ctx, entry); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if entry.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if entry.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	// Check — should be blacklisted
	blocked, err := repo.IsBlacklisted(ctx, entry.JTI)
	if err != nil {
		t.Fatalf("IsBlacklisted failed: %v", err)
	}
	if !blocked {
		t.Error("expected JTI to be blacklisted")
	}

	// Check non-existent JTI
	blocked, err = repo.IsBlacklisted(ctx, "NONEXISTENT-JTI-"+uuid.New().String())
	if err != nil {
		t.Fatalf("IsBlacklisted for nonexistent failed: %v", err)
	}
	if blocked {
		t.Error("expected non-existent JTI to NOT be blacklisted")
	}
}

func TestTokenBlacklistRepo_AddDuplicate(t *testing.T) {
	db, cleanup := setupTokenBlacklistTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewTokenBlacklistRepo(db)

	jti := "TEST-JTI-DUP-" + uuid.New().String()
	entry1 := &domain.TokenBlacklistEntry{
		JTI:       jti,
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	entry2 := &domain.TokenBlacklistEntry{
		JTI:       jti,
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}

	if err := repo.Add(ctx, entry1); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Adding the same JTI should fail
	err := repo.Add(ctx, entry2)
	if err == nil {
		t.Error("expected error when adding duplicate JTI")
	}
}

func TestTokenBlacklistRepo_DeleteExpired(t *testing.T) {
	db, cleanup := setupTokenBlacklistTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewTokenBlacklistRepo(db)

	// Add an already-expired token and a valid token
	expired := &domain.TokenBlacklistEntry{
		JTI:       "TEST-JTI-EXPIRED-" + uuid.New().String(),
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired 1 hour ago
	}
	valid := &domain.TokenBlacklistEntry{
		JTI:       "TEST-JTI-VALID-" + uuid.New().String(),
		UserID:    uuid.New(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // valid for 24 hours
	}

	if err := repo.Add(ctx, expired); err != nil {
		t.Fatalf("Add (expired) failed: %v", err)
	}
	if err := repo.Add(ctx, valid); err != nil {
		t.Fatalf("Add (valid) failed: %v", err)
	}

	// Delete expired tokens
	deleted, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted token, got %d", deleted)
	}

	// Expired should be gone
	blocked, err := repo.IsBlacklisted(ctx, expired.JTI)
	if err != nil {
		t.Fatalf("IsBlacklisted (expired) failed: %v", err)
	}
	if blocked {
		t.Error("expected expired JTI to be removed from blacklist")
	}

	// Valid should still be there
	blocked, err = repo.IsBlacklisted(ctx, valid.JTI)
	if err != nil {
		t.Fatalf("IsBlacklisted (valid) failed: %v", err)
	}
	if !blocked {
		t.Error("expected valid JTI to still be blacklisted")
	}
}

func TestTokenBlacklistRepo_DeleteExpired_None(t *testing.T) {
	db, cleanup := setupTokenBlacklistTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewTokenBlacklistRepo(db)

	// No expired tokens — should succeed and return 0
	deleted, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted tokens, got %d", deleted)
	}
}
