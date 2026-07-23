# Evolution Roadmap

> **Strict cap: 10 pending max.** Implement rounds do NOT add tasks. DISCOVER (pending < 3) refills to 10. GROOM (pending ≥ 8, every 5 rounds) prunes excess. Current total: 91 tasks.

## Phase 0: Foundation

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P0-01 | P0 | Project structure, Go module, Makefile, docker-compose | completed | 2026-07-20 | Initial seed |
| P0-02 | P0 | Domain models (Warehouse, Zone, Location, SKU, Inventory, Order, Task, User) | completed | 2026-07-20 | Initial seed |
| P0-03 | P0 | Database schema + migration 000001 | completed | 2026-07-20 | Initial seed |
| P0-04 | P0 | Auto-evolution scripts + Claude Code configs | completed | 2026-07-20 | Initial seed |
| P0-05 | P0 | Frontend scaffolds + proto definitions | completed | 2026-07-20 | Initial seed |
| P0-06 | P0 | Documentation (architecture, roadmap, domain model, ADR) | completed | 2026-07-20 | Initial seed |

## Phase 1: Core Backend

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P1-01 | P1 | Repository interfaces | completed | 2026-07-20 | |
| P1-02 | P1 | PostgreSQL repo: Warehouse + Zone + Location | completed | 2026-07-20 | |
| P1-03 | P1 | PostgreSQL repo: SKU + Inventory | completed | 2026-07-20 | |
| P1-04 | P1 | PostgreSQL repo: Order + OrderLine + ASN | completed | 2026-07-20 | |
| P1-05 | P1 | PostgreSQL repo: Task + Wave | completed | 2026-07-20 | |
| P1-06 | P1 | PostgreSQL repo: ASN lines | completed | 2026-07-20 | |
| P1-07 | P1 | PostgreSQL repo: User + Role + AuditLog | completed | 2026-07-20 | |
| P1-08 | P1 | HTTP middleware stack | completed | 2026-07-20 | Request ID, logging, recovery, CORS |
| P1-09 | P1 | Config + Logger integration | completed | 2026-07-20 | |
| P1-10 | P1 | Error handling (RFC 7807) | completed | 2026-07-20 | |
| P1-11 | P1 | Warehouse service + Admin API | completed | 2026-07-20 | |
| P1-12 | P1 | SKU service + Admin API | completed | 2026-07-20 | |
| P1-13 | P1 | Inventory service + Admin API | completed | 2026-07-20 | |
| P1-14 | P1 | Order service + Admin API | completed | 2026-07-20 | |
| P1-15 | P1 | Task service + PDA API | completed | 2026-07-20 | |

## Phase 2: Cross-Cutting & Frontend

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P2-01 | P1 | DB transaction support | completed | 2026-07-20 | |
| P2-02 | P1 | Pagination metadata | completed | 2026-07-20 | |
| P2-03 | P1 | Domain unit tests | completed | 2026-07-20 | |
| P2-04 | P1 | JWT auth (login, refresh, middleware) | completed | 2026-07-20 | |
| P2-05 | P1 | Makefile dev targets | completed | 2026-07-20 | |
| P2-06 | P1 | Seed data script | completed | 2026-07-20 | |
| P2-07 | P2 | Admin frontend scaffold | completed | 2026-07-20 | React + Ant Design + routing + API client |
| P2-08 | P2 | Admin: Warehouse pages | completed | 2026-07-20 | |
| P2-09 | P2 | FEFO/FIFO queries | completed | 2026-07-20 | |
| P2-10 | P2 | PDA frontend scaffold | completed | 2026-07-20 | Mobile layout + scanner + task list |
| P2-11 | P2 | Tx-aware repo helpers | completed | 2026-07-20 | |
| P2-12 | P2 | SELECT FOR UPDATE locking | completed | 2026-07-20 | |
| P2-13 | P2 | DI wiring for TxManager | completed | 2026-07-21 | |
| P2-14 | P2 | CountWaves + CountRoles | completed | 2026-07-21 | |
| P2-15 | P2 | AuditLog list endpoint | completed | 2026-07-21 | |
| P2-16 | P2 | User list endpoint | completed | 2026-07-21 | |
| P2-17 | P2 | Location state machine | completed | 2026-07-21 | |
| P2-18 | P2 | OrderLine + ASN status ops | completed | 2026-07-21 | |
| P2-19 | P2 | Inventory status transitions | completed | 2026-07-21 | |
| P2-20 | P1 | Seed: default admin user | completed | 2026-07-21 | |

## Phase 3: Next Up (max 10 pending)

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P3-01 | P1 | Admin: ProtectedRoute + Login page | completed | 2026-07-21 | Auth guard + login form with JWT; blocks all admin UI usage |
| P3-02 | P1 | Role-based authorization middleware | completed | 2026-07-21 | RequireRole middleware checks JWT role_names; admin-only /audit-logs and /users routes protected |
| P3-03 | P2 | Admin: SKU management pages | completed | 2026-07-21 | Full CRUD: list/search, create/edit modals with UOM and attributes editor |
| P3-04 | P2 | Admin: Inventory dashboard | completed | 2026-07-21 | Summary cards, low stock, warehouse breakdown; GET /api/v1/inventory/dashboard |
| P3-05 | P2 | Admin: Order management pages | completed | 2026-07-21 | Table with status badges, detail drawer, status transitions |
| P3-06 | P1 | Token blacklist / logout | completed | 2026-07-21 | Refresh token revocation DB-backed; logout endpoint; blacklist check on refresh |
| P3-07 | P2 | Health check endpoints (/health, /ready) | completed | 2026-07-21 | `/health` exists in both admin + PDA servers; no DB ping readiness check yet |
| P3-08 | P2 | PDA: Login + task list screen | completed | 2026-07-21 | Real JWT login via POST /api/v1/auth/login, auth middleware on PDA server, protected routes, React Query task fetching |
| P3-09 | P1 | i18n: Chinese (default) + English for Admin + PDA | completed | 2026-07-21 | react-i18next + i18next-browser-languagedetector, zh-CN default + en, language switcher in Admin header + PDA header, all page text + form labels + Ant Design locale translated |
| P3-10 | P2 | GitHub Pages: deploy PDA demo alongside admin | completed | 2026-07-21 | Build PDA to docs/pda/, cross-links admin ↔ PDA in headers |
| P3-11 | P2 | Redis client bootstrap | completed | 2026-07-21 | go-redis/v9 client, config-driven bootstrap, wired into admin + PDA entry points |
| P3-12 | P2 | Migration tracking table | completed | 2026-07-22 | schema_migrations table + runner; each .sql runs once via RunMigrationsFromDir |
| P3-13 | P1 | WMS standard benchmark / test scenarios | completed | 2026-07-22 | Benchmark tests (~12 benchmarks for state machines, inventory ops, FEFO/FIFO sorting, allocation at scale, concurrent transition), scenario tests (10 scenarios: inbound flow, outbound FEFO, shipment, cycle count, accuracy KPIs ≥99.5% inventory / ≥99.9% pick, concurrent adjust, concurrent reserve, allocate+pick race, partial fulfillment, quarantine, multi-order wave, transfer) |
| P3-14 | P1 | Wave service + Admin API | completed | 2026-07-22 | WaveService (CreateWave, GetWave, ListWaves, UpdateWaveStatus, ReleaseWave, AddWaveOrders, RemoveWaveOrders), WaveHandler with 7 REST endpoints, 13 unit tests |
| P3-15 | P1 | Extend UserService + Admin API: Create, Get, Update, UpdateStatus | completed | 2026-07-22 | UserService (CreateUser, GetUser, UpdateUser, UpdateUserStatus with state machine validation), UserHandler with 4 new REST endpoints, domain CanTransitionTo, 15 unit tests |
| P3-16 | P1 | ASN API endpoints via OrderService | completed | 2026-07-22 | Implemented CreateASN, GetASN, ListASNs service methods; added ASNFilter + ListASNs/CountASNs to repository; POST/GET /api/v1/asns, GET /api/v1/asns/{id}, PUT /api/v1/asns/{id}/status endpoints via OrderHandler |
| P3-17 | P1 | REST gap fill: inventory status, order line status, location barcode, SKU barcode | completed | 2026-07-22 | Added PATCH /inventory/{id}/status, PUT /orders/{id}/lines/{lineId}/status, GET /locations?barcode=X, GET /skus?code=X; added GetLocationByBarcode service method + 12 handler tests |
| P3-18 | P2 | Admin: Task management page | completed | 2026-07-22 | Replaced placeholder with full task list (table filtered by task_type/status/assigned_to), detail drawer with order/line refs, assignment modal, status transitions, completion modal, stat cards; fixed Task type definitions to match API shape; added comprehensive i18n translations (zh-CN + en) |
| P3-19 | P2 | Admin: Order create form | completed | 2026-07-22 | Modal form with order type selector, warehouse picker, priority, external ref, line items editor with SKU + qty + UOM; POST /orders API already wired; i18n EN+ZH |
| P3-20 | P1 | Order→Task generation on confirm | completed | 2026-07-22 | Auto-generates putaway tasks for inbound/transfer/return orders and pick tasks for outbound orders on confirm. Deduplicates: skips if tasks already exist. TaskFields carry order_id, order_line_id, SKU, qty, batch, priority. 5 new unit tests. |
| P3-21 | P1 | Task completion inventory effects + transaction audit | completed | 2026-07-22 | Putaway creates/increments inventory + receipt tx; pick decrements inventory + pick tx; replenish transfers between locations; all within TxManager for atomicity; respects "no negative qty" rule; 9 new unit tests |
| P3-22 | P1 | Inventory reservation for order allocation | completed | 2026-07-23 | ReserveInventory/UnreserveInventory with FEFO/FIFO strategy, TxManager atomic, transaction audit trail; 15 unit tests |
| P3-23 | P1 | Role management API (service + handler) | completed | 2026-07-23 | RoleService (ListRoles, GetRole, CreateRole, UpdateRole, DeleteRole) + RoleHandler with 5 REST endpoints under /api/v1/roles (admin-only). Added DeleteRole to repo interface + postgres impl. 13 service unit tests |
| P3-24 | P2 | Admin: Dashboard real data | completed | 2026-07-23 | New GET /api/v1/dashboard endpoint (DashboardHandler) aggregates warehouse_count, sku_count, inventory_stats, order_summary (by status), task_summary (by status). Added CountWarehouses/CountSKUs service methods, CountOrdersByStatus/CountTasksByStatus repo+service methods. Frontend Dashboard.tsx fetches real data with loading/error states and displays stat cards + order/task breakdowns |
| P3-25 | P2 | Admin: User & Role management pages | completed | 2026-07-23 | User list with status badges/filters, create/edit modal with role assignment, status transitions; Role list with permission config matrix (resource+actions), create/edit/delete modals, detail view; sidebar nav entries + i18n (zh-CN/en) |
| P3-26 | P2 | PDA: Camera barcode scanning integration | completed | 2026-07-23 | Wired @zxing/library into BarcodeScanner for camera scanning (auto-detect 1D/2D barcodes, rear-cam preference, scan-frame overlay); fixed Login double /api/v1 prefix bug; added barcode lookup API (location→putaway, SKU→inventory info); registered WarehouseHandler + SKUHandler on PDA service for barcode endpoints |
| P3-27 | P2 | Admin: ASN management pages | completed | 2026-07-23 | ASN list table with status badges and filters; create ASN modal (carrier, tracking, expected_at, line items); ASN detail drawer with line items; status transitions (pending→arrived→receiving→received). ASN API already exists from P3-16 |
| P3-28 | P1 | Admin: Location & Zone management pages | completed | 2026-07-23 | Zones page with warehouse filter, create/edit modal; Locations page with zone/warehouse filter, create/edit modal, status transitions; new API endpoints GET /api/v1/zones, PUT /api/v1/zones/{id}, PUT /api/v1/locations/{id} |
| P3-29 | P1 | PDA: Receiving workflow page | completed | 2026-07-23 | New PDA `/receive` page: scan ASN barcode → fetch ASN → view status + line items → receive each line with qty validation → confirm receipt. Auto-generates putaway tasks. ASN endpoints + ReceiveASNLine API wired to PDA service. i18n zh-CN+en |
| P3-30 | P2 | Admin: Wave management pages | completed | 2026-07-23 | Wave list table with type/status filters; create wave modal (name, type, warehouse); wave detail with order list sidebar; add/remove orders (only in created status); release wave button; status transitions. Wave API already exists from P3-14 |
| P3-31 | P2 | API handler integration tests | completed | 2026-07-23 | 29 integration tests across all handler groups (warehouse, zone, location, SKU, inventory, order, ASN, task, wave, user, role, dashboard, audit log). End-to-end through httptest.Server with in-memory mock repos. Build tag `integration` separates from unit tests. Covers CRUD, validation, pagination, not-found, state machines. |
| P3-32 | P1 | Repository integration tests | completed | 2026-07-23 | 34 new integration tests across all repos: warehouse (10), inventory (10), order (3), task (5), user (2), token_blacklist (4 new file). Total 146 repo tests covering ~90% of repository interface methods |
| P3-33 | P2 | Inventory transaction history API + Admin page | completed | 2026-07-23 | GET /api/v1/inventory-transactions with type/SKU/warehouse/date filters; Admin page with type badges, delta_qty colors, date range picker, detail modal. i18n zh-CN+en |
| P3-34 | P2 | PDA: Order lookup / detail page | completed | 2026-07-23 | New PDA page: scan/type order number → fetch order with lines → display status badge, type, priority, line item table (SKU, qty, fulfilled qty). i18n zh-CN+en. Added order_no filter to backend, registered order routes on PDA server |
| P3-35 | P2 | Admin: Audit log viewer page | completed | 2026-07-23 | Table with filters (action, user, resource, date range). Detail modal with old/new values JSON. Added date_from/date_to to backend AuditLogFilter. i18n zh-CN+en |
| P3-36 | P2 | Redis caching for hot inventory queries | cancelled | — | **Why cancelled:** Premature optimization. No performance bottleneck exists. Add caching when there's a measured problem, not before. |
| P3-37 | P2 | gRPC server bootstrap + TaskService | cancelled | — | **Why cancelled:** Speculative infrastructure. No downstream integrations (WCS/RCS/MES/ERP) exist yet. Bootstrap gRPC when there's something to integrate with. |
| P3-38 | P1 | PDA: Putaway workflow page | completed | 2026-07-23 | New PDA `/putaway` page: scan location barcode → select pending putaway tasks → confirm with qty validation → complete task with inventory effects. Complements Receiving (P3-29) and Scanning (P3-26). i18n zh-CN+en. |
| P3-39 | P1 | PDA: Picking workflow page | completed | 2026-07-23 | New PDA `/pick` page: multi-step verification (scan location → scan SKU → pick qty → complete), barcode scanning, qty validation, mismatch warnings. Route + i18n zh-CN+en. |
| P3-40 | P1 | Order cancellation: release reservations + cancel tasks | completed | 2026-07-23 | PUT /api/v1/orders/{id}/cancel — OrderService.CancelOrder: validates state transition → cancels non-terminal tasks → releases inventory reservations → cancels order lines → updates order status, all within TxManager. 10 unit tests. |
| P3-41 | P2 | PDA: Stock inquiry page | pending | — | New PDA page: scan location/SKU barcode → display current inventory levels (qty, available, reserved, batch_no, status). Inventory API exists from P1-13. i18n zh-CN+en. |
| P3-42 | P2 | PDA: Cycle count workflow page | pending | — | New PDA page: scan location → enter counted quantities per SKU/batch → submit variance report → approve/adjust if within tolerance. Requires new cycle count API endpoints (start count, submit count lines, finalize). i18n zh-CN+en. |
| P3-43 | P1 | Cycle count backend: domain model + service + API | pending | — | Domain: CycleCount/CycleCountLine entities with state machine (in_progress→pending_review→approved/adjusted). Service: StartCycleCount, SubmitLine, FinalizeCount, ApproveVariance. API: POST /cycle-counts, POST /cycle-counts/{id}/lines, POST /cycle-counts/{id}/finalize, PUT /cycle-counts/{id}/approve. Migration for cycle_counts + cycle_count_lines tables. Enables P3-42. |
| P3-44 | P1 | Shipment domain model + backend: entity, service, API, migration | pending | — | Domain: Shipment entity (tracking_no, carrier, carrier_service, status, estimated_delivery). Service: CreateShipment, GetShipment, ListShipments, UpdateStatus, AddTracking. API endpoints under /api/v1/shipments. Migration for shipments table. Completes outbound data model. |
| P3-45 | P1 | PDA: Ship confirmation workflow page | pending | — | New PDA page: scan/select confirmed outbound order → verify picked line items → enter carrier + tracking number → confirm shipment. Calls shipment API from P3-44. Completes the outbound flow (pick → ship). i18n zh-CN+en. |
| P3-46 | P1 | Dedicated ASN service: extract from OrderService | cancelled | — | **Why cancelled:** Pure internal refactoring. ASN works correctly under OrderService. No user-visible benefit, no new capability unlocked. Extract only when ASN complexity grows enough to justify the split. |
| P3-47 | P2 | Admin: Shipment management page | pending | — | Table listing shipments with status/carrier/tracking filters; detail drawer with order refs and item list; create shipment from confirmed outbound orders; status transitions (pending→in_transit→delivered). i18n zh-CN+en. |
| P3-48 | P2 | Health check /ready endpoint with DB ping | pending | — | Extends /health with GET /ready: pings PostgreSQL (pgxpool.Ping), pings Redis (PING command), returns 200 with per-service status or 503 if any dependency is unhealthy. Enables Kubernetes readiness probes. |
| P3-49 | P2 | Frontend API module extraction: dedicated api/*.ts per resource | cancelled | — | **Why cancelled:** Pure internal refactoring. No user-visible benefit. Extract incrementally as pages are modified rather than as a standalone task. |
| P3-50 | P2 | Admin: System configuration page | pending | — | New page at /settings: display/edit app-level settings (site name, default warehouse, low-stock threshold, default pagination size, JWT TTL). Store in app_config JSON column. Admin-only access. i18n zh-CN+en. |

<!-- GROOM round 85 on 2026-07-23: pruned from 10→8 pending. Cancelled P3-46 (ASN service refactor) and P3-49 (API module extraction) — both pure internal refactoring with no user-visible benefit. Remaining 8 tasks: cycle count (backend + PDA), shipment (backend + PDA + Admin), health check, stock inquiry, system config. -->
