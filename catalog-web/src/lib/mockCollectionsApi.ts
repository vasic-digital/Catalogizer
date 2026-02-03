import { Collection, SmartCollection, CollectionRule, CollectionTemplate, CollectionAnalytics } from '../types/collections';

// Mock collections data for testing
const MOCK_COLLECTIONS: Collection[] = [
  {
    id: '1',
    name: 'My Music Collection',
    description: 'All my favorite music tracks organized by genre and artist',
    item_count: 1250,
    is_public: false,
    is_smart: false,
    primary_media_type: 'music',
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-20T15:30:00Z',
    thumbnail_url: '/collections/music-thumb.jpg',
    owner_id: 'user1'
  },
  {
    id: '2',
    name: 'Movie Favorites',
    description: 'Best movies in my collection',
    item_count: 87,
    is_public: true,
    is_smart: false,
    primary_media_type: 'video',
    created_at: '2024-01-10T09:00:00Z',
    updated_at: '2024-01-18T20:15:00Z',
    thumbnail_url: '/collections/movies-thumb.jpg',
    owner_id: 'user1'
  },
  {
    id: '3',
    name: 'Recently Added',
    description: 'Items added in the last 30 days',
    item_count: 45,
    is_public: false,
    is_smart: true,
    primary_media_type: 'mixed',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-20T18:45:00Z',
    owner_id: 'user1'
  }
];

const MOCK_SMART_COLLECTIONS: SmartCollection[] = [
  {
    id: '3',
    name: 'Recently Added',
    description: 'Items added in the last 30 days',
    is_smart: true,
    smart_rules: [
      {
        id: 'rule1',
        field: 'date_added',
        operator: 'last_30_days',
        value: null,
        field_type: 'date',
        label: 'Date Added'
      }
    ],
    item_count: 45,
    last_updated: '2024-01-20T18:45:00Z',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-20T18:45:00Z',
    thumbnail_url: '/collections/recent-thumb.jpg',
    owner_id: 'user1',
    is_public: false,
    primary_media_type: 'mixed'
  },
  {
    id: '4',
    name: 'High Rated Movies',
    description: 'Movies with rating 4 stars or higher',
    is_smart: true,
    smart_rules: [
      {
        id: 'rule1',
        field: 'media_type',
        operator: 'equals',
        value: 'video',
        field_type: 'select',
        label: 'Media Type',
        condition: 'AND'
      },
      {
        id: 'rule2',
        field: 'rating',
        operator: 'greater_or_equal',
        value: 4,
        field_type: 'number',
        label: 'Rating',
        condition: 'AND'
      }
    ],
    item_count: 23,
    last_updated: '2024-01-19T22:30:00Z',
    created_at: '2024-01-05T12:00:00Z',
    updated_at: '2024-01-19T22:30:00Z',
    thumbnail_url: '/collections/high-rated-thumb.jpg',
    owner_id: 'user1',
    is_public: true,
    primary_media_type: 'video'
  }
];

const MOCK_TEMPLATES: CollectionTemplate[] = [
  {
    id: 'recently_added',
    name: 'Recently Added',
    description: 'Items added in the last 30 days',
    category: 'Time-based',
    rules: [
      {
        field: 'date_added',
        operator: 'last_30_days',
        value: null,
        field_type: 'date',
        label: 'Date Added'
      }
    ],
    icon: 'Clock'
  },
  {
    id: 'high_rated',
    name: 'High Rated',
    description: 'Items with rating 4 stars or higher',
    category: 'Quality-based',
    rules: [
      {
        field: 'rating',
        operator: 'greater_or_equal',
        value: 4,
        field_type: 'number',
        label: 'Rating'
      }
    ],
    icon: 'Star'
  }
];

const MOCK_ANALYTICS: CollectionAnalytics[] = [
  {
    collection_id: '1',
    total_items: 1250,
    media_type_distribution: {
      music: 1250,
      video: 0,
      image: 0,
      document: 0
    },
    size_distribution: {
      total_size_bytes: 5368709120, // ~5GB
      average_size_bytes: 4294967, // ~4MB
      largest_item_size_bytes: 52428800 // ~50MB
    },
    quality_distribution: {
      hd: 0,
      sd: 1250,
      uhd: 0
    },
    time_based_stats: {
      items_added_today: 5,
      items_added_this_week: 25,
      items_added_this_month: 68,
      oldest_item_date: '2023-01-15T10:00:00Z',
      newest_item_date: '2024-01-20T15:30:00Z'
    },
    engagement_stats: {
      total_views: 3420,
      unique_viewers: 45,
      total_plays: 2890,
      average_completion_rate: 0.78,
      last_accessed: '2024-01-20T18:45:00Z'
    },
    genre_distribution: {
      'Rock': 450,
      'Pop': 320,
      'Jazz': 280,
      'Electronic': 200
    }
  }
];

// Mock API functions
export const mockCollectionsApi = {
  // Basic CRUD
  getCollections: async (): Promise<Collection[]> => {
    await new Promise(resolve => setTimeout(resolve, 100));
    return MOCK_COLLECTIONS;
  },

  getCollection: async (id: string): Promise<Collection> => {
    await new Promise(resolve => setTimeout(resolve, 100));
    const collection = MOCK_COLLECTIONS.find(c => c.id === id);
    if (!collection) throw new Error('Collection not found');
    return collection;
  },

  createCollection: async (data: { name: string; description?: string; items?: string[] }): Promise<Collection> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const newCollection: Collection = {
      id: Date.now().toString(),
      name: data.name,
      description: data.description,
      item_count: data.items?.length || 0,
      is_public: false,
      is_smart: false,
      primary_media_type: 'mixed',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      owner_id: 'user1'
    };
    MOCK_COLLECTIONS.push(newCollection);
    return newCollection;
  },

  updateCollection: async (id: string, data: Partial<Collection>): Promise<Collection> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const index = MOCK_COLLECTIONS.findIndex(c => c.id === id);
    if (index === -1) throw new Error('Collection not found');
    MOCK_COLLECTIONS[index] = { ...MOCK_COLLECTIONS[index], ...data, updated_at: new Date().toISOString() };
    return MOCK_COLLECTIONS[index];
  },

  deleteCollection: async (id: string): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const index = MOCK_COLLECTIONS.findIndex(c => c.id === id);
    if (index === -1) throw new Error('Collection not found');
    MOCK_COLLECTIONS.splice(index, 1);
  },

  // Smart Collections
  getSmartCollections: async (): Promise<SmartCollection[]> => {
    await new Promise(resolve => setTimeout(resolve, 100));
    return MOCK_SMART_COLLECTIONS;
  },

  getSmartCollection: async (id: string): Promise<SmartCollection> => {
    await new Promise(resolve => setTimeout(resolve, 100));
    const collection = MOCK_SMART_COLLECTIONS.find(c => c.id === id);
    if (!collection) throw new Error('Smart Collection not found');
    return collection;
  },

  createSmartCollection: async (data: { name: string; description?: string; rules: CollectionRule[] }): Promise<SmartCollection> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    const newSmartCollection: SmartCollection = {
      id: Date.now().toString(),
      name: data.name,
      description: data.description,
      is_smart: true,
      smart_rules: data.rules,
      item_count: Math.floor(Math.random() * 100) + 10, // Mock item count
      last_updated: new Date().toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      owner_id: 'user1',
      is_public: false,
      primary_media_type: 'mixed'
    };
    MOCK_SMART_COLLECTIONS.push(newSmartCollection);
    return newSmartCollection;
  },

  updateSmartCollection: async (id: string, data: Partial<SmartCollection>): Promise<SmartCollection> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const index = MOCK_SMART_COLLECTIONS.findIndex(c => c.id === id);
    if (index === -1) throw new Error('Smart Collection not found');
    MOCK_SMART_COLLECTIONS[index] = { ...MOCK_SMART_COLLECTIONS[index], ...data, updated_at: new Date().toISOString() };
    return MOCK_SMART_COLLECTIONS[index];
  },

  deleteSmartCollection: async (id: string): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    const index = MOCK_SMART_COLLECTIONS.findIndex(c => c.id === id);
    if (index === -1) throw new Error('Smart Collection not found');
    MOCK_SMART_COLLECTIONS.splice(index, 1);
  },

  // Templates
  getTemplates: async (): Promise<CollectionTemplate[]> => {
    await new Promise(resolve => setTimeout(resolve, 100));
    return MOCK_TEMPLATES;
  },

  // Analytics
  getAnalytics: async (collectionId: string): Promise<CollectionAnalytics> => {
    await new Promise(resolve => setTimeout(resolve, 150));
    const analytics = MOCK_ANALYTICS.find(a => a.collection_id === collectionId);
    if (!analytics) throw new Error('Analytics not found');
    return analytics;
  },

  // Test Rules
  testRules: async (rules: CollectionRule[]): Promise<{ valid: boolean; sample_count: number; errors?: string[] }> => {
    await new Promise(resolve => setTimeout(resolve, 100));
    // Mock rule validation - always return success for testing
    return {
      valid: true,
      sample_count: Math.floor(Math.random() * 1000) + 10
    };
  }
};

// Helper function to determine if we should use mock API
export const shouldUseMockCollections = () => {
  // Use mock API when backend doesn't have collections endpoints
  return true;
};