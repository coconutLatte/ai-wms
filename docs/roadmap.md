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

> **Dependency order**: Repos → Config/Errors → Middleware → Server Bootstrap → Tx Support → Services → Quality → Tooling

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
| P1-26 | P1 | HTTP server bootstrap + graceful shutdown skeleton | pending | — | chi router init in cmd/admin + cmd/pda; listen on configured ports; SIGTERM/SIGINT handler; basic /healthz endpoint; services mount routes onto this skeleton |
| P1-16 | P1 | DB transaction support for atomic inventory operations | pending | — | txManager: wrap inventory change + location update + tx audit in single DB tx; needed before services |
| P1-09 | P1 | Warehouse service + Admin API (CRUD for warehouses, zones, locations) | pending | — | chi/v5 REST endpoints; thin handlers delegating to WarehouseService |
| P1-10 | P1 | SKU service + Admin API (CRUD for SKUs) | pending | — | chi/v5 REST endpoints; thin handlers delegating to SKUService |
| P1-11 | P1 | Inventory service + Admin API (query, adjust) | pending | — | With inventory transaction audit; check negative qty constraint |
| P1-12 | P1 | Order service + Admin API (create/manage orders) | pending | — | Inbound + Outbound order flows; status transitions; line-item management |
| P1-13 | P1 | Task service + PDA API (task assignment, status flow) | pending | — | Task lifecycle management; PDA endpoints; assignment logic |
| P1-20 | P1 | Domain unit tests (state machines, business rules, validation) | pending | — | Pure Go tests — no infrastructure; test Order/Task status transitions, Inventory invariants; promoted from P2 to P1 as foundational quality gate; P7-01 extends this to full 80%+ coverage |
| P1-17 | P2 | FEFO/FIFO inventory retrieval query method | pending | — | Add GetOldestInventory / GetExpiringInventory to InventoryRepository; blocks P5-02 |
| P1-18 | P1 | Pagination metadata for QueryInventory | pending | — | Return total count alongside filtered results; add to all list endpoints; promoted from P2 to P1 — every list endpoint needs this |
| P1-19 | P2 | Authentication service (JWT login, token refresh, session management) | pending | — | JWT generation + validation middleware; refresh token rotation; blocks P2-02 |
| P1-27 | P2 | In-process domain event bus (publish/subscribe for domain events) | pending | — | Simple typed event publisher; subscriber registration; events: InventoryChanged, OrderStatusChanged, TaskCompleted; used by notification + audit + WebSocket push |
| P1-21 | P1 | Proto code generation workflow (buf generate + CI check) | pending | — | Makefile `proto` target already runs buf generate; remaining work: CI step to verify generated code matches proto sources; go_package paths need verification |
| P1-22 | P1 | Makefile dev targets (run-admin, run-pda, migrate, remaining gaps) | pending | — | build/test/lint/fmt/proto/quality targets already exist; needs `make run-admin`, `make run-pda`, real `make migrate` via goose; partially complete |
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
| P2-09 | P2 | Dashboard home page (KPIs, charts, alerts summary) | pending | — | Inventory turnover, order volume, task completion rate, low-stock alerts; depends on P2-03 through P2-08 (aggregates data from all management pages); should be last Phase 2 UI task |
| P2-10 | P2 | User & Role management pages | pending | — | User CRUD; role editor with permission matrix; blocks P5-06 UI |
| P2-11 | P2 | Shared API client + TypeScript types (generated from OpenAPI or handwritten) | pending | — | Typed API client shared between admin and PDA; request/response types; error handling; reduces frontend integration bugs; blocks P2-03 through P2-08 (admin pages need typed API client); start with handwritten types; upgrade to auto-generated from OpenAPI when P6-08 is complete |

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
| P3-10 | P3 | PDA exception handling (item not found, damaged, wrong location) | pending | — | Exception flow: report issue, attach reason, trigger supervisor review; blocks P15-04 for standardized reason codes (initially works with hardcoded codes) |
| P3-11 | P3 | PDA offline queue + sync (background sync when connectivity restored) | pending | — | IndexedDB queue; task completion can be queued offline; conflict resolution |
| P3-12 | P3 | PDA task batching (group multiple tasks for single operator, optimized route) | pending | — | Batch assignment (N tasks per operator); optimized pick route within batch; sequential/parallel execution mode; batch start/complete/progress; reduces travel between tasks |

## Phase 4: Integrations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P4-01 | P4 | Integration adapter interface definition (common protocol + message format) | pending | — | Define IntegrationAdapter interface; standard message envelope; ack/nack protocol |
| P4-02 | P4 | WebSocket real-time events (inventory changes, task updates, alerts) | pending | — | Push events to admin + PDA clients; connection management; reconnection; depends on P1-27 domain event bus |
| P4-03 | P4 | Message queue integration (NATS for async task dispatch, event bus) | pending | — | Pub/sub for domain events; dead-letter queue; retry policies |
| P4-04 | P4 | API gateway + rate limiting + JWT auth enforcement | pending | — | Route-based rate limiting; JWT validation on all protected routes; API key for integrations; depends on P1-19 auth service for JWT validation logic |
| P4-05 | P4 | WCS adapter — conveyor control (divert, route, status query) | pending | — | Adapter for conveyor/sorter hardware; WebSocket or TCP protocol |
| P4-06 | P4 | WCS adapter — sorter interface (scan, sort, chute assignment) | pending | — | Scan-and-sort workflow; chute/door assignment; sort plan upload |
| P4-07 | P4 | RCS adapter — AGV/AMR task dispatch (move, dock, charge) | pending | — | Robot task dispatch via VDA 5050 or custom gRPC; position tracking |
| P4-08 | P4 | RCS adapter — robot fleet management (battery, errors, utilization) | pending | — | Fleet state monitoring; error recovery; zone/traffic control integration |
| P4-09 | P4 | MES adapter — production order sync (work orders → outbound raw material) | pending | — | Production order intake; material consumption + BOM component reservation |
| P4-10 | P4 | MES adapter — finished goods receipt (production → inventory) | pending | — | Auto-create inbound ASN from production output; quality check gate |
| P4-11 | P4 | ERP adapter — purchase order → inbound ASN | pending | — | PO sync; auto-create expected ASN; GRN (Goods Receipt Note) back to ERP |
| P4-12 | P4 | ERP adapter — sales order → outbound order → shipment confirmation | pending | — | SO sync; auto-create outbound order; ship confirmation + tracking back to ERP |
| P4-13 | P4 | Integration retry policy + dead-letter handling | pending | — | Configurable retry with exponential backoff per adapter; DLQ storage for failed messages; replay capability; alert on DLQ threshold breach; blocks Phase 4 production readiness |

## Phase 5: Advanced Features

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P5-01 | P5 | Wave planning engine (batch, zone, carrier-based wave creation) | pending | — | Wave generation from orders; configurable grouping strategies; depends on P1-05 Task+Wave repo and P1-12 Order service |
| P5-02 | P5 | Inventory allocation engine (FIFO, FEFO, lot-specific, manual override) | pending | — | Auto-allocate inventory to order lines; reservation lifecycle; depends on P1-17 |
| P5-03 | P5 | Multi-level inventory (warehouse → zone → location → container/LPN) | pending | — | Container/LPN domain entity; nested inventory; container movements |
| P5-04 | P5 | Inventory reports (turnover, aging, ABC analysis) | pending | — | Compute and display inventory KPIs; depends on P5-18 for export |
| P5-05 | P5 | Operational reports (order fill rate, pick accuracy, cycle count variance) | pending | — | Operational KPI dashboards; date-range filtering; drill-down |
| P5-06 | P5 | RBAC permissions (role-based access control UI + API enforcement) | pending | — | Permission middleware; resource/action checks on every API call; depends on P2-10 |
| P5-07 | P5 | Audit log viewer (operation history, traceability, compliance export) | pending | — | Filterable audit log UI; date range; export for compliance audit |
| P5-08 | P5 | Inventory alerts engine (low stock, expiry, stranded, slow-moving) | pending | — | Configurable thresholds; notification channels (in-app, email, webhook); depends on P8-08 for notification delivery |
| P5-09 | P5 | Multi-warehouse support (cross-warehouse transfers, global inventory view) | pending | — | Transfer orders between warehouses; consolidated inventory dashboard |
| P5-10 | P5 | Lot/batch traceability (raw material → WIP → finished goods — full chain) | pending | — | Lot genealogy; forward and backward trace; recall support |
| P5-11 | P5 | Replenishment engine (min/max levels, demand-driven, auto task generation) | pending | — | Replenishment rules per SKU/zone; auto-create replenishment tasks when low |
| P5-12 | P5 | Dynamic slotting engine (velocity-based SKU → location optimization) | pending | — | ABC classification by pick frequency; auto-suggest optimal storage locations |
| P5-13 | P5 | Quality inspection workflow (QC checkpoints, sampling rules, hold/release) | pending | — | Sampling plans (AQL); inspection results; hold/release inventory; NCR tracking |
| P5-14 | P5 | Cross-docking flow (receiving → sort → ship, bypass storage) | pending | — | Identify cross-dock candidates; sort-by-destination; time-window management |
| P5-15 | P5 | Kitting / de-kitting (bundle and unbundle SKUs) | pending | — | Kit BOM definition; kit assembly tasks; component consumption; kit disassembly |
| P5-16 | P5 | Cartonization engine (optimal packaging per order) | pending | — | Box selection by item dimensions/weight; multi-carton split; packing slip generation |
| P5-17 | P5 | Scheduled report generation (cron-based, email delivery, PDF) | pending | — | Schedule daily/weekly/monthly reports; email with PDF attachment; report history; depends on P5-18 export engine + P8-08 notification + P21-03 email sending |
| P5-18 | P5 | CSV/PDF export engine (generic export for all list views) | pending | — | Streaming CSV writer for large datasets; PDF with header/logo; used by P5-04 and P5-05 |
| P5-19 | P5 | Pick path optimization (shortest path routing through warehouse) | pending | — | Compute optimal pick sequence through warehouse zones/locations; reduce travel distance per wave |
| P5-20 | P5 | Task interleaving (combine putaway + pick for same operator/zone) | pending | — | Merge putaway and pick tasks in same zone to minimize empty travel; operator efficiency gains |
| P5-21 | P5 | Putaway strategy engine (rule-based target location selection) | pending | — | Configurable strategies: nearest available, zone-fixed, ABC velocity-based; auto-select best location on receipt; respects capacity + segregation constraints |
| P5-22 | P5 | Cycle count scheduling engine (ABC-based frequency, auto task generation) | pending | — | Schedule count tasks by ABC class (A=monthly, B=quarterly, C=annually); location-based rotation; calendar-aware scheduling; auto-create tasks on schedule |
| P5-23 | P5 | Order splitting engine (split large orders across waves/zones by availability) | pending | — | Split single outbound order into multiple sub-waves when inventory is spread across zones; partial fulfillment tracking; split-by-zone and split-by-availability strategies; consolidation at shipping |
| P5-24 | P5 | Pick face management (forward pick area replenishment from reserve storage) | pending | — | Define pick face locations with min/max levels per SKU; auto-trigger replenishment tasks when pick face drops below min; different from general replenishment P5-11 which is inventory-level; zone-level pick face configuration |
| P5-25 | P5 | Serial number tracking (per-unit inventory granularity for high-value items) | pending | — | Serial number domain entity; serial-level inventory records (1:1 instead of qty); serial capture on receiving/picking/shipping; serial genealogy for traceability; opt-in per SKU (not all SKUs need serial tracking) |
| P5-26 | P5 | Full-text search (PostgreSQL tsvector/tsquery for orders, SKUs, locations) | pending | — | GIN-indexed tsvector columns on searchable entities; search across SKU code/name/description, order_no, location code; ranked results with highlighting; search API endpoint; requires DB migration for tsvector columns + triggers |
| P5-27 | P5 | SKU substitution rules (alternative SKUs when primary is out of stock) | pending | — | Substitution entity (primary SKU → alternate SKU); priority ranking; auto-substitution during allocation when primary unavailable; substitution approval toggle (auto vs manual review); substitution audit log |
| P5-28 | P5 | Catch weight management (variable-weight SKUs, scale integration) | pending | — | Catch weight flag on SKU; expected vs actual weight capture at receiving/shipping; weight tolerance thresholds; scale integration via serial/TCP; price-by-weight support; catch weight audit trail |
| P5-29 | P5 | Blind receiving workflow (receive without pre-advice, identify on dock) | pending | — | Scan barcode → identify SKU → capture qty → generate putaway; unknown barcode exception flow; mobile-first receiving screen for PDA; inventory created without ASN pre-notification; links to P3-05/P3-10 |
| P5-30 | P5 | Order consolidation engine (merge orders by destination/customer, wave grouping) | pending | — | Identify consolidatable orders (same customer, destination, carrier, ship-date); auto-merge into consolidated pick wave; consolidated packing (all items for one destination packed together); split back if consolidation fails; consolidation audit log; reduces pick+pack labor for multi-order customers |
| P5-31 | P5 | Container/LPN domain entity + CRUD (container types, hierarchy, inventory assignment) | pending | — | Container entity (type: pallet/tote/case, barcode, dimensions, tare weight); container-inventory relationship (which SKUs in which container); nested container hierarchy (cases on pallet); container movement as single unit; container status (empty, partial, full, in-transit); foundational for P5-03 multi-level inventory |
| P5-32 | P5 | Backorder management (short-pick tracking, auto-release on inventory arrival) | pending | — | Track unfulfilled order lines with backorder status; auto-detect when inventory arrives and matches backordered SKU; priority-based release rules (FIFO, priority, customer-tier); backorder aging report; manual release override; backorder notification to customer service |
| P5-33 | P5 | Space utilization analytics (cube/weight %, empty location %, occupancy trends) | pending | — | Compute utilization by zone: cube % (occupied volume / total volume), weight % (occupied weight / max weight), empty location %; utilization heatmap by zone/aisle; trend charts over time; what-if re-slotting impact analysis; utilization alerts (zone > 90% full) |
| P5-34 | P5 | Inbound QC inspection routing (per-SKU rules, AQL sampling, disposition routing) | pending | — | Per-SKU inspection rules (skip, spot-check, full-inspect); AQL sampling tables (ANSI/ASQ Z1.4); inspection result: accept → putaway, reject → quarantine, rework → rework area; inspection task generation on ASN arrival; inspection history per SKU/supplier; extends P5-13 quality inspection with inbound-specific routing |
| P5-35 | P5 | Order hold management (hold types, release workflows, SLA tracking) | pending | — | Hold types: fraud_review, credit_check, manual_review, damaged_goods, carrier_delay; hold/release API; hold SLA timer with auto-escalation; hold audit log; blocked status propagation to wave/task; release with optional inventory re-allocation |
| P5-36 | P5 | Inventory reservation expiration (timeout-based auto-release, configurable TTL) | pending | — | Configurable reservation TTL per order priority (urgent: 4hr, high: 8hr, normal: 24hr, low: 72hr); background job to expire stale reservations; release reserved_qty back to available; reservation expiry audit log; per-SKU override for high-demand items; depends on P5-02 allocation engine + P6-25 scheduled jobs |
| P5-37 | P5 | Barcode generation service (GS1-128, QR, DataMatrix, label preview) | pending | — | Generate GS1-128 with Application Identifiers (00 SSCC, 01 GTIN, 10 batch, 17 expiry); QR code for mobile scanning; DataMatrix for small items; SVG/PNG output via HTTP endpoint; label preview with barcode + human-readable text; used by P15-02 label template engine |
| P5-38 | P5 | Temperature/cold chain zone management (temp ranges, sensor integration, excursion alerts) | pending | — | Temperature range per zone/location (min/max °C); SKU storage temp requirements; putaway validation (reject if zone temp incompatible with SKU); IoT sensor integration point (receive temp readings via API); temperature excursion alert (out-of-range > N minutes); excursion audit log for compliance; cold chain hold/release workflow |
| P5-39 | P5 | Task pause/resume workflow (operational pause, SLA timer handling) | pending | — | Pause reasons: break, shift_change, equipment_issue, waiting_material; pause/resume API with timestamp tracking; SLA timer pause (stop clock during pause, resume on unpause); max-pause-duration auto-escalation (e.g., pause > 2hr → alert supervisor); pause audit log; TaskStatusPaused already defined in domain — this adds the operational workflow |
| P5-40 | P5 | In-transit inventory tracking (inter-warehouse shipments, ownership, ETA) | pending | — | In-transit inventory entity (source_warehouse, dest_warehouse, SKU, qty, carrier, tracking_no, ETA, status); deduct from source on ship, add to dest on receipt; in-transit aging report; lost-in-transit auto-flag (past ETA + grace period); ownership transfer point configuration (FOB origin vs destination); depends on P5-09 multi-warehouse support |

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
| P6-08 | P6 | OpenAPI/Swagger documentation (auto-generated from code annotations) | pending | — | swaggo or similar; /api/docs endpoint; publish to API docs portal; blocks P2-11 (TypeScript types generation) + P7-10 (API client SDK) + P10-06 (developer portal) |
| P6-09 | P6 | Database backup & restore tooling (pg_dump automation, point-in-time recovery) | pending | — | Cron job for backups; restore procedure documented; backup verification |
| P6-10 | P6 | Health check endpoints (liveness, readiness, DB/Redis connectivity) | pending | — | /livez (process alive), /readyz (DB+Redis ok), /healthz (combined) |
| P6-11 | P6 | gRPC server implementation (inter-service communication + integration adapters) | pending | — | gRPC server bootstrap; reflection; interceptors (auth, logging, tracing) |
| P6-12 | P6 | Graceful shutdown + connection draining | pending | — | SIGTERM handler; drain HTTP connections; close DB pool; flush logs |
| P6-13 | P6 | Terraform/IaC for cloud infrastructure (DB, cache, compute, networking) | pending | — | Terraform modules for AWS/GCP; state management; environment workspaces |
| P6-14 | P6 | Blue-green deployment strategy (zero-downtime rollout, smoke tests, rollback) | pending | — | Deployment automation; smoke test after cutover; automated rollback on failure |
| P6-15 | P6 | Log aggregation pipeline (stdout → Loki/ELK → searchable archive) | pending | — | DaemonSet collectors; structured log parsing; retention policies; log-based alerts |
| P6-16 | P6 | Configuration hot-reload (no-restart config updates) | pending | — | File watcher or SIGHUP handler; reload log level, rate limits, feature flags without restart; apply to running server within seconds |
| P6-17 | P6 | Database connection pool monitoring (pool stats, slow query detection) | pending | — | Expose pgxpool.Stat() via /metrics; track acquire latency, idle, max; slow query log threshold; connection exhaustion alert |
| P6-18 | P6 | K8s Horizontal Pod Autoscaling (CPU/memory-based + custom metrics) | pending | — | HPA manifest per deployment; target 70% CPU / 80% memory; custom metrics from Prometheus (request rate, task queue depth); min/max replica bounds per env; scale-down stabilization window |
| P6-19 | P6 | K8s NetworkPolicy + PodDisruptionBudget | pending | — | NetworkPolicy: allow only required pod-to-pod and pod-to-DB/Redis traffic; deny-by-default posture; PDB: maxUnavailable=1 for admin/pda; prevents full outage during voluntary disruptions |
| P6-20 | P6 | Container image vulnerability scanning (Trivy in CI pipeline) | pending | — | Trivy scan on every image build; fail CI on CRITICAL/HIGH CVEs; SBOM attestation; scan results in GitHub Security tab; periodic rescan of latest base images |
| P6-21 | P6 | Database read replica support (read/write split for reporting queries) | pending | — | Read replica connection pool in DB struct; Write() and Read() pool accessors; reporting endpoints use replica; repo-level read-vs-write hint; lag-aware health check; blocks P5-04/P5-05 report performance |
| P6-22 | P6 | K8s pod topology spread constraints (zone/region-aware pod placement) | pending | — | topologySpreadConstraints per deployment; maxSkew=1 across zones; requiredDuringScheduling for critical services; prevent all replicas in single AZ; combined with P6-18 HPA for resilience |
| P6-23 | P6 | Database migration in CI/CD pipeline (automated migration run, pre-check, rollback) | pending | — | Run migrations as pre-deploy step in CI; migration dry-run validation before apply; rollback plan on migration failure; migration lock to prevent concurrent runs; depends on P1-25 migration tooling |
| P6-24 | P6 | Service mesh readiness (Istio/Envoy sidecar annotations, mTLS, traffic routing) | pending | — | Pod annotations for Istio sidecar injection; mTLS strict mode between services; VirtualService for canary routing; DestinationRule for circuit breaking at mesh layer; Gateway for ingress; complements P7-16 circuit breaker at app layer |
| P6-25 | P6 | Scheduled job infrastructure (in-process cron scheduler, job registry, execution history) | pending | — | Job registry with registered job types; cron expression scheduling per job; execution history with status (success/failure/running) and duration; concurrency control (prevent overlapping runs of same job); retry on failure with backoff; needed by P5-22 cycle count scheduling, P5-17 scheduled reports, P14-03 data retention purge |
| P6-26 | P6 | Dependency update automation (Renovate/Dependabot config for Go modules + npm packages) | pending | — | Renovate config in repo (.renovaterc); auto-PR for Go module updates (minor/patch auto-merge, major manual review); npm package updates grouped by scope; security vulnerability PRs flagged CRITICAL; schedule: weekly for non-critical, immediate for CVEs; PR labels for changelog generation |
| P6-27 | P6 | SAST security scanning in CI (gosec for Go, ESLint security rules for frontend) | pending | — | gosec in GitHub Actions (fail on HIGH severity); ESLint-plugin-security for frontend; npm audit --audit-level=high in CI; results in GitHub Security tab (SARIF upload); baseline exceptions file for accepted risks reviewed quarterly; false-positive suppression with justification comments |
| P6-28 | P6 | Database restore drill automation (periodic restore to staging, integrity verification) | pending | — | Automated weekly restore of latest backup to staging environment; integrity checks (row counts, FK referential integrity, index health); RTO measurement (time from backup to ready); alert if restore fails or exceeds RTO threshold; drill report for compliance audit; depends on P6-09 backup tooling |
| P6-29 | P6 | Go pprof profiling endpoints (CPU, heap, goroutine, mutex profiles) | pending | — | Register net/http/pprof on a separate internal port (e.g., :6060) not exposed publicly; auth-protected via shared secret or mTLS; CPU profile sampling, heap allocation profile, goroutine dump, mutex contention; flame graph generation via go tool pprof; used for P7-07 performance tuning |

## Phase 7: Quality, Security & Hardening

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P7-01 | P7 | Domain unit test suite (all entities, state machines, business invariants) | pending | — | 80%+ coverage on internal/domain/; test negative qty, status transitions, validation; extends P1-20 to full coverage |
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
| P7-13 | P7 | Redis caching layer (hot inventory data, session storage, rate limit counters) | pending | — | Cache-aside for QueryInventory; TTL policies; cache invalidation on inventory change; depends on P1-24 Redis client; blocks P7-19 idempotency dedup store |
| P7-14 | P7 | Data export + import (CSV bulk import for SKUs, orders; data export for reporting) | pending | — | Bulk endpoints; streaming CSV parser; validation on import |
| P7-15 | P7 | Request ID propagation (HTTP header → context → log → gRPC → DB span) | pending | — | Extract/inject X-Request-ID at every boundary; structured log field; trace correlation |
| P7-16 | P7 | Circuit breaker for integration adapters (WCS/RCS/MES/ERP failure isolation) | pending | — | Circuit breaker per adapter; half-open probing; fallback behavior; blocks Phase 4 integration reliability |
| P7-17 | P7 | Secrets management (vault integration, encrypted config, no secrets in code) | pending | — | Externalize all secrets; HashiCorp Vault or cloud secrets manager; CI secret scanning |
| P7-18 | P7 | Go fuzz testing (fuzz input parsers, validators, JSON unmarshal paths) | pending | — | go test -fuzz for CSV parser, JSON payloads, barcode validator; catch panics and edge cases |
| P7-19 | P7 | Idempotency key support (safe retry for mutating endpoints) | pending | — | Idempotency-Key header handling; dedup store (Redis-backed, depends on P7-13); return cached response on replay; applies to order create, inventory adjust, task complete |
| P7-20 | P7 | Feature flag system (runtime toggles, percentage rollout) | pending | — | Flag definitions in config/DB; per-request evaluation; admin UI for flag management; gradual rollout for risky changes; emergency kill-switch capability |
| P7-21 | P7 | Bulk API operations (batch create/update for high-volume endpoints) | pending | — | Bulk create SKUs, bulk inventory adjust; partial success response format; streaming request body; applicable to orders, inventory, SKUs |
| P7-22 | P7 | Optimistic concurrency control (row version column for concurrent write safety) | pending | — | Add `version` int column to inventory, orders, tasks; increment on update; WHERE version=$expected; detect conflict (RowsAffected=0); return 409 Conflict on concurrent modification; critical for inventory correctness under concurrent pick/putaway; builds on P1-16 tx patterns |
| P7-23 | P7 | Soft delete implementation (deleted_at timestamps, restore capability) | pending | — | Add deleted_at TIMESTAMPTZ to all major entities; soft-delete filter in base queries; restore endpoint; cascade rules (delete order → soft-delete lines); purge policy for GDPR compliance (P14-03); critical for WMS undo and audit retention |
| P7-24 | P7 | Test data factories (deterministic entity builders for all domain types) | pending | — | Go factory functions per domain entity; deterministic ID generation from seed; builder pattern for test customization; shared via internal/testutil/factory; used by P7-01, P7-02, P7-03; reduces test boilerplate and improves test reliability |
| P7-25 | P7 | Contract testing (API compatibility verification between services) | pending | — | Pact or similar consumer-driven contract tests; admin API consumer → server contract; PDA API consumer → server contract; verify in CI on every PR; prevents accidental API breakage between frontend and backend |
| P7-26 | P7 | API response compression (gzip/brotli middleware, content negotiation) | pending | — | chi compression middleware; gzip level 6 default, brotli for supporting clients; skip for already-compressed types (images); configurable min body size threshold; reduces bandwidth for mobile PDA on warehouse WiFi |
| P7-27 | P7 | Request validation middleware (input sanitization, content-type enforcement) | pending | — | Validate Content-Type header; enforce max body size (1MB default); sanitize string inputs (trim, null-byte check); reject unknown JSON fields (strict mode); extend P1-15 error handling with validation error details |
| P7-28 | P7 | Graceful degradation framework (degraded mode when dependencies are unhealthy) | pending | — | Degraded mode tiers: cache-down (skip Redis, direct DB), replica-down (route reads to primary), notification-down (queue alerts for later); auto-detect health via P6-10 checks; auto-recovery when dependency restores; admin alert on degradation; per-endpoint degradation behavior config |
| P7-29 | P7 | Database table partitioning (partition inventory_transactions + audit_logs by month) | pending | — | Declarative partitioning by month on high-volume tables; automated partition creation via cron; partition-aware queries (pruning); partition archival/detach for old data; complements P14-03 data retention; requires migration for partition setup |
| P7-30 | P7 | Application-level deadlock detection + retry (concurrent resource ordering) | pending | — | Detect deadlock cycles in concurrent inventory/location acquisition; deterministic resource ordering to prevent deadlocks; transaction retry with exponential backoff + jitter on deadlock; deadlock event logging for diagnostics; builds on P1-16 tx patterns + P7-22 optimistic locking |
| P7-31 | P7 | API versioning strategy (URL-based versioning, deprecation headers, sunset policy) | pending | — | URL-based versioning: /api/v1/ → /api/v2/; Sunset and Deprecation HTTP headers on deprecated endpoints; version deprecation policy (N-1 support: current + one previous); API changelog page; backwards-compatible changes within a version; breaking changes require new version |
| P7-32 | P7 | API key management for integrations (CRUD, scoped permissions, rotation, tracking) | pending | — | API key entity (hashed storage, reveal-once on creation); scoped permissions per key (read-only, specific resources); key rotation with overlap window; expiry with auto-disable; usage tracking (last used, request count); rate limit per key; needed for Phase 4 integrations + Phase 10 API ecosystem |
| P7-33 | P7 | Two-factor authentication (TOTP-based 2FA, enforced for admin roles, backup codes) | pending | — | TOTP setup wizard (QR code scan, verification); enforced per role (require 2FA for admin/manager roles); backup recovery codes (one-time use, bcrypt-hashed); remember-device option (30-day trust); 2FA challenge on login from new IP; account recovery flow for lost 2FA device; depends on P1-19 auth service |
| P7-34 | P7 | Session management hardening (concurrent limits, force-logout, device tracking) | pending | — | List active sessions per user with IP/device/user-agent; force-logout individual or all sessions; configurable concurrent session limit per role; device/browser fingerprint for anomaly detection; suspicious activity alerts (new country/IP); session idle timeout with sliding expiration; depends on P1-19 auth + P7-13 Redis cache for session store |
| P7-35 | P7 | Security.txt + vulnerability disclosure policy | pending | — | security.txt at /.well-known/security.txt per RFC 9116; vulnerability reporting contact and PGP key; disclosure policy (coordinated disclosure, safe harbor, response SLA); security contact page in admin UI; linked from API docs and README |

## Phase 8: Observability & Operations

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P8-01 | P8 | Grafana dashboards (pre-built WMS KPIs: order volume, pick rate, accuracy) | pending | — | Dashboard JSON in repo; importable via Grafana provisioning; depends on P6-05 metrics |
| P8-02 | P8 | AlertManager rules + notification routing (PagerDuty, Slack, email) | pending | — | Alert rules: service down, DB pool exhaustion, error rate spike, task backlog; severity routing; depends on P8-08 for notification delivery |
| P8-03 | P8 | SLO/SLI definitions + tracking (API latency, availability, throughput) | pending | — | Define SLOs (e.g., 99.9% API availability, p95 < 500ms); error budget burn alerts |
| P8-04 | P8 | Incident runbook documentation (common failure modes, recovery steps) | pending | — | Runbook per service; DB failover procedure; integration adapter recovery; escalation paths |
| P8-05 | P8 | Synthetic monitoring (external health probes simulating user flows) | pending | — | Blackbox exporter probes for critical API paths; login → dashboard → order create flow |
| P8-06 | P8 | Cloud resource tagging + cost allocation (per-environment cost tracking) | pending | — | Standard tags (env, service, owner); cost dashboards; unused resource detection |
| P8-07 | P8 | Chaos engineering baseline (controlled failure injection, resilience validation) | pending | — | Kill a DB replica, kill a pod, network partition; verify graceful degradation and recovery |
| P8-08 | P8 | Notification infrastructure (email + in-app notification delivery) | pending | — | SMTP email service with Go templates; in-app notification center (persisted, mark-read, real-time via WebSocket); used by P5-08 alerts + P8-02 alertmanager + P19-04 escalation; blocks P8-02 delivery channel; depends on P21-03 email sending service for email channel |
| P8-09 | P8 | SLO tracking & error budget dashboard (p50/p95/p99 per endpoint, burn rate) | pending | — | Grafana dashboard with per-endpoint latency histogram; SLO compliance gauges; error budget remaining/burn rate; 30-day rolling windows; depends on P6-05 metrics + P6-07 tracing; complements P8-03 SLO definitions |
| P8-10 | P8 | OpenTelemetry log-trace correlation (inject trace_id/span_id into structured logs) | pending | — | Inject OTel trace_id and span_id into all structured log entries; log correlation across services via trace_id; Grafana Loki→Tempo trace linking; log sampling by trace (keep all logs for error traces, sample healthy traces); depends on P6-06 structured logging + P6-07 distributed tracing |
| P8-11 | P8 | Integration adapter health monitoring (per-adapter dashboard, connection status, throughput) | pending | — | Per-adapter health metrics: connection status (up/down/degraded), message throughput (msgs/sec), error rate, last heartbeat timestamp, DLQ depth; Grafana dashboard row per adapter; alert on adapter down > 2min or DLQ depth > threshold; depends on P4-13 DLQ + P6-05 metrics + P8-01 Grafana |

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
| P9-10 | P9 | Equipment/forklift management (equipment entity, maintenance schedule, certification tracking) | pending | — | Equipment entity (type: forklift/reach-truck/pallet-jack/scanner/printer, barcode, status); maintenance schedule with calendar integration (P15-05); operator certification tracking (which operators certified for which equipment); equipment-to-task assignment; maintenance due alerts; utilization tracking per equipment |
| P9-11 | P9 | Customs documentation for international shipping (commercial invoice, packing list, certificates) | pending | — | Commercial invoice generation from order data (HS codes, country of origin per SKU); packing list with weights/dimensions per package; certificate of origin template; customs document templating (Go html/template); carrier-specific customs forms (FedEx/UPS/DHL electronic submission); document version tracking per shipment |
| P9-12 | P9 | Dock appointment scheduling (carrier/supplier self-service booking, time-slot, dwell tracking) | pending | — | Dock door calendar with configurable time slots (e.g., 30min blocks); carrier/supplier self-service booking portal (linked to P18-01 supplier portal); check-in/check-out workflow with timestamp tracking; dwell time tracking and alerts (dock occupancy > scheduled window); overbooking prevention per dock; recurring appointment templates; distinct from P9-01 yard management (trailer tracking) — this is the scheduling layer |

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
| P11-06 | P11 | Analytics data mart (star schema for operational reporting, ETL from OLTP) | pending | — | Star schema in separate analytics schema (fact_orders, fact_inventory_txns, dim_sku, dim_warehouse, dim_date, dim_customer); nightly ETL from OLTP to analytics tables via postgres_fdw or dbt; materialized views for common KPI dashboards (daily_order_volume, sku_velocity, zone_utilization); isolates reporting query load from operational DB; depends on P5-04/P5-05 reports; enables P11-03 ML forecasting with clean data |

## Phase 12: Internationalization & Localization

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P12-01 | P12 | i18n framework setup (react-i18next, locale detection, translation file structure) | pending | — | Shared i18n config for admin + PDA; locale auto-detection from browser/navigator; translation namespace organization; blocks P12-02 through P12-05 (translations depend on framework) |
| P12-02 | P12 | Chinese (zh-CN) translation — Admin UI | pending | — | Full translation of all admin pages, forms, tables, notifications; most common operator locale in APAC |
| P12-03 | P12 | Japanese (ja-JP) translation — Admin UI | pending | — | Full translation of admin UI; important for Japan-market warehouses |
| P12-04 | P12 | PDA UI translations (zh-CN, ja-JP) | pending | — | Mobile-optimized translations for PDA flows; short labels for small screens |
| P12-05 | P12 | Date/time/number formatting by locale (dayjs locale integration) | pending | — | Locale-aware date formats, number separators, timezone display; consistent across all UI components |
| P12-06 | P12 | RTL (right-to-left) layout foundation | pending | — | CSS logical properties; direction-aware components; Arabic/Hebrew readiness (UI only, no translation yet) |
| P12-07 | P12 | API error message i18n (locale-aware error responses via Accept-Language header) | pending | — | Translation files for API error messages in zh-CN, ja-JP; Accept-Language header parsing; fallback to en; error code remains stable across locales (only message text changes); error translation namespace separate from UI translations; used by PDA (operator-facing errors) + integration adapters (machine-readable codes unchanged) |

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
| P13-08 | P13 | Architecture Decision Record (ADR) template + automated status checks | pending | — | Standardized ADR template (status, context, decision, consequences); ADR directory under docs/adr/; ADR status lifecycle (proposed → accepted → superseded → deprecated); automated check: PRs that add architecture changes must include ADR; existing ADR 001 already in repo — this formalizes the process |
| P13-09 | P13 | Changelog generation from conventional commits (git-cliff config, automated release notes) | pending | — | git-cliff config in repo; categorize by conventional commit type (feat→Features, fix→Bug Fixes, refactor→Refactoring); group by scope; GitHub release integration (auto-populate release notes); CHANGELOG.md updated on every release; depends on P13-02 pre-commit conventional commit check |

## Phase 14: Compliance & Data Governance

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P14-01 | P14 | GDPR data export (user data export endpoint, right-to-access compliance) | pending | — | Aggregate all PII per user; JSON export endpoint; audit log of access requests |
| P14-02 | P14 | GDPR data deletion (right-to-erasure, anonymization cascade) | pending | — | Anonymize or delete user PII while preserving inventory audit integrity; cascade rules per entity |
| P14-03 | P14 | Data retention policies (configurable retention per entity type, auto-purge) | pending | — | Retention config (e.g., audit_logs: 7yr, inventory_transactions: 3yr); scheduled purge job; dry-run mode |
| P14-04 | P14 | PII data classification (identify and tag PII fields, masking in logs) | pending | — | Tag PII fields in domain models; log masking middleware; data flow diagram of PII; compliance documentation |
| P14-05 | P14 | Audit compliance report (pre-built report for SOC2/ISO27001 audit evidence) | pending | — | Pre-built report covering access controls, change management, data encryption, backup verification; exportable as PDF |
| P14-06 | P14 | Consent management (cookie consent banner, data processing consent records) | pending | — | Consent UI for admin/PDA; record consent timestamp + version; consent withdrawal flow; privacy policy page |
| P14-07 | P14 | License compliance scanning (Go + npm dependencies, attribution doc generation) | pending | — | Scan go.mod + package.json for license types; flag copyleft/GPL licenses for legal review; generate NOTICE file with all dependency attributions; CI check: fail build on prohibited license addition; quarterly license audit report; SPDX license identifiers throughout |

## Phase 15: Labeling, Printing & Carrier Management

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P15-01 | P15 | Carrier management (shipping carriers, service levels, accounts) | pending | — | Carrier entity (name, SCAC code, service levels, transit days); account credentials; rate tables; carrier assignment rules per order; tracking number integration |
| P15-02 | P15 | Label template engine (ZPL/EPL output, GS1-128, configurable templates) | pending | — | Template designer for barcode labels; GS1-128 application identifiers; ZPL/EPL printer output; templates per document type (location, SKU, pallet, ship); printer registry |
| P15-03 | P15 | Document printing service (pick list, pack slip, bill of lading, ship label) | pending | — | Print queue with status tracking; reprint capability; batch print for waves; PDF generation for non-label docs; printer assignment by zone/workstation |
| P15-04 | P15 | Reason code management (standardized codes for exceptions, adjustments, holds) | pending | — | Reason code entity (code, category, description); categories: adjustment, exception, QC hold, return disposition, damage; API for PDA/Admin lookups; blocks consistent exception handling in P3-10 |
| P15-05 | P15 | Holiday calendar & shift management (working days, shift schedules, SLA) | pending | — | Calendar entity per warehouse; holiday dates; shift definitions (morning/night, start/end); working hours; SLA calculation uses calendar (skip holidays, respect shifts); blocks P5-22 cycle count scheduling accuracy |
| P15-06 | P15 | Document number sequence engine (configurable prefixes, auto-increment) | pending | — | Sequence definition per document type (order, ASN, task, wave); configurable prefix + date component + counter; yearly/monthly/daily reset option; gapless mode for compliance; replaces inline number generation in repos |
| P15-07 | P15 | Wireless printer support (Bluetooth/WiFi printer discovery + printing from PDA) | pending | — | Web Bluetooth API for printer discovery on PDA; connect to Zebra/SATO/Brother mobile printers; print labels directly from PDA receiving/putaway flow; printer pairing management; fallback to server-side printing when direct connect unavailable; depends on P15-02 label template engine |

## Phase 16: Mobile/PDA Enhancements

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P16-01 | P16 | PDA push notifications (Web Push API for task assignment, alerts) | pending | — | Service worker push event handler; subscribe on PDA login; notify on new task assigned, task overdue, exception flagged; vibration/sound patterns per priority; permission management in settings |
| P16-02 | P16 | PDA voice input (Web Speech API for hands-free quantity entry + confirmation) | pending | — | Speech recognition for quantity entry ("five"), location codes ("A-01-02-03"), task confirmation ("confirm putaway"); wake-word "Hey WMS" for hands-free operation; language selection (en, zh, ja); fallback to manual input when speech unavailable |
| P16-03 | P16 | PDA NFC tag support (equipment check-in, location verification, tap-to-confirm) | pending | — | Web NFC API for reading NDEF tags; equipment check-in/out via NFC tag on forklift/cart; location verification (tap location tag = confirm you're at right place); tap-to-confirm for task completion; graceful fallback when NFC unavailable |
| P16-04 | P16 | PDA geofencing (zone-based auto task filtering, wrong-zone alerts) | pending | — | Geolocation API with warehouse zone polygons; auto-filter task list to current zone; wrong-zone alert when starting task outside expected zone; zone-entry/exit time logging for labor tracking; low-accuracy mode via BLE beacons for indoor positioning |
| P16-05 | P16 | PDA camera/image capture (photo documentation + barcode fallback) | pending | — | Camera capture via MediaDevices API; attach photos to receiving (damage evidence, pack condition); attach photos to exception reports; image compression before upload; photo gallery per task/receipt; barcode scanning from still image as fallback when live scanner fails; offline-queue photo uploads via P3-11 |

## Phase 17: Simulation & Digital Twin

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P17-01 | P17 | Warehouse layout designer (SVG-based drag-and-drop zone/location editor) | pending | — | Visual warehouse map editor; drag zones and locations onto grid; configure aisles, racks, conveyors; export layout as JSON config; import from existing warehouse data; measurement tools (distance, area); depends on warehouse/zone/location domain entities |
| P17-02 | P17 | Operations simulation engine (configurable order patterns, throughput modeling) | pending | — | Order pattern generator (configurable mix of inbound/outbound, SKU distribution, velocity); simulate pick/putaway/replenish cycles; measure throughput, utilization, queue depth, bottlenecks; configurable operator count and shift patterns; replay historical order data |
| P17-03 | P17 | What-if scenario comparison (layout changes, equipment, staffing scenarios) | pending | — | Define scenarios (baseline vs proposed); run simulations on each; side-by-side metrics comparison (throughput, labor hours, utilization); sensitivity analysis on key parameters; scenario library with saved baselines; report export |
| P17-04 | P17 | Digital twin dashboard (real-time warehouse state mirror, 3D visualization) | pending | — | Real-time reflection of warehouse state: inventory heatmap overlay on SVG layout, active tasks animated (pick paths, AGV movements), zone occupancy gauges, bottleneck highlighting; WebSocket-driven live updates (P4-02); time-slider for replay (last hour/day/week); depends on P17-01 warehouse layout designer + P11-02 live ops dashboard |

## Phase 18: Supplier & Partner Enablement

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P18-01 | P18 | Supplier self-service portal (ASN submission, label printing, shipment visibility) | pending | — | Supplier-facing web portal (separate from admin); supplier login with role-limited access; submit ASN with line items; print barcode labels (depends on P15-02); view shipment receiving status; view purchase order history; multi-tenant (supplier sees only their data) |
| P18-02 | P18 | Supplier performance scorecards (on-time delivery, accuracy, quality metrics) | pending | — | KPI computation per supplier: OTD%, fill rate, damage rate, documentation accuracy; scorecard dashboard with trend charts; configurable evaluation periods; supplier ranking; export for QBR (quarterly business review); depends on P5-04/P5-05 reporting |
| P18-03 | P18 | 3PL billing engine (activity-based billing, storage charges, invoice generation) | pending | — | Billable event capture (receipt, putaway, pick, pack, ship, storage-day); rate card per client (different rates per activity); storage charges by volume × days; minimum charges and tiered pricing; invoice generation with line-item detail; extends P9-06 3PL support with billing specifics |
| P18-04 | P18 | Customer inventory visibility portal (external view of owned inventory + orders) | pending | — | Customer-facing portal for 3PL clients; view own inventory levels by SKU/warehouse; order status tracking; shipment history with tracking numbers; stock arrival notification preferences; data scoped to client ownership only (security-critical); read-only access |

## Phase 19: Workflow & Approval Engine

> Approval workflows are essential for production WMS operations — inventory adjustments, order cancellations, and exceptions all need configurable approval chains before execution.

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P19-01 | P19 | Approval workflow engine (configurable multi-step approval chains) | pending | — | Domain entity: ApprovalRequest (resource type, resource ID, action, requested_by, status); ApprovalStep (step order, approver role/user, decision, comment); configurable workflow templates per action type; auto-approve on no matching approver; delegation support (approver can delegate); approval deadline with auto-reject; audit log of all decisions |
| P19-02 | P19 | Inventory adjustment approval workflow | pending | — | Define thresholds: qty-change < X% → auto-approve, < Y% → supervisor approval, ≥ Y% → manager approval; high-value SKU flag triggers mandatory approval regardless of qty; batch approval for cycle count variance reconciliation; depends on P19-01 |
| P19-03 | P19 | Order cancellation approval workflow | pending | — | Cancel request captures reason + impact (qty picked, inventory state); auto-approve if no inventory allocated yet; requires approval if picking has started; inventory restoration validation on cancel; customer credit check integration point; depends on P19-01 |
| P19-04 | P19 | Exception escalation workflow (SLA-based escalation, supervisor routing) | pending | — | Exception types linked to escalation paths (P15-04 reason codes); SLA timer per exception severity (critical: 5min, high: 15min, normal: 1hr); auto-escalate to next level on SLA breach; supervisor assignment by zone/shift (P15-05 calendar-aware); notification via P16-01 push + P8-08 email; resolution audit trail; depends on P19-01 + P15-04 |

## Phase 20: Data Migration & Onboarding

> Enterprise WMS adoption requires migrating data from legacy systems. This phase provides the tooling and processes for safe, validated data migration.

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P20-01 | P20 | Data migration framework (extract → validate → transform → load pipeline) | pending | — | Pipeline stages: Extract (CSV/JSON/DB reader) → Validate (schema + business rules) → Transform (field mapping, ID remapping) → Load (batch insert with transaction safety); progress tracking per entity type; error collection (row-level errors, not fail-all); rollback on critical failure; migration run metadata (started, completed, entity counts) |
| P20-02 | P20 | SKU master + inventory opening balance import | pending | — | CSV template with field mapping config; SKU dedup strategy (by code, by barcode); opening balance import: create inventory records with "migration" transaction type; batch insert with COPY protocol for speed; validation: required fields, UOM consistency, barcode format; linked to P20-01 framework; dry-run mode for validation without writing |
| P20-03 | P20 | Warehouse structure import (warehouses, zones, locations bulk import) | pending | — | Hierarchical import: warehouse → zones → locations; location barcode generation if missing; capacity validation; zone-type consistency check; parent-child referential integrity; location code format validation; depends on P20-01 |
| P20-04 | P20 | Open order import (in-progress orders at cutover point) | pending | — | Import partially-fulfilled orders at go-live; preserve original order_no + external_ref; line-level fulfilled_qty import; task generation for remaining work; inventory state reconciliation post-import; critical path — must be validated with P20-05 dry-run before go-live |
| P20-05 | P20 | Migration dry-run + validation report (pre-cutover verification suite) | pending | — | Execute full migration pipeline against staging; automated validation: row counts match, inventory qty sums balance, location occupancy consistent, order line sums correct; diff report (source vs target); performance benchmark (estimated cutover window); rollback test (restore from backup, verify clean state); pre-cutover checklist generation |

## Phase 21: Email & Communications

> Transactional emails are critical for WMS operations — order confirmations, shipment notifications, and alert delivery all need reliable email infrastructure.

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P21-01 | P21 | Email template engine (Go html/template, variable interpolation, i18n-aware) | pending | — | HTML email templates with Go html/template; variable substitution from domain events (order, shipment, alert); layout/base template with header/footer; i18n template lookup by locale; plaintext fallback generation; preview mode (render template with sample data in admin UI); template versioning |
| P21-02 | P21 | Transactional email templates (order confirm, ship notify, alert, report, welcome) | pending | — | Pre-built templates: order confirmation (inbound/outbound), shipment notification with tracking, inventory alert (low stock, expiry), scheduled report delivery (P5-17), user welcome + password reset; responsive HTML tested across email clients; unsubscribe header compliance; depends on P21-01 |
| P21-03 | P21 | Email sending service (SMTP + provider abstraction, delivery tracking) | pending | — | Provider interface: SMTP + SendGrid/Resend/Mailgun implementations; connection pooling for SMTP; delivery status webhook handling (delivered, bounced, complained, opened); retry with backoff on transient failures; send rate limiting per provider; email send audit log; blocks P8-08 notification delivery + P5-17 report email delivery |

## Phase 22: User Experience & Accessibility

> Accessibility and UX polish ensure the WMS is usable by all operators and administrators, including those with disabilities.

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P22-01 | P22 | Admin UI accessibility audit + remediation (WCAG 2.1 AA compliance) | pending | — | Keyboard navigation for all interactive elements (tables, forms, modals, dropdowns); screen reader labels (aria-label, aria-describedby, role attributes); focus management (focus trapping in modals, focus restoration on navigation); color contrast ≥ 4.5:1 for text, ≥ 3:1 for large text; skip-to-content link; axe-core or Lighthouse CI check in PR pipeline; accessible form validation errors |
| P22-02 | P22 | PDA UI accessibility (large touch targets, screen reader, high contrast) | pending | — | Minimum touch target size 44×44px (WCAG 2.5.5); high-contrast mode toggle for warehouse environments (bright/dim lighting); screen reader announcements for barcode scan results ("SKU 12345 scanned, qty 10"); vibration alternatives → visual flash + sound for alerts; reduced motion support (prefers-reduced-motion); font size controls (warehouse operators may have varying vision needs) |
| P22-03 | P22 | Admin UI dark mode + theme system | pending | — | Ant Design 5 ConfigProvider theme with dark/light/system toggle; persisted preference in localStorage; CSS variable-based design tokens for custom theming (brand colors, spacing, border-radius); automatic dark mode based on system preference on first visit; theme context provider for consistent access across admin + PDA |
| P22-04 | P22 | Admin UI responsive layout (tablet-functional admin dashboard) | pending | — | Collapsible sidebar with icon-only mode for small screens; responsive data tables → card/list view below breakpoint; dashboard KPI cards reflow to grid; form layouts single-column on mobile; navigation drawer for mobile; priority: functional on tablet (warehouse managers on floor), not optimized for phone |

## Phase 23: Training, Onboarding & Documentation

> Operator and administrator onboarding is critical for WMS adoption. This phase provides guided training, setup wizards, and embedded help.

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P23-01 | P23 | Operator training mode (guided PDA walkthroughs, progress tracking, certification) | pending | — | Training mode toggle per user; guided walkthrough for each PDA flow (receiving, putaway, picking, cycle count, shipping) with step-by-step overlay instructions; practice mode with test data (no real inventory impact); progress tracking per operator (which flows completed); supervisor sign-off on training completion; training certification records with expiry (re-certification reminders) |
| P23-02 | P23 | Admin onboarding wizard (guided initial setup: warehouse → zones → locations → SKUs → users) | pending | — | Multi-step wizard with progress indicator; Step 1: create first warehouse; Step 2: define zones (pre-built templates: standard 4-zone vs advanced); Step 3: generate location grid (aisle/rack/bin pattern); Step 4: import or create first SKUs (CSV upload or manual); Step 5: create admin users + assign roles; can skip steps and complete later; onboarding completion checklist on dashboard |
| P23-03 | P23 | Context-sensitive help + embedded documentation hub | pending | — | "?" help icon on every admin page linking to relevant docs section; embedded short tutorial GIFs/videos for common operations (create order, adjust inventory, assign task); searchable help drawer; link to full documentation site; help content versioned alongside app version; PDA help accessible from bottom tab (quick reference for task flows); depends on P6-08 OpenAPI docs for API reference section |

## Evolution Metrics

| Metric | Value |
|--------|-------|
| Total tasks | 272 |
| Completed | 9 |
| In progress | 0 |
| Pending | 263 |
| Success rate | — |
| Started | 2026-07-20 |
| Last evolution | 2026-07-20 (Round 3: P1-03 SKU+Inventory repos) |
| Last grooming | 2026-07-20 (Round 18: updated 4 task notes for accuracy; added 16 new tasks — P6-28 restore drills, P6-29 pprof, P7-33 2FA, P7-34 session hardening, P7-35 security.txt, P8-11 adapter health, P9-12 dock scheduling, P11-06 analytics mart, P12-07 API i18n errors; new Phase 22 UX/Accessibility (4 tasks), new Phase 23 Training/Onboarding (3 tasks)) |
