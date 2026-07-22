package domain

import (
	"time"

	"github.com/google/uuid"
)

// SchemaMigration represents a database migration that has been applied.
// The migration runner reads SQL files from the migrations/ directory and
// tracks each applied migration in the schema_migrations table. This ensures
// each .sql file runs exactly once.
type SchemaMigration struct {
	ID        uuid.UUID `json:"id"`
	Version   string    `json:"version"`    // e.g. "000001"
	Filename  string    `json:"filename"`   // e.g. "000001_init_schema.sql"
	Checksum  string    `json:"checksum"`   // SHA-256 of file contents
	AppliedAt time.Time `json:"applied_at"`
}

// NewSchemaMigration creates a new SchemaMigration record with a generated ID
// and the current timestamp.
func NewSchemaMigration(version, filename, checksum string) *SchemaMigration {
	return &SchemaMigration{
		ID:        uuid.New(),
		Version:   version,
		Filename:  filename,
		Checksum:  checksum,
		AppliedAt: time.Now(),
	}
}
