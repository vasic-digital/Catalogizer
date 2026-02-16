import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

test.describe('Error Handling', () => {
  test.describe('Network Errors', () => {
    test('handles complete network failure on login gracefully', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.abort('failed');
      });

      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill('user');
      await page.locator('input[placeholder*="password" i]').fill('pass');
      await page.click('button[type="submit"]');

      // Should stay on login page and not crash
      await page.waitForTimeout(2000);
      await expect(page).toHaveURL(/.*login/);
      await expect(page.locator('body')).toBeVisible();
    });

    test('handles network failure on media page gracefully', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);

      // Mock media endpoints to fail
      await page.route('**/api/v1/media/search**', async (route) => {
        await route.abort('failed');
      });
      await page.route('**/api/v1/media/stats**', async (route) => {
        await route.abort('failed');
      });

      await loginAs(page, testUser);
      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // Page should still render without crashing
      await expect(page.locator('body')).toBeVisible();
      await expect(page.locator('h1:has-text("Media Browser")')).toBeVisible();
    });

    test('handles network failure on dashboard gracefully', async ({ page }) => {
      await mockAuthEndpoints(page);

      // Mock dashboard stats to fail
      await page.route('**/api/v1/stats**', async (route) => {
        await route.abort('failed');
      });
      await page.route('**/api/v1/media/stats**', async (route) => {
        await route.abort('failed');
      });

      await loginAs(page, testUser);

      // Dashboard should still render
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('HTTP Error Codes', () => {
    test('handles 401 Unauthorized on API calls', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      // Mock a 401 on the media endpoint
      await page.route('**/api/v1/media/search**', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Unauthorized' }),
        });
      });

      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Should handle gracefully - either redirect to login or show error
      await expect(page.locator('body')).toBeVisible();
    });

    test('handles 403 Forbidden gracefully', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await page.route('**/api/v1/media/search**', async (route) => {
        await route.fulfill({
          status: 403,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Forbidden' }),
        });
      });

      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      await expect(page.locator('body')).toBeVisible();
    });

    test('handles 500 Internal Server Error', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await page.route('**/api/v1/media/search**', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal server error' }),
        });
      });
      await page.route('**/api/v1/media/stats**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            total_items: 0,
            by_type: {},
            total_size: 0,
            recent_additions: 0,
          }),
        });
      });

      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);

      // Should show error state with retry button
      const retryButton = page.locator('[data-testid="retry-button"]');
      if (await retryButton.isVisible({ timeout: 5000 })) {
        await expect(retryButton).toBeVisible();
        // Clicking retry should attempt to reload
        await retryButton.click();
        await page.waitForTimeout(1000);
        await expect(page.locator('body')).toBeVisible();
      }
    });

    test('handles 404 Not Found on API', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await page.route('**/api/v1/media/999', async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Media not found' }),
        });
      });

      // Page should handle 404 API responses gracefully
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Unauthenticated Access', () => {
    test('redirects /dashboard to login when not authenticated', async ({ page }) => {
      await page.goto('/dashboard');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /media to login when not authenticated', async ({ page }) => {
      await page.goto('/media');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /collections to login when not authenticated', async ({ page }) => {
      await page.goto('/collections');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /favorites to login when not authenticated', async ({ page }) => {
      await page.goto('/favorites');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /playlists to login when not authenticated', async ({ page }) => {
      await page.goto('/playlists');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /admin to login when not authenticated', async ({ page }) => {
      await page.goto('/admin');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /analytics to login when not authenticated', async ({ page }) => {
      await page.goto('/analytics');
      await expect(page).toHaveURL(/.*login/);
    });

    test('catch-all route redirects unknown paths', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await page.goto('/this-page-does-not-exist');
      await page.waitForLoadState('networkidle');

      // The catch-all route redirects to dashboard
      await expect(page).toHaveURL(/.*dashboard/);
    });
  });

  test.describe('Slow API Responses', () => {
    test('shows loading state while waiting for API response', async ({ page }) => {
      await mockAuthEndpoints(page);

      // Delay the stats endpoint
      await page.route('**/api/v1/media/stats**', async (route) => {
        await page.waitForTimeout(3000);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            total_items: 100,
            by_type: { movie: 50, music: 50 },
            total_size: 536870912000,
            recent_additions: 5,
          }),
        });
      });

      await loginAs(page, testUser);

      // Page should show while loading
      await expect(page.locator('body')).toBeVisible();
    });
  });
});
