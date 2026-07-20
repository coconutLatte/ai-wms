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
| P2-12 | P2 | SELECT FOR UPDATE row-level locking for inventory adjustments | pending | — | Prevent race condition in AdjustInventory between read and write within tx |
| P2-13 | P2 | DI wiring: wire TxManager into server startup (cmd/admin, cmd/pda) | pending | — | Create TxManager at startup, inject into services via NewInventoryServiceWithTx |
| P2-14 | P2 | CountWaves + CountRoles for TaskRepository and UserRepository | pending | — | ListWaves and ListRoles lack count methods; needed for future paginated wave/role list APIs |
| P2-15 | P2 | AuditLog list endpoint with pagination (service + Admin API) | pending | — | CountAuditLogs now exists; needs service + handler wrapping ListResponse[T] |
| P2-16 | P2 | User list endpoint with pagination (service + Admin API) | pending | — | CountUsers now exists; needs service + handler wrapping ListResponse[T] |
| P2-17 | P2 | Location status state machine (domain methods + service operations) | pending | — | empty→occupied→reserved→blocked transitions; formalize with CanTransitionTo on Location |
| P2-18 | P2 | Order line & ASN status transition operations in OrderService | pending | — | Services currently lack UpdateOrderLineStatus/UpdateASNStatus methods; domain state machines ready |
| P2-19 | P2 | Service tests for order line, ASN, and inventory status transitions | pending | — | Cover the remaining entity state machines at the service layer; domain tests already done |
| P2-20 | P1 | Seed data: create default admin user with hashed password | pending | — | Blocks login testing; seed script (P2-06) now updates admin password to real bcrypt hash at runtime; initial migration still has placeholder — update migration to use a real hash too |
| P2-21 | P2 | Role-based authorization middleware (check permissions on API routes) | pending | — | Depends on P2-04; parse role_ids from JWT, check Permission.Can() against resource+action; decorate endpoints with required permissions |
| P2-22 | P2 | User service + Admin API (CRUD users, register, /me profile) | pending | — | Create/list/update users, password change, proper GET /api/v1/auth/me response with full user profile |
| P2-23 | P2 | Token blacklist / logout (invalidate refresh tokens) | pending | — | Redis-backed JTI blacklist; logout endpoint; middleware checks blacklist on each request |
| P2-24 | P2 | Apply auth middleware to PDA server (cmd/pda) | pending | — | PDA endpoints (task assignment, status updates) should require auth; share middleware from api/middleware |
| P2-25 | P2 | Makefile `watch` target for hot-reload development | pending | — | Use `air` or similar for auto-rebuild on file changes; speeds up admin/PDA UI iteration |
| P2-26 | P2 | Migration tracking table (schema_migrations) | pending | — | Current `make migrate` blindly applies all .sql files; add a tracking table so each migration runs exactly once |
| P2-27 | P2 | UserRepository.UpdatePasswordHash method | pending | — | UpdateUser skips password_hash; seed script uses raw SQL. Add UpdatePasswordHash(ctx, userID, hash) to the repo interface + impl; update auth service's password change to use it |
| P2-28 | P2 | Makefile `setup-full` target (dev + migrate + seed in one step) | pending | — | Single command for fresh dev environment; depends on P2-06 + P2-26 |
| P2-29 | P1 | Admin: ProtectedRoute auth guard (redirect unauthenticated users to /login) | pending | — | Wrap admin layout routes; redirect to /login if no access token in zustand store; preserve intended destination for post-login redirect |
| P2-30 | P2 | Admin: ESLint + Prettier config | pending | — | Add .eslintrc.cjs with @typescript-eslint + react-hooks plugin; .prettierrc with consistent formatting; lint script already in package.json |
| P2-31 | P2 | Admin: API service modules (warehouses, SKUs, inventory, orders, tasks, auth) | pending | — | Per-entity API modules wrapping the axios client; typed request/response functions for each endpoint; used by Phase 3 page components |

## Phase 3: Admin & PDA UI

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P3-01 | P2 | Admin: splash page + login form | pending | — | Landing page with JWT login; redirect to dashboard on success |
| P3-02 | P2 | Admin: inventory dashboard (summary cards + charts) | pending | — | Total SKUs, low stock alerts, inventory by zone; charts via recharts or ant-design-charts |
| P3-03 | P2 | Admin: SKU management pages (list, create, edit, delete) | pending | — | Table with pagination, search, filters; form with JSONB attributes editor |
| P3-04 | P2 | Admin: order management pages (list, detail, status transitions) | pending | — | Order table with status badges; detail view with lines + ASN; status action buttons |
| P3-05 | P2 | Admin: task monitoring pages (list, detail, wave view) | pending | — | Filter by type/status/assignee; detail with scan history; wave grouping |
| P3-06 | P2 | Admin: user management pages (list, create, edit, roles) | pending | — | CRUD users; role assignment; password reset for admins |
| P3-07 | P2 | Admin: audit log viewer (filterable, paginated) | pending | — | Filter by entity type, action, user, date range; export to CSV |
| P3-08 | P2 | Admin: API client layer (axios/fetch wrapper with JWT refresh) | completed | 2026-07-20 | Implemented in P2-07; axios instance with Bearer token injection, 401 interceptor with JWT refresh + concurrent request queuing, shared API TypeScript types matching backend response envelope |
| P3-09 | P2 | PDA: login + task list screen | pending | — | Mobile-first login; swipe-to-refresh task list; status badges |
| P3-10 | P2 | PDA: barcode scanner component (webcam + keyboard wedge) | pending | — | Support both camera-based scanning (html5-qrcode) and hardware scanner input; vibrate on scan |
| P3-11 | P2 | PDA: receiving flow (scan ASN → confirm items → submit) | pending | — | Step-by-step wizard: scan ASN barcode, verify lines, enter received qty, confirm |
| P3-12 | P2 | PDA: putaway flow (scan location → confirm putaway) | pending | — | Scan item barcode, suggest location (FEFO), scan target location, confirm |
| P3-13 | P2 | PDA: picking flow (scan order → pick items → confirm) | pending | — | Show pick list with locations, scan item + location to confirm, flag shortages |
| P3-14 | P2 | PDA: cycle count flow (scan location → count → submit) | pending | — | Blind count entry; variance alert if qty differs from system; supervisor approval for large variances |
| P3-15 | P2 | PDA: exception handling (damage, shortage, wrong item) | pending | — | Quick exception buttons during any flow; photo capture; supervisor notification |

## Phase 4: Integration Adapters

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P4-01 | P3 | Integration adapter interface (Adapter, Connector, MessageRouter) | pending | — | Define Go interfaces in internal/integration/; decouple WCS/RCS/MES/ERP behind uniform contract |
| P4-02 | P3 | gRPC server bootstrap + service registration | pending | — | Start gRPC server alongside HTTP; register Warehouse/Inventory/Task/Order proto services |
| P4-03 | P3 | WCS adapter: conveyor/sorter task dispatch (gRPC client) | pending | — | Send pick/putaway/sort commands to conveyor system; receive completion/error events |
| P4-04 | P3 | RCS adapter: AGV/AMR move task dispatch (gRPC client) | pending | — | Send move commands to robot fleet (from→to location); receive position updates |
| P4-05 | P3 | MES adapter: production order → WMS outbound order | pending | — | Receive production orders from MES; create corresponding outbound WMS orders; send completion callbacks |
| P4-06 | P3 | ERP adapter: purchase order → ASN, sales order → outbound order | pending | — | Receive PO/SO from ERP via REST/gRPC; create ASN/outbound order; send fulfillment status back |
| P4-07 | P3 | WebSocket gateway for real-time task updates | pending | — | Push task status changes to connected admin/PDA clients; fallback to polling |
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
| P5-08 | P3 | Rate limiting middleware (token bucket, per-user/IP) | pending | — | Redis-backed rate limiter; configurable limits per endpoint; 429 responses with Retry-After |
| P5-09 | P3 | Redis caching layer for hot inventory queries | pending | — | Cache inventory by SKU+Location; invalidate on adjustment; configurable TTL; cache-hit metrics |
| P5-10 | P3 | Redis session store for refresh tokens | pending | — | Replace in-memory token store with Redis; survive server restarts; TTL auto-expiry |
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

## Phase 7: Operational Excellence

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P7-01 | P4 | Alerting rules + notification channels (Slack, email, webhook) | pending | — | Define alert thresholds (low stock, task backlog, error rate spike); route via alertmanager-compatible webhooks |
| P7-02 | P4 | Audit trail query API with advanced filters | pending | — | Full-text search across audit log; export to S3 for compliance; retention policy |
| P7-03 | P4 | Configuration hot-reload (feature flags, thresholds) | pending | — | Redis-backed config; watch for changes without restart; feature flag for gradual rollout |
| P7-04 | P4 | Data archival + purging (completed orders, old audit logs) | pending | — | Archive completed orders > 90 days to cold storage; purge audit logs > 1 year; configurable retention |
| P7-05 | P4 | Disaster recovery runbook + failover procedure | pending | — | Document RPO/RTO targets; DB replica setup; Redis sentinel; region failover steps |
| P7-06 | P4 | API versioning strategy (v2 coexistence) | pending | — | URL path versioning (/api/v1, /api/v2); deprecation headers; sunset policy documented |
| P7-07 | P4 | Internationalization (i18n) for admin + PDA UI | pending | — | zh-CN + en-US strings; react-i18next; language switcher; date/number locale formatting |
| P7-08 | P4 | Accessibility audit (WCAG 2.1 AA for admin UI) | pending | — | Keyboard navigation; screen reader labels; color contrast; focus management |
| P7-09 | P4 | Mobile offline support (PWA for PDA) | pending | — | Service worker for offline caching; IndexedDB task queue for offline operations; auto-sync on reconnect; conflict resolution strategy; offline indicator UI |
| P7-10 | P4 | Data encryption at rest | pending | — | Application-level encryption for PII/sensitive fields; key rotation policy; optionally leverage pg_tde or filesystem encryption; document encryption scope |
| P7-11 | P4 | Capacity planning dashboard | pending | — | Storage utilization trends (30/60/90 day); zone occupancy heatmap; projected full date based on inbound velocity; capacity alert thresholds |
| P7-12 | P4 | SLA monitoring dashboard | pending | — | Order cycle time vs SLA target; task completion rate by type/zone; inventory accuracy KPI (system vs physical); operator productivity trends; configurable SLA thresholds |

## Phase 8: Integration & Ecosystem

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P8-01 | P5 | EDI supplier integration (850/856/810 transaction sets) | pending | — | Inbound 850 (PO), outbound 856 (ASN), outbound 810 (invoice); ANSI X12 translation layer; SFTP/AS2 transport |
| P8-02 | P5 | Carrier API integration (rate shopping, label, tracking) | pending | — | Multi-carrier abstraction (FedEx, UPS, DHL); real-time rate quotes; label generation; tracking number assignment; carrier performance analytics |
| P8-03 | P5 | 3PL billing engine | pending | — | Storage fee calculation (per pallet/day); transaction fees (per receipt/pick/ship); activity-based billing; invoice generation; customer rate cards |
| P8-04 | P5 | IoT sensor integration (temperature, humidity monitoring) | pending | — | Sensor data ingestion via MQTT/HTTP; threshold alerts for cold chain; sensor-to-zone mapping; compliance reports for FDA/CFDA |
| P8-05 | P5 | API webhook system (outbound events) | pending | — | Event subscription registry; configurable webhook URLs per event type; retry with exponential backoff; delivery audit log; webhook secret signing |
| P8-06 | P5 | SSO/OAuth2 integration (Okta, Azure AD, Google Workspace) | pending | — | OAuth2/OIDC login flow; identity provider config; role mapping from IdP groups; coexist with existing JWT auth; admin SSO config UI |

## Phase 9: Analytics & Intelligence

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P9-01 | P5 | Demand forecasting (inbound/outbound volume prediction) | pending | — | Historical order pattern analysis; seasonal trend detection; SKU-level forecast; safety stock recommendations; forecast accuracy tracking |
| P9-02 | P5 | Labor planning & productivity analytics | pending | — | Task completion time tracking per operator; productivity KPIs (picks/hour, putaways/hour); staffing recommendations by shift; overtime alerting |
| P9-03 | P5 | Anomaly detection for inventory & operations | pending | — | Statistical outlier detection on inventory adjustments; unusual order cycle time detection; missing scan pattern detection; root cause suggestion |
| P9-04 | P5 | Pick path optimization | pending | — | Shortest-path routing through warehouse zones; order batching by proximity; travel distance estimation; before/after efficiency metrics |
| P9-05 | P5 | ABC classification automation | pending | — | Auto-classify SKUs as A/B/C based on transaction velocity; periodic reclassification (weekly/monthly); storage zone reassignment suggestions; pick face optimization |
| P9-06 | P5 | Cost-to-serve analysis | pending | — | Per-order operational cost breakdown (labor, packing, shipping); per-customer profitability dashboard; cost allocation rules engine; margin alerting |

<!-- DISCOVER will scan for new feature ideas when pending drops below 5. Last discovery: P8-P9 added 2026-07-20. -->
