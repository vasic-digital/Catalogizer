import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

test.describe('Accessibility', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
  });

  test('login page has no critical a11y violations', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze();

    const critical = results.violations.filter(
      v => v.impact === 'critical' || v.impact === 'serious'
    );

    if (critical.length > 0) {
      console.log('Login page accessibility violations:', JSON.stringify(critical, null, 2));
    }

    expect(critical).toEqual([]);
  });

  test('registration page has no critical a11y violations', async ({ page }) => {
    await page.goto('/register');
    await page.waitForLoadState('networkidle');

    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze();

    const critical = results.violations.filter(
      v => v.impact === 'critical' || v.impact === 'serious'
    );

    if (critical.length > 0) {
      console.log('Registration page accessibility violations:', JSON.stringify(critical, null, 2));
    }

    expect(critical).toEqual([]);
  });

  test('dashboard page has no critical a11y violations', async ({ page }) => {
    await mockDashboardEndpoints(page);

    // Login first to access protected dashboard
    await page.goto('/login');
    await page.locator('input[placeholder*="username" i]').fill(testUser.username);
    await page.locator('input[placeholder*="password" i]').fill(testUser.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    await page.waitForLoadState('networkidle');

    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa'])
      .analyze();

    const critical = results.violations.filter(
      v => v.impact === 'critical' || v.impact === 'serious'
    );

    if (critical.length > 0) {
      console.log('Dashboard accessibility violations:', JSON.stringify(critical, null, 2));
    }

    expect(critical).toEqual([]);
  });

  test('login page has proper heading hierarchy', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // Check that at least one heading exists
    const headings = page.locator('h1, h2, h3, h4, h5, h6');
    const headingCount = await headings.count();
    expect(headingCount).toBeGreaterThan(0);
  });

  test('form inputs have associated labels or aria-labels', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    const inputs = page.locator('input:not([type="hidden"])');
    const inputCount = await inputs.count();

    for (let i = 0; i < inputCount; i++) {
      const input = inputs.nth(i);
      const ariaLabel = await input.getAttribute('aria-label');
      const ariaLabelledBy = await input.getAttribute('aria-labelledby');
      const placeholder = await input.getAttribute('placeholder');
      const id = await input.getAttribute('id');

      // Input should have at least one accessible name source
      const hasLabel = ariaLabel || ariaLabelledBy || placeholder;
      const hasAssociatedLabel = id
        ? (await page.locator(`label[for="${id}"]`).count()) > 0
        : false;

      expect(
        hasLabel || hasAssociatedLabel,
        `Input at index ${i} lacks an accessible name`
      ).toBeTruthy();
    }
  });

  test('interactive elements are keyboard accessible', async ({ page }) => {
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // Tab through the page and verify focus moves to interactive elements
    await page.keyboard.press('Tab');
    const firstFocused = await page.evaluate(() => {
      const el = document.activeElement;
      return el ? el.tagName.toLowerCase() : null;
    });

    // First focused element should be an interactive element
    expect(['input', 'button', 'a', 'select', 'textarea']).toContain(firstFocused);
  });
});
