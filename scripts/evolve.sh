#!/usr/bin/env bash
# =============================================================================
# AI-WMS Auto-Evolution Engine
# =============================================================================
# Each round: reads roadmap → picks task → invokes Claude Code → commits.
# Triggered by crontab every 30 minutes, or manually: bash scripts/evolve.sh
#
# Usage:
#   bash scripts/evolve.sh             # Run one evolution cycle
#   bash scripts/evolve.sh --dry-run   # Show selected task, no execution
# =============================================================================

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="${LOG_DIR}/evolve-${TIMESTAMP}.log"

DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

log() { echo "[$(date '+%H:%M:%S')] $*" | tee -a "$LOG_FILE"; }
header() { log "==== $* ===="; }

# ── Step 1: Find next pending task ──────────────────────────

header "Reading Roadmap"

TASK_LINE=$(grep -E '^\| P[0-9]+-' "${REPO_ROOT}/docs/roadmap.md" | grep '| pending |' | head -1 || true)

if [[ -z "$TASK_LINE" ]]; then
    log "No pending tasks. Evolution complete!"
    exit 0
fi

TASK_ID=$(echo "$TASK_LINE"    | awk -F'|' '{print $2}' | xargs)
TASK_PRIO=$(echo "$TASK_LINE"  | awk -F'|' '{print $3}' | xargs)
TASK_DESC=$(echo "$TASK_LINE"  | awk -F'|' '{print $4}' | xargs)
TASK_NOTES=$(echo "$TASK_LINE" | awk -F'|' '{print $6}' | xargs)

log "Task: [$TASK_ID] $TASK_PRIO — $TASK_DESC"

if $DRY_RUN; then
    log "[DRY RUN] Would implement: $TASK_DESC"
    exit 0
fi

# ── Step 2: Git sync ────────────────────────────────────────

header "Git Sync"
git pull --rebase 2>/dev/null || log "(no remote, continuing locally)"

# ── Step 3: Build prompt ────────────────────────────────────

header "Building Prompt"

PROMPT_FILE="${LOG_DIR}/prompt-${TIMESTAMP}.md"

cat > "$PROMPT_FILE" << ENDPROMPT
You are the AI Evolution Engine for the ai-wms project at /root/workspace/ai-wms.
You are evolving this WMS (Warehouse Management System) autonomously.

## Current Task
- **ID**: ${TASK_ID}
- **Priority**: ${TASK_PRIO}
- **Description**: ${TASK_DESC}
- **Hint**: ${TASK_NOTES}

## Project Architecture (from CLAUDE.md)
- DDD layered: domain/ (zero deps) → service/ → repository/ → api/
- Backend: Go, chi/v5, pgx/v5, PostgreSQL 16
- Domain models in backend/internal/domain/ MUST have zero external dependencies
- Repository interfaces in backend/internal/repository/repository.go
- Repository implementations in backend/internal/repository/postgres/
- All IDs use github.com/google/uuid
- Errors wrapped: fmt.Errorf("doing X: %w", err)
- context.Context as first parameter in service/repository methods
- Commit format: feat(scope): description + Co-Authored-By line

## Your Mission
Implement ONLY the task described above. Steps:

1. Read any existing code files you need to understand the current state
2. Implement the feature — write production code and tests
3. Run: go build ./...   (fix any compilation errors)
4. Run: go test ./...    (fix any test failures)
5. Update docs/roadmap.md: change this task's status from "pending" to "completed | $(date +%Y-%m-%d) | <brief implementation note>"
6. Run: git add -A && git commit -m "feat(${TASK_PRIO}): ${TASK_DESC}

Co-Authored-By: deepseek-v4-pro <noreply@anthropic.com>"

## Constraints
- Only implement THIS task — don't go beyond scope
- Don't modify domain models unless the task explicitly requires it
- Don't change evolution scripts
- Keep changes minimal and focused
- Make sure go build && go test pass before committing
ENDPROMPT

log "Prompt written to $PROMPT_FILE ($(wc -c < "$PROMPT_FILE") bytes)"

# ── Step 4: Invoke Claude Code ──────────────────────────────

header "Invoking Claude Code"

if ! command -v claude &> /dev/null; then
    log "FATAL: claude CLI not found"
    exit 1
fi

cat "$PROMPT_FILE" | claude --print \
    --allowedTools "Read,Write,Edit,Bash,Glob" \
    --max-turns 40 \
    2>&1 | tee -a "$LOG_FILE"

CLAUDE_EXIT=$?
log "Claude Code exited with code $CLAUDE_EXIT"

# ── Step 5: Verify ──────────────────────────────────────────

header "Verification"

# Check if a commit was made
LATEST_COMMIT=$(git log --oneline -1 2>/dev/null)
log "Latest commit: $LATEST_COMMIT"

# Quick build check
if go build ./... 2>&1 | tee -a "$LOG_FILE"; then
    log "Build: PASS"
else
    log "Build: FAIL"
fi

log "Log: $LOG_FILE"
log "Evolution round complete."
exit 0
