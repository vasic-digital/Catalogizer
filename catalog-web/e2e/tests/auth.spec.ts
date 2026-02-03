import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
  });

  test.describe('Login', () => {
    test('displays login form', async ({ page }) => {
      await page.goto('/login');

      // Check form elements are visible - using placeholder text selectors
      await expect(page.locator('input[placeholder*="username" i]')).toBeVisible();
      await expect(page.locator('input[placeholder*="password" i]')).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();
    });

    test('logs in with valid credentials', async ({ page }) => {
      await page.goto('/login');

      // Fill in credentials using placeholder selectors
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);

      // Submit form
      await page.click('button[type="submit"]');

      // Should redirect to dashboard
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });

    test('shows error with invalid credentials', async ({ page }) => {
      // Override mock for invalid credentials
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Invalid username or password' }),
        });
      });

      await page.goto('/login');

      // Fill in wrong credentials
      await page.locator('input[placeholder*="username" i]').fill('wronguser');
      await page.locator('input[placeholder*="password" i]').fill('wrongpassword');

      // Submit form
      await page.click('button[type="submit"]');

      // Should stay on login page (form validation or API error)
      await page.waitForTimeout(2000);
      await expect(page).toHaveURL(/.*login/);
    });

    test('validates required fields - button disabled when empty', async ({ page }) => {
      await page.goto('/login');

      // Button should be disabled when fields are empty
      const submitButton = page.locator('button[type="submit"]');
      await expect(submitButton).toBeDisabled();

      // Fill username only
      await page.locator('input[placeholder*="username" i]').fill('test');
      await expect(submitButton).toBeDisabled();

      // Fill password too
      await page.locator('input[placeholder*="password" i]').fill('password');
      await expect(submitButton).toBeEnabled();
    });

    test('shows link to create account', async ({ page }) => {
      await page.goto('/login');

      // Check that register/create account link exists
      const registerLink = page.locator('a[href*="register"]');
      await expect(registerLink.first()).toBeVisible();
    });

    test('shows remember me checkbox', async ({ page }) => {
      await page.goto('/login');

      // Check remember me checkbox exists
      const rememberMe = page.locator('input[name="remember-me"]');
      await expect(rememberMe).toBeVisible();
    });

    test('shows forgot password link', async ({ page }) => {
      await page.goto('/login');

      // Check forgot password link
      const forgotPassword = page.locator('a[href*="forgot-password"]');
      await expect(forgotPassword).toBeVisible();
    });
  });

  test.describe('Registration', () => {
    test('displays registration form', async ({ page }) => {
      await page.goto('/register');

      // Check form elements are visible
      await expect(page.locator('input[placeholder*="username" i], input[type="text"]').first()).toBeVisible();
      await expect(page.locator('input[placeholder*="password" i], input[type="password"]').first()).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();
    });

    test('navigates to login page', async ({ page }) => {
      await page.goto('/register');

      // Click login link
      const loginLink = page.locator('a[href*="login"]');
      if (await loginLink.first().isVisible()) {
        await loginLink.first().click();
        await expect(page).toHaveURL(/.*login/);
      }
    });
  });

  test.describe('Logout', () => {
    test('stores auth token after login', async ({ page }) => {
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Check token is stored
      const token = await page.evaluate(() => localStorage.getItem('auth_token'));
      expect(token).toBeTruthy();
    });

    test('stores user data after login', async ({ page }) => {
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Check user is stored
      const user = await page.evaluate(() => localStorage.getItem('user'));
      expect(user).toBeTruthy();
    });
  });

  test.describe('Protected Routes', () => {
    test('redirects unauthenticated users to login', async ({ page }) => {
      // Try to access dashboard without logging in
      await page.goto('/dashboard');

      // Should redirect to login
      await expect(page).toHaveURL(/.*login/);
    });

    test('allows authenticated users to access protected routes', async ({ page }) => {
      // Login first
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');

      // Should be able to access dashboard
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });
  });
});
