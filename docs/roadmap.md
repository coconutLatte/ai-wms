# Evolution Roadmap

> **Auto-generated status file.** Claude Code reads this file each evolution round.
> Format: `| ID | Priority | Task | Status | Completed | Notes |`

## Phase 0: Foundation (Initial Seed)

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P0-01 | P0 | Project structure, Go module, Makefile, docker-compose | completed | 2026-07-20 | Initial seed |
| P0-02 | P0 | Domain models (Warehouse, Zone, Location, SKU, Inventory, Order, Task, User) | completed | 2026-07-20 | Initial seed |
| P0-03 | P0 | Database schema + migration 000001 | completed | 2026-07-20 | Initial seed |
| P0-04 | P0 | Auto-evolution scripts + Claude Code configs | completed | 2026-07-20 | Initial seed |
| P0-05 | P0 | Frontend scaffolds + proto definitions | completed | 2026-07-20 | Initial seed |
| P0-06 | P0 | Documentation (architecture, roadmap, domain model, ADR) | completed | 2026-07-20 | Initial seed |

## Phase 1: Core Domain Services & APIs

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P1-01 | P1 | Repository interfaces (Warehouse, Inventory, Order, Task) | completed | 2026-07-20 | Define interfaces in internal/repository/ |
| P1-02 | P1 | PostgreSQL repository implementation (Warehouse + Zone + Location) | completed | 2026-07-20 | Implement warehouse repo with pgx, 8 integration tests pass |
| P1-03 | P1 | PostgreSQL repository implementation (SKU + Inventory) | pending | — | Implement inventory repo with pgx |
| P1-04 | P1 | PostgreSQL repository implementation (Order + OrderLine) | pending | — | Implement order repo with pgx |
| P1-05 | P1 | PostgreSQL repository implementation (Task + Wave) | pending | — | Implement task repo with pgx |
| P1-06 | P1 | Warehouse service + Admin API (CRUD for warehouses, zones, locations) | pending | — | chi/v5 REST endpoints |
| P1-07 | P1 | SKU service + Admin API (CRUD for SKUs) | pending | — | chi/v5 REST endpoints |
| P1-08 | P1 | Inventory service + Admin API (query, adjust) | pending | — | With inventory transaction audit |
| P1-09 | P1 | Order service + Admin API (create/manage orders) | pending | — | Inbound + Outbound order flows |
| P1-10 | P1 | Task service + PDA API (task assignment, status flow) | pending | — | Task lifecycle management |
| P1-11 | P1 | Config management + Logger + Error handling packages | pending | — | Shared utilities in pkg/ |

## Phase 2: Admin Frontend

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P2-01 | P2 | Admin frontend scaffold (React + Ant Design + Vite + React Router) | pending | — | |
| P2-02 | P2 | Login/Auth page + API client setup | pending | — | |
| P2-03 | P2 | Warehouse management pages (list, create, edit zones/locations) | pending | — | |
| P2-04 | P2 | Inventory overview page (list, search, filter, detail) | pending | — | |
| P2-05 | P2 | Order management pages (inbound list, outbound list, create, detail) | pending | — | |
| P2-06 | P2 | Task monitoring dashboard (task list, status filter, assign) | pending | — | |
| P2-07 | P2 | Dashboard home page (KPIs, charts, alerts) | pending | — | |

## Phase 3: PDA Operations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P3-01 | P3 | PDA frontend scaffold (React + Vite + mobile-first CSS) | pending | — | |
| P3-02 | P3 | Barcode scanner component (camera-based + manual input) | pending | — | |
| P3-03 | P3 | Receiving flow (scan ASN → confirm receipt → generate putaway tasks) | pending | — | |
| P3-04 | P3 | Putaway flow (scan location → scan SKU → confirm) | pending | — | |
| P3-05 | P3 | Picking flow (wave task → scan location → scan SKU → confirm pick) | pending | — | |
| P3-06 | P3 | Cycle counting flow (count task → scan → submit → variance review) | pending | — | |
| P3-07 | P3 | Shipping flow (scan outbound order → verify → confirm ship) | pending | — | |
| P3-08 | P3 | PDA task list + offline queue + sync | pending | — | |

## Phase 4: Integrations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P4-01 | P4 | Integration adapter interface definition (common protocol) | pending | — | |
| P4-02 | P4 | WCS adapter (conveyor control, sorter interface) | pending | — | |
| P4-03 | P4 | RCS adapter (AGV/AMR task dispatch) | pending | — | |
| P4-04 | P4 | MES adapter (production order sync, material consumption) | pending | — | |
| P4-05 | P4 | ERP adapter (purchase order → inbound, sales order → outbound) | pending | — | |
| P4-06 | P4 | WebSocket real-time events (inventory changes, task updates) | pending | — | |
| P4-07 | P4 | API gateway + rate limiting + JWT auth | pending | — | |
| P4-08 | P4 | Message queue integration (NATS/Redis Streams for async tasks) | pending | — | |

## Phase 5: Advanced Features

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P5-01 | P5 | Wave planning strategies (batch, zone, carrier-based) | pending | — | |
| P5-02 | P5 | Inventory allocation strategies (FIFO, FEFO, lot-specific) | pending | — | |
| P5-03 | P5 | Multi-level inventory (warehouse → zone → location → container/LPN) | pending | — | |
| P5-04 | P5 | Report engine (inventory turnover, aging, ABC analysis) | pending | — | |
| P5-05 | P5 | RBAC permissions (role-based access control UI + enforcement) | pending | — | |
| P5-06 | P5 | Audit log viewer (operation history, traceability) | pending | — | |
| P5-07 | P5 | Inventory alerts (low stock, expiry, stranded inventory) | pending | — | |
| P5-08 | P5 | Multi-warehouse support (cross-warehouse transfers) | pending | — | |
| P5-09 | P5 | Lot traceability (raw material → WIP → finished goods) | pending | — | |
| P5-10 | P5 | Performance optimization + load testing | pending | — | |

## Evolution Metrics

| Metric | Value |
|--------|-------|
| Total tasks | 50 |
| Completed | 8 |
| In progress | 0 |
| Pending | 42 |
| Success rate | — |
| Started | 2026-07-20 |
| Last evolution | 2026-07-20 (Round 1: P1-01, P1-02) |
