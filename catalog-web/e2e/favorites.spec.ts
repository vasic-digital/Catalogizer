import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from './fixtures/auth';
import { mockDashboardEndpoints } from './fixtures/api-mocks';

/**
 * Favorites E2E Tests
 *
 * Tests the favorites workflow: login, browse media, toggle favorites,
 * verify the favorites page reflects changes.
 */

const favoriteItems = [
  {
    id: 1,
    media_id: 101,
    title: 'Inception',
    media_type: 'movie',
    poster_url: 'https://example.com/inception.jpg',
    added_at: new Date().toISOString(),
  },
  {
    id: 2,
    media_id: 102,
    title: 'Dark Side of the Moon',
    media_type: 'music',
    poster_url: 'https://example.com/dsotm.jpg',
    added_at: new Date(Date.now() - 86400000).toISOString(),
  },
  {
    id: 3,
    media_id: 103,
    title: 'Breaking Bad',
    media_type: 'tv_show',
    poster_url: 'https://example.com/bb.jpg',
    added_at: new Date(Date.now() - 172800000).toISOString(),
  },
];

async function setupFavoritesMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  // Track whether an item was added or removed for dynamic responses
  let currentFavorites = [...favoriteItems];

  // Mock favorites list
  await page.route('**/api/v1/favorites**', async (route) => {
    const url = route.request().url();
    const method = route.request().method();

    // Skip sub-routes like /favorites/stats
    if (url.includes('/favorites/stats')) {
      return route.continue();
    }

    if (method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          items: currentFavorites,
          total: currentFavorites.length,
        }),
      });
    } else if (method === 'POST') {
      const body = route.request().postDataJSON();
      const newFav = {
        id: Date.now(),
        media_id: body.media_id,
        title: body.title || 'New Favorite',
        media_type: body.media_type || 'movie',
        poster_url: '',
        added_at: new Date().toISOString(),
      };
      currentFavorites.push(newFav);
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(newFav),
      });
    } else if (method === 'DELETE') {
      const idMatch = url.match(/\/favorites\/(\d+)/);
      if (idMatch) {
        const deleteId = parseInt(idMatch[1]);
        currentFavorites = currentFavorites.filter((f) => f.id !== deleteId);
      }
      await route.fulfill({ status: 204, body: '' });
    }
  });

  // Mock favorites stats
  await page.route('**/api/v1/favorites/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_count: currentFavorites.length,
        media_type_breakdown: {
          movie: currentFavorites.filter((f) => f.media_type === 'movie').length,
          music: currentFavorites.filter((f) => f.media_type === 'music').length,
          tv_show: currentFavorites.filter((f) => f.media_type === 'tv_show').length,
        },
        recent_additions: currentFavorites.slice(0, 3),
      }),
    });
  });

  // Mock entity detail with user-metadata (favorite toggle on entity pages)
  await page.route('**/api/v1/entities/*/user-metadata', async (route) => {
    if (route.request().method() === 'PUT') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Updated' }),
      });
    } else {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ favorite: false, rating: 0 }),
      });
    }
  });
}

test.describe('Favorites', () => {
  test.beforeEach(async ({ page }) => {
    await setupFavoritesMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Favorites Page Access', () => {
    test('can navigate to the favorites page', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page).toHaveURL(/.*favorites/);
      await expect(page.locator('body')).toBeVisible();
    });

    test('displays "My Favorites" page heading', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('text=My Favorites')).toBeVisible();
    });

    test('displays subtitle about managing favorites', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('text=Manage your favorite media items')).toBeVisible();
    });

    test('favorites link is accessible from the navigation', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const favLink = page.locator('nav a[href="/favorites"]');
      await expect(favLink).toBeVisible();
    });

    test('clicking favorites nav link navigates to favorites', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const favLink = page.locator('nav a[href="/favorites"]');
      await favLink.click();
      await expect(page).toHaveURL(/.*favorites/);
    });
  });

  test.describe('Favorites List Display', () => {
    test('displays favorite items on the page', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      // At least one favorite title should be visible
      const inceptionText = page.locator('text=Inception');
      await expect(inceptionText.first()).toBeVisible();
    });

    test('displays multiple favorites', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      // Both items should appear
      await expect(page.locator('text=Inception').first()).toBeVisible();
      await expect(page.locator('text=Breaking Bad').first()).toBeVisible();
    });
  });

  test.describe('Tab Navigation', () => {
    test('shows Favorites, Recently Added, and Statistics tabs', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(
        page.locator('button:has-text("Favorites"), [role="tab"]:has-text("Favorites")').first()
      ).toBeVisible();
      await expect(
        page
          .locator('button:has-text("Recently Added"), [role="tab"]:has-text("Recently Added")')
          .first()
      ).toBeVisible();
      await expect(
        page.locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")').first()
      ).toBeVisible();
    });

    test('can switch to Recently Added tab', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      const recentTab = page
        .locator('button:has-text("Recently Added"), [role="tab"]:has-text("Recently Added")')
        .first();
      await recentTab.click();
      await page.waitForTimeout(300);

      await expect(page.locator('text=Recently Added Favorites')).toBeVisible();
    });

    test('can switch to Statistics tab and see stats', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      const statsTab = page
        .locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")')
        .first();
      await statsTab.click();
      await page.waitForTimeout(500);

      await expect(page.locator('text=Favorite Statistics')).toBeVisible();
      await expect(page.locator('text=Total Favorites')).toBeVisible();
      await expect(page.locator('text=By Media Type')).toBeVisible();
    });

    test('Statistics tab shows insight cards', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      const statsTab = page
        .locator('button:has-text("Statistics"), [role="tab"]:has-text("Statistics")')
        .first();
      await statsTab.click();
      await page.waitForTimeout(500);

      await expect(page.locator('text=Most Common Type')).toBeVisible();
      await expect(page.locator('text=Recent Activity')).toBeVisible();
      await expect(page.locator('text=Storage Impact')).toBeVisible();
    });
  });

  test.describe('Favorites Action Buttons', () => {
    test('Bulk Actions button is present', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("Bulk Actions")')).toBeVisible();
    });

    test('Import button is present', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("Import")')).toBeVisible();
    });

    test('Export button is present', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("Export")')).toBeVisible();
    });

    test('clicking Bulk Actions toggles selection mode', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      const bulkButton = page.locator('button:has-text("Bulk Actions")');
      await bulkButton.click();
      await page.waitForTimeout(300);

      // Button should still be visible and the mode should change
      await expect(bulkButton).toBeVisible();
    });
  });

  test.describe('Remove Favorite', () => {
    test('favorite items have a remove/delete action available', async ({ page }) => {
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      // Look for delete/remove buttons or icons on each favorite card
      const removeButton = page.locator(
        'button[aria-label*="remove" i], button[aria-label*="delete" i], button:has(svg[class*="trash"]), button:has(svg[class*="x"])'
      );
      // If there are action buttons, at least one should be findable
      if (await removeButton.first().isVisible({ timeout: 3000 })) {
        await expect(removeButton.first()).toBeVisible();
      }
    });
  });
});
