#!/usr/bin/env bash
# =============================================================================
# AI-WMS Auto-Evolution Engine (Self-Evolving Roadmap Edition)
# =============================================================================
# The roadmap is NOT static. The AI can add, split, re-prioritize, and discover
# new tasks. Evolution never ends — it only changes direction.
#
# Modes (auto-selected based on roadmap state):
#   - implement: Pick highest-priority pending task and build it
#   - groom:     Every N rounds, review roadmap health and expand
#   - discover:  When < 5 pending tasks remain, AI explores and fills roadmap
#
# Usage:
#   bash scripts/evolve.sh              # Auto-detect mode and run
#   bash scripts/evolve.sh --dry-run    # Show what would happen
#   bash scripts/evolve.sh --groom      # Force roadmap grooming round
#   bash scripts/evolve.sh --discover   # Force discovery round
# =============================================================================

set -euo pipefail

# ── Cron needs explicit PATH and env vars ────────────────────
export PATH="/home/claude-dev/.local/bin:/usr/local/go/bin:$PATH"
export ANTHROPIC_BASE_URL="${ANTHROPIC_BASE_URL:-http://one-api.server22.jz}"
export ANTHROPIC_AUTH_TOKEN="${ANTHROPIC_AUTH_TOKEN:-sk-4v2AKtxcYlM3RNbemF3SMMoZTzxbBJt5fRqYpawSLKR4xGE1}"
export ANTHROPIC_MODEL="${ANTHROPIC_MODEL:-deepseek-v4-pro[1m]}"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="${LOG_DIR}/evolve-${TIMESTAMP}.log"

log()   { echo "[$(date '+%H:%M:%S')] $*" | tee -a "$LOG_FILE"; }
header(){ log "==== $* ===="; }

# ── Determine Mode ──────────────────────────────────────────

MODE="${1:-auto}"
DRY_RUN=false
[[ "$MODE" == "--dry-run" ]] && { DRY_RUN=true; MODE="auto"; }

# ── Lean Roadmap Mode ───────────────────────────────────────
# Max 10 pending tasks. When < 3 remain, DISCOVER to refill to ~10.
# GROOM only if pending >= 8 (don't groom a slim roadmap).
# No infinite grooming — every round is IMPLEMENT unless critically low.

MAX_PENDING=10
MIN_PENDING=3

ROADMAP="${REPO_ROOT}/docs/roadmap.md"
PENDING_COUNT=$(grep -cE '^\| P[0-9]+-.*\| pending \|' "$ROADMAP" 2>/dev/null || echo 0)
TOTAL_COUNT=$(grep -cE '^\| P[0-9]+-' "$ROADMAP" 2>/dev/null || echo 0)
COMPLETED_COUNT=$(grep -cE '^\| P[0-9]+-.*\| completed \|' "$ROADMAP" 2>/dev/null || echo 0)

ROUND_FILE="${REPO_ROOT}/.evolution-round"
if [[ -f "$ROUND_FILE" ]]; then
    CURRENT_ROUND=$(cat "$ROUND_FILE")
else
    CURRENT_ROUND=1
fi

LAST_MODE_FILE="${REPO_ROOT}/.evolution-last-mode"
LAST_MODE=$(cat "$LAST_MODE_FILE" 2>/dev/null || echo "implement")

NEXT_ROUND=$((CURRENT_ROUND + 1))

# Auto-detect mode
if [[ "$MODE" == "auto" || "$MODE" == "--now" ]]; then
    if [[ "$PENDING_COUNT" -lt "$MIN_PENDING" ]]; then
        MODE="discover"
        log "Auto-mode: only $PENDING_COUNT pending (min $MIN_PENDING) → DISCOVER (target $MAX_PENDING)"
    elif [[ "$PENDING_COUNT" -ge 8 ]] && [[ $((CURRENT_ROUND % 5)) -eq 0 ]] && [[ "$LAST_MODE" != "groom" ]]; then
        MODE="groom"
        log "Auto-mode: round $CURRENT_ROUND checkpoint, $PENDING_COUNT pending → GROOM"
    else
        MODE="implement"
        log "Auto-mode: $PENDING_COUNT pending → IMPLEMENT"
    fi
fi

# ── Git Sync ─────────────────────────────────────────────────

header "Evolution Round $CURRENT_ROUND — Mode: $MODE"
log "Roadmap: $COMPLETED_COUNT done / $TOTAL_COUNT total / $PENDING_COUNT pending"

if $DRY_RUN; then
    log "[DRY RUN] Would run mode: $MODE"
    exit 0
fi

git pull --rebase 2>/dev/null || log "(no remote, continuing locally)"

# ── Build Prompt ─────────────────────────────────────────────

header "Building $MODE Prompt"
PROMPT_FILE="${LOG_DIR}/prompt-${TIMESTAMP}.md"

case "$MODE" in

# ═══════════════════════════════════════════════════════════════
# IMPLEMENT: Pick a pending task and build it
# ═══════════════════════════════════════════════════════════════
implement)
    # awk with exit avoids SIGPIPE from grep|head when many matches exist
    TASK_LINE=$(awk -F'|' '/^\| P[0-9]+-/ && /\| pending \|/ {print; exit}' "$ROADMAP")
    TASK_ID=$(echo "$TASK_LINE"   | awk -F'|' '{print $2}' | xargs)
    TASK_PRIO=$(echo "$TASK_LINE" | awk -F'|' '{print $3}' | xargs)
    TASK_DESC=$(echo "$TASK_LINE" | awk -F'|' '{print $4}' | xargs)
    TASK_NOTE=$(echo "$TASK_LINE" | awk -F'|' '{print $6}' | xargs)

    log "Implementing: [$TASK_ID] $TASK_DESC"

    cat > "$PROMPT_FILE" << ENDPROMPT
You are evolving the ai-wms project at /root/workspace/ai-wms.

## Task: ${TASK_ID} — ${TASK_DESC}
Hint: ${TASK_NOTE}

## Architecture
- DDD: domain/ (zero deps) → service/ → repository/ → api/
- Go + chi/v5 + pgx/v5 + PostgreSQL 16
- Domain models: zero external dependencies. Repository interfaces in repository/. Impls in postgres/.
- UUID IDs, fmt.Errorf wrapping, context.Context first param

## Steps
1. Read existing code to understand current state
2. Implement the task — write production code AND tests
3. Run: go build ./... && go test ./...   (fix any errors)
4. Update docs/roadmap.md: change this task status to "completed | $(date +%Y-%m-%d) | <note>"
5. **IMPORTANT — Roadmap Self-Evolution**: After implementing, review:
   - Are there follow-up tasks this implementation enables? Add them.
   - Are there edge cases or improvements you noticed? Add them as new roadmap entries.
   - Is a related task now obsolete? Mark it.
   - Use new task IDs: P<priority>-<next-available-number>
   - Append new tasks to the appropriate Phase section in docs/roadmap.md
6. Commit & push: git add -A && git commit -m "feat(${TASK_PRIO}): ${TASK_DESC}

Co-Authored-By: deepseek-v4-pro <noreply@anthropic.com>"
ENDPROMPT
    ;;

# ═══════════════════════════════════════════════════════════════
# GROOM: Review roadmap health, expand future tasks, re-prioritize
# ═══════════════════════════════════════════════════════════════
groom)
    log "Grooming roadmap (round $CURRENT_ROUND)"

    cat > "$PROMPT_FILE" << ENDPROMPT
You are grooming the ai-wms project roadmap at /root/workspace/ai-wms.

## Mission: Roadmap Grooming

This is a meta-evolution round. Do NOT implement features. Instead, review and improve the roadmap itself.

### Steps
1. Read docs/roadmap.md to see all tasks and their status
2. Read CLAUDE.md and docs/architecture.md for project vision
3. Review the current codebase: what's implemented? What patterns exist?
4. Run: go vet ./... and go test ./... to find quality gaps

### Then update docs/roadmap.md:

**Expand the roadmap:**
- What Phase 5/6 features are missing? (monitoring, alerting, API docs, deployment, k8s, etc.)
- What Phase 4 integration details need more granularity?
- What cross-cutting concerns need tasks? (security hardening, performance tuning, code quality, documentation)
- What new features would make this a more complete WMS?

**Re-prioritize if needed:**
- Are there dependencies between tasks that aren't reflected?
- Should any P2/P3 tasks be promoted to P1 based on what's built?

**Clean up:**
- Mark any truly obsolete tasks as "cancelled" with reason
- Split overly large tasks into smaller ones
- Ensure task IDs are unique and sequential within phases

**Update GitHub ecosystem files:**
- Update README.md <!-- EVOLUTION-STATS --> block with current task counts and rounds
- Verify LICENSE and CONTRIBUTING.md are present and current
- If new major features were added, update README feature list and architecture diagram

**Format for new tasks:**
| P<phase>-<NN> | P<phase> | <description> | pending | — | <hint> |

Add new phases (P6, P7...) if the existing phases are full or need expansion.

5. Commit & push: git add -A && git commit -m "chore(roadmap): grooming round $CURRENT_ROUND — expanded roadmap

Co-Authored-By: deepseek-v4-pro <noreply@anthropic.com>"
ENDPROMPT
    ;;

# ═══════════════════════════════════════════════════════════════
# DISCOVER: Deep exploration to find new tasks
# ═══════════════════════════════════════════════════════════════
discover)
    TARGET=$MAX_PENDING
    log "Discovering new tasks (only $PENDING_COUNT pending, target $TARGET)"

    cat > "$PROMPT_FILE" << ENDPROMPT
You are discovering new tasks for the ai-wms project at /root/workspace/ai-wms.

## Mission: Task Discovery (Lean Roadmap)

The roadmap has only $PENDING_COUNT pending tasks (minimum is $MIN_PENDING). You must refill to approximately $TARGET pending tasks. Current total: $TOTAL_COUNT tasks.

### Steps
1. Read docs/roadmap.md to see completed tasks and current pending
2. Read CLAUDE.md and docs/architecture.md for project vision
3. Scan the codebase to find the most impactful gaps:
   - backend/ — missing repos, services, APIs, tests
   - frontend/ — scaffold only, needs real pages
   - docs/ — README, architecture diagrams, API docs
4. Run: go test ./... -cover to find coverage gaps

### Then update docs/roadmap.md:

**Add new tasks to reach ~$TARGET pending total (NOT more).** Focus on the NEXT most valuable work, not a 5-year master plan.

Priority of new tasks:
- **P0**: Anything blocking the next concrete milestone
- **P1**: Directly builds on completed work (next repo → next service → next API)
- **P2**: Frontend pages, developer experience, testing

**Crucial rules:**
- DO NOT add speculative Phase 5+ tasks (report engine, ML slotting, etc.) unless they are the logical next step
- DO NOT create new phases beyond the next 2 phases
- Keep task descriptions specific and implementable in one round
- If completed tasks create new gaps (e.g., "we have repos but no services"), add THOSE

Format: | P<phase>-<NN> | P<phase> | <task> | pending | — | <hint> |

6. Commit & push: git add -A && git commit -m "feat(roadmap): discovery round — refilled to $TARGET pending tasks

Co-Authored-By: deepseek-v4-pro <noreply@anthropic.com>"
ENDPROMPT
ENDPROMPT
    ;;

*)
    log "Unknown mode: $MODE"
    exit 1
    ;;
esac

# ── Invoke Claude Code ──────────────────────────────────────

header "Invoking Claude Code ($MODE)"

if ! command -v claude &> /dev/null; then
    log "FATAL: claude CLI not found"
    exit 1
fi

log "Prompt: $PROMPT_FILE ($(wc -c < "$PROMPT_FILE") bytes)"

cat "$PROMPT_FILE" | claude --print \
    --allowedTools "Read,Write,Edit,Bash,Glob" \
    2>&1 | tee -a "$LOG_FILE"

CLAUDE_EXIT=$?

# ── Verify ──────────────────────────────────────────────────

header "Verify"
NEW_PENDING=$(grep -cE '^\| P[0-9]+-.*\| pending \|' "$ROADMAP" 2>/dev/null || echo 0)
NEW_COMPLETED=$(grep -cE '^\| P[0-9]+-.*\| completed \|' "$ROADMAP" 2>/dev/null || echo 0)
LATEST_COMMIT=$(git log --oneline -1 2>/dev/null)

log "Pending: $PENDING_COUNT → $NEW_PENDING"
log "Completed: $COMPLETED_COUNT → $NEW_COMPLETED"
log "Latest commit: $LATEST_COMMIT"
log "Exit code: $CLAUDE_EXIT"

# Push if remote is configured
git push 2>/dev/null || log "(no remote or push failed)"

# Save persistent state
echo "$NEXT_ROUND" > "$ROUND_FILE"
echo "$MODE" > "$LAST_MODE_FILE"

log "Round: $CURRENT_ROUND → $NEXT_ROUND"
log "Log: $LOG_FILE"
log "Done. Next round in ~10 min."
exit 0
