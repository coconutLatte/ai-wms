-- Shipment Management
-- Migration: 000005_shipments

BEGIN;

-- ── Shipments ─────────────────────────────────────────────────

CREATE TABLE shipments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipment_no       VARCHAR(50)  NOT NULL UNIQUE,
    order_id          UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    warehouse_id      UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT,
    status            VARCHAR(20)  NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'in_transit', 'delivered', 'cancelled')),
    carrier           VARCHAR(100) NOT NULL,
    tracking_no       VARCHAR(100),
    carrier_service   VARCHAR(50),
    estimated_delivery TIMESTAMPTZ,
    actual_delivery   TIMESTAMPTZ,
    notes             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    shipped_at        TIMESTAMPTZ,
    delivered_at      TIMESTAMPTZ
);

CREATE INDEX idx_shipments_order ON shipments(order_id);
CREATE INDEX idx_shipments_warehouse ON shipments(warehouse_id);
CREATE INDEX idx_shipments_status ON shipments(status);
CREATE INDEX idx_shipments_carrier ON shipments(carrier);
CREATE INDEX idx_shipments_tracking ON shipments(tracking_no);

COMMIT;
