import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Setup auth mocks
    await mockAuthEndpoints(page);
    await mockDashboardEndpoints(page);

    // Login
    await page.goto('/login');
    await page.locator('input[placeholder*="username" i]').fill(testUser.username);
    await page.locator('input[placeholder*="password" i]').fill(testUser.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });

  test('displays dashboard page', async ({ page }) => {
    await expect(page).toHaveURL(/.*dashboard/);

    // Should show some dashboard content
    const content = page.locator('main, [role="main"], .dashboard');
    await expect(content.first()).toBeVisible();
  });

  test('displays navigation', async ({ page }) => {
    // Check for navigation
    const navigation = page.locator('nav, [role="navigation"], aside, header');
    await expect(navigation.first()).toBeVisible();
  });

  test('page renders without errors', async ({ page }) => {
    // Verify page renders
    await expect(page.locator('body')).toBeVisible();

    // Check that main content area exists
    const content = page.locator('main, [role="main"]');
    await expect(content.first()).toBeVisible();
  });
});
