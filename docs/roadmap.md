# Evolution Roadmap

> **Strict cap: 10 pending max.** Implement rounds do NOT add tasks. DISCOVER (pending < 3) refills to 10. GROOM (pending ≥ 8, every 5 rounds) prunes excess. Current total: 107 tasks.

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
| P3-41 | P2 | PDA: Stock inquiry page | completed | 2026-07-23 | New PDA page at /stock-inquiry: scan location/SKU barcode → resolved via GET /api/v1/stock-inquiry?barcode=X → displays entity info + inventory cards (qty, reserved, available, batch_no, status) + aggregated totals. Link from Scan hub's "Locate" card. i18n zh-CN+en. Backend handler uses WarehouseService + SKUService + InventoryService. |
| P3-42 | P2 | PDA: Cycle count workflow page | completed | 2026-07-24 | New PDA page at /cycle-count: scan location → enter counted quantities per SKU/batch → submit variance report for supervisor review. Three-step workflow (scan→count→submit). Link from Scan hub's "Cycle Count" card. i18n zh-CN+en. |
| P3-43 | P1 | Cycle count backend: domain model + service + API | completed | 2026-07-24 | Domain: CycleCount (state machine: draft→in_progress→pending_review→approved/adjusted/cancelled) + CycleCountLine entities. Service: StartCycleCount (creates lines from current inventory), SubmitLine, FinalizeCount, ApproveCount (applies inventory adjustments atomically), CancelCycleCount. API: POST/GET /cycle-counts, POST /cycle-counts/{id}/lines, POST /cycle-counts/{id}/finalize, PUT /cycle-counts/{id}/approve, PUT /cycle-counts/{id}/cancel. Migration 000004: cycle_counts + cycle_count_lines tables. Wired into both Admin and PDA servers. |
| P3-44 | P1 | Shipment domain model + backend: entity, service, API, migration | completed | 2026-07-24 | Domain: Shipment entity (state machine: pending→in_transit→delivered/cancelled). Service: CreateShipment (validates order is outbound + confirmed), GetShipment, ListShipments, UpdateStatus, UpdateTracking, DeliverShipment. API: POST/GET /shipments, PUT /{id}/status, PUT /{id}/tracking, PUT /{id}/deliver. Migration 000005: shipments table. Wired into both Admin and PDA servers. |
| P3-45 | P1 | PDA: Ship confirmation workflow page | completed | 2026-07-24 | New PDA page: scan/select confirmed outbound order → verify picked line items → enter carrier + tracking number → confirm shipment via POST /api/v1/shipments. Completed outbound flow: pick (P3-39) → ship (P3-45). Added to Scan hub nav. i18n zh-CN+en. |
| P3-46 | P1 | Dedicated ASN service: extract from OrderService | cancelled | — | **Why cancelled:** Pure internal refactoring. ASN works correctly under OrderService. No user-visible benefit, no new capability unlocked. Extract only when ASN complexity grows enough to justify the split. |
| P3-47 | P2 | Admin: Shipment management page | completed | 2026-07-24 | Table with status/carrier/tracking filters; detail drawer with order refs; create shipment from confirmed outbound orders; status transitions (pending→in_transit→delivered); update tracking modal. i18n zh-CN+en. |
| P3-48 | P2 | Health check /ready endpoint with DB ping | completed | 2026-07-24 | GET /ready with PostgreSQL + Redis ping; per-service status; 200 ok / 503 degraded. Enables Kubernetes readiness probes. |
| P3-49 | P2 | Frontend API module extraction: dedicated api/*.ts per resource | cancelled | — | **Why cancelled:** Pure internal refactoring. No user-visible benefit. Extract incrementally as pages are modified rather than as a standalone task. |
| P3-50 | P2 | Admin: System configuration page | completed | 2026-07-24 | Settings page at /settings: site name, default warehouse, low-stock threshold, pagination size, JWT TTL. Stored in app_config JSONB. Admin-only. i18n zh-CN+en. |
| P3-51 | P1 | PDA: Replenishment workflow page | completed | 2026-07-24 | New PDA page at /replenish: scan pick location → view inventory → select SKU → system finds reserve stock → confirm source/destination/qty → create+start+complete replenish task (atomic inventory transfer). Added to Scan hub (🔄 card) + i18n zh-CN+en (35 new keys). Uses existing task/inventory endpoints — no backend changes needed. |
| P3-52 | P1 | PDA: Transfer workflow page | completed | 2026-07-24 | New PDA page at /transfer: scan source location → scan destination → enter SKU/batch/qty → confirm → complete transfer task. Added to Scan hub (📦 card) + i18n zh-CN+en (35 keys). Uses existing transfer task type + backend endpoints. |
| P3-53 | P2 | Docker multi-stage builds for admin + PDA services | completed | 2026-07-24 | Multi-stage Dockerfiles (golang:1.25-alpine → alpine:3.20) for admin + PDA. Added services to docker-compose.yml with depends_on, health checks, env vars. Added .dockerignore. Enhanced Makefile with docker-up/docker-build-admin/docker-build-pda targets. Both images ~10 MB. |
| P3-54 | P2 | GitHub Actions CI pipeline for backend | completed | 2026-07-24 | CI workflow with build, test, vet, lint, Docker verify; badge in README |
| P3-55 | P2 | Frontend test infrastructure: Vitest + React Testing Library | completed | 2026-07-24 | Installed vitest, @testing-library/react, @testing-library/jest-dom, jsdom in both frontend/admin and frontend/pda. Created vitest.config.ts with jsdom environment, globals, @/ path alias, and setup files. Wrote 8 smoke test files (17 tests total): Admin — NotFound, Login, Dashboard, Settings; PDA — NotFound, Login, Profile, Tasks. Added test/test:run scripts to package.json. All pass. |
| P3-56 | P2 | API documentation with Swagger UI | completed | 2026-07-24 | Embedded OpenAPI 3.0 spec + Swagger UI at GET /api/docs on both admin and PDA servers. Covers all 18 resource groups across 70+ endpoints. No external Go dependencies. |
| P3-57 | P1 | Seed data: operational demo data (orders, ASNs, tasks, waves, cycle counts, shipments, non-admin users) | pending | — | Extend cmd/seed to create realistic demo data: 5-8 orders (inbound + outbound across statuses), 3-5 ASNs with lines, 10-15 tasks (putaway/pick/replenish across statuses), 2-3 waves with orders, 1-2 cycle counts, 1-2 shipments, and 2-3 non-admin users (operator + picker). Current seed only creates warehouses/zones/locations/SKUs/inventory + admin user. Unblocks full frontend demo and testing of all workflows. |
| P3-58 | P2 | PDA MSW demo mocks + ESLint config for both apps | pending | — | Admin app already has MSW demo mocks for GitHub Pages; PDA has none — its demo at /pwa/ makes real API calls that 404. Add MSW to PDA devDependencies + src/mocks/handlers.ts with mock APIs (auth, tasks, ASNs, orders, cycle counts, stock inquiry, shipments, barcode lookups). Also add eslint.config.mjs (flat config, TypeScript + React plugins) to both admin + PDA — lint scripts exist in package.json but no ESLint config file, so `npm run lint` fails. |
| P3-59 | P1 | Service tests: CycleCount + Shipment | pending | — | `cycle_count.go` and `shipment.go` are the only two service files without `_test.go`. Every other service (warehouse, SKU, inventory, order, task, user, role, wave, auth, audit_log, app_config) has unit tests. Add CycleCountServiceTest (~8 tests: StartCycleCount with/without location filter, SubmitLine, FinalizeCount, ApproveCount, CancelCycleCount, state machine guards) + ShipmentServiceTest (~8 tests: CreateShipment, UpdateStatus transitions, UpdateTracking, DeliverShipment, validation errors). |
| P3-60 | P2 | Repo integration tests: CycleCount + Shipment repos | pending | — | `postgres/cycle_count_repo.go` and `postgres/shipment_repo.go` lack `_test.go` files. Every other postgres repo has integration tests. Add cycle_count_repo_test.go (~6 tests: Create, GetByID, List/filter, CreateLines, UpdateLine, UpdateStatus) + shipment_repo_test.go (~6 tests: Create, GetByID, List/filter by warehouse/order/status/carrier, UpdateStatus, UpdateTracking). |
| P3-61 | P2 | Domain tests: AppConfig entity | pending | — | `app_config.go` is the only domain entity file without a `_test.go`. Every other domain entity (warehouse, inventory, order, task, user, cycle_count, shipment, schema_migration) has tests. Add app_config_test.go (~5 tests: NewAppConfig defaults, GetValue/SetValue, JSONB marshal/unmarshal, key validation). |
| P3-62 | P2 | PDA component tests: key workflow pages | pending | — | PDA pages have zero component tests except Login, NotFound, Profile, Tasks (P3-55). Add smoke + interaction tests for Receiving (scan ASN → receive line), Picking (scan location → pick SKU), Putaway (scan location → complete putaway task), and Transfer (scan source/dest → complete). ~12 tests total across 4 pages. Mock API responses with existing api/ modules. |
| P3-63 | P2 | Admin component tests: key management pages | pending | — | Admin pages have zero component tests except Login, NotFound, Dashboard, Settings (P3-55). Add smoke + interaction tests for Orders (list, create, detail, cancel), Inventory (list, status transition, adjust), and Tasks (list, assign, complete). ~12 tests total across 3 pages. Mock API responses using existing MSW handlers. |
| P3-64 | P2 | E2E test scaffolding: Playwright + critical flow smoke tests | pending | — | No end-to-end tests exist. Scaffold Playwright in a new e2e/ directory with playwright.config.ts (base URL, screenshot-on-failure, CI reporter). Smoke test 3 critical flows against a running dev server: (a) inbound: login → create ASN → receive → putaway, (b) outbound: create order → confirm → pick → ship, (c) inventory: cycle count → approve adjustment. Add `make test-e2e` target. |
| P3-65 | P2 | Redis client unit test expansion | pending | — | Current coverage in `pkg/redis` is 41.7%. Expand to >80% with tests for: connection failure (wrong port/addr), ping timeout handling, Close idempotency, nil client safety checks, Redis unavailable → graceful degradation in /ready endpoint. Use miniredis (github.com/alicebob/miniredis/v2) for isolated unit tests — no real Redis needed. |
| P3-66 | P2 | Deployment guide: environment variables, config reference, runbook | pending | — | README is minimal (project overview only). Add PRODUCTION.md with: all env vars and defaults (DB_*, REDIS_*, JWT_*, SERVER_*), PostgreSQL setup (users, extensions), Redis setup, docker-compose for production (profiles, resource limits), health check usage (/health vs /ready), backup/restore basics (pg_dump/pg_restore), and a quick-start runbook (clone → env → migrate → seed → up). |

<!-- DISCOVER round 87 on 2026-07-24: refilled from 2→10 pending. Focused on test coverage gaps: missing service tests (cycle_count, shipment), missing repo tests, missing domain test (app_config), missing component tests for PDA workflow pages and admin management pages, E2E scaffolding with Playwright, Redis client tests (41.7%→80%+), and deployment documentation. All tasks directly address measurable quality gaps in the existing codebase — no speculative features. -->
