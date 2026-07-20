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
| P1-01 | P0 | Repository interfaces (Warehouse, Inventory, Order, Task) | completed | 2026-07-20 | Define interfaces in internal/repository/ |
| P1-02 | P0 | PostgreSQL repository implementation (Warehouse + Zone + Location) | completed | 2026-07-20 | Implement warehouse repo with pgx, 8 integration tests pass |
| P1-03 | P0 | PostgreSQL repository implementation (SKU + Inventory) | completed | 2026-07-20 | SKU CRUD + Inventory CRUD + Query filter + Tx audit, 13 integration tests pass |
| P1-04 | P1 | PostgreSQL repository implementation (Order + OrderLine) | pending | — | Implement order repo with pgx; create, get, list, status update, line management |
| P1-05 | P1 | PostgreSQL repository implementation (Task + Wave) | pending | — | Implement task repo with pgx; create, assign, status flow, wave grouping |
| P1-06 | P1 | PostgreSQL repository implementation (ASN) | pending | — | ASN CRUD + ASN line management; split from original P1-04 for granularity |
| P1-07 | P1 | PostgreSQL repository implementation (User + Role + AuditLog) | pending | — | User CRUD, role management, audit log insertion; needed for auth later |
| P1-08 | P1 | HTTP middleware stack (request ID, logging, recovery, CORS) | pending | — | chi/v5 middleware; req-id propagation, structured request logging, panic recovery, CORS config |
| P1-09 | P1 | Warehouse service + Admin API (CRUD for warehouses, zones, locations) | pending | — | chi/v5 REST endpoints; thin handlers delegating to WarehouseService |
| P1-10 | P1 | SKU service + Admin API (CRUD for SKUs) | pending | — | chi/v5 REST endpoints; thin handlers delegating to SKUService |
| P1-11 | P1 | Inventory service + Admin API (query, adjust) | pending | — | With inventory transaction audit; check negative qty constraint |
| P1-12 | P1 | Order service + Admin API (create/manage orders) | pending | — | Inbound + Outbound order flows; status transitions; line-item management |
| P1-13 | P1 | Task service + PDA API (task assignment, status flow) | pending | — | Task lifecycle management; PDA endpoints; assignment logic |
| P1-14 | P1 | Config management + Logger integration into services | pending | — | Wire pkg/config and pkg/logger into cmd entry points; env/file config loading |
| P1-15 | P1 | Standardized error handling (API error codes, validation errors, problem details) | pending | — | RFC 7807 problem details; consistent JSON error shape; input validation helpers |
| P1-16 | P1 | DB transaction support for atomic inventory operations | pending | — | txManager: wrap inventory change + location update + tx audit in single DB tx |
| P1-17 | P2 | FEFO/FIFO inventory retrieval query method | pending | — | Add GetOldestInventory / GetExpiringInventory to InventoryRepository; blocks P5-02 |
| P1-18 | P2 | Pagination metadata for QueryInventory | pending | — | Return total count alongside filtered results; add to all list endpoints |
| P1-19 | P2 | Authentication service (JWT login, token refresh, session management) | pending | — | JWT generation + validation middleware; refresh token rotation; blocks P2-02 |
| P1-20 | P2 | Domain unit tests (state machines, business rules, validation) | pending | — | Pure Go tests — no infrastructure; test Order/Task status transitions, Inventory invariants |

## Phase 2: Admin Frontend

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P2-01 | P2 | Admin frontend scaffold (React + Ant Design + Vite + React Router) | pending | — | Layout, navigation, theme, API client (axios/fetch wrapper with JWT) |
| P2-02 | P2 | Login/Auth page + JWT integration | pending | — | Login form, token storage, auto-refresh, redirect; depends on P1-19 |
| P2-03 | P2 | Warehouse management pages (list, create, edit zones/locations) | pending | — | CRUD tables, forms, zone/location hierarchy views |
| P2-04 | P2 | SKU management pages (list, create, edit, attribute editor) | pending | — | SKU CRUD with dynamic attribute form; barcode display/generation |
| P2-05 | P2 | Inventory overview page (list, search, filter, detail, transaction log) | pending | — | Filterable inventory table; drill-down to transaction history |
| P2-06 | P2 | Order management pages (inbound list, outbound list, create, detail) | pending | — | Order CRUD; order line editor; status timeline; external_ref linking |
| P2-07 | P2 | Task monitoring dashboard (task list, status filter, assignment) | pending | — | Task table by status; assign worker; view task detail with instructions |
| P2-08 | P2 | Dashboard home page (KPIs, charts, alerts summary) | pending | — | Inventory turnover, order volume, task completion rate, low-stock alerts |
| P2-09 | P2 | User & Role management pages | pending | — | User CRUD; role editor with permission matrix; blocks P5-05 UI |

## Phase 3: PDA Operations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P3-01 | P3 | PDA frontend scaffold (React + Vite + mobile-first CSS + PWA) | pending | — | Mobile-optimized layout; bottom tabs; offline-ready with service worker |
| P3-02 | P3 | PDA auth + bootstrap (login, warehouse selection, task list) | pending | — | Simple login; session persistence; pull-to-refresh task list |
| P3-03 | P3 | Barcode scanner component (camera-based + manual input) | pending | — | Browser Barcode Detection API or ZXing; fallback to manual keyboard input |
| P3-04 | P3 | Receiving flow (scan ASN → confirm receipt → generate putaway tasks) | pending | — | Scan ASN barcode; verify expected vs actual qty; auto-create putaway tasks |
| P3-05 | P3 | Putaway flow (scan location → scan SKU → confirm → inventory update) | pending | — | Guided putaway: scan target location → scan item barcode → confirm qty |
| P3-06 | P3 | Picking flow (wave task → scan location → scan SKU → confirm pick) | pending | — | Pick list view; scan source location; validate correct SKU; report shorts |
| P3-07 | P3 | Cycle counting flow (count task → scan → submit → variance review) | pending | — | Blind count option; variance auto-detection; recount workflow |
| P3-08 | P3 | Shipping flow (scan outbound order → verify → confirm ship) | pending | — | Load verification; carrier integration; shipment confirmation |
| P3-09 | P3 | PDA exception handling (item not found, damaged, wrong location) | pending | — | Exception flow: report issue, attach reason, trigger supervisor review |
| P3-10 | P3 | PDA offline queue + sync (background sync when connectivity restored) | pending | — | IndexedDB queue; task completion can be queued offline; conflict resolution |

## Phase 4: Integrations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P4-01 | P4 | Integration adapter interface definition (common protocol + message format) | pending | — | Define IntegrationAdapter interface; standard message envelope; ack/nack protocol |
| P4-02 | P4 | WebSocket real-time events (inventory changes, task updates, alerts) | pending | — | Push events to admin + PDA clients; connection management; reconnection |
| P4-03 | P4 | Message queue integration (NATS for async task dispatch, event bus) | pending | — | Pub/sub for domain events; dead-letter queue; retry policies |
| P4-04 | P4 | API gateway + rate limiting + JWT auth enforcement | pending | — | Route-based rate limiting; JWT validation on all protected routes; API key for integrations |
| P4-05 | P4 | WCS adapter — conveyor control (divert, route, status query) | pending | — | Adapter for conveyor/sorter hardware; WebSocket or TCP protocol |
| P4-06 | P4 | WCS adapter — sorter interface (scan, sort, chute assignment) | pending | — | Scan-and-sort workflow; chute/door assignment; sort plan upload |
| P4-07 | P4 | RCS adapter — AGV/AMR task dispatch (move, dock, charge) | pending | — | Robot task dispatch via VDA 5050 or custom gRPC; position tracking |
| P4-08 | P4 | RCS adapter — robot fleet management (battery, errors, utilization) | pending | — | Fleet state monitoring; error recovery; zone/traffic control integration |
| P4-09 | P4 | MES adapter — production order sync (work orders → outbound raw material) | pending | — | Production order intake; material consumption + BOM component reservation |
| P4-10 | P4 | MES adapter — finished goods receipt (production → inventory) | pending | — | Auto-create inbound ASN from production output; quality check gate |
| P4-11 | P4 | ERP adapter — purchase order → inbound ASN | pending | — | PO sync; auto-create expected ASN; GRN (Goods Receipt Note) back to ERP |
| P4-12 | P4 | ERP adapter — sales order → outbound order → shipment confirmation | pending | — | SO sync; auto-create outbound order; ship confirmation + tracking back to ERP |

## Phase 5: Advanced Features

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P5-01 | P5 | Wave planning engine (batch, zone, carrier-based wave creation) | pending | — | Wave generation from orders; configurable grouping strategies |
| P5-02 | P5 | Inventory allocation engine (FIFO, FEFO, lot-specific, manual override) | pending | — | Auto-allocate inventory to order lines; reservation lifecycle; depends on P1-17 |
| P5-03 | P5 | Multi-level inventory (warehouse → zone → location → container/LPN) | pending | — | Container/LPN domain entity; nested inventory; container movements |
| P5-04 | P5 | Report engine (inventory turnover, aging, ABC analysis, order fill rate) | pending | — | Scheduled + on-demand reports; CSV/PDF export; email delivery |
| P5-05 | P5 | RBAC permissions (role-based access control UI + API enforcement) | pending | — | Permission middleware; resource/action checks on every API call; depends on P2-09 |
| P5-06 | P5 | Audit log viewer (operation history, traceability, compliance export) | pending | — | Filterable audit log UI; date range; export for compliance audit |
| P5-07 | P5 | Inventory alerts engine (low stock, expiry, stranded, slow-moving) | pending | — | Configurable thresholds; notification channels (in-app, email, webhook) |
| P5-08 | P5 | Multi-warehouse support (cross-warehouse transfers, global inventory view) | pending | — | Transfer orders between warehouses; consolidated inventory dashboard |
| P5-09 | P5 | Lot/batch traceability (raw material → WIP → finished goods — full chain) | pending | — | Lot genealogy; forward and backward trace; recall support |
| P5-10 | P5 | Replenishment engine (min/max levels, demand-driven, auto task generation) | pending | — | Replenishment rules per SKU/zone; auto-create replenishment tasks when low |

## Phase 6: Production Operations & DevOps

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P6-01 | P6 | Production Docker images (multi-stage builds, distroless, health checks) | pending | — | Optimized Dockerfiles for admin + PDA; HEALTHCHECK; non-root user |
| P6-02 | P6 | CI/CD pipeline (GitHub Actions: lint → test → build → deploy) | pending | — | On push/PR; go vet + test + build matrix; Docker image build + push |
| P6-03 | P6 | Kubernetes manifests (Deployment, Service, Ingress, ConfigMap, Secrets) | pending | — | k8s base configs; resource limits; readiness/liveness probes |
| P6-04 | P6 | Helm chart (parameterized deployment, environment overrides) | pending | — | Single helm install; values per env (dev/staging/prod); secrets management |
| P6-05 | P6 | Prometheus metrics instrumentation (HTTP latency, DB pool, business KPIs) | pending | — | /metrics endpoint; custom metrics (order count, task throughput, inventory changes) |
| P6-06 | P6 | Structured logging to stdout (JSON format, log levels, sampling) | pending | — | slog or zerolog integration; structured fields; sampling for high-volume paths |
| P6-07 | P6 | Distributed tracing (OpenTelemetry — trace propagation across services) | pending | — | OTLP exporter; span context in middleware; DB query spans |
| P6-08 | P6 | OpenAPI/Swagger documentation (auto-generated from code annotations) | pending | — | swaggo or similar; /api/docs endpoint; publish to API docs portal |
| P6-09 | P6 | Database backup & restore tooling (pg_dump automation, point-in-time recovery) | pending | — | Cron job for backups; restore procedure documented; backup verification |
| P6-10 | P6 | Health check endpoints (liveness, readiness, DB/Redis connectivity) | pending | — | /livez (process alive), /readyz (DB+Redis ok), /healthz (combined) |
| P6-11 | P6 | gRPC server implementation (inter-service communication + integration adapters) | pending | — | gRPC server bootstrap; reflection; interceptors (auth, logging, tracing) |
| P6-12 | P6 | Graceful shutdown + connection draining | pending | — | SIGTERM handler; drain HTTP connections; close DB pool; flush logs |

## Phase 7: Quality, Security & Hardening

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P7-01 | P7 | Domain unit test suite (all entities, state machines, business invariants) | pending | — | 80%+ coverage on internal/domain/; test negative qty, status transitions, validation |
| P7-02 | P7 | Service integration tests (WarehouseService, InventoryService, OrderService) | pending | — | Mock repositories; test orchestration logic; error paths |
| P7-03 | P7 | API integration tests (HTTP handler tests with mock services) | pending | — | httptest + chi router; status codes, response shapes, error scenarios |
| P7-04 | P7 | E2E tests (Playwright — admin login → create order → verify task created) | pending | — | Critical path smoke tests; run in CI against test DB |
| P7-05 | P7 | Input validation hardening (request body, query params, path params, UUID format) | pending | — | Validation middleware; SQL injection prevention; content-type checking |
| P7-06 | P7 | Security hardening (bcrypt cost tuning, rate limiting, CORS policy, CSP headers) | pending | — | Security headers; brute-force protection on login; sensitive data masking |
| P7-07 | P7 | Performance benchmarks (load testing with k6 — order creation, inventory query) | pending | — | Baseline benchmarks; identify N+1 queries; connection pool tuning |
| P7-08 | P7 | Database query optimization (missing indexes, query plan review, connection pool) | pending | — | EXPLAIN ANALYZE review; add indexes for hot queries; pg_stat_statements |
| P7-09 | P7 | Error code catalog (documented error codes, consistent across all APIs) | pending | — | Central error registry; codes for validation, auth, business rules, system errors |
| P7-10 | P7 | API client SDK generation (Go + TypeScript from OpenAPI spec) | pending | — | Auto-generated typed clients; reduces frontend API errors |
| P7-11 | P7 | godoc documentation pass (all exported symbols, package-level docs, examples) | pending | — | Ensure every exported symbol has doc comment; add example tests |
| P7-12 | P7 | Dependency auditing + SBOM generation (govulncheck, npx audit, SPDX) | pending | — | Vulnerability scanning in CI; SBOM for compliance; automated updates |
| P7-13 | P7 | Redis caching layer (hot inventory data, session storage, rate limit counters) | pending | — | Cache-aside for QueryInventory; TTL policies; cache invalidation on inventory change |
| P7-14 | P7 | Data export + import (CSV bulk import for SKUs, orders; data export for reporting) | pending | — | Bulk endpoints; streaming CSV parser; validation on import |

## Evolution Metrics

| Metric | Value |
|--------|-------|
| Total tasks | 84 |
| Completed | 9 |
| In progress | 0 |
| Pending | 75 |
| Success rate | — |
| Started | 2026-07-20 |
| Last evolution | 2026-07-20 (Round 3: P1-03 SKU+Inventory repos) |
| Last grooming | 2026-07-20 (Round 10: expanded Phases 6-7, split large tasks, reprioritized) |
