# OpenDesign Consumability Verdict

**Revision:** 1
**Last modified:** 2026-06-25T00:00:00Z
**Authority:** §11.4.99 latest-source verification · §11.4.6 no-guessing — every claim cited to a fetched URL
**Scope:** Resolves the shared `UNCONFIRMED` blocker in the OpenDesign (§11.4.162) integration plans for the catalogizer web, Tauri desktop, and androidtv clients.

---

## VERDICT

**NOT-CLEANLY-CONSUMABLE → token-module fallback.**

`nexu-io/open-design` is a **local-first desktop application + CLI + MCP server**, NOT a publishable design-token / design-system code dependency. It publishes **no npm package that exports design tokens** for programmatic import by any application stack. Its "design systems" are `DESIGN.md` **markdown prose files** authored to be fed into an AI coding agent's prompt — they are not an importable token artifact.

- **npm token package for direct import: NONE.** (See evidence below — the only nexu-io-adjacent npm packages are an MCP bridge and markdown-template installers, neither of which is a consumable token library.)
- **Hand-authored design-token module fallback: CONFIRMED CORRECT for all three stacks** (web/React+Tailwind+CSS-custom-properties, Tauri desktop, Android/Jetpack Compose).

---

## What `nexu-io/open-design` actually is

A **native desktop app** (Electron shell; Next.js 16 / React 18 / TypeScript frontend; Node 24 / Express / SQLite daemon) — self-described as a "Local-first, open-source Claude Design alternative." It generates design artifacts (web/desktop/mobile prototypes, slides, images, videos, HyperFrames) and exports HTML/PDF/PPTX/MP4/ZIP/Markdown.

Its distribution model is explicitly **"skills, a CLI, and an MCP server — not an npm package"** (consumed as a native app, or wired into coding agents via `od mcp install <agent>` over the Model Context Protocol). Each "design system" is a portable `DESIGN.md` file (a markdown schema following the `awesome-claude-design` / VoltAgent `awesome-design-md` 9-section format) read as part of an agent's system prompt. Some systems additionally ship a `tokens.css` ("canonical compiled CSS custom properties") — but this is a per-system file inside the repo / template install, NOT a versioned, importable npm token library.

## npm reality (the blocker, resolved)

Registry search for the nexu-io/open-design design-token package returns **none**. The closest npm artifacts are:

| npm package | What it is | Consumable as a token dependency? |
|---|---|---|
| `open-design-mcp` (v0.16.1, publisher `nano-step001`) | MCP stdio server bridging coding agents to the Open Design daemon | No — it is agent tooling, not a token export |
| `open-design-templates` (v1.1.0, publisher `deceivedleaped`) | CLI (`bin: od-templates`) that **copies `DESIGN.md` / `SKILL.md` markdown** into Open Design's directories | No — installs markdown, not JS/JSON/CSS tokens; "not a React, Tailwind, Android, or Compose dependency" |
| `getdesign` (v0.6.24, repo `VoltAgent/awesome-design-md`, **not** nexu-io) | CLI (`bin: getdesign`) — `npx getdesign add <slug>` installs pre-built `DESIGN.md` markdown templates | No — distributes markdown templates, not token objects |

There is no `@nexu-io/open-design`, no `@nexu/open-design`, no nexu-io-published design-token library. The `od` command referenced in the README is the desktop app's CLI, not an installable token package.

## Per-stack consumption mechanism

### (a) React + Tailwind + CSS-custom-properties (catalogizer web & Tauri desktop)
**No clean dependency path.** There is no npm package to `npm install` that yields importable tokens. The only OpenDesign-native artifact is `DESIGN.md` (markdown prose) plus, per-system, a `tokens.css`. A `tokens.css` could in principle be hand-copied into the web/desktop app and referenced as CSS custom properties + a Tailwind `theme.extend` mapping — but that is a manual copy of one repo file, not a maintained, versioned dependency, and it does not carry the project's "Catalogizer Blue" palette (those tokens would still have to be authored). **Concrete mechanism: author the tokens locally** (see fallback).

### (b) Jetpack Compose / Android (catalogizer androidtv)
**No consumption mechanism exists.** OpenDesign produces no Kotlin/Compose artifact, no Android resource set, and no Gradle-publishable token library. `tokens.css` is not consumable by Compose. There is no path from `nexu-io/open-design` into an Android build. **Compose tokens must be hand-authored.**

## Confirmed fallback (the plans' call is correct)

For **all three stacks**, the correct approach is a **hand-authored design-token module that mirrors the OpenDesign `DESIGN.md` / `tokens.css` token structure** (the 9-section schema: visual theme, color palette light+dark, typography, spacing, etc.), encoding the project's **"Catalogizer Blue" light + dark palette**:

- **Web / Tauri desktop:** a single source-of-truth token file (e.g. `tokens.css` CSS custom properties for light + dark `:root` / `[data-theme="dark"]`), consumed by Tailwind via `theme.extend` referencing `var(--…)`. Mirror OpenDesign's section structure so a future OpenDesign export can be diffed against it.
- **Android TV / Compose:** a Kotlin token object / `MaterialTheme` `ColorScheme` (light + dark) mirroring the same token names/values.

This satisfies §11.4.162's intent (use a design-token/theme system with light+dark variants from brand assets, no ad-hoc one-off CSS) while honestly reflecting that OpenDesign is not consumable as a dependency. The §11.4.74 "extend upstream, don't reimplement" obligation does not bite here: there is no upstream token *library* to extend — the upstream is an app + markdown, so authoring a project-local token module mirroring its schema is the sanctioned path, not a duplication of a reusable package.

---

## Sources verified 2026-06-25

- https://github.com/nexu-io/open-design — repo README, tech stack, "skills/CLI/MCP — not an npm package" distribution model (fetched 2026-06-25)
- https://github.com/nexu-io (org), https://open-design.ai/ — project identity: local-first Claude Design alternative, Apache-2.0, BYOK (via web search, 2026-06-25)
- https://github.com/nexu-io/open-design/blob/main/design-systems/README.md — `DESIGN.md` = "canonical design prose for agents", `tokens.css` = "canonical compiled CSS custom properties" (fetched 2026-06-25)
- https://github.com/nexu-io/open-design/blob/main/docs/spec.md — design systems follow the external `awesome-claude-design` 9-section schema; the 9 sections are defined in that external repo, not inline (fetched 2026-06-25)
- https://registry.npmjs.org/getdesign — `getdesign` v0.6.24, "CLI for installing DESIGN.md templates", repo `VoltAgent/awesome-design-md`, `bin: getdesign`, NOT nexu-io (fetched 2026-06-25)
- https://registry.npmjs.org/-/v1/search?text=nexu-io%20open-design — full npm search; no nexu-io design-token library; surfaces `open-design-mcp` (MCP bridge) + `open-design-templates` (template installer) (fetched 2026-06-25)
- https://registry.npmjs.org/open-design-templates — `open-design-templates` v1.1.0, installs `DESIGN.md`/`SKILL.md` markdown, `bin: od-templates`, "not a React, Tailwind, Android, or Compose dependency" (fetched 2026-06-25)
- https://www.npmjs.com/package/getdesign, https://www.npmjs.com/search?q=open-design — npmjs.com returned HTTP 403 to the fetcher; the authoritative `registry.npmjs.org` API (above) was used instead (attempted 2026-06-25)
