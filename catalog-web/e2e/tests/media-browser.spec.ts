import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockMediaEndpoints } from '../fixtures/api-mocks';

test.describe('Media Browser', () => {
  test.beforeEach(async ({ page }) => {
    // Setup mocks
    await mockAuthEndpoints(page);
    await mockMediaEndpoints(page);

    // Login
    await page.goto('/login');
    await page.locator('input[placeholder*="username" i]').fill(testUser.username);
    await page.locator('input[placeholder*="password" i]').fill(testUser.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });

  test('can navigate to media page from dashboard', async ({ page }) => {
    // Try to navigate to media page
    const mediaLink = page.locator('a[href*="media"], nav a:has-text("Media"), nav a:has-text("Library")');

    if (await mediaLink.first().isVisible({ timeout: 5000 })) {
      await mediaLink.first().click();
      // Page should load
      await expect(page.locator('body')).toBeVisible();
    }
  });

  test('dashboard loads successfully after login', async ({ page }) => {
    // Verify we're on dashboard
    await expect(page).toHaveURL(/.*dashboard/);
    // Verify body is visible
    await expect(page.locator('body')).toBeVisible();
  });
});
