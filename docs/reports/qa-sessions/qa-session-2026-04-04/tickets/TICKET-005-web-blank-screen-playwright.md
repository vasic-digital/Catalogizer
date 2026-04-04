# [CRITICAL] Web Platform: Playwright sees blank white page

**Platform**: Web (Playwright headless)
**Screen**: All pages
**Severity**: CRITICAL (blocks all web autonomous QA)
**Discovered by**: HelixQA autonomous web session

## Description

The Playwright browser launched by HelixQA captures completely white/blank screenshots for every page. All 39 execute-phase tests and all curiosity-phase steps show only white pages. The web app at `http://localhost:3000` loads normally in a regular browser.

## Evidence

- Screenshots: `web-001-layout.png` through `web-049-*.png` — all blank white
- The LLM correctly identified this as "completely blank" and tried multiple recovery actions (scroll, escape, click, wait, key)

## Root Cause (suspected)

1. HelixQA's Playwright adapter may not be navigating to the correct URL
2. The Vite dev server may require specific headers or cookies that headless Playwright doesn't provide
3. The React app may have JavaScript errors in headless mode that prevent rendering
4. The Playwright version may be incompatible with the Chromium bundled version

## Impact

All web platform autonomous QA is blocked. No screens, features, or UI/UX can be tested via HelixQA web platform.

## Fix Required

1. Investigate HelixQA's Playwright adapter URL configuration
2. Verify Playwright can reach `http://localhost:3000` in the current environment
3. Check browser console for JavaScript errors in headless mode
4. Ensure the Playwright adapter is using a compatible Chromium version
