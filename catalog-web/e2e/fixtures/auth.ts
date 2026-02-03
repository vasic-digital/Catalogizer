import { test as base, expect, Page } from '@playwright/test';

/**
 * Test user credentials for E2E testing
 */
export const testUser = {
  username: 'testuser',
  password: 'testpassword123',
  firstName: 'Test',
  lastName: 'User',
};

export const adminUser = {
  username: 'admin',
  password: 'adminpassword123',
  firstName: 'Admin',
  lastName: 'User',
};

/**
 * Extended test fixtures with authentication helpers
 */
export const test = base.extend<{
  authenticatedPage: Page;
  adminPage: Page;
}>({
  /**
   * A page that is already authenticated as a regular user
   */
  authenticatedPage: async ({ page }, use) => {
    await mockAuthEndpoints(page);
    await loginAs(page, testUser);
    await use(page);
  },

  /**
   * A page that is already authenticated as an admin user
   */
  adminPage: async ({ page }, use) => {
    await mockAuthEndpoints(page, true);
    await loginAs(page, adminUser);
    await use(page);
  },
});

/**
 * Mock authentication API endpoints
 * The auth state is determined by checking if auth_token exists in localStorage
 */
export async function mockAuthEndpoints(page: Page, isAdmin = false) {
  const mockUser = {
    id: isAdmin ? 1 : 2,
    username: isAdmin ? adminUser.username : testUser.username,
    email: isAdmin ? 'admin@example.com' : 'test@example.com',
    first_name: isAdmin ? 'Admin' : 'Test',
    last_name: 'User',
    role: isAdmin ? 'admin' : 'user',
    permissions: isAdmin
      ? ['read:media', 'write:media', 'delete:media', 'admin:all']
      : ['read:media', 'write:media'],
    created_at: new Date().toISOString(),
  };

  // Mock login endpoint
  await page.route('**/api/v1/auth/login', async (route) => {
    const request = route.request();
    const body = request.postDataJSON();

    if (body.username && body.password) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          token: 'mock-jwt-token-' + Date.now(),
          user: mockUser,
        }),
      });
    } else {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Invalid credentials' }),
      });
    }
  });

  // Mock logout endpoint
  await page.route('**/api/v1/auth/logout', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'Logged out successfully' }),
    });
  });

  // Mock profile endpoint
  await page.route('**/api/v1/auth/profile', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(mockUser),
    });
  });

  // Mock auth status endpoint - always return authenticated with user data
  // The actual auth state is managed by localStorage in the browser
  await page.route('**/api/v1/auth/status', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        authenticated: true,
        user: mockUser,
        permissions: mockUser.permissions,
      }),
    });
  });

  // Mock permissions endpoint
  await page.route('**/api/v1/auth/permissions', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        permissions: mockUser.permissions,
      }),
    });
  });

  // Mock register endpoint
  await page.route('**/api/v1/auth/register', async (route) => {
    const body = route.request().postDataJSON();
    await route.fulfill({
      status: 201,
      contentType: 'application/json',
      body: JSON.stringify({
        token: 'mock-jwt-token-' + Date.now(),
        user: {
          id: 3,
          username: body.username,
          email: body.email,
          first_name: body.first_name,
          last_name: body.last_name,
          role: 'user',
          permissions: ['read:media', 'write:media'],
          created_at: new Date().toISOString(),
        },
      }),
    });
  });

  // Mock init-status endpoint
  await page.route('**/api/v1/auth/init-status', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        initialized: true,
        admin_exists: true,
      }),
    });
  });
}

/**
 * Login as a specific user
 */
export async function loginAs(page: Page, user: typeof testUser) {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  // Fill login form - target by placeholder text since the form uses custom Input component
  const usernameInput = page.locator('input[placeholder*="username" i]');
  const passwordInput = page.locator('input[placeholder*="password" i]');

  await usernameInput.fill(user.username);
  await passwordInput.fill(user.password);

  // Submit form - click the Sign in button
  await page.click('button[type="submit"]');

  // Wait for navigation to dashboard
  await page.waitForURL('**/dashboard', { timeout: 10000 });
}

/**
 * Logout the current user
 */
export async function logout(page: Page) {
  // Click user menu or logout button
  const logoutButton = page.locator('[data-testid="logout-button"], button:has-text("Logout"), button:has-text("Sign out")');
  if (await logoutButton.isVisible()) {
    await logoutButton.click();
  } else {
    // Try user menu dropdown
    const userMenu = page.locator('[data-testid="user-menu"], [aria-label="User menu"]');
    if (await userMenu.isVisible()) {
      await userMenu.click();
      await page.click('text=Logout');
    }
  }

  // Wait for redirect to login
  await page.waitForURL('**/login', { timeout: 10000 });
}

export { expect };
