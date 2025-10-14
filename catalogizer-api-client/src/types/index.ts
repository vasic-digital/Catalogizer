// Base types
export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
  status: number;
  success: boolean;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
  has_next: boolean;
  has_previous: boolean;
}

// Media types
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
  metadata?: Record<string, any>;
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
  sort_order?: 'asc' | 'desc';
  limit?: number;
  offset?: number;
}

export interface MediaStats {
  total_items: number;
  by_type: Record<string, number>;
  by_quality: Record<string, number>;
  total_size: number;
  recent_additions: number;
}

// Authentication types
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

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  first_name: string;
  last_name: string;
}

export interface UpdateProfileRequest {
  first_name?: string;
  last_name?: string;
  email?: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

// SMB types
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

export interface CreateSMBConfigRequest {
  name: string;
  host: string;
  port: number;
  share_name: string;
  username: string;
  password: string;
  domain?: string;
  mount_point: string;
}

// Playback types
export interface PlaybackProgress {
  media_id: number;
  position: number;
  duration: number;
  timestamp?: number;
}

export interface StreamInfo {
  url: string;
  mime_type: string;
  file_size: number;
  duration?: number;
}

// Download types
export interface DownloadJob {
  id: number;
  media_id: number;
  status: 'pending' | 'downloading' | 'completed' | 'failed' | 'paused' | 'cancelled';
  progress: number;
  file_path: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

// WebSocket types
export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

export interface DownloadProgressMessage extends WebSocketMessage {
  type: 'download_progress';
  data: {
    job_id: number;
    progress: number;
    status: string;
    error?: string;
  };
}

export interface ScanProgressMessage extends WebSocketMessage {
  type: 'scan_progress';
  data: {
    config_id: number;
    progress: number;
    status: string;
    found_items: number;
    error?: string;
  };
}

// Client configuration
export interface ClientConfig {
  baseURL: string;
  timeout?: number;
  retryAttempts?: number;
  retryDelay?: number;
  enableWebSocket?: boolean;
  webSocketURL?: string;
  headers?: Record<string, string>;
}

// Error types
export class CatalogizerError extends Error {
  constructor(
    message: string,
    public status?: number,
    public code?: string
  ) {
    super(message);
    this.name = 'CatalogizerError';
  }
}

export class AuthenticationError extends CatalogizerError {
  constructor(message = 'Authentication failed') {
    super(message, 401, 'AUTH_ERROR');
    this.name = 'AuthenticationError';
  }
}

export class NetworkError extends CatalogizerError {
  constructor(message = 'Network request failed') {
    super(message, 0, 'NETWORK_ERROR');
    this.name = 'NetworkError';
  }
}

export class ValidationError extends CatalogizerError {
  constructor(message = 'Validation failed') {
    super(message, 400, 'VALIDATION_ERROR');
    this.name = 'ValidationError';
  }
}

// Event types for the client
export interface ClientEvents {
  'auth:login': (user: User) => void;
  'auth:logout': () => void;
  'auth:token_refresh': (token: string) => void;
  'download:progress': (progress: DownloadProgressMessage['data']) => void;
  'scan:progress': (progress: ScanProgressMessage['data']) => void;
  'connection:open': () => void;
  'connection:close': () => void;
  'connection:error': (error: Error) => void;
}