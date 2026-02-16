import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, adminUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints } from '../fixtures/api-mocks';

/**
 * Setup admin-related API mocks
 */
async function setupAdminMocks(page: Page) {
  // Mock admin system info
  await page.route('**/api/v1/admin/system**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        version: '1.0.0',
        uptime: 86400,
        cpuUsage: 45,
        memoryUsage: 62,
        diskUsage: {
          total: 1073741824000,
          used: 536870912000,
          free: 536870912000,
        },
        activeConnections: 12,
        totalRequests: 15420,
      }),
    });
  });

  // Mock admin users list
  await page.route('**/api/v1/admin/users**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: '1',
          username: 'admin',
          email: 'admin@example.com',
          role: 'admin',
          status: 'active',
          created_at: new Date().toISOString(),
          last_login: new Date().toISOString(),
        },
        {
          id: '2',
          username: 'testuser',
          email: 'test@example.com',
          role: 'user',
          status: 'active',
          created_at: new Date().toISOString(),
          last_login: new Date(Date.now() - 3600000).toISOString(),
        },
      ]),
    });
  });

  // Mock storage info
  await page.route('**/api/v1/admin/storage**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          path: '/media/videos',
          total_size: 536870912000,
          used_size: 268435456000,
          type: 'local',
          status: 'online',
        },
        {
          path: '//nas/media',
          total_size: 1073741824000,
          used_size: 268435456000,
          type: 'smb',
          status: 'online',
        },
      ]),
    });
  });

  // Mock backups
  await page.route('**/api/v1/admin/backups**', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 'backup-1',
            type: 'full',
            status: 'completed',
            size: 52428800,
            created_at: new Date().toISOString(),
          },
        ]),
      });
    } else if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'backup-new',
          type: 'full',
          status: 'in_progress',
          created_at: new Date().toISOString(),
        }),
      });
    }
  });

  // Mock settings endpoints
  await page.route('**/api/v1/settings**', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          theme: 'light',
          language: 'en',
          notifications_enabled: true,
          auto_scan: true,
          scan_interval: 3600,
          default_media_quality: 'high',
          subtitle_language: 'en',
        }),
      });
    } else if (route.request().method() === 'PUT') {
      const body = route.request().postDataJSON();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Settings updated', ...body }),
      });
    }
  });
}

test.describe('Settings and Admin', () => {
  test.describe('Admin Page Access', () => {
    test('admin link is visible for admin users', async ({ page }) => {
      await mockAuthEndpoints(page, true);
      await mockDashboardEndpoints(page);
      await setupAdminMocks(page);
      await loginAs(page, adminUser);

      const adminLink = page.locator('nav a[href="/admin"]');
      await expect(adminLink).toBeVisible();
    });

    test('admin link is hidden for regular users', async ({ page }) => {
      await mockAuthEndpoints(page, false);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      const adminLink = page.locator('nav a[href="/admin"]');
      await expect(adminLink).not.toBeVisible();
    });

    test('regular user is redirected from admin page to dashboard', async ({ page }) => {
      await mockAuthEndpoints(page, false);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await page.goto('/admin');
      await page.waitForLoadState('networkidle');

      // Should be redirected away from admin
      await expect(page).toHaveURL(/.*dashboard/);
    });

    test('admin user can access the admin page', async ({ page }) => {
      await mockAuthEndpoints(page, true);
      await mockDashboardEndpoints(page);
      await setupAdminMocks(page);
      await loginAs(page, adminUser);

      await page.goto('/admin');
      await page.waitForLoadState('networkidle');

      // Should stay on admin page
      await expect(page).toHaveURL(/.*admin/);
    });
  });

  test.describe('Dashboard System Status', () => {
    test('dashboard shows system status indicators', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      // Dashboard should show system status
      await expect(page.locator('text=System Status')).toBeVisible();
    });

    test('dashboard shows CPU usage', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('text=CPU Usage')).toBeVisible();
    });

    test('dashboard shows Memory usage', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('text=Memory Usage')).toBeVisible();
    });

    test('dashboard shows Disk usage', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('text=Disk Usage')).toBeVisible();
    });

    test('dashboard shows Network status', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('text=Network')).toBeVisible();
      await expect(page.locator('text=Online')).toBeVisible();
    });

    test('dashboard shows Uptime', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('text=Uptime')).toBeVisible();
    });
  });

  test.describe('Dashboard Quick Actions', () => {
    test('shows Quick Actions section', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('text=Quick Actions')).toBeVisible();
    });

    test('shows Upload Media quick action', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('button:has-text("Upload Media")')).toBeVisible();
    });

    test('shows Scan Library quick action', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      await expect(page.locator('button:has-text("Scan Library")')).toBeVisible();
    });

    test('shows Search quick action', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      // Quick Actions section has a Search button
      const searchButton = page.locator('button:has-text("Search")').last();
      await expect(searchButton).toBeVisible();
    });

    test('shows Settings quick action', async ({ page }) => {
      await mockAuthEndpoints(page);
      await mockDashboardEndpoints(page);
      await loginAs(page, testUser);

      // Quick Actions section has a Settings button
      const settingsButton = page.locator('button:has-text("Settings")').last();
      await expect(settingsButton).toBeVisible();
    });
  });
});
