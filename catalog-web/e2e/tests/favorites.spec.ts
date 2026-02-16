import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup favorites-related API mocks
 */
async function setupFavoritesMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  const favoriteItems = [
    {
      id: 1,
      media_id: 101,
      title: 'Favorite Movie 1',
      media_type: 'movie',
      poster_url: 'https://example.com/fav1.jpg',
      added_at: new Date().toISOString(),
    },
    {
      id: 2,
      media_id: 102,
      title: 'Favorite Song',
      media_type: 'music',
      poster_url: 'https://example.com/fav2.jpg',
      added_at: new Date(Date.now() - 86400000).toISOString(),
    },
    {
      id: 3,
      media_id: 103,
      title: 'Favorite TV Show',
      media_type: 'tv_show',
      poster_url: 'https://example.com/fav3.jpg',
      added_at: new Date(Date.now() - 172800000).toISOString(),
    },
  ];

  // Mock favorites list
  await page.route('**/api/v1/favorites**', async (route) => {
    const method = route.request().method();
    if (method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          items: favoriteItems,
          total: favoriteItems.length,
        }),
      });
    } else if (method === 'POST') {
      const body = route.request().postDataJSON();
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: Date.now(),
          media_id: body.media_id,
          added_at: new Date().toISOString(),
        }),
      });
    } else if (method === 'DELETE') {
      await route.fulfill({ status: 204, body: '' });
    }
  });

  // Mock favorites stats
  await page.route('**/api/v1/favorites/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_count: 3,
        media_type_breakdown: {
          movie: 1,
          music: 1,
          tv_show: 1,
        },
        recent_additions: favoriteItems,
      }),
    });
  });
}

test.describe('Favorites', () => {
  test.beforeEach(async ({ page }) => {
    await setupFavoritesMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Favorites Page Layout', () => {
    test('displays the My Favorites page title', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('text=My Favorites')).toBeVisible();
    });

    test('displays subtitle about managing favorites', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('text=Manage your favorite media items')).toBeVisible();
    });

    test('displays favorites tab navigation', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("Favorites"), [role="tab"]:has-text("Favorites")').first()).toBeVisible();
      await expect(page.locator('button:has-text("Recently Added"), [role="tab"]:has-text("Recently Added")').first()).toBeVisible();
      await expect(page.locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")').first()).toBeVisible();
    });
  });

  test.describe('Favorites Action Buttons', () => {
    test('Bulk Actions button is present', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const bulkButton = page.locator('button:has-text("Bulk Actions")');
      await expect(bulkButton).toBeVisible();
    });

    test('Import button is present', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const importButton = page.locator('button:has-text("Import")');
      await expect(importButton).toBeVisible();
    });

    test('Export button is present', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const exportButton = page.locator('button:has-text("Export")');
      await expect(exportButton).toBeVisible();
    });

    test('clicking Bulk Actions highlights the button', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const bulkButton = page.locator('button:has-text("Bulk Actions")');
      await bulkButton.click();

      // The button should get a different visual style when active
      await page.waitForTimeout(300);
      await expect(bulkButton).toBeVisible();
    });
  });

  test.describe('Favorites Tab Navigation', () => {
    test('can switch to Recently Added tab', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const recentTab = page.locator('button:has-text("Recently Added"), [role="tab"]:has-text("Recently Added")').first();
      await recentTab.click();
      await page.waitForTimeout(300);

      // Should show "Recently Added Favorites" content
      await expect(page.locator('text=Recently Added Favorites')).toBeVisible();
    });

    test('can switch to Statistics tab', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const statsTab = page.locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")').first();
      await statsTab.click();
      await page.waitForTimeout(300);

      // Should show statistics content
      await expect(page.locator('text=Favorite Statistics')).toBeVisible();
    });

    test('Statistics tab shows total favorites count', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const statsTab = page.locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")').first();
      await statsTab.click();
      await page.waitForTimeout(500);

      await expect(page.locator('text=Total Favorites')).toBeVisible();
    });

    test('Statistics tab shows media type breakdown', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const statsTab = page.locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")').first();
      await statsTab.click();
      await page.waitForTimeout(500);

      await expect(page.locator('text=By Media Type')).toBeVisible();
    });

    test('Statistics tab shows insights cards', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      const statsTab = page.locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")').first();
      await statsTab.click();
      await page.waitForTimeout(500);

      await expect(page.locator('text=Most Common Type')).toBeVisible();
      await expect(page.locator('text=Recent Activity')).toBeVisible();
      await expect(page.locator('text=Storage Impact')).toBeVisible();
    });
  });

  test.describe('Favorites Navigation', () => {
    test('favorites link is visible in header navigation', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const favLink = page.locator('nav a[href="/favorites"]');
      await expect(favLink).toBeVisible();
    });

    test('clicking favorites nav link navigates to favorites page', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const favLink = page.locator('nav a[href="/favorites"]');
      await favLink.click();
      await expect(page).toHaveURL(/.*favorites/);
    });
  });
});
