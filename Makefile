.PHONY: help dev build test lint clean db-up db-down db-reset evolve

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Development ──────────────────────────────────────────────

dev: ## Start development infrastructure (PostgreSQL + Redis)
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 3
	@docker-compose ps

dev-down: ## Stop development infrastructure
	docker-compose down

dev-reset: ## Reset development infrastructure (WARNING: deletes all data)
	docker-compose down -v
	docker-compose up -d

# ── Build & Test ────────────────────────────────────────────

build: ## Build all Go binaries
	go build ./...

test: ## Run all tests
	go test -v -race -count=1 ./...

test-cover: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	golangci-lint run ./... 2>/dev/null || echo "golangci-lint not installed — run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

fmt: ## Format Go code
	go fmt ./...

tidy: ## Tidy Go modules
	go mod tidy

# ── Database ────────────────────────────────────────────────

db-up: ## Start database only
	docker-compose up -d postgres
	@sleep 3

db-migrate: ## Run database migrations
	@echo "Migrations are auto-applied via docker-entrypoint-initdb.d on first start"
	@echo "For manual migrations, use: psql -h localhost -U wms -d wms -f migrations/000001_init_schema.sql"

db-reset: ## Reset database
	docker-compose down -v postgres
	docker-compose up -d postgres
	@sleep 3
	@echo "Database reset complete"

# ── Code Generation ─────────────────────────────────────────

proto: ## Generate Go code from proto files
	buf generate proto/

# ── Quality ─────────────────────────────────────────────────

quality: build test lint ## Run full quality check (build + test + lint)

# ── Auto-Evolution ──────────────────────────────────────────

evolve: ## Run one evolution cycle manually
	@bash scripts/evolve.sh

evolve-dry: ## Dry-run evolution (no changes)
	@bash scripts/evolve.sh --dry-run

# ── Setup ───────────────────────────────────────────────────

setup: ## Full project setup
	@bash scripts/setup-dev.sh

# ── Docker ──────────────────────────────────────────────────

docker-build: ## Build Docker images
	docker build -t ai-wms/admin:latest -f backend/cmd/admin/Dockerfile .
	docker build -t ai-wms/pda:latest -f backend/cmd/pda/Dockerfile .
