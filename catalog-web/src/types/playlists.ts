import { Music, Video, Image, FileText } from 'lucide-react'

export interface Playlist {
  id: string
  name: string
  description?: string
  user_id: number
  is_public: boolean
  is_smart: boolean
  thumbnail_url?: string
  item_count: number
  total_duration?: number
  created_at: string
  updated_at: string
  last_item_added?: string
  primary_media_type?: string
  items?: PlaylistItem[]
}

export interface PlaylistItem {
  id: string
  playlist_id: string
  media_id: number
  position: number
  media_item: {
    id: number
    title: string
    media_type: string
    year?: number
    cover_image?: string
    duration?: number
    rating?: number
    quality?: string
    file_path?: string
    thumbnail_url?: string
    artist?: string
    album?: string
    description?: string
    file_size?: number
  }
  added_at: string
}

// Helper properties for easier access
export interface PlaylistItemWithMedia extends PlaylistItem {
  item_id: string
  title: string
  media_type: string
  artist?: string
  album?: string
  description?: string
  duration?: number
  quality?: string
  rating?: number
  thumbnail_url?: string
  file_path?: string
  file_size?: number
}

// Helper function to flatten PlaylistItem properties
export const flattenPlaylistItem = (item: PlaylistItem): PlaylistItemWithMedia => {
  const { media_item, ...rest } = item
  return {
    ...rest,
    media_item, // Keep the original media_item
    item_id: media_item.id.toString(),
    title: media_item.title,
    media_type: media_item.media_type,
    artist: media_item.artist,
    album: media_item.album,
    description: media_item.description,
    duration: media_item.duration,
    quality: media_item.quality,
    rating: media_item.rating,
    thumbnail_url: media_item.thumbnail_url,
    file_path: media_item.file_path,
    file_size: media_item.file_size
  }
}

// Helper function to get media type icon name
export const getMediaIconName = (mediaType: string): string => {
  switch (mediaType.toLowerCase()) {
    case 'video':
      return 'video'
    case 'audio':
    case 'music':
      return 'music'
    case 'image':
      return 'image'
    default:
      return 'document'
  }
}

// Helper function to get media type icon component (for generic use)
export const getMediaIcon = (mediaType: string, iconMap: { [key: string]: any }) => {
  const iconName = getMediaIconName(mediaType)
  return iconMap[iconName] || iconMap.document
}

// Simplified helper function that uses a default icon map
export const getMediaIconWithMap = (mediaType: string) => {
  const iconName = getMediaIconName(mediaType)
  
  // Default icon mapping
  const defaultIcons = {
    music: Music,
    video: Video,
    image: Image,
    document: FileText
  }
  
  return defaultIcons[iconName as keyof typeof defaultIcons] || FileText
}

export interface PlaylistCreateRequest {
  name: string
  description?: string
  is_public?: boolean
  is_smart?: boolean
  smart_rules?: SmartPlaylistRule[]
  items?: PlaylistItem[]
}

export interface PlaylistUpdateRequest {
  name?: string
  description?: string
  is_public?: boolean
}

export interface SmartPlaylistRule {
  field: 'media_type' | 'year' | 'rating' | 'quality' | 'genre' | 'created_at'
  operator: 'equals' | 'not_equals' | 'greater_than' | 'less_than' | 'contains' | 'starts_with' | 'ends_with' | 'in' | 'not_in'
  value: string | number | string[]
  condition?: 'and' | 'or'
}

export interface PlaylistResponse {
  playlists: Playlist[]
  total: number
  limit: number
  offset: number
}

export interface PlaylistItemsResponse {
  items: PlaylistItem[]
  total: number
  playlist: Playlist
}

export interface PlaylistShareInfo {
  share_url: string
  share_token: string
  expires_at?: string
  permissions: {
    can_view: boolean
    can_comment: boolean
    can_download: boolean
  }
}

export interface PlaylistAnalytics {
  playlist_id: string
  total_plays: number
  unique_viewers: number
  average_completion_rate: number
  popular_items: {
    media_id: number
    title: string
    play_count: number
  }[]
  viewing_stats: {
    date: string
    plays: number
    viewers: number
  }[]
}

export const PLAYLIST_TYPES = [
  'user_created',
  'recently_played',
  'most_played',
  'favorites',
  'watch_later',
  'continue_watching',
  'recommended'
] as const

export type PlaylistType = typeof PLAYLIST_TYPES[number]

// Aliases for compatibility with components
export type PlaylistViewMode = 'grid' | 'list'
export type PlaylistSortBy = 'name' | 'name_desc' | 'created_at' | 'updated_at' | 'duration' | 'item_count'
export type CreatePlaylistRequest = PlaylistCreateRequest
export type UpdatePlaylistRequest = PlaylistUpdateRequest