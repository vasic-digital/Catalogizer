# OpenDesign Web-Client Integration Plan (catalog-web)

**Revision:** 1
**Last modified:** 2026-06-25T13:18:00Z
**Authority:** Constitution §11.4.162 (OpenDesign UI design system mandate), §11.4.99 (latest-source cross-reference), §11.4.6 (no-guessing)
**Scope:** `/Volumes/T7/Projects/catalogizer/catalog-web` — the React/Vite web client only.
**Status:** PLAN — documentation only. No app source modified. No commit.
**Upstream context:** `docs/design/OPENDESIGN_INTEGRATION_AUDIT.md` (unified "Catalogizer Blue" token spec).

---

## §0 — Executive summary (FACTS)

- **Toolchain:** React 18 + TypeScript 4.9 + Vite 6, styled with **Tailwind CSS 3.2 + shadcn-style HSL CSS variables**, theme tokens in `catalog-web/src/index.css` and `catalog-web/tailwind.config.js`. (`package.json:31,78,80`; `index.css:1-49`; `tailwind.config.js:1-154`.)
- **Light/dark:** `darkMode: 'class'` is configured (`tailwind.config.js:7`) and a full `.dark{}` block exists (`index.css:29-49`), **but NOTHING in `src/` ever applies the `.dark` class** — confirmed by grep (only `src/index.css` matches `.dark`; zero `documentElement`/`classList`/`'dark'` toggles in `src/`). The dark theme is currently **dead CSS**. No theme-toggle component exists.
- **OpenDesign npm-consumable? NO.** Verified against latest sources (§11.4.99): `github.com/nexu-io/open-design` is a **standalone desktop/web AI design application** ("open-source Claude Design alternative"), publishes **no npm package and no design-token export**; the only npm package named `opendesign` (v0.3.0) is an **unrelated** `.octopus` file CLI from `opendesigndev`. Therefore the **design-token-file fallback** (a tokens module encoding "Catalogizer Blue") is the honest, mandated path — see §3.
- **Files to ADD:** **4** · **Files to MODIFY:** **4** (see §5 manifest).

---

## §1 — Toolchain (file:line evidence)

| Aspect | Finding | Evidence |
|---|---|---|
| Framework | React 18 + TypeScript + Vite 6 | `catalog-web/package.json:31-32` (`react`/`react-dom` `^18.2.0`), `:79` (`typescript ^4.9.3`), `:80` (`vite ^6.2.0`) |
| Build scripts | `dev`=vite, `build`=`tsc && vite build`, `test`=vitest, `test:e2e`=playwright | `package.json:84-99` |
| Styling system | **Tailwind CSS 3.2** utility-first | `package.json:78` (`tailwindcss ^3.2.7`); `index.css:1-3` (`@tailwind base/components/utilities`) |
| Token strategy | shadcn-style **HSL CSS custom properties** mapped into Tailwind `colors` | `index.css:6-49` (`:root`/`.dark` HSL vars); `tailwind.config.js:11-15` (`border/input/ring/background/foreground = hsl(var(--…))`) |
| **Where the palette lives** | TWO parallel sources: (a) HSL CSS vars in `src/index.css:6-49`; (b) a **static `primary` 50–950 scale** + `secondary/accent/success/warning/error` scales in `tailwind.config.js:16-93` | both files cited |
| Fonts | Inter (sans) + JetBrains Mono (mono), loaded via Google Fonts | `tailwind.config.js:95-98`; `index.html:11` |
| Radius | `--radius: 0.75rem` | `index.css:26` |
| Spacing extras | `18`=4.5rem, `88`=22rem, `128`=32rem; `borderRadius 4xl`=2rem; `maxWidth 8xl/9xl` | `tailwind.config.js:135-146` |
| Tailwind plugins | forms, typography, aspect-ratio | `tailwind.config.js:149-153` |
| Shared UI lib | `@vasic-digital/ui-components` (`file:../submodules/ui_components_react`) | `package.json:21` |

### §1.1 — Load-bearing complication: ui-components hardcodes colors

The shared `@vasic-digital/ui-components` `Button` does **NOT** consume the CSS-var/Tailwind tokens — it hardcodes literal Tailwind palette classes:

- `submodules/ui_components_react/src/components/Button.tsx:23` → `primary: 'bg-blue-600 text-white hover:bg-blue-700 …'`
- `:27` → `destructive: 'bg-red-600 …'`
- `Switch.tsx:35-38` → `bg-blue-600 dark:bg-blue-500` / `focus:ring-blue-500`

Consequence: **changing the `--primary` CSS var alone will NOT recolor these shared components.** The plan must either (a) re-point ui-components to token classes (`bg-primary` / `bg-primary-600`), or (b) keep `primary-600` aligned to the brand blue so `bg-blue-600`≈brand by coincidence. Option (a) is correct per §11.4.162 single-source intent; it is an **owned submodule** (§11.4.28 equal-codebase) so the change belongs there, but is OUT of this web-only plan's file scope and is recorded as a cross-cutting dependency (§4.4).

---

## §2 — Light/dark mechanism

**FACT:** There is currently **no working dark mode and no theme toggle.** Evidence:

- `darkMode: 'class'` set (`tailwind.config.js:7`) and `.dark{}` overrides defined (`index.css:29-49`).
- `src/main.tsx:1-66` mounts `HelmetProvider → QueryClientProvider → App` with **no theme provider** and **no `.dark` class application**.
- Grep for `'dark'` / `documentElement` / `classList` / `toggleTheme` / `ThemeProvider` / `prefers-color-scheme` across `src/` → **zero matches** (only `src/index.css` matches `.dark`).

So the `.dark` palette is dead code: the toggle was never wired. OpenDesign (§11.4.162) requires **every component ship light + dark variants** — this is currently UNMET because dark is unreachable.

### §2.1 — Minimal idiomatic dark-mode mechanism for this stack (Tailwind `class` strategy)

Idiomatic for Tailwind `darkMode:'class'` + React + Zustand (already a dep, `package.json:46`):

1. **Add `src/contexts/ThemeContext.tsx`** — a `ThemeProvider` + `useTheme()` hook that:
   - reads initial theme from `localStorage('catalog-web-theme')` else `window.matchMedia('(prefers-color-scheme: dark)')`;
   - applies/removes `'dark'` on `document.documentElement.classList` in a `useEffect`;
   - persists the choice to `localStorage`.
2. **Add `src/components/ThemeToggle.tsx`** — a button using the existing `lucide-react` (`package.json:30`) Sun/Moon icons + the `ui-components` `Switch` or `Button`, calling `useTheme().toggle()`. Reuse the global focus-visible ring already defined (`index.css:61-69`).
3. **Wire** `<ThemeProvider>` into `src/main.tsx` (wrap `<App/>`) and drop `<ThemeToggle/>` into the app header/settings.

No new dependency required — Zustand or plain Context both suffice; Context is the lightest. (`UNCONFIRMED:` exact header/settings component to host the toggle — not yet located; resolve during implementation by reading `src/components/` layout shell.)

---

## §3 — OpenDesign reality check + token-file fallback (§11.4.99, §11.4.6)

### §3.1 — Verified findings (latest-source, 2026-06-25)

| Question | Finding (FACT) | Source |
|---|---|---|
| Does `github.com/nexu-io/open-design` publish an npm package? | **NO.** It is a standalone local-first **AI design application** (desktop `.app`/Docker/`pnpm` from source), "open-source Claude Design alternative". No npm module, no token export. | WebFetch `github.com/nexu-io/open-design` README, 2026-06-25 |
| Is there ANY npm `opendesign`? | Yes but **UNRELATED** — `opendesign` v0.3.0 by `opendesigndev` is a CLI for `.octopus` design files, not a token/theme system. | WebFetch `registry.npmjs.org/opendesign`, 2026-06-25 |
| Can this stack `npm install` OpenDesign tokens cleanly? | **NO** — neither package exposes consumable design tokens (CSS vars / JSON / Tailwind preset) for a React+Tailwind app. | both above |

**Honest conclusion (§11.4.6):** OpenDesign cannot be cleanly consumed as an npm token package by catalog-web. Claiming an `npm install @nexu-io/open-design` integration would be a bluff. The §11.4.74 "extend-don't-reimplement / consume as dependency" path does not apply to a token package because none exists. We therefore implement the **mandated fallback the task specifies**: a project-owned **design-token module** that encodes the unified "Catalogizer Blue" palette + typography + spacing, structured to mirror OpenDesign's token taxonomy (`color.*`, `typography.*`, `space.*`, `radius.*`) from the audit §3, so a future real OpenDesign export can drop in as the upstream source.

### §3.2 — Token-file design (the fallback)

Create a single canonical TS token module — **`src/styles/tokens.ts`** — that is the one source of truth and matches the OpenDesign taxonomy from `OPENDESIGN_INTEGRATION_AUDIT.md §3.1-§3.3`:

```ts
// src/styles/tokens.ts  (illustrative shape — values from audit §3.1)
export const catalogizerBlue = {
  color: {
    brand:    { primary:{light:'#1565C0',dark:'#9ECAFF'},
                primaryHover:{light:'#1976D2',dark:'#7FB4F5'},
                onPrimary:{light:'#FFFFFF',dark:'#003258'},
                secondary:{light:'#535F70',dark:'#BBC7DB'},
                onSecondary:{light:'#FFFFFF',dark:'#253140'},
                accent:{light:'#6B5778',dark:'#D6BEE4'} },
    semantic: { success:{light:'#16A34A',dark:'#22C55E'},
                warning:{light:'#D97706',dark:'#F59E0B'},
                error:{light:'#DC2626',dark:'#EF4444'} },
    neutral:  { background:{light:'#FFFFFF',dark:'#020817'},
                surface:{light:'#F8FAFC',dark:'#0F172A'},
                foreground:{light:'#020817',dark:'#F8FAFC'},
                muted:{light:'#64748B',dark:'#94A3B8'},
                border:{light:'#E2E8F0',dark:'#1E293B'} },
    state:    { ring:{light:'#1565C0',dark:'#9ECAFF'} },
  },
  typography: { fontFamily:{ sans:'Inter, system-ui, sans-serif',
                             mono:'JetBrains Mono, monospace' } },
  space:  { '18':'4.5rem','88':'22rem','128':'32rem' },
  radius: { base:'0.75rem', '4xl':'2rem' },
} as const;
```

All values above are copied verbatim from the audit §3.1–§3.3 (themselves traced to real per-app source files). The CSS-var layer (`index.css`) and the Tailwind config consume FROM this module so a brand change is one edit.

> Note (§11.4.6): shadcn Tailwind tokens use **HSL channel triplets** in the CSS vars (e.g. `--primary: 221.2 83.2% 53.3%`). The token module stores hex; an implementation-time helper (`hexToHslTriplet`) converts hex→`H S% L%` so `index.css` keeps its `hsl(var(--primary))` contract intact. (`#1565C0` → `210 79% 42%`; recompute exactly during implementation, do not hand-eyeball.)

---

## §4 — Concrete change list (the unified Catalogizer Blue tokens, light + dark)

### §4.1 — `src/index.css` (MODIFY `:root` lines 6-27 and `.dark` lines 29-49)

Replace the literal HSL triplets so they resolve to the audit §3.1 brand values. Specific changes:

| CSS var | Current (`index.css`) | Target (Catalogizer Blue, from audit §3.1) |
|---|---|---|
| `--primary` (`:root`) | `221.2 83.2% 53.3%` (`#2563EB`) | `#1565C0` → recompute HSL (`:13`) |
| `--primary` (`.dark`) | `217.2 91.2% 59.8%` (`#3B82F6`) | `#9ECAFF` → recompute HSL (`:36`) |
| `--primary-foreground` light/dark | `210 40% 98%` / `222.2 84% 4.9%` | onPrimary `#FFFFFF` / `#003258` (`:14,37`) |
| `--ring` light | `221.2 83.2% 53.3%` | `#1565C0` (= primary) (`:25`) |
| `--ring` dark | `224.3 76.3% 94.1%` | `#9ECAFF` (fix the divergent value) (`:48`) |
| `--accent` light/dark | `210 40% 96%` / `217.2 32.6% 17.5%` | brand accent `#6B5778` / `#D6BEE4` (`:19,42`) — *operator decision per §11.4.66 if product wants to keep the neutral accent* |
| `--destructive` | `0 84.2% 60.2%` / `0 62.8% 30.6%` | error `#DC2626` / `#EF4444` (audit §3.1) (`:21,44`) |
| `--background`/`--foreground`/`--border`/`--input` | already match audit neutrals (`#FFFFFF`/`#020817`/`#E2E8F0`/`#1E293B`) | **no change** (`:7-8,23-24,30-31,46-47`) |
| `--radius` | `0.75rem` | **no change** (already = `radius.base`) (`:26`) |

Also update the hardcoded focus ring `ring-blue-500` (`index.css:68`) → `ring-ring` (or `ring-[hsl(var(--ring))]`) so focus colour follows the brand token instead of a literal blue.

### §4.2 — `tailwind.config.js` (MODIFY lines 16-98)

- **Collapse the duplicate static `primary` 50–950 scale** (`:16-28`) onto the brand token (remove drift vs the CSS var). Keep the 50–950 ramp only if a consumer needs shades; if kept, re-anchor `600` to the brand `#1565C0` so `primary-600` and `--primary` agree.
- **Re-point `accent`** (currently fuchsia `#d946ef`, `:42-54`) → brand accent `#6B5778` ramp (audit §4.1) — *operator decision per §11.4.66 if the fuchsia accent is intentional product branding.*
- Keep `success`/`warning`/`error` scales (`:55-93`) — already match audit semantic anchors (`success-600 #16a34a`, `warning-600 #d97706`, `error-600 #dc2626`).
- Keep `fontFamily` (`:95-98`), spacing (`:135-139`), radius (`:140-142`) — already aligned with audit §3.2/§3.3.
- (Optional, cleaner) add a token import: `import { catalogizerBlue } from './src/styles/tokens'` and reference it instead of inline literals, so config + CSS share one source.

### §4.3 — Files to ADD

| File | Purpose | §ref |
|---|---|---|
| `src/styles/tokens.ts` | Canonical Catalogizer Blue token module (OpenDesign taxonomy) | §3.2 |
| `src/contexts/ThemeContext.tsx` | `ThemeProvider` + `useTheme()` (applies `.dark`, persists to localStorage) | §2.1 |
| `src/components/ThemeToggle.tsx` | Sun/Moon toggle wired to `useTheme()` | §2.1 |
| `e2e/tests/visual-regression-dark.spec.ts` *(or extend existing)* | Dark-theme screenshot baselines for key surfaces | §6 |

### §4.4 — Cross-cutting dependency (OUT of web-only file scope, MUST be tracked)

`@vasic-digital/ui-components` (`submodules/ui_components_react/src/components/Button.tsx:23,27`, `Switch.tsx:35-38`) hardcodes `bg-blue-600`/`bg-red-600`/`ring-blue-500`. To make the single-token-source guarantee real, those literals must become token classes (`bg-primary` / `bg-destructive` / `ring-ring`). This is an **owned submodule** (§11.4.28) change — file it as a separate tracked work item; it is a prerequisite for full §11.4.162 compliance but lives outside catalog-web's tree.

---

## §5 — File manifest (counts)

**ADD (4):** `src/styles/tokens.ts`, `src/contexts/ThemeContext.tsx`, `src/components/ThemeToggle.tsx`, `e2e/tests/visual-regression-dark.spec.ts`.
**MODIFY (4):** `src/index.css`, `tailwind.config.js`, `src/main.tsx` (wrap `<ThemeProvider>`), host header/settings component for `<ThemeToggle/>` (`UNCONFIRMED:` exact file — locate during impl).
**CROSS-CUTTING (tracked separately, NOT in web tree):** `submodules/ui_components_react/src/components/{Button,Switch}.tsx`.

---

## §6 — Visual-regression testing (per §11.4.162)

**FACT — VR tooling already exists.** `catalog-web` already runs Playwright VR via `expect(page).toHaveScreenshot(...)` in `e2e/tests/visual-regression.spec.ts` (e.g. `:15` `login-page.png`, `:71` `dashboard.png`, `threshold: 0.2`). Config: `playwright.config.ts` (`testDir './e2e'`, projects chromium/firefox/webkit/Mobile-Chrome, `webServer` auto-starts `npm run dev`).

### §6.1 — Exact commands

```bash
cd /Volumes/T7/Projects/catalogizer/catalog-web
npm run playwright:install                       # one-time browser install (package.json:99)
npm run test:e2e                                 # run all e2e incl. VR (package.json:94)
npx playwright test e2e/tests/visual-regression.spec.ts                 # VR only
npx playwright test e2e/tests/visual-regression.spec.ts --update-snapshots   # (re)baseline after intended token change
npx playwright test --project=chromium e2e/tests/visual-regression.spec.ts   # single browser
```

### §6.2 — What to add for OpenDesign tokens + dark mode

1. **Dark-theme baselines:** for every existing VR surface (login, dashboard, browse, media-detail, settings) add a `.dark`-toggled screenshot. Toggle by adding the class before capture:
   `await page.evaluate(() => document.documentElement.classList.add('dark'));` then `await expect(page).toHaveScreenshot('dashboard-dark.png', {fullPage:true, threshold:0.2});`.
2. **Token-source unit test** (`src/styles/__tests__/tokens.test.ts`, vitest): assert the generated CSS-var values equal the `tokens.ts` canonical set (catches drift — the single-source guarantee). Run via `npm run test`.
3. **Contrast/overlap assertions:** the project already has `@axe-core/playwright` (`package.json:49`) and `e2e/accessibility.spec.ts` — extend it to assert `onPrimary`-on-`primary` and `foreground`-on-`background` meet WCAG 2.1 AA in BOTH themes (carry the Android HELIX-175 methodology cited in the audit §5).
4. **Self-validated comparator (§11.4.107(10) / §1.1):** add a paired mutation test that deliberately corrupts one baseline pixel-set and asserts `toHaveScreenshot` FAILs — proving the VR gate is not a bluff. Without this, a green VR run is not trustworthy evidence.

### §6.3 — Honest gap (§11.4.6)

Playwright `toHaveScreenshot` baselines are platform-font-renderer sensitive; baselines committed on one OS may diff on another. Generate/refresh baselines in the same container the CI/operator uses (rootless Podman per project policy), or pin a renderer. Flag as `UNCONFIRMED:` until the baseline-generation environment is fixed.

---

## §7 — Sequencing (impl order)

1. Add `src/styles/tokens.ts` (§3.2) — values from audit §3.1, no behaviour change yet.
2. Re-point `index.css` (§4.1) + `tailwind.config.js` (§4.2) at the brand values; run `npm run test:e2e ... --update-snapshots` to rebaseline light theme.
3. Add `ThemeContext` + `ThemeToggle`, wire into `main.tsx` (§2.1); now `.dark` is reachable.
4. Add dark VR baselines + token-source unit test + axe contrast checks + self-validated mutation (§6.2).
5. (Tracked separately) re-point ui-components hardcoded colours (§4.4).

---

## Sources verified (2026-06-25)

- In-repo: `catalog-web/package.json`, `src/index.css`, `tailwind.config.js`, `index.html`, `src/main.tsx`, `playwright.config.ts`, `e2e/tests/visual-regression.spec.ts`, `e2e/` listing, `submodules/ui_components_react/src/components/{Button,Switch}.tsx`, `docs/design/OPENDESIGN_INTEGRATION_AUDIT.md`.
- Online (§11.4.99): WebFetch `github.com/nexu-io/open-design` (standalone AI design app, no npm/token export); WebFetch `registry.npmjs.org/opendesign` (unrelated `.octopus` CLI v0.3.0). Both fetched 2026-06-25.

> OpenDesign provides NO consumable token package for this stack; the design-token-file fallback (§3) is the mandated, honest path. Re-verify OpenDesign's surface before any future "real OpenDesign export" swap-in.
