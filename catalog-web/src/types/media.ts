export interface MediaItem {
  id: number
  title: string
  media_type: string
  year?: number
  description?: string
  cover_image?: string
  rating?: number
  quality?: string
  file_size?: number
  duration?: number
  directory_path: string
  storage_root_name?: string
  storage_root_protocol?: string
  created_at: string
  updated_at: string
  external_metadata?: ExternalMetadata[]
  versions?: MediaVersion[]
}

export interface ExternalMetadata {
  id: number
  media_id: number
  provider: string
  external_id: string
  title: string
  description?: string
  year?: number
  rating?: number
  poster_url?: string
  backdrop_url?: string
  genres?: string[]
  cast?: string[]
  crew?: string[]
  metadata: Record<string, any>
  last_updated: string
}

export interface MediaVersion {
  id: number
  media_id: number
  version: string
  quality: string
  file_path: string
  file_size: number
  codec?: string
  resolution?: string
  bitrate?: number
  language?: string
}

export interface QualityInfo {
  overall_score: number
  video_quality?: number
  audio_quality?: number
  resolution: string
  bitrate?: number
  codec: string
  file_size: number
}

export interface MediaSearchRequest {
  query?: string
  media_type?: string
  year_min?: number
  year_max?: number
  rating_min?: number
  quality?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  limit?: number
  offset?: number
}

export interface MediaSearchResponse {
  items: MediaItem[]
  total: number
  limit: number
  offset: number
}

export const MEDIA_TYPES = [
  'movie',
  'tv_show',
  'tv_season',
  'tv_episode',
  'music_artist',
  'music_album',
  'song',
  'game',
  'software',
  'book',
  'comic',
] as const

export type MediaType = typeof MEDIA_TYPES[number]

// --- Entity types for structured media browsing ---

export interface MediaEntity {
  id: number
  media_type_id: number
  title: string
  original_title?: string
  year?: number
  description?: string
  genre?: string[]
  director?: string
  rating?: number
  runtime?: number
  language?: string
  status: string
  parent_id?: number
  season_number?: number
  episode_number?: number
  track_number?: number
  first_detected: string
  last_updated: string
}

export interface MediaEntityDetail extends MediaEntity {
  media_type: string
  file_count: number
  children_count: number
  external_metadata: EntityExternalMetadata[]
}

export interface EntityExternalMetadata {
  id: number
  media_item_id: number
  provider: string
  external_id: string
  data?: Record<string, any>
  rating?: number
  review_url?: string
  cover_url?: string
  trailer_url?: string
  last_fetched: string
}

export interface EntityUserMetadata {
  user_rating?: number
  watched_status?: string
  favorite?: boolean
  personal_notes?: string
  tags?: string[]
}

export interface MediaTypeInfo {
  id: number
  name: string
  description: string
  count: number
}

export interface EntityListResponse {
  items: MediaEntity[]
  total: number
  limit: number
  offset: number
}

export interface EntityStatsResponse {
  total_entities: number
  by_type: Record<string, number>
}

export interface EntityFile {
  id: number
  media_item_id: number
  file_id: number
  quality_info?: string
  language?: string
  is_primary: boolean
  created_at: string
}

export const QUALITY_LEVELS = [
  'cam',
  'ts',
  'dvdrip',
  'brrip',
  '720p',
  '1080p',
  '4k',
  'hdr',
  'dolby_vision',
] as const

export type QualityLevel = typeof QUALITY_LEVELS[number]

export interface StorageRoot {
  id: number
  name: string
  protocol: string
  enabled: boolean
  max_depth: number
  enable_duplicate_detection: boolean
  enable_metadata_extraction: boolean
  created_at: string
  updated_at: string
  last_scan_at?: string
}

export const SUPPORTED_PROTOCOLS = [
  'smb',
  'ftp',
  'nfs',
  'webdav',
  'local',
] as const

export type StorageProtocol = typeof SUPPORTED_PROTOCOLS[number]