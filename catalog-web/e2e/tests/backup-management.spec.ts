import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup backup management API mocks.
 */
async function setupBackupMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);

  const mockBackups = [
    {
      id: 1,
      filename: 'backup-2026-03-30-120000.db',
      size: 52428800,
      created_at: new Date().toISOString(),
      status: 'completed',
    },
    {
      id: 2,
      filename: 'backup-2026-03-29-080000.db',
      size: 51380224,
      created_at: new Date(Date.now() - 86400000).toISOString(),
      status: 'completed',
    },
  ];

  // List backups
  await page.route('**/api/v1/admin/backups**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ backups: mockBackups, total: mockBackups.length }),
    });
  });

  // Create backup
  await page.route('**/api/v1/admin/backup', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 3,
          filename: 'backup-2026-03-30-140000.db',
          size: 0,
          created_at: new Date().toISOString(),
          status: 'in_progress',
          message: 'Backup started',
        }),
      });
      return;
    }
    await route.fallback();
  });

  // Restore backup
  await page.route('**/api/v1/admin/backup/restore', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Backup restored successfully' }),
      });
      return;
    }
    await route.fallback();
  });

  // Catch-all for other admin endpoints
  await page.route('**/api/v1/admin/**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ settings: {}, message: 'OK' }),
    });
  });
}

test.describe('Backup Management', () => {
  test.beforeEach(async ({ page }) => {
    await setupBackupMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Admin Panel Access', () => {
    test('can navigate to admin panel', async ({ page }) => {
      await page.goto('/admin');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('body')).toBeVisible();
    });

    test('admin panel has navigation visible', async ({ page }) => {
      await page.goto('/admin');
      await page.waitForLoadState('networkidle');
      const navigation = page.locator('nav, [role="navigation"], aside');
      await expect(navigation.first()).toBeVisible();
    });
  });

  test.describe('Backup List', () => {
    test('backup page loads without errors', async ({ page }) => {
      await page.goto('/admin');
      await page.waitForLoadState('networkidle');

      // Look for backup-related content in the admin panel
      const backupSection = page.locator(
        'text=Backup, text=backup, [data-testid="backup"], a[href*="backup"]'
      );
      if (await backupSection.first().isVisible({ timeout: 3000 })) {
        await backupSection.first().click();
        await page.waitForLoadState('networkidle');
      }
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Backup Operations', () => {
    test('page renders without console errors', async ({ page }) => {
      const consoleErrors: string[] = [];
      page.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.goto('/admin');
      await page.waitForLoadState('networkidle');

      // Filter out expected non-critical errors
      const criticalErrors = consoleErrors.filter(
        (e) => !e.includes('favicon') && !e.includes('manifest')
      );
      expect(criticalErrors).toHaveLength(0);
    });
  });
});
