export interface MediaItem {
  id: number;
  title: string;
  media_type: string;
  year?: number;
  description?: string;
  cover_image?: string;
  rating?: number;
  quality?: string;
  file_size?: number;
  duration?: number;
  directory_path: string;
  smb_path?: string;
  created_at: string;
  updated_at: string;
  external_metadata?: ExternalMetadata[];
  versions?: MediaVersion[];
  is_favorite?: boolean;
  watch_progress?: number;
  last_watched?: string;
}

export interface ExternalMetadata {
  id: number;
  media_id: number;
  provider: string;
  external_id: string;
  title: string;
  description?: string;
  year?: number;
  rating?: number;
  poster_url?: string;
  backdrop_url?: string;
  genres?: string[];
  cast?: string[];
  crew?: string[];
  metadata?: Record<string, string>;
  last_updated: string;
}

export interface MediaVersion {
  id: number;
  media_id: number;
  version: string;
  quality: string;
  file_path: string;
  file_size: number;
  codec?: string;
  resolution?: string;
  bitrate?: number;
  language?: string;
  frame_rate?: number;
  audio_channels?: number;
  sample_rate?: number;
}

export interface MediaSearchRequest {
  query?: string;
  media_type?: string;
  year_min?: number;
  year_max?: number;
  rating_min?: number;
  quality?: string;
  sort_by?: string;
  sort_order?: string;
  limit?: number;
  offset?: number;
}

export interface MediaSearchResponse {
  items: MediaItem[];
  total: number;
  limit: number;
  offset: number;
}

export interface MediaStats {
  total_items: number;
  by_type: Record<string, number>;
  by_quality: Record<string, number>;
  total_size: number;
  recent_additions: number;
}

export interface User {
  id: number;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  is_admin: boolean;
  permissions?: string[];
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

export interface AuthStatus {
  authenticated: boolean;
  user?: User;
  expires_at?: string;
}

export interface AppConfig {
  server_url?: string;
  auth_token?: string;
  theme: string;
  auto_start: boolean;
}

export interface SMBConfig {
  id: number;
  name: string;
  host: string;
  port: number;
  share_name: string;
  username: string;
  password: string;
  domain?: string;
  is_active: boolean;
  mount_point: string;
  created_at: string;
  updated_at: string;
}

export interface SMBStatus {
  config_id: number;
  is_connected: boolean;
  last_check: string;
  error_message?: string;
  mount_point?: string;
}

export interface PlaybackProgress {
  media_id: number;
  position: number;
  duration: number;
  timestamp: number;
}

export interface UiPlaybackProgress {
  mediaItemId: number;
  positionUnit: string;
  durationTotal: number | null;
  lastPosition: number;
  lastSessionAmount: number;
  totalReproductions: number;
  aggregateAmount: number;
  lastSessionEndedAt: string | null;
}

export interface UiPlaybackSession {
  id: number;
  positionUnit: string;
  startPosition: number;
  endPosition: number;
  totalAmount: number;
  startedAt: string;
  endedAt: string | null;
  completed: boolean;
}

export interface DownloadJob {
  id: number;
  media_id: number;
  status: 'pending' | 'downloading' | 'completed' | 'failed' | 'paused';
  progress: number;
  file_path: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export type MediaType =
  | 'movie'
  | 'tv_show'
  | 'documentary'
  | 'anime'
  | 'music'
  | 'audiobook'
  | 'podcast'
  | 'game'
  | 'software'
  | 'ebook'
  | 'training'
  | 'concert'
  | 'youtube_video'
  | 'sports'
  | 'news'
  | 'other';

export type QualityLevel =
  | 'cam'
  | 'ts'
  | 'dvdrip'
  | 'brrip'
  | '720p'
  | '1080p'
  | '4k'
  | 'hdr'
  | 'dolby_vision';

export type SortOption =
  | 'title'
  | 'year'
  | 'rating'
  | 'updated_at'
  | 'created_at'
  | 'file_size'
  | 'duration';

export type SortOrder = 'asc' | 'desc';

export type Theme = 'light' | 'dark' | 'system';