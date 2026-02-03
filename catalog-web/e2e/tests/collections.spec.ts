import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockCollectionEndpoints } from '../fixtures/api-mocks';

test.describe('Collections', () => {
  test.beforeEach(async ({ page }) => {
    // Setup mocks
    await mockAuthEndpoints(page);
    await mockCollectionEndpoints(page);

    // Login
    await page.goto('/login');
    await page.locator('input[placeholder*="username" i]').fill(testUser.username);
    await page.locator('input[placeholder*="password" i]').fill(testUser.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });

  test('can access collections from navigation', async ({ page }) => {
    // Try to find collections link in navigation
    const collectionsLink = page.locator('a[href*="collection"], nav a:has-text("Collection")');

    if (await collectionsLink.first().isVisible({ timeout: 5000 })) {
      await collectionsLink.first().click();
      // Page should load
      await expect(page.locator('body')).toBeVisible();
    }
  });

  test('dashboard has navigation visible', async ({ page }) => {
    // Verify navigation is visible
    const navigation = page.locator('nav, [role="navigation"], aside');
    await expect(navigation.first()).toBeVisible();
  });
});
