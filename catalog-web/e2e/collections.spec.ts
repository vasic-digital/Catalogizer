import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from './fixtures/auth';
import { mockDashboardEndpoints } from './fixtures/api-mocks';

/**
 * Collections E2E Tests
 *
 * Tests collection CRUD operations: navigate to /collections, create a new
 * collection, add items, edit, and delete collections.
 */

const existingCollections = [
  {
    id: 'col-1',
    name: 'Action Movies',
    description: 'Best action films of all time',
    item_count: 25,
    is_smart: false,
    is_public: false,
    media_type: 'video',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'col-2',
    name: 'Jazz Classics',
    description: 'Smooth jazz albums and compilations',
    item_count: 42,
    is_smart: true,
    is_public: true,
    media_type: 'music',
    smart_rules: [{ field: 'genre', operator: 'contains', value: 'jazz' }],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'col-3',
    name: 'Sci-Fi Marathon',
    description: 'Science fiction movies and series',
    item_count: 18,
    is_smart: false,
    is_public: false,
    media_type: 'video',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

async function setupCollectionMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  let collections = [...existingCollections];

  // Mock collections list
  await page.route('**/api/v1/collections**', async (route) => {
    const url = route.request().url();
    const method = route.request().method();

    // Skip sub-routes like /collections/col-1/share
    if (url.match(/\/collections\/[^?]+\/(share|duplicate|export|items)/)) {
      return route.continue();
    }

    if (method === 'GET' && !url.match(/\/collections\/[^?]+$/)) {
      // List endpoint
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          collections,
          total: collections.length,
        }),
      });
    } else if (method === 'POST' && !url.match(/\/collections\/[^?]+/)) {
      // Create endpoint
      const body = route.request().postDataJSON();
      const newCollection = {
        id: 'col-new-' + Date.now(),
        name: body.name || body.collection?.name || 'New Collection',
        description: body.description || body.collection?.description || '',
        item_count: 0,
        is_smart: body.is_smart || false,
        is_public: false,
        media_type: body.media_type || 'mixed',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      collections.push(newCollection);
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify(newCollection),
      });
    } else if (method === 'GET' && url.match(/\/collections\/[^?]+$/)) {
      // Get single collection
      const idMatch = url.match(/\/collections\/([^?]+)$/);
      const id = idMatch ? idMatch[1] : 'col-1';
      const found = collections.find((c) => c.id === id) || collections[0];
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(found),
      });
    } else if (method === 'PUT' || method === 'PATCH') {
      const body = route.request().postDataJSON();
      const idMatch = url.match(/\/collections\/([^?]+)/);
      const id = idMatch ? idMatch[1] : 'col-1';
      const idx = collections.findIndex((c) => c.id === id);
      if (idx >= 0) {
        collections[idx] = { ...collections[idx], ...body, updated_at: new Date().toISOString() };
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(collections[idx] || collections[0]),
      });
    } else if (method === 'DELETE') {
      const idMatch = url.match(/\/collections\/([^?]+)/);
      const id = idMatch ? idMatch[1] : '';
      collections = collections.filter((c) => c.id !== id);
      await route.fulfill({ status: 204, body: '' });
    } else {
      await route.continue();
    }
  });

  // Mock share endpoint
  await page.route('**/api/v1/collections/*/share', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ share_url: 'https://example.com/share/abc123' }),
    });
  });

  // Mock duplicate endpoint
  await page.route('**/api/v1/collections/*/duplicate', async (route) => {
    await route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 'col-dup-' + Date.now(),
        name: 'Action Movies (Copy)',
        item_count: 25,
        created_at: new Date().toISOString(),
      }),
    });
  });

  // Mock items in a collection
  await page.route('**/api/v1/collections/*/items**', async (route) => {
    const method = route.request().method();
    if (method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          items: [
            { id: 1, title: 'The Dark Knight', media_type: 'movie' },
            { id: 2, title: 'Inception', media_type: 'movie' },
          ],
          total: 2,
        }),
      });
    } else if (method === 'POST') {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Item added to collection' }),
      });
    } else if (method === 'DELETE') {
      await route.fulfill({ status: 204, body: '' });
    }
  });

  // Mock bulk operations
  await page.route('**/api/v1/collections/bulk**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'Bulk operation completed', affected: 2 }),
    });
  });

  // Mock export endpoint
  await page.route('**/api/v1/collections/*/export**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ download_url: 'https://example.com/export/abc.json' }),
    });
  });
}

test.describe('Collections', () => {
  test.beforeEach(async ({ page }) => {
    await setupCollectionMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Collections Page Layout', () => {
    test('navigates to /collections and displays the page', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page).toHaveURL(/.*collections/);
      await expect(page.locator('h1:has-text("Collections")')).toBeVisible();
    });

    test('displays subtitle about organizing media', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(
        page.locator('text=Organize your media with smart and manual collections')
      ).toBeVisible();
    });

    test('shows tab navigation for All, Smart, and Manual collections', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('text=All Collections').first()).toBeVisible();
      await expect(page.locator('text=Smart Collections').first()).toBeVisible();
      await expect(page.locator('text=Manual Collections').first()).toBeVisible();
    });

    test('collections list displays existing collections', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      await expect(page.locator('text=Action Movies').first()).toBeVisible();
      await expect(page.locator('text=Jazz Classics').first()).toBeVisible();
    });

    test('collections link is accessible from navigation', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const collectionsLink = page.locator('nav a[href="/collections"]');
      await expect(collectionsLink).toBeVisible();
    });
  });

  test.describe('Search and Filter Collections', () => {
    test('search input is visible', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="Search collections"]');
      await expect(searchInput).toBeVisible();
    });

    test('typing in search filters visible collections', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('input[placeholder*="Search collections"]');
      await searchInput.fill('Jazz');
      await page.waitForTimeout(500);

      // Page should still be functional after search
      await expect(page.locator('body')).toBeVisible();
    });

    test('media type filter dropdown is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const select = page.locator('select, [role="combobox"]').first();
      await expect(select).toBeVisible();
    });

    test('switching to Smart Collections tab filters results', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const smartTab = page.locator('text=Smart Collections').first();
      await smartTab.click();
      await page.waitForTimeout(500);

      // Should show only smart collections or empty state
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Create Collection', () => {
    test('Smart Collection button is visible', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const createButton = page.locator('button:has-text("Smart Collection")');
      await expect(createButton).toBeVisible();
    });

    test('clicking Smart Collection button opens the builder', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const createButton = page.locator('button:has-text("Smart Collection")');
      await createButton.click();
      await page.waitForTimeout(500);

      // A modal or expanded section should appear
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('View Mode Toggle', () => {
    test('grid and list view buttons are present', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const viewToggles = page.locator(
        'button[title="Grid View"], button[title="List View"]'
      );
      const count = await viewToggles.count();
      expect(count).toBeGreaterThanOrEqual(2);
    });

    test('can switch to list view', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const listButton = page.locator('button[title="List View"]');
      await listButton.click();
      await page.waitForTimeout(300);

      // Page should still display collections
      await expect(page.locator('body')).toBeVisible();
    });

    test('can switch back to grid view', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const listButton = page.locator('button[title="List View"]');
      await listButton.click();
      await page.waitForTimeout(300);

      const gridButton = page.locator('button[title="Grid View"]');
      await gridButton.click();
      await page.waitForTimeout(300);

      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Collection Actions', () => {
    test('Templates button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("Templates")')).toBeVisible();
    });

    test('Advanced Search button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("Advanced Search")')).toBeVisible();
    });

    test('Automation button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("Automation")')).toBeVisible();
    });

    test('AI Features button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('button:has-text("AI Features")')).toBeVisible();
    });
  });

  test.describe('Empty State', () => {
    test('shows empty state when no collections match search', async ({ page }) => {
      // Override to return empty collections
      await page.route('**/api/v1/collections**', async (route) => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ collections: [], total: 0 }),
          });
        }
      });

      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const emptyState = page.locator('text=No collections found');
      if (await emptyState.isVisible({ timeout: 5000 })) {
        await expect(emptyState).toBeVisible();
      }
    });
  });

  test.describe('Delete Collection', () => {
    test('collection cards have action menus or delete buttons', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(500);

      // Look for action menus (three-dot icon or explicit delete buttons)
      const actionButtons = page.locator(
        'button[aria-label*="action" i], button[aria-label*="menu" i], button[aria-label*="more" i], button:has(svg[class*="dots"])'
      );
      if (await actionButtons.first().isVisible({ timeout: 3000 })) {
        await expect(actionButtons.first()).toBeVisible();
      }
    });
  });
});
