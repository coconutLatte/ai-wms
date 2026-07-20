# Contributing to AI-WMS

## Two Ways to Contribute

### 1. Human Contributions

Found a bug? Have a feature idea? Want to improve docs?

- **Issues**: Open a GitHub issue — the AI evolution system will read it and may pick it up
- **Pull Requests**: Fork, branch, and PR — humans and AI co-evolve this codebase
- **Roadmap**: Suggest tasks by opening an issue with the label `roadmap`

### 2. Let the AI Evolve It

The self-evolution engine runs every 10 minutes. It:
1. Reads `docs/roadmap.md` for pending tasks
2. Implements the highest-priority one
3. Runs tests and quality checks
4. Commits and pushes

You can watch it happen: `tail -f logs/cron-evolve.log`

## Development Setup

```bash
# Prerequisites
go >= 1.26
docker

# Clone and setup
git clone git@github.com:coconutLatte/ai-wms.git
cd ai-wms
make setup

# Run tests
make test
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for the full architecture overview.

## Commit Conventions

- `feat(scope): description` — new feature
- `fix(scope): description` — bug fix
- `chore(scope): description` — maintenance
- `docs(scope): description` — documentation

All commits include: `Co-Authored-By: <model> <noreply@anthropic.com>`
