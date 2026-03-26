# Splash Screens & Branding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add enterprise-quality branded splash screens with Vasic Digital footer to all 5 Catalogizer applications, bump versions to 1.1.0, validate with comprehensive testing on 4 ADB devices, and produce release builds.

**Architecture:** Each app gets a `SplashScreen` component/composable that shows the branded loading screen with minimum duration enforcement (5s first launch, 2.5s regular). React apps share a common design pattern; Android/TV use Compose with SplashScreen API. Asset images are copied into each app's assets directory. Testing uses Playwright (web/desktop), ADB screenshots (Android/TV), and HelixQA autonomous sessions.

**Tech Stack:** React 18, TypeScript, Tailwind CSS, Tauri 2, Kotlin, Jetpack Compose, SplashScreen API, Playwright, ADB, HelixQA, k6

---

## File Structure

### New Files
| File | Responsibility |
|------|---------------|
| `catalog-web/src/components/SplashScreen.tsx` | Web app splash screen component |
| `catalog-web/src/components/__tests__/SplashScreen.test.tsx` | Web splash tests |
| `catalog-web/src/assets/app-icon.jpeg` | Copy of Application_Icon.jpeg |
| `catalog-web/src/assets/vasic-digital-logo.jpeg` | Copy of Vasic_Digital_Logo.jpeg |
| `catalogizer-desktop/src/components/SplashScreen.tsx` | Desktop splash (replaces LoadingScreen) |
| `catalogizer-desktop/src/components/__tests__/SplashScreen.test.tsx` | Desktop splash tests |
| `catalogizer-desktop/src/assets/app-icon.jpeg` | Copy of Application_Icon.jpeg |
| `catalogizer-desktop/src/assets/vasic-digital-logo.jpeg` | Copy of Vasic_Digital_Logo.jpeg |
| `installer-wizard/src/components/SplashScreen.tsx` | Wizard splash screen |
| `installer-wizard/src/components/__tests__/SplashScreen.test.tsx` | Wizard splash tests |
| `installer-wizard/src/assets/app-icon.jpeg` | Copy of Application_Icon.jpeg |
| `installer-wizard/src/assets/vasic-digital-logo.jpeg` | Copy of Vasic_Digital_Logo.jpeg |
| `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/splash/SplashContent.kt` | Android branded splash composable |
| `catalogizer-android/app/src/main/res/drawable/vasic_digital_logo.jpeg` | VD logo for Android |
| `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/splash/SplashContent.kt` | TV branded splash composable |
| `catalogizer-androidtv/app/src/main/res/drawable/vasic_digital_logo.jpeg` | VD logo for TV |

### Modified Files
| File | Changes |
|------|---------|
| `catalog-web/src/App.tsx` | Wrap with SplashScreen |
| `catalog-web/src/components/layout/Layout.tsx` | Add version footer |
| `catalog-web/package.json` | Version 1.0.0 → 1.1.0 |
| `catalogizer-desktop/src/App.tsx` | Replace LoadingScreen with SplashScreen |
| `catalogizer-desktop/src/components/Layout.tsx` | Add version footer |
| `catalogizer-desktop/package.json` | Version 1.0.0 → 1.1.0 |
| `catalogizer-desktop/src-tauri/tauri.conf.json` | Version 1.0.0 → 1.1.0 |
| `installer-wizard/src/App.tsx` | Wrap with SplashScreen |
| `installer-wizard/src/components/layout/WizardLayout.tsx` | Add version footer |
| `installer-wizard/package.json` | Version 1.0.0 → 1.1.0 |
| `installer-wizard/src-tauri/tauri.conf.json` | Version 1.0.0 → 1.1.0 |
| `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/MainActivity.kt` | Add branded splash composable |
| `catalogizer-android/app/src/main/res/values/styles.xml` | Extend splash duration |
| `catalogizer-android/app/build.gradle.kts` | versionCode + versionName 1.1.0 |
| `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/MainActivity.kt` | Add SplashScreen API + branded composable |
| `catalogizer-androidtv/app/src/main/res/values/styles.xml` | Add splash theme |
| `catalogizer-androidtv/app/build.gradle.kts` | versionCode + versionName 1.1.0 |

---

## Task 1: Copy Brand Assets to All Apps

**Files:**
- Create: `catalog-web/src/assets/app-icon.jpeg` (copy from `Assets/Logo/Application_Icon.jpeg`)
- Create: `catalog-web/src/assets/vasic-digital-logo.jpeg` (copy from `Assets/Logo/Vasic_Digital_Logo.jpeg`)
- Create: `catalogizer-desktop/src/assets/app-icon.jpeg`
- Create: `catalogizer-desktop/src/assets/vasic-digital-logo.jpeg`
- Create: `installer-wizard/src/assets/app-icon.jpeg`
- Create: `installer-wizard/src/assets/vasic-digital-logo.jpeg`
- Create: `catalogizer-android/app/src/main/res/drawable/vasic_digital_logo.jpeg`
- Create: `catalogizer-android/app/src/main/res/drawable/app_icon.jpeg`
- Create: `catalogizer-androidtv/app/src/main/res/drawable/vasic_digital_logo.jpeg`
- Create: `catalogizer-androidtv/app/src/main/res/drawable/app_icon.jpeg`

- [ ] **Step 1: Copy assets to all app directories**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer

# Web
cp Assets/Logo/Application_Icon.jpeg catalog-web/src/assets/app-icon.jpeg
cp Assets/Logo/Vasic_Digital_Logo.jpeg catalog-web/src/assets/vasic-digital-logo.jpeg

# Desktop
mkdir -p catalogizer-desktop/src/assets
cp Assets/Logo/Application_Icon.jpeg catalogizer-desktop/src/assets/app-icon.jpeg
cp Assets/Logo/Vasic_Digital_Logo.jpeg catalogizer-desktop/src/assets/vasic-digital-logo.jpeg

# Wizard
cp Assets/Logo/Application_Icon.jpeg installer-wizard/src/assets/app-icon.jpeg
cp Assets/Logo/Vasic_Digital_Logo.jpeg installer-wizard/src/assets/vasic-digital-logo.jpeg

# Android
cp Assets/Logo/Application_Icon.jpeg catalogizer-android/app/src/main/res/drawable/app_icon.jpeg
cp Assets/Logo/Vasic_Digital_Logo.jpeg catalogizer-android/app/src/main/res/drawable/vasic_digital_logo.jpeg

# Android TV
cp Assets/Logo/Application_Icon.jpeg catalogizer-androidtv/app/src/main/res/drawable/app_icon.jpeg
cp Assets/Logo/Vasic_Digital_Logo.jpeg catalogizer-androidtv/app/src/main/res/drawable/vasic_digital_logo.jpeg
```

- [ ] **Step 2: Commit**

```bash
git add catalog-web/src/assets/app-icon.jpeg catalog-web/src/assets/vasic-digital-logo.jpeg
git add catalogizer-desktop/src/assets/ installer-wizard/src/assets/
git commit -m "feat: copy brand assets to all application directories"

cd catalogizer-android && git add app/src/main/res/drawable/vasic_digital_logo.jpeg app/src/main/res/drawable/app_icon.jpeg && git commit -m "feat: add brand assets for splash screen" && cd ..
cd catalogizer-androidtv && git add app/src/main/res/drawable/vasic_digital_logo.jpeg app/src/main/res/drawable/app_icon.jpeg && git commit -m "feat: add brand assets for splash screen" && cd ..
```

---

## Task 2: Create Web App Splash Screen

**Files:**
- Create: `catalog-web/src/components/SplashScreen.tsx`
- Create: `catalog-web/src/components/__tests__/SplashScreen.test.tsx`
- Modify: `catalog-web/src/App.tsx`

- [ ] **Step 1: Write failing test**

Create `catalog-web/src/components/__tests__/SplashScreen.test.tsx`:

```tsx
import { render, screen, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SplashScreen } from '../SplashScreen';

describe('SplashScreen', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders splash with app icon, title, and footer', () => {
    render(<SplashScreen onComplete={() => {}} />);
    expect(screen.getByText('Catalogizer')).toBeInTheDocument();
    expect(screen.getByText(/Advanced Multi-Protocol/)).toBeInTheDocument();
    expect(screen.getByText(/Made with/)).toBeInTheDocument();
    expect(screen.getByText(/Vasic Digital/)).toBeInTheDocument();
    expect(screen.getByText(/v1.1.0/)).toBeInTheDocument();
  });

  it('shows for at least 5 seconds on first launch', async () => {
    const onComplete = vi.fn();
    render(<SplashScreen onComplete={onComplete} />);

    act(() => { vi.advanceTimersByTime(4900); });
    expect(onComplete).not.toHaveBeenCalled();

    act(() => { vi.advanceTimersByTime(200); });
    expect(onComplete).toHaveBeenCalled();
  });

  it('shows for at least 2.5 seconds on regular launch', async () => {
    localStorage.setItem('catalogizer_launched', 'true');
    const onComplete = vi.fn();
    render(<SplashScreen onComplete={onComplete} />);

    act(() => { vi.advanceTimersByTime(2400); });
    expect(onComplete).not.toHaveBeenCalled();

    act(() => { vi.advanceTimersByTime(200); });
    expect(onComplete).toHaveBeenCalled();
  });

  it('sets launched flag in localStorage', () => {
    render(<SplashScreen onComplete={() => {}} />);
    expect(localStorage.getItem('catalogizer_launched')).toBe('true');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd catalog-web && npm run test -- --run src/components/__tests__/SplashScreen.test.tsx`
Expected: FAIL — SplashScreen not found

- [ ] **Step 3: Create SplashScreen component**

Create `catalog-web/src/components/SplashScreen.tsx`:

```tsx
import { useEffect, useState, useRef } from 'react';
import appIcon from '@/assets/app-icon.jpeg';
import vdLogo from '@/assets/vasic-digital-logo.jpeg';

interface SplashScreenProps {
  onComplete: () => void;
  appTitle?: string;
  subtitle?: string;
}

export function SplashScreen({
  onComplete,
  appTitle = 'Catalogizer',
  subtitle = 'Advanced Multi-Protocol Media Collection Management System',
}: SplashScreenProps) {
  const [visible, setVisible] = useState(true);
  const completeCalled = useRef(false);

  useEffect(() => {
    const isFirstLaunch = !localStorage.getItem('catalogizer_launched');
    localStorage.setItem('catalogizer_launched', 'true');
    const minDuration = isFirstLaunch ? 5000 : 2500;

    const timer = setTimeout(() => {
      if (!completeCalled.current) {
        completeCalled.current = true;
        setVisible(false);
        onComplete();
      }
    }, minDuration);

    return () => clearTimeout(timer);
  }, [onComplete]);

  if (!visible) return null;

  return (
    <div className="fixed inset-0 z-[9999] flex flex-col items-center justify-center"
         style={{ background: 'linear-gradient(180deg, #0F172A 0%, #1E293B 100%)' }}>
      <div className="flex-1 flex flex-col items-center justify-center">
        <img src={appIcon} alt="Catalogizer" className="w-28 h-28 rounded-2xl mb-8 shadow-2xl" />
        <h1 className="text-3xl font-bold text-white mb-2">{appTitle}</h1>
        <p className="text-sm text-slate-400 text-center max-w-md px-4">{subtitle}</p>
        <div className="mt-8">
          <div className="w-8 h-8 border-3 border-blue-500 border-t-transparent rounded-full animate-spin" />
        </div>
      </div>
      <div className="pb-8 flex flex-col items-center gap-2">
        <div className="flex items-center gap-2">
          <img src={vdLogo} alt="Vasic Digital" className="w-6 h-6 rounded" />
          <span className="text-slate-500 text-xs">Made with ♥ by Vasic Digital</span>
        </div>
        <span className="text-slate-600 text-xs">v1.1.0</span>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd catalog-web && npm run test -- --run src/components/__tests__/SplashScreen.test.tsx`
Expected: PASS

- [ ] **Step 5: Wire SplashScreen into App.tsx**

In `catalog-web/src/App.tsx`, add the splash screen wrapper. Replace the `App` function body:

```tsx
import { SplashScreen } from '@/components/SplashScreen'

// Add state at the top of App():
const [splashComplete, setSplashComplete] = useState(false);

// Wrap the return: show splash when not complete
if (!splashComplete) {
  return <SplashScreen onComplete={() => setSplashComplete(true)} />;
}

// ... existing return with ErrorBoundary/Router/etc
```

- [ ] **Step 6: Run all web tests**

Run: `cd catalog-web && npm run test -- --run`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
cd catalog-web
git add src/components/SplashScreen.tsx src/components/__tests__/SplashScreen.test.tsx src/App.tsx
git commit -m "feat: add branded splash screen to web app"
```

---

## Task 3: Create Desktop App Splash Screen

**Files:**
- Create: `catalogizer-desktop/src/components/SplashScreen.tsx`
- Create: `catalogizer-desktop/src/components/__tests__/SplashScreen.test.tsx`
- Modify: `catalogizer-desktop/src/App.tsx`
- Delete: `catalogizer-desktop/src/components/LoadingScreen.tsx` (replaced)

Same pattern as Task 2 but adapted for the Tauri desktop app. The SplashScreen component is the same design. In `App.tsx`, replace the `LoadingScreen` import with `SplashScreen` and add minimum duration enforcement that combines with the existing `isInitialized` state.

- [ ] **Step 1: Write test** (same structure as Task 2 but in desktop test directory)
- [ ] **Step 2: Create SplashScreen.tsx** (same component, copy from web, adjust import paths)
- [ ] **Step 3: Update App.tsx** — replace `LoadingScreen` usage with `SplashScreen`, show splash until BOTH `isInitialized` AND minimum time elapsed
- [ ] **Step 4: Delete LoadingScreen.tsx**
- [ ] **Step 5: Run tests**: `cd catalogizer-desktop && npm test`
- [ ] **Step 6: Commit**

```bash
cd catalogizer-desktop
git add src/components/SplashScreen.tsx src/components/__tests__/SplashScreen.test.tsx src/App.tsx
git rm src/components/LoadingScreen.tsx
git commit -m "feat: add branded splash screen to desktop app, replace LoadingScreen"
```

---

## Task 4: Create Installer Wizard Splash Screen

**Files:**
- Create: `installer-wizard/src/components/SplashScreen.tsx`
- Create: `installer-wizard/src/components/__tests__/SplashScreen.test.tsx`
- Modify: `installer-wizard/src/App.tsx`

Same component pattern. Title is "Catalogizer Installation Wizard". Wire into App.tsx before the ConfigurationProvider/WizardProvider mount.

- [ ] **Step 1: Write test** (title assertion uses "Catalogizer Installation Wizard")
- [ ] **Step 2: Create SplashScreen.tsx** (same design, customized title)
- [ ] **Step 3: Update App.tsx** — add splash state, show splash before router
- [ ] **Step 4: Run tests**: `cd installer-wizard && npm test`
- [ ] **Step 5: Commit**

```bash
cd installer-wizard
git add src/components/SplashScreen.tsx src/components/__tests__/SplashScreen.test.tsx src/App.tsx
git commit -m "feat: add branded splash screen to installer wizard"
```

---

## Task 5: Create Android Branded Splash Composable

**Files:**
- Create: `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/splash/SplashContent.kt`
- Modify: `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/MainActivity.kt`
- Modify: `catalogizer-android/app/src/main/res/values/styles.xml`

- [ ] **Step 1: Update splash theme for longer display**

In `catalogizer-android/app/src/main/res/values/styles.xml`, change animation duration:

```xml
<item name="windowSplashScreenAnimationDuration">5000</item>
```

- [ ] **Step 2: Create SplashContent composable**

Create `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/splash/SplashContent.kt`:

```kotlin
package com.catalogizer.android.ui.splash

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.catalogizer.android.BuildConfig
import com.catalogizer.android.R
import kotlinx.coroutines.delay

@Composable
fun SplashContent(
    isAppReady: Boolean,
    onSplashComplete: () -> Unit
) {
    val isFirstLaunch = remember { true } // Will be set from SharedPreferences
    val minDuration = if (isFirstLaunch) 5000L else 2500L

    LaunchedEffect(isAppReady) {
        delay(minDuration)
        if (isAppReady) {
            onSplashComplete()
        }
    }

    LaunchedEffect(isAppReady, minDuration) {
        if (isAppReady) {
            delay(minDuration)
            onSplashComplete()
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(Color(0xFF0F172A), Color(0xFF1E293B))
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Image(
                painter = painterResource(R.drawable.app_icon),
                contentDescription = "Catalogizer",
                modifier = Modifier.size(96.dp),
                contentScale = ContentScale.Fit
            )
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = "Catalogizer",
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Advanced Multi-Protocol Media\nCollection Management System",
                fontSize = 14.sp,
                color = Color(0xFF94A3B8),
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(32.dp))
            CircularProgressIndicator(
                modifier = Modifier.size(32.dp),
                color = Color(0xFF4A90E2),
                strokeWidth = 3.dp
            )
        }

        // Footer
        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .padding(bottom = 32.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Image(
                    painter = painterResource(R.drawable.vasic_digital_logo),
                    contentDescription = "Vasic Digital",
                    modifier = Modifier.size(24.dp)
                )
                Text(
                    text = "Made with \u2764\uFE0F by Vasic Digital",
                    fontSize = 12.sp,
                    color = Color(0xFF64748B)
                )
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = "v${BuildConfig.VERSION_NAME}",
                fontSize = 11.sp,
                color = Color(0xFF475569)
            )
        }
    }
}
```

- [ ] **Step 3: Update MainActivity.kt to show branded splash**

In `catalogizer-android/app/src/main/java/com/catalogizer/android/ui/MainActivity.kt`, modify the `CatalogizerApp` composable to show `SplashContent` while loading:

```kotlin
// Replace the if (!isLoading) block with:
var splashComplete by remember { mutableStateOf(false) }

if (!splashComplete) {
    SplashContent(
        isAppReady = !isLoading,
        onSplashComplete = { splashComplete = true }
    )
} else {
    CatalogizerNavigation(
        isAuthenticated = authState.isAuthenticated,
        authViewModel = authViewModel,
        homeViewModel = homeViewModel,
        searchViewModel = searchViewModel
    )
}
```

- [ ] **Step 4: Commit in submodule**

```bash
cd catalogizer-android
git add app/src/main/java/com/catalogizer/android/ui/splash/SplashContent.kt
git add app/src/main/java/com/catalogizer/android/ui/MainActivity.kt
git add app/src/main/res/values/styles.xml
git commit -m "feat: add branded splash screen with Vasic Digital footer"
```

---

## Task 6: Create Android TV Branded Splash Composable

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/splash/SplashContent.kt`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/MainActivity.kt`
- Modify: `catalogizer-androidtv/app/src/main/res/values/styles.xml`

Same as Task 5 but adapted for TV: larger text sizes (TV is viewed from distance), uses `androidx.tv.material3` components, D-pad safe layout.

- [ ] **Step 1: Add splash theme to styles.xml**

Add to `catalogizer-androidtv/app/src/main/res/values/styles.xml`:

```xml
<style name="Theme.CatalogizerTV.Splash" parent="Theme.SplashScreen">
    <item name="windowSplashScreenBackground">@android:color/black</item>
    <item name="windowSplashScreenAnimatedIcon">@mipmap/ic_launcher</item>
    <item name="windowSplashScreenAnimationDuration">5000</item>
    <item name="postSplashScreenTheme">@style/Theme.CatalogizerTV</item>
</style>
```

- [ ] **Step 2: Create SplashContent.kt** (TV version — larger sizes: 36sp title, 120dp icon, 16sp subtitle)

- [ ] **Step 3: Add SplashScreen API to MainActivity.kt**

```kotlin
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen

// In onCreate, before super.onCreate:
val splashScreen = installSplashScreen()

// After viewmodel init:
splashScreen.setKeepOnScreenCondition { mainViewModel.isLoading.value }
```

And modify `CatalogizerTVApp` composable same as Task 5 Step 3.

- [ ] **Step 4: Update AndroidManifest.xml theme** to use `Theme.CatalogizerTV.Splash`

- [ ] **Step 5: Commit in submodule**

```bash
cd catalogizer-androidtv
git add app/src/main/java/com/catalogizer/androidtv/ui/splash/SplashContent.kt
git add app/src/main/java/com/catalogizer/androidtv/ui/MainActivity.kt
git add app/src/main/res/values/styles.xml app/src/main/AndroidManifest.xml
git commit -m "feat: add branded splash screen with Vasic Digital footer"
```

---

## Task 7: Add Version Footer to Landing Screens

**Files:**
- Modify: `catalog-web/src/components/layout/Layout.tsx`
- Modify: `catalogizer-desktop/src/components/Layout.tsx`
- Modify: `installer-wizard/src/components/layout/WizardLayout.tsx`

- [ ] **Step 1: Add version footer to web Layout**

In `catalog-web/src/components/layout/Layout.tsx`, add at the bottom of the layout:

```tsx
<span className="fixed bottom-2 right-3 text-xs text-slate-400/50 select-none">v1.1.0</span>
```

- [ ] **Step 2: Add version footer to desktop Layout**

Same pattern, positioned bottom-left of sidebar.

- [ ] **Step 3: Add version footer to wizard WizardLayout**

Bottom-center of wizard container.

- [ ] **Step 4: Run all frontend tests**

```bash
cd catalog-web && npm run test -- --run
cd ../catalogizer-desktop && npm test
cd ../installer-wizard && npm test
```

- [ ] **Step 5: Commit**

```bash
git add catalog-web/src/components/layout/Layout.tsx
git add catalogizer-desktop/src/components/Layout.tsx
git add installer-wizard/src/components/layout/WizardLayout.tsx
git commit -m "feat: add version footer to all app landing screens"
```

---

## Task 8: Bump All Versions to 1.1.0

**Files:**
- Modify: `catalog-web/package.json`
- Modify: `catalogizer-desktop/package.json`
- Modify: `catalogizer-desktop/src-tauri/tauri.conf.json`
- Modify: `installer-wizard/package.json`
- Modify: `installer-wizard/src-tauri/tauri.conf.json`
- Modify: `catalogizer-android/app/build.gradle.kts`
- Modify: `catalogizer-androidtv/app/build.gradle.kts`

- [ ] **Step 1: Bump all version numbers**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer

# Web
sed -i 's/"version": "1.0.0"/"version": "1.1.0"/' catalog-web/package.json

# Desktop
sed -i 's/"version": "1.0.0"/"version": "1.1.0"/' catalogizer-desktop/package.json
sed -i 's/"version": "1.0.0"/"version": "1.1.0"/' catalogizer-desktop/src-tauri/tauri.conf.json

# Wizard
sed -i 's/"version": "1.0.0"/"version": "1.1.0"/' installer-wizard/package.json
sed -i 's/"version": "1.0.0"/"version": "1.1.0"/' installer-wizard/src-tauri/tauri.conf.json
```

For Android/TV, update `versionCode` and `versionName` in each `build.gradle.kts`.

- [ ] **Step 2: Commit**

```bash
git add catalog-web/package.json catalogizer-desktop/package.json catalogizer-desktop/src-tauri/tauri.conf.json
git add installer-wizard/package.json installer-wizard/src-tauri/tauri.conf.json
git commit -m "chore: bump all app versions to 1.1.0"

cd catalogizer-android && git add app/build.gradle.kts && git commit -m "chore: bump version to 1.1.0" && cd ..
cd catalogizer-androidtv && git add app/build.gradle.kts && git commit -m "chore: bump version to 1.1.0" && cd ..
```

---

## Task 9: Run Tests and Capture Evidence

- [ ] **Step 1: Run all web tests with coverage**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm run test -- --run
npm run type-check
npm run lint
```

- [ ] **Step 2: Run desktop and wizard tests**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalogizer-desktop && npm test
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/installer-wizard && npm test
```

- [ ] **Step 3: Run Go backend tests** (verify no regressions)

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1 -short
```

- [ ] **Step 4: Build and install Android APKs on devices**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalogizer-android
./gradlew assembleDebug

cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalogizer-androidtv
./gradlew assembleDebug

# Install on all 4 devices
adb -s 19bbb528a1dbbc4d install -r catalogizer-android/app/build/outputs/apk/debug/app-debug.apk
adb -s 1acdceab90248933 install -r catalogizer-android/app/build/outputs/apk/debug/app-debug.apk
adb -s 192.168.0.134:5555 install -r catalogizer-androidtv/app/build/outputs/apk/debug/app-debug.apk
adb -s 192.168.0.214:5555 install -r catalogizer-androidtv/app/build/outputs/apk/debug/app-debug.apk
```

- [ ] **Step 5: Capture splash screen screenshots from all 4 devices**

```bash
mkdir -p qa-results/splash-screens-$(date +%Y%m%d)

# Launch apps and capture splash
for device in 19bbb528a1dbbc4d 1acdceab90248933; do
  adb -s $device shell am force-stop com.catalogizer.android
  adb -s $device shell am start -n com.catalogizer.android/.ui.MainActivity
  sleep 1
  adb -s $device exec-out screencap -p > "qa-results/splash-screens-$(date +%Y%m%d)/android-$device.png"
done

for device in 192.168.0.134:5555 192.168.0.214:5555; do
  adb -s $device shell am force-stop com.catalogizer.androidtv
  adb -s $device shell am start -n com.catalogizer.androidtv/.ui.MainActivity
  sleep 1
  adb -s $device exec-out screencap -p > "qa-results/splash-screens-$(date +%Y%m%d)/tv-${device%%:*}.png"
done
```

- [ ] **Step 6: Record video of splash screens**

```bash
for device in 19bbb528a1dbbc4d 1acdceab90248933; do
  adb -s $device shell am force-stop com.catalogizer.android
  adb -s $device shell screenrecord --time-limit 10 /sdcard/splash.mp4 &
  sleep 1
  adb -s $device shell am start -n com.catalogizer.android/.ui.MainActivity
  sleep 10
  adb -s $device pull /sdcard/splash.mp4 "qa-results/splash-screens-$(date +%Y%m%d)/android-$device.mp4"
done
```

- [ ] **Step 7: Commit evidence**

Screenshots and videos go to `qa-results/` (gitignored). Document locations in the final report.

---

## Task 10: Analyze Screenshots and Fix Issues

- [ ] **Step 1: Visually inspect all screenshots**

Read each captured screenshot file. Check:
- App icon is centered and properly sized
- Title text is white, readable, properly positioned
- Subtitle is visible in slate-400 color
- Loading indicator is visible and animated
- VD logo is visible in footer
- "Made with ♥ by Vasic Digital" text is present
- Version number is displayed
- No visual artifacts, clipping, or misalignment

- [ ] **Step 2: Fix any issues found**

For each issue, apply the fix in the relevant component and rebuild/retest.

- [ ] **Step 3: Re-capture and verify**

Repeat screenshot capture after fixes. Verify all issues resolved.

---

## Task 11: Create Release Builds

- [ ] **Step 1: Build release APKs**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalogizer-android
./gradlew assembleRelease
cp app/build/outputs/apk/release/app-release.apk ../../releases/catalogizer-android-1.1.0.apk

cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalogizer-androidtv
./gradlew assembleRelease
cp app/build/outputs/apk/release/app-release.apk ../../releases/catalogizer-androidtv-1.1.0.apk
```

- [ ] **Step 2: Build web production bundle**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-web
npm run build
cp -r dist/ ../releases/catalog-web-1.1.0/
```

- [ ] **Step 3: Verify all releases in releases/ directory**

---

## Task 12: Final Commit, Submodule Update, and Push

- [ ] **Step 1: Update submodule pointers**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
git add catalogizer-android catalogizer-androidtv
git commit -m "chore: update Android submodule pointers for v1.1.0"
```

- [ ] **Step 2: Create final report**

Create `docs/status/SPLASH_SCREEN_IMPLEMENTATION_REPORT_2026-03-26.md` with all results, screenshots locations, test counts, and issue log.

- [ ] **Step 3: Push to all remotes**

```bash
# Submodules first
cd catalogizer-android && git push origin main && cd ..
cd catalogizer-androidtv && git push origin main && cd ..

# Main repo
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

---

## Summary

| Task | What | Apps |
|------|------|------|
| 1 | Copy brand assets | All 5 |
| 2 | Web splash screen | catalog-web |
| 3 | Desktop splash screen | catalogizer-desktop |
| 4 | Wizard splash screen | installer-wizard |
| 5 | Android splash screen | catalogizer-android |
| 6 | Android TV splash screen | catalogizer-androidtv |
| 7 | Version footer on landing screens | Web, Desktop, Wizard |
| 8 | Version bump to 1.1.0 | All 5 |
| 9 | Test execution + evidence capture | All |
| 10 | Screenshot analysis + fixes | All |
| 11 | Release builds | All |
| 12 | Final commit + push | All |
