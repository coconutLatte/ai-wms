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

## Phase 2: Next Up (max 10 pending)

| ID | Priority | Task | Status | Completed | Notes |
|----|----------|------|--------|-----------|-------|
| P2-01 | P1 | DB transaction support for atomic inventory operations | pending | — | Wrap inventory change + location update + audit in single DB tx |
| P2-02 | P1 | Pagination metadata for all list endpoints | pending | — | Total count + page_token; every list API needs this |
| P2-03 | P1 | Domain unit tests (state machines, business rules) | pending | — | Pure Go tests for Order/Task status transitions, Inventory invariants |
| P2-04 | P1 | Authentication (JWT login, token refresh, middleware) | pending | — | Blocks admin login page |
| P2-05 | P1 | Makefile: run-admin, run-pda, migrate targets | pending | — | Dev workflow; build/test/lint already work |
| P2-06 | P1 | Seed data script (demo warehouse, zones, SKUs) | pending | — | Enables UI development; basic seed already in migration |
| P2-07 | P2 | Admin frontend scaffold (React + Ant Design + routing) | pending | — | Layout, navigation, theme, API client |
| P2-08 | P2 | Admin: Warehouse management pages (list, create, edit) | pending | — | Warehouse + zone + location CRUD UI |
| P2-09 | P2 | FEFO/FIFO inventory retrieval queries | pending | — | GetOldestInventory / GetExpiringInventory methods |
| P2-10 | P2 | PDA frontend scaffold (React + mobile-first) | pending | — | Mobile layout, barcode scanner component, task list |

<!-- DISCOVER will refill when pending < 3 -->
