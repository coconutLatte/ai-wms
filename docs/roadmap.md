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
| P3-13 | P1 | WMS standard benchmark / test scenarios | pending | — | Core flow tests: ASN→收货→上架, 波次→分配→拣货→复核, 出库→称重→发货, 盘点→差异→调整; accuracy KPIs (inventory ≥99.5%, pick ≥99.9%); race conditions (concurrent adjust, allocate+pick same SKU) |
| P3-14 | P1 | Wave service + Admin API | pending | — | Wave repo supports full CRUD (CreateWave, ListWaves, GetWave, UpdateWaveStatus, AddWaveOrders, RemoveWaveOrders); needs WaveService layer + REST handlers; proto already defines CreateWave/ReleaseWave/GetWave |
| P3-15 | P1 | Extend UserService + Admin API: Create, Get, Update, UpdateStatus | pending | — | UserRepo has full CRUD but UserService only exposes ListUsers; add remaining service methods + POST/PUT /users, GET /users/{id}, PUT /users/{id}/status routes |
| P3-16 | P1 | ASN API endpoints via OrderService | pending | — | OrderRepo supports ASN CRUD (CreateASN, GetASN, CreateASNLine, UpdateASNStatus); OrderService has status operations; add ASN HTTP handler + routes: POST/GET /asns, GET /asns/{id}, PUT /asns/{id}/status, POST /asns/{id}/lines; requires ListASNs repo method |
| P3-17 | P1 | REST gap fill: inventory status, order line status, location barcode, SKU barcode | pending | — | 4 small isolated gaps with existing repo support: PATCH /inventory/{id}/status, PUT /orders/{id}/lines/{lineId}/status, GET /locations?barcode=X, GET /skus?code=X; add service methods + handler routes |
| P3-18 | P2 | Admin: Task management page | pending | — | Replace placeholder with real task list (table filtered by type/status/worker), task detail drawer with order/line refs, assignment dropdown, status transitions |
| P3-19 | P2 | Admin: Order create form | pending | — | Modal form: order type selector (inbound/outbound/transfer/return), warehouse picker, line items editor with SKU search + qty; POST /orders API already wired |
| P3-20 | P2 | Admin: User & Role management pages | pending | — | User table (list, status badges, create/edit modal with role assignment); Role list with permission config (resource+actions matrix); depends on P3-15 API |
| P3-21 | P2 | PDA: Camera barcode scanning integration | pending | — | Wire @zxing/library (already installed) into BarcodeScanner component for real camera scanning; add barcode lookup API call with result navigation (location→putaway task, SKU→inventory info) |

<!-- DISCOVER refilled to 10 pending on 2026-07-22. -->
