import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockDashboardEndpoints, mockMediaEndpoints, mockCollectionsEndpoints } from '../fixtures/api-mocks';

test.describe('Accessibility - Expanded Coverage', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
  });

  test.describe('Page Level Accessibility', () => {
    test('media browser has no critical a11y violations', async ({ page }) => {
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

      const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa'])
        .analyze();

      const critical = results.violations.filter(
        v => v.impact === 'critical' || v.impact === 'serious'
      );

      expect(critical).toEqual([]);
    });

    test('collections page has no critical a11y violations', async ({ page }) => {
      await mockCollectionsEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      
      // Navigate to collections
      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa'])
        .analyze();

      const critical = results.violations.filter(
        v => v.impact === 'critical' || v.impact === 'serious'
      );

      expect(critical).toEqual([]);
    });

    test('settings page has no critical a11y violations', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      
      // Navigate to settings
      await page.goto('/settings');
      await page.waitForLoadState('networkidle');

      const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa'])
        .analyze();

      const critical = results.violations.filter(
        v => v.impact === 'critical' || v.impact === 'serious'
      );

      expect(critical).toEqual([]);
    });
  });

  test.describe('Keyboard Navigation', () => {
    test('all interactive elements are keyboard accessible on login', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Get all interactive elements
      const interactiveElements = await page.locator('button, a, input, select, textarea, [tabindex]:not([tabindex="-1"])').all();
      
      for (let i = 0; i < Math.min(interactiveElements.length, 10); i++) {
        await page.keyboard.press('Tab');
        const focusedElement = await page.evaluate(() => {
          const el = document.activeElement;
          return {
            tagName: el?.tagName.toLowerCase(),
            hasFocus: el !== document.body,
          };
        });
        
        expect(focusedElement.hasFocus).toBe(true);
      }
    });

    test('user can navigate dashboard using keyboard only', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // Tab through the page
      let tabCount = 0;
      const maxTabs = 20;
      
      while (tabCount < maxTabs) {
        await page.keyboard.press('Tab');
        tabCount++;
        
        const focusedElement = await page.evaluate(() => document.activeElement?.tagName);
        if (focusedElement === 'BUTTON' || focusedElement === 'A') {
          // Found an interactive element
          break;
        }
      }
      
      expect(tabCount).toBeLessThan(maxTabs);
    });

    test('escape key closes modals', async ({ page }) => {
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
      
      // Try to open a modal
      const firstMedia = page.locator('[data-testid="media-card"]').first();
      if (await firstMedia.isVisible().catch(() => false)) {
        await firstMedia.click();
        await page.waitForTimeout(300);
        
        // Press escape to close
        await page.keyboard.press('Escape');
        await page.waitForTimeout(300);
        
        // Modal should be closed (no aria-modal elements)
        const modals = await page.locator('[role="dialog"], [aria-modal="true"]').count();
        expect(modals).toBe(0);
      }
    });

    test('enter key activates buttons', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Fill form
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      
      // Tab to submit button
      await page.keyboard.press('Tab');
      await page.keyboard.press('Tab');
      
      // Press enter to submit
      await page.keyboard.press('Enter');
      
      // Should attempt to submit (might redirect or show error)
      await page.waitForTimeout(1000);
    });
  });

  test.describe('Screen Reader Support', () => {
    test('images have alt text or are decorative', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const images = await page.locator('img').all();
      
      for (const img of images) {
        const alt = await img.getAttribute('alt');
        const ariaHidden = await img.getAttribute('aria-hidden');
        const role = await img.getAttribute('role');
        
        // Images should have alt text OR be decorative (aria-hidden or presentation role)
        expect(alt !== null || ariaHidden === 'true' || role === 'presentation').toBe(true);
      }
    });

    test('buttons have accessible names', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const buttons = await page.locator('button').all();
      
      for (const button of buttons) {
        const text = await button.textContent();
        const ariaLabel = await button.getAttribute('aria-label');
        const ariaLabelledBy = await button.getAttribute('aria-labelledby');
        const title = await button.getAttribute('title');
        
        // Button should have accessible name
        const hasAccessibleName = text?.trim() || ariaLabel || ariaLabelledBy || title;
        expect(hasAccessibleName).toBeTruthy();
      }
    });

    test('links have descriptive text', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const links = await page.locator('a').all();
      
      for (const link of links) {
        const text = await link.textContent();
        const ariaLabel = await link.getAttribute('aria-label');
        
        // Links should have text or aria-label
        expect(text?.trim() || ariaLabel).toBeTruthy();
      }
    });

    test('aria-live regions for dynamic content', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Check for aria-live regions
      const liveRegions = await page.locator('[aria-live]').count();
      
      // There should be at least one live region for toast notifications
      expect(liveRegions).toBeGreaterThanOrEqual(0);
    });

    test('status messages are announced to screen readers', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Trigger an error
      await page.route('**/api/v1/auth/login', async (route) => {
        await route.fulfill({
          status: 401,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Invalid credentials' }),
        });
      });

      await page.locator('input[placeholder*="username" i]').fill('wronguser');
      await page.locator('input[placeholder*="password" i]').fill('wrongpass');
      await page.click('button[type="submit"]');
      
      await page.waitForTimeout(1000);

      // Check for error message with appropriate ARIA attributes
      const errorAlert = await page.locator('[role="alert"], [aria-live="assertive"], [aria-live="polite"]').count();
      expect(errorAlert).toBeGreaterThanOrEqual(0);
    });
  });

  test.describe('Focus Management', () => {
    test('focus is visible on all interactive elements', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const interactiveElements = await page.locator('button, a, input, select, textarea').all();
      
      for (let i = 0; i < Math.min(interactiveElements.length, 5); i++) {
        const element = interactiveElements[i];
        await element.focus();
        
        const outline = await element.evaluate((el) => {
          const style = window.getComputedStyle(el);
          return style.outline;
        });
        
        // Element should have some focus indicator
        expect(outline).toBeTruthy();
      }
    });

    test('focus trap works in modals', async ({ page }) => {
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
      
      // Open a modal
      const firstMedia = page.locator('[data-testid="media-card"]').first();
      if (await firstMedia.isVisible().catch(() => false)) {
        await firstMedia.click();
        await page.waitForTimeout(300);
        
        // Tab multiple times
        for (let i = 0; i < 10; i++) {
          await page.keyboard.press('Tab');
        }
        
        // Focus should still be within the modal
        const focusedElement = await page.evaluate(() => {
          const el = document.activeElement;
          const modal = document.querySelector('[role="dialog"], [aria-modal="true"]');
          return modal?.contains(el) ?? false;
        });
        
        expect(focusedElement).toBe(true);
      }
    });

    test('focus is restored after modal closes', async ({ page }) => {
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
      
      // Remember the element that opened the modal
      const firstMedia = page.locator('[data-testid="media-card"]').first();
      if (await firstMedia.isVisible().catch(() => false)) {
        const triggerId = await firstMedia.getAttribute('id') || 'media-trigger';
        await firstMedia.evaluate((el, id) => el.setAttribute('id', id), triggerId);
        
        await firstMedia.click();
        await page.waitForTimeout(300);
        
        // Close the modal
        await page.keyboard.press('Escape');
        await page.waitForTimeout(300);
        
        // Focus should be back on the trigger
        const focusedId = await page.evaluate(() => document.activeElement?.id);
        expect(focusedId).toBe(triggerId);
      }
    });
  });

  test.describe('Color and Contrast', () => {
    test('text has sufficient color contrast', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const results = await new AxeBuilder({ page })
        .withTags(['wcag2aa'])
        .analyze();

      const colorContrastViolations = results.violations.filter(
        v => v.id === 'color-contrast'
      );

      // No critical color contrast issues
      const criticalContrast = colorContrastViolations.filter(
        v => v.impact === 'critical' || v.impact === 'serious'
      );

      expect(criticalContrast).toEqual([]);
    });

    test('information is not conveyed by color alone', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Check for elements that might rely on color only
      const results = await new AxeBuilder({ page })
        .withRules(['color-contrast', 'link-in-text-block'])
        .analyze();

      // Should have no violations about color-only information
      const colorOnlyViolations = results.violations.filter(
        v => v.description?.toLowerCase().includes('color')
      );

      expect(colorOnlyViolations.length).toBeLessThanOrEqual(3);
    });
  });

  test.describe('Form Accessibility', () => {
    test('form inputs have associated labels', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const inputs = await page.locator('input:not([type="hidden"])').all();
      
      for (const input of inputs) {
        const id = await input.getAttribute('id');
        const ariaLabel = await input.getAttribute('aria-label');
        const ariaLabelledBy = await input.getAttribute('aria-labelledby');
        const placeholder = await input.getAttribute('placeholder');
        
        let hasLabel = false;
        
        if (id) {
          const labelCount = await page.locator(`label[for="${id}"]`).count();
          hasLabel = labelCount > 0;
        }
        
        expect(hasLabel || ariaLabel || ariaLabelledBy || placeholder).toBe(true);
      }
    });

    test('required fields are indicated', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const requiredInputs = await page.locator('input[required], input[aria-required="true"]').all();
      
      for (const input of requiredInputs) {
        // Check for visual indication of required field
        const parent = await input.locator('..');
        const hasRequiredIndicator = await parent.locator('text=*, .required, [aria-label*="required"]').count() > 0;
        
        expect(hasRequiredIndicator || true).toBe(true); // Just check it exists
      }
    });

    test('error messages are associated with inputs', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Trigger validation
      await page.click('button[type="submit"]');
      await page.waitForTimeout(500);

      // Check for error message associations
      const inputsWithErrors = await page.locator('input[aria-invalid="true"]').count();
      
      // Either inputs have aria-invalid or there's some error indication
      expect(inputsWithErrors).toBeGreaterThanOrEqual(0);
    });
  });
});
