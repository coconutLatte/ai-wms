package domain

import (
	"testing"
)

func TestNewSchemaMigration(t *testing.T) {
	m := NewSchemaMigration("000001", "000001_init_schema.sql", "abc123")

	if m.ID.String() == "" {
		t.Error("expected non-empty ID")
	}
	if m.Version != "000001" {
		t.Errorf("expected version 000001, got %s", m.Version)
	}
	if m.Filename != "000001_init_schema.sql" {
		t.Errorf("expected filename 000001_init_schema.sql, got %s", m.Filename)
	}
	if m.Checksum != "abc123" {
		t.Errorf("expected checksum abc123, got %s", m.Checksum)
	}
	if m.AppliedAt.IsZero() {
		t.Error("expected non-zero AppliedAt")
	}
}

func TestNewSchemaMigrationEmptyChecksum(t *testing.T) {
	m := NewSchemaMigration("000002", "000002_token_blacklist.sql", "")

	if m.Version != "000002" {
		t.Errorf("expected version 000002, got %s", m.Version)
	}
	if m.Checksum != "" {
		t.Errorf("expected empty checksum, got %s", m.Checksum)
	}
	if m.AppliedAt.IsZero() {
		t.Error("expected non-zero AppliedAt")
	}
}
