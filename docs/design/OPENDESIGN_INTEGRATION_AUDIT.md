# OpenDesign Integration Audit & Unified Token Plan

**Revision:** 1
**Last modified:** 2026-06-25T12:02:41Z
**Authority:** Constitution §11.4.162 (OpenDesign UI design system mandate)
**Scope:** catalogizer-androidtv, catalogizer-android, catalog-web, catalogizer-desktop
**Status:** AUDIT — documentation only. No app source modified. No commit.

---

## Executive summary (FACTS)

- **OpenDesign is NOT integrated anywhere.** A repo-wide search for `nexu-io/open-design`,
  `@open-design`, `opendesign`, `nexu.io`, and `open-design` across every app's
  `package.json` / `build.gradle*` / `*.kt` / `*.ts` / `*.tsx` / `*.css` returned ZERO real
  matches (the only grep hits were unrelated tokens like `CoverQuality` and the `nexus-*`
  HelixQA test banks). Confirmed exit-code-0 with an explicit exclusion filter.
- **Three divergent ad-hoc palettes** exist across the four client apps, each defining its own
  primary color with no shared source of truth:
  - **Android phone + Android TV** — Material 3 scheme in Kotlin; primary **blue** (`#1976D2`
    phone / `#1565C0` TV light, `#9ECAFF` dark).
  - **catalog-web** — Tailwind + shadcn HSL CSS variables; primary **blue** (`#2563EB` light /
    `#3B82F6` dark) PLUS a separate static Tailwind `primary` 50–950 scale also blue (`blue-*`).
  - **catalogizer-desktop** — Tailwind + shadcn HSL CSS variables; primary **near-black slate**
    (`#0F172A` light / `#F8FAFC` dark) — the unmodified shadcn monochrome default, a *different
    brand color entirely* from the other three apps.
- **No brand `colors.xml` exists** in either Android app — colors live only inline in
  `Theme.kt`. No central brand-color asset was found in any app.

The "blue" brand (≈ Material/Tailwind blue) is the dominant signal across 3 of 4 apps; the
desktop app is the outlier. The unified spec below adopts the existing blue brand and converges
the desktop app onto it.

---

## §1 — Current-state audit

### §1.1 Color palettes (actual values, with file references)

| App | Token role | Light (hex) | Dark (hex) | Source file |
|---|---|---|---|---|
| **android (phone)** | primary | `#1976D2` | `#9ECAFF` | `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/theme/Theme.kt:15,44` |
| android | onPrimary | `#FFFFFF` | `#003258` | same `:16,45` |
| android | secondary | `#535F70` | `#BBC7DB` | same `:20,49` |
| android | tertiary | `#6B5778` | `#D6BEE4` | same `:25,54` |
| android | error | `#BA1A1A` | `#FFB4AB` | same `:30,59` |
| android | background/surface | `#FDFCFF` | `#101214` | same `:35-37,64-66` |
| android | onBackground/onSurface | `#1A1C1E` | `#E2E2E6` | same `:36-38,65-67` |
| android | outline | `#73777F` | `#8D9199` | same `:41,70` |
| **androidtv** | primary | `#1565C0` | `#9ECAFF` | `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/theme/Theme.kt:55,17` |
| androidtv | onPrimary | `#FFFFFF` | `#003258` | same `:56,18` |
| androidtv | secondary | `#535F70` | `#BBC7DB` | same `:59,21` |
| androidtv | tertiary | `#6B5778` | `#D6BEE4` | same `:63,25` |
| androidtv | error | `#BA1A1A` | `#FFB4AB` | same `:67,29` |
| androidtv | background/surface | `#FDFCFF` | `#101214` | same `:71-73,33-35` |
| androidtv | onBackground/onSurface | `#1A1C1E` | `#E2E2E6` | same `:72-74,34-36` |
| **catalog-web** | primary (CSS var) | `#2563EB` (`221.2 83.2% 53.3%`) | `#3B82F6` (`217.2 91.2% 59.8%`) | `catalog-web/src/index.css:13,36` |
| catalog-web | background | `#FFFFFF` (`0 0% 100%`) | `#020817` (`222.2 84% 4.9%`) | same `:7,30` |
| catalog-web | foreground | `#020817` | `#F8FAFC` (`210 40% 98%`) | same `:8,31` |
| catalog-web | destructive | `#EF4444` (`0 84.2% 60.2%`) | `0 62.8% 30.6%` | same `:21,44` |
| catalog-web | border/input | `214.3 31.8% 91.4%` | `217.2 32.6% 17.5%` | same `:23-24,46-47` |
| catalog-web | ring | `221.2 83.2% 53.3%` | `224.3 76.3% 94.1%` | same `:25,48` |
| catalog-web | static `primary` scale 50–950 | `#eff6ff` … `#2563eb` (600) … `#172554` | (same scale, theme-independent) | `catalog-web/tailwind.config.js:16-28` |
| catalog-web | `accent` scale (fuchsia) | `#fdf4ff` … `#d946ef` (500) … `#4a044e` | — | `tailwind.config.js:42-54` |
| catalog-web | success/warning/error scales | `#22c55e` / `#f59e0b` / `#ef4444` (500) | — | `tailwind.config.js:55-93` |
| **catalogizer-desktop** | primary (CSS var) | `#0F172A` (`222.2 47.4% 11.2%`) | `#F8FAFC` (`210 40% 98%`) | `catalogizer-desktop/src/styles/globals.css:13,42` |
| catalogizer-desktop | background | `#FFFFFF` (`0 0% 100%`) | `#020817` (`222.2 84% 4.9%`) | same `:7,36` |
| catalogizer-desktop | foreground | `#020817` | `#F8FAFC` | same `:8,37` |
| catalogizer-desktop | destructive | `#EF4444` (`0 84.2% 60.2%`) | `0 62.8% 30.6%` | same `:25,54` |
| catalogizer-desktop | border/input | `214.3 31.8% 91.4%` | `217.2 32.6% 17.5%` | same `:28-29,57-58` |
| catalogizer-desktop | ring | `222.2 84% 4.9%` | `212.7 26.8% 83.9%` | same `:30,59` |

> HSL→hex conversions above were computed via `colorsys` and are exact for the cited HSL triplets.

### §1.2 Typography (actual values)

| App | Font family | Type scale source | Notes |
|---|---|---|---|
| android (phone) | Compose default (`Typography()`) | `Theme.kt:154` | Uses Material 3 default type scale — no custom scale defined. |
| androidtv | `FontFamily.Default` | `Type.kt:11-117` (`TVTypography`) | Custom 10-foot scale: displayLarge 26sp → labelSmall 10sp; weights Normal/Medium. |
| catalog-web | `Inter` (sans), `JetBrains Mono` (mono) | `tailwind.config.js:95-98`; loaded `catalog-web/index.html:11` | Inter weights 300–900, JetBrains Mono 300–700 via Google Fonts. |
| catalogizer-desktop | **none declared** (browser/OS default) | — | No `fontFamily` extend in `tailwind.config.js`; no font in `globals.css` or `index.html`. |

### §1.3 Spacing & radius (actual values)

| App | Spacing extensions | Radius | Source |
|---|---|---|---|
| android / androidtv | Material default 4dp grid (no custom spacing tokens found) | Material default shapes | — |
| catalog-web | `18`=4.5rem, `88`=22rem, `128`=32rem; `borderRadius 4xl`=2rem; `maxWidth 8xl/9xl` | `--radius: 0.75rem` | `tailwind.config.js:135-146`; `index.css:26` |
| catalogizer-desktop | none beyond Tailwind defaults | `--radius: 0.5rem` (lg/md/sm derived) | `tailwind.config.js:40-44`; `globals.css:32` |

### §1.4 Key divergences (the problem)

1. **Primary brand color is inconsistent across 3 distinct values:**
   - Android phone `#1976D2`, Android TV `#1565C0` (intentional WCAG-AA bump per HELIX-175
     comment), web `#2563EB`, desktop `#0F172A` (slate, not blue).
   - **Desktop is the worst divergence** — its primary is monochrome slate, a wholly different
     brand identity from the blue the other 3 share.
2. **Two parallel color systems in catalog-web** — shadcn HSL CSS vars AND a static
   Tailwind `primary` 50–950 scale; their `primary-600` (`#2563eb`) and the CSS-var primary
   (`#2563EB`) happen to align, but they are two independent sources that can drift.
3. **Radius differs** — web `0.75rem` vs desktop `0.5rem`.
4. **Typography unmanaged** — desktop ships no brand font; android phone has no custom scale.
5. **No shared token source** — each app re-declares everything; a brand change requires four
   independent edits with no enforcement.

---

## §2 — OpenDesign gap

| Question | Finding (FACT) |
|---|---|
| Is OpenDesign a dependency in any app? | **No.** Not in any `package.json` (web/desktop) nor any `build.gradle.kts` (android/androidtv). |
| Are OpenDesign design tokens consumed anywhere? | **No.** No `opendesign`/`nexu` import or token reference in any source file. |
| Is there a shared brand-token package? | **No.** The closest shared layer is `@vasic-digital/ui-components` (a `file:../` local dep referenced in catalog-web `CLAUDE.md`), but it is not OpenDesign and was not found to define brand color tokens. |
| Light + dark variants present today? | Partially — all four apps have light+dark, but with **divergent** palettes, not a single token source. |
| Visual-regression tests for theming? | None found wired to a design-token source. |

**Conclusion:** §11.4.162 is currently UNMET across all four client apps. OpenDesign must be
introduced as the single token source, the divergent palettes reconciled onto one brand, and
light+dark variants + visual-regression coverage added.

---

## §3 — Unified design-token spec (proposed, light + dark)

Brand decision (derived from existing assets, not invented): adopt the **Catalogizer Blue** that
3 of 4 apps already use. Anchor the primary on the web/Material blue family and converge the
desktop app off its monochrome slate default. Neutrals follow the shadcn slate ramp already used
by web + desktop (and close to the Android M3 neutrals).

### §3.1 Core color tokens

| Token (OpenDesign `color.*`) | Light hex | Dark hex | Rationale / origin |
|---|---|---|---|
| `color.brand.primary` | `#1565C0` | `#9ECAFF` | Android TV light (AA-bumped per HELIX-175) + Android dark; the most accessibility-vetted blue in the repo. |
| `color.brand.primary.hover` | `#1976D2` | `#7FB4F5` | Android phone primary as the lighter hover step. |
| `color.brand.onPrimary` | `#FFFFFF` | `#003258` | From existing Android M3 onPrimary pair. |
| `color.brand.secondary` | `#535F70` | `#BBC7DB` | Shared Android secondary (already identical phone+TV). |
| `color.brand.onSecondary` | `#FFFFFF` | `#253140` | Android M3 pair. |
| `color.brand.accent` | `#6B5778` | `#D6BEE4` | Android tertiary (the only existing accent shared by mobile); replaces web's stray fuchsia. |
| `color.semantic.success` | `#16A34A` | `#22C55E` | web `success-600/500` (`tailwind.config.js:62,61`). |
| `color.semantic.warning` | `#D97706` | `#F59E0B` | web `warning-600/500` (`:74,73`). |
| `color.semantic.error` | `#DC2626` | `#EF4444` | web `error-600`/shadcn destructive (`:88` / `index.css:21`). |
| `color.neutral.background` | `#FFFFFF` | `#020817` | shadcn `--background` shared by web+desktop. |
| `color.neutral.surface` | `#F8FAFC` | `#0F172A` | slate-50 / slate-900 (card/elevated). |
| `color.neutral.foreground` | `#020817` | `#F8FAFC` | shadcn `--foreground` shared by web+desktop. |
| `color.neutral.muted` | `#64748B` | `#94A3B8` | slate-500/400 (`muted-foreground`). |
| `color.neutral.border` | `#E2E8F0` | `#1E293B` | shadcn border (`214.3 31.8% 91.4%` / `217.2 32.6% 17.5%`). |
| `color.state.ring` | `#1565C0` | `#9ECAFF` | focus ring = brand primary (unifies web/desktop divergent ring values). |

> The dark/light pairs above are the *target*; exact perceptual contrast must be re-verified
> against WCAG 2.1 AA during implementation (see §5), as the Android comments already do.

### §3.2 Typography tokens

| Token (`typography.*`) | Value | Origin |
|---|---|---|
| `font.family.sans` | `Inter, system-ui, sans-serif` | catalog-web `tailwind.config.js:96`. |
| `font.family.mono` | `JetBrains Mono, monospace` | catalog-web `:97`. |
| `font.scale.displayLarge … labelSmall` | adopt androidtv `TVTypography` ratios for TV; web/desktop use the M3-equivalent rem scale | `Type.kt:11-117`. |
| Desktop action | **add** `Inter` (currently has no brand font) | gap from §1.2. |

### §3.3 Spacing & radius tokens

| Token | Value | Origin |
|---|---|---|
| `space.*` | 4px base grid; named steps `18`=4.5rem, `88`=22rem, `128`=32rem | catalog-web `tailwind.config.js:135-139` (superset). |
| `radius.base` | `0.75rem` (`lg`); `md`=calc-2px; `sm`=calc-4px | unify on web's `0.75rem`; desktop raises from `0.5rem`. |
| `radius.4xl` | `2rem` | catalog-web `:141`. |

### §3.4 Component tokens (minimum set)

Map the above to OpenDesign component-level tokens: `button.primary.{bg,fg,hover,ring}`,
`card.{bg,fg,border,radius}`, `input.{bg,border,ring}`, `surface.elevated.bg`,
`focus.ring.{color,width,offset}` — each resolving to the §3.1–§3.3 primitives so a single
brand change cascades to all components and all four apps.

---

## §4 — Per-app integration steps

> All steps are PLAN only. Sequence: introduce OpenDesign as the token source → generate
> platform-specific outputs → replace each app's ad-hoc tokens with generated ones.

### §4.0 Foundation (shared)
1. Add OpenDesign (`github.com/nexu-io/open-design`) per §11.4.74 (extend-don't-reimplement;
   PR upstream any missing pattern rather than fork). For JS apps install as an npm dependency;
   for Android consume generated outputs (see §4.3).
2. Author the canonical token set from §3 as OpenDesign source-of-truth tokens (light+dark).
3. Generate per-platform artifacts: CSS custom properties (web+desktop) and a Kotlin
   `Color`/`Typography` token file (android+androidtv).

### §4.1 catalog-web
- Replace the hand-written `:root`/`.dark` HSL block in `src/index.css:6-49` with
  OpenDesign-generated CSS variables.
- Collapse the duplicate static `primary` 50–950 scale in `tailwind.config.js:16-93` onto the
  generated brand tokens (remove the drift risk; keep semantic success/warning/error mapped to
  §3.1).
- Re-point `accent` (currently fuchsia `#d946ef`) to `color.brand.accent` (`#6B5778`) for brand
  consistency, unless product explicitly wants the fuchsia accent (operator decision per
  §11.4.66 if so).
- Keep Inter / JetBrains Mono (already aligned with §3.2).

### §4.2 catalogizer-desktop
- **Highest-impact change:** replace the monochrome slate `--primary` (`#0F172A`/`#F8FAFC`,
  `globals.css:13,42`) with the brand blue tokens from §3.1 so desktop matches the other apps.
- Replace the full `:root`/`.dark` block (`globals.css:6-60`) with OpenDesign-generated vars.
- Raise `--radius` `0.5rem → 0.75rem` (`:32`) to match the unified `radius.base`.
- Add `font.family.sans = Inter` (currently absent): extend `tailwind.config.js` `fontFamily`
  and load the font (Google Fonts link or bundled).

### §4.3 catalogizer-android (phone)
- Replace the inline `Light*`/`Dark*` `Color(...)` constants in `Theme.kt:15-70` with the
  OpenDesign-generated Kotlin token file.
- Align light `primary` `#1976D2 → #1565C0` (the AA-vetted brand primary) — note the existing
  HELIX-175 contrast rationale lives in the TV app; carry that comment forward.
- Decide on `dynamicColor` (`Theme.kt:131,135-138`): Android-12+ dynamic color currently
  OVERRIDES the brand palette. For brand consistency, gate dynamic color behind a user setting
  defaulting to the brand scheme (operator decision per §11.4.66).
- Add a custom `Typography` from `font.scale` (phone currently uses bare `Typography()`).

### §4.4 catalogizer-androidtv
- Replace inline `TVDarkColorScheme`/`TVLightColorScheme` in `Theme.kt:16-77` with generated
  tokens; the TV light primary `#1565C0` already equals the chosen brand primary (no change to
  primary value — this app is the brand anchor).
- Keep the 10-foot `TVTypography` scale (`Type.kt`) but source its family/weights from
  `typography.*` tokens.

---

## §5 — Visual-regression test plan (per §11.4.162)

§11.4.162 requires every UI change be covered by standard test types **including visual
regression** (before/after screenshots with per-pixel or perceptual diff PASS/FAIL), every
component ship **light + dark** variants, and **no element/label overlap**.

| App | Tooling (already present) | VR approach |
|---|---|---|
| catalog-web | Playwright (`playwright.config.ts`, `e2e/`) | Add `expect(page).toHaveScreenshot()` snapshots for key surfaces (home, library, media detail, login, settings) in BOTH `light` and `dark` (toggle `.dark` class). Baseline → diff on every change; PASS only on perceptual match within tolerance. |
| catalogizer-desktop | Playwright/Vitest (`vitest.config.ts`) + Tauri | Component-level VR via Playwright against the Vite frontend for the same surfaces, light+dark. Tauri window VR optional/operator-attended where headless infeasible (SKIP-with-reason per §11.4.3). |
| catalogizer-android | Compose UI test (`androidTest/`, Espresso+Compose) | Compose screenshot tests (Roborazzi/Paparazzi-class) for each screen in light+dark theme; assert against committed golden PNGs with a self-validated diff analyzer (golden-good/golden-bad pair per §11.4.107(10)). |
| catalogizer-androidtv | Compose-for-TV UI test (`androidTest/`) | Same Compose screenshot approach, focused on 10-foot surfaces + D-pad focus-state rendering, light+dark. |

Cross-cutting test requirements:
1. **Token-source test** — a unit test asserting each app's generated tokens equal the
   OpenDesign canonical set (catches drift; the single-source guarantee).
2. **Light + dark coverage** — every VR snapshot exists in both themes.
3. **Overlap/contrast assertions** — automated check that no label overlays another and that
   `foreground`-on-`background` / `onPrimary`-on-`primary` meet WCAG 2.1 AA (carry the Android
   HELIX-175 contrast methodology to all apps).
4. **Self-validated analyzer** — the perceptual-diff comparator must FAIL a deliberately
   degraded golden (anti-bluff per §11.4.107(10) / §1.1 paired mutation).

---

## Sources verified (in-repo, 2026-06-25)

- `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/theme/Theme.kt`
- `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/theme/Type.kt`
- `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/theme/Theme.kt`
- `catalog-web/tailwind.config.js`, `catalog-web/src/index.css`, `catalog-web/index.html`
- `catalogizer-desktop/tailwind.config.js`, `catalogizer-desktop/src/styles/globals.css`,
  `catalogizer-desktop/src/main.tsx`
- Repo-wide OpenDesign grep (exit 0, zero real matches).

> No external/online source was fetched for this audit (in-repo facts only). OpenDesign's exact
> token-API surface (`github.com/nexu-io/open-design`) MUST be cross-referenced against its
> latest docs per §11.4.99 before §4.0 implementation begins.
