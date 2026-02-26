import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { playlistsApi } from '@/lib/playlistsApi'
import type { 
  PlaylistCreateRequest, 
  PlaylistUpdateRequest
} from '@/types/playlists'
import toast from 'react-hot-toast'

export const usePlaylists = (params?: {
  limit?: number
  offset?: number
  include_smart?: boolean
  type?: string
}) => {
  const queryClient = useQueryClient()

  const {
    data: playlistsData,
    isLoading: isLoadingPlaylists,
    error: playlistsError,
    refetch: refetchPlaylists
  } = useQuery({
    queryKey: ['playlists', params],
    queryFn: () => playlistsApi.getPlaylists(params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    cacheTime: 1000 * 60 * 10, // 10 minutes
  })

  const createPlaylistMutation = useMutation({
    mutationFn: (request: PlaylistCreateRequest) => playlistsApi.createPlaylist(request),
    onSuccess: (playlist) => {
      toast.success(`Playlist "${playlist.name}" created successfully`)
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
      return playlist
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to create playlist')
    }
  })

  const updatePlaylistMutation = useMutation({
    mutationFn: ({ id, request }: { id: string; request: PlaylistUpdateRequest }) => 
      playlistsApi.updatePlaylist(id, request),
    onSuccess: (playlist) => {
      toast.success(`Playlist "${playlist.name}" updated successfully`)
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
      queryClient.invalidateQueries({ queryKey: ['playlist', playlist.id] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to update playlist')
    }
  })

  const deletePlaylistMutation = useMutation({
    mutationFn: playlistsApi.deletePlaylist,
    onSuccess: (_, playlistId) => {
      toast.success('Playlist deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
      queryClient.removeQueries({ queryKey: ['playlist', playlistId] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to delete playlist')
    }
  })

  const addItemsMutation = useMutation({
    mutationFn: ({ playlistId, mediaIds }: { playlistId: string; mediaIds: number[] }) => 
      playlistsApi.addItemsToPlaylist(playlistId, mediaIds),
    onSuccess: (data, variables) => {
      toast.success(`Added ${data.added} items to playlist`)
      queryClient.invalidateQueries({ queryKey: ['playlist-items', variables.playlistId] })
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to add items to playlist')
    }
  })

  const removeItemMutation = useMutation({
    mutationFn: ({ playlistId, itemId }: { playlistId: string; itemId: string }) => 
      playlistsApi.removeFromPlaylist(playlistId, itemId),
    onSuccess: (_, variables) => {
      toast.success('Item removed from playlist')
      queryClient.invalidateQueries({ queryKey: ['playlist-items', variables.playlistId] })
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to remove item from playlist')
    }
  })

  const reorderItemsMutation = useMutation({
    mutationFn: ({ playlistId, itemOrders }: { playlistId: string; itemOrders: { id: string; position: number }[] }) => 
      playlistsApi.reorderPlaylistItems(playlistId, itemOrders),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['playlist-items', variables.playlistId] })
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || 'Failed to reorder playlist')
    }
  })

  const createPlaylist = (request: PlaylistCreateRequest) => {
    return createPlaylistMutation.mutateAsync(request)
  }

  const updatePlaylist = (id: string, request: PlaylistUpdateRequest) => {
    updatePlaylistMutation.mutate({ id, request })
  }

  const deletePlaylist = (id: string) => {
    if (window.confirm('Are you sure you want to delete this playlist? This action cannot be undone.')) {
      deletePlaylistMutation.mutate(id)
    }
  }

  const addItemsToPlaylist = (playlistId: string, mediaIds: number[]) => {
    addItemsMutation.mutate({ playlistId, mediaIds })
  }

  const removeFromPlaylist = (playlistId: string, itemId: string) => {
    removeItemMutation.mutate({ playlistId, itemId })
  }

  const reorderPlaylistItems = (playlistId: string, itemOrders: { id: string; position: number }[]) => {
    reorderItemsMutation.mutate({ playlistId, itemOrders })
  }

  return {
    playlists: playlistsData?.playlists || [],
    total: playlistsData?.total || 0,
    isLoading: isLoadingPlaylists,
    error: playlistsError,
    refetchPlaylists,
    createPlaylist,
    updatePlaylist,
    deletePlaylist,
    addItemsToPlaylist,
    removeFromPlaylist,
    reorderPlaylistItems,
    isCreating: createPlaylistMutation.isLoading,
    isUpdating: updatePlaylistMutation.isLoading,
    isDeleting: deletePlaylistMutation.isLoading,
    isAddingItems: addItemsMutation.isLoading
  }
}

// Hook for playlist items
export const usePlaylistItems = (playlistId: string, params?: {
  limit?: number
  offset?: number
  sort_by?: 'position' | 'added_at' | 'title' | 'duration'
  sort_order?: 'asc' | 'desc'
}) => {
  const {
    data: playlistItemsData,
    isLoading: isLoadingItems,
    error: itemsError,
    refetch: refetchItems
  } = useQuery({
    queryKey: ['playlist-items', playlistId, params],
    queryFn: () => playlistsApi.getPlaylistItems(playlistId, params),
    staleTime: 1000 * 60 * 5, // 5 minutes
    cacheTime: 1000 * 60 * 10, // 10 minutes,
    enabled: !!playlistId
  })

  return {
    items: playlistItemsData?.items || [],
    total: playlistItemsData?.total || 0,
    playlist: playlistItemsData?.playlist,
    isLoading: isLoadingItems,
    error: itemsError,
    refetchItems
  }
}

// Hook for playlist analytics
export const usePlaylistAnalytics = (playlistId: string) => {
  return useQuery({
    queryKey: ['playlist-analytics', playlistId],
    queryFn: () => playlistsApi.getPlaylistAnalytics(playlistId),
    staleTime: 1000 * 60 * 10, // 10 minutes
    cacheTime: 1000 * 60 * 30, // 30 minutes
    enabled: !!playlistId
  })
}