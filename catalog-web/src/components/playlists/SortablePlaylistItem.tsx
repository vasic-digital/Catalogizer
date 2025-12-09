import React from 'react';
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
  FileText
} from 'lucide-react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Button } from '../ui/Button';
import { PlaylistItem, flattenPlaylistItem, getMediaIconWithMap } from '../../types/playlists';
import { FavoriteToggle } from '../favorites/FavoriteToggle';

interface SortablePlaylistItemProps {
  item: PlaylistItem;
  index: number;
  isPlaying: boolean;
  isCurrent: boolean;
  showActions?: boolean;
  onClick?: (item: PlaylistItem, index: number) => void;
  onPlay?: (item: PlaylistItem, index: number) => void;
  onPause?: () => void;
  onRemove?: (item: PlaylistItem, index: number) => void;
  onToggleFavorite?: (itemId: string, itemType: string) => void;
  isFavorite?: boolean;
}

export const SortablePlaylistItem: React.FC<SortablePlaylistItemProps> = ({
  item,
  index,
  isPlaying,
  isCurrent,
  showActions = true,
  onClick,
  onPlay,
  onPause,
  onRemove,
  onToggleFavorite,
  isFavorite,
}) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: item.id });

  const [showDropdown, setShowDropdown] = React.useState(false);
  const [imageError, setImageError] = React.useState(false);
  const [isHovered, setIsHovered] = React.useState(false);

  const flattenedItem = flattenPlaylistItem(item);
  const Icon = getMediaIconWithMap(flattenedItem.media_type);

  const handleItemClick = () => {
    if (onClick) {
      onClick(item, index);
    }
  };

  const handlePlayClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (isPlaying && isCurrent) {
      if (onPause) onPause();
    } else {
      if (onPlay) onPlay(item, index);
    }
  };

  const handleRemoveClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (onRemove) onRemove(item, index);
  };

  const handleToggleFavorite = () => {
    if (onToggleFavorite) {
      onToggleFavorite(item.id, 'playlist_item');
    }
  };

  const duration = flattenedItem.duration ? parseInt(String(flattenedItem.duration)) : null;
  const formattedDuration = duration ? formatDuration(duration) : null;

  function formatDuration(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    } else {
      return `${minutes}:${secs.toString().padStart(2, '0')}`;
    }
  }

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <motion.div
      ref={setNodeRef}
      style={style}
      layout
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      whileHover={{ scale: 1.02 }}
      className={`flex items-center gap-4 p-4 rounded-lg border transition-all cursor-pointer ${
        isCurrent 
          ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800' 
          : 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-600'
      } ${isDragging ? 'cursor-grabbing' : 'cursor-grab'}`}
      onClick={handleItemClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      {...attributes}
    >
      {/* Drag Handle */}
      <div 
        className="flex items-center justify-center text-gray-400"
        {...listeners}
      >
        <GripVertical className="w-5 h-5" />
      </div>

      {/* Thumbnail/Image */}
      <div className="relative w-16 h-16 rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-700 flex-shrink-0">
        {flattenedItem.thumbnail_url && !imageError ? (
          <img
            src={flattenedItem.thumbnail_url}
            alt={flattenedItem.title}
            className="w-full h-full object-cover"
            onError={() => setImageError(true)}
          />
        ) : flattenedItem.file_path && flattenedItem.file_path.match(/\.(jpg|jpeg|png|gif|webp)$/i) && !imageError ? (
          <img
            src={flattenedItem.file_path}
            alt={flattenedItem.title}
            className="w-full h-full object-cover"
            onError={() => setImageError(true)}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <Icon className="w-8 h-8 text-gray-400" />
          </div>
        )}
        
        {/* Play/Pause Button Overlay */}
        {(isHovered || isCurrent) && (
          <div className="absolute inset-0 bg-black bg-opacity-40 flex items-center justify-center">
            <Button
              size="sm"
              variant="ghost"
              onClick={handlePlayClick}
              className="rounded-full w-8 h-8 bg-white/90 hover:bg-white text-black"
            >
              {isPlaying && isCurrent ? (
                <Pause className="w-3 h-3" />
              ) : (
                <Play className="w-3 h-3" />
              )}
            </Button>
          </div>
        )}
      </div>

      {/* Item Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <h3 className="font-medium text-gray-900 dark:text-white truncate">
            {flattenedItem.title}
          </h3>
          {isCurrent && isPlaying && (
            <div className="flex items-center gap-1">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-xs text-green-600 dark:text-green-400 font-medium">
                Playing
              </span>
            </div>
          )}
        </div>
        
        <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 mt-1">
          {flattenedItem.artist && (
            <span className="truncate">{flattenedItem.artist}</span>
          )}
          {flattenedItem.artist && formattedDuration && (
            <span>•</span>
          )}
          {formattedDuration && (
            <span>{formattedDuration}</span>
          )}
          {flattenedItem.media_type && (
            <>
              {(flattenedItem.artist || formattedDuration) && (
                <span>•</span>
              )}
              <span className="text-xs uppercase">{flattenedItem.media_type}</span>
            </>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        {/* Favorite Toggle */}
        {onToggleFavorite && (
          <FavoriteToggle
            mediaId={item.media_id}
            className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          />
        )}

        {/* Remove Button */}
        {onRemove && (
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRemoveClick}
            className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:text-red-400 dark:hover:text-red-300 dark:hover:bg-red-900/20 p-2"
          >
            <X className="w-4 h-4" />
          </Button>
        )}

        {/* Dropdown Menu */}
        {showActions && (
          <div className="relative">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowDropdown(!showDropdown)}
              className="p-2"
            >
              <MoreHorizontal className="w-4 h-4" />
            </Button>
            
            <AnimatePresence>
              {showDropdown && (
                <motion.div
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  className="absolute right-0 top-full mt-1 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50"
                >
                  <button
                    className="w-full text-left px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                    onClick={() => {
                      // Add to playlist
                      setShowDropdown(false);
                    }}
                  >
                    Add to Playlist
                  </button>
                  <button
                    className="w-full text-left px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                    onClick={() => {
                      // Share
                      setShowDropdown(false);
                    }}
                  >
                    Share
                  </button>
                  <button
                    className="w-full text-left px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                    onClick={() => {
                      // Download
                      setShowDropdown(false);
                    }}
                  >
                    Download
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        )}
      </div>
    </motion.div>
  );
};