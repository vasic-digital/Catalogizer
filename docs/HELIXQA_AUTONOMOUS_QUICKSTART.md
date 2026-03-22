# HelixQA Autonomous QA Session - User Guide

## Quick Start

```bash
# Run autonomous QA session
cd HelixQA
./helixqa autonomous \
  --project /path/to/project \
  --platforms desktop \
  --env ../.env \
  --output ./qa-results
```

## Configuration

Create `.env` file:
```bash
HELIX_AUTONOMOUS_ENABLED=true
HELIX_AUTONOMOUS_PLATFORMS=desktop,web
ANTHROPIC_API_KEY=sk-ant-...
HELIX_AGENTS_ENABLED=opencode,claude-code
```

## Session Phases

1. **Setup** - Initialize agents and start recording
2. **Doc-Driven** - Verify documented features
3. **Curiosity** - Explore and find issues
4. **Report** - Generate comprehensive report

## Reports

Located in `qa-results/`:
- `qa-report.md` - Main report
- `tickets/HQA-*.md` - Individual issues
- `screenshots/` - Evidence
- `videos/` - Session recordings

## Troubleshooting

**"No agents available"** - Check CLI agent installation
**"Vision provider failed"** - Verify API key
**"Session timeout"** - Increase timeout or reduce coverage target

See full documentation in [Implementation Plan](../HELIXQA_AUTONOMOUS_QA_IMPLEMENTATION_PLAN.md)
