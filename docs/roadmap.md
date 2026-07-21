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
| P3-04 | P2 | Admin: Inventory dashboard | pending | — | Summary cards, low stock, zone breakdown |
| P3-05 | P2 | Admin: Order management pages | pending | — | Table with status badges, detail view |
| P3-06 | P1 | Token blacklist / logout | pending | — | Invalidate refresh tokens on logout |
| P3-07 | P2 | Health check endpoints (/health, /ready) | completed | 2026-07-21 | `/health` exists in both admin + PDA servers; no DB ping readiness check yet |
| P3-08 | P2 | PDA: Login + task list screen | pending | — | Mobile login, swipe-to-refresh task list |
| P3-09 | P2 | Redis client bootstrap | pending | — | Blocks rate limiting, caching, session store |
| P3-10 | P2 | Migration tracking table | pending | — | schema_migrations so each .sql runs once |

<!-- DISCOVER refills when pending < 3. Last trim: 2026-07-21 (9→8, removed completed P3-07). -->
