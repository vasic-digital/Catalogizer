# CLAUDE.md - Catalogizer Website

## Overview

Project website for Catalogizer, built with VitePress. Contains product documentation, user guides, developer reference, FAQ, download links, and video course outline. Static site generated from Markdown.

**Package**: `catalogizer-website` (VitePress 1.x / Markdown)

## Build & Test

```bash
npm install
npm run dev          # vitepress dev (live reload)
npm run build        # vitepress build (static output to .vitepress/dist)
npm run preview      # serve built site locally
```

## Content Structure

| Path | Purpose |
|------|---------|
| `index.md` | Landing page |
| `features.md` | Product features overview |
| `download.md` | Platform download links and instructions |
| `getting-started.md` | Quick start guide |
| `documentation.md` | Documentation hub |
| `faq.md` | Frequently asked questions |
| `support.md` | Support channels and resources |
| `changelog.md` | Release changelog |
| `course.md` | Video course outline |
| `guides/` | User guides: web-app, desktop, android, android-tv, configuration, security, monitoring |
| `developer/` | Developer docs: architecture, api, testing, contributing |
| `docs/` | Additional docs: getting-started, testing-strategy |
| `.vitepress/config.ts` | VitePress config: nav, sidebar, social links, footer |

## Writing Conventions

- Markdown files at root level for top-level pages
- Subdirectories (`guides/`, `developer/`) for grouped content
- Front matter not required (VitePress infers title from first `#` heading)
- Internal links use relative paths without `.md` extension (e.g., `/features`, `/guides/web-app`)
- `ignoreDeadLinks: true` is set in config (links to unimplemented pages are tolerated)

## Site Configuration

VitePress config in `.vitepress/config.ts` defines:
- **Nav**: Home, Features, Download, Documentation, Video Course, FAQ, Support
- **Sidebar**: 5 groups (Getting Started, User Guide, Administration, Developer, Resources)
- **Footer**: MIT License, Vasic Digital copyright
- **Social**: GitHub link

## Dependencies

- **Dev**: `vitepress ^1.5.0`

## Commit Style

Conventional Commits: `docs(website): add Android TV user guide`
