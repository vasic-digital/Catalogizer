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


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** use `su` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Container-Based Solutions
When a build or runtime environment requires system-level dependencies, use containers instead of elevation:

- **Use the `Containers` submodule** (`https://github.com/vasic-digital/Containers`) for containerized build and runtime environments
- **Add the `Containers` submodule as a Git dependency** and configure it for local use within the project
- **Build and run inside containers** to avoid any need for privilege escalation
- **Rootless Podman/Docker** is the preferred container runtime

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo` or `su`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Use the `Containers` submodule for containerized builds
5. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**


