-- WMS Initial Schema
-- Migration: 000001_init_schema

BEGIN;

-- ── Warehouses ─────────────────────────────────────────────

CREATE TABLE warehouses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    address     TEXT,
    status      VARCHAR(20)  NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'archived')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Zones ──────────────────────────────────────────────────

CREATE TABLE zones (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    code          VARCHAR(50)  NOT NULL,
    name          VARCHAR(200) NOT NULL,
    zone_type     VARCHAR(20)  NOT NULL
        CHECK (zone_type IN ('receiving', 'storage', 'picking', 'shipping', 'returns', 'staging')),
    status        VARCHAR(20)  NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'full')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(warehouse_id, code)
);

-- ── Locations ──────────────────────────────────────────────

CREATE TABLE locations (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id       UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    code          VARCHAR(100) NOT NULL,
    barcode       VARCHAR(200),
    location_type VARCHAR(20)  NOT NULL
        CHECK (location_type IN ('pallet', 'shelf', 'floor', 'conveyor', 'agv')),
    status        VARCHAR(20)  NOT NULL DEFAULT 'empty'
        CHECK (status IN ('empty', 'occupied', 'reserved', 'blocked')),
    max_weight    NUMERIC(10,3),   -- kg
    max_volume    NUMERIC(10,6),   -- m³
    max_qty       INT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(warehouse_id, code)
);

-- ── SKUs ───────────────────────────────────────────────────

CREATE TABLE skus (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(100) NOT NULL UNIQUE,
    name        VARCHAR(300) NOT NULL,
    description TEXT,
    barcode     VARCHAR(200) UNIQUE,
    base_unit   VARCHAR(10)  NOT NULL DEFAULT 'EA',
    pack_unit   VARCHAR(10),
    pack_qty    INT          NOT NULL DEFAULT 1,
    weight      NUMERIC(10,3),   -- kg per base unit
    volume      NUMERIC(10,6),   -- m³ per base unit
    length      NUMERIC(10,2),   -- cm
    width       NUMERIC(10,2),   -- cm
    height      NUMERIC(10,2),   -- cm
    category    VARCHAR(200),
    attributes  JSONB        DEFAULT '{}',
    status      VARCHAR(20)  NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'discontinued')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Inventory ──────────────────────────────────────────────

CREATE TABLE inventory (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku_id          UUID NOT NULL REFERENCES skus(id),
    location_id     UUID NOT NULL REFERENCES locations(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    batch_no        VARCHAR(100),
    qty             NUMERIC(15,3) NOT NULL DEFAULT 0,
    reserved_qty    NUMERIC(15,3) NOT NULL DEFAULT 0,
    status          VARCHAR(20)    NOT NULL DEFAULT 'available'
        CHECK (status IN ('available', 'quarantine', 'damaged', 'expired')),
    production_date DATE,
    expiry_date     DATE,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(sku_id, location_id, batch_no)
);

CREATE INDEX idx_inventory_sku ON inventory(sku_id);
CREATE INDEX idx_inventory_location ON inventory(location_id);
CREATE INDEX idx_inventory_warehouse ON inventory(warehouse_id);
CREATE INDEX idx_inventory_expiry ON inventory(expiry_date) WHERE expiry_date IS NOT NULL;

-- ── Inventory Transactions ─────────────────────────────────

CREATE TABLE inventory_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_id    UUID NOT NULL REFERENCES inventory(id),
    sku_id          UUID NOT NULL REFERENCES skus(id),
    location_id     UUID NOT NULL REFERENCES locations(id),
    type            VARCHAR(20) NOT NULL
        CHECK (type IN ('receipt', 'putaway', 'pick', 'ship', 'transfer', 'adjustment', 'return')),
    delta_qty       NUMERIC(15,3) NOT NULL,
    resulting_qty   NUMERIC(15,3) NOT NULL,
    reference_type  VARCHAR(50),
    reference_id    UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      VARCHAR(100)
);

CREATE INDEX idx_inv_tx_inventory ON inventory_transactions(inventory_id);
CREATE INDEX idx_inv_tx_created_at ON inventory_transactions(created_at);

-- ── Orders ─────────────────────────────────────────────────

CREATE TABLE orders (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no      VARCHAR(100) NOT NULL UNIQUE,
    order_type    VARCHAR(20)  NOT NULL
        CHECK (order_type IN ('inbound', 'outbound', 'transfer', 'return')),
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id),
    status        VARCHAR(20)  NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'confirmed', 'processing', 'partial', 'completed', 'cancelled')),
    priority      VARCHAR(10)  NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    external_ref  VARCHAR(200),
    external_type VARCHAR(50),
    notes         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at  TIMESTAMPTZ,
    created_by    VARCHAR(100)
);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_external_ref ON orders(external_ref);

-- ── Order Lines ────────────────────────────────────────────

CREATE TABLE order_lines (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id       UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    line_no        INT  NOT NULL,
    sku_id         UUID NOT NULL REFERENCES skus(id),
    ordered_qty    NUMERIC(15,3) NOT NULL,
    fulfilled_qty  NUMERIC(15,3) NOT NULL DEFAULT 0,
    uom            VARCHAR(10)   NOT NULL DEFAULT 'EA',
    batch_no       VARCHAR(100),
    status         VARCHAR(20)   NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'allocated', 'partial', 'fulfilled', 'cancelled')),
    notes          TEXT,
    UNIQUE(order_id, line_no)
);

-- ── ASN ────────────────────────────────────────────────────

CREATE TABLE asns (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asn_no        VARCHAR(100) NOT NULL UNIQUE,
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id),
    order_id      UUID REFERENCES orders(id),
    carrier       VARCHAR(200),
    tracking_no   VARCHAR(200),
    expected_at   TIMESTAMPTZ NOT NULL,
    arrived_at    TIMESTAMPTZ,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'arrived', 'receiving', 'partial', 'received')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE asn_lines (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asn_id       UUID NOT NULL REFERENCES asns(id) ON DELETE CASCADE,
    sku_id       UUID NOT NULL REFERENCES skus(id),
    expected_qty NUMERIC(15,3) NOT NULL,
    received_qty NUMERIC(15,3) NOT NULL DEFAULT 0,
    batch_no     VARCHAR(100),
    status       VARCHAR(20)  NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'partial', 'received'))
);

-- ── Tasks ──────────────────────────────────────────────────

CREATE TABLE tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_no         VARCHAR(100) NOT NULL UNIQUE,
    task_type       VARCHAR(20)  NOT NULL
        CHECK (task_type IN ('putaway', 'pick', 'replenish', 'transfer', 'cycle_count', 'load', 'unload')),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    order_id        UUID REFERENCES orders(id),
    order_line_id   UUID REFERENCES order_lines(id),
    priority        VARCHAR(10)  NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'assigned', 'in_progress', 'paused', 'completed', 'cancelled', 'exception')),
    assigned_to     VARCHAR(100),
    from_location_id UUID REFERENCES locations(id),
    to_location_id   UUID REFERENCES locations(id),
    sku_id          UUID NOT NULL REFERENCES skus(id),
    expected_qty    NUMERIC(15,3) NOT NULL,
    actual_qty      NUMERIC(15,3) NOT NULL DEFAULT 0,
    uom             VARCHAR(10) NOT NULL DEFAULT 'EA',
    batch_no        VARCHAR(100),
    instructions    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assigned ON tasks(assigned_to) WHERE status = 'assigned';

-- ── Waves ──────────────────────────────────────────────────

CREATE TABLE waves (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wave_no       VARCHAR(100) NOT NULL UNIQUE,
    warehouse_id  UUID NOT NULL REFERENCES warehouses(id),
    wave_type     VARCHAR(20)  NOT NULL
        CHECK (wave_type IN ('single_order', 'batch', 'zone', 'carrier')),
    status        VARCHAR(20)  NOT NULL DEFAULT 'created'
        CHECK (status IN ('created', 'released', 'in_progress', 'completed')),
    order_ids     UUID[]       NOT NULL DEFAULT '{}',
    task_ids      UUID[]       NOT NULL DEFAULT '{}',
    total_orders  INT NOT NULL DEFAULT 0,
    total_lines   INT NOT NULL DEFAULT 0,
    total_qty     NUMERIC(15,3) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at   TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ
);

-- ── Users ──────────────────────────────────────────────────

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(100) NOT NULL UNIQUE,
    email         VARCHAR(200) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name  VARCHAR(200),
    role_ids      UUID[]       NOT NULL DEFAULT '{}',
    status        VARCHAR(20)  NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'locked')),
    last_login    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Roles ──────────────────────────────────────────────────

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Audit Logs ─────────────────────────────────────────────

CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    username    VARCHAR(100) NOT NULL,
    action      VARCHAR(100) NOT NULL,
    resource    VARCHAR(100) NOT NULL,
    resource_id VARCHAR(200),
    details     TEXT,
    ip_address  VARCHAR(50),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);

-- ── Seed Data ──────────────────────────────────────────────

-- Default roles
INSERT INTO roles (id, name, description, permissions) VALUES
    (gen_random_uuid(), 'admin', 'System Administrator',
     '[{"resource":"*","actions":["*"]}]'),
    (gen_random_uuid(), 'operator', 'Warehouse Operator',
     '[{"resource":"warehouse","actions":["read"]},{"resource":"inventory","actions":["read","update"]},{"resource":"order","actions":["read","create","update"]},{"resource":"task","actions":["read","update"]}]'),
    (gen_random_uuid(), 'picker', 'Picker (PDA User)',
     '[{"resource":"task","actions":["read","update"]},{"resource":"inventory","actions":["read"]}]');

-- Default admin user (password: admin123 — CHANGE IN PRODUCTION)
INSERT INTO users (id, username, email, password_hash, display_name, role_ids) VALUES
    (gen_random_uuid(), 'admin', 'admin@wms.local',
     '$2a$10$dummy_hash_replace_with_real_bcrypt', -- placeholder
     'System Admin',
     ARRAY(SELECT id FROM roles WHERE name = 'admin'));

COMMIT;
