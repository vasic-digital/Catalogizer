import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockAllEndpoints } from '../fixtures/api-mocks';

test.describe('Protected Routes', () => {
  test.describe('Unauthenticated Access', () => {
    test('redirects / to login', async ({ page }) => {
      await page.goto('/');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /dashboard to login', async ({ page }) => {
      await page.goto('/dashboard');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /media to login', async ({ page }) => {
      await page.goto('/media');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /collections to login', async ({ page }) => {
      await page.goto('/collections');
      await expect(page).toHaveURL(/.*login/);
    });

    test('redirects /favorites to login', async ({ page }) => {
      await page.goto('/favorites');
      await expect(page).toHaveURL(/.*login/);
    });

    test('allows access to /login', async ({ page }) => {
      await page.goto('/login');
      await expect(page).toHaveURL(/.*login/);

      // Should show login form
      const form = page.locator('form');
      await expect(form).toBeVisible();
    });

    test('allows access to /register', async ({ page }) => {
      await page.goto('/register');
      await expect(page).toHaveURL(/.*register/);

      // Should show register form
      const form = page.locator('form');
      await expect(form).toBeVisible();
    });
  });

  test.describe('Authenticated User Access', () => {
    test.beforeEach(async ({ page }) => {
      await mockAuthEndpoints(page, false);
      await mockAllEndpoints(page);

      // Login as regular user
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });

    test('stays on dashboard after login', async ({ page }) => {
      // Already logged in from beforeEach, verify we're on dashboard
      await expect(page).toHaveURL(/.*dashboard/);
      // Verify body loads
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Navigation Guards', () => {
    test('login redirects to dashboard', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockAllEndpoints(page);

      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');

      // Should redirect to dashboard after login
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });
  });
});
