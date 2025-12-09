import React, { useState } from 'react';
import { motion } from 'framer-motion';
import {
  Plus,
  Search,
  Filter,
  Grid,
  List,
  Heart,
  Clock,
  BarChart3,
  Share,
  MoreHorizontal,
  Edit,
  Trash2,
  Copy,
  Download,
  Eye
} from 'lucide-react';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { Tabs } from '../components/ui/Tabs';
import { Card } from '../components/ui/Card';
import { SmartCollectionBuilder } from '../components/collections/SmartCollectionBuilder';
import { useCollections } from '../hooks/useCollections';
import { SmartCollection } from '../types/collections';
import { toast } from 'react-hot-toast';

const COLLECTIONS_TABS = [
  { id: 'all', label: 'All Collections' },
  { id: 'smart', label: 'Smart Collections' },
  { id: 'manual', label: 'Manual Collections' },
  { id: 'favorites', label: 'Favorites' },
];

const MEDIA_TYPE_OPTIONS = [
  { value: 'all', label: 'All Media' },
  { value: 'music', label: 'Music' },
  { value: 'video', label: 'Video' },
  { value: 'image', label: 'Images' },
  { value: 'document', label: 'Documents' }
];

const VIEW_OPTIONS = [
  { value: 'grid', label: 'Grid View', icon: Grid },
  { value: 'list', label: 'List View', icon: List },
];

const SORT_OPTIONS = [
  { value: 'name', label: 'Name' },
  { value: 'created_at', label: 'Date Created' },
  { value: 'updated_at', label: 'Date Updated' },
  { value: 'item_count', label: 'Item Count' },
];

export const Collections: React.FC = () => {
  const [activeTab, setActiveTab] = useState('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [filterMediaType, setFilterMediaType] = useState('all');
  const [sortBy, setSortBy] = useState('name');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [showSmartBuilder, setShowSmartBuilder] = useState(false);
  const [selectedCollection, setSelectedCollection] = useState<SmartCollection | null>(null);

  const {
    collections,
    isLoading,
    error,
    refetchCollections,
    createCollection,
    updateCollection,
    deleteCollection,
    shareCollection,
    duplicateCollection,
    exportCollection,
    isSharing,
    isDuplicating,
    isExporting,
  } = useCollections();

  // Filter and sort collections
  const filteredCollections = React.useMemo(() => {
    let filtered = [...collections];

    // Apply tab filter
    switch (activeTab) {
      case 'smart':
        filtered = filtered.filter((c: SmartCollection) => c.is_smart);
        break;
      case 'manual':
        filtered = filtered.filter((c: SmartCollection) => !c.is_smart);
        break;
      case 'favorites':
        // This would need favorites integration
        filtered = filtered.filter(() => false); // Placeholder
        break;
    }

    // Apply search filter
    if (searchQuery) {
      filtered = filtered.filter((c: SmartCollection) =>
        c.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        c.description?.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Apply media type filter
    if (filterMediaType !== 'all') {
      filtered = filtered.filter((c: SmartCollection) => c.primary_media_type === filterMediaType);
    }

    // Apply sorting
    filtered.sort((a: SmartCollection, b: SmartCollection) => {
      const aValue = a[sortBy as keyof SmartCollection];
      const bValue = b[sortBy as keyof SmartCollection];
      
      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return aValue.localeCompare(bValue);
      }
      if (typeof aValue === 'number' && typeof bValue === 'number') {
        return aValue - bValue;
      }
      return 0;
    });

    return filtered;
  }, [collections, activeTab, searchQuery, filterMediaType, sortBy]);

  const handleCreateSmartCollection = () => {
    setShowSmartBuilder(true);
  };

  const handleSaveSmartCollection = async (name: string, description: string, rules: any[]) => {
    try {
      await createCollection({
        collection: {
          name,
          description,
          is_public: false,
          is_smart: true,
          smart_rules: rules,
        }
      });
      setShowSmartBuilder(false);
    } catch (error) {
      console.error('Failed to create smart collection:', error);
    }
  };

  const handleShareCollection = async (collection: SmartCollection) => {
    try {
      await shareCollection({
        id: collection.id,
        shareRequest: {
          can_view: true,
          can_comment: false,
          can_download: false,
        }
      });
    } catch (error) {
      console.error('Failed to share collection:', error);
    }
  };

  const handleDuplicateCollection = async (collection: SmartCollection) => {
    try {
      await duplicateCollection({
        id: collection.id,
        newName: `${collection.name} (Copy)`
      });
    } catch (error) {
      console.error('Failed to duplicate collection:', error);
    }
  };

  const handleExportCollection = async (collection: SmartCollection, format: 'json' | 'csv' | 'm3u') => {
    try {
      await exportCollection({
        id: collection.id,
        format
      });
    } catch (error) {
      console.error('Failed to export collection:', error);
    }
  };

  const handleDeleteCollection = async (collection: SmartCollection) => {
    if (window.confirm(`Are you sure you want to delete "${collection.name}"? This action cannot be undone.`)) {
      try {
        await deleteCollection({
          id: collection.id
        });
      } catch (error) {
        console.error('Failed to delete collection:', error);
      }
    }
  };

  const renderCollectionCard = (collection: SmartCollection) => (
    <motion.div
      key={collection.id}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ scale: 1.02 }}
      className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 cursor-pointer hover:shadow-md transition-shadow"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <h3 className="font-semibold text-gray-900 dark:text-white mb-1">
            {collection.name}
          </h3>
          {collection.description && (
            <p className="text-sm text-gray-600 dark:text-gray-400 line-clamp-2">
              {collection.description}
            </p>
          )}
        </div>
        
        <div className="flex items-center gap-1">
          {collection.is_smart && (
            <div className="w-6 h-6 bg-purple-100 dark:bg-purple-900 rounded-full flex items-center justify-center">
              <Clock className="w-3 h-3 text-purple-600 dark:text-purple-400" />
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
        <span>{collection.item_count} items</span>
        <span>{new Date(collection.created_at).toLocaleDateString()}</span>
      </div>

      <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleShareCollection(collection)}
            disabled={isSharing}
            title="Share collection"
          >
            <Share className="w-4 h-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleDuplicateCollection(collection)}
            disabled={isDuplicating}
            title="Duplicate collection"
          >
            <Copy className="w-4 h-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleExportCollection(collection, 'json')}
            disabled={isExporting}
            title="Export collection"
          >
            <Download className="w-4 h-4" />
          </Button>
        </div>
        
        <Button
          variant="ghost"
          size="sm"
          onClick={() => handleDeleteCollection(collection)}
          className="text-red-600 hover:text-red-700"
          title="Delete collection"
        >
          <Trash2 className="w-4 h-4" />
        </Button>
      </div>
    </motion.div>
  );

  const renderCollectionListItem = (collection: SmartCollection) => (
    <motion.div
      key={collection.id}
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md transition-shadow"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 flex-1">
          <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-pink-600 rounded-lg flex items-center justify-center">
            {collection.is_smart ? (
              <Clock className="w-6 h-6 text-white" />
            ) : (
              <Grid className="w-6 h-6 text-white" />
            )}
          </div>
          
          <div className="flex-1">
            <h3 className="font-semibold text-gray-900 dark:text-white mb-1">
              {collection.name}
            </h3>
            {collection.description && (
              <p className="text-sm text-gray-600 dark:text-gray-400 line-clamp-1">
                {collection.description}
              </p>
            )}
          </div>
          
          <div className="text-right">
            <div className="text-lg font-bold text-gray-900 dark:text-white">
              {collection.item_count.toLocaleString()}
            </div>
            <div className="text-xs text-gray-500 dark:text-gray-400">items</div>
          </div>
        </div>

        <div className="flex items-center gap-2 ml-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleShareCollection(collection)}
            disabled={isSharing}
            title="Share collection"
          >
            <Share className="w-4 h-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleDuplicateCollection(collection)}
            disabled={isDuplicating}
            title="Duplicate collection"
          >
            <Copy className="w-4 h-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleDeleteCollection(collection)}
            className="text-red-600 hover:text-red-700"
            title="Delete collection"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </motion.div>
  );

  if (showSmartBuilder) {
    return (
      <div className="max-w-4xl mx-auto">
        <SmartCollectionBuilder
          onSave={handleSaveSmartCollection}
          onCancel={() => setShowSmartBuilder(false)}
          className="mb-6"
        />
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Collections
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Organize your media with smart and manual collections
        </p>
      </div>

      {/* Tabs */}
      <Tabs
        tabs={COLLECTIONS_TABS}
        activeTab={activeTab}
        onChangeTab={setActiveTab}
        className="mb-6"
      />

      {/* Controls */}
      <div className="mb-6 flex flex-col lg:flex-row gap-4 items-start lg:items-center justify-between">
        <div className="flex flex-col sm:flex-row gap-4 flex-1">
          {/* Search */}
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <Input
              placeholder="Search collections..."
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

          {/* Sort */}
          <Select
            value={sortBy}
            onChange={setSortBy}
            options={SORT_OPTIONS}
            className="w-40"
          />
        </div>

        <div className="flex items-center gap-2">
          {/* View Mode Toggle */}
          <div className="flex items-center bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
            {VIEW_OPTIONS.map((option) => {
              const IconComponent = option.icon;
              return (
                <button
                  key={option.value}
                  onClick={() => setViewMode(option.value as 'grid' | 'list')}
                  className={`p-2 rounded ${
                    viewMode === option.value
                      ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                      : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
                  }`}
                  title={option.label}
                >
                  <IconComponent className="w-4 h-4" />
                </button>
              );
            })}
          </div>

          {/* Create Actions */}
          <Button
            onClick={handleCreateSmartCollection}
            className="flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Smart Collection
          </Button>
        </div>
      </div>

      {/* Collections Display */}
      <div className="min-h-96">
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : filteredCollections.length === 0 ? (
          <div className="text-center py-12">
            <div className="w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-4">
              <Grid className="w-8 h-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              No collections found
            </h3>
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              {searchQuery || filterMediaType !== 'all' || activeTab !== 'all'
                ? 'Try adjusting your search or filters'
                : 'Create your first collection to get started'
              }
            </p>
            {!searchQuery && filterMediaType === 'all' && activeTab === 'all' && (
              <Button
                onClick={handleCreateSmartCollection}
                className="flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                Create Smart Collection
              </Button>
            )}
          </div>
        ) : viewMode === 'grid' ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredCollections.map(renderCollectionCard)}
          </div>
        ) : (
          <div className="space-y-4">
            {filteredCollections.map(renderCollectionListItem)}
          </div>
        )}
      </div>
    </div>
  );
};