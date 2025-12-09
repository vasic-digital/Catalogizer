import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  ArrowLeft, 
  Save, 
  X, 
  Plus, 
  Trash2, 
  Search,
  Filter,
  Shuffle,
  Play,
  Music,
  Film,
  Image,
  FileText
} from 'lucide-react';

import { PageHeader } from '../components/layout/PageHeader';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { Tabs } from '../components/ui/Tabs';
import { PlaylistGrid } from '../components/playlists/PlaylistGrid';
import { PlaylistManager } from '../components/playlists/PlaylistManager';
import { PlaylistPlayer } from '../components/playlists/PlaylistPlayer';
import { PlaylistItemComponent } from '../components/playlists/PlaylistItem';
import { playlistsApi } from '../lib/playlistsApi';
import { usePlaylists } from '../hooks/usePlaylists';
import { Playlist, PlaylistItem, CreatePlaylistRequest, UpdatePlaylistRequest, flattenPlaylistItem, getMediaIconWithMap } from '../types/playlists';
import { toast } from 'react-hot-toast';

const MEDIA_TYPE_OPTIONS = [
  { value: 'all', label: 'All Media' },
  { value: 'music', label: 'Music' },
  { value: 'video', label: 'Video' },
  { value: 'image', label: 'Images' },
  { value: 'document', label: 'Documents' }
];

const MEDIA_TYPE_ICONS = {
  music: Music,
  video: Film,
  image: Image,
  document: FileText,
};

interface CreatePlaylistFormData extends Omit<CreatePlaylistRequest, 'items'> {
  selectedItems: PlaylistItem[];
}

export const PlaylistsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'all' | 'my' | 'public' | 'favorites'>('all');
  const [isCreating, setIsCreating] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editingPlaylist, setEditingPlaylist] = useState<Playlist | null>(null);
  const [showPlayer, setShowPlayer] = useState(false);
  const [selectedPlaylist, setSelectedPlaylist] = useState<Playlist | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterMediaType, setFilterMediaType] = useState('all');

  // Form state for creating/editing playlists
  const [formData, setFormData] = useState<CreatePlaylistFormData>({
    name: '',
    description: '',
    is_public: false,
    selectedItems: []
  });

  const {
    playlists,
    isLoading,
    error,
    refetchPlaylists
  } = usePlaylists();

  // Filter playlists based on active tab
  const filteredPlaylists = React.useMemo(() => {
    let filtered = playlists;

    switch (activeTab) {
      case 'my':
        filtered = filtered.filter((p: Playlist) => !p.is_public);
        break;
      case 'public':
        filtered = filtered.filter((p: Playlist) => p.is_public);
        break;
      case 'favorites':
        // This would need favorites integration
        filtered = filtered.filter(() => false); // Placeholder
        break;
    }

    // Apply search filter
    if (searchQuery) {
      filtered = filtered.filter((p: Playlist) =>
        p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        p.description?.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Apply media type filter
    if (filterMediaType !== 'all') {
      filtered = filtered.filter((p: Playlist) => p.primary_media_type === filterMediaType);
    }

    return filtered;
  }, [playlists, activeTab, searchQuery, filterMediaType]);

  const handleCreatePlaylist = async () => {
    try {
      const playlistData: CreatePlaylistRequest = {
        name: formData.name,
        description: formData.description,
        is_public: formData.is_public,
        items: formData.selectedItems
      };

      const newPlaylist = await playlistsApi.createPlaylist(playlistData);
      toast.success(`Created playlist: ${newPlaylist.name}`);
      setIsCreating(false);
      resetForm();
      refetchPlaylists();
    } catch (error) {
      console.error('Failed to create playlist:', error);
      toast.error('Failed to create playlist');
    }
  };

  const handleUpdatePlaylist = async () => {
    if (!editingPlaylist) return;

    try {
      const updateData: UpdatePlaylistRequest = {
        name: formData.name,
        description: formData.description,
        is_public: formData.is_public
      };

      const updated = await playlistsApi.updatePlaylist(editingPlaylist.id, updateData);
      toast.success(`Updated playlist: ${updated.name}`);
      setIsEditing(false);
      setEditingPlaylist(null);
      resetForm();
      refetchPlaylists();
    } catch (error) {
      console.error('Failed to update playlist:', error);
      toast.error('Failed to update playlist');
    }
  };

  const handleDeletePlaylist = async (playlist: Playlist) => {
    if (!window.confirm(`Are you sure you want to delete "${playlist.name}"?`)) {
      return;
    }

    try {
      await playlistsApi.deletePlaylist(playlist.id);
      toast.success(`Deleted playlist: ${playlist.name}`);
      refetchPlaylists();
    } catch (error) {
      console.error('Failed to delete playlist:', error);
      toast.error('Failed to delete playlist');
    }
  };

  const handleEditPlaylist = (playlist: Playlist) => {
    setEditingPlaylist(playlist);
    setFormData({
      name: playlist.name,
      description: playlist.description || '',
      is_public: playlist.is_public,
      selectedItems: playlist.items || []
    });
    setIsEditing(true);
  };

  const handlePlayPlaylist = async (playlist: Playlist) => {
    try {
      const items = await playlistsApi.getPlaylistItems(playlist.id);
      setSelectedPlaylist({ ...playlist, items: items.items });
      setShowPlayer(true);
    } catch (error) {
      console.error('Failed to load playlist items:', error);
      toast.error('Failed to load playlist items');
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      is_public: false,
      selectedItems: []
    });
  };

  const handleFormChange = (field: keyof CreatePlaylistFormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const renderCreateEditForm = () => {
    const isEditMode = isEditing && editingPlaylist;
    const title = isEditMode ? 'Edit Playlist' : 'Create New Playlist';

    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
      >
        <motion.div
          initial={{ y: 20 }}
          animate={{ y: 0 }}
          className="bg-white dark:bg-gray-900 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto"
        >
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                {title}
              </h2>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setIsCreating(false);
                  setIsEditing(false);
                  setEditingPlaylist(null);
                  resetForm();
                }}
              >
                <X className="w-4 h-4" />
              </Button>
            </div>
          </div>

          <div className="p-6 space-y-6">
            {/* Basic Info */}
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Playlist Name *
                </label>
                <Input
                  value={formData.name}
                  onChange={(e) => handleFormChange('name', e.target.value)}
                  placeholder="Enter playlist name"
                  className="w-full"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Description
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => handleFormChange('description', e.target.value)}
                  placeholder="Enter playlist description (optional)"
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:text-white"
                />
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="is_public"
                  checked={formData.is_public}
                  onChange={(e) => handleFormChange('is_public', e.target.checked)}
                  className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                />
                <label htmlFor="is_public" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
                  Make this playlist public
                </label>
              </div>
            </div>

            {/* Selected Items */}
            {formData.selectedItems.length > 0 && (
              <div>
                <div className="flex items-center justify-between mb-3">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                    Selected Items ({formData.selectedItems.length})
                  </label>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleFormChange('selectedItems', [])}
                    className="text-red-600 hover:text-red-700"
                  >
                    Clear All
                  </Button>
                </div>
                <div className="max-h-48 overflow-y-auto space-y-2 border border-gray-200 dark:border-gray-700 rounded-lg p-3">
                  {formData.selectedItems.map((item, index) => {
                    const flattenedItem = flattenPlaylistItem(item);
                    const Icon = getMediaIconWithMap(flattenedItem.media_type);
                    return (
                      <div key={`${flattenedItem.item_id}-${index}`} className="flex items-center gap-3 p-2 bg-gray-50 dark:bg-gray-800 rounded">
                        <Icon className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                            {flattenedItem.title}
                          </p>
                          {flattenedItem.artist && (
                            <p className="text-xs text-gray-600 dark:text-gray-400 truncate">
                              {flattenedItem.artist}
                            </p>
                          )}
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            const newItems = formData.selectedItems.filter((_, i) => i !== index);
                            handleFormChange('selectedItems', newItems);
                          }}
                          className="text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="w-3 h-3" />
                        </Button>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Item Search/Selection */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Add Items to Playlist
              </label>
              <div className="flex gap-2">
                <Input
                  placeholder="Search media items..."
                  className="flex-1"
                  // TODO: Implement media item search
                />
                <Button variant="outline">
                  <Search className="w-4 h-4" />
                </Button>
              </div>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Media item search will be available in the next update
              </p>
            </div>
          </div>

          <div className="p-6 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
            <Button
              variant="outline"
              onClick={() => {
                setIsCreating(false);
                setIsEditing(false);
                setEditingPlaylist(null);
                resetForm();
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={isEditMode ? handleUpdatePlaylist : handleCreatePlaylist}
              disabled={!formData.name.trim()}
              className="flex items-center gap-2"
            >
              <Save className="w-4 h-4" />
              {isEditMode ? 'Update' : 'Create'} Playlist
            </Button>
          </div>
        </motion.div>
      </motion.div>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <PageHeader
        title="Playlists"
        subtitle="Organize and manage your media collections"
        breadcrumbs={[
          { label: 'Home', href: '/' },
          { label: 'Playlists' }
        ]}
      />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Tabs */}
        <Tabs
          tabs={[
            { id: 'all', label: 'All Playlists' },
            { id: 'my', label: 'My Playlists' },
            { id: 'public', label: 'Public' },
            { id: 'favorites', label: 'Favorites' }
          ]}
          activeTab={activeTab}
          onChangeTab={(tab) => setActiveTab(tab as any)}
        />

        {/* Filters and Controls */}
        <div className="mt-6 flex flex-col lg:flex-row gap-4 items-start lg:items-center justify-between">
          <div className="flex flex-col sm:flex-row gap-4 flex-1">
            {/* Search */}
            <div className="relative max-w-md">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                placeholder="Search playlists..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Media Type Filter */}
            <Select
              value={filterMediaType}
              onChange={setFilterMediaType}
              options={MEDIA_TYPE_OPTIONS}
              className="w-40"
            />
          </div>

          {/* Actions */}
          <Button
            onClick={() => setIsCreating(true)}
            className="flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Create Playlist
          </Button>
        </div>

        {/* Playlists Grid */}
        <div className="mt-8">
          <PlaylistGrid
            onCreatePlaylist={() => setIsCreating(true)}
            onEditPlaylist={handleEditPlaylist}
          />
        </div>

        {/* Create/Edit Modal */}
        <AnimatePresence>
          {(isCreating || isEditing) && renderCreateEditForm()}
        </AnimatePresence>

        {/* Player Modal */}
        <AnimatePresence>
          {showPlayer && selectedPlaylist && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
              onClick={() => setShowPlayer(false)}
            >
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.95 }}
                transition={{ duration: 0.2 }}
                className="bg-white dark:bg-gray-900 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto"
                onClick={(e) => e.stopPropagation()}
              >
                <PlaylistPlayer
                  playlist={selectedPlaylist}
                  items={selectedPlaylist.items || []}
                  onClose={() => setShowPlayer(false)}
                />
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
};