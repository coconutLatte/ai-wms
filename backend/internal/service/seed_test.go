package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestSeededAdminPassword verifies that the hardcoded bcrypt hash
// in migrations/000001_init_schema.sql correctly validates the
// default admin password "admin123".
func TestSeededAdminPassword(t *testing.T) {
	// This hash is hardcoded in migrations/000001_init_schema.sql
	const hardcodedHash = "$2a$10$3oMF9qsdLBRGrIs8oQD/Z.Eon1bZ.EHLPToHmwsrbpwTlZq8YE5sW"

	err := bcrypt.CompareHashAndPassword([]byte(hardcodedHash), []byte("admin123"))
	if err != nil {
		t.Fatalf("seeded admin password hash does NOT match 'admin123': %v", err)
	}

	// Also verify that a wrong password does NOT match.
	err = bcrypt.CompareHashAndPassword([]byte(hardcodedHash), []byte("wrongpassword"))
	if err == nil {
		t.Fatal("expected wrong password to fail validation, but it succeeded")
	}
}

// TestHashPasswordRoundtrip verifies that HashPassword produces
// hashes that can be validated by bcrypt.
func TestHashPasswordRoundtrip(t *testing.T) {
	password := "admin123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Fatalf("roundtrip validation failed: %v", err)
	}

	// Wrong password should not validate.
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrong")); err == nil {
		t.Fatal("wrong password should not validate")
	}
}
