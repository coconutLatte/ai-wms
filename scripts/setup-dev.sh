#!/usr/bin/env bash
# =============================================================================
# AI-WMS Development Environment Setup
# =============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

echo "========================================"
echo "  AI-WMS Development Setup"
echo "========================================"

# Check prerequisites
echo ""
echo "Checking prerequisites..."

check_cmd() {
    if command -v "$1" &> /dev/null; then
        echo "  ✅ $1 found: $($1 version 2>/dev/null || $1 --version 2>/dev/null || echo 'ok')"
    else
        echo "  ❌ $1 not found — please install it"
    fi
}

check_cmd go
check_cmd node
check_cmd docker
check_cmd claude

# Install Go dependencies
echo ""
echo "Installing Go dependencies..."
go mod tidy
go mod download

# Start infrastructure
echo ""
echo "Starting PostgreSQL + Redis..."
docker-compose up -d
echo "Waiting for services to be ready..."
sleep 3

# Verify database
echo ""
echo "Verifying database connection..."
if docker-compose exec -T postgres pg_isready -U wms 2>/dev/null; then
    echo "  ✅ PostgreSQL is ready"
else
    echo "  ⚠️  PostgreSQL may still be starting — migrations will apply shortly"
fi

# Run initial build
echo ""
echo "Running initial build..."
go build ./...

# Run tests
echo ""
echo "Running tests..."
go test ./... || echo "  ⚠️  Some tests failed (may be expected during initial setup)"

# Make scripts executable
chmod +x scripts/*.sh

echo ""
echo "========================================"
echo "  Setup Complete!"
echo "========================================"
echo ""
echo "  Database:  postgresql://wms:wms_dev_2026@localhost:5432/wms"
echo "  Redis:     redis://localhost:6379"
echo ""
echo "  Next steps:"
echo "    make evolve   — Run one manual evolution cycle"
echo "    make test     — Run all tests"
echo "    make help     — Show all available commands"
echo ""
echo "  To enable auto-evolution every 30 minutes:"
echo "    crontab -l | { cat; echo '*/30 * * * * cd ${REPO_ROOT} && bash scripts/evolve.sh >> logs/cron-evolve.log 2>&1'; } | crontab -"
echo ""
