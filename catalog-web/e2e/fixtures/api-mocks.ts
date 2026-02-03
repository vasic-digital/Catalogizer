import { Page } from '@playwright/test';

/**
 * Mock media data for E2E tests
 */
export const mockMedia = {
  items: [
    {
      id: 1,
      title: 'Test Movie 1',
      media_type: 'movie',
      year: 2023,
      rating: 8.5,
      poster_url: 'https://example.com/poster1.jpg',
      description: 'A test movie for E2E testing',
      duration: 120,
      created_at: new Date().toISOString(),
    },
    {
      id: 2,
      title: 'Test TV Show',
      media_type: 'tv_show',
      year: 2022,
      rating: 9.0,
      poster_url: 'https://example.com/poster2.jpg',
      description: 'A test TV show for E2E testing',
      seasons: 3,
      created_at: new Date().toISOString(),
    },
    {
      id: 3,
      title: 'Test Music Album',
      media_type: 'music',
      year: 2024,
      artist: 'Test Artist',
      poster_url: 'https://example.com/poster3.jpg',
      tracks: 12,
      created_at: new Date().toISOString(),
    },
  ],
  total: 3,
  limit: 20,
  offset: 0,
};

/**
 * Mock collections data
 */
export const mockCollections = [
  {
    id: 1,
    name: 'Favorites',
    description: 'My favorite media',
    item_count: 5,
    cover_image: 'https://example.com/collection1.jpg',
    created_at: new Date().toISOString(),
  },
  {
    id: 2,
    name: 'Watch Later',
    description: 'Media to watch later',
    item_count: 10,
    cover_image: 'https://example.com/collection2.jpg',
    created_at: new Date().toISOString(),
  },
];

/**
 * Mock dashboard stats
 */
export const mockDashboardStats = {
  total_items: 150,
  movies_count: 80,
  tv_shows_count: 40,
  music_count: 30,
  total_size: '250 GB',
  recent_additions: 12,
  storage_used: 75,
};

/**
 * Mock recent activity
 */
export const mockActivity = [
  {
    id: 1,
    type: 'media_added',
    title: 'New movie added',
    description: 'Test Movie 1 was added to the library',
    timestamp: new Date().toISOString(),
  },
  {
    id: 2,
    type: 'scan_completed',
    title: 'Library scan completed',
    description: 'Scanned 50 new files',
    timestamp: new Date(Date.now() - 3600000).toISOString(),
  },
];

/**
 * Setup all API mocks for media browsing
 */
export async function mockMediaEndpoints(page: Page) {
  // Mock media list/search endpoint
  await page.route('**/api/v1/media**', async (route) => {
    const url = new URL(route.request().url());
    const query = url.searchParams.get('q') || url.searchParams.get('query');

    let filteredItems = [...mockMedia.items];

    // Filter by search query
    if (query) {
      filteredItems = filteredItems.filter(item =>
        item.title.toLowerCase().includes(query.toLowerCase())
      );
    }

    // Filter by media type
    const mediaType = url.searchParams.get('media_type') || url.searchParams.get('type');
    if (mediaType) {
      filteredItems = filteredItems.filter(item => item.media_type === mediaType);
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: filteredItems,
        total: filteredItems.length,
        limit: 20,
        offset: 0,
      }),
    });
  });

  // Mock single media item endpoint
  await page.route('**/api/v1/media/*', async (route) => {
    const url = route.request().url();
    const idMatch = url.match(/\/media\/(\d+)/);
    const id = idMatch ? parseInt(idMatch[1]) : 1;

    const item = mockMedia.items.find(m => m.id === id) || mockMedia.items[0];

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(item),
    });
  });
}

/**
 * Setup collection API mocks
 */
export async function mockCollectionEndpoints(page: Page) {
  // Mock collections list
  await page.route('**/api/v1/collections', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          collections: mockCollections,
          total: mockCollections.length,
        }),
      });
    } else if (route.request().method() === 'POST') {
      const body = route.request().postDataJSON();
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          id: Date.now(),
          ...body,
          item_count: 0,
          created_at: new Date().toISOString(),
        }),
      });
    }
  });

  // Mock single collection
  await page.route('**/api/v1/collections/*', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockCollections[0]),
      });
    } else if (route.request().method() === 'DELETE') {
      await route.fulfill({
        status: 204,
        body: '',
      });
    }
  });
}

/**
 * Setup dashboard API mocks
 */
export async function mockDashboardEndpoints(page: Page) {
  // Mock stats endpoint
  await page.route('**/api/v1/stats**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(mockDashboardStats),
    });
  });

  // Mock activity endpoint
  await page.route('**/api/v1/activity**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        activities: mockActivity,
        total: mockActivity.length,
      }),
    });
  });

  // Mock recent media endpoint
  await page.route('**/api/v1/media/recent**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: mockMedia.items.slice(0, 5),
        total: 5,
      }),
    });
  });
}

/**
 * Setup all API mocks
 */
export async function mockAllEndpoints(page: Page) {
  await mockMediaEndpoints(page);
  await mockCollectionEndpoints(page);
  await mockDashboardEndpoints(page);
}
