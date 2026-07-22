-- Migration Tracking Table
-- Migration: 000003_schema_migrations
-- Tracks which SQL migration files have been applied so each runs exactly once.
-- The migration runner reads .sql files from the migrations/ directory in order,
-- checks this table, and runs only the unapplied ones.

BEGIN;

CREATE TABLE schema_migrations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version     VARCHAR(50)   NOT NULL UNIQUE,     -- e.g. "000001", "000002"
    filename    VARCHAR(200)  NOT NULL,             -- e.g. "000001_init_schema.sql"
    checksum    VARCHAR(128),                       -- SHA-256 of file contents (future: detect drift)
    applied_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Index for fast lookup by version during startup.
CREATE INDEX idx_schema_migrations_version ON schema_migrations (version);

COMMIT;
