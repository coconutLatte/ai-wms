export const meta = {
  name: 'wms-evolve',
  description: 'Multi-agent WMS evolution: architect reviews task, developer implements, reviewer verifies',
  phases: [
    { title: 'Understand', detail: 'Read roadmap, architecture, and domain model to understand the task' },
    { title: 'Architect', detail: 'Design approach and validate against existing architecture' },
    { title: 'Implement', detail: 'Write code, tests, and update documentation' },
    { title: 'Review', detail: 'Adversarial code review — find bugs and violations' },
    { title: 'Fix', detail: 'Fix issues found in review' },
    { title: 'Verify', detail: 'Build, test, and quality gate' },
  ],
}

// ── Phase 1: Understand ──────────────────────────────────────
phase('Understand')

const roadmap = await agent(
  'Read docs/roadmap.md and find the highest-priority pending task (P0 first, then P1, etc.). Return: { task_id, priority, description } as JSON.',
  { label: 'read-roadmap', schema: { type: 'object', properties: { task_id: { type: 'string' }, priority: { type: 'string' }, description: { type: 'string' } }, required: ['task_id', 'priority', 'description'] } }
)

log(`Selected task: [${roadmap.task_id}] ${roadmap.priority} — ${roadmap.description}`)

const architecture = await agent(
  'Read docs/architecture.md and docs/domain-model.md. Return a summary of: key architectural constraints, relevant domain models for this task, and any patterns that must be followed.',
  { label: 'read-architecture' }
)

log(`Architecture context loaded: ${architecture.slice(0, 200)}...`)

// ── Phase 2: Architect ──────────────────────────────────────
phase('Architect')

const design = await agent(
  `Design the implementation approach for task [${roadmap.task_id}]: ${roadmap.description}.

Architecture context:
${architecture}

Provide:
1. Which files need to be created/modified
2. Which domain models are involved
3. API endpoints (if any) to be created
4. Database changes (if any)
5. Testing strategy
6. Any architectural concerns

Be specific about file paths and function signatures.`,
  { label: 'design-approach' }
)

log(`Design complete: ${design.slice(0, 300)}...`)

// ── Phase 3: Implement ──────────────────────────────────────
phase('Implement')

const implResult = await agent(
  `Implement task [${roadmap.task_id}]: ${roadmap.description}.

Design approach:
${design}

Project rules from CLAUDE.md:
- Domain models in internal/domain/ must have ZERO external dependencies
- Repository interfaces in internal/repository/, implementations in internal/repository/postgres/
- Services take interfaces, not concretions
- API handlers are thin — parse, call service, respond
- Errors wrapped with fmt.Errorf("context: %w", err)
- context.Context as first parameter in service/repository methods
- All IDs use github.com/google/uuid
- Every new function should have a test

After implementation:
1. Run 'go build ./...' to verify compilation
2. Run 'go test ./...' to verify tests pass
3. Fix any compilation or test errors

Update docs/roadmap.md to mark the task as completed when done.`,
  { label: 'implement-task', agentType: 'wms-developer' }
)

log(`Implementation complete`)

// ── Phase 4: Review ─────────────────────────────────────────
phase('Review')

const reviewFindings = await agent(
  `Review the code changes made for task [${roadmap.task_id}]: ${roadmap.description}.

Check for:
- 🔴 Critical: compilation errors, missing tests, architectural violations (domain models with external deps, business logic in handlers)
- 🟡 Warning: missing error handling, potential bugs, test gaps
- 🟢 Suggestion: code style, naming, simplification opportunities

For each finding, specify:
- File and approximate line
- Severity (critical/warning/suggestion)
- Description
- Suggested fix`,
  { label: 'review-changes', agentType: 'wms-reviewer', schema: {
    type: 'object',
    properties: {
      findings: {
        type: 'array',
        items: {
          type: 'object',
          properties: {
            file: { type: 'string' },
            severity: { type: 'string', enum: ['critical', 'warning', 'suggestion'] },
            description: { type: 'string' },
            suggestion: { type: 'string' },
          },
          required: ['file', 'severity', 'description'],
        },
      },
    },
    required: ['findings'],
  } })
)

const criticalFindings = reviewFindings.findings.filter(f => f.severity === 'critical')
log(`Review found ${reviewFindings.findings.length} issues (${criticalFindings.length} critical)`)

// ── Phase 5: Fix Critical Issues ────────────────────────────
if (criticalFindings.length > 0) {
  phase('Fix')

  await agent(
    `Fix these critical issues found during code review for task [${roadmap.task_id}]:

${JSON.stringify(criticalFindings, null, 2)}

Fix each issue. After fixing, run 'go build ./...' and 'go test ./...' to verify.`,
    { label: 'fix-critical-issues', agentType: 'wms-developer' }
  )

  log('Critical issues fixed')
}

// ── Phase 6: Verify ─────────────────────────────────────────
phase('Verify')

const verifyResult = await agent(
  `Run the full quality check:
1. go build ./...
2. go test ./...
3. go vet ./...

Report: { build_passed: boolean, test_passed: boolean, vet_passed: boolean, test_output_summary: string }`,
  { label: 'quality-gate', schema: {
    type: 'object',
    properties: {
      build_passed: { type: 'boolean' },
      test_passed: { type: 'boolean' },
      vet_passed: { type: 'boolean' },
      test_output_summary: { type: 'string' },
    },
    required: ['build_passed', 'test_passed', 'vet_passed'],
  } }
)

if (!verifyResult.build_passed || !verifyResult.test_passed) {
  log(`❌ Quality gate FAILED: ${verifyResult.test_output_summary}`)
} else {
  log(`✅ All quality checks passed`)
}

return {
  task_id: roadmap.task_id,
  priority: roadmap.priority,
  description: roadmap.description,
  findings_count: reviewFindings.findings.length,
  critical_count: criticalFindings.length,
  quality: verifyResult,
}
