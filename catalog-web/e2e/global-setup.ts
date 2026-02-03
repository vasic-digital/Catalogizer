import { FullConfig } from '@playwright/test';

/**
 * Global setup for Playwright E2E tests
 * Runs once before all tests
 */
async function globalSetup(_config: FullConfig) {
  console.log('Starting E2E test setup...');

  // Ensure Playwright browsers are installed
  // Note: Run `npx playwright install` if browsers are missing

  // You can add additional global setup here:
  // - Start mock servers
  // - Seed test database
  // - Create test users
  // - Set up authentication state

  console.log('E2E test setup complete');
}

export default globalSetup;
