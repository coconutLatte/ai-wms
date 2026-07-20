# ADR-001: Go + chi/v5 + PostgreSQL

**Status**: Accepted
**Date**: 2026-07-20
**Author**: AI Evolution System (initial seed)

## Context

We are building a self-evolving WMS that will be developed incrementally by AI agents. The technology stack must:
1. Be easy for AI to generate correct code (strong typing, clear patterns)
2. Compile and test quickly (under 30 seconds for quality gate)
3. Support the domain complexity of WMS (inventory transactions, order state machines)
4. Scale from prototype to production without fundamental rewrites

## Decision

- **Language**: Go 1.26+
- **HTTP Framework**: chi/v5
- **Database**: PostgreSQL 16 with pgx/v5 driver
- **API Style**: REST (chi) + gRPC (for internal/integration services)
- **Cache**: Redis 7 with go-redis/v9

## Alternatives Considered

### Python (FastAPI)
- Pro: Fast to prototype, great ecosystem
- Con: Dynamic typing means AI-generated bugs surface at runtime; slower performance for WMS throughput; GIL limits concurrency
- Verdict: Rejected. Compile-time safety is critical for AI-generated code.

### Node.js/TypeScript
- Pro: Same language as frontend, large ecosystem
- Con: ORM complexity (Prisma/TypeORM), runtime type system still allows many errors; npm ecosystem churn
- Verdict: Rejected. Go's simplicity is preferred for AI-driven development.

### Java/Spring Boot
- Pro: Enterprise-grade, battle-tested for WMS
- Con: Slow compilation, verbose, heavy framework that AI tends to misuse
- Verdict: Rejected. Too heavyweight for rapid 30-minute evolution cycles.

### Rust
- Pro: Maximum safety and performance
- Con: Steep learning curve, slow compilation, AI frequently generates non-compiling code
- Verdict: Rejected. Compilation speed and AI error rate make it unsuitable.

### Go with Gin
- Pro: Most popular Go HTTP framework
- Con: Non-standard handler signatures, magic internals, less AI-friendly
- Verdict: Rejected. chi's stdlib compatibility makes it easier for AI to generate correct code.

## Consequences

### Positive
- Go compiles in seconds → fast quality gate in evolution cycle
- Strong typing → AI errors caught at compile time
- chi's `net/http` compatibility → predictable, well-documented patterns
- pgx is the fastest Go PostgreSQL driver
- Single binary deployment simplifies operations

### Negative
- gRPC setup adds build complexity (protoc + protoc-gen-go required)
- Go's error handling is verbose (but explicit — good for AI comprehension)
- No built-in ORM (intentional — we use raw SQL with pgx for maximum control)

## References

- [chi/v5](https://github.com/go-chi/chi)
- [pgx/v5](https://github.com/jackc/pgx)
- [Domain-Driven Design in Go](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
