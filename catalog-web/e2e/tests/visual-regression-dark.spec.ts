import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Dark-mode visual regression — exercises the REAL theme system (§11.4.162).
 *
 * Unlike the prefers-color-scheme block in visual-regression.spec.ts (which
 * emulates the OS preference), these tests drive the actual `ThemeToggle`
 * button and assert the `.dark` class is applied/removed on the document root,
 * proving the Catalogizer Blue light + dark palettes BOTH render. Each surface
 * is captured in light AND dark for a perceptual-diff baseline.
 */

/** Click the ThemeToggle and wait for the `.dark` class to reach `wantDark`. */
async function setDarkViaToggle(page: import('@playwright/test').Page, wantDark: boolean) {
  const isDark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
  if (isDark !== wantDark) {
    await page.locator('[data-testid="theme-toggle"]').first().click();
  }
  await expect
    .poll(() => page.evaluate(() => document.documentElement.classList.contains('dark')))
    .toBe(wantDark);
}

test.describe('Visual Regression — Dark Mode (real toggle)', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
    // Force a deterministic starting theme regardless of host OS preference,
    // so the first screenshot is always the light baseline (§11.4.50).
    await page.addInitScript(() => {
      try {
        window.localStorage.setItem('catalog-web-theme', 'light');
      } catch {
        /* localStorage may be unavailable in the test browser */
      }
    });
  });

  test('login page renders in light then dark via toggle', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // Light baseline (toggle is present on the public login header).
    await setDarkViaToggle(page, false);
    await expect(page).toHaveScreenshot('login-page-light-toggle.png', {
      fullPage: true,
      threshold: 0.2,
    });

    // Flip to dark via the real ThemeToggle and assert the class actually applied.
    await setDarkViaToggle(page, true);
    await expect(page).toHaveScreenshot('login-page-dark-toggle.png', {
      fullPage: true,
      threshold: 0.2,
    });
  });

  test('dashboard renders in light then dark via toggle', async ({ page }) => {
    await mockDashboardEndpoints(page);

    await page.goto('/login');
    await page.locator('input[placeholder*="username" i]').fill(testUser.username);
    await page.locator('input[placeholder*="password" i]').fill(testUser.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    await page.waitForLoadState('networkidle');

    await setDarkViaToggle(page, false);
    await expect(page).toHaveScreenshot('dashboard-light-toggle.png', {
      fullPage: true,
      threshold: 0.2,
    });

    await setDarkViaToggle(page, true);
    await expect(page).toHaveScreenshot('dashboard-dark-toggle.png', {
      fullPage: true,
      threshold: 0.2,
    });
  });

  test('theme choice persists across reload', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    await setDarkViaToggle(page, true);
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Persisted dark choice must survive a reload (localStorage path).
    await expect
      .poll(() => page.evaluate(() => document.documentElement.classList.contains('dark')))
      .toBe(true);
  });
});
