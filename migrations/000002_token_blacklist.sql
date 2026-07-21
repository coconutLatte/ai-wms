-- Token Blacklist
-- Migration: 000002_token_blacklist
-- Stores revoked refresh token JTIs so they cannot be used to obtain new access tokens.
-- Rows are automatically cleaned up after their natural expiry via the expires_at column.
-- Access tokens (short-lived, ~15 min) do NOT need blacklisting.

BEGIN;

CREATE TABLE token_blacklist (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jti         VARCHAR(64)  NOT NULL UNIQUE,    -- JWT Token ID (jti claim)
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ  NOT NULL,            -- When the original token would have expired
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index for fast JTI lookups during refresh token validation.
CREATE INDEX idx_token_blacklist_jti ON token_blacklist (jti);

-- Index for periodic cleanup of expired entries.
CREATE INDEX idx_token_blacklist_expires ON token_blacklist (expires_at);

COMMIT;
