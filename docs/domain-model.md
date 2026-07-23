# Domain Model

## Entity Overview

```
Warehouse (1) ──→ (*) Zone (1) ──→ (*) Location
                                          │
                                          │ stores
                                          ▼
SKU (1) ──→ (*) Inventory ──── at ─── Location
                │
                │ tracks changes via
                ▼
          InventoryTransaction

Order (1) ──→ (*) OrderLine ─── references ──→ SKU
  │
  │ inbound orders have
  ▼
ASN (1) ──→ (*) ASNLine ─── references ──→ SKU

Task ─── operates on ──→ Inventory (at Location)
  │                       references Order/OrderLine
  │
  │ grouped into
  ▼
Wave (1) ──→ (*) Order → (*) Task

CycleCount (1) ──→ (*) CycleCountLine ─── references ──→ SKU, Location
  │
  │ applied via
  ▼
InventoryTransaction (adjustment)

User (1) ──→ (*) Role ──→ (*) Permission
```

## Entity Details

### Warehouse
The top-level organizational unit. Represents a physical building.
- **Status**: active, inactive, archived
- A Warehouse contains multiple Zones

### Zone
A logical area within a warehouse with a specific function.
- **Types**: receiving, storage, picking, shipping, returns, staging
- **Status**: active, inactive, full
- A Zone contains multiple Locations

### Location
A specific storage position (shelf, bin, pallet position, conveyor slot, AGV dock).
- **Types**: pallet, shelf, floor, conveyor, agv
- **Status**: empty, occupied, reserved, blocked
- **Status machine** (enforced via `CanTransitionTo`):
  ```
  empty ──→ occupied
  empty ──→ reserved
  empty ──→ blocked
  occupied ──→ empty
  occupied ──→ blocked
  reserved ──→ occupied
  reserved ──→ empty
  reserved ──→ blocked
  blocked ──→ empty
  ```
  - `IsTerminal()` returns true only for `blocked` (location is out of service)
  - Same-status transitions are rejected (no-op guard)
- Has optional Capacity (max weight, volume, quantity)
- Identified by a barcode label

### SKU (Stock Keeping Unit)
A unique product variant. The master data for everything stored in the warehouse.
- Has physical attributes (weight, dimensions)
- Has UOM (base unit, pack unit, conversion ratio)
- Has flexible Attributes (JSONB in DB)
- **Status**: active, inactive, discontinued

### Inventory
The physical presence of a SKU at a Location. This is the **most critical** entity.
- `qty` = on-hand quantity
- `reserved_qty` = allocated to orders but not yet picked
- `available_qty` = qty - reserved_qty (computed, not stored)
- Batch/lot tracked per inventory record
- **Status**: available, quarantine, damaged, expired
- FEFO/FIFO based on production_date and expiry_date

### InventoryTransaction
Immutable audit record of every inventory change.
- **Types**: receipt, putaway, pick, ship, transfer, adjustment, return
- Records `delta_qty` (change) and `resulting_qty` (after)
- Links to reference document (order, task, etc.)

### Order
A business document requesting inventory movement.
- **Types**: inbound (receiving), outbound (shipping), transfer, return
- **Status flow**: draft → confirmed → processing → partial → completed (or cancelled)
- Has Priority: low, normal, high, urgent
- Links to external systems via external_ref

### OrderLine
A single line item in an order. Tracks ordered vs fulfilled quantities.
- **Status**: pending → allocated → partial → fulfilled (or cancelled)

### ASN (Advanced Shipping Notice)
Pre-notification of an inbound delivery. Links to an inbound Order.
- **Status**: pending → arrived → receiving → partial → received

### Task
A single warehouse operation assigned to a worker.
- **Types**: putaway, pick, replenish, transfer, cycle_count, load, unload
- **Status flow**: pending → assigned → in_progress → completed
- Has from_location and to_location for movement tasks
- Has expected_qty vs actual_qty for variance tracking

### Wave
A group of orders batched for efficient picking.
- **Types**: single_order, batch, zone, carrier
- Contains a set of Order IDs and generated Task IDs
- **Status flow**: created → released → in_progress → completed
- Managed via WaveService with Admin API (POST/GET /waves, PUT /waves/{id}/status, POST /waves/{id}/release, POST/DELETE /waves/{id}/orders)

### CycleCount
A physical inventory count for a specific warehouse area (location or zone).
- Created via `POST /cycle-counts` — auto-generates lines from current inventory records
- Contains CycleCountLine entries, each representing a SKU/batch at a location
- **Status flow**: draft → in_progress → pending_review → approved/adjusted → (terminal)
  - Any non-terminal stage can also transition to cancelled
- **State machine** implemented in `domain.CycleCount.CanTransitionTo()`
- Lines track system_qty vs counted_qty, computing variance automatically
- Approval can apply inventory adjustments atomically (CycleCountService.ApproveCount with "approve" action)

### User
A system user (admin or warehouse operator).
- Linked to Roles for permissions
- **Status**: active, inactive, locked

### Role
A named collection of Permissions.
- Standard roles: admin, operator, picker

### Permission
Resource + Actions pair. Example: `{resource: "warehouse", actions: ["read","create"]}`
- Resource "*" = all resources
- Action "*" = all actions

## State Machines

### Order Lifecycle
```
draft → confirmed → processing → completed
  │         │            │
  └─────────┴────────────┴──→ cancelled
                    │
                    └──→ partial (some lines still pending)
```

### Task Lifecycle
```
pending → assigned → in_progress → completed
  │          │           │
  └──────────┴───────────┴──→ cancelled
                     │
                     └──→ exception (needs human)
                              │
                              └──→ in_progress (resolved)
```

### Wave Lifecycle
```
created → released → in_progress → completed
```
- Orders can only be added/removed in `created` status.
- `completed` is terminal — no further transitions allowed.

### CycleCount Lifecycle
```
draft → in_progress → pending_review → approved → (terminal)
  │         │             │              │
  │         │             │              └──→ adjusted → (terminal)
  └─────────┴─────────────┴──────────────┴──→ cancelled → (terminal)
```
- Lines are generated from current inventory when count starts.
- Operator submits counted qty per line; variance = counted_qty - system_qty.
- At finalize, all lines must be counted; count moves to pending_review.
- Approve action applies inventory adjustments; adjust action marks without applying changes.

### Shipment (Outbound)
- **Fields**: id, shipment_no, order_id, warehouse_id, status, carrier, tracking_no, carrier_service, estimated_delivery, actual_delivery, notes, timestamps
- **Statuses**: pending, in_transit, delivered, cancelled

### Shipment Lifecycle
```
pending → in_transit → delivered → (terminal)
  │          │
  └──────────┴──────────→ cancelled → (terminal)
```
- Created for confirmed outbound orders only.
- Carrier and tracking number can be updated while non-terminal.
- Marking delivered records the actual delivery timestamp.

### App Configuration (System Settings)
- **Single-row configuration** (app_config table, JSONB).
- **Fields**: site_name, default_warehouse_id, low_stock_threshold, default_page_size, jwt_access_ttl
- Stored as a single JSONB document with upsert semantics.
- Admin-only access via GET/PUT /api/v1/settings.
- Sensible defaults seeded by migration 000006.

### Inventory Status
```
available   → quarantine (quality hold)
available   → damaged    (inspection failure)
available   → expired    (past expiry date)
quarantine  → available  (release from hold)
quarantine  → damaged    (inspection failure)
quarantine  → expired    (past expiry date)
damaged     → available  (re-graded / repaired)
damaged     → expired    (past expiry date)
expired     → (terminal)
```

## Business Rules

1. **Inventory cannot be negative.** Before any deduction, verify sufficient available qty.
2. **Location capacity.** If location has Capacity set, total inventory qty must not exceed it.
3. **Order → Task generation.** Confirming an order generates the necessary tasks (putaway for inbound, pick+ship for outbound).
4. **FIFO/FEFO allocation.** When allocating inventory for an outbound order, prefer:
   - FEFO: earliest expiry_date first (if set)
   - FIFO: earliest received_at first (if no expiry)
5. **Task completion updates inventory.** Completing a pick task decrements inventory; completing a putaway task increments inventory.
6. **All inventory changes are audited.** Every change creates an InventoryTransaction.
7. **Shipments are for outbound orders only.** A shipment can only be created for confirmed, processing, partial, or completed outbound orders (not drafts or cancelled).
