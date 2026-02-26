import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { collectionsApi } from '../lib/collectionsApi';
import { CreateCollectionRequest, UpdateCollectionRequest, ShareCollectionRequest } from '../types/collections';
import { toast } from 'react-hot-toast';

export const useCollections = () => {
  const queryClient = useQueryClient();

  const {
    data: collections = [],
    isLoading,
    error,
    refetch: refetchCollections,
  } = useQuery({
    queryKey: ['collections'],
    queryFn: () => collectionsApi.getCollections(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const createCollectionMutation = useMutation({
    mutationFn: ({ collection }: { collection: CreateCollectionRequest }) =>
      collectionsApi.createCollection(collection),
    onSuccess: (newCollection) => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
      toast.success(`Created collection: ${newCollection.name}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create collection');
    },
  });

  const updateCollectionMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: UpdateCollectionRequest }) =>
      collectionsApi.updateCollection(id, updates),
    onSuccess: (updatedCollection) => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
      queryClient.invalidateQueries({ queryKey: ['collection', updatedCollection.id] });
      toast.success(`Updated collection: ${updatedCollection.name}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update collection');
    },
  });

  const deleteCollectionMutation = useMutation({
    mutationFn: ({ id }: { id: string }) => collectionsApi.deleteCollection(id),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
      queryClient.removeQueries({ queryKey: ['collection', id] });
      toast.success('Collection deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete collection');
    },
  });

  const refreshCollectionMutation = useMutation({
    mutationFn: ({ id }: { id: string }) => collectionsApi.refreshCollection(id),
    onSuccess: (refreshedCollection) => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
      queryClient.invalidateQueries({ queryKey: ['collection', refreshedCollection.id] });
      toast.success('Collection refreshed successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to refresh collection');
    },
  });

  const shareCollectionMutation = useMutation({
    mutationFn: ({ id, shareRequest }: { id: string; shareRequest: ShareCollectionRequest }) =>
      collectionsApi.shareCollection(id, shareRequest),
    onSuccess: async (shareInfo) => {
      try {
        await navigator.clipboard.writeText(shareInfo.share_url);
        toast.success('Share link copied to clipboard!');
      } catch (error) {
        toast.success('Collection shared successfully!');
        console.error('Failed to copy to clipboard:', error);
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to share collection');
    },
  });

  const duplicateCollectionMutation = useMutation({
    mutationFn: ({ id, newName }: { id: string; newName?: string }) =>
      collectionsApi.duplicateCollection(id, newName),
    onSuccess: (duplicatedCollection) => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
      toast.success(`Duplicated collection: ${duplicatedCollection.name}`);
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to duplicate collection');
    },
  });

  const exportCollectionMutation = useMutation({
    mutationFn: ({ id, format }: { id: string; format: 'json' | 'csv' | 'm3u' }) =>
      collectionsApi.exportCollection(id, format),
    onSuccess: (blob, { id, format }) => {
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `collection-${id}.${format}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      toast.success('Collection exported successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to export collection');
    },
  });

  const bulkDeleteCollectionsMutation = useMutation({
    mutationFn: ({ collectionIds }: { collectionIds: string[] }) =>
      collectionsApi.bulkDeleteCollections(collectionIds),
    onSuccess: (_, { collectionIds }) => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
      collectionIds.forEach(id => {
        queryClient.removeQueries({ queryKey: ['collection', id] });
      });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete collections');
    },
  });

  const bulkShareCollectionsMutation = useMutation({
    mutationFn: ({ collectionIds, shareRequest }: { collectionIds: string[]; shareRequest: ShareCollectionRequest }) =>
      collectionsApi.bulkShareCollections(collectionIds, shareRequest),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to share collections');
    },
  });

  const bulkExportCollectionsMutation = useMutation({
    mutationFn: ({ collectionIds, format }: { collectionIds: string[]; format: 'json' | 'csv' | 'm3u' }) =>
      collectionsApi.bulkExportCollections(collectionIds, format),
    onSuccess: (blob, { format }) => {
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `collections-bulk-export.${format}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      toast.success('Collections exported successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to export collections');
    },
  });

  const bulkUpdateCollectionsMutation = useMutation({
    mutationFn: ({ collectionIds, action, updates }: { collectionIds: string[]; action: string; updates?: UpdateCollectionRequest }) =>
      collectionsApi.bulkUpdateCollections(collectionIds, action, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['collections'] });
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update collections');
    },
  });

  return {
    collections,
    isLoading,
    error,
    refetchCollections,
    createCollection: createCollectionMutation.mutateAsync,
    updateCollection: updateCollectionMutation.mutateAsync,
    deleteCollection: deleteCollectionMutation.mutateAsync,
    refreshCollection: refreshCollectionMutation.mutateAsync,
    shareCollection: shareCollectionMutation.mutateAsync,
    duplicateCollection: duplicateCollectionMutation.mutateAsync,
    exportCollection: exportCollectionMutation.mutateAsync,
    bulkDeleteCollections: bulkDeleteCollectionsMutation.mutateAsync,
    bulkShareCollections: bulkShareCollectionsMutation.mutateAsync,
    bulkExportCollections: bulkExportCollectionsMutation.mutateAsync,
    bulkUpdateCollections: bulkUpdateCollectionsMutation.mutateAsync,
    isCreating: createCollectionMutation.isPending,
    isUpdating: updateCollectionMutation.isPending,
    isDeleting: deleteCollectionMutation.isPending,
    isRefreshing: refreshCollectionMutation.isPending,
    isSharing: shareCollectionMutation.isPending,
    isDuplicating: duplicateCollectionMutation.isPending,
    isExporting: exportCollectionMutation.isPending,
    isBulkDeleting: bulkDeleteCollectionsMutation.isPending,
    isBulkSharing: bulkShareCollectionsMutation.isPending,
    isBulkExporting: bulkExportCollectionsMutation.isPending,
    isBulkUpdating: bulkUpdateCollectionsMutation.isPending,
  };
};

export const useCollection = (id: string) => {
  const {
    data: collection,
    isLoading,
    error,
    refetch: refetchCollection,
  } = useQuery({
    queryKey: ['collection', id],
    queryFn: () => collectionsApi.getCollection(id),
    enabled: !!id,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const {
    data: collectionItems,
    isLoading: isLoadingItems,
    refetch: refetchItems,
  } = useQuery({
    queryKey: ['collection-items', id],
    queryFn: () => collectionsApi.getCollectionItems(id),
    enabled: !!id,
    staleTime: 2 * 60 * 1000, // 2 minutes
  });

  return {
    collection,
    collectionItems,
    isLoading,
    isLoadingItems,
    error,
    refetchCollection,
    refetchItems,
  };
};

export const useCollectionAnalytics = (id: string) => {
  const {
    data: analytics,
    isLoading,
    error,
    refetch: refetchAnalytics,
  } = useQuery({
    queryKey: ['collection-analytics', id],
    queryFn: () => collectionsApi.getCollectionAnalytics(id),
    enabled: !!id,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

  return {
    analytics,
    isLoading,
    error,
    refetchAnalytics,
  };
};

export const useSharedCollection = (shareId: string) => {
  const {
    data: collection,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['shared-collection', shareId],
    queryFn: () => collectionsApi.getSharedCollection(shareId),
    enabled: !!shareId,
    staleTime: 15 * 60 * 1000, // 15 minutes
  });

  return {
    collection,
    isLoading,
    error,
  };
};