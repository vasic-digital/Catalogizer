# Splash Screen & Branding Design Specification

> **Status:** APPROVED
> **Date:** 2026-03-26

---

## Goal

Add enterprise-quality branded splash screens to all 5 Catalogizer applications (Web, Desktop, Installer Wizard, Android, Android TV) with the Vasic Digital footer, app version display, and minimum display durations. Validate with comprehensive testing across all platforms including 4 physical ADB devices.

## Visual Design

### Layout (all platforms)

```
┌─────────────────────────────────┐
│                                 │
│         [App Icon]              │
│    Application_Icon.jpeg        │
│         (centered)              │
│                                 │
│       "Catalogizer"             │
│  Advanced Multi-Protocol Media  │
│  Collection Management System   │
│                                 │
│     [loading indicator]         │
│                                 │
│   ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│   [VD logo]  Made with ♥ by    │
│              Vasic Digital      │
│              v1.1.0             │
└─────────────────────────────────┘
```

- **Background**: Dark gradient `#0F172A` to `#1E293B` (Tailwind slate-900 to slate-800)
- **App icon**: `Application_Icon.jpeg` centered, 120px (web/desktop) or 96dp (Android)
- **Title**: "Catalogizer" — white, bold, 32px/24sp
- **Subtitle**: "Advanced Multi-Protocol Media Collection Management System" — `#94A3B8` (slate-400), 14px/14sp
- **Loading indicator**: Subtle pulsing dot or spinning ring, `#4A90E2` (brand blue)
- **Footer**: `Vasic_Digital_Logo.jpeg` (32px height) + "Made with ♥ by Vasic Digital" in `#64748B` (slate-500)
- **Version**: `v1.1.0` below footer in `#475569` (slate-600), 12px/12sp
- **Installer Wizard title**: "Catalogizer Installation Wizard" instead of just "Catalogizer"

### Duration Rules

| Condition | Minimum Duration |
|-----------|-----------------|
| First launch (no prior session detected) | 5.0 seconds |
| Regular launch | 2.5 seconds |

Detection mechanism:
- **Web**: `localStorage.getItem('catalogizer_launched')`
- **Desktop/Wizard**: Tauri `Store` plugin or filesystem check
- **Android/TV**: `SharedPreferences.getBoolean("has_launched", false)`

Splash hides only when BOTH conditions are true: minimum time elapsed AND app initialization complete.

### Version Footer on Landing Screens

After splash dismisses, each app's main landing screen shows version:
- **Web**: Bottom-right corner, small muted text
- **Desktop**: Bottom-left of sidebar
- **Wizard**: Bottom-center of wizard container
- **Android**: Bottom-center of home screen
- **Android TV**: Bottom-right of home screen

## Assets

| Asset | Source | Usage |
|-------|--------|-------|
| `Application_Icon.jpeg` | `Assets/Logo/` | Splash center icon |
| `Vasic_Digital_Logo.jpeg` | `Assets/Logo/` (downloaded from vasic-digital GitHub) | Footer branding |
| `Application_Logo.svg` | `Assets/Logo/` | Loading indicator accent |

Assets must be copied/imported into each app's asset directory in appropriate formats and resolutions.

## Per-Platform Implementation

### 1. catalog-web (React)
- New `src/components/SplashScreen.tsx`
- Wraps entire app in `App.tsx`, shows before Suspense content
- CSS animation for loading indicator
- Copy logo assets to `src/assets/`

### 2. catalogizer-desktop (Tauri/React)
- Replace existing `LoadingScreen.tsx` with branded `SplashScreen.tsx`
- Same React component pattern as web
- Copy logo assets to `src/assets/`

### 3. installer-wizard (Tauri/React)
- New `src/components/SplashScreen.tsx`
- Shows before WizardProvider/ConfigurationProvider mount
- Title: "Catalogizer Installation Wizard"
- Copy logo assets to `src/assets/`

### 4. catalogizer-android (Kotlin/Compose)
- Extend existing SplashScreen API usage with longer duration
- Add custom post-splash Composable screen with branded layout
- Copy logo assets to `res/drawable/`
- SharedPreferences for first-launch detection

### 5. catalogizer-androidtv (Kotlin/Compose)
- Add SplashScreen API (currently missing)
- Custom Composable splash screen with TV-optimized layout (larger text, D-pad friendly)
- Copy logo assets to `res/drawable/`
- SharedPreferences for first-launch detection

## Version Bump

All apps bump from `1.0.0` to `1.1.0`, versionCode incremented:
- `catalog-web/package.json`
- `catalogizer-desktop/package.json` + `tauri.conf.json`
- `installer-wizard/package.json` + `tauri.conf.json`
- `catalogizer-android/app/build.gradle.kts`
- `catalogizer-androidtv/app/build.gradle.kts`

## Testing Plan

### Unit/Component Tests
- Splash renders with correct elements (icon, title, footer, version)
- First-launch detection works
- Minimum duration enforced (5s first, 2.5s regular)
- Splash dismisses after init completes

### E2E Tests (Playwright)
- Web splash visible on load, disappears after init
- Screenshot capture of splash screen
- Video recording of full startup sequence

### Android Tests (ADB)
- 4 devices in parallel (2 phones Android 15, 2 TV devices)
- `adb shell screencap` during splash
- `adb shell screenrecord` for video
- Verify splash duration via logcat timestamps

### HelixQA Bank Entries
- New test cases for splash screen verification per platform
- Autonomous LLM-driven curiosity testing with video analysis

### Challenges
- CH-099: Web splash screen renders with all required elements
- CH-100: Desktop splash screen renders with all required elements
- CH-101: Wizard splash screen renders with all required elements
- CH-102: Android splash screen meets duration requirements
- CH-103: Android TV splash screen meets duration requirements

## Deliverables

1. Splash screens in all 5 apps
2. Version footer on all landing screens
3. Version bump to 1.1.0 across all apps
4. All tests passing (unit, E2E, challenges)
5. Video recordings in `qa-results/`
6. Release APKs in `releases/`
7. Final report with all findings and fix log
