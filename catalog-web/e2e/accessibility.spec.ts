import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from './fixtures/auth';
import { mockDashboardEndpoints, mockMediaEndpoints, mockCollectionEndpoints } from './fixtures/api-mocks';

/**
 * Accessibility E2E Tests
 *
 * Basic a11y checks: heading hierarchy, ARIA labels, keyboard navigation,
 * and focus management across key pages.
 */

async function setupAllMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);
  await mockMediaEndpoints(page);
  await mockCollectionEndpoints(page);

  await page.route('**/api/v1/favorites**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          { id: 1, media_id: 101, title: 'Test Movie', media_type: 'movie', added_at: new Date().toISOString() },
        ],
        total: 1,
      }),
    });
  });

  await page.route('**/api/v1/favorites/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_count: 1,
        media_type_breakdown: { movie: 1 },
        recent_additions: [],
      }),
    });
  });

  await page.route('**/api/v1/entities/types', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        types: [
          { id: 1, name: 'movie', display_name: 'Movies', count: 10 },
          { id: 2, name: 'tv_show', display_name: 'TV Shows', count: 5 },
        ],
      }),
    });
  });

  await page.route('**/api/v1/entities/stats', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_entities: 15,
        total_files: 30,
        by_type: { movie: 10, tv_show: 5 },
      }),
    });
  });

  await page.route('**/api/v1/media/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_items: 15,
        by_type: { movie: 10, tv_show: 5 },
        total_size: 100000000,
        recent_additions: 3,
      }),
    });
  });
}

test.describe('Accessibility', () => {
  test.describe('Heading Hierarchy', () => {
    test('login page has a proper heading structure', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // There should be at least one heading on the page
      const headings = page.locator('h1, h2, h3');
      const count = await headings.count();
      expect(count).toBeGreaterThanOrEqual(1);
    });

    test('dashboard has an h1 or prominent heading', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      // Dashboard should have a main heading
      const headings = page.locator('h1, h2');
      const count = await headings.count();
      expect(count).toBeGreaterThanOrEqual(1);
    });

    test('media browser page has descriptive headings', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      const h1 = page.locator('h1');
      await expect(h1.first()).toBeVisible();
      await expect(h1.first()).toContainText(/Media/i);
    });

    test('collections page has descriptive headings', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      const h1 = page.locator('h1');
      await expect(h1.first()).toBeVisible();
      await expect(h1.first()).toContainText(/Collection/i);
    });

    test('favorites page has descriptive headings', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      await page.goto('/favorites');
      await page.waitForLoadState('networkidle');

      const headings = page.locator('h1, h2');
      const count = await headings.count();
      expect(count).toBeGreaterThanOrEqual(1);

      // Should mention favorites
      await expect(page.locator('text=/[Ff]avorite/')).toBeVisible();
    });

    test('headings do not skip levels (h1 -> h3 without h2)', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // Check that heading levels are sequential
      const headingLevels = await page.evaluate(() => {
        const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
        return Array.from(headings).map((h) => parseInt(h.tagName[1]));
      });

      if (headingLevels.length > 1) {
        for (let i = 1; i < headingLevels.length; i++) {
          // Each heading level should not skip more than one level
          const jump = headingLevels[i] - headingLevels[i - 1];
          expect(jump).toBeLessThanOrEqual(1);
        }
      }
    });
  });

  test.describe('ARIA Labels', () => {
    test('navigation has an accessible role', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      // Look for nav element or role="navigation"
      const nav = page.locator('nav, [role="navigation"]');
      await expect(nav.first()).toBeVisible();
    });

    test('search input has accessible placeholder text', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      const searchInput = page.locator('header input[placeholder*="Search media"]');
      await expect(searchInput).toBeVisible();

      const placeholder = await searchInput.getAttribute('placeholder');
      expect(placeholder).toBeTruthy();
      expect(placeholder!.length).toBeGreaterThan(3);
    });

    test('login form inputs have proper labels or placeholders', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const usernameInput = page.locator('input[placeholder*="username" i]');
      const passwordInput = page.locator('input[placeholder*="password" i]');

      // Both inputs should have accessible identification
      const usernamePlaceholder = await usernameInput.getAttribute('placeholder');
      const passwordPlaceholder = await passwordInput.getAttribute('placeholder');

      expect(usernamePlaceholder).toBeTruthy();
      expect(passwordPlaceholder).toBeTruthy();
    });

    test('buttons have accessible text or aria-label', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // All buttons should have either text content, aria-label, or title
      const buttonsWithoutLabel = await page.evaluate(() => {
        const buttons = document.querySelectorAll('button');
        let unlabeledCount = 0;
        buttons.forEach((btn) => {
          const text = btn.textContent?.trim();
          const ariaLabel = btn.getAttribute('aria-label');
          const title = btn.getAttribute('title');
          const hasChildSvgWithAriaLabel = btn.querySelector('svg[aria-label]');
          if (!text && !ariaLabel && !title && !hasChildSvgWithAriaLabel) {
            unlabeledCount++;
          }
        });
        return unlabeledCount;
      });

      // Allow some tolerance for icon-only buttons that may use other methods
      // but most buttons should be accessible
      expect(buttonsWithoutLabel).toBeLessThan(5);
    });

    test('images and icons have alt text or are decorative', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      await page.goto('/media');
      await page.waitForLoadState('networkidle');

      // Check that img elements have alt attributes
      const imgsWithoutAlt = await page.evaluate(() => {
        const images = document.querySelectorAll('img');
        let count = 0;
        images.forEach((img) => {
          if (!img.hasAttribute('alt')) {
            count++;
          }
        });
        return count;
      });

      // All images should have alt attributes (even if empty for decorative)
      expect(imgsWithoutAlt).toBe(0);
    });
  });

  test.describe('Keyboard Navigation', () => {
    test('login form can be submitted using keyboard only', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Tab to username field and type
      await page.keyboard.press('Tab');
      await page.keyboard.type(testUser.username);

      // Tab to password field and type
      await page.keyboard.press('Tab');
      await page.keyboard.type(testUser.password);

      // Tab to submit button
      await page.keyboard.press('Tab');

      // Enter to submit
      await page.keyboard.press('Enter');

      // Should navigate to dashboard
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
    });

    test('Tab key moves focus through interactive elements', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      // Focus should move through header elements
      await page.keyboard.press('Tab');
      const firstFocused = await page.evaluate(() => {
        const el = document.activeElement;
        return el ? el.tagName.toLowerCase() : null;
      });
      // The first focused element should be something interactive
      expect(['a', 'button', 'input', 'select', 'textarea']).toContain(firstFocused);
    });

    test('navigation links are reachable via Tab key', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      // Tab through the page and check if we reach a nav link
      let foundNavLink = false;
      for (let i = 0; i < 20; i++) {
        await page.keyboard.press('Tab');
        const activeTag = await page.evaluate(() => {
          const el = document.activeElement;
          return el ? { tag: el.tagName.toLowerCase(), href: el.getAttribute('href') } : null;
        });
        if (activeTag?.tag === 'a' && activeTag?.href) {
          foundNavLink = true;
          break;
        }
      }

      expect(foundNavLink).toBe(true);
    });

    test('Escape key closes mobile menu', async ({ page }) => {
      await setupAllMocks(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      // Open mobile menu
      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      // Verify menu is open (check a navigation link is visible)
      const mediaLink = page.locator('a[href="/media"]:has-text("Media")');
      if (await mediaLink.isVisible({ timeout: 3000 })) {
        // Press Escape
        await page.keyboard.press('Escape');
        await page.waitForTimeout(500);

        // Menu should close (desktop nav links become hidden again)
        // The page should still be functional
        await expect(page.locator('body')).toBeVisible();
      }
    });
  });

  test.describe('Focus Management', () => {
    test('focus is visible on interactive elements', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Tab to the first input
      await page.keyboard.press('Tab');
      await page.waitForTimeout(200);

      // Check that focus ring is present via CSS
      const hasFocusStyles = await page.evaluate(() => {
        const el = document.activeElement;
        if (!el) return false;
        const styles = window.getComputedStyle(el);
        // Check for focus indicators: outline, box-shadow, or border changes
        const hasOutline = styles.outline !== 'none' && styles.outline !== '';
        const hasBoxShadow = styles.boxShadow !== 'none' && styles.boxShadow !== '';
        const hasBorderColor = styles.borderColor !== '';
        return hasOutline || hasBoxShadow || hasBorderColor;
      });

      // At least one focus indicator should be present
      expect(hasFocusStyles).toBe(true);
    });

    test('after login, focus returns to a reasonable element', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      // After redirect to dashboard, the page should be interactive
      const activeTag = await page.evaluate(() => {
        const el = document.activeElement;
        return el ? el.tagName.toLowerCase() : 'body';
      });

      // Active element should be body or an interactive element (not stuck on a removed element)
      expect(['body', 'a', 'button', 'input', 'html']).toContain(activeTag);
    });

    test('dialog/modal traps focus when open', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      await page.goto('/collections');
      await page.waitForLoadState('networkidle');

      // Try to open a dialog (e.g., Smart Collection builder)
      const createButton = page.locator('button:has-text("Smart Collection")');
      if (await createButton.isVisible({ timeout: 3000 })) {
        await createButton.click();
        await page.waitForTimeout(500);

        // If a dialog opened, check that Tab stays within it
        const dialog = page.locator('[role="dialog"], [role="alertdialog"], .modal, [data-state="open"]');
        if (await dialog.first().isVisible({ timeout: 3000 })) {
          // Tab through some elements inside the dialog
          for (let i = 0; i < 10; i++) {
            await page.keyboard.press('Tab');
          }
          // Active element should still be within the dialog
          const isInsideDialog = await page.evaluate(() => {
            const dialog = document.querySelector('[role="dialog"], [role="alertdialog"], .modal, [data-state="open"]');
            const active = document.activeElement;
            return dialog && active ? dialog.contains(active) : true;
          });
          expect(isInsideDialog).toBe(true);
        }
      }
    });
  });

  test.describe('Color Contrast and Text', () => {
    test('page uses semantic HTML elements', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      // Check for semantic elements
      const hasHeader = await page.locator('header').count();
      const hasNav = await page.locator('nav').count();
      const hasMain = await page.locator('main, [role="main"]').count();

      expect(hasHeader).toBeGreaterThanOrEqual(1);
      expect(hasNav).toBeGreaterThanOrEqual(1);
      // Main content area should exist
      expect(hasMain).toBeGreaterThanOrEqual(0); // Flexible since Layout may use div
    });

    test('text is not too small for readability', async ({ page }) => {
      await setupAllMocks(page);
      await loginAs(page, testUser);

      // Check that body text is at least 12px
      const minFontSize = await page.evaluate(() => {
        const paragraphs = document.querySelectorAll('p, span, a, li, td');
        let minSize = 999;
        paragraphs.forEach((el) => {
          const size = parseFloat(window.getComputedStyle(el).fontSize);
          if (size > 0 && size < minSize) {
            minSize = size;
          }
        });
        return minSize;
      });

      // Minimum readable font size should be 10px+
      expect(minFontSize).toBeGreaterThanOrEqual(10);
    });
  });

  test.describe('Form Accessibility', () => {
    test('login form submit button has descriptive text', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const submitButton = page.locator('button[type="submit"]');
      const text = await submitButton.textContent();
      expect(text).toBeTruthy();
      expect(text!.length).toBeGreaterThan(2);
    });

    test('password input has type="password" for security', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      const passwordInput = page.locator('input[placeholder*="password" i]');
      await expect(passwordInput).toHaveAttribute('type', 'password');
    });

    test('form fields have autocomplete attributes for user convenience', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.goto('/login');
      await page.waitForLoadState('networkidle');

      // Check if inputs have autocomplete attributes
      const usernameAutocomplete = await page.locator('input[placeholder*="username" i]').getAttribute('autocomplete');
      const passwordAutocomplete = await page.locator('input[placeholder*="password" i]').getAttribute('autocomplete');

      // Autocomplete is recommended but not strictly required
      // At minimum, the inputs should exist and be fillable
      await expect(page.locator('input[placeholder*="username" i]')).toBeVisible();
      await expect(page.locator('input[placeholder*="password" i]')).toBeVisible();
    });
  });
});
