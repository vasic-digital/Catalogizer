/**
 * Test data generator for Catalogizer Desktop tests
 */

export interface MediaItem {
  id: number;
  title: string;
  type: 'movie' | 'tv_show' | 'music_album' | 'game' | 'book';
  year: number;
  posterPath?: string;
  backdropPath?: string;
  overview: string;
  rating: number;
  runtime?: number;
  genres: string[];
}

export interface User {
  id: number;
  username: string;
  email: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

/**
 * Generates mock media items for testing
 */
export function generateMediaItems(count: number = 10): MediaItem[] {
  const items: MediaItem[] = [];
  const types: MediaItem['type'][] = ['movie', 'tv_show', 'music_album', 'game', 'book'];
  const genres = [
    'Action', 'Adventure', 'Comedy', 'Drama', 'Horror',
    'Sci-Fi', 'Fantasy', 'Romance', 'Thriller', 'Documentary'
  ];

  for (let i = 1; i <= count; i++) {
    const type = types[i % types.length];
    const title = `Test ${type.replace('_', ' ')} ${i}`;
    const itemGenres = [...genres].sort(() => Math.random() - 0.5).slice(0, 3);

    items.push({
      id: i,
      title,
      type,
      year: 2010 + (i % 15),
      posterPath: `/posters/poster_${i}.jpg`,
      backdropPath: `/backdrops/backdrop_${i}.jpg`,
      overview: `This is a test overview for ${title}. It's a great piece of media that everyone should experience.`,
      rating: 5.0 + (i % 5),
      runtime: type === 'movie' || type === 'tv_show' ? 90 + (i % 60) : undefined,
      genres: itemGenres,
    });
  }

  return items;
}

/**
 * Generates mock users for testing
 */
export function generateUsers(count: number = 5): User[] {
  const users: User[] = [];

  for (let i = 1; i <= count; i++) {
    users.push({
      id: i,
      username: `user${i}`,
      email: `user${i}@example.com`,
      createdAt: new Date(Date.now() - i * 86400000),
      updatedAt: new Date(),
    });
  }

  return users;
}

/**
 * Creates a successful API response
 */
export function createSuccessResponse<T>(data: T): ApiResponse<T> {
  return {
    success: true,
    data,
  };
}

/**
 * Creates an error API response
 */
export function createErrorResponse(message: string): ApiResponse<never> {
  return {
    success: false,
    error: message,
  };
}

/**
 * Creates a loading API response
 */
export function createLoadingResponse<T>(): ApiResponse<T> {
  return {
    success: false,
    message: 'Loading...',
  };
}

/**
 * Mock Tauri API functions for testing
 */
export const mockTauriApi = {
  invoke: vi.fn(),
  listen: vi.fn(),
  emit: vi.fn(),
};

/**
 * Resets all mock Tauri API functions
 */
export function resetTauriMocks() {
  mockTauriApi.invoke.mockReset();
  mockTauriApi.listen.mockReset();
  mockTauriApi.emit.mockReset();
}

/**
 * Sets up a successful Tauri API response
 */
export function setupTauriSuccessResponse(data: any) {
  mockTauriApi.invoke.mockResolvedValue(data);
}

/**
 * Sets up a failed Tauri API response
 */
export function setupTauriErrorResponse(error: string) {
  mockTauriApi.invoke.mockRejectedValue(new Error(error));
}
