# End-to-End (E2E) Test Plan

**Date**: 2026-02-10
**Status**: ⏳ In Progress - Foundation Complete, Expansion Planned

## Overview

This document outlines the comprehensive E2E testing strategy for Catalogizer across all platforms: Web (Playwright), Android (Espresso/Maestro), and AndroidTV.

---

## Current State

### catalog-web (Playwright)

**Location**: `catalog-web/e2e/tests/`

**Current Tests**: 5 spec files

| File | Tests | Coverage |
|------|-------|----------|
| `auth.spec.ts` | 14 | Login, registration, logout, protected routes |
| `dashboard.spec.ts` | - | Dashboard functionality |
| `media-browser.spec.ts` | - | Media browsing, search, playback |
| `collections.spec.ts` | - | Collection management |
| `protected-routes.spec.ts` | - | Route protection |

**Framework**: Playwright
**Execution**: `npm run test:e2e`
**CI/CD**: Configured but GitHub Actions disabled

### catalogizer-android

**Status**: ⏳ Planned
**Framework**: Espresso / Maestro
**Test Files**: To be created

### catalogizer-androidtv

**Status**: ⏳ Planned
**Framework**: Espresso / Maestro
**Test Files**: To be created

---

## Target Coverage

### Web (catalog-web) - Target: 80+ tests

#### 1. Authentication (20 tests)

**Login Flow** (10 tests):
- ✅ Display login form
- ✅ Login with valid credentials
- ✅ Show error with invalid credentials
- ✅ Validate required fields
- ✅ Show register link
- ✅ Show remember me checkbox
- ✅ Show forgot password link
- [ ] Remember me functionality
- [ ] Password visibility toggle
- [ ] Login with Enter key

**Registration Flow** (5 tests):
- ✅ Display registration form
- ✅ Navigate to login page
- [ ] Register with valid data
- [ ] Show validation errors
- [ ] Check password strength indicator

**Session Management** (5 tests):
- ✅ Store auth token after login
- ✅ Store user data after login
- [ ] Logout clears session
- [ ] Session persistence on page reload
- [ ] Session expiration handling

#### 2. Dashboard (10 tests)

- [ ] Display welcome message with username
- [ ] Show recent media stats
- [ ] Display storage usage chart
- [ ] Show quick actions panel
- [ ] Display recent activity feed
- [ ] Navigate to media browser
- [ ] Navigate to analytics
- [ ] Show user profile dropdown
- [ ] Display notifications badge
- [ ] Responsive layout on mobile/tablet

#### 3. Media Browser (15 tests)

**Browsing** (8 tests):
- [ ] Display media grid/list view
- [ ] Toggle between grid and list views
- [ ] Filter by media type (movies, TV, music)
- [ ] Sort by name, date, size
- [ ] Pagination works correctly
- [ ] Virtual scrolling for large libraries
- [ ] Thumbnail lazy loading
- [ ] Display media metadata (title, year, rating)

**Search** (4 tests):
- [ ] Search with keyword
- [ ] Search autocomplete suggestions
- [ ] Advanced search with filters
- [ ] Clear search results

**Media Details** (3 tests):
- [ ] Display media details modal
- [ ] Show cover art and metadata
- [ ] Play media file

#### 4. Media Playback (10 tests)

- [ ] Play video file
- [ ] Pause/resume playback
- [ ] Seek to position
- [ ] Adjust volume
- [ ] Toggle fullscreen
- [ ] Display subtitles
- [ ] Change subtitle track
- [ ] Change audio track
- [ ] Picture-in-picture mode
- [ ] Playback error handling

#### 5. Collections (8 tests)

- [ ] Display collections list
- [ ] Create new collection
- [ ] Add media to collection
- [ ] Remove media from collection
- [ ] Rename collection
- [ ] Delete collection
- [ ] Collection detail view
- [ ] Share collection

#### 6. Playlists (6 tests)

- [ ] Display playlists
- [ ] Create playlist
- [ ] Add items to playlist
- [ ] Reorder playlist items
- [ ] Play playlist
- [ ] Delete playlist

#### 7. Favorites (4 tests)

- [ ] Add to favorites
- [ ] View favorites page
- [ ] Remove from favorites
- [ ] Sort favorites

#### 8. Search & Filters (6 tests)

- [ ] Global search
- [ ] Filter by type
- [ ] Filter by year
- [ ] Filter by rating
- [ ] Combined filters
- [ ] Clear all filters

#### 9. Admin Panel (5 tests)

- [ ] Access admin panel (admin only)
- [ ] View user list
- [ ] Create new user
- [ ] Edit user permissions
- [ ] View system logs

#### 10. Settings (4 tests)

- [ ] View user profile
- [ ] Update profile information
- [ ] Change password
- [ ] Update preferences

#### 11. WebSocket Real-time Updates (4 tests)

- [ ] Receive scan progress updates
- [ ] Receive download progress updates
- [ ] Receive new media notifications
- [ ] WebSocket reconnection

#### 12. Offline Functionality (4 tests)

- [ ] Show offline indicator
- [ ] Cache media for offline viewing
- [ ] Sync when back online
- [ ] Queue operations while offline

#### 13. Accessibility (4 tests)

- [ ] Keyboard navigation works
- [ ] Screen reader compatibility
- [ ] Focus management
- [ ] ARIA labels present

---

## Android/AndroidTV - Target: 60+ tests per app

### 1. Authentication (10 tests)

- [ ] Display splash screen
- [ ] Show server configuration screen
- [ ] Connect to server
- [ ] Login with credentials
- [ ] Remember credentials
- [ ] Logout
- [ ] Session persistence
- [ ] Biometric authentication (mobile only)
- [ ] PIN code lock
- [ ] Handle network errors

### 2. Media Browsing (12 tests)

- [ ] Display media grid (TV: leanback layout)
- [ ] Navigate with D-pad (TV)
- [ ] Filter by type
- [ ] Sort options
- [ ] Search functionality
- [ ] Voice search (TV)
- [ ] Thumbnail loading
- [ ] Smooth scrolling
- [ ] Load more pagination
- [ ] Pull to refresh (mobile)
- [ ] Grid/list toggle (mobile)
- [ ] Media details screen

### 3. Media Playback (15 tests)

- [ ] Play video file
- [ ] ExoPlayer integration
- [ ] Pause/resume
- [ ] Seek forward/backward
- [ ] D-pad controls (TV)
- [ ] Subtitle selection
- [ ] Audio track selection
- [ ] Playback speed control
- [ ] Picture-in-picture (mobile)
- [ ] Background playback (music)
- [ ] Notification controls (mobile)
- [ ] Cast to Chromecast
- [ ] Resume from last position
- [ ] Next/previous episode
- [ ] Playback quality selection

### 4. Offline Mode (8 tests)

- [ ] Download media for offline
- [ ] Manage downloaded files
- [ ] Play offline content
- [ ] Delete downloads
- [ ] Download progress notification
- [ ] Storage space management
- [ ] Sync downloaded metadata
- [ ] Offline favorites

### 5. Collections & Playlists (6 tests)

- [ ] View collections
- [ ] Add to collection
- [ ] Create playlist
- [ ] Play playlist
- [ ] Reorder items
- [ ] Delete collection/playlist

### 6. Search (4 tests)

- [ ] Text search
- [ ] Voice search
- [ ] Search suggestions
- [ ] Recent searches

### 7. Settings (5 tests)

- [ ] Access settings
- [ ] Change playback quality
- [ ] Update subtitle preferences
- [ ] Configure downloads
- [ ] Clear cache

---

## Test Implementation Patterns

### Playwright Pattern (Web)

```typescript
import { test, expect, Page } from '@playwright/test';

test.describe('Feature Name', () => {
  let page: Page;

  test.beforeEach(async ({ page: testPage }) => {
    page = testPage;
    // Setup: login, navigate, mock APIs
    await login(page);
  });

  test('should do something', async () => {
    // Arrange
    await page.goto('/path');

    // Act
    await page.click('button[data-testid="action"]');

    // Assert
    await expect(page.locator('[data-testid="result"]')).toBeVisible();
    await expect(page).toHaveURL(/expected-url/);
  });

  test('should handle errors', async () => {
    // Mock API error
    await page.route('**/api/endpoint', route =>
      route.fulfill({ status: 500, body: 'Error' })
    );

    // Act & Assert
    await page.click('button');
    await expect(page.locator('.error-message')).toBeVisible();
  });
});
```

### Espresso Pattern (Android)

```kotlin
@RunWith(AndroidJUnit4::class)
class MediaBrowserTest {
    @get:Rule
    val activityRule = ActivityScenarioRule(MainActivity::class.java)

    @Before
    fun setup() {
        // Setup: mock server, login
    }

    @Test
    fun shouldDisplayMediaGrid() {
        // Arrange
        val testMedia = createTestMedia()

        // Act
        onView(withId(R.id.mediaGrid))
            .perform(scrollTo())

        // Assert
        onView(withId(R.id.mediaGrid))
            .check(matches(isDisplayed()))
        onView(withText("Test Movie"))
            .check(matches(isDisplayed()))
    }

    @Test
    fun shouldPlayMediaOnClick() {
        // Act
        onView(withId(R.id.mediaItem))
            .perform(click())

        // Assert
        onView(withId(R.id.playerView))
            .check(matches(isDisplayed()))
        assertTrue(player.isPlaying)
    }
}
```

### Maestro Pattern (Android/AndroidTV)

```yaml
# media-playback.yaml
appId: com.catalogizer.android
---
- launchApp
- tapOn: "Movies"
- assertVisible: "The Matrix"
- tapOn: "The Matrix"
- assertVisible: "Play"
- tapOn: "Play"
- assertVisible: "Player"
- waitForAnimationToEnd
- assertTrue: ${player.isPlaying}
```

---

## Execution Strategy

### Local Execution

**Web (Playwright)**:
```bash
cd catalog-web
npm run test:e2e                    # All tests
npm run test:e2e -- --headed        # With browser UI
npm run test:e2e -- auth.spec.ts   # Single spec
npm run test:e2e -- --debug         # Debug mode
```

**Android (Espresso)**:
```bash
cd catalogizer-android
./gradlew connectedAndroidTest      # On connected device
./gradlew testDebugUnitTest         # Unit tests
```

**Android (Maestro)**:
```bash
maestro test media-playback.yaml
maestro test --format junit flows/  # All flows
```

### CI/CD Integration

**GitHub Actions Example** (when re-enabled):

```yaml
name: E2E Tests
on: [pull_request]

jobs:
  web-e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - name: Install dependencies
        run: cd catalog-web && npm ci
      - name: Run Playwright tests
        run: cd catalog-web && npm run test:e2e
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: catalog-web/playwright-report/

  android-e2e:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-java@v3
      - name: Run Espresso tests
        uses: reactivecircus/android-emulator-runner@v2
        with:
          api-level: 30
          script: ./gradlew connectedAndroidTest
```

---

## Test Data Management

### Fixtures (Web)

**Location**: `catalog-web/e2e/fixtures/`

```typescript
// auth.ts
export const testUser = {
  username: 'testuser',
  password: 'testpass123',
  email: 'test@example.com'
};

export async function login(page: Page) {
  await page.goto('/login');
  await page.fill('[placeholder*="username"]', testUser.username);
  await page.fill('[placeholder*="password"]', testUser.password);
  await page.click('button[type="submit"]');
  await page.waitForURL('**/dashboard');
}

// media.ts
export const testMedia = {
  movie: {
    id: 1,
    title: 'Test Movie',
    year: 2024,
    type: 'movie',
    path: '/path/to/movie.mp4'
  },
  tvShow: {
    id: 2,
    title: 'Test Show',
    year: 2024,
    type: 'tv',
    seasons: 1,
    episodes: 10
  }
};

export function mockMediaEndpoints(page: Page) {
  page.route('**/api/v1/media', route =>
    route.fulfill({
      status: 200,
      body: JSON.stringify([testMedia.movie, testMedia.tvShow])
    })
  );
}
```

### Mock Server (Android)

```kotlin
// MockWebServerRule.kt
class MockWebServerRule : TestWatcher() {
    private val server = MockWebServer()

    override fun starting(description: Description) {
        server.start()
    }

    override fun finished(description: Description) {
        server.shutdown()
    }

    fun enqueue(response: MockResponse) {
        server.enqueue(response)
    }

    fun url(path: String): String = server.url(path).toString()
}
```

---

## Coverage Tracking

### Web E2E Coverage

**Tool**: Playwright Coverage Plugin

```typescript
// playwright.config.ts
export default {
  use: {
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  reporter: [
    ['html'],
    ['junit', { outputFile: 'results.xml' }],
    ['json', { outputFile: 'results.json' }]
  ]
};
```

**Coverage Report**:
```bash
npm run test:e2e
open playwright-report/index.html
```

### Android E2E Coverage

**Tool**: Jacoco

```gradle
android {
    buildTypes {
        debug {
            testCoverageEnabled true
        }
    }
}

tasks.withType(Test) {
    jacoco.includeNoLocationClasses = true
    jacoco.excludes = ['jdk.internal.*']
}
```

**Generate Report**:
```bash
./gradlew createDebugCoverageReport
open app/build/reports/coverage/debug/index.html
```

---

## Best Practices

### 1. Test Independence

✅ **Good**:
```typescript
test('should create collection', async ({ page }) => {
  // Setup in test
  await login(page);
  await page.goto('/collections');

  // Test logic
  await page.click('[data-testid="create-collection"]');
  // ...
});
```

❌ **Bad**:
```typescript
// Don't rely on previous test state
test('should edit collection', async ({ page }) => {
  // Assumes collection exists from previous test
  await page.click('[data-testid="edit"]');
});
```

### 2. Use Data Attributes

✅ **Good**:
```typescript
await page.click('[data-testid="submit-button"]');
```

❌ **Bad**:
```typescript
await page.click('.btn-primary.submit');  // Fragile
```

### 3. Wait for Elements

✅ **Good**:
```typescript
await page.waitForSelector('[data-testid="result"]', { state: 'visible' });
await expect(page.locator('[data-testid="result"]')).toBeVisible();
```

❌ **Bad**:
```typescript
await page.waitForTimeout(2000);  // Flaky
```

### 4. Mock External Dependencies

✅ **Good**:
```typescript
await page.route('**/api/external/**', route =>
  route.fulfill({ body: mockData })
);
```

### 5. Use Page Objects

```typescript
// pages/LoginPage.ts
export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
  }

  async login(username: string, password: string) {
    await this.page.fill('[data-testid="username"]', username);
    await this.page.fill('[data-testid="password"]', password);
    await this.page.click('[data-testid="submit"]');
  }
}

// Usage
const loginPage = new LoginPage(page);
await loginPage.goto();
await loginPage.login('user', 'pass');
```

---

## Debugging

### Playwright

**Interactive Mode**:
```bash
npm run test:e2e -- --debug
```

**Trace Viewer**:
```bash
npx playwright show-trace trace.zip
```

**Inspector**:
```typescript
await page.pause();  // Pause execution
```

### Android Espresso

**Layout Inspector**:
- Tools → Layout Inspector in Android Studio

**Espresso Test Recorder**:
- Run → Record Espresso Test

**Logcat Filtering**:
```bash
adb logcat -s TestRunner
```

---

## Performance Considerations

### 1. Parallel Execution

**Playwright**:
```typescript
// playwright.config.ts
export default {
  workers: 4,  // Run 4 tests in parallel
  fullyParallel: true
};
```

### 2. Test Sharding

```bash
npm run test:e2e -- --shard=1/4
npm run test:e2e -- --shard=2/4
# ...
```

### 3. Selective Testing

```bash
# Run only auth tests
npm run test:e2e -- auth.spec.ts

# Run tests matching pattern
npm run test:e2e -- --grep="login"
```

---

## Roadmap

### Phase 1: Foundation (✅ Complete)
- [x] Set up Playwright for web
- [x] Create basic auth tests
- [x] Create test fixtures
- [x] Configure CI/CD structure

### Phase 2: Web Expansion (⏳ In Progress)
- [ ] Expand to 80+ Playwright tests
- [ ] Add visual regression tests
- [ ] Add accessibility tests
- [ ] Add performance tests

### Phase 3: Android (🔜 Planned)
- [ ] Set up Espresso/Maestro
- [ ] Create 60+ Android tests
- [ ] Add UI automator tests
- [ ] Add integration tests

### Phase 4: AndroidTV (🔜 Planned)
- [ ] Set up leanback testing
- [ ] Create 60+ AndroidTV tests
- [ ] Add D-pad navigation tests
- [ ] Add voice search tests

---

## Resources

### Documentation

- **Playwright**: https://playwright.dev/
- **Espresso**: https://developer.android.com/training/testing/espresso
- **Maestro**: https://maestro.mobile.dev/
- **Testing Library**: https://testing-library.com/

### Tools

- **Playwright Test Generator**: `npx playwright codegen`
- **Espresso Test Recorder**: Android Studio → Run → Record Espresso Test
- **Maestro Studio**: `maestro studio`

---

## Conclusion

**Current Status**:
- ✅ E2E testing infrastructure established
- ✅ 5 Playwright spec files created
- ✅ Test patterns and fixtures defined
- ⏳ Expansion to 80+ web tests planned
- 🔜 Android/AndroidTV E2E tests planned

**Next Steps**:
1. Expand Playwright tests to cover all critical user flows
2. Add visual regression testing
3. Set up Android Espresso/Maestro framework
4. Create comprehensive test suite for mobile apps
5. Integrate E2E tests into CI/CD pipeline

**Testing Philosophy**: Comprehensive E2E testing ensures production readiness by validating complete user workflows across all platforms.

---

**Last Updated**: 2026-02-10
**Status**: ⏳ Foundation Complete, Expansion Planned
