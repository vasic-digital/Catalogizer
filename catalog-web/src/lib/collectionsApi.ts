import type { Collection } from '@/types/collections';

export const collectionsApi = {
  async getCollections(): Promise<Collection[]> {
    // Mock implementation - would be replaced with actual API call
    return [
      {
        id: '1',
        name: 'Favorite Movies',
        description: 'My all-time favorite movies',
        mediaCount: 42,
        duration: 6300, // minutes
        isSmart: false,
        createdAt: '2023-01-15T10:30:00Z',
        updatedAt: '2023-12-08T15:45:00Z'
      },
      {
        id: '2',
        name: 'Recent Documentaries',
        description: 'Auto-generated collection of documentaries from the last 6 months',
        mediaCount: 18,
        duration: 2100,
        isSmart: true,
        criteria: {
          genres: ['Documentary'],
          yearRange: [2023, 2023]
        },
        createdAt: '2023-06-01T09:00:00Z',
        updatedAt: '2023-12-09T12:30:00Z'
      },
      {
        id: '3',
        name: 'Highly Rated TV Shows',
        description: 'TV shows with rating 8.0 or higher',
        mediaCount: 25,
        duration: 18750,
        isSmart: true,
        criteria: {
          ratingRange: [8.0, 10.0],
          tags: ['tv_show']
        },
        createdAt: '2023-02-20T14:15:00Z',
        updatedAt: '2023-12-05T18:20:00Z'
      }
    ];
  },

  async createCollection(collectionData: Omit<Collection, 'id' | 'createdAt' | 'updatedAt'>): Promise<Collection> {
    // Mock implementation
    const newCollection: Collection = {
      ...collectionData,
      id: Date.now().toString(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    return newCollection;
  },

  async updateCollection(id: string, updates: Partial<Collection>): Promise<void> {
    // Mock implementation
    console.log(`Updating collection ${id}:`, updates);
  },

  async deleteCollection(id: string): Promise<void> {
    // Mock implementation
    console.log(`Deleting collection ${id}`);
  }
};