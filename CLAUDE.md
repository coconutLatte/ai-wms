# AI-WMS: Self-Evolving Warehouse Management System

## Project Identity

This is a **self-evolving** WMS (Warehouse Management System). Every ~30 minutes, Claude Code automatically reads the roadmap, selects the highest-priority pending task, implements it, runs quality checks, and commits. The codebase grows organically, one small increment at a time.

## What We're Building

A complete WMS with:
- **Admin System** — React + Ant Design web dashboard for warehouse managers
- **PDA System** — React mobile web app for warehouse operators (barcode scanning, task execution)
- **Downward Integrations** — WCS (conveyors/sorters), RCS (AGV/AMR robots)
- **Upward Integrations** — MES (production orders), ERP (purchase/sales orders)

## Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Backend | Go (module: `github.com/ai-wms/ai-wms`) | Fast compilation for rapid iteration; strong typing catches AI errors |
| HTTP Router | chi/v5 | Lightweight, idiomatic Go, stdlib-compatible |
| Database | PostgreSQL 16 | Mature RDBMS, excellent for complex inventory queries |
| Cache | Redis 7 | Hot inventory data, session storage |
| RPC | gRPC + Protobuf | Efficient service-to-service and hardware system communication |
| Admin UI | React 18 + Ant Design 5 + Vite | Enterprise-grade component library |
| PDA UI | React 18 + Vite (mobile-first) | Same stack, mobile-optimized UI |
| Deployment | Docker Compose (dev), Kubernetes (prod) | |

## Architecture: DDD Layered

```
cmd/          → Application entry points (admin server, pda server)
internal/
  domain/     → Pure domain entities, zero dependencies (THE SOURCE OF TRUTH)
  service/    → Business logic orchestration
  repository/ → Data access interfaces + PostgreSQL implementation
  api/        → HTTP handlers, middleware, gRPC services
  integration/→ External system adapters (WCS, RCS, MES, ERP)
pkg/          → Shared utilities (config, logger, errors)
```

**CRITICAL RULES:**
1. Domain models in `internal/domain/` must have ZERO external dependencies — no database drivers, no HTTP, no gRPC. Pure Go structs only.
2. Repository interfaces are defined in `internal/repository/` and implemented in `internal/repository/postgres/`.
3. Every service in `internal/service/` takes interfaces, not concrete implementations.
4. API handlers are thin — they parse requests, call services, return responses. No business logic in handlers.

## Evolution Protocol

Each evolution round follows this exact protocol:

1. **Read State**: Load `docs/roadmap.md`, find the first `status: pending` task ordered by priority (P0 > P1 > P2 > ...)
2. **Understand Context**: Read relevant files in `internal/domain/`, `docs/architecture.md`, `docs/domain-model.md`
3. **Implement**: Write code that is compilable, testable, and follows existing patterns
4. **Quality Gate**: `go build ./...` MUST pass. `go test ./...` MUST pass.
5. **Commit**: Conventional commits format: `feat(scope): description`
6. **Update Roadmap**: Mark task as `completed` with date, move to next

## Coding Conventions

### Go
- Package names: lowercase, single word, no underscores
- File names: snake_case.go
- Structs: PascalCase
- JSON tags: snake_case
- Errors: always wrapped with context using `fmt.Errorf("context: %w", err)`
- Use `github.com/google/uuid` for all IDs
- Use `context.Context` as first parameter in all service/repository methods

### Commit Messages
Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat(scope): description` — new feature
- `fix(scope): description` — bug fix
- `refactor(scope): description` — code change without feature/fix
- `docs(scope): description` — documentation only
- `test(scope): description` — adding tests

### SQL
- Table names: plural (warehouses, zones, locations, skus, orders, tasks)
- Primary keys: UUID, default gen_random_uuid()
- Timestamps: TIMESTAMPTZ, default NOW()
- All schema changes in `migrations/` with sequential numbering

## Evolution Log
<!-- Evolution history will be appended here automatically -->
<!-- Last evolution: (initial seed) -->
