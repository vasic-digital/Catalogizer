# [MAJOR] Web Dashboard: "Something went wrong" error after login

**Platform**: Web (Playwright + manual browser)
**Screen**: Dashboard (post-login)
**Severity**: MAJOR
**Discovered by**: HelixQA autonomous web session (curiosity step #1)
**Session**: session-1775320720

## Description

After successful login (green "Successfully logged in!" toast visible), the Dashboard page shows an error: "Something went wrong in Dashboard. This page encountered an error. Other pages should still work normally." with a "Reload Page" button.

## Evidence

- **Screenshot**: `web-curiosity-001-after.png` (shows error + success toast simultaneously)
- **Session log**: `helixqa-web-prelogin-v3.log` — curiosity step #1 after login
- Navigation bar fully rendered (Media, Browse, Favorites, Playlists, Analytics, Subtitles, Collections, Convert, Admin)

## Reproduction Steps

1. Open http://localhost:3000
2. Login with admin/admin123
3. Dashboard loads with error

## Expected Behavior

Dashboard should load with media statistics, recent activity, and quick access cards.

## Actual Behavior

Error boundary catches an exception: "Something went wrong in Dashboard"

## Root Cause (suspected)

The Dashboard component may throw when API endpoints return unexpected data format, or a required API endpoint is missing/returning an error. The fresh database with NAS-scanned data may have a different schema than expected by the Dashboard's data fetching logic.
