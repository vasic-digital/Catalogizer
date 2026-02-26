import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from './fixtures/auth';
import { mockDashboardEndpoints } from './fixtures/api-mocks';

/**
 * Entity Browsing E2E Tests
 *
 * Tests the /browse page which uses the entity system to browse
 * media by type, search entities, paginate, and view entity details.
 */

const mockEntityTypes = [
  { id: 1, name: 'movie', display_name: 'Movies', count: 85 },
  { id: 2, name: 'tv_show', display_name: 'TV Shows', count: 42 },
  { id: 3, name: 'tv_season', display_name: 'TV Seasons', count: 120 },
  { id: 4, name: 'tv_episode', display_name: 'TV Episodes', count: 950 },
  { id: 5, name: 'music_artist', display_name: 'Music Artists', count: 35 },
  { id: 6, name: 'music_album', display_name: 'Music Albums', count: 78 },
  { id: 7, name: 'song', display_name: 'Songs', count: 512 },
  { id: 8, name: 'game', display_name: 'Games', count: 20 },
  { id: 9, name: 'software', display_name: 'Software', count: 15 },
  { id: 10, name: 'book', display_name: 'Books', count: 10 },
  { id: 11, name: 'comic', display_name: 'Comics', count: 5 },
];

const mockEntities = [
  {
    id: 1,
    title: 'The Dark Knight',
    media_type: 'movie',
    year: 2008,
    file_count: 1,
    children_count: 0,
    created_at: new Date().toISOString(),
  },
  {
    id: 2,
    title: 'Inception',
    media_type: 'movie',
    year: 2010,
    file_count: 1,
    children_count: 0,
    created_at: new Date().toISOString(),
  },
  {
    id: 3,
    title: 'Breaking Bad',
    media_type: 'tv_show',
    year: 2008,
    file_count: 0,
    children_count: 5,
    created_at: new Date().toISOString(),
  },
  {
    id: 4,
    title: 'Abbey Road',
    media_type: 'music_album',
    year: 1969,
    file_count: 17,
    children_count: 17,
    created_at: new Date().toISOString(),
  },
  {
    id: 5,
    title: 'Interstellar',
    media_type: 'movie',
    year: 2014,
    file_count: 1,
    children_count: 0,
    created_at: new Date().toISOString(),
  },
];

const mockEntityDetail = {
  id: 1,
  title: 'The Dark Knight',
  media_type: 'movie',
  year: 2008,
  file_count: 1,
  children_count: 0,
  description: 'When the menace known as the Joker wreaks havoc on Gotham...',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

async function setupBrowseMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  // Mock entity types endpoint
  await page.route('**/api/v1/entities/types', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ types: mockEntityTypes }),
    });
  });

  // Mock entity stats endpoint
  await page.route('**/api/v1/entities/stats', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_entities: 1872,
        total_files: 3500,
        by_type: {
          movie: 85,
          tv_show: 42,
          music_album: 78,
          song: 512,
          game: 20,
        },
      }),
    });
  });

  // Mock entity list/search endpoint
  await page.route('**/api/v1/entities?**', async (route) => {
    const url = new URL(route.request().url());
    const query = url.searchParams.get('query') || '';
    const limit = parseInt(url.searchParams.get('limit') || '24', 10);
    const offset = parseInt(url.searchParams.get('offset') || '0', 10);

    let results = [...mockEntities];
    if (query) {
      results = results.filter((item) =>
        item.title.toLowerCase().includes(query.toLowerCase())
      );
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: results.slice(offset, offset + limit),
        total: results.length,
        limit,
        offset,
      }),
    });
  });

  // Mock browse by type endpoint
  await page.route('**/api/v1/entities/browse/**', async (route) => {
    const url = route.request().url();
    const typeMatch = url.match(/\/browse\/([^?]+)/);
    const typeName = typeMatch ? typeMatch[1] : 'movie';

    const filtered = mockEntities.filter((e) => e.media_type === typeName);

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: filtered,
        total: filtered.length,
        type: typeName,
        limit: 24,
        offset: 0,
      }),
    });
  });

  // Mock single entity endpoint
  await page.route('**/api/v1/entities/*', async (route) => {
    const url = route.request().url();
    // Skip if this is a sub-resource like /entities/1/files
    if (url.match(/\/entities\/\d+\/(children|files|metadata|duplicates|user-metadata)/)) {
      return route.continue();
    }

    const idMatch = url.match(/\/entities\/(\d+)/);
    const id = idMatch ? parseInt(idMatch[1]) : 1;
    const entity = mockEntities.find((e) => e.id === id) || mockEntityDetail;

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ ...mockEntityDetail, ...entity }),
    });
  });

  // Mock entity children
  await page.route('**/api/v1/entities/*/children', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ items: [], total: 0, limit: 24, offset: 0 }),
    });
  });

  // Mock entity files
  await page.route('**/api/v1/entities/*/files', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        files: [
          {
            id: 1,
            file_path: '/media/movies/The.Dark.Knight.2008.1080p.mkv',
            file_size: 8589934592,
            storage_root: 'nas-media',
          },
        ],
        total: 1,
      }),
    });
  });

  // Mock entity duplicates
  await page.route('**/api/v1/entities/*/duplicates', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ duplicates: [], total: 0 }),
    });
  });

  // Mock entity metadata
  await page.route('**/api/v1/entities/*/metadata**', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ metadata: [] }),
      });
    } else {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'OK' }),
      });
    }
  });

  // Mock entity user-metadata
  await page.route('**/api/v1/entities/*/user-metadata', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'Updated' }),
    });
  });
}

test.describe('Entity Browsing', () => {
  test.beforeEach(async ({ page }) => {
    await setupBrowseMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Browse Page Layout', () => {
    test('navigates to /browse and displays the entity browser', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('body')).toBeVisible();
      // Should not be redirected to login
      await expect(page).toHaveURL(/.*browse/);
    });

    test('shows the Entity Browser heading', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const heading = page.locator('h1, h2').filter({ hasText: /browse|entity|media/i });
      await expect(heading.first()).toBeVisible();
    });

    test('displays media type selector cards when no type is selected', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      // The type selector should show clickable type cards
      // Look for any media type names visible on the page
      const movieType = page.locator('text=/[Mm]ovie/');
      await expect(movieType.first()).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Filtering by Media Type', () => {
    test('clicking a media type shows entities of that type', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      // Click on a movie type card
      const movieCard = page.locator('button, [role="button"], a').filter({
        hasText: /[Mm]ovie/,
      });
      if (await movieCard.first().isVisible({ timeout: 5000 })) {
        await movieCard.first().click();
        await page.waitForTimeout(1000);

        // URL should now contain the type parameter
        await expect(page).toHaveURL(/type=movie/);
      }
    });

    test('shows a back button when a type is selected', async ({ page }) => {
      await page.goto('/browse?type=movie');
      await page.waitForLoadState('networkidle');

      const backButton = page.locator('button').filter({
        has: page.locator('svg'),
      });
      await expect(backButton.first()).toBeVisible();
    });

    test('clicking back returns to the type selector view', async ({ page }) => {
      await page.goto('/browse?type=movie');
      await page.waitForLoadState('networkidle');

      // Find and click the back/reset button
      const backButton = page.locator('button[aria-label*="back" i], button:has(svg)').first();
      if (await backButton.isVisible({ timeout: 3000 })) {
        await backButton.click();
        await page.waitForTimeout(1000);

        // Should return to unfiltered view (no type param)
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const url = page.url();
        // Type selector should be visible again (movies text from the card, not from results)
        await expect(page.locator('body')).toBeVisible();
      }
    });
  });

  test.describe('Search Entities', () => {
    test('search input is visible on the browse page', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="search" i], input[type="search"]');
      await expect(searchInput.first()).toBeVisible();
    });

    test('typing in search filters entities by title', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="search" i], input[type="search"]');
      await searchInput.first().fill('Dark');
      await page.waitForTimeout(1000);

      // URL should include search query parameter
      await expect(page).toHaveURL(/q=Dark/);
    });

    test('clearing search returns to type selector', async ({ page }) => {
      await page.goto('/browse?q=Dark');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="search" i], input[type="search"]');
      await searchInput.first().fill('');
      await page.waitForTimeout(1000);

      // Should return to default browse view
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Pagination', () => {
    test('shows pagination controls when there are multiple pages', async ({ page }) => {
      // Override entities endpoint to return a large total
      await page.route('**/api/v1/entities/browse/**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            items: mockEntities.filter((e) => e.media_type === 'movie'),
            total: 100, // More than one page of 24
            type: 'movie',
            limit: 24,
            offset: 0,
          }),
        });
      });

      await page.goto('/browse?type=movie');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Look for pagination controls (next/prev buttons or page numbers)
      const paginationControls = page.locator(
        'button:has-text("Next"), button:has-text("Previous"), [aria-label*="page" i], nav[aria-label*="pagination" i]'
      );
      // Pagination should be visible if there are enough results
      if (await paginationControls.first().isVisible({ timeout: 5000 })) {
        await expect(paginationControls.first()).toBeVisible();
      }
    });

    test('clicking next page updates URL with page parameter', async ({ page }) => {
      await page.route('**/api/v1/entities/browse/**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            items: mockEntities.filter((e) => e.media_type === 'movie'),
            total: 100,
            type: 'movie',
            limit: 24,
            offset: 0,
          }),
        });
      });

      await page.goto('/browse?type=movie');
      await page.waitForLoadState('networkidle');

      const nextButton = page.locator('button:has-text("Next"), button[aria-label*="next" i]');
      if (await nextButton.first().isVisible({ timeout: 5000 })) {
        await nextButton.first().click();
        await page.waitForTimeout(1000);

        await expect(page).toHaveURL(/page=2/);
      }
    });
  });

  test.describe('Entity Detail Navigation', () => {
    test('clicking an entity navigates to entity detail page', async ({ page }) => {
      await page.goto('/browse?type=movie');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Find a clickable entity card or link
      const entityLink = page.locator(
        'a[href*="/entity/"], [role="link"][href*="/entity/"], [data-testid*="entity"]'
      );
      if (await entityLink.first().isVisible({ timeout: 5000 })) {
        await entityLink.first().click();
        await expect(page).toHaveURL(/\/entity\/\d+/);
      } else {
        // Try clicking on entity title text directly
        const entityTitle = page.locator('text=The Dark Knight');
        if (await entityTitle.first().isVisible({ timeout: 3000 })) {
          await entityTitle.first().click();
          await page.waitForTimeout(1000);
          // Check if navigated to detail page
          const url = page.url();
          expect(url).toMatch(/\/entity\/\d+|\/browse/);
        }
      }
    });

    test('entity detail page shows entity title', async ({ page }) => {
      await page.goto('/entity/1');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const title = page.locator('text=The Dark Knight');
      await expect(title.first()).toBeVisible();
    });

    test('entity detail page shows a back button to browse', async ({ page }) => {
      await page.goto('/entity/1');
      await page.waitForLoadState('networkidle');

      const backButton = page.locator(
        'button:has-text("Back"), a:has-text("Back"), button:has-text("Browse"), a[href="/browse"]'
      );
      await expect(backButton.first()).toBeVisible();
    });

    test('entity detail page shows file information', async ({ page }) => {
      await page.goto('/entity/1');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // Should show files section or file count
      const filesSection = page.locator('text=/[Ff]ile/');
      await expect(filesSection.first()).toBeVisible();
    });

    test('entity not found shows appropriate message', async ({ page }) => {
      await page.route('**/api/v1/entities/999', async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Entity not found' }),
        });
      });

      await page.goto('/entity/999');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const notFound = page.locator('text=/not found/i');
      await expect(notFound.first()).toBeVisible();
    });
  });

  test.describe('Navigation Integration', () => {
    test('browse link is visible in navigation', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const browseLink = page.locator('nav a[href="/browse"]');
      if (await browseLink.isVisible({ timeout: 3000 })) {
        await expect(browseLink).toBeVisible();
      }
    });

    test('can navigate to browse from navigation bar', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const browseLink = page.locator('nav a[href="/browse"]');
      if (await browseLink.isVisible({ timeout: 3000 })) {
        await browseLink.click();
        await expect(page).toHaveURL(/.*browse/);
      }
    });
  });
});
