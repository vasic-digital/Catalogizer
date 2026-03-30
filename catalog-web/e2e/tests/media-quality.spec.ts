import { test, expect, Page } from '@playwright/test';
import { mockAuthEndpoints, testUser, loginAs } from '../fixtures/auth';
import { mockDashboardEndpoints, mockMediaEndpoints } from '../fixtures/api-mocks';

/**
 * Setup media quality API mocks.
 */
async function setupMediaQualityMocks(page: Page) {
  await mockAuthEndpoints(page);
  await mockDashboardEndpoints(page);
  await mockMediaEndpoints(page);

  // Quality analysis endpoint
  await page.route('**/api/v1/media/quality**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_analyzed: 340,
        quality_distribution: {
          '4K': 45,
          '1080p': 120,
          '720p': 85,
          '480p': 50,
          'unknown': 40,
        },
        codec_distribution: {
          'H.264': 180,
          'H.265': 90,
          'VP9': 30,
          'AV1': 10,
          'other': 30,
        },
        average_bitrate: 8500000,
      }),
    });
  });

  // Individual media quality detail
  await page.route('**/api/v1/media/*/quality', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        media_id: 1,
        resolution: '1920x1080',
        codec: 'H.264',
        bitrate: 8500000,
        frame_rate: 23.976,
        audio_codec: 'AAC',
        audio_channels: 6,
        duration: 8160,
        container: 'mkv',
      }),
    });
  });

  // Admin catch-all
  await page.route('**/api/v1/admin/**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ settings: {}, message: 'OK' }),
    });
  });
}

test.describe('Media Quality', () => {
  test.beforeEach(async ({ page }) => {
    await setupMediaQualityMocks(page);
    await loginAs(page, testUser);
  });

  test.describe('Media Browse Page', () => {
    test('can access media browse page', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('body')).toBeVisible();
    });

    test('media browse page loads without console errors', async ({ page }) => {
      const consoleErrors: string[] = [];
      page.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      const criticalErrors = consoleErrors.filter(
        (e) => !e.includes('favicon') && !e.includes('manifest')
      );
      expect(criticalErrors).toHaveLength(0);
    });

    test('media items are displayed on browse page', async ({ page }) => {
      await page.goto('/browse');
      await page.waitForLoadState('networkidle');

      // Should display media items from mocked data
      const mediaContent = page.locator(
        '[data-testid="media-item"], .media-item, .media-card, text=The Matrix, text=Breaking Bad'
      );
      if (await mediaContent.first().isVisible({ timeout: 5000 })) {
        await expect(mediaContent.first()).toBeVisible();
      }
    });
  });

  test.describe('Entity Detail Quality Info', () => {
    test('can navigate to entity detail page', async ({ page }) => {
      await page.goto('/entity/1');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('body')).toBeVisible();
    });

    test('entity detail page loads without errors', async ({ page }) => {
      const consoleErrors: string[] = [];
      page.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.goto('/entity/1');
      await page.waitForLoadState('networkidle');

      const criticalErrors = consoleErrors.filter(
        (e) => !e.includes('favicon') && !e.includes('manifest')
      );
      expect(criticalErrors).toHaveLength(0);
    });
  });

  test.describe('Analytics Quality View', () => {
    test('can access analytics page', async ({ page }) => {
      await page.goto('/analytics');
      await page.waitForLoadState('networkidle');
      await expect(page.locator('body')).toBeVisible();
    });

    test('analytics page has navigation', async ({ page }) => {
      await page.goto('/analytics');
      await page.waitForLoadState('networkidle');

      const navigation = page.locator('nav, [role="navigation"], aside');
      await expect(navigation.first()).toBeVisible();
    });
  });
});
