# AI-WMS Architecture

## System Overview

AI-WMS is a self-evolving Warehouse Management System designed for continuous AI-driven development. It manages warehouse operations across web-based admin and PDA interfaces, integrating with hardware control systems (WCS/RCS) and enterprise business systems (MES/ERP).

## Architectural Principles

1. **Domain-Driven Design** — The domain model is the single source of truth. All business rules live in `internal/domain/`.
2. **Ports & Adapters** — External systems (WCS, RCS, MES, ERP) connect through defined adapter interfaces.
3. **Evolution-First** — Every design decision accounts for AI-driven incremental development. Simplicity over cleverness.
4. **Compile-Time Safety** — Go's type system catches AI-generated errors before runtime.
5. **Testability** — Interfaces enable mocking; pure domain logic is trivially testable.

## Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  ┌──────────────────────┐  ┌──────────────────────────┐    │
│  │  Admin UI (React)    │  │  PDA UI (React Mobile)   │    │
│  └──────────┬───────────┘  └──────────┬───────────────┘    │
│             │                         │                     │
│             ▼                         ▼                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              API Layer (HTTP/gRPC)                    │  │
│  │  chi/v5 REST API  │  gRPC Services  │  WebSocket      │  │
│  └────────────────────────┬─────────────────────────────┘  │
│                           │                                 │
│                           ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Service Layer                            │  │
│  │  WarehouseSvc │ InventorySvc │ OrderSvc │ TaskSvc     │  │
│  └────────────────────────┬─────────────────────────────┘  │
│                           │                                 │
│                           ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Domain Layer (Pure, Zero-Dependency)      │  │
│  │  Warehouse │ Zone │ Location │ SKU │ Inventory        │  │
│  │  Order │ ASN │ Task │ Wave │ User │ Role             │  │
│  └────────────────────────┬─────────────────────────────┘  │
│                           │                                 │
│                           ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Infrastructure Layer                     │  │
│  │  PostgreSQL │ Redis │ gRPC │ REST │ Message Queue     │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Integration Adapters                     │  │
│  │  WCS Adapter │ RCS Adapter │ MES Adapter │ ERP Adapter│  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Go Backend (Monolith → Modular Monolith → Microservices)
- **Now (Phase 0-3)**: Single Go module, multiple entry points (`cmd/admin`, `cmd/pda`)
- **Later (Phase 4+)**: Extract integration adapters as separate services if needed
- **Why**: Monolith is simpler for AI to reason about during early evolution; we split when the boundary is clear

### 2. chi/v5 over Gin/Fiber
- chi is standard-library compatible (`net/http` handlers)
- No magic — explicit middleware chains
- Lighter than Gin, more idiomatic

### 3. PostgreSQL as Primary Store
- ACID transactions for inventory accuracy (critical)
- JSONB for flexible SKU attributes
- Mature Go driver (pgx/v5)

### 4. UUID Primary Keys
- AI-generated code doesn't need to manage sequences
- Safe for distributed systems from day one
- Predictable pattern for AI: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`

### 5. Inventory Double-Entry
- Every inventory change creates an `inventory_transactions` row
- `qty` = sum of all `delta_qty` for that inventory record (eventual consistency via DB trigger or service)
- Full audit trail

### 6. Security Middleware Stack
- **RequestID** — Extracts or generates a request ID (X-Request-ID header), stored in context for tracing
- **Recovery** — Catches panics, logs stack traces, returns 500
- **Logger** — Structured request logging (method, path, status, duration, request_id)
- **CORS** — Configurable cross-origin headers including preflight handling
- **Auth** — JWT Bearer token validation; injects user_id, username, role_ids, and role_names into request context
- **RequireRole** — RBAC middleware checking role names from JWT claims; returns 403 if unauthorized

Auth flow: `RequestID → Recovery → Logger → CORS → Auth → RequireRole → Handler`

Admin-only routes (`/api/v1/users`, `/api/v1/audit-logs`) are wrapped in `RequireRole("admin")`.

## API Design

### REST Endpoints (Admin)

```
GET    /api/v1/warehouses          — List warehouses
POST   /api/v1/warehouses          — Create warehouse
GET    /api/v1/warehouses/:id      — Get warehouse
PUT    /api/v1/warehouses/:id      — Update warehouse
GET    /api/v1/warehouses/:id/zones — List zones

GET    /api/v1/inventory           — Query inventory (with filters)
POST   /api/v1/inventory/adjust    — Manual inventory adjustment

GET    /api/v1/orders              — List orders
POST   /api/v1/orders              — Create order
GET    /api/v1/orders/:id          — Get order with lines
PUT    /api/v1/orders/:id/status   — Update order status

GET    /api/v1/tasks               — List tasks
POST   /api/v1/tasks/:id/assign    — Assign task
PUT    /api/v1/tasks/:id/complete  — Complete task
```

### REST Endpoints (PDA)

```
GET    /pda/v1/tasks               — My assigned tasks
PUT    /pda/v1/tasks/:id/start     — Start task
PUT    /pda/v1/tasks/:id/complete  — Complete task (with scanned barcodes)
POST   /pda/v1/tasks/:id/exception — Report exception

POST   /pda/v1/receiving/scan      — Scan ASN barcode → receive
POST   /pda/v1/putaway/confirm     — Confirm putaway to location
POST   /pda/v1/count/submit        — Submit cycle count result
```

### gRPC Services (Internal + Integration)

```protobuf
service WarehouseService { ... }
service InventoryService { ... }
service TaskService { ... }
service IntegrationService { ... }  // For WCS/RCS/MES/ERP
```

## Data Flow: Inbound Order (Example)

```
ERP → Order (inbound) → ASN → Receiving (PDA) → Putaway Task → Inventory Update
  │         │              │         │                │              │
  │    order_svc      asn_svc   task_svc         task_svc     inventory_svc
  │         │              │         │                │              │
  └─────────┴──────────────┴─────────┴────────────────┴──────────────┘
                             PostgreSQL
```

## Evolution Considerations

### What AI Can Safely Modify
- New domain types and constants
- Service implementations (taking interfaces)
- API handlers
- Frontend pages and components
- Tests
- Documentation

### What AI Should Be Conservative With
- Database schema changes (require migration files)
- Repository interfaces (affect all implementations)
- Proto definitions (affect generated code)
- Evolution scripts themselves (affect the evolution process)

### Rollback Strategy
If an evolution round breaks things:
1. `git log` to find the last good commit
2. `git revert` the broken commit
3. Mark the task as `failed` in roadmap
4. Evolution picks up the next task automatically
