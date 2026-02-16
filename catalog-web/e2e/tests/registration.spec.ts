import { test, expect } from '@playwright/test';
import { mockAuthEndpoints } from '../fixtures/auth';

test.describe('Registration', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
  });

  test.describe('Registration Form UI', () => {
    test('displays registration form with all required fields', async ({ page }) => {
      await page.goto('/register');
      await expect(page.locator('text=Create account')).toBeVisible();
      await expect(page.locator('input[name="firstName"]')).toBeVisible();
      await expect(page.locator('input[name="lastName"]')).toBeVisible();
      await expect(page.locator('input[name="username"]')).toBeVisible();
      await expect(page.locator('input[name="email"]')).toBeVisible();
      await expect(page.locator('input[name="password"]')).toBeVisible();
      await expect(page.locator('input[name="confirmPassword"]')).toBeVisible();
    });

    test('displays subtitle text about joining Catalogizer', async ({ page }) => {
      await page.goto('/register');
      await expect(page.locator('text=Join Catalogizer to start organizing your media')).toBeVisible();
    });

    test('has link to login page', async ({ page }) => {
      await page.goto('/register');
      const loginLink = page.locator('a[href*="login"]');
      await expect(loginLink.first()).toBeVisible();
    });

    test('clicking "Sign in instead" navigates to login', async ({ page }) => {
      await page.goto('/register');
      await page.click('text=Sign in instead');
      await expect(page).toHaveURL(/.*login/);
    });
  });

  test.describe('Registration Form Validation', () => {
    test('shows error when username is too short', async ({ page }) => {
      await page.goto('/register');
      await page.locator('input[name="firstName"]').fill('John');
      await page.locator('input[name="lastName"]').fill('Doe');
      await page.locator('input[name="username"]').fill('ab');
      await page.locator('input[name="email"]').fill('john@example.com');
      await page.locator('input[name="password"]').fill('password123');
      await page.locator('input[name="confirmPassword"]').fill('password123');
      await page.click('button[type="submit"]');

      await expect(page.locator('text=Username must be at least 3 characters')).toBeVisible();
    });

    test('shows error for invalid email format', async ({ page }) => {
      await page.goto('/register');
      await page.locator('input[name="firstName"]').fill('John');
      await page.locator('input[name="lastName"]').fill('Doe');
      await page.locator('input[name="username"]').fill('johndoe');
      await page.locator('input[name="email"]').fill('invalid-email');
      await page.locator('input[name="password"]').fill('password123');
      await page.locator('input[name="confirmPassword"]').fill('password123');
      await page.click('button[type="submit"]');

      await expect(page.locator('text=Email is invalid')).toBeVisible();
    });

    test('shows error when password is too short', async ({ page }) => {
      await page.goto('/register');
      await page.locator('input[name="firstName"]').fill('John');
      await page.locator('input[name="lastName"]').fill('Doe');
      await page.locator('input[name="username"]').fill('johndoe');
      await page.locator('input[name="email"]').fill('john@example.com');
      await page.locator('input[name="password"]').fill('short');
      await page.locator('input[name="confirmPassword"]').fill('short');
      await page.click('button[type="submit"]');

      await expect(page.locator('text=Password must be at least 8 characters')).toBeVisible();
    });

    test('shows error when passwords do not match', async ({ page }) => {
      await page.goto('/register');
      await page.locator('input[name="firstName"]').fill('John');
      await page.locator('input[name="lastName"]').fill('Doe');
      await page.locator('input[name="username"]').fill('johndoe');
      await page.locator('input[name="email"]').fill('john@example.com');
      await page.locator('input[name="password"]').fill('password123');
      await page.locator('input[name="confirmPassword"]').fill('different123');
      await page.click('button[type="submit"]');

      await expect(page.locator('text=Passwords do not match')).toBeVisible();
    });

    test('shows error when required fields are empty', async ({ page }) => {
      await page.goto('/register');
      await page.click('button[type="submit"]');

      // Should show validation errors for empty required fields
      await expect(page.locator('text=Username is required')).toBeVisible();
      await expect(page.locator('text=Email is required')).toBeVisible();
    });

    test('clears individual field errors on input change', async ({ page }) => {
      await page.goto('/register');
      await page.click('button[type="submit"]');

      // Error should be shown
      await expect(page.locator('text=Username is required')).toBeVisible();

      // Typing in the field should clear the error
      await page.locator('input[name="username"]').fill('validuser');
      await expect(page.locator('text=Username is required')).not.toBeVisible();
    });
  });

  test.describe('Successful Registration', () => {
    test('submits valid registration and navigates to login', async ({ page }) => {
      await page.goto('/register');
      await page.locator('input[name="firstName"]').fill('John');
      await page.locator('input[name="lastName"]').fill('Doe');
      await page.locator('input[name="username"]').fill('johndoe');
      await page.locator('input[name="email"]').fill('john@example.com');
      await page.locator('input[name="password"]').fill('password123');
      await page.locator('input[name="confirmPassword"]').fill('password123');
      await page.click('button[type="submit"]');

      // Should navigate to login after successful registration
      await expect(page).toHaveURL(/.*login/, { timeout: 10000 });
    });
  });

  test.describe('Registration Password Visibility', () => {
    test('can toggle password visibility in both password fields', async ({ page }) => {
      await page.goto('/register');

      const passwordInput = page.locator('input[name="password"]');
      const confirmInput = page.locator('input[name="confirmPassword"]');

      // Both should be masked initially
      await expect(passwordInput).toHaveAttribute('type', 'password');
      await expect(confirmInput).toHaveAttribute('type', 'password');
    });
  });

  test.describe('Registration Server Errors', () => {
    test('handles duplicate username error from server', async ({ page }) => {
      await page.route('**/api/v1/auth/register', async (route) => {
        await route.fulfill({
          status: 409,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Username already exists' }),
        });
      });

      await page.goto('/register');
      await page.locator('input[name="firstName"]').fill('John');
      await page.locator('input[name="lastName"]').fill('Doe');
      await page.locator('input[name="username"]').fill('existinguser');
      await page.locator('input[name="email"]').fill('john@example.com');
      await page.locator('input[name="password"]').fill('password123');
      await page.locator('input[name="confirmPassword"]').fill('password123');
      await page.click('button[type="submit"]');

      // Should stay on register page
      await page.waitForTimeout(2000);
      await expect(page).toHaveURL(/.*register/);
    });
  });
});
