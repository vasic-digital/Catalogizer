import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  GripVertical,
  Play,
  Pause,
  MoreHorizontal,
  X,
  Music,
  Film,
  Image,
  FileText,
  Clock,
  Star,
  Heart
} from 'lucide-react';

import { Button } from '../ui/Button';
import { PlaylistItem, flattenPlaylistItem, getMediaIconWithMap } from '../../types/playlists';
import { MediaPlayer } from '../media/MediaPlayer';
import { FavoriteToggle } from '../favorites/FavoriteToggle';
import { usePlayerState } from '../../hooks/usePlayerState';

interface PlaylistItemComponentProps {
  item: PlaylistItem;
  index: number;
  isPlaying: boolean;
  isCurrent: boolean;
  isDraggable?: boolean;
  showActions?: boolean;
  onClick?: (item: PlaylistItem, index: number) => void;
  onPlay?: (item: PlaylistItem, index: number) => void;
  onPause?: () => void;
  onRemove?: (item: PlaylistItem, index: number) => void;
  onToggleFavorite?: (itemId: string, itemType: string) => void;
  isFavorite?: boolean;
  className?: string;
}

const MEDIA_TYPE_ICONS = {
  music: Music,
  video: Film,
  image: Image,
  document: FileText,
};

const DURATION_FORMATTER = new Intl.DateTimeFormat('en-US', {
  minute: '2-digit',
  second: '2-digit'
});

export const PlaylistItemComponent: React.FC<PlaylistItemComponentProps> = ({
  item,
  index,
  isPlaying,
  isCurrent,
  isDraggable = true,
  showActions = true,
  onClick,
  onPlay,
  onPause,
  onRemove,
  onToggleFavorite,
  isFavorite = false,
  className = ''
}) => {
  const [showDropdown, setShowDropdown] = useState(false);
  const [imageError, setImageError] = useState(false);
  const [isHovered, setIsHovered] = useState(false);

  const flattenedItem = flattenPlaylistItem(item);
  const Icon = getMediaIconWithMap(flattenedItem.media_type);

  const handleItemClick = () => {
    if (onClick) {
      onClick(item, index);
    } else if (isCurrent && isPlaying) {
      onPause?.();
    } else {
      onPlay?.(item, index);
    }
  };

  const handlePlayPause = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isCurrent && isPlaying) {
      onPause?.();
    } else {
      onPlay?.(item, index);
    }
  };

  const handleRemove = (e: React.MouseEvent) => {
    e.stopPropagation();
    onRemove?.(item, index);
  };

  const handleFavorite = (e: React.MouseEvent) => {
    e.stopPropagation();
    onToggleFavorite?.(flattenedItem.item_id, flattenedItem.media_type);
  };

  const formatDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${minutes}:${secs.toString().padStart(2, '0')}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  };

  const getQualityBadge = () => {
    if (flattenedItem.media_type !== 'video' && flattenedItem.media_type !== 'music') return null;

    const quality = flattenedItem.quality || 'standard';
    const colors = {
      sd: 'bg-gray-500',
      hd: 'bg-blue-500',
      '4k': 'bg-purple-500',
      high: 'bg-green-500',
      standard: 'bg-gray-500',
      low: 'bg-orange-500'
    };

    return (
      <span
        className={`text-xs px-2 py-1 rounded-full text-white ${colors[quality as keyof typeof colors] || 'bg-gray-500'}`}
      >
        {quality.toUpperCase()}
      </span>
    );
  };

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      whileHover={{ y: -2 }}
      className={`group flex items-center gap-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-all duration-200 ${isCurrent ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-300 dark:border-blue-700' : ''} ${className}`}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={handleItemClick}
    >
      {/* Drag Handle */}
      {isDraggable && (
        <div className="cursor-grab active:cursor-grabbing opacity-0 group-hover:opacity-100 transition-opacity">
          <GripVertical className="w-4 h-4 text-gray-400" />
        </div>
      )}

      {/* Play/Pause Button */}
      <div className="relative">
        <Button
          variant="ghost"
          size="sm"
          onClick={handlePlayPause}
          className="opacity-0 group-hover:opacity-100 transition-opacity"
        >
          {isCurrent && isPlaying ? (
            <Pause className="w-4 h-4" />
          ) : (
            <Play className="w-4 h-4" />
          )}
        </Button>
        {isCurrent && isPlaying && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="w-2 h-2 bg-blue-600 rounded-full animate-pulse" />
          </div>
        )}
      </div>

      {/* Thumbnail/Image */}
      <div className="relative flex-shrink-0">
        {flattenedItem.thumbnail_url && !imageError ? (
          <img
            src={flattenedItem.thumbnail_url}
            alt={flattenedItem.title}
            className="w-12 h-12 rounded object-cover"
            onError={() => setImageError(true)}
          />
        ) : (
          <div className={`w-12 h-12 rounded flex items-center justify-center ${
            isCurrent 
              ? 'bg-blue-100 dark:bg-blue-900/50' 
              : 'bg-gradient-to-br from-blue-500 to-purple-600'
          }`}>
            <Icon className={`w-6 h-6 ${isCurrent ? 'text-blue-600 dark:text-blue-400' : 'text-white/80'}`} />
          </div>
        )}
        
        {/* Playing Indicator */}
        {isCurrent && isPlaying && (
          <div className="absolute inset-0 bg-black/40 rounded flex items-center justify-center">
            <div className="flex gap-1">
              <div className="w-1 h-3 bg-white rounded-full animate-pulse" />
              <div className="w-1 h-3 bg-white rounded-full animate-pulse" style={{ animationDelay: '0.2s' }} />
              <div className="w-1 h-3 bg-white rounded-full animate-pulse" style={{ animationDelay: '0.4s' }} />
            </div>
          </div>
        )}
      </div>

      {/* Item Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <h4 className="font-medium text-sm truncate text-gray-900 dark:text-white">
            {flattenedItem.title}
          </h4>
          {getQualityBadge()}
        </div>
        <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
          {flattenedItem.artist && (
            <span className="truncate">{flattenedItem.artist}</span>
          )}
          {flattenedItem.artist && flattenedItem.album && <span>•</span>}
          {flattenedItem.album && (
            <span className="truncate">{flattenedItem.album}</span>
          )}
          {flattenedItem.duration && (
            <span className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              {formatDuration(flattenedItem.duration)}
            </span>
          )}
        </div>
        {flattenedItem.description && (
          <p className="text-xs text-gray-600 dark:text-gray-400 mt-1 line-clamp-2">
            {flattenedItem.description}
          </p>
        )}
      </div>

      {/* Metadata */}
      <div className="flex items-center gap-2">
        {/* File Size */}
        {flattenedItem.file_size && (
          <span className="text-xs text-gray-500 dark:text-gray-400">
            {(flattenedItem.file_size / 1024 / 1024).toFixed(1)} MB
          </span>
        )}

        {/* Quality Badge */}
        {flattenedItem.quality && (
          <span className={`text-xs px-2 py-1 rounded ${
            flattenedItem.quality === '4k' ? 'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300' :
            flattenedItem.quality === 'hd' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300' :
            'bg-gray-100 text-gray-800 dark:bg-gray-900/50 dark:text-gray-300'
          }`}>
            {flattenedItem.quality.toUpperCase()}
          </span>
        )}

        {/* Rating */}
        {flattenedItem.rating && flattenedItem.rating > 0 && (
          <div className="flex items-center gap-1">
            <Star className="w-3 h-3 text-yellow-500 fill-current" />
            <span className="text-xs text-gray-600 dark:text-gray-400">
              {flattenedItem.rating.toFixed(1)}
            </span>
          </div>
        )}
      </div>

      {/* Actions */}
      {showActions && (
        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <FavoriteToggle
            mediaId={parseInt(flattenedItem.item_id)}
            size="sm"
            className="p-1"
          />
          
          <div className="relative">
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                setShowDropdown(!showDropdown);
              }}
              className="p-1"
            >
              <MoreHorizontal className="w-4 h-4" />
            </Button>

            {/* Dropdown Menu */}
            <AnimatePresence>
              {showDropdown && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.95, y: -10 }}
                  animate={{ opacity: 1, scale: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.95, y: -10 }}
                  transition={{ duration: 0.15 }}
                  className="absolute right-0 mt-1 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-10"
                >
                  <div className="py-1">
                    <button
                      className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
                      onClick={(e) => {
                        e.stopPropagation();
                        handlePlayPause(e);
                        setShowDropdown(false);
                      }}
                    >
                      {isCurrent && isPlaying ? (
                        <>
                          <Pause className="w-4 h-4" />
                          Pause
                        </>
                      ) : (
                        <>
                          <Play className="w-4 h-4" />
                          Play
                        </>
                      )}
                    </button>
                    <button
                      className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleFavorite(e);
                        setShowDropdown(false);
                      }}
                    >
                      <Heart className="w-4 h-4" />
                      {isFavorite ? 'Remove from Favorites' : 'Add to Favorites'}
                    </button>
                    <button
                      className="w-full px-4 py-2 text-left text-sm text-red-600 dark:text-red-400 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleRemove(e);
                        setShowDropdown(false);
                      }}
                    >
                      <X className="w-4 h-4" />
                      Remove
                    </button>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      )}
    </motion.div>
  );
};