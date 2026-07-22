# Evolution Roadmap

> **Strict cap: 10 pending max.** Implement rounds do NOT add tasks. DISCOVER (pending < 3) refills to 10. GROOM (pending ≥ 8, every 5 rounds) prunes excess.

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
| P3-23 | P1 | Role management API (service + handler) | pending | — | Create RoleService (ListRoles, GetRole, CreateRole, UpdateRole, DeleteRole) and RoleHandler with REST endpoints under /api/v1/roles. Roles currently only seedable — this makes them manageable via admin UI. Needed by P3-25 for role CRUD in frontend |
| P3-24 | P2 | Admin: Dashboard real data | pending | — | Wire inventory dashboard API (GET /inventory/dashboard) for summary stat cards; add order summary (counts by status) and task summary (counts by status) API calls. Replace all "—" placeholders with live data. Dashboard is the landing page and currently shows no real data |
| P3-25 | P2 | Admin: User & Role management pages | pending | — | User table (list, status badges, create/edit modal with role assignment); Role list with permission config (resource+actions matrix); depends on P3-23 (Role API) and P3-15 (User API) |
| P3-26 | P2 | PDA: Camera barcode scanning integration | pending | — | Wire @zxing/library (already installed) into BarcodeScanner component for real camera scanning; fix Login double /api/v1 prefix bug; add barcode lookup API call with result navigation (location→putaway task, SKU→inventory info) |
| P3-27 | P2 | Admin: ASN management pages | pending | — | ASN list table with status badges and filters; create ASN modal (carrier, tracking, expected_at, line items); ASN detail drawer with line items; status transitions (pending→arrived→receiving→received). ASN API already exists from P3-16 |
| P3-28 | P2 | Admin: Wave management pages | pending | — | Wave list table with type/status filters; create wave modal (name, type, warehouse); wave detail with order list sidebar; add/remove orders (only in created status); release wave button; status transitions. Wave API already exists from P3-14 |
| P3-29 | P2 | API handler integration tests | pending | — | Add integration tests using httptest.Server with real service+repo layers. Cover warehouse CRUD, inventory query+adjust, order lifecycle, task lifecycle, wave operations, and auth flow. Target 40%+ handler coverage (up from 9.3%). Use build tags to separate from unit tests |

<!-- GROOMED round 60 on 2026-07-22: re-sorted by true priority (P1→P2), 10 pending, 0 cancelled. -->
