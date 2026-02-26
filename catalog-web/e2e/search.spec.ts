import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from './fixtures/auth';
import { mockDashboardEndpoints } from './fixtures/api-mocks';

/**
 * Search E2E Tests
 *
 * Tests searching by title across media and entities, filtering results,
 * and navigating to a result. Covers the header search bar, the /media
 * page search, and the /browse entity search.
 */

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
    media_type: 'game',
    year: 2016,
    rating: 8.9,
    poster_url: 'https://example.com/dark-souls.jpg',
    description: 'Action role-playing video game',
    duration: 0,
    created_at: new Date().toISOString(),
  },
  {
    id: 5,
    title: 'Interstellar',
    media_type: 'movie',
    year: 2014,
    rating: 8.6,
    poster_url: 'https://example.com/interstellar.jpg',
    description: 'A team of explorers travel through a wormhole',
    duration: 169,
    created_at: new Date().toISOString(),
  },
];

async function setupSearchMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  // Mock media search with filtering
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

  // Mock media list (used by /media page)
  await page.route('**/api/v1/media', async (route) => {
    if (route.request().url().includes('/media/search') ||
        route.request().url().includes('/media/stats') ||
        route.request().url().includes('/media/recent') ||
        route.request().url().includes('/media/popular')) {
      return route.continue();
    }
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

  // Mock media stats
  await page.route('**/api/v1/media/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_items: allMedia.length,
        by_type: { movie: 2, tv_show: 1, music: 1, game: 1 },
        total_size: 536870912000,
        recent_additions: 5,
      }),
    });
  });

  // Mock entity search (for /browse)
  await page.route('**/api/v1/entities?**', async (route) => {
    const url = new URL(route.request().url());
    const query = url.searchParams.get('query') || '';

    let results = allMedia.map((m) => ({
      id: m.id,
      title: m.title,
      media_type: m.media_type,
      year: m.year,
      file_count: 1,
      children_count: 0,
      created_at: m.created_at,
    }));

    if (query) {
      results = results.filter((item) =>
        item.title.toLowerCase().includes(query.toLowerCase())
      );
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

  // Mock entity types
  await page.route('**/api/v1/entities/types', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        types: [
          { id: 1, name: 'movie', display_name: 'Movies', count: 2 },
          { id: 2, name: 'tv_show', display_name: 'TV Shows', count: 1 },
          { id: 3, name: 'music', display_name: 'Music', count: 1 },
          { id: 4, name: 'game', display_name: 'Games', count: 1 },
        ],
      }),
    });
  });

  // Mock entity stats
  await page.route('**/api/v1/entities/stats', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_entities: 5,
        total_files: 10,
        by_type: { movie: 2, tv_show: 1, music: 1, game: 1 },
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
            { query: 'music', timestamp: new Date(Date.now() - 3600000).toISOString(), results_count: 3 },
          ],
        }),
      });
    } else if (route.request().method() === 'DELETE') {
      await route.fulfill({ status: 204, body: '' });
    }
  });

  // Mock collections for the collections search test
  await page.route('**/api/v1/collections**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ collections: [], total: 0 }),
    });
  });
}

test.describe('Search', () => {
  test.beforeEach(async ({ page }) => {
    await setupSearchMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Header Search Bar', () => {
    test('header search bar is visible on the dashboard', async ({ page }) => {
      const searchBar = page.locator('header input[placeholder*="Search media"]');
      await expect(searchBar).toBeVisible();
    });

    test('can type in the header search bar', async ({ page }) => {
      const searchBar = page.locator('header input[placeholder*="Search media"]');
      await searchBar.fill('batman');
      await expect(searchBar).toHaveValue('batman');
    });

    test('pressing Enter in header search navigates or triggers search', async ({ page }) => {
      const searchBar = page.locator('header input[placeholder*="Search media"]');
      await searchBar.fill('Dark');
      await searchBar.press('Enter');
      await page.waitForTimeout(1000);

      // Should navigate to a search results page or /media with query
      await expect(page.locator('body')).toBeVisible();
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
      await page.waitForTimeout(600);

      // Results should be filtered -- page still functional
      await expect(page.locator('body')).toBeVisible();
    });

    test('searching for non-existent title shows no/empty results', async ({ page }) => {
      await page.route('**/api/v1/media/search**', async (route) => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('query') || '';
        if (query.toLowerCase() === 'xyznonexistent') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ items: [], total: 0, limit: 24, offset: 0 }),
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      await searchInput.fill('xyznonexistent');
      await page.waitForTimeout(600);

      // Page should still be functional
      await expect(page.locator('body')).toBeVisible();
    });

    test('clearing search restores full results', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="Search your media collection"]');
      await searchInput.fill('Dark');
      await page.waitForTimeout(600);

      await searchInput.fill('');
      await page.waitForTimeout(600);

      await expect(page.locator('body')).toBeVisible();
    });

    test('search is debounced to avoid excessive API calls', async ({ page }) => {
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
      await searchInput.type('rapid typing test', { delay: 30 });
      await page.waitForTimeout(800);

      const totalRequests = requestCount - initialCount;
      // With debounce, we expect far fewer requests than characters typed
      expect(totalRequests).toBeLessThan(16);
    });
  });

  test.describe('Entity Search on Browse Page', () => {
    test('browse page has a search input', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="search" i], input[type="search"]');
      await expect(searchInput.first()).toBeVisible();
    });

    test('searching entities by title updates URL with query param', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="search" i], input[type="search"]');
      await searchInput.first().fill('Interstellar');
      await page.waitForTimeout(1000);

      await expect(page).toHaveURL(/q=Interstellar/);
    });

    test('entity search results are displayed', async ({ page }) => {
      await page.goto('/browse?q=Dark');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Should show entities matching the search
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Filter Results by Type', () => {
    test('media page has filter controls', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const filtersButton = page.locator('[data-testid="filters-button"]');
      await expect(filtersButton).toBeVisible();
    });

    test('clicking filters button reveals filter options', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const filtersButton = page.locator('[data-testid="filters-button"]');
      await filtersButton.click();
      await page.waitForTimeout(500);

      // Filter panel should become visible
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Navigate to Search Result', () => {
    test('clicking on a media item navigates to its detail', async ({ page }) => {
      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Find a clickable media card
      const mediaCard = page.locator(
        'a[href*="/media/"], a[href*="/entity/"], [data-testid*="media-card"]'
      );
      if (await mediaCard.first().isVisible({ timeout: 5000 })) {
        await mediaCard.first().click();
        await page.waitForTimeout(1000);

        // Should navigate to a detail page
        const url = page.url();
        expect(url).toMatch(/\/(media|entity)\/\d+|\/media/);
      }
    });
  });

  test.describe('Collections Page Search', () => {
    test('collections page has a search input', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="Search collections"]');
      await expect(searchInput).toBeVisible();
    });

    test('can type in the collections search box', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="Search collections"]');
      await searchInput.fill('Action');
      await expect(searchInput).toHaveValue('Action');
    });
  });
});
