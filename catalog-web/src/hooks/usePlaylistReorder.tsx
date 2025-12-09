import { useMutation, useQueryClient } from '@tanstack/react-query';
import { playlistApi } from '../lib/playlistsApi';
import { toast } from 'react-hot-toast';
import type { PlaylistItem } from '../types/playlists';

export const usePlaylistReorder = () => {
  const queryClient = useQueryClient();

  const reorderPlaylistMutation = useMutation({
    mutationFn: async ({ 
      playlistId, 
      items 
    }: { 
      playlistId: string; 
      items: PlaylistItem[] 
    }) => {
      // Create ordered array of item IDs
      const itemIds = items.map(item => item.id);
      
      // Call API to reorder playlist
      await playlistApi.reorderPlaylist(playlistId, itemIds);
      
      return { playlistId, items };
    },
    onMutate: async ({ playlistId, items }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['playlists'] });
      await queryClient.cancelQueries({ queryKey: ['playlist', playlistId] });

      // Snapshot the previous value
      const previousPlaylists = queryClient.getQueryData(['playlists']);
      
      // Optimistically update the playlist in cache
      queryClient.setQueryData(['playlists'], (old: any) => {
        if (!old) return old;
        
        return {
          ...old,
          playlists: old.playlists.map((playlist: any) => 
            playlist.id === playlistId 
              ? { ...playlist, items }
              : playlist
          )
        };
      });

      queryClient.setQueryData(['playlist', playlistId], (old: any) => {
        if (!old) return old;
        return {
          ...old,
          items
        };
      });

      return { previousPlaylists };
    },
    onError: (error, variables, context) => {
      // Rollback on error
      if (context?.previousPlaylists) {
        queryClient.setQueryData(['playlists'], context.previousPlaylists);
      }
      
      toast.error('Failed to reorder playlist: ' + (error instanceof Error ? error.message : 'Unknown error'));
    },
    onSettled: (data, error, variables) => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: ['playlists'] });
      queryClient.invalidateQueries({ queryKey: ['playlist', variables.playlistId] });
    },
    onSuccess: () => {
      toast.success('Playlist reordered successfully');
    }
  });

  return reorderPlaylistMutation;
};