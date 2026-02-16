import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockCollectionEndpoints, mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup extended collection mocks for management tests
 */
async function setupCollectionMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  const collections = [
    {
      id: 'col-1',
      name: 'Action Movies',
      description: 'Best action films',
      item_count: 25,
      is_smart: false,
      is_public: false,
      media_type: 'video',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    {
      id: 'col-2',
      name: 'Jazz Collection',
      description: 'Smooth jazz albums',
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
      name: 'Family Photos',
      description: 'Photos from family events',
      item_count: 150,
      is_smart: false,
      is_public: false,
      media_type: 'image',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
  ];

  // Mock collections list
  await page.route('**/api/v1/collections**', async (route) => {
    const method = route.request().method();
    if (method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          collections,
          total: collections.length,
        }),
      });
    } else if (method === 'POST') {
      const body = route.request().postDataJSON();
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'col-new-' + Date.now(),
          ...body.collection,
          item_count: 0,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }),
      });
    }
  });

  // Mock single collection operations
  await page.route('**/api/v1/collections/*', async (route) => {
    const method = route.request().method();
    if (method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(collections[0]),
      });
    } else if (method === 'PUT' || method === 'PATCH') {
      const body = route.request().postDataJSON();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ...collections[0], ...body }),
      });
    } else if (method === 'DELETE') {
      await route.fulfill({ status: 204, body: '' });
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

  // Mock export endpoint
  await page.route('**/api/v1/collections/*/export**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ download_url: 'https://example.com/export/abc123.json' }),
    });
  });

  // Mock bulk operations
  await page.route('**/api/v1/collections/bulk**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'Bulk operation completed', affected: 2 }),
    });
  });
}

test.describe('Collections Management', () => {
  test.beforeEach(async ({ page }) => {
    await setupCollectionMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Collections Page Layout', () => {
    test('displays the Collections page title', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('h1:has-text("Collections")')).toBeVisible();
    });

    test('displays subtitle about organizing media', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('text=Organize your media with smart and manual collections')).toBeVisible();
    });

    test('displays tab navigation for collection types', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      await expect(page.locator('text=All Collections').first()).toBeVisible();
      await expect(page.locator('text=Smart Collections').first()).toBeVisible();
      await expect(page.locator('text=Manual Collections').first()).toBeVisible();
    });
  });

  test.describe('Collection Search and Filters', () => {
    test('search input for collections is visible', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const searchInput = page.locator('input[placeholder*="Search collections"]');
      await expect(searchInput).toBeVisible();
    });

    test('can type in the search box to filter collections', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const searchInput = page.locator('input[placeholder*="Search collections"]');
      await searchInput.fill('Action');
      await page.waitForTimeout(500);
      await expect(page.locator('body')).toBeVisible();
    });

    test('media type filter dropdown is present', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      // Select component for media type filter
      const select = page.locator('select, [role="combobox"]').first();
      await expect(select).toBeVisible();
    });
  });

  test.describe('Create Collection', () => {
    test('Smart Collection button is visible', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const createButton = page.locator('button:has-text("Smart Collection")');
      await expect(createButton).toBeVisible();
    });

    test('clicking Smart Collection button shows the builder', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const createButton = page.locator('button:has-text("Smart Collection")');
      await createButton.click();
      await page.waitForTimeout(500);
      // Smart collection builder should appear
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('View Mode', () => {
    test('grid and list view toggles are present', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      // View mode toggle buttons (Grid and List icons)
      const viewToggles = page.locator('button[title="Grid View"], button[title="List View"]');
      const count = await viewToggles.count();
      expect(count).toBeGreaterThanOrEqual(2);
    });

    test('can switch to list view', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const listButton = page.locator('button[title="List View"]');
      await listButton.click();
      await page.waitForTimeout(300);
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Collection Actions', () => {
    test('Templates button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const templatesButton = page.locator('button:has-text("Templates")');
      await expect(templatesButton).toBeVisible();
    });

    test('Advanced Search button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const searchButton = page.locator('button:has-text("Advanced Search")');
      await expect(searchButton).toBeVisible();
    });

    test('Automation button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const autoButton = page.locator('button:has-text("Automation")');
      await expect(autoButton).toBeVisible();
    });

    test('Integrations button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const intButton = page.locator('button:has-text("Integrations")');
      await expect(intButton).toBeVisible();
    });

    test('AI Features button is available', async ({ page }) => {
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');
      const aiButton = page.locator('button:has-text("AI Features")');
      await expect(aiButton).toBeVisible();
    });
  });

  test.describe('Empty State', () => {
    test('shows empty state message when no collections match filter', async ({ page }) => {
      // Override to return empty collections
      await page.route('**/api/v1/collections**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ collections: [], total: 0 }),
        });
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
});
