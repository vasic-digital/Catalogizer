import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockDashboardEndpoints, mockMediaEndpoints } from '../fixtures/api-mocks';

test.describe('Visual Regression', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
  });

  test.describe('Login Page', () => {
    test('login page matches baseline screenshot', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('login-page.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('login page with error state matches baseline', async ({ page }) => {
      await page.goto('/login');
      
      // Trigger an error state
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Invalid credentials' }),
        });
      });
      
      await page.locator('input[placeholder*="username" i]').fill('testuser');
      await page.locator('input[placeholder*="password" i]').fill('wrongpass');
      await page.click('button[type="submit"]');
      
      await page.waitForTimeout(500);
      
      await expect(page).toHaveScreenshot('login-page-error.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('login form focus states', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      const usernameInput = page.locator('input[placeholder*="username" i]');
      await usernameInput.focus();
      
      await expect(page).toHaveScreenshot('login-page-focus.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });
  });

  test.describe('Dashboard', () => {
    test('dashboard matches baseline screenshot', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('dashboard.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('dashboard loading state', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      
      // Capture immediately after navigation
      await expect(page).toHaveScreenshot('dashboard-loading.png', {
        fullPage: true,
        threshold: 0.3,
        maxDiffPixelRatio: 0.1,
      });
    });
  });

  test.describe('Media Browser', () => {
    test('media browser grid view matches baseline', async ({ page }) => {
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
      
      await expect(page).toHaveScreenshot('media-browser-grid.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('media browser list view matches baseline', async ({ page }) => {
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
      
      // Switch to list view
      const listViewButton = page.locator('[data-testid="view-mode-list"]').first();
      if (await listViewButton.isVisible().catch(() => false)) {
        await listViewButton.click();
        await page.waitForTimeout(500);
      }
      
      await expect(page).toHaveScreenshot('media-browser-list.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('media detail modal matches baseline', async ({ page }) => {
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
      
      // Click on first media item
      const firstMedia = page.locator('[data-testid="media-card"]').first();
      if (await firstMedia.isVisible().catch(() => false)) {
        await firstMedia.click();
        await page.waitForTimeout(500);
        
        await expect(page).toHaveScreenshot('media-detail-modal.png', {
          fullPage: false,
          threshold: 0.2,
        });
      }
    });
  });

  test.describe('Responsive Layouts', () => {
    test('login page on mobile viewport', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('login-page-mobile.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('login page on tablet viewport', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 });
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('login-page-tablet.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('dashboard on mobile viewport', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      await page.setViewportSize({ width: 375, height: 667 });
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('dashboard-mobile.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('dashboard on large desktop viewport', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      await page.setViewportSize({ width: 1920, height: 1080 });
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('dashboard-desktop.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });
  });

  test.describe('Dark Mode', () => {
    test('login page in dark mode', async ({ page }) => {
      // Set dark mode preference
      await page.emulateMedia({ colorScheme: 'dark' });
      
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('login-page-dark.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('dashboard in dark mode', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      await page.emulateMedia({ colorScheme: 'dark' });
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('dashboard-dark.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });
  });

  test.describe('Component States', () => {
    test('empty state screens', async ({ page }) => {
      await mockMediaEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      
      // Navigate to favorites (likely empty)
      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');
      
      await expect(page).toHaveScreenshot('empty-state.png', {
        fullPage: true,
        threshold: 0.2,
      });
    });

    test('loading skeletons', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      
      // Capture immediately while loading
      await expect(page).toHaveScreenshot('loading-skeletons.png', {
        fullPage: true,
        threshold: 0.3,
        maxDiffPixelRatio: 0.1,
      });
    });

    test('dropdown menus', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      await page.waitForLoadState('networkidle');
      
      // Click on user menu
      const userMenu = page.locator('[data-testid="user-menu"]').first();
      if (await userMenu.isVisible().catch(() => false)) {
        await userMenu.click();
        await page.waitForTimeout(300);
        
        await expect(page).toHaveScreenshot('dropdown-menu.png', {
          fullPage: false,
          threshold: 0.2,
        });
      }
    });
  });
});
