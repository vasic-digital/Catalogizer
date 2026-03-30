import {
  SUBTITLE_PROVIDERS,
  SUBTITLE_FORMATS,
  COMMON_LANGUAGES,
} from '../subtitles'
import type {
  SubtitleProvider,
  SubtitleFormat,
  SubtitleSearchRequest,
  SubtitleSearchResult,
  SubtitleSearchResponse,
  SubtitleDownloadRequest,
  SubtitleDownloadResponse,
  SubtitleTrack,
  SubtitleSyncVerificationRequest,
  SubtitleSyncVerificationResponse,
  SubtitleTranslationRequest,
  SubtitleTranslationResponse,
  SupportedLanguage,
  SupportedProvider,
  SubtitleMediaInfo,
} from '../subtitles'

describe('subtitles types and constants', () => {
  describe('SUBTITLE_PROVIDERS', () => {
    it('contains all 5 expected providers', () => {
      expect(SUBTITLE_PROVIDERS).toHaveLength(5)
    })

    it('includes opensubtitles', () => {
      expect(SUBTITLE_PROVIDERS).toContain('opensubtitles')
    })

    it('includes subdb', () => {
      expect(SUBTITLE_PROVIDERS).toContain('subdb')
    })

    it('includes yify', () => {
      expect(SUBTITLE_PROVIDERS).toContain('yify')
    })

    it('includes subscene', () => {
      expect(SUBTITLE_PROVIDERS).toContain('subscene')
    })

    it('includes addic7ed', () => {
      expect(SUBTITLE_PROVIDERS).toContain('addic7ed')
    })

    it('is a readonly array', () => {
      const providers: readonly string[] = SUBTITLE_PROVIDERS
      expect(Array.isArray(providers)).toBe(true)
    })
  })

  describe('SUBTITLE_FORMATS', () => {
    it('contains all 5 expected formats', () => {
      expect(SUBTITLE_FORMATS).toHaveLength(5)
    })

    it('includes srt format', () => {
      expect(SUBTITLE_FORMATS).toContain('srt')
    })

    it('includes vtt format', () => {
      expect(SUBTITLE_FORMATS).toContain('vtt')
    })

    it('includes ass and ssa formats', () => {
      expect(SUBTITLE_FORMATS).toContain('ass')
      expect(SUBTITLE_FORMATS).toContain('ssa')
    })

    it('includes sub format', () => {
      expect(SUBTITLE_FORMATS).toContain('sub')
    })

    it('is a readonly array', () => {
      const formats: readonly string[] = SUBTITLE_FORMATS
      expect(Array.isArray(formats)).toBe(true)
    })
  })

  describe('COMMON_LANGUAGES', () => {
    it('contains 20 languages', () => {
      expect(COMMON_LANGUAGES).toHaveLength(20)
    })

    it('includes English', () => {
      const english = COMMON_LANGUAGES.find(l => l.code === 'en')
      expect(english).toBeDefined()
      expect(english!.name).toBe('English')
      expect(english!.native_name).toBe('English')
    })

    it('includes Spanish', () => {
      const spanish = COMMON_LANGUAGES.find(l => l.code === 'es')
      expect(spanish).toBeDefined()
      expect(spanish!.name).toBe('Spanish')
    })

    it('includes Japanese', () => {
      const japanese = COMMON_LANGUAGES.find(l => l.code === 'ja')
      expect(japanese).toBeDefined()
      expect(japanese!.name).toBe('Japanese')
    })

    it('includes Russian', () => {
      const russian = COMMON_LANGUAGES.find(l => l.code === 'ru')
      expect(russian).toBeDefined()
      expect(russian!.name).toBe('Russian')
    })

    it('each language has code, name, and native_name', () => {
      COMMON_LANGUAGES.forEach(lang => {
        expect(lang.code).toBeTruthy()
        expect(lang.name).toBeTruthy()
        expect(lang.native_name).toBeTruthy()
      })
    })

    it('all language codes are unique', () => {
      const codes = COMMON_LANGUAGES.map(l => l.code)
      expect(new Set(codes).size).toBe(codes.length)
    })

    it('all language codes are 2 characters', () => {
      COMMON_LANGUAGES.forEach(lang => {
        expect(lang.code).toHaveLength(2)
      })
    })
  })

  describe('SubtitleProvider type', () => {
    it('accepts valid provider values', () => {
      const os: SubtitleProvider = 'opensubtitles'
      const yify: SubtitleProvider = 'yify'
      expect(os).toBe('opensubtitles')
      expect(yify).toBe('yify')
    })
  })

  describe('SubtitleFormat type', () => {
    it('accepts valid format values', () => {
      const srt: SubtitleFormat = 'srt'
      const vtt: SubtitleFormat = 'vtt'
      expect(srt).toBe('srt')
      expect(vtt).toBe('vtt')
    })
  })

  describe('interface shape validation', () => {
    it('SubtitleSearchRequest has all optional fields', () => {
      const request: SubtitleSearchRequest = {}
      expect(request.query).toBeUndefined()
      expect(request.language).toBeUndefined()
      expect(request.media_path).toBeUndefined()
      expect(request.title).toBeUndefined()
      expect(request.year).toBeUndefined()
      expect(request.season).toBeUndefined()
      expect(request.episode).toBeUndefined()
      expect(request.providers).toBeUndefined()
      expect(request.limit).toBeUndefined()
      expect(request.offset).toBeUndefined()
    })

    it('SubtitleSearchRequest accepts all fields', () => {
      const request: SubtitleSearchRequest = {
        query: 'Breaking Bad',
        language: 'en',
        media_path: '/media/shows/bb',
        title: 'Breaking Bad',
        year: 2008,
        season: 1,
        episode: 1,
        providers: ['opensubtitles', 'yify'],
        limit: 20,
        offset: 0,
      }
      expect(request.query).toBe('Breaking Bad')
      expect(request.providers).toHaveLength(2)
    })

    it('SubtitleSearchResult has required fields', () => {
      const result: SubtitleSearchResult = {
        id: 'sub-1',
        title: 'Breaking Bad S01E01',
        language: 'en',
        language_name: 'English',
        provider: 'opensubtitles',
        download_url: 'https://example.com/sub.srt',
        hearing_impaired: false,
        foreign_parts_only: false,
        machine_translated: false,
      }
      expect(result.id).toBe('sub-1')
      expect(result.hearing_impaired).toBe(false)
      expect(result.media_path).toBeUndefined()
      expect(result.year).toBeUndefined()
      expect(result.rating).toBeUndefined()
    })

    it('SubtitleSearchResponse has required fields', () => {
      const response: SubtitleSearchResponse = {
        results: [],
        total: 0,
        limit: 20,
        offset: 0,
        query_time_ms: 150,
      }
      expect(response.results).toEqual([])
      expect(response.query_time_ms).toBe(150)
    })

    it('SubtitleDownloadRequest has required field', () => {
      const request: SubtitleDownloadRequest = {
        id: 'sub-1',
      }
      expect(request.id).toBe('sub-1')
      expect(request.media_path).toBeUndefined()
      expect(request.language).toBeUndefined()
      expect(request.encoding).toBeUndefined()
    })

    it('SubtitleDownloadResponse has required field', () => {
      const response: SubtitleDownloadResponse = {
        success: true,
      }
      expect(response.success).toBe(true)
      expect(response.subtitle_id).toBeUndefined()
      expect(response.error).toBeUndefined()
    })

    it('SubtitleTrack has required fields', () => {
      const track: SubtitleTrack = {
        id: 'track-1',
        media_id: 42,
        language: 'en',
        language_name: 'English',
        provider: 'opensubtitles',
        file_path: '/subs/en.srt',
        format: 'srt',
        encoding: 'utf-8',
        file_size: 45000,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        hearing_impaired: false,
        foreign_parts_only: false,
        machine_translated: false,
        verified: true,
      }
      expect(track.id).toBe('track-1')
      expect(track.verified).toBe(true)
      expect(track.sync_offset).toBeUndefined()
      expect(track.rating).toBeUndefined()
    })

    it('SubtitleSyncVerificationRequest has required fields', () => {
      const request: SubtitleSyncVerificationRequest = {
        subtitle_id: 'sub-1',
        media_id: 42,
      }
      expect(request.subtitle_id).toBe('sub-1')
      expect(request.sample_duration).toBeUndefined()
      expect(request.sensitivity).toBeUndefined()
    })

    it('SubtitleSyncVerificationResponse has required fields', () => {
      const response: SubtitleSyncVerificationResponse = {
        success: true,
        status: 'perfect',
      }
      expect(response.success).toBe(true)
      expect(response.status).toBe('perfect')
      expect(response.sync_offset).toBeUndefined()
    })

    it('SubtitleSyncVerificationResponse accepts all status values', () => {
      const statuses: SubtitleSyncVerificationResponse['status'][] = [
        'perfect', 'good', 'acceptable', 'poor', 'unusable',
      ]
      expect(statuses).toHaveLength(5)
    })

    it('SubtitleTranslationRequest has required fields', () => {
      const request: SubtitleTranslationRequest = {
        text: 'Hello, world!',
        from_language: 'en',
        to_language: 'es',
      }
      expect(request.text).toBe('Hello, world!')
      expect(request.provider).toBeUndefined()
    })

    it('SubtitleTranslationResponse has required field', () => {
      const response: SubtitleTranslationResponse = {
        success: true,
      }
      expect(response.success).toBe(true)
      expect(response.translated_text).toBeUndefined()
    })

    it('SupportedLanguage has required fields', () => {
      const lang: SupportedLanguage = {
        code: 'en',
        name: 'English',
        native_name: 'English',
      }
      expect(lang.code).toBe('en')
    })

    it('SupportedProvider has required fields', () => {
      const provider: SupportedProvider = {
        name: 'opensubtitles',
        display_name: 'OpenSubtitles',
        enabled: true,
        features: ['search', 'download', 'upload'],
      }
      expect(provider.name).toBe('opensubtitles')
      expect(provider.features).toHaveLength(3)
    })

    it('SubtitleMediaInfo has required fields', () => {
      const info: SubtitleMediaInfo = {
        media_id: 42,
        title: 'The Matrix',
        media_type: 'movie',
        directory_path: '/media/movies/the-matrix',
        subtitles: [],
      }
      expect(info.media_id).toBe(42)
      expect(info.subtitles).toEqual([])
      expect(info.year).toBeUndefined()
    })
  })
})
