# OpenDesign Desktop Integration Plan (catalogizer-desktop)

**Revision:** 1
**Last modified:** 2026-06-25T00:00:00Z
**Authority:** Constitution §11.4.162 (OpenDesign UI design system mandate)
**Scope:** `catalogizer-desktop` (Tauri 2 + React 18 + TypeScript + Vite + Tailwind CSS)
**Companion audit:** `docs/design/OPENDESIGN_INTEGRATION_AUDIT.md` (Rev 1) — defines the unified "Catalogizer Blue" token spec
**Status:** PLAN — documentation only. No app source modified. No commit. (Anti-bluff §11.4: every claim below cites a file I read.)

---

## §0 — Executive summary (FACTS)

- The desktop client is at `/Volumes/T7/Projects/catalogizer/catalogizer-desktop`. It is a **Tauri 2** app: React 18 frontend + Rust backend.
  Confirmed from `catalogizer-desktop/package.json:2` (`"name": "catalogizer-desktop"`), `package.json:19` (`@tauri-apps/api ^2.10.1`), `package.json:47` (`tailwindcss ^3.3.2`), and `catalogizer-desktop/src-tauri/Cargo.toml` + `src-tauri/tauri.conf.json` (Rust backend present).
- The color palette lives in **one place**: `catalogizer-desktop/src/styles/globals.css:6-60` (shadcn-style HSL CSS custom properties in `:root` / `.dark`), surfaced to Tailwind via `catalogizer-desktop/tailwind.config.js:9-44`.
- The current `--primary` is **monochrome slate** (`222.2 47.4% 11.2%` ≈ `#0F172A` light / `210 40% 98%` ≈ `#F8FAFC` dark — `globals.css:13,42`), the unmodified shadcn default. This is the brand-divergence the audit flags (§1.4 of the audit): desktop is the only app whose primary is NOT blue.
- Light/dark **does** exist and is wired at runtime via a `.dark` class on `document.documentElement` toggled in `catalogizer-desktop/src/stores/configStore.ts:36-41,67-72,110-111`. **Gap:** the toggle only special-cases `'dark'`; `'light'` and `'system'` both fall into the same `else`/remove branch, so the `'system'` value declared in `src/types/index.ts:225` is NOT honoured (no `prefers-color-scheme` query anywhere — confirmed by grep).
- **OpenDesign is NOT a dependency** of this app (zero matches in `package.json`). Whether it is cleanly consumable as an npm package is **UNCONFIRMED** (see §3). The safe, working path is the **design-token-module fallback** (§3.2) that encodes the audit's Catalogizer Blue tokens in the same shadcn-CSS-variable shape the app already consumes.
- **No visual-regression tooling exists.** Playwright is NOT a devDependency (zero in `package.json`; the two `package-lock.json` hits are transitive peer refs, no `e2e/` dir, `node_modules/.bin/playwright` absent). Only **Vitest** is real (`package.json:13-15`, `vitest.config.ts`).

---

## §1 — Toolchain (answers question 1)

| Aspect | FACT | Evidence (file:line) |
|---|---|---|
| Framework | Tauri 2 (Rust backend + React 18 webview frontend) | `package.json:19` `@tauri-apps/api ^2.10.1`; `package.json:31` `@tauri-apps/cli >=2.0.0`; `src-tauri/tauri.conf.json` present |
| Frontend language | TypeScript 5, React 18.2 | `package.json:24-25,48`; `tsconfig.json` |
| Bundler / dev server | Vite 4 (dev server :1420 per `CLAUDE.md`) | `package.json:7-9,49`; `vite.config.ts` |
| Styling | Tailwind CSS 3 + shadcn-style HSL CSS variables | `package.json:47`; `tailwind.config.js:1-48`; `src/styles/globals.css:1-70` |
| Class-merge util | `clsx` + `tailwind-merge` (via `src/utils/cn.ts`) | `package.json:22,27`; `CLAUDE.md` directory table |
| Build script | `tsc && vite build` | `package.json:8` |
| Test script | `vitest run` (+ `:watch`, `:coverage`) | `package.json:13-15`; `vitest.config.ts` |
| Tauri build | `tauri build` (AppImage/dmg/msi) | `package.json:12` |
| **Palette / theme location** | **`src/styles/globals.css:6-60`** (CSS vars) → exposed in **`tailwind.config.js:9-44`** | read both files |
| CSS entry import | `globals.css` imported once in `src/main.tsx:6` | `grep` hit |
| Font family | **NONE declared** (browser/OS default; no `fontFamily` extend) | `tailwind.config.js:7-46` has no `fontFamily`; no font `<link>` in `index.html:1-13` |
| Radius | `--radius: 0.5rem` | `globals.css:32` |

---

## §2 — Light/dark mechanism (answers question 2)

**Mechanism exists.** It is a class-based dark mode driven by the Zustand config store:

- The `.dark` CSS class is added/removed on `document.documentElement` in three places:
  `configStore.ts:36-41` (on `loadConfig`), `configStore.ts:67-72` (on `setTheme`), `configStore.ts:110-111` (on `resetConfig`).
- The corresponding token override block is `globals.css .dark { … }` at `globals.css:35-60`.
- The user-facing control is the theme `<select>` in `src/pages/SettingsPage.tsx:236` (`<option value="dark">Dark</option>`), wired through `SettingsPage.tsx:11,122` → `configStore.setTheme`.
- The `Theme` union is `'light' | 'dark' | 'system'` (`src/types/index.ts:225`).

> NOTE on Tailwind dark-mode config: `tailwind.config.js` does NOT set `darkMode: 'class'`. The app does not rely on Tailwind's `dark:` variant utilities — it relies on the CSS variables flipping inside the `.dark` selector block (`globals.css:35-60`), which works regardless of the Tailwind `darkMode` setting because the variables are plain CSS, not `dark:`-prefixed utilities. If any future component uses `dark:` utility variants, `darkMode: 'class'` MUST be added to `tailwind.config.js`. **UNCONFIRMED:** whether any current component uses `dark:` utilities (not exhaustively grepped); the variable-flip path is the proven mechanism.

**Real gaps to fix as part of this work:**
1. `'system'` is a declared `Theme` value (`types/index.ts:225`) but `configStore.ts` has no `window.matchMedia('(prefers-color-scheme: dark)')` handling — selecting "system" silently behaves as light. The minimal idiomatic fix: in `configStore.ts`, resolve `'system'` via `matchMedia` and add a `change` listener.
2. `SettingsPage.tsx:236` only renders `<option value="dark">`; confirm `light`/`system` options exist (the audit/grep shows only the `dark` option line). **UNCONFIRMED** whether `light`/`system` options are rendered — read `SettingsPage.tsx:225-245` before editing.

---

## §3 — OpenDesign reality check (answers question 3)

### §3.1 Honest finding

- OpenDesign is **not** installed: zero matches for `open-design` / `opendesign` / `nexu` in `catalogizer-desktop/package.json`.
- **UNCONFIRMED:** whether `github.com/nexu-io/open-design` publishes a consumable npm package, a design-token export format (Style Dictionary / W3C DTCG JSON / CSS-vars generator), or only a hosted design tool. As a subagent I did not fetch the upstream repo/npm registry. Per §11.4.99 + §11.4.123, the conductor MUST cross-reference the latest OpenDesign docs/npm before committing to a runtime-dependency integration. **I will not claim a working npm integration I have not verified.**
- This stack **can** consume any design-token system that emits **CSS custom properties** (HSL or hex), because the app's entire theming surface is already CSS variables (`globals.css:6-60`) read through Tailwind (`tailwind.config.js:10-38`). So: IF OpenDesign emits CSS-vars (directly or via a Style-Dictionary "css" platform), it is cleanly consumable here. IF it is only a hosted/Figma design tool with no code export, it is NOT consumable as a runtime dependency and the fallback (§3.2) is the answer.

### §3.2 Recommended path — design-token-module fallback (works regardless of OpenDesign packaging)

Encode the audit's "Catalogizer Blue" tokens (audit §3.1–§3.3) as a **single source-of-truth token module** in the app, shaped exactly like OpenDesign's structure (color / typography / spacing / radius), so that when/if OpenDesign's exact export format is confirmed, the file is swapped for generated output with no consumer changes. Two equivalent encodings — pick the CSS-vars one (zero new build step):

- **Tokens as CSS variables** (recommended, no build tooling): a `src/styles/tokens.css` authored from the audit, `@import`ed at the top of `globals.css`. `globals.css`'s `:root`/`.dark` then reference the token vars. This is a drop-in for the existing variable mechanism.
- **Tokens as a typed TS module** (optional, enables the §5 token-source test): `src/styles/tokens.ts` exporting the canonical light/dark token maps; a small generator (or hand-mirrored CSS) keeps `tokens.css` in sync. The TS module is what the drift test (§5.1) asserts against.

This fallback satisfies §11.4.162's "use design tokens/themes for color palette (light+dark), typography, spacing" using the audit's values, without claiming a non-working OpenDesign runtime wire-up.

### §3.3 Audit token values to encode (light / dark)

From the audit §3.1–§3.3 (the conductor MUST re-verify WCAG AA per audit §5 before finalizing):

| Token | Light | Dark |
|---|---|---|
| brand.primary | `#1565C0` | `#9ECAFF` |
| brand.primary.hover | `#1976D2` | `#7FB4F5` |
| brand.onPrimary | `#FFFFFF` | `#003258` |
| brand.secondary | `#535F70` | `#BBC7DB` |
| brand.accent | `#6B5778` | `#D6BEE4` |
| semantic.success | `#16A34A` | `#22C55E` |
| semantic.warning | `#D97706` | `#F59E0B` |
| semantic.error | `#DC2626` | `#EF4444` |
| neutral.background | `#FFFFFF` | `#020817` |
| neutral.surface | `#F8FAFC` | `#0F172A` |
| neutral.foreground | `#020817` | `#F8FAFC` |
| neutral.muted | `#64748B` | `#94A3B8` |
| neutral.border | `#E2E8F0` | `#1E293B` |
| state.ring | `#1565C0` | `#9ECAFF` |
| font.family.sans | `Inter, system-ui, sans-serif` | (same) |
| font.family.mono | `JetBrains Mono, monospace` | (same) |
| radius.base | `0.75rem` (`lg`); `md`=−2px; `sm`=−4px | (same) |

> The app's CSS variables are **HSL triplets** (e.g. `--primary: 222.2 47.4% 11.2%` at `globals.css:13`). The hex values above MUST be converted to the `H S% L%` triplet form when written into `globals.css` (the audit already did the reverse conversion; reuse `colorsys`/equivalent and keep both forms documented).

---

## §4 — Concrete change list (answers question 4)

> All paths absolute. ADD = new file; MODIFY = edit existing. No commit (conductor owns commits per §11.4.84).

### ADD
1. `/Volumes/T7/Projects/catalogizer/catalogizer-desktop/src/styles/tokens.css`
   — Catalogizer Blue tokens (§3.3) as CSS variables, light values under `:root`, dark overrides under `.dark`, in `H S% L%` triplet form. Single source of truth for the desktop palette.
2. `/Volumes/T7/Projects/catalogizer/catalogizer-desktop/src/styles/tokens.ts` (optional but recommended)
   — typed `{ light, dark }` token maps mirroring `tokens.css`; consumed by the §5.1 drift test.
3. `/Volumes/T7/Projects/catalogizer/catalogizer-desktop/src/styles/__tests__/tokens.test.ts`
   — token-source drift test (Vitest) asserting `globals.css`/`tokens.css` values equal the canonical `tokens.ts` set (anti-bluff per §11.4.107(10) / §1.1; self-validates with a deliberately-wrong fixture that MUST FAIL).
4. Visual-regression scaffold (see §5): `playwright.config.ts` + `e2e/theme.spec.ts` + Playwright as a devDependency (NEW — none exists today).

### MODIFY
5. `catalogizer-desktop/src/styles/globals.css`
   — `:6-33` (`:root`): repoint `--primary`, `--primary-foreground`, `--secondary*`, `--accent*`, `--ring`, `--border`, `--input`, `--muted*` to the §3.3 Catalogizer Blue light values (HSL triplets). Critically `--primary` `222.2 47.4% 11.2%` → blue `#1565C0` triplet (`globals.css:13`).
   — `:35-60` (`.dark`): same repoint to §3.3 dark values; `--primary` → `#9ECAFF` triplet, `--ring` → `#9ECAFF` (`globals.css:42,59`).
   — `:32` (`--radius`): `0.5rem` → `0.75rem` (unify with audit `radius.base`; web app already at `0.75rem`).
   — `:1-3`: add `@import "./tokens.css";` (or fold tokens directly) so `:root`/`.dark` reference token vars.
6. `catalogizer-desktop/tailwind.config.js`
   — `:7-46` `theme.extend`: ADD a `fontFamily` block (`sans: ['Inter', 'system-ui', 'sans-serif']`, `mono: ['JetBrains Mono', 'monospace']`) — desktop currently declares NO brand font (`tailwind.config.js` has no `fontFamily`). Optionally ADD `darkMode: 'class'` (top level) to make the `.dark` toggle authoritative for any future `dark:` utilities.
7. `catalogizer-desktop/index.html`
   — `:3-8` `<head>`: add Inter + JetBrains Mono font loading (Google Fonts `<link>` OR a bundled `@fontsource/*` import in `main.tsx`) so `font.family.sans` resolves. (`index.html` currently loads no font.)
8. `catalogizer-desktop/src/stores/configStore.ts`
   — `:36-41,67-72,110-111`: replace the `if 'dark' … else` branches with a `resolveTheme(theme)` helper that handles `'light'` (remove `.dark`), `'dark'` (add `.dark`), and `'system'` (read `window.matchMedia('(prefers-color-scheme: dark)')` + attach a `change` listener). Closes the `'system'` gap (`types/index.ts:225`).
9. `catalogizer-desktop/src/pages/SettingsPage.tsx`
   — `:225-245` (the theme `<select>`): ensure `light` and `system` `<option>`s are present (only the `dark` option is confirmed at `:236`). **Read first** — may already exist.

### NO CHANGE NEEDED
- `tailwind.config.js:9-39` color mappings already read every `hsl(var(--token))` — repointing the variable values cascades automatically; no Tailwind color map edits required.
- `src/utils/cn.ts`, all component files — they consume the semantic Tailwind classes (`bg-primary`, `text-foreground`, …); the brand change is purely token-level.

---

## §5 — Visual-regression / UI test plan (answers question 5)

### §5.0 What exists today (FACT)
- **Unit/component tests:** Vitest + React Testing Library, jsdom env (`vitest.config.ts:6-9`), 80% coverage thresholds (`vitest.config.ts:22-27`), 100+ test files under `src/**/__tests__/`. Run: `npm run test` (`package.json:13`), coverage `npm run test:coverage` (`package.json:15`).
- **Visual regression:** **NONE.** No Playwright devDependency (`package.json` has zero), no `e2e/` dir, no `playwright.config.ts`, no `node_modules/.bin/playwright`. The companion audit (§5) lists "Playwright/Vitest … + Tauri" for desktop — only Vitest is real today; Playwright VR must be ADDED.

### §5.1 Token-source drift test (anti-bluff, runs in the existing Vitest suite)
- File: `src/styles/__tests__/tokens.test.ts` (ADD).
- Asserts the values in `globals.css`/`tokens.css` equal the canonical `tokens.ts` map (catches palette drift = the single-source guarantee per audit §5 item 1).
- Self-validated per §11.4.107(10): a paired mutation that corrupts one token MUST make the test FAIL.
- Command: `npm run test` (already wired).

### §5.2 Pixel visual-regression (ADD — the §11.4.162 visual-regression requirement)
- Drive the Vite frontend with **Playwright** (`@playwright/test`) headless against `npm run dev` (:1420); Tauri-window VR is operator-attended where headless is infeasible (SKIP-with-reason per §11.4.3).
- Snapshots in **both** themes (toggle `.dark` via the Settings select or by calling `configStore.setTheme`) for key surfaces: Home, Library, MediaDetail, Search, Login, Settings (pages enumerated in `CLAUDE.md` directory table).
- `expect(page).toHaveScreenshot()` baseline → perceptual diff; PASS only within tolerance.
- The diff comparator MUST FAIL a deliberately-degraded golden (anti-bluff §11.4.107(10) / §1.1 paired mutation).
- Commands to ADD:
  - install: `npm i -D @playwright/test && npx playwright install chromium`
  - run: `npx playwright test` (baseline: `npx playwright test --update-snapshots`).
- Cross-cutting (audit §5): light+dark coverage for every snapshot; WCAG 2.1 AA contrast assertions for `foreground`-on-`background` and `onPrimary`-on-`primary` (carry the Android HELIX-175 methodology); no label-overlap assertion.

---

## §6 — Sequencing (PLAN only)
1. Conductor cross-references OpenDesign latest docs/npm (§11.4.99) → decide §3.1 (runtime dep) vs §3.2 (token-module fallback). Fallback is the no-risk default.
2. ADD `tokens.css`/`tokens.ts` from §3.3 (re-verify WCAG AA per audit §5).
3. MODIFY `globals.css` + `tailwind.config.js` + `index.html` (palette, radius, font) per §4.
4. MODIFY `configStore.ts` + `SettingsPage.tsx` to close the `'system'`/`light` gap.
5. ADD token-drift test (§5.1) + Playwright VR scaffold (§5.2), light+dark.
6. Independent code review (§11.4.142) + visual-regression GREEN before any commit (conductor-owned).

---

## Sources verified (in-repo, read this session)
- `catalogizer-desktop/package.json`, `package-lock.json` (playwright transitive-only), `CLAUDE.md`, `index.html`, `tailwind.config.js`, `vitest.config.ts`
- `catalogizer-desktop/src/styles/globals.css`
- `catalogizer-desktop/src/stores/configStore.ts`
- `catalogizer-desktop/src/types/index.ts` (grep: `Theme` at `:225`)
- `catalogizer-desktop/src/pages/SettingsPage.tsx` (grep: `:11,122,236`)
- `docs/design/OPENDESIGN_INTEGRATION_AUDIT.md` (Rev 1)
- `node_modules` absent (not installed) — verified Playwright binary not present.

> No external/online source fetched (subagent). OpenDesign's exact token-API surface (`github.com/nexu-io/open-design`) is **UNCONFIRMED** and MUST be cross-referenced against latest docs per §11.4.99 before choosing §3.1 over the §3.2 fallback.
