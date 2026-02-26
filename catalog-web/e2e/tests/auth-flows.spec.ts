import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';

test.describe('Authentication Flows', () => {
  test.describe('Login Page UI', () => {
    test.beforeEach(async ({ page }) => {
      await mockAuthEndpoints(page);
    });

    test('displays the Catalogizer branding and welcome text', async ({ page }) => {
      await page.goto('/login');
      await expect(page.locator('text=Welcome back')).toBeVisible();
      await expect(page.locator('text=Enter your credentials to access Catalogizer')).toBeVisible();
    });

    test('shows the logo with "C" initial', async ({ page }) => {
      await page.goto('/login');
      const logo = page.locator('.text-white.font-bold.text-xl');
      await expect(logo).toHaveText('C');
    });

    test('password field is masked by default', async ({ page }) => {
      await page.goto('/login');
      const passwordInput = page.locator('input[placeholder*="password" i]');
      await expect(passwordInput).toHaveAttribute('type', 'password');
    });

    test('toggles password visibility with eye icon', async ({ page }) => {
      await page.goto('/login');
      const passwordInput = page.locator('input[placeholder*="password" i]');

      // Password is initially hidden
      await expect(passwordInput).toHaveAttribute('type', 'password');

      // Click the eye toggle button (button sibling near the password input)
      // Use a more specific selector for the password visibility toggle
      const eyeButton = passwordInput.locator('..').locator('button[type="button"]');
      await eyeButton.click();

      // Password should now be visible
      await expect(passwordInput).toHaveAttribute('type', 'text');

      // Click again to hide
      await eyeButton.click();
      await expect(passwordInput).toHaveAttribute('type', 'password');
    });

    test('submit button is disabled when fields are empty', async ({ page }) => {
      await page.goto('/login');
      const submitButton = page.locator('button[type="submit"]');
      await expect(submitButton).toBeDisabled();
    });

    test('submit button enables when both fields have values', async ({ page }) => {
      await page.goto('/login');
      const submitButton = page.locator('button[type="submit"]');

      await page.locator('input[placeholder*="username" i]').fill('user');
      await expect(submitButton).toBeDisabled();

      await page.locator('input[placeholder*="password" i]').fill('pass');
      await expect(submitButton).toBeEnabled();
    });

    test('submit button disables again when a field is cleared', async ({ page }) => {
      await page.goto('/login');
      const submitButton = page.locator('button[type="submit"]');

      await page.locator('input[placeholder*="username" i]').fill('user');
      await page.locator('input[placeholder*="password" i]').fill('pass');
      await expect(submitButton).toBeEnabled();

      await page.locator('input[placeholder*="username" i]').fill('');
      await expect(submitButton).toBeDisabled();
    });
  });

  test.describe('Successful Login', () => {
    test.beforeEach(async ({ page }) => {
      await mockAuthEndpoints(page);
    });

    test('redirects to dashboard after successful login', async ({ page }) => {
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });

    test('stores auth_token in localStorage after login', async ({ page }) => {
      await loginAs(page, testUser);
      const token = await page.evaluate(() => localStorage.getItem('auth_token'));
      expect(token).toBeTruthy();
      expect(token).toContain('mock-jwt-token');
    });

    test('stores user object in localStorage after login', async ({ page }) => {
      await loginAs(page, testUser);
      const userStr = await page.evaluate(() => localStorage.getItem('user'));
      expect(userStr).toBeTruthy();
      const userData = JSON.parse(userStr!);
      expect(userData.username).toBe(testUser.username);
    });

    test('displays welcome message with username on dashboard', async ({ page }) => {
      await loginAs(page, testUser);
      const welcomeText = page.locator('text=Welcome back');
      await expect(welcomeText.first()).toBeVisible();
    });
  });

  test.describe('Failed Login', () => {
    test('stays on login page with invalid credentials', async ({ page }) => {
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Invalid username or password' }),
        });
      });
      await mockAuthEndpoints(page);

      // Re-mock login to return 401
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Invalid username or password' }),
        });
      });

      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill('wronguser');
      await page.locator('input[placeholder*="password" i]').fill('wrongpassword');
      await page.click('button[type="submit"]');

      await page.waitForTimeout(2000);
      await expect(page).toHaveURL(/.*login/);
    });

    test('handles server error (500) gracefully', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal server error' }),
        });
      });

      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill('user');
      await page.locator('input[placeholder*="password" i]').fill('pass');
      await page.click('button[type="submit"]');

      await page.waitForTimeout(2000);
      await expect(page).toHaveURL(/.*login/);
    });

    test('handles network timeout gracefully', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.route('**/api/v1/auth/login', async (route) => {
        await page.waitForTimeout(15000);
        await route.abort('timedout');
      });

      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill('user');
      await page.locator('input[placeholder*="password" i]').fill('pass');
      await page.click('button[type="submit"]');

      // Should stay on login page
      await page.waitForTimeout(2000);
      await expect(page).toHaveURL(/.*login/);
    });
  });

  test.describe('Session Persistence', () => {
    test('maintains auth state across page reload when token exists', async ({ page }) => {
      await mockAuthEndpoints(page);
      await loginAs(page, testUser);

      // Reload the page
      await page.reload();
      await page.waitForLoadState('networkidle');

      // Should remain on dashboard (not redirected to login)
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });
  });

  test.describe('Logout', () => {
    test('logout button is accessible from header', async ({ page }) => {
      await mockAuthEndpoints(page);
      await loginAs(page, testUser);

      // Look for logout icon button in header
      const logoutButton = page.locator('header button').filter({
        has: page.locator('svg')
      });
      // At least one button should be visible in the header
      await expect(logoutButton.first()).toBeVisible();
    });
  });

  test.describe('Navigation Links on Login Page', () => {
    test('has link to registration page', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      const registerLink = page.locator('a[href*="register"]');
      await expect(registerLink.first()).toBeVisible();
    });

    test('has link to forgot password', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      const forgotLink = page.locator('a[href*="forgot-password"]');
      await expect(forgotLink).toBeVisible();
    });

    test('remember me checkbox is present', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      const checkbox = page.locator('input[name="remember-me"]');
      await expect(checkbox).toBeVisible();
    });

    test('clicking "Create new account" navigates to register', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      await page.click('text=Create new account');
      await expect(page).toHaveURL(/.*register/);
    });
  });
});
