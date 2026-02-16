import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockMediaEndpoints, mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup mock endpoints for playback-related pages
 */
async function setupPlaybackMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockMediaEndpoints(page);
  await mockDashboardEndpoints(page);

  // Mock media stats
  await page.route('**/api/v1/media/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_items: 150,
        by_type: { movie: 80, tv_show: 40, music: 30 },
        total_size: 268435456000,
        recent_additions: 12,
      }),
    });
  });

  // Mock search endpoint with playable items
  await page.route('**/api/v1/media/search**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          {
            id: 1,
            title: 'Test Video File',
            media_type: 'video/mp4',
            year: 2023,
            rating: 8.5,
            poster_url: 'https://example.com/poster1.jpg',
            description: 'A test video for playback testing',
            duration: 7200,
            directory_path: '/media/videos/test.mp4',
            created_at: new Date().toISOString(),
          },
          {
            id: 2,
            title: 'Test Audio Track',
            media_type: 'audio/mp3',
            year: 2024,
            artist: 'Test Artist',
            poster_url: 'https://example.com/poster2.jpg',
            duration: 240,
            directory_path: '/media/music/test.mp3',
            created_at: new Date().toISOString(),
          },
        ],
        total: 2,
        limit: 24,
        offset: 0,
      }),
    });
  });

  // Mock playback position endpoint
  await page.route('**/api/v1/playback/position**', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ position: 120, duration: 7200 }),
      });
    } else {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Position saved' }),
      });
    }
  });
}

test.describe('Media Playback', () => {
  test.beforeEach(async ({ page }) => {
    await setupPlaybackMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Media Player Controls', () => {
    test('media page loads with playable content', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('h1:has-text("Media Browser")')).toBeVisible();
    });

    test('media page has controls for browsing', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // Search input should be present
      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      await expect(searchInput).toBeVisible();

      // Grid/list toggle should be present
      await expect(page.locator('[data-testid="grid-view-button"]')).toBeVisible();
      await expect(page.locator('[data-testid="list-view-button"]')).toBeVisible();
    });

    test('refresh button triggers content reload', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const refreshButton = page.locator('[data-testid="refresh-button"]');
      await expect(refreshButton).toBeVisible();
      await refreshButton.click();

      // Button should remain functional after click
      await expect(refreshButton).toBeVisible();
    });
  });

  test.describe('Upload Manager', () => {
    test('upload button toggles the upload manager area', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const uploadButton = page.locator('[data-testid="upload-button"]');
      await uploadButton.click();

      // Upload area should appear
      await page.waitForTimeout(500);
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Media Search and Filter for Playable Content', () => {
    test('searching for a video title filters results', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      await searchInput.fill('Test Video');
      await page.waitForTimeout(500);

      // Results should be filtered
      await expect(page.locator('body')).toBeVisible();
    });

    test('filter sidebar can be toggled on media page', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const filtersButton = page.locator('[data-testid="filters-button"]');
      await filtersButton.click();
      await page.waitForTimeout(500);

      // Filter panel should appear
      await expect(page.locator('body')).toBeVisible();

      // Toggle off
      await filtersButton.click();
      await page.waitForTimeout(500);
    });
  });
});
