# Discovered Findings — 2026-06-26 (during §11.4.170 work)

**Status:** OPEN — tracked follow-ups. Captured by independent read-only subagents (§11.4.142/.150). No code changed. Evidence is the subagents' computed contrast ratios + file:line references.

---

## FINDING-1 — Web + desktop carry the SAME sub-AA error-token-as-surface class as the phone (latent)

**Type:** Bug (accessibility / WCAG). **Severity:** Medium (latent, not yet user-reachable). **Same class as phone fix `d3d00eab`.**

The defect: a dark error color token (`#EF4444`) used as an error **surface/background** paired with a light foreground (`#F8FAFC`) → contrast **3.60:1**, below WCAG AA 4.5:1 for normal text (passes only as large text/UI ≥3:1).

| Client | file:line | Mode | BG | FG | Ratio | AA | State |
|---|---|---|---|---|---|---|---|
| desktop | `catalogizer-desktop/src/styles/globals.css:76-77` | dark | `--destructive`=#EF4444 | `--destructive-foreground`=#F8FAFC | **3.60:1** | FAIL (normal) | **WIRED** via `tailwind.config.js:41-44` → `bg-destructive text-destructive-foreground` is a live class; one component away from rendering (current consumer `src/pages/LoginPage.tsx:63` is a safe 10% tint). LATENT. |
| desktop | `globals.css:46-47` | light | #DC2626 | #FFFFFF | 4.83:1 | PASS | ok |
| web | `catalog-web/src/index.css:52-53` | dark | #EF4444 | #F8FAFC | **3.60:1** | FAIL | DEFINED but DORMANT — not wired into `tailwind.config.js` (no `destructive` color), consumed nowhere. Real destructive surface = `src/components/ui/Button.tsx:11` `bg-red-600 text-white` = #DC2626/#FFFFFF = 4.83:1 PASS. |

Neither client has a WCAG contrast oracle guarding the destructive pair (desktop `tokens.test.ts` only checks tokens.ts↔tokens.css drift) — the exact gap that let the phone defect ship.

**Recommended remediation (NEEDS §11.4.142 review — production change):** apply the same M3-tonal fix as the phone (raise the dark error *surface* tone toward `#FFB4AB` and/or darken the on-error foreground) so dark-mode error surfaces reach ≥4.5:1; add a WCAG contrast oracle test (assert every error-surface↔foreground pair ≥4.5:1) to BOTH web and desktop. Desktop is the priority (wired + latent); web is dormant but should be fixed at the token source to prevent future wiring.

---

## FINDING-2 — Upstream constitution `04d71da` has duplicate anchor numbers §11.4.140 and §11.4.141

**Type:** Task (governance integrity). **Severity:** Medium. **Pre-existing upstream debt — NOT introduced by, and correctly NOT touched by, the §11.4.170 commit `fc84675`.**

In `submodules/constitution/Constitution.md` at upstream HEAD `04d71da`:
- `### §11.4.140` appears **twice**: "Universal action-prefix system" (~line 8996, 2026-06-09) AND "Mandatory HelixTranslate canonical translation pipeline" (~line 9821, 2026-06-25, renamed from old §11.4.133 by `04d71da`).
- `### §11.4.141` appears **twice**: "Token-efficiency mandate" (~line 9091, 2026-06-09) AND "Mandatory independent per-language translation review" (~line 9851, 2026-06-25, new in `04d71da`).

The `04d71da` commit ("fix §11.4.140 numbering collision") moved the translation block off the §11.4.133 collision but onto already-occupied 140/141.

**Why not fixed here:** renumbering an anchor changes its identity + every cross-reference + the propagation-gate literal (`CM-COVENANT-114-140/141-PROPAGATION`) across the whole consumer fleet — high-blast-radius, operator-gated (§11.4.101). Also the translation clauses are the upstream author's; renumbering them without coordination risks colliding with their next push.

**Recommended:** operator decides the free numbers for the translation pipeline + translation review clauses (e.g. §11.4.171/§11.4.172), then a dedicated commit renumbers the two blocks + their gate literals + carrier mirrors, with §11.4.157 lockstep (the carriers also lag — `04d71da` added 140/141 to Constitution.md only, not to CLAUDE/AGENTS/QWEN/GEMINI).
