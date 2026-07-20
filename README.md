# 🤖 AI-WMS — Self-Evolving Warehouse Management System

[![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis)](https://redis.io/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Evolving](https://img.shields.io/badge/evolving-every%2030min-ff69b4)](https://github.com/coconutLatte/ai-wms)

> **A WMS that writes itself.** Every 30 minutes, Claude Code picks the highest-priority task from a self-expanding roadmap, implements it, runs tests, and pushes to this repo — fully autonomous.

## 🎯 What Is This?

AI-WMS is a complete Warehouse Management System with:

- **📊 Admin Dashboard** — React + Ant Design web UI for warehouse managers
- **📱 PDA Operations** — React mobile web app for warehouse operators (barcode scanning, receiving, picking, putaway)
- **⬇️ Downward Integration** — WCS (conveyors, sorters), RCS (AGV/AMR robots)
- **⬆️ Upward Integration** — MES (production), ERP (purchase/sales)

The twist: **the codebase evolves autonomously.** A cron job triggers `scripts/evolve.sh` every 30 minutes, which reads the roadmap, constructs a prompt, invokes Claude Code to implement the task, runs quality gates, commits, and pushes. The roadmap itself self-expands via periodic grooming rounds.

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────┐
│                   Presentation                       │
│  ┌──────────────┐  ┌──────────────┐                 │
│  │ Admin (React)│  │ PDA (React)  │                 │
│  └──────┬───────┘  └──────┬───────┘                 │
│         │                 │                          │
│         ▼                 ▼                          │
│  ┌─────────────────────────────────────────────┐    │
│  │           API Layer (chi/v5 + gRPC)         │    │
│  └──────────────────────┬──────────────────────┘    │
│                         │                            │
│                         ▼                            │
│  ┌─────────────────────────────────────────────┐    │
│  │           Service Layer                      │    │
│  │  Warehouse │ Inventory │ Order │ Task        │    │
│  └──────────────────────┬──────────────────────┘    │
│                         │                            │
│                         ▼                            │
│  ┌─────────────────────────────────────────────┐    │
│  │           Domain Layer (pure Go, zero deps)  │    │
│  │  Warehouse │ Zone │ Location │ SKU │ Order  │    │
│  └──────────────────────┬──────────────────────┘    │
│                         │                            │
│                         ▼                            │
│  ┌─────────────────────────────────────────────┐    │
│  │           Infrastructure                     │    │
│  │  PostgreSQL 16 │ Redis 7 │ pgx/v5            │    │
│  └─────────────────────────────────────────────┘    │
│                                                     │
│  ┌─────────────────────────────────────────────┐    │
│  │     Integration Adapters (WCS/RCS/MES/ERP)  │    │
│  └─────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────┘
```

## 🧬 Self-Evolution Mechanism

```
┌──────────────────────────────────────────────────┐
│              Every 30 Minutes                    │
│                                                  │
│  📖 Roadmap → 🎯 Pick Task → 🤖 Claude Code      │
│       │                                      │    │
│       ▼                                      ▼    │
│  ✅ go build + go test ←── Auto-fix ←── Fail?    │
│       │                                      │    │
│       ▼                                      │    │
│  📝 Commit → 🚀 Push → 🔄 Update Roadmap         │
│                                                  │
│  Three modes:                                    │
│  🛠️  implement — build the next feature          │
│  🌿 groom     — expand roadmap (every 5 rounds)  │
│  🔎 discover  — explore for new tasks            │
└──────────────────────────────────────────────────┘
```

## 📂 Project Structure

```
ai-wms/
├── CLAUDE.md                 # AI instructions & conventions
├── docs/
│   ├── architecture.md       # System architecture
│   ├── roadmap.md            # Self-expanding task list
│   └── domain-model.md       # Domain entity reference
├── backend/
│   ├── cmd/                  # Entry points (admin + pda)
│   ├── internal/
│   │   ├── domain/           # Pure domain entities
│   │   ├── service/          # Business logic
│   │   ├── repository/       # Data access (PostgreSQL)
│   │   ├── api/              # HTTP handlers + middleware
│   │   └── integration/      # External adapters
│   └── pkg/                  # Shared utilities
├── frontend/
│   ├── admin/                # Admin SPA (React + Ant Design)
│   └── pda/                  # PDA mobile app (React)
├── proto/                    # gRPC service definitions
├── migrations/               # PostgreSQL migrations
├── scripts/
│   └── evolve.sh             # Auto-evolution engine
└── .claude/                  # Claude Code agents & workflows
```

## 🚀 Quick Start

```bash
# Prerequisites
go version  # ≥ 1.26
docker ps   # PostgreSQL + Redis

# Setup
make setup          # Install deps, start DB, run migrations
make build          # Build all services
make test           # Run tests

# Start services
make run-admin      # Admin API on :8080
make run-pda        # PDA API on :8081

# Trigger evolution manually
make evolve         # Run one evolution cycle
make evolve-dry     # Preview next task only
```

## 📊 Evolution Progress

<!-- EVOLUTION-STATS-START -->
| Metric | Value |
|--------|-------|
| Total tasks | 156 |
| Completed | 38 |
| Pending | 118 |
| Evolution rounds | 34 |
| Last evolution | 2026-07-21 |
<!-- EVOLUTION-STATS-END -->

See [docs/roadmap.md](docs/roadmap.md) for the full task list.

## 🔧 Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26, chi/v5, pgx/v5 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| RPC | gRPC + Protobuf |
| Admin UI | React 18, Ant Design 5, Vite |
| PDA UI | React 18, antd-mobile, Vite |
| Deployment | Docker Compose (dev), TBD (prod) |

## 🤝 Contributing

This project evolves autonomously, but human contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for ways to get involved.

## 📄 License

MIT — see [LICENSE](LICENSE) for details.

---

*Built by Claude Code, every 30 minutes.*
