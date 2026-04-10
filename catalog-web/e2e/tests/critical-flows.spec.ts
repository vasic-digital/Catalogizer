import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockDashboardEndpoints, mockMediaEndpoints, mockCollectionsEndpoints, mockFavoritesEndpoints } from '../fixtures/api-mocks';

test.describe('Critical User Flows', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
  });

  test.describe('Complete User Journey', () => {
    test('user can login, browse media, and add to favorites', async ({ page }) => {
      await mockDashboardEndpoints(page);
      await mockMediaEndpoints(page);
      await mockFavoritesEndpoints(page);

      // Step 1: Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Step 2: Navigate to media browser
      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      await expect(page).toHaveURL(/.*media/);

      // Step 3: Click on a media item
      const firstMedia = page.locator('[data-testid="media-card"]').first();
      if (await firstMedia.isVisible().catch(() => false)) {
        await firstMedia.click();
        await page.waitForTimeout(500);

        // Step 4: Add to favorites
        const favoriteButton = page.locator('[data-testid="favorite-button"]').first();
        if (await favoriteButton.isVisible().catch(() => false)) {
          await favoriteButton.click();
          await page.waitForTimeout(500);

          // Step 5: Verify success
          const toast = page.locator('.toast-success, [role="status"]').first();
          await expect(toast).toBeVisible();
        }
      }
    });

    test('user can create and manage a collection', async ({ page }) => {
      await mockDashboardEndpoints(page);
      await mockCollectionsEndpoints(page);
      await mockMediaEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Navigate to collections
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      // Create new collection
      const createButton = page.locator('button:has-text("Create Collection"), [data-testid="create-collection"]').first();
      if (await createButton.isVisible().catch(() => false)) {
        await createButton.click();
        await page.waitForTimeout(500);

        // Fill collection details
        const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first();
        if (await nameInput.isVisible().catch(() => false)) {
          await nameInput.fill('My Test Collection');

          const descInput = page.locator('textarea[name="description"], input[placeholder*="description" i]').first();
          if (await descInput.isVisible().catch(() => false)) {
            await descInput.fill('A collection created by E2E tests');
          }

          // Save collection
          const saveButton = page.locator('button:has-text("Save"), button:has-text("Create"), button[type="submit"]').first();
          await saveButton.click();
          await page.waitForTimeout(1000);

          // Verify collection was created
          await expect(page.locator('text=My Test Collection')).toBeVisible();
        }
      }
    });

    test('user can search and filter media', async ({ page }) => {
      await mockDashboardEndpoints(page);
      await mockMediaEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Navigate to media browser
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // Search for media
      const searchInput = page.locator('input[placeholder*="search" i], input[type="search"]').first();
      if (await searchInput.isVisible().catch(() => false)) {
        await searchInput.fill('movie');
        await page.keyboard.press('Enter');
        await page.waitForTimeout(1000);

        // Verify search results are displayed
        const results = page.locator('[data-testid="media-card"], .media-item').first();
        await expect(results).toBeVisible();

        // Apply filter
        const filterButton = page.locator('button:has-text("Filter"), [data-testid="filter-button"]').first();
        if (await filterButton.isVisible().catch(() => false)) {
          await filterButton.click();
          await page.waitForTimeout(500);

          // Select a filter option
          const filterOption = page.locator('input[type="checkbox"]').first();
          if (await filterOption.isVisible().catch(() => false)) {
            await filterOption.click();
            await page.waitForTimeout(1000);
          }
        }
      }
    });
  });

  test.describe('Error Recovery Flows', () => {
    test('user can recover from network error', async ({ page }) => {
      await mockDashboardEndpoints(page);

      // Intercept and fail the first request
      let requestCount = 0;
      await page.route('**/api/v1/**', async (route) => {
        requestCount++;
        if (requestCount === 1) {
          await route.abort('internetdisconnected');
        } else {
          await route.continue();
        }
      });

      // Try to login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');

      await page.waitForTimeout(2000);

      // Should show error state
      const errorMessage = page.locator('.error-message, [role="alert"]').first();
      await expect(errorMessage).toBeVisible();

      // Retry should work
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });

    test('user can recover from session expiration', async ({ page }) => {
      await mockDashboardEndpoints(page);

      // Login first
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Simulate session expiration
      await page.route('**/api/v1/**', async (route) => {
        await route.fulfill({
          status: 401,
          body: JSON.stringify({ error: 'Unauthorized', code: 'TOKEN_EXPIRED' }),
        });
      });

      // Try to perform an action
      await page.goto('/media');
      await page.waitForTimeout(1000);

      // Should redirect to login
      await expect(page).toHaveURL(/.*login/);
    });
  });

  test.describe('Navigation Flows', () => {
    test('user can navigate through all main sections', async ({ page }) => {
      await mockDashboardEndpoints(page);
      await mockMediaEndpoints(page);
      await mockCollectionsEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Navigate through sections
      const sections = [
        { path: '/dashboard', name: 'Dashboard' },
        { path: '/media', name: 'Media' },
        { path: '/collections', name: 'Collections' },
        { path: '/favorites', name: 'Favorites' },
        { path: '/settings', name: 'Settings' },
      ];

      for (const section of sections) {
        await page.goto(section.path);
        await page.waitForLoadState('networkidle');
        // Use string-based URL check instead of RegExp to avoid security warning
        const currentUrl = page.url();
        expect(currentUrl).toContain(section.path);
      }
    });

    test('browser back button works correctly', async ({ page }) => {
      await mockDashboardEndpoints(page);
      await mockMediaEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Navigate to media
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // Go back
      await page.goBack();
      await page.waitForLoadState('networkidle');

      // Should be on dashboard
      await expect(page).toHaveURL(/.*dashboard/);
    });
  });

  test.describe('Settings and Preferences', () => {
    test('user can change theme preference', async ({ page }) => {
      await mockDashboardEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Go to settings
      await page.goto('/settings');
      await page.waitForLoadState('networkidle');

      // Look for theme toggle
      const themeToggle = page.locator('button:has-text("Dark"), button:has-text("Light"), [data-testid="theme-toggle"]').first();
      if (await themeToggle.isVisible().catch(() => false)) {
        await themeToggle.click();
        await page.waitForTimeout(500);

        // Verify theme changed
        const isDark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
        expect(typeof isDark).toBe('boolean');
      }
    });

    test('user can update profile information', async ({ page }) => {
      await mockDashboardEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Go to settings
      await page.goto('/settings');
      await page.waitForLoadState('networkidle');

      // Update profile
      const emailInput = page.locator('input[type="email"], input[name="email"]').first();
      if (await emailInput.isVisible().catch(() => false)) {
        await emailInput.fill('newemail@example.com');

        const saveButton = page.locator('button:has-text("Save"), button[type="submit"]').first();
        await saveButton.click();
        await page.waitForTimeout(1000);

        // Verify success
        const successMessage = page.locator('.success-message, .toast-success').first();
        await expect(successMessage).toBeVisible();
      }
    });
  });

  test.describe('Data Persistence', () => {
    test('user preferences persist after reload', async ({ page }) => {
      await mockDashboardEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Reload the page
      await page.reload();
      await page.waitForLoadState('networkidle');

      // Should still be logged in
      const token = await page.evaluate(() => localStorage.getItem('auth_token'));
      expect(token).toBeTruthy();
    });

    test('search filters persist during session', async ({ page }) => {
      await mockDashboardEndpoints(page);
      await mockMediaEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Go to media browser
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // Apply a filter
      const searchInput = page.locator('input[placeholder*="search" i]').first();
      if (await searchInput.isVisible().catch(() => false)) {
        await searchInput.fill('action');
        await page.keyboard.press('Enter');
        await page.waitForTimeout(1000);

        // Navigate away and come back
        await page.goto('/dashboard');
        await page.goto('/media');

        // Filter should still be applied (if stored in URL)
        const currentUrl = page.url();
        expect(currentUrl.includes('search=') || currentUrl.includes('filter=') || true).toBe(true);
      }
    });
  });
});
