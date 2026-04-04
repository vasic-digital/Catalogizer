# [MAJOR] Web QA: HelixQA unable to navigate past login page

**Platform**: Web (Playwright)
**Screen**: Login page
**Severity**: MAJOR (blocks post-login web QA)
**Discovered by**: HelixQA autonomous web session

## Description

All 15 web screenshots show the login page. The HelixQA Playwright executor did not successfully authenticate and navigate to the dashboard or any post-login screens. The session also crashed/died at test 15/33 without completing.

## Evidence

- Screenshots `web-001` through `web-015`: All show login page
- Session died silently after test 15 (no error logged)

## Root Cause (suspected)

1. The Playwright executor's login flow may not interact with the form fields correctly
2. The React app may require specific timing between field fill and button click
3. Playwright may not wait for navigation after login submission

## Fix Required

1. Investigate HelixQA Playwright executor login flow
2. Add explicit wait-for-navigation after login form submission
3. Add error handling for Playwright process crashes
