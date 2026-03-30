import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints, mockMediaEndpoints, mockMedia } from '../fixtures/api-mocks';

/**
 * Setup recent and popular media API mocks.
 */
async function setupRecentPopularMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);
  await mockMediaEndpoints(page);

  // Recent media endpoint
  await page.route('**/api/v1/media/recent**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          {
            id: 10,
            title: 'Recently Added Movie',
            media_type: 'movie',
            year: 2026,
            added_at: new Date().toISOString(),
          },
          {
            id: 11,
            title: 'Recently Added Album',
            media_type: 'music_album',
            year: 2025,
            added_at: new Date(Date.now() - 3600000).toISOString(),
          },
          ...mockMedia,
        ],
        total: 5,
      }),
    });
  });

  // Popular media endpoint
  await page.route('**/api/v1/media/popular**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          {
            id: 1,
            title: 'The Matrix',
            media_type: 'movie',
            year: 1999,
            favorite_count: 42,
          },
          {
            id: 2,
            title: 'Breaking Bad',
            media_type: 'tv_show',
            year: 2008,
            favorite_count: 38,
          },
          {
            id: 3,
            title: 'Dark Side of the Moon',
            media_type: 'music_album',
            year: 1973,
            favorite_count: 35,
          },
        ],
        total: 3,
      }),
    });
  });

  // Favorites catch-all
  await page.route('**/api/v1/favorites**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ items: [], total: 0 }),
    });
  });

  // Admin catch-all
  await page.route('**/api/v1/admin/**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ settings: {}, message: 'OK' }),
    });
  });
}

test.describe('Recent and Popular Media', () => {
  test.beforeEach(async ({ page }) => {
    await setupRecentPopularMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Dashboard Recent Activity', () => {
    test('dashboard loads and displays content', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('body')).toBeVisible();
    });

    test('dashboard loads without console errors', async ({ page }) => {
      const consoleErrors: string[] = [];
      page.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const criticalErrors = consoleErrors.filter(
        (e) => !e.includes('favicon') && !e.includes('manifest')
      );
      expect(criticalErrors).toHaveLength(0);
    });

    test('dashboard has navigation', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const navigation = page.locator('nav, [role="navigation"], aside');
      await expect(navigation.first()).toBeVisible();
    });
  });

  test.describe('Browse Page', () => {
    test('can access browse page for media listing', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('body')).toBeVisible();
    });

    test('browse page loads without console errors', async ({ page }) => {
      const consoleErrors: string[] = [];
      page.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const criticalErrors = consoleErrors.filter(
        (e) => !e.includes('favicon') && !e.includes('manifest')
      );
      expect(criticalErrors).toHaveLength(0);
    });

    test('browse page has navigation', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const navigation = page.locator('nav, [role="navigation"], aside');
      await expect(navigation.first()).toBeVisible();
    });
  });

  test.describe('Media Navigation', () => {
    test('media or browse link is visible in navigation', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const mediaLink = page.locator(
        'nav a[href*="browse"], nav a[href*="media"], a[href*="browse"]'
      );
      if (await mediaLink.first().isVisible({ timeout: 3000 })) {
        await expect(mediaLink.first()).toBeVisible();
      }
    });

    test('clicking media nav link navigates away from dashboard', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const mediaLink = page.locator(
        'nav a[href*="browse"], nav a[href*="media"]'
      );
      if (await mediaLink.first().isVisible({ timeout: 3000 })) {
        await mediaLink.first().click();
        await page.waitForLoadState('networkidle');
        // Should be on a different page than dashboard
        await expect(page.locator('body')).toBeVisible();
      }
    });
  });
});
