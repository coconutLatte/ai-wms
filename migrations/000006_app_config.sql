-- System Configuration
-- Migration: 000006_app_config
-- Stores application-level settings in a single JSONB document.

BEGIN;

CREATE TABLE app_config (
    id          INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1), -- single-row table
    config      JSONB NOT NULL DEFAULT '{}',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default configuration.
INSERT INTO app_config (config) VALUES ('{
    "site_name": "AI-WMS",
    "default_warehouse_id": "",
    "low_stock_threshold": 10,
    "default_page_size": 20,
    "jwt_access_ttl": 3600
}'::jsonb);

COMMIT;
