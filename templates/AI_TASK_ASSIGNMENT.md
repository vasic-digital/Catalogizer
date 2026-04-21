# AI Task Assignment Template

> Use this when delegating an implementation task to an AI agent
> (Claude Code, OpenCode, Gemini CLI, Junie, Qwen Code). Paste the
> template, fill the bracketed sections, remove the surrounding
> commentary, and send.
>
> Source: Master Plan v2 §8.2.

---

```
[SYSTEM DIRECTIVE: VERIFICATION MODE ENABLED]

TASK: [FEATURE NAME] in [MODULE NAME].

REQUIRED READING (read before any code change):
1. docs/LANDMINES.md — Section [Go Backend / React/Web / Android / Android TV / Desktop / HelixQA]
2. docs/API_CONTRACTS.md — Endpoint [METHOD PATH] (if API work)
3. CONSTITUTION.md — Article VII (Full-QA Master Cycle), Article VIII (Device State Preservation), Article IX (HelixQA Tool Hygiene)
4. CLAUDE.md project instructions (read the auto-loaded version in your session)
5. AGENTS.md (agent-specific constraints)

CONTEXT:
- Current branch: [branch]
- Related existing work: [file paths, line numbers, or commit SHAs]
- What's already tried or ruled out: [bullet list]

VERIFICATION COMMANDS (must all exit 0 before claiming done):
[pick the relevant chain from templates/VERIFICATION_COMMANDS.md]

CONSTRAINTS:
- Do not claim task complete until ALL verification commands pass
- Do not remove or comment out error handling
- Do not add t.Skip(), it.skip(), or @Ignore without linking to an open issue
- Do not create .disabled files
- Do not write to .env files (gitignored); use .env.example with placeholders
- If you discover a new production rule, append a RULE entry to docs/LANDMINES.md in the same commit as the fix
- Never run commands with sudo/root — RULE-CONST-001
- Every commit goes to all upstream remotes via the standard git push origin main

IF VERIFICATION FAILS:
- Step A: Read the error log in full (do not truncate)
- Step B: Fix the root cause, not the symptom
- Step C: Re-run verification
- Step D: If still failing after 2 attempts, output exactly: "VERIFICATION BLOCKED: <concrete reason>" and stop — do not "mostly fix" the issue

BEGIN TASK.
```

## Notes for the assignor

- Keep the bracket fill-ins concrete. "Fix the login bug" is too vague;
  "fix auth rate-limiter IP-spoof bypass in `catalog-api/internal/auth/middleware.go:285`" is actionable.
- Always include the verification chain. Without it the agent has no
  unambiguous signal that the work is complete.
- The RULE-reporting clause is what keeps `docs/LANDMINES.md` alive —
  every landmine that the agent stepped on should be captured so the
  next agent walks around it.
