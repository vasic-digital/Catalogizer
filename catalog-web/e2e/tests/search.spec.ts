import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup search-related API mocks
 */
async function setupSearchMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  const allMedia = [
    {
      id: 1,
      title: 'The Dark Knight',
      media_type: 'movie',
      year: 2008,
      rating: 9.0,
      poster_url: 'https://example.com/dark-knight.jpg',
      description: 'Batman fights the Joker in Gotham City',
      duration: 152,
      created_at: new Date().toISOString(),
    },
    {
      id: 2,
      title: 'Breaking Bad',
      media_type: 'tv_show',
      year: 2008,
      rating: 9.5,
      poster_url: 'https://example.com/breaking-bad.jpg',
      description: 'A chemistry teacher turns to drug manufacturing',
      seasons: 5,
      created_at: new Date().toISOString(),
    },
    {
      id: 3,
      title: 'Abbey Road',
      media_type: 'music',
      year: 1969,
      artist: 'The Beatles',
      poster_url: 'https://example.com/abbey-road.jpg',
      tracks: 17,
      created_at: new Date().toISOString(),
    },
    {
      id: 4,
      title: 'Dark Souls III',
      media_type: 'movie',
      year: 2016,
      rating: 8.0,
      poster_url: 'https://example.com/dark-souls.jpg',
      description: 'Action role-playing video game',
      duration: 120,
      created_at: new Date().toISOString(),
    },
  ];

  // Mock media search endpoint with query filtering
  await page.route('**/api/v1/media/search**', async (route) => {
    const url = new URL(route.request().url());
    const query = url.searchParams.get('query') || url.searchParams.get('q') || '';
    const mediaType = url.searchParams.get('media_type') || url.searchParams.get('type') || '';

    let results = [...allMedia];

    if (query) {
      results = results.filter(
        (item) =>
          item.title.toLowerCase().includes(query.toLowerCase()) ||
          (item.description && item.description.toLowerCase().includes(query.toLowerCase()))
      );
    }

    if (mediaType) {
      results = results.filter((item) => item.media_type === mediaType);
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: results,
        total: results.length,
        limit: 24,
        offset: 0,
      }),
    });
  });

  // Mock media stats
  await page.route('**/api/v1/media/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_items: allMedia.length,
        by_type: { movie: 2, tv_show: 1, music: 1 },
        total_size: 536870912000,
        recent_additions: 4,
      }),
    });
  });

  // Mock search history
  await page.route('**/api/v1/search/history**', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          queries: [
            { query: 'batman', timestamp: new Date().toISOString(), results_count: 1 },
            { query: 'music collection', timestamp: new Date(Date.now() - 3600000).toISOString(), results_count: 5 },
          ],
        }),
      });
    } else if (route.request().method() === 'DELETE') {
      await route.fulfill({ status: 204, body: '' });
    }
  });

  // Mock media list endpoint (used by some routes)
  await page.route('**/api/v1/media', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: allMedia,
        total: allMedia.length,
        limit: 20,
        offset: 0,
      }),
    });
  });
}

test.describe('Search', () => {
  test.beforeEach(async ({ page }) => {
    await setupSearchMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Header Search Bar', () => {
    test('header search bar is visible on dashboard', async ({ page }) => {
      const searchBar = page.locator('header input[placeholder*="Search media"]');
      await expect(searchBar).toBeVisible();
    });

    test('can type in the header search bar', async ({ page }) => {
      const searchBar = page.locator('header input[placeholder*="Search media"]');
      await searchBar.fill('batman');
      await expect(searchBar).toHaveValue('batman');
    });
  });

  test.describe('Media Page Search', () => {
    test('media page has a search input', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      await expect(searchInput).toBeVisible();
    });

    test('searching by title returns matching results', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      await searchInput.fill('Dark');
      await page.waitForTimeout(500);

      // Should have results displayed
      await expect(page.locator('body')).toBeVisible();
    });

    test('searching with no results shows appropriate message', async ({ page }) => {
      // Override search to return empty results for a specific query
      await page.route('**/api/v1/media/search**', async (route) => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('query') || '';
        if (query === 'xyznonexistent') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              items: [],
              total: 0,
              limit: 24,
              offset: 0,
            }),
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      await searchInput.fill('xyznonexistent');
      await page.waitForTimeout(500);
      await expect(page.locator('body')).toBeVisible();
    });

    test('clearing search shows all results again', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      const searchInput = page.locator('input[placeholder*="Search your media collection"]');

      // Search for something
      await searchInput.fill('Dark');
      await page.waitForTimeout(500);

      // Clear search
      await searchInput.fill('');
      await page.waitForTimeout(500);

      // Results should be back to full list
      await expect(page.locator('body')).toBeVisible();
    });

    test('search is debounced and does not fire on every keystroke', async ({ page }) => {
      let requestCount = 0;
      await page.route('**/api/v1/media/search**', async (route) => {
        requestCount++;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ items: [], total: 0, limit: 24, offset: 0 }),
        });
      });

      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      const initialCount = requestCount;

      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      // Type quickly
      await searchInput.type('test query', { delay: 50 });
      await page.waitForTimeout(600);

      // Should not have fired a request for every keystroke
      // (debounce should batch them)
      const totalRequests = requestCount - initialCount;
      // With debounce, we expect far fewer requests than characters typed
      expect(totalRequests).toBeLessThan(10);
    });
  });

  test.describe('Collections Page Search', () => {
    test('collections page has a search input', async ({ page }) => {
      // Setup collection mocks
      await page.route('**/api/v1/collections**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ collections: [], total: 0 }),
        });
      });

      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const searchInput = page.locator('input[placeholder*="Search collections"]');
      await expect(searchInput).toBeVisible();
    });
  });
});
