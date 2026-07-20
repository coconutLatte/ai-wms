#!/usr/bin/env bash
# =============================================================================
# AI-WMS Auto-Evolution Script
# =============================================================================
# This script is the engine of the self-evolving WMS. It:
# 1. Reads the roadmap to find the highest-priority pending task
# 2. Constructs a prompt with full context
# 3. Invokes Claude Code to implement the task
# 4. Runs quality gates (build + test)
# 5. Commits and pushes if successful
# 6. Updates the roadmap
#
# Usage:
#   ./scripts/evolve.sh          # Run one evolution cycle
#   ./scripts/evolve.sh --dry-run # Show what would be done without doing it
# =============================================================================

set -euo pipefail

# ── Configuration ────────────────────────────────────────────

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "$LOG_DIR"

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="${LOG_DIR}/evolve-${TIMESTAMP}.log"
ROADMAP_FILE="${REPO_ROOT}/docs/roadmap.md"
ARCHITECTURE_FILE="${REPO_ROOT}/docs/architecture.md"
CLAUDE_MD="${REPO_ROOT}/CLAUDE.md"

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
    DRY_RUN=true
    echo "[DRY RUN] No changes will be made"
fi

# ── Logging ──────────────────────────────────────────────────

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
    echo "$msg" | tee -a "$LOG_FILE"
}

log_section() {
    log "========================================"
    log "  $*"
    log "========================================"
}

# ── Step 1: Read State ──────────────────────────────────────

log_section "Step 1: Reading Roadmap State"

if [[ ! -f "$ROADMAP_FILE" ]]; then
    log "ERROR: roadmap.md not found at $ROADMAP_FILE"
    exit 1
fi

# Find the first pending task (ordered by priority: P0 > P1 > P2 > ...)
# Parse markdown table rows looking for | status | pending |
TASK_LINE=$(grep -E '^\| P[0-9]+-' "$ROADMAP_FILE" | grep '| pending |' | head -1 || true)

if [[ -z "$TASK_LINE" ]]; then
    log "No pending tasks found. Evolution complete!"
    exit 0
fi

# Extract fields from the markdown table row
TASK_ID=$(echo "$TASK_LINE" | awk -F'|' '{print $2}' | xargs)
TASK_PRIORITY=$(echo "$TASK_LINE" | awk -F'|' '{print $3}' | xargs)
TASK_DESC=$(echo "$TASK_LINE" | awk -F'|' '{print $4}' | xargs)

log "Selected task: [$TASK_ID] $TASK_PRIORITY — $TASK_DESC"
log "Roadmap line: $TASK_LINE"

if $DRY_RUN; then
    log "[DRY RUN] Would implement: $TASK_DESC"
    exit 0
fi

# ── Step 2: Git Sync ────────────────────────────────────────

log_section "Step 2: Git Sync"

# Check if this is a git repository
if git rev-parse --git-dir > /dev/null 2>&1; then
    log "Fetching latest from origin..."
    git fetch origin 2>&1 | tee -a "$LOG_FILE" || log "WARNING: git fetch failed (no remote?)"

    # Check if we can pull (has upstream)
    if git rev-parse --abbrev-ref @{u} > /dev/null 2>&1; then
        log "Pulling latest changes..."
        git pull --rebase 2>&1 | tee -a "$LOG_FILE" || {
            log "WARNING: git pull failed, continuing with local state"
        }
    else
        log "No upstream configured, working locally"
    fi
else
    log "Not a git repository yet. Initializing..."
    git init
    git add -A
    git commit -m "chore: initial commit before first evolution
Co-Authored-By: deepseek-v4-pro <noreply@anthropic.com>"
fi

# ── Step 3: Build Evolution Prompt ──────────────────────────

log_section "Step 3: Building Evolution Prompt"

# Read architecture context
ARCH_CONTEXT=""
if [[ -f "$ARCHITECTURE_FILE" ]]; then
    ARCH_CONTEXT=$(head -100 "$ARCHITECTURE_FILE")
fi

# Read domain model context
DOMAIN_CONTEXT=""
if [[ -f "${REPO_ROOT}/docs/domain-model.md" ]]; then
    DOMAIN_CONTEXT=$(head -100 "${REPO_ROOT}/docs/domain-model.md")
fi

# Construct the prompt
EVOLVE_PROMPT=$(cat <<PROMPT
You are the AI Evolution Engine for the ai-wms project — a self-evolving Warehouse Management System.

## Project Context
${ARCH_CONTEXT}

## Domain Model
${DOMAIN_CONTEXT}

## Current Evolution Task
- **Task ID**: ${TASK_ID}
- **Priority**: ${TASK_PRIORITY}
- **Description**: ${TASK_DESC}

## Your Mission
Implement ONLY this task. Do not go beyond scope.

### Requirements:
1. Write compilable Go code following existing patterns in backend/internal/
2. Domain models go in backend/internal/domain/ (ZERO external dependencies)
3. Service logic goes in backend/internal/service/
4. Repository implementations go in backend/internal/repository/postgres/
5. API handlers go in backend/internal/api/
6. Every new function/method should have a corresponding test
7. Run 'go build ./...' to verify compilation before finishing
8. Run 'go test ./...' to verify tests pass before finishing
9. Update docs/roadmap.md to mark this task as completed when done
10. Follow all conventions from CLAUDE.md

### Constraints:
- Do NOT modify domain model files unless the task explicitly requires it
- Do NOT change the database schema without adding a migration file
- Do NOT change the evolution scripts or workflow
- Keep changes minimal and focused

### When Complete:
1. Verify: go build ./... && go test ./...
2. Mark task as completed in docs/roadmap.md
3. Commit with message: "feat: ${TASK_DESC}"
PROMPT
)

# Write the prompt to a file for debugging
PROMPT_FILE="${LOG_DIR}/prompt-${TIMESTAMP}.md"
echo "$EVOLVE_PROMPT" > "$PROMPT_FILE"
log "Prompt written to $PROMPT_FILE"

# ── Step 4: Invoke Claude Code ──────────────────────────────

log_section "Step 4: Invoking Claude Code"

# Check if claude CLI is available
if command -v claude &> /dev/null; then
    log "Claude Code CLI available, invoking..."

    # Run Claude Code with the prompt
    # --print: non-interactive mode, output to stdout
    # --allowedTools: restrict to safe tools
    CLAUDE_EXIT_CODE=0
    echo "$EVOLVE_PROMPT" | claude --print \
        --allowedTools "Read,Write,Edit,Bash(search_files_here_pattern),Glob,NotebookEdit" \
        2>&1 | tee -a "$LOG_FILE" || CLAUDE_EXIT_CODE=$?

    if [[ $CLAUDE_EXIT_CODE -ne 0 ]]; then
        log "ERROR: Claude Code exited with code $CLAUDE_EXIT_CODE"
        # Continue to quality gate anyway — Claude may have made partial progress
    fi
else
    log "WARNING: Claude Code CLI not found on PATH"
    log "Evolution cannot proceed without Claude Code CLI"
    log "Please install: npm install -g @anthropic-ai/claude-code"
    exit 1
fi

# ── Step 5: Quality Gate ────────────────────────────────────

log_section "Step 5: Quality Gate"

# Run quality checks
bash "${REPO_ROOT}/scripts/quality-check.sh" 2>&1 | tee -a "$LOG_FILE" || {
    log_section "Quality Gate FAILED — Attempting Auto-Fix"

    # Try to fix with Claude Code (max 2 attempts)
    for attempt in 1 2; do
        log "Auto-fix attempt $attempt/2..."

        FIX_PROMPT="The quality check failed for the ai-wms project after implementing task ${TASK_ID}: ${TASK_DESC}.

        Error output is in the log above. Please fix ALL compilation errors and test failures.
        Run 'go build ./...' and 'go test ./...' to verify your fixes.
        Do NOT make any other changes beyond fixing errors."

        echo "$FIX_PROMPT" | claude --print \
            --allowedTools "Read,Write,Edit,Bash,Glob" \
            2>&1 | tee -a "$LOG_FILE" || true

        # Re-check quality
        if bash "${REPO_ROOT}/scripts/quality-check.sh" 2>&1 | tee -a "$LOG_FILE"; then
            log "Auto-fix succeeded on attempt $attempt"
            break
        fi

        if [[ $attempt -eq 2 ]]; then
            log_section "FATAL: Quality gate still failing after 2 fix attempts"
            log "Task ${TASK_ID} will be marked as 'failed'"

            # Mark task as failed in roadmap
            sed -i "s/| ${TASK_ID} | ${TASK_PRIORITY} | ${TASK_DESC} | pending |/| ${TASK_ID} | ${TASK_PRIORITY} | ${TASK_DESC} | failed |/g" "$ROADMAP_FILE"

            # Rollback changes
            log "Rolling back changes..."
            git checkout -- .
            git clean -fd

            exit 1
        fi
    done
}

log "Quality gate PASSED"

# ── Step 6: Commit & Update Roadmap ─────────────────────────

log_section "Step 6: Commit & Update Roadmap"

# Update roadmap: mark as completed with today's date
TODAY=$(date +%Y-%m-%d)
sed -i "s/| ${TASK_ID} | ${TASK_PRIORITY} | ${TASK_DESC} | pending |/| ${TASK_ID} | ${TASK_PRIORITY} | ${TASK_DESC} | completed | ${TODAY} |/g" "$ROADMAP_FILE"

# Get list of changed files
CHANGED_FILES=$(git diff --name-only 2>/dev/null || git status --short | awk '{print $2}')
log "Changed files:"
echo "$CHANGED_FILES" | tee -a "$LOG_FILE"

# Stage all changes
git add -A

# Commit
COMMIT_MSG="feat(${TASK_PRIORITY}): ${TASK_DESC}

Task: ${TASK_ID}
Evolution round: ${TIMESTAMP}

Co-Authored-By: deepseek-v4-pro <noreply@anthropic.com>"

git commit -m "$COMMIT_MSG" 2>&1 | tee -a "$LOG_FILE"
log "Committed with message: feat(${TASK_PRIORITY}): ${TASK_DESC}"

# Push if remote is configured
if git rev-parse --abbrev-ref @{u} > /dev/null 2>&1; then
    git push 2>&1 | tee -a "$LOG_FILE"
    log "Pushed to remote"
else
    log "No remote configured, skipping push"
fi

# ── Step 7: Update CLAUDE.md with Evolution Log ─────────────

log_section "Step 7: Updating Evolution Log"

# Append evolution entry to CLAUDE.md
cat >> "$CLAUDE_MD" << EVO_LOG

## Evolution ${TIMESTAMP}
- **Task**: [${TASK_ID}] ${TASK_DESC}
- **Status**: completed
- **Files changed**: $(echo "$CHANGED_FILES" | wc -l) files
EVO_LOG

log_section "EVOLUTION COMPLETE"
log "Task: ${TASK_ID} — ${TASK_DESC}"
log "Log: ${LOG_FILE}"
log "Next evolution scheduled in ~30 minutes"

exit 0
