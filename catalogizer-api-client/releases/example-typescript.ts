// TypeScript Catalogizer API Client Usage Example

import CatalogizerClient, {
  MediaSearchRequest,
  User,
  CatalogizerError,
  AuthenticationError
} from '@catalogizer/api-client';

class CatalogizerService {
  private client: CatalogizerClient;
  private currentUser: User | null = null;

  constructor(baseURL: string) {
    this.client = new CatalogizerClient({
      baseURL,
      enableWebSocket: true,
      webSocketURL: baseURL.replace('http', 'ws') + '/ws',
      timeout: 30000,
      retryAttempts: 3
    });

    this.setupEventListeners();
  }

  private setupEventListeners(): void {
    this.client.on('auth:login', (user: User) => {
      this.currentUser = user;
      console.log('User logged in:', user.username);
    });

    this.client.on('auth:logout', () => {
      this.currentUser = null;
      console.log('User logged out');
    });

    this.client.on('download:progress', (progress) => {
      console.log(`Download progress: ${progress.progress}%`);
    });
  }

  async login(username: string, password: string): Promise<User> {
    try {
      const response = await this.client.connect({ username, password });
      if (response) {
        this.currentUser = response.user;
        return response.user;
      }
      throw new Error('Login failed');
    } catch (error) {
      if (error instanceof AuthenticationError) {
        throw new Error('Invalid credentials');
      }
      throw error;
    }
  }

  async searchMedia(query: string): Promise<any[]> {
    const searchRequest: MediaSearchRequest = {
      query,
      sort_by: 'rating',
      sort_order: 'desc',
      limit: 50
    };

    try {
      const results = await this.client.media.search(searchRequest);
      return results.items;
    } catch (error) {
      console.error('Search failed:', error);
      throw error;
    }
  }

  async getMediaDetails(id: number) {
    return this.client.media.getById(id);
  }

  getCurrentUser(): User | null {
    return this.currentUser;
  }

  async disconnect(): Promise<void> {
    await this.client.disconnect();
  }
}

// Usage
async function example() {
  const service = new CatalogizerService('http://localhost:8080');

  try {
    const user = await service.login('demo', 'demo');
    console.log('Logged in as:', user.username);

    const movies = await service.searchMedia('action movies');
    console.log('Found movies:', movies.length);

    await service.disconnect();
  } catch (error) {
    console.error('Error:', error);
  }
}

export { CatalogizerService };
