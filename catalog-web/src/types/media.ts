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
  'music',
  'game',
  'software',
  'documentary',
  'concert',
  'training',
  'audiobook',
  'ebook',
  'podcast',
  'youtube_video',
  'adult_content',
  'anime',
  'sports',
  'news',
] as const

export type MediaType = typeof MEDIA_TYPES[number]

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