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

> **Dependency order**: Repos → Config/Errors → Middleware → Tx Support → Services → Quality/Gen

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P1-01 | P0 | Repository interfaces (Warehouse, Inventory, Order, Task) | completed | 2026-07-20 | Define interfaces in internal/repository/ |
| P1-02 | P0 | PostgreSQL repository implementation (Warehouse + Zone + Location) | completed | 2026-07-20 | Implement warehouse repo with pgx, 8 integration tests pass |
| P1-03 | P0 | PostgreSQL repository implementation (SKU + Inventory) | completed | 2026-07-20 | SKU CRUD + Inventory CRUD + Query filter + Tx audit, 13 integration tests pass |
| P1-04 | P1 | PostgreSQL repository implementation (Order + OrderLine) | pending | — | Implement order repo with pgx; create, get, list, status update, line management |
| P1-05 | P1 | PostgreSQL repository implementation (Task + Wave) | pending | — | Implement task repo with pgx; create, assign, status flow, wave grouping |
| P1-06 | P1 | PostgreSQL repository implementation (ASN) | pending | — | ASN CRUD + ASN line management; split from original P1-04 for granularity |
| P1-07 | P1 | PostgreSQL repository implementation (User + Role + AuditLog) | pending | — | User CRUD, role management, audit log insertion; needed for auth later |
| P1-14 | P1 | Config management + Logger integration into services | pending | — | Wire pkg/config and pkg/logger into cmd entry points; env/file config loading; should precede middleware + service tasks |
| P1-15 | P1 | Standardized error handling (API error codes, validation errors, problem details) | pending | — | RFC 7807 problem details; consistent JSON error shape; input validation helpers; pkg/errors domain sentinels already done; this adds API-layer formatting |
| P1-08 | P1 | HTTP middleware stack (request ID, logging, recovery, CORS) | pending | — | chi/v5 middleware; req-id propagation, structured request logging, panic recovery, CORS config |
| P1-16 | P1 | DB transaction support for atomic inventory operations | pending | — | txManager: wrap inventory change + location update + tx audit in single DB tx; needed before services |
| P1-09 | P1 | Warehouse service + Admin API (CRUD for warehouses, zones, locations) | pending | — | chi/v5 REST endpoints; thin handlers delegating to WarehouseService |
| P1-10 | P1 | SKU service + Admin API (CRUD for SKUs) | pending | — | chi/v5 REST endpoints; thin handlers delegating to SKUService |
| P1-11 | P1 | Inventory service + Admin API (query, adjust) | pending | — | With inventory transaction audit; check negative qty constraint |
| P1-12 | P1 | Order service + Admin API (create/manage orders) | pending | — | Inbound + Outbound order flows; status transitions; line-item management |
| P1-13 | P1 | Task service + PDA API (task assignment, status flow) | pending | — | Task lifecycle management; PDA endpoints; assignment logic |
| P1-20 | P1 | Domain unit tests (state machines, business rules, validation) | pending | — | Pure Go tests — no infrastructure; test Order/Task status transitions, Inventory invariants; promoted from P2 to P1 as foundational quality gate |
| P1-17 | P2 | FEFO/FIFO inventory retrieval query method | pending | — | Add GetOldestInventory / GetExpiringInventory to InventoryRepository; blocks P5-02 |
| P1-18 | P2 | Pagination metadata for QueryInventory | pending | — | Return total count alongside filtered results; add to all list endpoints |
| P1-19 | P2 | Authentication service (JWT login, token refresh, session management) | pending | — | JWT generation + validation middleware; refresh token rotation; blocks P2-02 |
| P1-21 | P1 | Proto code generation workflow (buf generate + CI check) | pending | — | Run buf generate to produce Go stubs; add CI step to verify generated code matches proto sources |
| P1-22 | P1 | Makefile dev targets (run-admin, run-pda, migrate, remaining gaps) | pending | — | build/test/lint/proto targets already exist; needs `make run-admin`, `make run-pda`, real `make migrate` via goose; partially complete |
| P1-23 | P2 | Development seed data script (sample warehouse, zones, locations, SKUs) | pending | — | CLI or SQL script to populate dev DB with realistic demo data; enables UI development; basic role/user seed already in 000001 |
| P1-24 | P1 | Redis client initialization + connection pool + health check | pending | — | Wire go-redis/v9 into cmd entry points; RedisAddr() in config already; connection pool config; /readyz Redis ping; needed for sessions + caching |
| P1-25 | P1 | Database migration tooling (golang-migrate or goose CLI integration) | pending | — | Replace docker-entrypoint auto-migration with explicit migration tool; `make migrate-up` / `make migrate-down`; migration version tracking table |

## Phase 2: Admin Frontend

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P2-01 | P2 | Admin frontend scaffold (React + Ant Design + Vite + React Router) | pending | — | Layout, navigation, theme, API client (axios/fetch wrapper with JWT) |
| P2-02 | P2 | Login/Auth page + JWT integration | pending | — | Login form, token storage, auto-refresh, redirect; depends on P1-19 |
| P2-03 | P2 | Warehouse management pages (list, create, edit) | pending | — | Warehouse CRUD tables + forms; split from zone/location for granularity |
| P2-04 | P2 | Zone & Location hierarchy management pages | pending | — | Zone list/create/edit per warehouse; location grid/table; drag-and-drop zone assignment; split from P2-03 |
| P2-05 | P2 | SKU management pages (list, create, edit, attribute editor) | pending | — | SKU CRUD with dynamic attribute form; barcode display/generation |
| P2-06 | P2 | Inventory overview page (list, search, filter, detail, transaction log) | pending | — | Filterable inventory table; drill-down to transaction history |
| P2-07 | P2 | Order management pages (inbound list, outbound list, create, detail) | pending | — | Order CRUD; order line editor; status timeline; external_ref linking |
| P2-08 | P2 | Task monitoring dashboard (task list, status filter, assignment) | pending | — | Task table by status; assign worker; view task detail with instructions |
| P2-09 | P2 | Dashboard home page (KPIs, charts, alerts summary) | pending | — | Inventory turnover, order volume, task completion rate, low-stock alerts |
| P2-10 | P2 | User & Role management pages | pending | — | User CRUD; role editor with permission matrix; blocks P5-06 UI |
| P2-11 | P2 | Shared API client + TypeScript types (generated from OpenAPI or handwritten) | pending | — | Typed API client shared between admin and PDA; request/response types; error handling; reduces frontend integration bugs |

## Phase 3: PDA Operations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P3-01 | P3 | PDA frontend scaffold (React + Vite + mobile-first CSS) | pending | — | Mobile-optimized layout; bottom tabs; touch-friendly components; responsive breakpoints |
| P3-02 | P3 | PDA PWA setup (service worker, offline manifest, install prompt) | pending | — | Offline-ready with service worker; app manifest; install-to-home-screen; split from P3-01 for granularity |
| P3-03 | P3 | PDA auth + bootstrap (login, warehouse selection, task list) | pending | — | Simple login; session persistence; pull-to-refresh task list |
| P3-04 | P3 | Barcode scanner component (camera-based + manual input) | pending | — | Browser Barcode Detection API or ZXing; fallback to manual keyboard input |
| P3-05 | P3 | Receiving flow (scan ASN → confirm receipt → generate putaway tasks) | pending | — | Scan ASN barcode; verify expected vs actual qty; auto-create putaway tasks |
| P3-06 | P3 | Putaway flow (scan location → scan SKU → confirm → inventory update) | pending | — | Guided putaway: scan target location → scan item barcode → confirm qty |
| P3-07 | P3 | Picking flow (wave task → scan location → scan SKU → confirm pick) | pending | — | Pick list view; scan source location; validate correct SKU; report shorts |
| P3-08 | P3 | Cycle counting flow (count task → scan → submit → variance review) | pending | — | Blind count option; variance auto-detection; recount workflow |
| P3-09 | P3 | Shipping flow (scan outbound order → verify → confirm ship) | pending | — | Load verification; carrier integration; shipment confirmation |
| P3-10 | P3 | PDA exception handling (item not found, damaged, wrong location) | pending | — | Exception flow: report issue, attach reason, trigger supervisor review |
| P3-11 | P3 | PDA offline queue + sync (background sync when connectivity restored) | pending | — | IndexedDB queue; task completion can be queued offline; conflict resolution |

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
| P5-04 | P5 | Inventory reports (turnover, aging, ABC analysis) | pending | — | Compute and display inventory KPIs; depends on P5-18 for export |
| P5-05 | P5 | Operational reports (order fill rate, pick accuracy, cycle count variance) | pending | — | Operational KPI dashboards; date-range filtering; drill-down |
| P5-06 | P5 | RBAC permissions (role-based access control UI + API enforcement) | pending | — | Permission middleware; resource/action checks on every API call; depends on P2-10 |
| P5-07 | P5 | Audit log viewer (operation history, traceability, compliance export) | pending | — | Filterable audit log UI; date range; export for compliance audit |
| P5-08 | P5 | Inventory alerts engine (low stock, expiry, stranded, slow-moving) | pending | — | Configurable thresholds; notification channels (in-app, email, webhook) |
| P5-09 | P5 | Multi-warehouse support (cross-warehouse transfers, global inventory view) | pending | — | Transfer orders between warehouses; consolidated inventory dashboard |
| P5-10 | P5 | Lot/batch traceability (raw material → WIP → finished goods — full chain) | pending | — | Lot genealogy; forward and backward trace; recall support |
| P5-11 | P5 | Replenishment engine (min/max levels, demand-driven, auto task generation) | pending | — | Replenishment rules per SKU/zone; auto-create replenishment tasks when low |
| P5-12 | P5 | Dynamic slotting engine (velocity-based SKU → location optimization) | pending | — | ABC classification by pick frequency; auto-suggest optimal storage locations |
| P5-13 | P5 | Quality inspection workflow (QC checkpoints, sampling rules, hold/release) | pending | — | Sampling plans (AQL); inspection results; hold/release inventory; NCR tracking |
| P5-14 | P5 | Cross-docking flow (receiving → sort → ship, bypass storage) | pending | — | Identify cross-dock candidates; sort-by-destination; time-window management |
| P5-15 | P5 | Kitting / de-kitting (bundle and unbundle SKUs) | pending | — | Kit BOM definition; kit assembly tasks; component consumption; kit disassembly |
| P5-16 | P5 | Cartonization engine (optimal packaging per order) | pending | — | Box selection by item dimensions/weight; multi-carton split; packing slip generation |
| P5-17 | P5 | Scheduled report generation (cron-based, email delivery, PDF) | pending | — | Schedule daily/weekly/monthly reports; email with PDF attachment; report history |
| P5-18 | P5 | CSV/PDF export engine (generic export for all list views) | pending | — | Streaming CSV writer for large datasets; PDF with header/logo; used by P5-04 and P5-05 |
| P5-19 | P5 | Pick path optimization (shortest path routing through warehouse) | pending | — | Compute optimal pick sequence through warehouse zones/locations; reduce travel distance per wave |
| P5-20 | P5 | Task interleaving (combine putaway + pick for same operator/zone) | pending | — | Merge putaway and pick tasks in same zone to minimize empty travel; operator efficiency gains |

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
| P6-13 | P6 | Terraform/IaC for cloud infrastructure (DB, cache, compute, networking) | pending | — | Terraform modules for AWS/GCP; state management; environment workspaces |
| P6-14 | P6 | Blue-green deployment strategy (zero-downtime rollout, smoke tests, rollback) | pending | — | Deployment automation; smoke test after cutover; automated rollback on failure |
| P6-15 | P6 | Log aggregation pipeline (stdout → Loki/ELK → searchable archive) | pending | — | DaemonSet collectors; structured log parsing; retention policies; log-based alerts |

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
| P7-15 | P7 | Request ID propagation (HTTP header → context → log → gRPC → DB span) | pending | — | Extract/inject X-Request-ID at every boundary; structured log field; trace correlation |
| P7-16 | P7 | Circuit breaker for integration adapters (WCS/RCS/MES/ERP failure isolation) | pending | — | Circuit breaker per adapter; half-open probing; fallback behavior; blocks Phase 4 integration reliability |
| P7-17 | P7 | Secrets management (vault integration, encrypted config, no secrets in code) | pending | — | Externalize all secrets; HashiCorp Vault or cloud secrets manager; CI secret scanning |
| P7-18 | P7 | Go fuzz testing (fuzz input parsers, validators, JSON unmarshal paths) | pending | — | go test -fuzz for CSV parser, JSON payloads, barcode validator; catch panics and edge cases |

## Phase 8: Observability & Operations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P8-01 | P8 | Grafana dashboards (pre-built WMS KPIs: order volume, pick rate, accuracy) | pending | — | Dashboard JSON in repo; importable via Grafana provisioning; depends on P6-05 metrics |
| P8-02 | P8 | AlertManager rules + notification routing (PagerDuty, Slack, email) | pending | — | Alert rules: service down, DB pool exhaustion, error rate spike, task backlog; severity routing |
| P8-03 | P8 | SLO/SLI definitions + tracking (API latency, availability, throughput) | pending | — | Define SLOs (e.g., 99.9% API availability, p95 < 500ms); error budget burn alerts |
| P8-04 | P8 | Incident runbook documentation (common failure modes, recovery steps) | pending | — | Runbook per service; DB failover procedure; integration adapter recovery; escalation paths |
| P8-05 | P8 | Synthetic monitoring (external health probes simulating user flows) | pending | — | Blackbox exporter probes for critical API paths; login → dashboard → order create flow |
| P8-06 | P8 | Cloud resource tagging + cost allocation (per-environment cost tracking) | pending | — | Standard tags (env, service, owner); cost dashboards; unused resource detection |
| P8-07 | P8 | Chaos engineering baseline (controlled failure injection, resilience validation) | pending | — | Kill a DB replica, kill a pod, network partition; verify graceful degradation and recovery |

## Phase 9: Advanced WMS Features

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P9-01 | P9 | Yard management (truck check-in, dock door scheduling, trailer tracking) | pending | — | Dock calendar; trailer status (arrived, at-dock, loaded, departed); live yard map |
| P9-02 | P9 | Labor management (productivity tracking, standard times, performance dashboards) | pending | — | Engineered standards per task type; actual vs standard time; operator scorecards |
| P9-03 | P9 | Returns management (RMA workflow, disposition, inspection, refurbishment) | pending | — | RMA intake; disposition (restock, refurb, scrap, return-to-vendor); credit memo trigger |
| P9-04 | P9 | Value-added services (labeling, gift wrap, assembly, quality check) | pending | — | VAS work order; task generation for VAS steps; cost tracking per VAS type |
| P9-05 | P9 | Dangerous goods handling (segregation rules, storage zones, compliance labels) | pending | — | DG classification per SKU; segregation validation on putaway; hazmat shipping docs |
| P9-06 | P9 | Multi-client / 3PL support (client segregation, billing, client portal) | pending | — | Client entity; ownership on inventory; client-isolated views; activity-based billing data |
| P9-07 | P9 | Voice picking integration (headset-directed picking via Vocollect/Honeywell) | pending | — | Voice template per task type; audio feedback; hands-free confirmation; error correction dialogue |
| P9-08 | P9 | Pick-to-light / put-to-light hardware integration | pending | — | Light device registry; pick/put commands via MQTT or TCP; confirmation sensor handling |
| P9-09 | P9 | Automated palletizing/de-palletizing (robotic arm integration via RCS) | pending | — | Pallet build plan; layer-by-layer sequence; mix-SKU pallets; de-palletizing for receiving |

## Phase 10: Integration Hub & API Ecosystem

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P10-01 | P10 | Webhook system (configurable outbound webhooks for domain events) | pending | — | Event subscription per external system; retry with backoff; webhook delivery log |
| P10-02 | P10 | Integration testing sandbox (mock WCS/RCS/MES/ERP adapters) | pending | — | Mock adapters return realistic responses; record/replay mode; sandbox UI for manual testing |
| P10-03 | P10 | EDI support (EDIFACT DESADV/ORDERS/INVOIC messages for enterprise) | pending | — | EDI translator; mapping between WMS domain and EDIFACT; AS2 or SFTP transport |
| P10-04 | P10 | CSV/flat-file integration adapter (for legacy system onboarding) | pending | — | Watch folder → parse CSV → validate → import; configurable field mappings; error file output |
| P10-05 | P10 | API usage analytics + rate-limit quota management | pending | — | Per-client metrics; usage dashboards; hard/soft quota enforcement; overage billing data |
| P10-06 | P10 | Partner API developer portal (docs, sandbox keys, interactive console) | pending | — | Public-facing API docs; sandbox API key self-service; API explorer/console; depends on P6-08 |

## Phase 11: Data & Analytics

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P11-01 | P11 | CDC pipeline (PostgreSQL WAL → analytical store for reporting) | pending | — | Debezium or pgoutput logical replication; materialized views in analytics DB |
| P11-02 | P11 | Real-time warehouse operations dashboard (live map, heatmap, bottleneck view) | pending | — | SVG-based warehouse map; task heatmap by zone; bottleneck detection (queue depth, wait time) |
| P11-03 | P11 | ML-based demand forecasting (predict inbound/outbound volume by SKU) | pending | — | Time-series forecasting; seasonal adjustment; staff planning integration; depends on P11-01 |
| P11-04 | P11 | Slotting optimization (ML-based ABC + dynamic assignment) | pending | — | Combines velocity data + cube movement; continuous re-slot suggestions; what-if simulation |
| P11-05 | P11 | Anomaly detection (unusual adjustments, suspicious pick patterns, shrink analysis) | pending | — | Statistical outlier detection on inventory adjustments; pick variance clustering; shrink root-cause |

## Phase 12: Internationalization & Localization

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P12-01 | P12 | i18n framework setup (react-i18next, locale detection, translation file structure) | pending | — | Shared i18n config for admin + PDA; locale auto-detection from browser/navigator; translation namespace organization |
| P12-02 | P12 | Chinese (zh-CN) translation — Admin UI | pending | — | Full translation of all admin pages, forms, tables, notifications; most common operator locale in APAC |
| P12-03 | P12 | Japanese (ja-JP) translation — Admin UI | pending | — | Full translation of admin UI; important for Japan-market warehouses |
| P12-04 | P12 | PDA UI translations (zh-CN, ja-JP) | pending | — | Mobile-optimized translations for PDA flows; short labels for small screens |
| P12-05 | P12 | Date/time/number formatting by locale (dayjs locale integration) | pending | — | Locale-aware date formats, number separators, timezone display; consistent across all UI components |
| P12-06 | P12 | RTL (right-to-left) layout foundation | pending | — | CSS logical properties; direction-aware components; Arabic/Hebrew readiness (UI only, no translation yet) |

## Phase 13: Developer Experience & Tooling

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P13-01 | P13 | golangci-lint configuration (.golangci.yml with recommended linters) | pending | — | Enable errcheck, gosec, govet, staticcheck, revive; tune for DDD project conventions; CI integration |
| P13-02 | P13 | Pre-commit hooks (go fmt, go vet, conventional commit check, no-debug-print) | pending | — | pre-commit or lefthook config; fast checks only (<5s); blocks commits that fail quality gate |
| P13-03 | P13 | Go hot reload for development (air/CompileDaemon config) | pending | — | Watch .go files; auto-rebuild + restart on change; .air.toml config in repo; excludes vendor and testdata |
| P13-04 | P13 | Frontend npm workspace + shared package (types, API client, hooks) | pending | — | Monorepo structure with packages/shared; shared TypeScript types, API client, auth hooks; admin + PDA both consume |
| P13-05 | P13 | Frontend component testing (Vitest + React Testing Library setup) | pending | — | Test setup with jsdom; example tests for shared components; CI integration; snapshot testing for regression |
| P13-06 | P13 | Environment-specific config templates (.env.dev, .env.staging, .env.prod) | pending | — | Documented env var templates per environment; .env.example checked in; .env in .gitignore; Docker Compose overrides per env |
| P13-07 | P13 | Postman/Bruno API collection (pre-built requests for all endpoints) | pending | — | Collection file in repo; environment variables for local/dev/staging; pre-request scripts for auth token; self-documenting |

## Phase 14: Compliance & Data Governance

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P14-01 | P14 | GDPR data export (user data export endpoint, right-to-access compliance) | pending | — | Aggregate all PII per user; JSON export endpoint; audit log of access requests |
| P14-02 | P14 | GDPR data deletion (right-to-erasure, anonymization cascade) | pending | — | Anonymize or delete user PII while preserving inventory audit integrity; cascade rules per entity |
| P14-03 | P14 | Data retention policies (configurable retention per entity type, auto-purge) | pending | — | Retention config (e.g., audit_logs: 7yr, inventory_transactions: 3yr); scheduled purge job; dry-run mode |
| P14-04 | P14 | PII data classification (identify and tag PII fields, masking in logs) | pending | — | Tag PII fields in domain models; log masking middleware; data flow diagram of PII; compliance documentation |
| P14-05 | P14 | Audit compliance report (pre-built report for SOC2/ISO27001 audit evidence) | pending | — | Pre-built report covering access controls, change management, data encryption, backup verification; exportable as PDF |
| P14-06 | P14 | Consent management (cookie consent banner, data processing consent records) | pending | — | Consent UI for admin/PDA; record consent timestamp + version; consent withdrawal flow; privacy policy page |

## Evolution Metrics

| Metric | Value |
|--------|-------|
| Total tasks | 164 |
| Completed | 9 |
| In progress | 0 |
| Pending | 155 |
| Success rate | — |
| Started | 2026-07-20 |
| Last evolution | 2026-07-20 (Round 3: P1-03 SKU+Inventory repos) |
| Last grooming | 2026-07-20 (Round 11: reordered P1 for correct dependency flow; added P1-24 Redis, P1-25 migration tooling; split P2-03→P2-03+P2-04; added P2-11 shared API client; added P5-19 pick path + P5-20 task interleaving; added P7-18 fuzz testing; new Phases 12 i18n, 13 DevX, 14 Compliance — 35 net new tasks) |
