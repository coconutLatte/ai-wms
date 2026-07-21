# Evolution Roadmap

> **Lean mode**: Max 10 pending tasks. When < 3 remain, DISCOVER refills to ~10.
> Format: `| ID | Priority | Task | Status | Completed | Notes |`

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
| P1-01 | P1 | Repository interfaces | completed | 2026-07-20 | 4 interfaces: Warehouse, Inventory, Order, Task |
| P1-02 | P1 | PostgreSQL repo: Warehouse + Zone + Location | completed | 2026-07-20 | 8 integration tests |
| P1-03 | P1 | PostgreSQL repo: SKU + Inventory | completed | 2026-07-20 | 13 integration tests |
| P1-04 | P1 | PostgreSQL repo: Order + OrderLine + ASN | completed | 2026-07-20 | 15 integration tests |
| P1-05 | P1 | PostgreSQL repo: Task + Wave | completed | 2026-07-20 | 24 integration tests |
| P1-06 | P1 | PostgreSQL repo: ASN lines | completed | 2026-07-20 | 7 integration tests |
| P1-07 | P1 | PostgreSQL repo: User + Role + AuditLog | completed | 2026-07-20 | 19 integration tests |
| P1-08 | P1 | HTTP middleware stack | completed | 2026-07-20 | Request ID, logging, recovery, CORS |
| P1-09 | P1 | Config + Logger integration | completed | 2026-07-20 | Env-based config, structured logging |
| P1-10 | P1 | Error handling (RFC 7807, validation, domain errors) | completed | 2026-07-20 | 125 tests |
| P1-11 | P1 | Warehouse service + Admin API | completed | 2026-07-20 | CRUD warehouses, zones, locations |
| P1-12 | P1 | SKU service + Admin API | completed | 2026-07-20 | CRUD with validation |
| P1-13 | P1 | Inventory service + Admin API | completed | 2026-07-20 | Query, adjust, transactions |
| P1-14 | P1 | Order service + Admin API | completed | 2026-07-20 | Status machine, line management |
| P1-15 | P1 | Task service + PDA API | completed | 2026-07-20 | Assignment, status flow, PDA endpoints |

## Phase 2: Cross-Cutting & Frontend Foundation

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P2-01 | P1 | DB transaction support for atomic inventory operations | completed | 2026-07-20 | TxManager interface + pgx impl; inventory qty+txn are now atomic via WithTx; 5 integration tests |
| P2-02 | P1 | Pagination metadata for all list endpoints | completed | 2026-07-20 | `ListResponse[T]` generic envelope with total/page/page_size/total_pages; Count* methods on all repos; 8 endpoints updated |
| P2-03 | P1 | Domain unit tests (state machines, business rules) | completed | 2026-07-20 | 88 tests across 4 files: Order/Task/Wave/ASN/OrderLine state machines, Inventory biz rules (CanDeduct, CanReserve, FEFO/FIFO helpers), Permission.Can, struct validation; moved state transition logic from service → domain (proper DDD) |
| P2-04 | P1 | Authentication (JWT login, token refresh, middleware) | completed | 2026-07-20 | AuthService with bcrypt passwords, JWT HS256 access (15min) + refresh (7d) tokens, Auth middleware with OptionalAuth variant, context helpers (GetUserID/GetUsername/GetUserRoleIDs), POST /api/v1/auth/login + refresh + GET /me endpoints, UpdateLastLogin repo method; 17 unit tests (service + middleware) |
| P2-05 | P1 | Makefile: run-admin, run-pda, migrate targets | completed | 2026-07-20 | `make run-admin` (port 8080), `make run-pda` (port 8081) via go run; `make migrate` applies all SQL files via docker exec psql; db-migrate now aliases migrate |
| P2-06 | P1 | Seed data script (demo warehouse, zones, SKUs) | completed | 2026-07-20 | `backend/cmd/seed/main.go`: idempotent Go binary seeding 1 warehouse, 7 zones, 35 locations, 16 SKUs across 6 categories, 25 inventory records with batch tracking. Also fixes admin password to real bcrypt hash. `make seed` target added. |
| P2-07 | P2 | Admin frontend scaffold (React + Ant Design + routing) | completed | 2026-07-20 | React 18 + Ant Design 5 + react-router-dom v6; AdminLayout with collapsible sidebar + header with user dropdown; all module routes (dashboard, warehouses, SKUs, inventory, orders, tasks); axios client with JWT auth interceptor + auto-refresh on 401 with request queuing; zustand auth store with localStorage persistence; Ant Design ConfigProvider theme tokens; shared API types matching backend ListResponse[T] envelope; Login + 404 placeholder pages; global CSS with layout overrides |
| P2-08 | P2 | Admin: Warehouse management pages (list, create, edit) | completed | 2026-07-20 | Full CRUD UI: warehouse list/create/edit, zone list/create, location list/create; nested navigation with breadcrumbs; paginated tables with status/type tag colors; also fixed Zone.type→zone_type and added missing Location fields in TS types |
| P2-09 | P2 | FEFO/FIFO inventory retrieval queries | completed | 2026-07-20 | GetOldestInventory (FIFO) and GetExpiringInventory (FEFO) repo/service/API endpoints with ORDER BY received_at/expiry_date; only returns available+non-zero stock; warehouse/sku/limit filters; GET /api/v1/inventory/fifo and GET /api/v1/inventory/fefo |
| P2-10 | P2 | PDA frontend scaffold (React + mobile-first) | completed | 2026-07-20 | Mobile layout, bottom tab nav, barcode scanner component, task list/detail pages, login + profile + scan pages, axios client with JWT refresh, zustand auth store, antd-mobile-ready CSS; 10 source files scaffolded |
| P2-11 | P2 | Tx-aware helpers for remaining repos (warehouse, order, task, user) | completed | 2026-07-20 | Added exec/query/queryRow tx dispatch helpers to WarehouseRepo, OrderRepo, TaskRepo, UserRepo; all repos now participate in TxManager.WithTx transactions |
| P2-12 | P2 | SELECT FOR UPDATE row-level locking for inventory adjustments | completed | 2026-07-20 | Added GetAndLockInventory (SELECT ... FOR UPDATE) to repo interface+postgres; moved read+validate into tx boundary in AdjustInventory to prevent race between concurrent adjustments |
| P2-13 | P2 | DI wiring: wire TxManager into server startup (cmd/admin, cmd/pda) | completed | 2026-07-21 | TxManager created at startup in both cmd/admin and cmd/pda; inventory service uses NewInventoryServiceWithTx for atomic adjust operations |
| P2-14 | P2 | CountWaves + CountRoles for TaskRepository and UserRepository | completed | 2026-07-21 | Added CountWaves(warehouseID) to TaskRepository + postgres impl with 1 integration test; Added CountRoles() to UserRepository + postgres impl with 1 integration test; updated mockTaskRepo and stubUserRepo in service tests |
| P2-15 | P2 | AuditLog list endpoint with pagination (service + Admin API) | completed | 2026-07-21 | AuditLogService.ListAuditLogs + AuditLogHandler.ListAuditLogs (GET /api/v1/audit-logs) with user_id/action/resource filters |
| P2-16 | P2 | User list endpoint with pagination (service + Admin API) | completed | 2026-07-21 | UserService.ListUsers + UserHandler.ListUsers (GET /api/v1/users) with status filter + pagination; 3 unit tests
| P2-17 | P2 | Location status state machine (domain methods + service operations) | completed | 2026-07-21 | IsTerminal + CanTransitionTo on Location; UpdateLocationStatus enforces state machine; 11 domain tests + 4 service tests |
| P2-18 | P2 | Order line & ASN status transition operations in OrderService | completed | 2026-07-21 | Added UpdateOrderLineStatus + UpdateASNStatus service methods; added GetOrderLine to OrderRepository; 10 unit tests |
| P2-19 | P2 | Service tests for order line, ASN, and inventory status transitions | completed | 2026-07-21 | Inventory status state machine (IsTerminal, CanTransitionTo) on Inventory domain (11 tests); UpdateInventoryStatus service method (10 tests); all transitions covered (available↔quarantine, damaged↔available, expired terminal) |
| P2-20 | P1 | Seed data: create default admin user with hashed password | completed | 2026-07-21 | Migration now has real bcrypt hash; seed script always ensures admin user + roles exist before demo data |
| P2-21 | P2 | Role-based authorization middleware (check permissions on API routes) | pending | — | Parse role_ids from JWT, check Permission.Can() against resource+action; decorate endpoints with required permissions |
| P2-22 | P2 | User service + Admin API (CRUD users, register, /me profile) | pending | — | Create/list/update users, password change, proper GET /api/v1/auth/me response with full user profile |
| P2-23 | P2 | Token blacklist / logout (invalidate refresh tokens) | pending | — | Redis-backed JTI blacklist; logout endpoint; middleware checks blacklist on each request ⚠ Depends on P5-23 (Redis bootstrap) |
| P2-24 | P2 | Apply auth middleware to PDA server (cmd/pda) | pending | — | PDA endpoints (task assignment, status updates) should require auth; share middleware from api/middleware |
| P2-25 | P2 | Makefile `watch` target for hot-reload development | pending | — | Use `air` or similar for auto-rebuild on file changes; speeds up admin/PDA UI iteration |
| P2-26 | P2 | Migration tracking table (schema_migrations) | pending | — | Current `make migrate` blindly applies all .sql files; add a tracking table so each migration runs exactly once |
| P2-27 | P2 | UserRepository.UpdatePasswordHash method | pending | — | UpdateUser skips password_hash; seed script uses raw SQL. Add UpdatePasswordHash(ctx, userID, hash) to the repo interface + impl; update auth service's password change to use it |
| P2-28 | P2 | Makefile `setup-full` target (dev + migrate + seed in one step) | pending | — | Single command for fresh dev environment ⚠ Depends on P2-06 + P2-26 |
| P2-29 | P1 | Admin: ProtectedRoute auth guard (redirect unauthenticated users to /login) | pending | — | Wrap admin layout routes; redirect to /login if no access token in zustand store; preserve intended destination for post-login redirect ⚠ Depends on P2-20 (seed admin user for testing) |
| P2-30 | P2 | Admin: ESLint + Prettier config | pending | — | Add .eslintrc.cjs with @typescript-eslint + react-hooks plugin; .prettierrc with consistent formatting; lint script already in package.json |
| P2-31 | P2 | Admin: API service modules (warehouses, SKUs, inventory, orders, tasks, auth) | pending | — | Per-entity API modules wrapping the axios client; typed request/response functions for each endpoint; used by Phase 3 page components |

## Phase 3: Admin & PDA UI

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P3-01 | P2 | Admin: splash page + login form | pending | — | Landing page with JWT login; redirect to dashboard on success ⚠ Depends on P2-29 (ProtectedRoute) |
| P3-02 | P2 | Admin: inventory dashboard (summary cards + charts) | pending | — | Total SKUs, low stock alerts, inventory by zone; charts via recharts or ant-design-charts |
| P3-03 | P2 | Admin: SKU management pages (list, create, edit, delete) | pending | — | Table with pagination, search, filters; form with JSONB attributes editor |
| P3-04 | P2 | Admin: order management pages (list, detail, status transitions) | pending | — | Order table with status badges; detail view with lines + ASN; status action buttons |
| P3-05 | P2 | Admin: task monitoring pages (list, detail, wave view) | pending | — | Filter by type/status/assignee; detail with scan history; wave grouping |
| P3-06 | P2 | Admin: user management pages (list, create, edit, roles) | pending | — | CRUD users; role assignment; password reset for admins |
| P3-07 | P2 | Admin: audit log viewer (filterable, paginated) | pending | — | Filter by entity type, action, user, date range; export to CSV |
| P3-08 | P2 | Admin: API client layer (axios/fetch wrapper with JWT refresh) | cancelled | — | Merged into P2-07 (admin scaffold) — API client with JWT refresh + request queuing already built |
| P3-09 | P2 | PDA: login + task list screen | pending | — | Mobile-first login; swipe-to-refresh task list; status badges |
| P3-10 | P2 | PDA: barcode scanner component (webcam + keyboard wedge) | pending | — | Support both camera-based scanning (html5-qrcode) and hardware scanner input; vibrate on scan |
| P3-11 | P2 | PDA: receiving flow (scan ASN → confirm items → submit) | pending | — | Step-by-step wizard: scan ASN barcode, verify lines, enter received qty, confirm |
| P3-12 | P2 | PDA: putaway flow (scan location → confirm putaway) | pending | — | Scan item barcode, suggest location (FEFO), scan target location, confirm |
| P3-13 | P2 | PDA: picking flow (scan order → pick items → confirm) | pending | — | Show pick list with locations, scan item + location to confirm, flag shortages |
| P3-14 | P2 | PDA: cycle count flow (scan location → count → submit) | pending | — | Blind count entry; variance alert if qty differs from system; supervisor approval for large variances |
| P3-15 | P2 | PDA: exception handling (damage, shortage, wrong item) | pending | — | Quick exception buttons during any flow; photo capture; supervisor notification |
| P3-16 | P2 | PDA: shipping flow (verify order → generate label → confirm ship) | pending | — | Scan order barcode, verify picked items, trigger label print, confirm shipment; carrier/service selection; tracking number capture |
| P3-17 | P2 | PDA: stock take / physical inventory count flow | pending | — | Full warehouse count mode (different from cycle count); area-by-area count sheets; blind count entry; variance reconciliation; adjust inventory after approval |

## Phase 4: Integration Adapters

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P4-01 | P3 | Integration adapter interface (Adapter, Connector, MessageRouter) | pending | — | Define Go interfaces in internal/integration/; decouple WCS/RCS/MES/ERP behind uniform contract |
| P4-02 | P3 | gRPC server bootstrap + service registration | pending | — | Start gRPC server alongside HTTP; register Warehouse/Inventory/Task/Order proto services |
| P4-03 | P3 | WCS adapter: conveyor/sorter task dispatch (gRPC client) | pending | — | Send pick/putaway/sort commands to conveyor system; receive completion/error events |
| P4-04 | P3 | RCS adapter: AGV/AMR move task dispatch (gRPC client) | pending | — | Send move commands to robot fleet (from→to location); receive position updates |
| P4-05 | P3 | MES adapter: production order → WMS outbound order | pending | — | Receive production orders from MES; create corresponding outbound WMS orders; send completion callbacks |
| P4-06 | P3 | ERP adapter: purchase order → ASN, sales order → outbound order | pending | — | Receive PO/SO from ERP via REST/gRPC; create ASN/outbound order; send fulfillment status back |
| P4-07 | P3 | WebSocket gateway for real-time task updates | pending | — | Push task status changes to connected admin/PDA clients; fallback to polling ⚠ Depends on P5-23 (Redis pub/sub for multi-instance fan-out) |
| P4-08 | P3 | Integration message queue (Redis Streams or NATS) | pending | — | Async message bus for integration events; retry + DLQ for failed deliveries |

## Phase 5: Production Readiness

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P5-01 | P2 | Health check endpoints (/health, /ready) with dependency checks | pending | — | Liveness probe (always OK); readiness probe (DB + Redis reachable); /healthz aggregation |
| P5-02 | P3 | Dockerfiles for admin + PDA services (multi-stage build) | pending | — | Build Go binary in golang:1.26; run in distroless; non-root user |
| P5-03 | P3 | Kubernetes manifests (deployment, service, configmap, ingress) | pending | — | Namespace isolation; resource limits; liveness/readiness probes; ConfigMap for env |
| P5-04 | P3 | Prometheus metrics endpoint + instrumentation | pending | — | HTTP request duration/histogram; DB query latency; active tasks gauge; inventory levels gauge |
| P5-05 | P2 | Structured request logging with trace IDs | pending | — | Inject trace_id at middleware; propagate through context; structured log fields (method, path, status, duration, user_id) |
| P5-06 | P3 | OpenAPI/Swagger specification (auto-generated from handlers) | pending | — | Annotate handlers with swaggo comments; serve /swagger/index.html; publish to API docs |
| P5-07 | P3 | GitHub Actions CI/CD pipeline | pending | — | Lint + test on PR; build + push Docker images on merge to master; deploy to staging |
| P5-08 | P3 | Rate limiting middleware (token bucket, per-user/IP) | pending | — | Redis-backed rate limiter; configurable limits per endpoint; 429 responses with Retry-After ⚠ Depends on P5-23 |
| P5-09 | P3 | Redis caching layer for hot inventory queries | pending | — | Cache inventory by SKU+Location; invalidate on adjustment; configurable TTL; cache-hit metrics ⚠ Depends on P5-23 |
| P5-10 | P3 | Redis session store for refresh tokens | pending | — | Replace in-memory token store with Redis; survive server restarts; TTL auto-expiry ⚠ Depends on P5-23 |
| P5-11 | P3 | Database backup/restore scripts (pg_dump cron + S3 upload) | pending | — | Daily pg_dump; compress + upload to S3/MinIO; restore procedure documented |
| P5-12 | P3 | End-to-end integration tests (critical business flows) | pending | — | Order creation → ASN → receiving → putaway → inventory verification; run in CI with testcontainers |
| P5-13 | P3 | Performance benchmark suite (wrk/k6 load tests) | pending | — | Baseline latency/throughput for key endpoints; run in CI to catch regressions; target p99 < 200ms |
| P5-14 | P3 | Graceful shutdown + signal handling | pending | — | SIGTERM/SIGINT → drain HTTP connections → close DB pool → flush logs → exit; respect K8s terminationGracePeriodSeconds |
| P5-15 | P3 | CORS configuration hardening + CSP headers | pending | — | Restrict allowed origins to actual frontend URLs; add Content-Security-Policy, X-Content-Type-Options headers |
| P5-16 | P3 | Input validation hardening (SQL injection, XSS, parameter tampering) | pending | — | Audit all user inputs; sanitize query params; validate UUID format; max request body size |
| P5-17 | P3 | TLS/HTTPS configuration for API servers | pending | — | Auto-cert via ACME/Let's Encrypt in dev; manual cert config for prod; redirect HTTP→HTTPS; HSTS header |
| P5-18 | P3 | Secrets management (DB password, JWT secret, API keys) | pending | — | Load secrets from env vars or file; never hard-code; document required secrets in .env.example; validate at startup |
| P5-19 | P3 | golangci-lint configuration + code quality automation | pending | — | `.golangci.yml` with recommended linters (errcheck, gosec, govet, staticcheck, revive); `make lint` already exists; add to quality gate |
| P5-20 | P3 | Response compression (gzip/brotli) middleware | pending | — | Compress JSON responses > 1KB; respect Accept-Encoding; skip for already-compressed types (images); configurable level |
| P5-21 | P3 | Grafana dashboard templates for WMS metrics | pending | — | Provisioned dashboards: API overview (latency/errors), inventory health, task throughput, DB pool stats; JSON export |
| P5-22 | P3 | Request timeout & deadline propagation | pending | — | Per-route timeout config; context deadline propagation through service→repo chain; 504 Gateway Timeout response; respect K8s pod termination |
| P5-23 | P2 | Redis client bootstrap + connection pool | pending | — | Initialize Redis client at server startup; connection pool config; health check integration ⚠ BLOCKS: P2-23 (token blacklist), P5-08 (rate limiting), P5-09 (caching), P5-10 (session store), P5-24 (idempotency), P7-03 (config hot-reload), P4-07 (WebSocket pub/sub) |
| P5-24 | P3 | Idempotency key support for mutation endpoints | pending | — | Idempotency-Key header parsing; Redis-backed key dedup with TTL; return cached response for duplicate keys; prevent double-creates on retry ⚠ Depends on P5-23 |
| P5-25 | P2 | Soft delete pattern across domain entities | pending | — | Add deleted_at TIMESTAMPTZ to all tables; domain IsDeleted() method; repo List methods filter out soft-deleted by default; include_deleted query param for admin views |
| P5-26 | P3 | Bulk operation endpoints (batch create/update) | pending | — | POST /api/v1/inventory/bulk-adjust; POST /api/v1/tasks/bulk-assign; transactional bulk operations with partial-failure reporting |
| P5-27 | P3 | Database connection pool configuration + health | pending | — | Configurable max_connections, min_connections, max_idle_time; pool stats exported as metrics; connection acquisition timeout; /ready checks pool health |
| P5-28 | P3 | Request body size limits + content-type validation | pending | — | Max request body size middleware (default 1MB, configurable per-route); strict Content-Type checking (application/json only for JSON endpoints); 413/415 responses |
| P5-29 | P3 | Circuit breaker patterns for external integration adapters | pending | — | Configurable failure thresholds, half-open state, timeout; per-adapter circuit breaker; metrics export (open/closed/half-open count); graceful fallback responses |
| P5-30 | P3 | OpenTelemetry distributed tracing instrumentation | pending | — | OTel SDK for Go; auto-instrument HTTP + gRPC + DB calls; trace context propagation via W3C TraceContext headers; export to Jaeger/Tempo/OTLP collector; span attributes (user_id, warehouse_id, task_id) |
| P5-31 | P3 | Graceful degradation when Redis is unavailable | pending | — | Fallback to in-memory rate limiter (single-instance only); skip caching layer; log warning; auto-recover when Redis reconnects; must not crash on Redis connection loss |
| P5-32 | P3 | Database migration testing + rollback support | pending | — | Test each new migration against a copy of production schema; verify both up and down migrations work; migration dry-run mode; document rollback procedure |
| P5-33 | P3 | Slow query logging + query analysis | pending | — | Log queries exceeding configurable threshold (default 100ms); track query plan changes; expose slow query stats via metrics endpoint; periodic query plan review |
| P5-34 | P3 | Container-ready structured logging (stdout JSON) | pending | — | JSON log format for container log aggregation; log level configurable per env; include service name + version in every log line; structured fields for log aggregation tools (Loki/ELK) |

## Phase 6: Advanced WMS Features

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P6-01 | P4 | Wave management service (batch picking optimization) | pending | — | Group outbound orders by zone/aisle; generate pick waves; release to PDA operators; wave status tracking |
| P6-02 | P4 | Cycle counting engine (ABC classification, schedule, reconcile) | pending | — | ABC analysis by transaction frequency; auto-generate count tasks; variance review workflow; adjust inventory |
| P6-03 | P4 | Replenishment automation (min/max thresholds, auto-tasks) | pending | — | Per-location min/max qty config; auto-generate replenishment tasks when below min; FIFO source selection |
| P6-04 | P4 | Batch/lot tracking & traceability | pending | — | Lot number on inventory; forward trace (lot→orders) and backward trace (order→lot); expiry alerts |
| P6-05 | P4 | Barcode label printing (ZPL/PDF templates) | pending | — | Generate barcode labels for locations, SKUs, pallets; ZPL for Zebra printers; PDF fallback |
| P6-06 | P4 | Reporting engine (inventory valuation, order fulfillment, productivity) | pending | — | Scheduled + on-demand reports; CSV/PDF export; inventory aging, order cycle time, operator productivity |
| P6-07 | P4 | Dashboard KPI widgets (inventory health, order SLA, task throughput) | pending | — | Real-time KPI cards on admin dashboard; configurable date range; drill-down to detail views |
| P6-08 | P4 | Returns management (RMA, inspection, restock/scrap) | pending | — | Create return order; inspection workflow (good/damaged); restock or scrap disposition; credit memo trigger |
| P6-09 | P4 | Quality inspection workflow (receiving gate, hold/release) | pending | — | Sampling rules (AQL); inspection checklist; hold/release decision; quarantine inventory integration |
| P6-10 | P4 | Cross-docking (inbound→outbound without storage) | pending | — | Flag orders for cross-dock; match inbound ASN to pending outbound; direct flow-through task generation |
| P6-11 | P4 | Slotting optimization (velocity-based location assignment) | pending | — | Analyze pick frequency per SKU; suggest optimal zone/location; auto-relocate slow movers; heat map |
| P6-12 | P4 | Multi-warehouse inventory visibility + transfers | pending | — | Cross-warehouse inventory query; inter-warehouse transfer orders; consolidated dashboard |
| P6-13 | P4 | Pallet/case management (LPN hierarchy) | pending | — | License Plate Number (LPN) tracking; parent-child unit hierarchy (pallet→case→each); LPN barcode labels; move/repackage operations |
| P6-14 | P4 | Serial number tracking | pending | — | Per-unit serial number capture during receiving; serial traceability from receipt to ship; serial validation during picks; serial-level inventory status |
| P6-15 | P4 | Dock door scheduling | pending | — | Truck appointment booking; dock door assignment; arrival/departure time tracking; carrier notification; yard check-in/check-out |
| P6-16 | P4 | Value-added services (kitting, labeling, co-packing) | pending | — | Kit BOM definition; kit assembly work orders; custom labeling rules per customer; co-packing work instructions; service billing |
| P6-17 | P4 | Putaway strategy engine (rule-based location assignment) | pending | — | Configurable strategies: fixed-location, random, zoned, velocity-based; strategy selection per SKU/zone; auto-suggest putaway location during receiving |
| P6-18 | P4 | Wave release rules engine | pending | — | Release waves by carrier cutoff time, order priority, zone affinity; configurable release triggers; partial wave release for urgent orders |
| P6-19 | P4 | Task interleaving optimization | pending | — | Combine pick + putaway tasks for single operator trip; reduce empty travel; configurable interleave rules per zone; efficiency tracking |
| P6-20 | P4 | Cartonization / packing optimization | pending | — | Suggest optimal carton size for order contents; pack station workflow; carton content manifest; shipping label generation trigger |
| P6-21 | P3 | Inventory unreserve / reservation release API | pending | — | POST /api/v1/inventory/{id}/unreserve to release reserved qty; auto-release on order cancellation; reservation expiry TTL for stale holds |
| P6-22 | P4 | Voice picking integration (headset-based) | pending | — | Voice-directed picking via headset API; text-to-speech location/item prompts; voice command recognition (confirm, skip, short); hands-free operation mode for PDA |
| P6-23 | P5 | Pick-to-light / put-to-light hardware interface | pending | — | Light module controller integration (Modbus/TCP or HTTP); activate location lights for pick/put instructions; confirm via button press; zone-level light controller management |
| P6-24 | P5 | Automated weigh station + dimensioning integration | pending | — | Capture weight + dimensions from in-line scales/dimensioners; validate against SKU master data; flag discrepancies; auto-update shipping weight; MQTT or serial protocol integration |

## Phase 7: Operational Excellence

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P7-01 | P4 | Alerting rules + notification channels (Slack, email, webhook) | pending | — | Define alert thresholds (low stock, task backlog, error rate spike); route via alertmanager-compatible webhooks |
| P7-02 | P4 | Audit trail query API with advanced filters | pending | — | Full-text search across audit log; export to S3 for compliance; retention policy |
| P7-03 | P4 | Configuration hot-reload (feature flags, thresholds) | pending | — | Redis-backed config; watch for changes without restart; feature flag for gradual rollout ⚠ Depends on P5-23 |
| P7-04 | P4 | Data archival + purging (completed orders, old audit logs) | pending | — | Archive completed orders > 90 days to cold storage; purge audit logs > 1 year; configurable retention |
| P7-05 | P4 | Disaster recovery runbook + failover procedure | pending | — | Document RPO/RTO targets; DB replica setup; Redis sentinel; region failover steps |
| P7-06 | P4 | API versioning strategy (v2 coexistence) | pending | — | URL path versioning (/api/v1, /api/v2); deprecation headers; sunset policy documented |
| P7-07 | P4 | Internationalization (i18n) for admin + PDA UI | pending | — | zh-CN + en-US strings; react-i18next; language switcher; date/number locale formatting |
| P7-08 | P4 | Accessibility audit (WCAG 2.1 AA for admin UI) | pending | — | Keyboard navigation; screen reader labels; color contrast; focus management |
| P7-09 | P4 | Mobile offline support (PWA for PDA) | pending | — | Service worker for offline caching; IndexedDB task queue for offline operations; auto-sync on reconnect; conflict resolution strategy; offline indicator UI |
| P7-10 | P4 | Data encryption at rest | pending | — | Application-level encryption for PII/sensitive fields; key rotation policy; optionally leverage pg_tde or filesystem encryption; document encryption scope |
| P7-11 | P4 | Capacity planning dashboard | pending | — | Storage utilization trends (30/60/90 day); zone occupancy heatmap; projected full date based on inbound velocity; capacity alert thresholds |
| P7-12 | P4 | SLA monitoring dashboard | pending | — | Order cycle time vs SLA target; task completion rate by type/zone; inventory accuracy KPI (system vs physical); operator productivity trends; configurable SLA thresholds |
| P7-13 | P4 | Multi-tenancy data isolation strategy | pending | — | Tenant ID on all tenant-owned entities; query filtering by tenant; connection pool per tenant or row-level security; tenant provisioning/de-provisioning workflow; cross-tenant access prevention |
| P7-14 | P4 | PII data masking in logs + audit trail | pending | — | Identify PII fields (email, username, IP address) in log output; auto-mask in structured logs; audit log of PII access; configurable masking rules per environment |
| P7-15 | P5 | GDPR/CCPA data subject request handling | pending | — | API endpoints for data export (all PII for a user) and data deletion; right-to-be-forgotten workflow; consent tracking for cookies/analytics; data processing agreement template |
| P7-16 | P5 | Security penetration testing checklist + remediation | pending | — | OWASP Top 10 checklist; auth bypass testing; injection testing; dependency vulnerability scanning (govulncheck); automated DAST in CI; findings→Jira/GitHub issue workflow |

## Phase 8: Integration & Ecosystem

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P8-01 | P5 | EDI supplier integration (850/856/810 transaction sets) | pending | — | Inbound 850 (PO), outbound 856 (ASN), outbound 810 (invoice); ANSI X12 translation layer; SFTP/AS2 transport |
| P8-02 | P5 | Carrier API integration (rate shopping, label, tracking) | pending | — | Multi-carrier abstraction (FedEx, UPS, DHL); real-time rate quotes; label generation; tracking number assignment; carrier performance analytics |
| P8-03 | P5 | 3PL billing engine | pending | — | Storage fee calculation (per pallet/day); transaction fees (per receipt/pick/ship); activity-based billing; invoice generation; customer rate cards ⚠ Depends on P7-13 (multi-tenancy) |
| P8-04 | P5 | IoT sensor integration (temperature, humidity monitoring) | pending | — | Sensor data ingestion via MQTT/HTTP; threshold alerts for cold chain; sensor-to-zone mapping; compliance reports for FDA/CFDA |
| P8-05 | P5 | API webhook system (outbound events) | pending | — | Event subscription registry; configurable webhook URLs per event type; retry with exponential backoff; delivery audit log; webhook secret signing |
| P8-06 | P5 | SSO/OAuth2 integration (Okta, Azure AD, Google Workspace) | pending | — | OAuth2/OIDC login flow; identity provider config; role mapping from IdP groups; coexist with existing JWT auth; admin SSO config UI |
| P8-07 | P5 | Supplier self-service portal | pending | — | Supplier-facing web UI for ASN submission; delivery appointment scheduling; PO acknowledgment; shipment status visibility; document upload (COA, packing list) |
| P8-08 | P5 | Customer self-service portal | pending | — | Customer-facing web UI for order status tracking; inventory availability check; return request (RMA) submission; invoice/statement download; shipment tracking |

## Phase 9: Analytics & Intelligence

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P9-01 | P5 | Demand forecasting (inbound/outbound volume prediction) | pending | — | Historical order pattern analysis; seasonal trend detection; SKU-level forecast; safety stock recommendations; forecast accuracy tracking |
| P9-02 | P5 | Labor planning & productivity analytics | pending | — | Task completion time tracking per operator; productivity KPIs (picks/hour, putaways/hour); staffing recommendations by shift; overtime alerting |
| P9-03 | P5 | Anomaly detection for inventory & operations | pending | — | Statistical outlier detection on inventory adjustments; unusual order cycle time detection; missing scan pattern detection; root cause suggestion |
| P9-04 | P5 | Pick path optimization | pending | — | Shortest-path routing through warehouse zones; order batching by proximity; travel distance estimation; before/after efficiency metrics |
| P9-05 | P5 | ABC classification automation | pending | — | Auto-classify SKUs as A/B/C based on transaction velocity; periodic reclassification (weekly/monthly); storage zone reassignment suggestions; pick face optimization |
| P9-06 | P5 | Cost-to-serve analysis | pending | — | Per-order operational cost breakdown (labor, packing, shipping); per-customer profitability dashboard; cost allocation rules engine; margin alerting |

## Phase 10: Developer Experience & Community

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P10-01 | P4 | Auto-generated API reference documentation | pending | — | Annotate all handlers with swaggo/openapi comments; serve /api/docs with Swagger UI; publish static docs; keep in sync with code via CI check |
| P10-02 | P4 | Postman / Insomnia API collection | pending | — | Export all Admin + PDA endpoints as Postman collection; include auth flow (login → use token); example request bodies; environment variables for local/dev/staging |
| P10-03 | P4 | Docker Compose production-like environment | pending | — | Multi-service compose with admin, pda, postgres, redis, nginx reverse proxy; health check dependencies; volume mounts for dev iteration |
| P10-04 | P4 | Development troubleshooting guide | pending | — | Common issues (DB connection refused, port conflicts, migration failures); diagnostic commands; log locations; how to reset to clean state |
| P10-05 | P5 | Changelog automation from conventional commits | pending | — | Generate CHANGELOG.md from git history using conventional commit format; categorize by feat/fix/refactor; link to roadmap task IDs; auto-update on release |
| P10-06 | P5 | Architecture Decision Records (ADR) for key choices | pending | — | Document rationale for UUID PKs, DDD layering, chi/v5 selection, pgx over GORM, JSONB for SKU attrs; ADR template in docs/adr/; numbered sequentially |
| P10-07 | P5 | Frontend test suite (Vitest + React Testing Library) | pending | — | Unit tests for auth store, API client, hooks; component tests for Login, TaskList, warehouse forms; snapshot tests for layout; CI integration |
| P10-08 | P5 | CLI admin tool (cobra-based) | pending | — | `wms-cli` binary for admin operations: seed DB, reset password, list users, health check, trigger migration, export data; useful for devops and scripting |

## Phase 11: Observability & Site Reliability

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P11-01 | P4 | Centralized log aggregation (Loki/ELK stack) | pending | — | Ship structured JSON logs from all services to Loki; log retention policy; logQL dashboards for error rates, slow requests; correlate logs with traces via trace_id |
| P11-02 | P4 | Distributed tracing visualization (Jaeger/Grafana Tempo) | pending | — | Deploy tracing backend; service dependency graph; span waterfall for critical flows (order→task→inventory); trace search by user/order/task ID ⚠ Depends on P5-30 (OTel instrumentation) |
| P11-03 | P4 | SLO/SLI definition + error budget dashboards | pending | — | Define SLOs: API availability (99.9%), p99 latency (< 200ms), inventory accuracy; error budget burn rate alerts; SLO compliance dashboards in Grafana |
| P11-04 | P4 | Synthetic monitoring / heartbeat checks | pending | — | Periodic synthetic requests to critical endpoints (login, query inventory, create order); measure end-to-end latency; alert on synthetic failure; runs from external locations |
| P11-05 | P5 | Incident response runbook + post-mortem template | pending | — | Runbook per incident type (DB down, Redis down, high error rate, inventory discrepancy); escalation path; communication template; post-mortem doc with timeline, root cause, action items |
| P11-06 | P5 | On-call rotation integration (PagerDuty/Opsgenie webhook) | pending | — | Alert routing to on-call via webhook; acknowledge/snooze from alert dashboard; escalation policy; schedule override for holidays; incident auto-creation from critical alerts |
| P11-07 | P4 | Database slow query analysis dashboard | pending | — | Grafana dashboard for slow query trends; per-query latency distribution; query plan change detection; index usage statistics; unused index identification |
| P11-08 | P5 | Resource usage forecasting + auto-scaling triggers | pending | — | CPU/memory/connection pool trend analysis; predict saturation point; auto-scaling recommendations (HPA thresholds); cost optimization suggestions (right-sizing) |

<!-- DISCOVER will scan for new feature ideas when pending drops below 5. Last discovery: P8-P9 added 2026-07-20. Last grooming: round 35, added P3-16..17 (shipping + stock take), P5-29..34 (circuit breakers, OTel, graceful degradation, migration testing, slow queries, container logging), P6-22..24 (voice picking, pick-to-light, weigh station), P7-13..16 (multi-tenancy, PII masking, GDPR, pen testing), P8-07..08 (supplier/customer portals), P11 (Observability & SRE, 8 tasks); cancelled P3-08 (merged into P2-07). (2026-07-21). -->
