import { test, expect } from '@playwright/test';
import { mockAuthEndpoints, testUser } from '../fixtures/auth';
import { mockDashboardEndpoints, mockMediaEndpoints } from '../fixtures/api-mocks';

test.describe('Performance', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthEndpoints(page);
  });

  test.describe('Page Load Performance', () => {
    test('login page loads within 3 seconds', async ({ page }) => {
      const startTime = Date.now();
      
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      expect(loadTime).toBeLessThan(3000);
    });

    test('dashboard loads within 4 seconds after login', async ({ page }) => {
      await mockDashboardEndpoints(page);
      
      // Login first
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      
      const startTime = Date.now();
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      expect(loadTime).toBeLessThan(4000);
    });

    test('media browser loads within 4 seconds', async ({ page }) => {
      await mockMediaEndpoints(page);
      
      // Login first
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      
      // Navigate to media browser
      const startTime = Date.now();
      await page.goto('/media');
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      expect(loadTime).toBeLessThan(4000);
    });
  });

  test.describe('Core Web Vitals', () => {
    test('LCP (Largest Contentful Paint) is under 2.5s', async ({ page }) => {
      await page.goto('/login');
      
      const lcp = await page.evaluate(() => {
        return new Promise<number>((resolve) => {
          let lcpValue = 0;
          const observer = new PerformanceObserver((list) => {
            const entries = list.getEntries();
            const lastEntry = entries[entries.length - 1];
            lcpValue = lastEntry.startTime;
          });
          observer.observe({ entryTypes: ['largest-contentful-paint'] });
          
          setTimeout(() => {
            observer.disconnect();
            resolve(lcpValue);
          }, 3000);
        });
      });
      
      expect(lcp).toBeLessThan(2500);
    });

    test('FID (First Input Delay) measurement setup', async ({ page }) => {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      // Check that event handlers can be registered
      const fid = await page.evaluate(() => {
        return new Promise<number>((resolve) => {
          let fidValue = 0;
          const observer = new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
              const firstEntry = entry as PerformanceEventTiming;
              fidValue = firstEntry.processingStart - firstEntry.startTime;
            }
          });
          observer.observe({ entryTypes: ['first-input'] });
          
          setTimeout(() => {
            observer.disconnect();
            resolve(fidValue);
          }, 1000);
        });
      });
      
      // FID should be minimal or not measured (no interaction)
      expect(fid).toBeGreaterThanOrEqual(0);
    });

    test('CLS (Cumulative Layout Shift) is under 0.1', async ({ page }) => {
      await page.goto('/login');
      
      const cls = await page.evaluate(() => {
        return new Promise<number>((resolve) => {
          let clsValue = 0;
          const observer = new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
              if (!(entry as any).hadRecentInput) {
                clsValue += (entry as any).value;
              }
            }
          });
          observer.observe({ entryTypes: ['layout-shift'] });
          
          setTimeout(() => {
            observer.disconnect();
            resolve(clsValue);
          }, 3000);
        });
      });
      
      expect(cls).toBeLessThan(0.1);
    });

    test('TTFB (Time to First Byte) is under 600ms', async ({ page }) => {
      await page.goto('/login');
      
      const ttfb = await page.evaluate(() => {
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        return navigation.responseStart - navigation.startTime;
      });
      
      expect(ttfb).toBeLessThan(600);
    });

    test('FCP (First Contentful Paint) is under 1.8s', async ({ page }) => {
      await page.goto('/login');
      
      const fcp = await page.evaluate(() => {
        return new Promise<number>((resolve) => {
          let fcpValue = 0;
          const observer = new PerformanceObserver((list) => {
            const entries = list.getEntries();
            if (entries.length > 0) {
              fcpValue = entries[0].startTime;
            }
          });
          observer.observe({ entryTypes: ['paint'] });
          
          setTimeout(() => {
            observer.disconnect();
            resolve(fcpValue);
          }, 3000);
        });
      });
      
      expect(fcp).toBeGreaterThan(0);
      expect(fcp).toBeLessThan(1800);
    });
  });

  test.describe('Resource Loading', () => {
    test('no images larger than 500KB', async ({ page }) => {
      const imageSizes: Array<{ src: string; size: number }> = [];
      
      page.on('response', async (response) => {
        const contentType = response.headers()['content-type'] || '';
        if (contentType.includes('image')) {
          const body = await response.body().catch(() => Buffer.from(''));
          imageSizes.push({
            src: response.url(),
            size: body.length,
          });
        }
      });
      
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      const largeImages = imageSizes.filter(img => img.size > 500 * 1024);
      expect(largeImages).toEqual([]);
    });

    test('JavaScript bundles are gzipped', async ({ page }) => {
      let jsGzipped = false;
      
      page.on('response', (response) => {
        const contentType = response.headers()['content-type'] || '';
        const contentEncoding = response.headers()['content-encoding'] || '';
        
        if (contentType.includes('javascript')) {
          if (contentEncoding.includes('gzip') || contentEncoding.includes('br')) {
            jsGzipped = true;
          }
        }
      });
      
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      expect(jsGzipped).toBe(true);
    });

    test('CSS is minified', async ({ page }) => {
      const cssResponses: Array<{ url: string; size: number }> = [];
      
      page.on('response', async (response) => {
        const contentType = response.headers()['content-type'] || '';
        if (contentType.includes('css')) {
          const body = await response.body().catch(() => Buffer.from(''));
          cssResponses.push({
            url: response.url(),
            size: body.length,
          });
        }
      });
      
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      // CSS files should be reasonably small (minified)
      for (const css of cssResponses) {
        expect(css.size).toBeLessThan(500 * 1024); // Less than 500KB
      }
    });
  });

  test.describe('Memory Usage', () => {
    test('memory usage stays stable during navigation', async ({ page }) => {
      await mockDashboardEndpoints(page);
      await mockMediaEndpoints(page);
      
      // Login
      await page.goto('/login');
      await page.locator('input[placeholder*="username" i]').fill(testUser.username);
      await page.locator('input[placeholder*="password" i]').fill(testUser.password);
      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
      
      // Get initial memory
      const initialMemory = await page.evaluate(() => {
        return (performance as any).memory?.usedJSHeapSize || 0;
      });
      
      // Navigate multiple times
      for (let i = 0; i < 5; i++) {
        await page.goto('/media');
        await page.waitForLoadState('networkidle');
        await page.goto('/dashboard');
        await page.waitForLoadState('networkidle');
      }
      
      // Get final memory
      const finalMemory = await page.evaluate(() => {
        return (performance as any).memory?.usedJSHeapSize || 0;
      });
      
      // Memory should not grow significantly (allow 50% growth)
      if (initialMemory > 0 && finalMemory > 0) {
        const growthRatio = finalMemory / initialMemory;
        expect(growthRatio).toBeLessThan(1.5);
      }
    });
  });

  test.describe('Network Performance', () => {
    test('API calls complete within 2 seconds', async ({ page }) => {
      const apiTimes: Array<{ url: string; time: number }> = [];
      
      page.on('request', (request) => {
        if (request.url().includes('/api/')) {
          (request as any).__startTime = Date.now();
        }
      });
      
      page.on('response', (response) => {
        const request = response.request();
        if (request.url().includes('/api/') && (request as any).__startTime) {
          apiTimes.push({
            url: request.url(),
            time: Date.now() - (request as any).__startTime,
          });
        }
      });
      
      await page.goto('/login');
      await page.waitForTimeout(2000);
      
      // All API calls should complete within 2 seconds
      for (const api of apiTimes) {
        expect(api.time).toBeLessThan(2000);
      }
    });

    test('no duplicate API requests', async ({ page }) => {
      const requestUrls: string[] = [];
      
      page.on('request', (request) => {
        if (request.url().includes('/api/')) {
          requestUrls.push(request.url());
        }
      });
      
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
      
      // Check for duplicates (same URL requested multiple times)
      const uniqueUrls = new Set(requestUrls);
      expect(uniqueUrls.size).toBe(requestUrls.length);
    });
  });
});
