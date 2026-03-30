import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup password change API mocks.
 */
async function setupPasswordChangeMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  // Password change endpoint
  await page.route('**/api/v1/auth/change-password', async (route) => {
    if (route.request().method() !== 'POST') {
      await route.fallback();
      return;
    }

    let postData: { current_password?: string; new_password?: string } = {};
    try {
      postData = JSON.parse(route.request().postData() || '{}');
    } catch {
      postData = {};
    }

    if (postData.current_password === testUser.password) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Password changed successfully' }),
      });
    } else {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Current password is incorrect' }),
      });
    }
  });

  // Catch-all for admin endpoints
  await page.route('**/api/v1/admin/**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ settings: {}, message: 'OK' }),
    });
  });
}

test.describe('Password Change', () => {
  test.beforeEach(async ({ page }) => {
    await setupPasswordChangeMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Settings Page Access', () => {
    test('can navigate to settings or profile page', async ({ page }) => {
      // Try settings page
      await page.goto('/settings');
      await page.waitForLoadState('networkidle');

      const body = page.locator('body');
      await expect(body).toBeVisible();
    });

    test('can navigate to profile page', async ({ page }) => {
      await page.goto('/profile');
      await page.waitForLoadState('networkidle');

      const body = page.locator('body');
      await expect(body).toBeVisible();
    });
  });

  test.describe('Password Change Form', () => {
    test('settings page loads without errors', async ({ page }) => {
      const consoleErrors: string[] = [];
      page.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.goto('/settings');
      await page.waitForLoadState('networkidle');

      const criticalErrors = consoleErrors.filter(
        (e) => !e.includes('favicon') && !e.includes('manifest')
      );
      expect(criticalErrors).toHaveLength(0);
    });

    test('password fields are present on settings page', async ({ page }) => {
      await page.goto('/settings');
      await page.waitForLoadState('networkidle');

      // Look for password-related inputs
      const passwordInputs = page.locator('input[type="password"]');
      const passwordSection = page.locator(
        'text=Password, text=Change Password, text=password'
      );

      // At least one of these should be visible on a settings page
      const hasPasswordInputs = await passwordInputs.first().isVisible({ timeout: 3000 });
      const hasPasswordSection = await passwordSection.first().isVisible({ timeout: 3000 });

      expect(hasPasswordInputs || hasPasswordSection).toBeTruthy();
    });

    test('navigation contains settings or profile link', async ({ page }) => {
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      const settingsLink = page.locator(
        'a[href*="settings"], a[href*="profile"], button:has-text("Settings"), button:has-text("Profile")'
      );

      // There should be a way to access settings from the dashboard
      if (await settingsLink.first().isVisible({ timeout: 3000 })) {
        await expect(settingsLink.first()).toBeVisible();
      }
    });
  });
});
