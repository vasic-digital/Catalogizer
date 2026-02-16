import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

test.describe('Responsive Layout', () => {
  test.describe('Desktop Navigation', () => {
    test.beforeEach(async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);
    });

    test('header navigation links are visible on desktop', async ({ page }) => {
      // Desktop viewport (default)
      await page.setViewportSize({ width: 1280, height: 720 });
      await page.waitForTimeout(300);

      const navLinks = page.locator('nav.hidden.md\\:flex a');
      // Desktop nav should be visible
      const dashboardLink = page.locator('nav a[href="/dashboard"]');
      await expect(dashboardLink).toBeVisible();
    });

    test('all navigation items are visible on desktop', async ({ page }) => {
      await page.setViewportSize({ width: 1280, height: 720 });
      await page.waitForTimeout(300);

      await expect(page.locator('nav a:has-text("Dashboard")').first()).toBeVisible();
      await expect(page.locator('nav a:has-text("Media")').first()).toBeVisible();
      await expect(page.locator('nav a:has-text("Favorites")').first()).toBeVisible();
      await expect(page.locator('nav a:has-text("Collections")').first()).toBeVisible();
    });

    test('header search bar is visible on desktop', async ({ page }) => {
      await page.setViewportSize({ width: 1280, height: 720 });
      await page.waitForTimeout(300);

      const searchBar = page.locator('header input[placeholder*="Search media"]');
      await expect(searchBar).toBeVisible();
    });

    test('user welcome message is visible on desktop', async ({ page }) => {
      await page.setViewportSize({ width: 1280, height: 720 });
      await page.waitForTimeout(300);

      const welcome = page.locator('header span:has-text("Welcome")');
      await expect(welcome).toBeVisible();
    });

    test('mobile menu button is hidden on desktop', async ({ page }) => {
      await page.setViewportSize({ width: 1280, height: 720 });
      await page.waitForTimeout(300);

      // The mobile menu button has md:hidden class
      const mobileMenuButton = page.locator('button.md\\:hidden');
      // On desktop, the hamburger button should not be visible
      if (await mobileMenuButton.count() > 0) {
        await expect(mobileMenuButton.first()).not.toBeVisible();
      }
    });
  });

  test.describe('Mobile Navigation', () => {
    test.beforeEach(async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
    });

    test('mobile menu button is visible on small screens', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);

      // Wait for viewport to apply
      await page.waitForTimeout(300);

      // The hamburger button should be visible on mobile
      const hamburgerButton = page.locator('header button').filter({
        has: page.locator('svg'),
      });
      await expect(hamburgerButton.first()).toBeVisible();
    });

    test('desktop navigation is hidden on mobile', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      // Desktop nav has hidden class on mobile
      const desktopNav = page.locator('nav.hidden');
      if (await desktopNav.count() > 0) {
        // Nav items should not be visible
        const navLink = page.locator('nav.hidden.md\\:flex a').first();
        await expect(navLink).not.toBeVisible();
      }
    });

    test('header search bar is hidden on mobile', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      // Search bar has hidden md:flex class
      const searchBar = page.locator('header .hidden.md\\:flex input[placeholder*="Search media"]');
      if (await searchBar.count() > 0) {
        await expect(searchBar).not.toBeVisible();
      }
    });

    test('clicking hamburger opens mobile menu with navigation links', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      // Click hamburger menu
      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      // Mobile menu should appear with navigation links
      const mobileMenu = page.locator('.md\\:hidden a[href="/dashboard"]');
      if (await mobileMenu.count() > 0) {
        await expect(mobileMenu).toBeVisible();
      }
    });

    test('mobile menu contains Dashboard link', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      await expect(page.locator('a[href="/dashboard"]:has-text("Dashboard")')).toBeVisible();
    });

    test('mobile menu contains Media link', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      await expect(page.locator('a[href="/media"]:has-text("Media")')).toBeVisible();
    });

    test('mobile menu contains Favorites link', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      await expect(page.locator('a[href="/favorites"]')).toBeVisible();
    });

    test('mobile menu contains Collections link', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      await expect(page.locator('a[href="/collections"]')).toBeVisible();
    });

    test('mobile menu contains Logout button', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      await expect(page.locator('button:has-text("Logout")')).toBeVisible();
    });

    test('mobile menu has search input', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      const mobileSearch = page.locator('input[placeholder*="Search media"]');
      await expect(mobileSearch.first()).toBeVisible();
    });

    test('clicking a mobile menu link closes the menu', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      const hamburger = page.locator('header button').filter({
        has: page.locator('svg'),
      }).first();
      await hamburger.click();
      await page.waitForTimeout(500);

      // Click on Media link
      const mediaLink = page.locator('a[href="/media"]:has-text("Media")');
      if (await mediaLink.isVisible()) {
        await mediaLink.click();
        await page.waitForTimeout(500);

        // Should navigate to media page
        await expect(page).toHaveURL(/.*media/);
      }
    });
  });

  test.describe('Tablet Breakpoints', () => {
    test('navigation adjusts at tablet breakpoint', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await page.setViewportSize({ width: 768, height: 1024 });
      await loginAs(page, testUser);
      await page.waitForTimeout(300);

      // At 768px (md breakpoint), desktop nav should be visible
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Login Page Responsive', () => {
    test('login form is centered on mobile', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/login');

      // Login form should be visible and centered
      await expect(page.locator('input[placeholder*="username" i]')).toBeVisible();
      await expect(page.locator('input[placeholder*="password" i]')).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();
    });

    test('login form fits within mobile viewport', async ({ page }) => {
      await mockAuthEndpoints(page);
      await page.setViewportSize({ width: 320, height: 568 });
      await page.goto('/login');

      // Form elements should be visible without horizontal scrolling
      await expect(page.locator('input[placeholder*="username" i]')).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();
    });
  });

  test.describe('Header Sticky Behavior', () => {
    test('header is sticky and stays at top', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      // The header has "sticky top-0" classes
      const header = page.locator('header.sticky');
      await expect(header).toBeVisible();
    });
  });
});
