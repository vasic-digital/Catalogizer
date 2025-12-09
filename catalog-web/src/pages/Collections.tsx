import React, { useState, useEffect } from 'react';
import { CollectionsManager } from '@/components/collections/CollectionsManager';
import { collectionsApi } from '@/lib/collectionsApi';
import { useQuery } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type { Collection } from '@/types/collections';

export const Collections: React.FC = () => {
  const [collections, setCollections] = useState<Collection[]>([]);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['collections'],
    queryFn: () => collectionsApi.getCollections(),
    staleTime: 1000 * 60 * 5,
  });

  useEffect(() => {
    if (data) {
      setCollections(data);
    }
  }, [data]);

  const handleCreateCollection = async (collectionData: any) => {
    try {
      const newCollection = await collectionsApi.createCollection(collectionData);
      setCollections(prev => [newCollection, ...prev]);
      toast.success('Collection created successfully');
    } catch (error) {
      toast.error(`Failed to create collection: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleUpdateCollection = async (id: string, updates: Partial<Collection>) => {
    try {
      await collectionsApi.updateCollection(id, updates);
      setCollections(prev => 
        prev.map(collection => 
          collection.id === id ? { ...collection, ...updates } : collection
        )
      );
      toast.success('Collection updated successfully');
    } catch (error) {
      toast.error(`Failed to update collection: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleDeleteCollection = async (id: string) => {
    try {
      await collectionsApi.deleteCollection(id);
      setCollections(prev => prev.filter(collection => collection.id !== id));
      toast.success('Collection deleted successfully');
    } catch (error) {
      toast.error(`Failed to delete collection: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handlePlayCollection = async (id: string) => {
    try {
      const collection = collections.find(c => c.id === id);
      if (collection) {
        toast.success(`Playing collection: ${collection.name}`);
        // Implementation would open a media player with collection items
      }
    } catch (error) {
      toast.error(`Failed to play collection: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Collections
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Organize your media into custom collections
        </p>
      </div>
      
      <CollectionsManager
        collections={collections}
        onCreateCollection={handleCreateCollection}
        onUpdateCollection={handleUpdateCollection}
        onDeleteCollection={handleDeleteCollection}
        onPlayCollection={handlePlayCollection}
      />
    </div>
  );
};