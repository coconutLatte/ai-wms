-- Cycle Count Management
-- Migration: 000004_cycle_counts

BEGIN;

-- ── Cycle Counts ─────────────────────────────────────────────

CREATE TABLE cycle_counts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    count_no      VARCHAR(50)  NOT NULL UNIQUE,
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    location_id   UUID REFERENCES locations(id) ON DELETE SET NULL,
    zone_id       UUID REFERENCES zones(id) ON DELETE SET NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'in_progress', 'pending_review', 'approved', 'adjusted', 'cancelled')),
    counted_by    VARCHAR(100),
    notes         TEXT,
    total_lines   INT NOT NULL DEFAULT 0,
    matched_lines INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    approved_at   TIMESTAMPTZ,
    approved_by   VARCHAR(100)
);

CREATE INDEX idx_cycle_counts_warehouse ON cycle_counts(warehouse_id);
CREATE INDEX idx_cycle_counts_status ON cycle_counts(status);
CREATE INDEX idx_cycle_counts_location ON cycle_counts(location_id);

-- ── Cycle Count Lines ────────────────────────────────────────

CREATE TABLE cycle_count_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_count_id  UUID NOT NULL REFERENCES cycle_counts(id) ON DELETE CASCADE,
    sku_id          UUID NOT NULL REFERENCES skus(id) ON DELETE RESTRICT,
    location_id     UUID NOT NULL REFERENCES locations(id) ON DELETE RESTRICT,
    batch_no        VARCHAR(100),
    system_qty      NUMERIC(18,6) NOT NULL DEFAULT 0,
    counted_qty     NUMERIC(18,6),
    variance        NUMERIC(18,6),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'counted', 'reviewed')),
    counted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cycle_count_lines_count ON cycle_count_lines(cycle_count_id);
CREATE INDEX idx_cycle_count_lines_sku ON cycle_count_lines(sku_id);
CREATE INDEX idx_cycle_count_lines_status ON cycle_count_lines(status);

COMMIT;
