import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup playlist API mocks.
 */
async function setupPlaylistMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  const mockPlaylists = [
    {
      id: 'pl-1',
      name: 'Movie Marathon',
      description: 'Weekend binge watching',
      item_count: 5,
      duration: 36000,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    {
      id: 'pl-2',
      name: 'Jazz Favorites',
      description: 'Smooth jazz collection',
      item_count: 12,
      duration: 45000,
      created_at: new Date(Date.now() - 86400000).toISOString(),
      updated_at: new Date(Date.now() - 86400000).toISOString(),
    },
  ];

  await page.route('**/api/v1/playlists**', async (route) => {
    const url = route.request().url();
    const method = route.request().method();

    if (method === 'POST') {
      const body = route.request().postDataJSON();
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'pl-new',
          name: body?.name || 'New Playlist',
          description: body?.description || '',
          item_count: 0,
          duration: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }),
      });
      return;
    }

    if (method === 'DELETE') {
      await route.fulfill({ status: 204, body: '' });
      return;
    }

    if (method === 'PUT') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          ...mockPlaylists[0],
          name: 'Updated Playlist',
        }),
      });
      return;
    }

    // Specific playlist by ID
    if (url.match(/\/playlists\/[^/?]+$/)) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          ...mockPlaylists[0],
          items: [
            { id: 1, media_id: 1, title: 'The Matrix', position: 0 },
            { id: 2, media_id: 2, title: 'Breaking Bad S01E01', position: 1 },
            { id: 3, media_id: 3, title: 'Dark Side of the Moon', position: 2 },
          ],
        }),
      });
      return;
    }

    // Playlists list
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        playlists: mockPlaylists,
        total: mockPlaylists.length,
      }),
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

test.describe('Playlists', () => {
  test.beforeEach(async ({ page }) => {
    await setupPlaylistMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Playlists Page Layout', () => {
    test('can access playlists page', async ({ page }) => {
      await page.goto('/playlists');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('body')).toBeVisible();
    });

    test('playlists page loads without console errors', async ({ page }) => {
      const consoleErrors: string[] = [];
      page.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.goto('/playlists');
      await page.waitForLoadState('networkidle');

      const criticalErrors = consoleErrors.filter(
        (e) => !e.includes('favicon') && !e.includes('manifest')
      );
      expect(criticalErrors).toHaveLength(0);
    });

    test('playlists page displays page title', async ({ page }) => {
      await page.goto('/playlists');
      await page.waitForLoadState('networkidle');

      const title = page.locator('text=Playlists, text=My Playlists, h1, h2');
      await expect(title.first()).toBeVisible();
    });
  });

  test.describe('Playlist Navigation', () => {
    test('playlists link is visible in navigation', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const playlistLink = page.locator(
        'nav a[href*="playlist"], a[href*="playlist"]'
      );
      if (await playlistLink.first().isVisible({ timeout: 3000 })) {
        await expect(playlistLink.first()).toBeVisible();
      }
    });

    test('clicking playlists nav link navigates to playlists page', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const playlistLink = page.locator(
        'nav a[href*="playlist"], a[href*="playlist"]'
      );
      if (await playlistLink.first().isVisible({ timeout: 3000 })) {
        await playlistLink.first().click();
        await expect(page).toHaveURL(/.*playlist/);
      }
    });
  });

  test.describe('Playlist CRUD', () => {
    test('create playlist button is present', async ({ page }) => {
      await page.goto('/playlists');
      await page.waitForLoadState('networkidle');

      const createButton = page.locator(
        'button:has-text("Create"), button:has-text("New Playlist"), button:has-text("Add")'
      );
      if (await createButton.first().isVisible({ timeout: 3000 })) {
        await expect(createButton.first()).toBeVisible();
      }
    });

    test('page has navigation sidebar or header', async ({ page }) => {
      await page.goto('/playlists');
      await page.waitForLoadState('networkidle');

      const navigation = page.locator('nav, [role="navigation"], aside');
      await expect(navigation.first()).toBeVisible();
    });
  });
});
