#!/usr/bin/env bash
# =============================================================================
# AI-WMS Quality Check Script
# =============================================================================
# Runs all quality checks: build, test, lint.
# Exits with 0 if all pass, non-zero on failure.
# =============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

PASSED=0
FAILED=0

check() {
    local name="$1"
    shift
    echo "━━━ $name ━━━"
    if "$@" 2>&1; then
        echo "✅ $name PASSED"
        PASSED=$((PASSED + 1))
    else
        echo "❌ $name FAILED"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# ── Go Build ────────────────────────────────────────────────
check "Go Build" go build ./... || true

# ── Go Test ─────────────────────────────────────────────────
check "Go Test" go test -v -count=1 ./... 2>&1 || true

# ── Go Vet ──────────────────────────────────────────────────
check "Go Vet" go vet ./... || true

# ── Go Format Check ─────────────────────────────────────────
check "Go Format" bash -c '
    unformatted=$(gofmt -l . 2>/dev/null)
    if [[ -n "$unformatted" ]]; then
        echo "Unformatted files:"
        echo "$unformatted"
        exit 1
    fi
' || true

# ── Summary ─────────────────────────────────────────────────
echo ""
echo "========================================"
echo "  Quality Check Summary"
echo "========================================"
echo "  Passed: $PASSED"
echo "  Failed: $FAILED"
echo "========================================"

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi

exit 0
