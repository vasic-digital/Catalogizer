import { SmartCollection, CollectionRule, CreateCollectionRequest, UpdateCollectionRequest, CollectionAnalytics, ShareCollectionRequest, CollectionShareInfo, Collection, CollectionTemplate } from '../types/collections';
import { api } from './api';
import { mockCollectionsApi, shouldUseMockCollections } from './mockCollectionsApi';

class CollectionsApi {
  private baseUrl = '/api/collections';

  private async tryApiCall<T>(apiCall: () => Promise<T>, mockCall: () => Promise<T>): Promise<T> {
    if (shouldUseMockCollections()) {
      return mockCall();
    }
    
    try {
      return await apiCall();
    } catch (error: any) {
      // If API endpoint not found (404), fall back to mock data
      if (error.response?.status === 404) {
        return mockCall();
      }
      throw error;
    }
  }

  async getCollections(): Promise<SmartCollection[]> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}`);
        return response.data;
      },
      () => mockCollectionsApi.getSmartCollections()
    );
  }

  async getCollection(id: string): Promise<SmartCollection> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}/${id}`);
        return response.data;
      },
      () => mockCollectionsApi.getSmartCollection(id)
    );
  }

  async createCollection(collection: CreateCollectionRequest): Promise<SmartCollection> {
    return this.tryApiCall(
      async () => {
        const response = await api.post(`${this.baseUrl}`, collection);
        return response.data;
      },
      () => mockCollectionsApi.createSmartCollection({
        name: collection.name,
        description: collection.description,
        rules: collection.smart_rules || []
      })
    );
  }

  async updateCollection(id: string, updates: UpdateCollectionRequest): Promise<SmartCollection> {
    return this.tryApiCall(
      async () => {
        const response = await api.put(`${this.baseUrl}/${id}`, updates);
        return response.data;
      },
      () => mockCollectionsApi.updateSmartCollection(id, updates)
    );
  }

  async deleteCollection(id: string): Promise<void> {
    return this.tryApiCall(
      async () => {
        await api.delete(`${this.baseUrl}/${id}`);
      },
      () => mockCollectionsApi.deleteSmartCollection(id)
    );
  }

  async getCollectionItems(id: string, page = 1, limit = 50): Promise<any> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}/${id}/items`, {
          params: { page, limit }
        });
        return response.data;
      },
      async () => {
        // Mock items data
        const mockItems = Array.from({ length: 20 }, (_, i) => ({
          id: `item${i + 1}`,
          title: `Sample Item ${i + 1}`,
          artist: `Sample Artist ${i + 1}`,
          album: `Sample Album ${i + 1}`,
          duration: Math.floor(Math.random() * 300) + 120,
          media_type: ['music', 'video', 'image'][Math.floor(Math.random() * 3)],
          file_size: Math.floor(Math.random() * 10000000) + 1000000,
          date_added: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString()
        }));
        
        return {
          items: mockItems,
          total: mockItems.length,
          page,
          limit
        };
      }
    );
  }

  async refreshCollection(id: string): Promise<SmartCollection> {
    return this.tryApiCall(
      async () => {
        const response = await api.post(`${this.baseUrl}/${id}/refresh`);
        return response.data;
      },
      async () => {
        await new Promise(resolve => setTimeout(resolve, 1000));
        const collection = await mockCollectionsApi.getSmartCollection(id);
        collection.item_count = Math.floor(Math.random() * 100) + 10;
        collection.last_updated = new Date().toISOString();
        return collection;
      }
    );
  }

  async getCollectionAnalytics(id: string): Promise<CollectionAnalytics> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}/${id}/analytics`);
        return response.data;
      },
      () => mockCollectionsApi.getAnalytics(id)
    );
  }

  async shareCollection(id: string, shareRequest: ShareCollectionRequest): Promise<CollectionShareInfo> {
    return this.tryApiCall(
      async () => {
        const response = await api.post(`${this.baseUrl}/${id}/share`, shareRequest);
        return response.data;
      },
      async () => {
        const mockShareInfo: CollectionShareInfo = {
          share_id: `share_${id}_${Date.now()}`,
          share_url: `http://localhost:3006/shared/share_${id}_${Date.now()}`,
          expires_at: shareRequest.expires_at || new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
          created_at: new Date().toISOString(),
          access_count: 0,
          permissions: shareRequest
        };
        return mockShareInfo;
      }
    );
  }

  async getSharedCollection(shareId: string): Promise<SmartCollection> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}/shared/${shareId}`);
        return response.data;
      },
      async () => {
        return mockCollectionsApi.getSmartCollection('3'); // Return recently added as mock shared
      }
    );
  }

  async exportCollection(id: string, format: 'json' | 'csv' | 'm3u' = 'json'): Promise<Blob> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}/${id}/export`, {
          params: { format },
          responseType: 'blob'
        });
        return response.data;
      },
      async () => {
        const collection = await mockCollectionsApi.getSmartCollection(id);
        let data: string;
        
        if (format === 'json') {
          data = JSON.stringify(collection, null, 2);
        } else {
          data = `name,description\n${collection.name},"${collection.description || ''}"`;
        }
        
        return new Blob([data], { type: format === 'json' ? 'application/json' : 'text/csv' });
      }
    );
  }

  async importCollection(file: File): Promise<SmartCollection> {
    return this.tryApiCall(
      async () => {
        const formData = new FormData();
        formData.append('file', file);
        
        const response = await api.post(`${this.baseUrl}/import`, formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        });
        return response.data;
      },
      async () => {
        const text = await file.text();
        const data = JSON.parse(text) as SmartCollection;
        return mockCollectionsApi.createSmartCollection({
          name: data.name,
          description: data.description,
          rules: data.smart_rules || []
        });
      }
    );
  }

  async duplicateCollection(id: string, newName?: string): Promise<SmartCollection> {
    return this.tryApiCall(
      async () => {
        const response = await api.post(`${this.baseUrl}/${id}/duplicate`, {
          name: newName
        });
        return response.data;
      },
      async () => {
        const original = await mockCollectionsApi.getSmartCollection(id);
        return mockCollectionsApi.createSmartCollection({
          name: newName || `${original.name} (Copy)`,
          description: original.description,
          rules: original.smart_rules
        });
      }
    );
  }

  async addItemsToCollection(id: string, itemIds: string[]): Promise<void> {
    return this.tryApiCall(
      async () => {
        await api.post(`${this.baseUrl}/${id}/items`, { item_ids: itemIds });
      },
      async () => {
        // Mock implementation - just delay
        await new Promise(resolve => setTimeout(resolve, 200));
      }
    );
  }

  async removeItemsFromCollection(id: string, itemIds: string[]): Promise<void> {
    return this.tryApiCall(
      async () => {
        await api.delete(`${this.baseUrl}/${id}/items`, {
          data: { item_ids: itemIds }
        });
      },
      async () => {
        // Mock implementation - just delay
        await new Promise(resolve => setTimeout(resolve, 200));
      }
    );
  }

  async getCollectionSuggestions(): Promise<string[]> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}/suggestions`);
        return response.data;
      },
      async () => {
        return [
          'Summer Mix 2024',
          'Workout Music',
          'Evening Jazz',
          'Movie Night Favorites',
          'Travel Photography'
        ];
      }
    );
  }

  async getTemplates(): Promise<CollectionTemplate[]> {
    return this.tryApiCall(
      async () => {
        const response = await api.get(`${this.baseUrl}/templates`);
        return response.data;
      },
      () => mockCollectionsApi.getTemplates()
    );
  }

  async testRules(rules: CollectionRule[]): Promise<{ valid: boolean; sample_count: number; errors?: string[] }> {
    return this.tryApiCall(
      async () => {
        const response = await api.post(`${this.baseUrl}/test-rules`, { rules });
        return response.data;
      },
      () => mockCollectionsApi.testRules(rules)
    );
  }

  // Bulk operations
  async bulkDeleteCollections(collectionIds: string[]): Promise<void> {
    return this.tryApiCall(
      async () => {
        await api.delete(`${this.baseUrl}/bulk`, { data: { collection_ids: collectionIds } });
      },
      async () => {
        for (const id of collectionIds) {
          await mockCollectionsApi.deleteSmartCollection(id);
        }
      }
    );
  }

  async bulkShareCollections(collectionIds: string[], shareRequest: ShareCollectionRequest): Promise<CollectionShareInfo[]> {
    return this.tryApiCall(
      async () => {
        const response = await api.post(`${this.baseUrl}/bulk/share`, {
          collection_ids: collectionIds,
          share_request: shareRequest
        });
        return response.data;
      },
      async () => {
        const results: CollectionShareInfo[] = [];
        for (const id of collectionIds) {
          const mockShareInfo: CollectionShareInfo = {
            share_id: `share_${id}_${Date.now()}`,
            share_url: `http://localhost:3006/shared/share_${id}_${Date.now()}`,
            expires_at: shareRequest.expires_at || new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
            created_at: new Date().toISOString(),
            access_count: 0,
            permissions: shareRequest
          };
          results.push(mockShareInfo);
        }
        return results;
      }
    );
  }

  async bulkExportCollections(collectionIds: string[], format: 'json' | 'csv' | 'm3u' = 'json'): Promise<Blob> {
    return this.tryApiCall(
      async () => {
        const response = await api.post(`${this.baseUrl}/bulk/export`, {
          collection_ids: collectionIds,
          format
        }, {
          responseType: 'blob'
        });
        return response.data;
      },
      async () => {
        let data: string;
        const collections = [];
        
        for (const id of collectionIds) {
          const collection = await mockCollectionsApi.getSmartCollection(id);
          collections.push(collection);
        }
        
        if (format === 'json') {
          data = JSON.stringify(collections, null, 2);
        } else {
          data = 'name,description\n' + collections.map(c => `${c.name},"${c.description || ''}"`).join('\n');
        }
        
        return new Blob([data], { type: format === 'json' ? 'application/json' : 'text/csv' });
      }
    );
  }

  async bulkUpdateCollections(collectionIds: string[], action: string, updates?: UpdateCollectionRequest): Promise<SmartCollection[]> {
    return this.tryApiCall(
      async () => {
        const response = await api.put(`${this.baseUrl}/bulk`, {
          collection_ids: collectionIds,
          action,
          updates
        });
        return response.data;
      },
      async () => {
        const results: SmartCollection[] = [];
        for (const id of collectionIds) {
          if (action === 'duplicate') {
            const original = await mockCollectionsApi.getSmartCollection(id);
            const duplicate = await mockCollectionsApi.createSmartCollection({
              name: `${original.name} (Copy)`,
              description: original.description,
              rules: original.smart_rules
            });
            results.push(duplicate);
          } else if (updates) {
            const updated = await mockCollectionsApi.updateSmartCollection(id, updates);
            results.push(updated);
          }
        }
        return results;
      }
    );
  }
}

export const collectionsApi = new CollectionsApi();