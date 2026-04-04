# [CRITICAL] HelixQA Playwright cannot authenticate — blocks ALL web QA

**Platform**: Web (Playwright headless)
**Severity**: CRITICAL (blocks entire web platform QA)
**Sessions affected**: All web autonomous sessions (3 attempts today)

## Description

The HelixQA Playwright executor cannot interact with the Catalogizer React login form. All web QA sessions complete technically (30+ tests) but ALL screenshots show only the login page. The LLM correctly identifies the login form, types credentials, and submits — but the React form does not respond to Playwright's input actions.

## Evidence

- **Video**: N/A (web sessions use Playwright, no ADB screenrecord)
- **Screenshots**: `web-001-loginform.png` through `web-118-*.png` — ALL show login page
- **Session logs**: `helixqa-full-web.log` — 40+ curiosity steps, all "screen unchanged"
- **Session ID**: session-1775317470

## Root Cause Analysis

The Playwright executor sends actions (click, type, key) but the React app's form elements do not respond. Possible causes:

1. **Coordinate mismatch**: Playwright clicks at LLM-suggested coordinates, but the React app renders differently in headless mode (no GPU rendering)
2. **Missing element wait**: The form may not be fully interactive when Playwright starts clicking (React hydration timing)
3. **CSP/CORS blocking**: The headless browser may block certain JavaScript or API requests needed for form interactivity
4. **Playwright version incompatibility**: The bundled Chromium may not support the React app's rendering

## Impact

- 0% of web UI features tested autonomously
- Login, dashboard, media browser, search, collections, playlists, settings — ALL untested
- Web platform is effectively excluded from autonomous QA

## Required Fix

Investigate and fix the HelixQA Playwright executor:
1. Add explicit `waitForSelector` before interacting with form elements
2. Use Playwright's built-in `fill()` and `click()` with element selectors instead of coordinate-based clicks
3. Add a login pre-step that authenticates via API and injects the session cookie into the browser
4. Alternatively, implement a Playwright-specific login flow that uses the DOM directly
