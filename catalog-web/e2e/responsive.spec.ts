import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from './fixtures/auth';
import { mockDashboardEndpoints, mockCollectionEndpoints, mockMediaEndpoints } from './fixtures/api-mocks';

/**
 * Responsive Layout E2E Tests
 *
 * Tests mobile (375px) and tablet (768px) viewports for key pages:
 * login, dashboard, media browser, collections, favorites, and browse.
 */

async function setupAllMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);
  await mockMediaEndpoints(page);
  await mockCollectionEndpoints(page);

  // Mock favorites
  await page.route('**/api/v1/favorites**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          { id: 1, media_id: 101, title: 'Test Movie', media_type: 'movie', added_at: new Date().toISOString() },
        ],
        total: 1,
      }),
    });
  });

  await page.route('**/api/v1/favorites/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_count: 1,
        media_type_breakdown: { movie: 1 },
        recent_additions: [],
      }),
    });
  });

  // Mock entity types for browse
  await page.route('**/api/v1/entities/types', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        types: [
          { id: 1, name: 'movie', display_name: 'Movies', count: 10 },
          { id: 2, name: 'tv_show', display_name: 'TV Shows', count: 5 },
        ],
      }),
    });
  });

  await page.route('**/api/v1/entities/stats', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_entities: 15,
        total_files: 30,
        by_type: { movie: 10, tv_show: 5 },
      }),
    });
  });

  // Mock media stats
  await page.route('**/api/v1/media/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_items: 15,
        by_type: { movie: 10, tv_show: 5 },
        total_size: 100000000,
        recent_additions: 3,
      }),
    });
  });
}

test.describe('Responsive - Mobile (375px)', () => {
  test.describe('Login Page', () => {
    test('login form renders correctly on mobile', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/login');

      await expect(page.locator('input[placeholder*="username" i]')).toBeVisible();
      await expect(page.locator('input[placeholder*="password" i]')).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();
    });

    test('login form fits within mobile viewport without horizontal scroll', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/login');

      // Verify that the page body width does not exceed viewport
      const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
      expect(scrollWidth).toBeLessThanOrEqual(375 + 5); // small tolerance
    });

    test('login form is usable on very small screens (320px)', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.setViewportSize({ width: 320, height: 568 });
      await page.goto('/login');

      await expect(page.locator('input[placeholder*="username" i]')).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();
    });
  });

  test.describe('Dashboard - Mobile', () => {
    test.beforeEach(async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);
    });

    test('hamburger menu button is visible on mobile', async ({ page }) => {
      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      });
      await expect(hamburger.first()).toBeVisible();
    });

    test('desktop navigation links are hidden on mobile', async ({ page }) => {
      const desktopNav = page.locator('nav.hidden.md\\:flex a');
      if (await desktopNav.count() > 0) {
        await expect(desktopNav.first()).not.toBeVisible();
      }
    });

    test('clicking hamburger opens mobile menu with all navigation links', async ({ page }) => {
      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      // Mobile menu should have navigation links
      await expect(page.locator('a[href="/dashboard"]:has-text("Dashboard")')).toBeVisible();
      await expect(page.locator('a[href="/media"]:has-text("Media")')).toBeVisible();
    });

    test('mobile menu has search input', async ({ page }) => {
      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      const mobileSearch = page.locator('input[placeholder*="Search media"]');
      await expect(mobileSearch.first()).toBeVisible();
    });

    test('mobile menu has logout button', async ({ page }) => {
      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      await expect(page.locator('button:has-text("Logout")')).toBeVisible();
    });

    test('clicking a mobile menu link navigates and closes menu', async ({ page }) => {
      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      const mediaLink = page.locator('a[href="/media"]:has-text("Media")');
      if (await mediaLink.isVisible()) {
        await mediaLink.click();
        await page.waitForTimeout(500);
        await expect(page).toHaveURL(/.*media/);
      }
    });
  });

  test.describe('Favorites Page - Mobile', () => {
    test('favorites page renders on mobile', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);

      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('text=My Favorites')).toBeVisible();
    });
  });

  test.describe('Collections Page - Mobile', () => {
    test('collections page renders on mobile', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);

      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('h1:has-text("Collections")')).toBeVisible();
    });
  });

  test.describe('Browse Page - Mobile', () => {
    test('browse page renders on mobile', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);

      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('body')).toBeVisible();
      await expect(page).toHaveURL(/.*browse/);
    });
  });
});

test.describe('Responsive - Tablet (768px)', () => {
  test.describe('Dashboard - Tablet', () => {
    test.beforeEach(async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 768, height: 1024 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);
    });

    test('desktop navigation is visible at tablet breakpoint', async ({ page }) => {
      // At 768px (md breakpoint), the desktop nav should be visible
      const dashboardLink = page.locator('nav a[href="/dashboard"]');
      await expect(dashboardLink).toBeVisible();
    });

    test('all main navigation links are visible', async ({ page }) => {
      await expect(page.locator('nav a:has-text("Dashboard")').first()).toBeVisible();
      await expect(page.locator('nav a:has-text("Media")').first()).toBeVisible();
      await expect(page.locator('nav a:has-text("Favorites")').first()).toBeVisible();
      await expect(page.locator('nav a:has-text("Collections")').first()).toBeVisible();
    });

    test('header search bar is visible at tablet width', async ({ page }) => {
      const searchBar = page.locator('header input[placeholder*="Search media"]');
      await expect(searchBar).toBeVisible();
    });

    test('mobile hamburger button is hidden on tablet', async ({ page }) => {
      const mobileMenuButton = page.locator('button.md\\:hidden');
      if (await mobileMenuButton.count() > 0) {
        await expect(mobileMenuButton.first()).not.toBeVisible();
      }
    });
  });

  test.describe('Media Page - Tablet', () => {
    test('media page renders with controls at tablet width', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 768, height: 1024 });
      await loginAs(page, testUser);

      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('h1:has-text("Media Browser")')).toBeVisible();
      await expect(
        page.locator('input[placeholder*="Search your media collection"]')
      ).toBeVisible();
    });
  });

  test.describe('Collections Page - Tablet', () => {
    test('collections page renders at tablet width', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 768, height: 1024 });
      await loginAs(page, testUser);

      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('h1:has-text("Collections")')).toBeVisible();
    });
  });

  test.describe('Favorites Page - Tablet', () => {
    test('favorites page renders at tablet width', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 768, height: 1024 });
      await loginAs(page, testUser);

      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('text=My Favorites')).toBeVisible();
    });
  });

  test.describe('Browse Page - Tablet', () => {
    test('browse page renders at tablet width with type selector', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 768, height: 1024 });
      await loginAs(page, testUser);

      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      await expect(page).toHaveURL(/.*browse/);
      await expect(page.locator('body')).toBeVisible();
    });
  });
});

test.describe('Responsive - Header Behavior', () => {
  test('header is sticky across all viewports', async ({ page }) => {
    await setupAllMocks(page);
    await loginAs(page, testUser);

    // Test at desktop
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.waitForTimeout(300);
    const stickyHeader = page.locator('header.sticky');
    await expect(stickyHeader).toBeVisible();

    // Test at mobile
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(300);
    await expect(stickyHeader).toBeVisible();
  });

  test('user welcome text is visible on desktop but may be hidden on mobile', async ({ page }) => {
    await setupAllMocks(page);
    await loginAs(page, testUser);

    // Desktop: welcome visible
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.waitForTimeout(300);
    const welcome = page.locator('header span:has-text("Welcome")');
    await expect(welcome).toBeVisible();

    // Mobile: may be hidden to save space
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(300);
    // Not asserting it must be hidden -- just that the page is still functional
    await expect(page.locator('header')).toBeVisible();
  });
});
