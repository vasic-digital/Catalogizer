export interface SubtitleSearchRequest {
  query?: string
  language?: string
  media_path?: string
  title?: string
  year?: number
  season?: number
  episode?: number
  providers?: string[]
  limit?: number
  offset?: number
}

export interface SubtitleSearchResult {
  id: string
  media_path?: string
  title: string
  year?: number
  season?: number
  episode?: number
  language: string
  language_name: string
  provider: string
  download_url: string
  rating?: number
  downloads?: number
  fps?: number
  format?: string
  release?: string
  hearing_impaired: boolean
  foreign_parts_only: boolean
  machine_translated: boolean
  upload_date?: string
}

export interface SubtitleSearchResponse {
  results: SubtitleSearchResult[]
  total: number
  limit: number
  offset: number
  query_time_ms: number
}

export interface SubtitleDownloadRequest {
  id: string
  media_path?: string
  language?: string
  encoding?: string
}

export interface SubtitleDownloadResponse {
  success: boolean
  subtitle_id?: string
  media_id?: number
  file_path?: string
  language?: string
  format?: string
  encoding?: string
  message?: string
  error?: string
}

export interface SubtitleTrack {
  id: string
  media_id: number
  language: string
  language_name: string
  provider: string
  file_path: string
  format: string
  encoding: string
  file_size: number
  created_at: string
  updated_at: string
  sync_offset?: number
  hearing_impaired: boolean
  foreign_parts_only: boolean
  machine_translated: boolean
  rating?: number
  verified: boolean
}

export interface SubtitleSyncVerificationRequest {
  subtitle_id: string
  media_id: number
  sample_duration?: number
  sensitivity?: number
}

export interface SubtitleSyncVerificationResponse {
  success: boolean
  sync_offset?: number
  sync_score?: number
  confidence?: number
  status: 'perfect' | 'good' | 'acceptable' | 'poor' | 'unusable'
  message?: string
  error?: string
}

export interface SubtitleTranslationRequest {
  text: string
  from_language: string
  to_language: string
  provider?: string
}

export interface SubtitleTranslationResponse {
  success: boolean
  translated_text?: string
  from_language?: string
  to_language?: string
  provider?: string
  confidence?: number
  message?: string
  error?: string
}

export interface SupportedLanguage {
  code: string
  name: string
  native_name: string
}

export interface SupportedProvider {
  name: string
  display_name: string
  enabled: boolean
  features: string[]
}

export interface SubtitleMediaInfo {
  media_id: number
  title: string
  year?: number
  media_type: string
  directory_path: string
  subtitles: SubtitleTrack[]
}

export const SUBTITLE_PROVIDERS = [
  'opensubtitles',
  'subdb',
  'yify',
  'subscene',
  'addic7ed'
] as const

export type SubtitleProvider = typeof SUBTITLE_PROVIDERS[number]

export const SUBTITLE_FORMATS = [
  'srt',
  'vtt',
  'ass',
  'ssa',
  'sub'
] as const

export type SubtitleFormat = typeof SUBTITLE_FORMATS[number]

export const COMMON_LANGUAGES = [
  { code: 'en', name: 'English', native_name: 'English' },
  { code: 'es', name: 'Spanish', native_name: 'Español' },
  { code: 'fr', name: 'French', native_name: 'Français' },
  { code: 'de', name: 'German', native_name: 'Deutsch' },
  { code: 'it', name: 'Italian', native_name: 'Italiano' },
  { code: 'pt', name: 'Portuguese', native_name: 'Português' },
  { code: 'ru', name: 'Russian', native_name: 'Русский' },
  { code: 'ja', name: 'Japanese', native_name: '日本語' },
  { code: 'zh', name: 'Chinese', native_name: '中文' },
  { code: 'ko', name: 'Korean', native_name: '한국어' },
  { code: 'ar', name: 'Arabic', native_name: 'العربية' },
  { code: 'hi', name: 'Hindi', native_name: 'हिन्दी' },
  { code: 'sv', name: 'Swedish', native_name: 'Svenska' },
  { code: 'no', name: 'Norwegian', native_name: 'Norsk' },
  { code: 'da', name: 'Danish', native_name: 'Dansk' },
  { code: 'nl', name: 'Dutch', native_name: 'Nederlands' },
  { code: 'fi', name: 'Finnish', native_name: 'Suomi' },
  { code: 'pl', name: 'Polish', native_name: 'Polski' },
  { code: 'tr', name: 'Turkish', native_name: 'Türkçe' },
  { code: 'cs', name: 'Czech', native_name: 'Čeština' }
]