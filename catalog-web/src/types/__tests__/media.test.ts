import {
  MEDIA_TYPES,
  QUALITY_LEVELS,
  SUPPORTED_PROTOCOLS,
} from '../media'
import type {
  MediaType,
  QualityLevel,
  StorageProtocol,
  MediaEntityDetail,
  EntityUserMetadata,
  MediaTypeInfo,
  EntityListResponse,
  EntityStatsResponse,
  EntityFile,
  StorageRoot,
} from '../media'

describe('media types and constants', () => {
  describe('MEDIA_TYPES', () => {
    it('contains all 11 expected media types', () => {
      expect(MEDIA_TYPES).toHaveLength(11)
    })

    it('includes movie type', () => {
      expect(MEDIA_TYPES).toContain('movie')
    })

    it('includes tv types', () => {
      expect(MEDIA_TYPES).toContain('tv_show')
      expect(MEDIA_TYPES).toContain('tv_season')
      expect(MEDIA_TYPES).toContain('tv_episode')
    })

    it('includes music types', () => {
      expect(MEDIA_TYPES).toContain('music_artist')
      expect(MEDIA_TYPES).toContain('music_album')
      expect(MEDIA_TYPES).toContain('song')
    })

    it('includes game and software types', () => {
      expect(MEDIA_TYPES).toContain('game')
      expect(MEDIA_TYPES).toContain('software')
    })

    it('includes book and comic types', () => {
      expect(MEDIA_TYPES).toContain('book')
      expect(MEDIA_TYPES).toContain('comic')
    })

    it('is a readonly array', () => {
      const types: readonly string[] = MEDIA_TYPES
      expect(Array.isArray(types)).toBe(true)
    })
  })

  describe('QUALITY_LEVELS', () => {
    it('contains all 9 expected quality levels', () => {
      expect(QUALITY_LEVELS).toHaveLength(9)
    })

    it('includes standard definition levels', () => {
      expect(QUALITY_LEVELS).toContain('cam')
      expect(QUALITY_LEVELS).toContain('ts')
      expect(QUALITY_LEVELS).toContain('dvdrip')
    })

    it('includes blu-ray and HD levels', () => {
      expect(QUALITY_LEVELS).toContain('brrip')
      expect(QUALITY_LEVELS).toContain('720p')
      expect(QUALITY_LEVELS).toContain('1080p')
    })

    it('includes UHD and HDR levels', () => {
      expect(QUALITY_LEVELS).toContain('4k')
      expect(QUALITY_LEVELS).toContain('hdr')
      expect(QUALITY_LEVELS).toContain('dolby_vision')
    })

    it('is a readonly array', () => {
      const levels: readonly string[] = QUALITY_LEVELS
      expect(Array.isArray(levels)).toBe(true)
    })
  })

  describe('SUPPORTED_PROTOCOLS', () => {
    it('contains all 5 expected protocols', () => {
      expect(SUPPORTED_PROTOCOLS).toHaveLength(5)
    })

    it('includes network protocols', () => {
      expect(SUPPORTED_PROTOCOLS).toContain('smb')
      expect(SUPPORTED_PROTOCOLS).toContain('ftp')
      expect(SUPPORTED_PROTOCOLS).toContain('nfs')
      expect(SUPPORTED_PROTOCOLS).toContain('webdav')
    })

    it('includes local protocol', () => {
      expect(SUPPORTED_PROTOCOLS).toContain('local')
    })

    it('is a readonly array', () => {
      const protocols: readonly string[] = SUPPORTED_PROTOCOLS
      expect(Array.isArray(protocols)).toBe(true)
    })
  })

  describe('MediaType type', () => {
    it('accepts valid media type values', () => {
      const movieType: MediaType = 'movie'
      const tvType: MediaType = 'tv_show'
      const songType: MediaType = 'song'
      expect(movieType).toBe('movie')
      expect(tvType).toBe('tv_show')
      expect(songType).toBe('song')
    })
  })

  describe('QualityLevel type', () => {
    it('accepts valid quality levels', () => {
      const hd: QualityLevel = '1080p'
      const uhd: QualityLevel = '4k'
      expect(hd).toBe('1080p')
      expect(uhd).toBe('4k')
    })
  })

  describe('StorageProtocol type', () => {
    it('accepts valid protocols', () => {
      const smb: StorageProtocol = 'smb'
      const local: StorageProtocol = 'local'
      expect(smb).toBe('smb')
      expect(local).toBe('local')
    })
  })

  describe('interface shape validation', () => {
    it('MediaEntityDetail has required fields', () => {
      const entity: MediaEntityDetail = {
        id: 1,
        media_type_id: 1,
        title: 'The Matrix',
        status: 'active',
        first_detected: '2024-01-01T00:00:00Z',
        last_updated: '2024-01-01T00:00:00Z',
        media_type: 'movie',
        file_count: 3,
        children_count: 0,
        external_metadata: [],
      }
      expect(entity.id).toBe(1)
      expect(entity.title).toBe('The Matrix')
      expect(entity.media_type).toBe('movie')
      expect(entity.original_title).toBeUndefined()
      expect(entity.year).toBeUndefined()
      expect(entity.description).toBeUndefined()
      expect(entity.genre).toBeUndefined()
      expect(entity.director).toBeUndefined()
      expect(entity.rating).toBeUndefined()
      expect(entity.runtime).toBeUndefined()
      expect(entity.parent_id).toBeUndefined()
    })

    it('MediaEntityDetail accepts all optional fields', () => {
      const entity: MediaEntityDetail = {
        id: 2,
        media_type_id: 3,
        title: 'Breaking Bad S01E01',
        original_title: 'Pilot',
        year: 2008,
        description: 'A high school chemistry teacher...',
        genre: ['drama', 'crime'],
        director: 'Vince Gilligan',
        rating: 9.5,
        runtime: 58,
        language: 'en',
        status: 'active',
        parent_id: 1,
        season_number: 1,
        episode_number: 1,
        track_number: undefined,
        first_detected: '2024-01-01T00:00:00Z',
        last_updated: '2024-01-02T00:00:00Z',
        media_type: 'tv_episode',
        file_count: 1,
        children_count: 0,
        external_metadata: [],
      }
      expect(entity.year).toBe(2008)
      expect(entity.genre).toEqual(['drama', 'crime'])
      expect(entity.season_number).toBe(1)
      expect(entity.episode_number).toBe(1)
    })

    it('EntityUserMetadata has all optional fields', () => {
      const metadata: EntityUserMetadata = {}
      expect(metadata.user_rating).toBeUndefined()
      expect(metadata.watched_status).toBeUndefined()
      expect(metadata.favorite).toBeUndefined()
      expect(metadata.personal_notes).toBeUndefined()
      expect(metadata.tags).toBeUndefined()
    })

    it('EntityUserMetadata accepts all fields', () => {
      const metadata: EntityUserMetadata = {
        user_rating: 8.5,
        watched_status: 'completed',
        favorite: true,
        personal_notes: 'Great movie',
        tags: ['must-watch', 'sci-fi'],
      }
      expect(metadata.user_rating).toBe(8.5)
      expect(metadata.favorite).toBe(true)
      expect(metadata.tags).toHaveLength(2)
    })

    it('MediaTypeInfo has required fields', () => {
      const info: MediaTypeInfo = {
        id: 1,
        name: 'movie',
        description: 'Feature films',
        count: 500,
      }
      expect(info.id).toBe(1)
      expect(info.name).toBe('movie')
      expect(info.count).toBe(500)
    })

    it('EntityListResponse has required fields', () => {
      const response: EntityListResponse = {
        items: [],
        total: 0,
        limit: 20,
        offset: 0,
      }
      expect(response.items).toEqual([])
      expect(response.total).toBe(0)
      expect(response.limit).toBe(20)
    })

    it('EntityStatsResponse has required fields', () => {
      const stats: EntityStatsResponse = {
        total_entities: 1500,
        by_type: { movie: 500, tv_show: 300, song: 700 },
      }
      expect(stats.total_entities).toBe(1500)
      expect(stats.by_type.movie).toBe(500)
    })

    it('EntityFile has required fields', () => {
      const file: EntityFile = {
        id: 1,
        media_item_id: 42,
        file_id: 100,
        is_primary: true,
        created_at: '2024-01-01T00:00:00Z',
      }
      expect(file.id).toBe(1)
      expect(file.is_primary).toBe(true)
      expect(file.quality_info).toBeUndefined()
      expect(file.language).toBeUndefined()
    })

    it('StorageRoot has required fields', () => {
      const root: StorageRoot = {
        id: 1,
        name: 'NAS Media',
        protocol: 'smb',
        enabled: true,
        max_depth: 5,
        enable_duplicate_detection: true,
        enable_metadata_extraction: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      expect(root.id).toBe(1)
      expect(root.protocol).toBe('smb')
      expect(root.enabled).toBe(true)
      expect(root.last_scan_at).toBeUndefined()
    })

    it('StorageRoot accepts optional last_scan_at', () => {
      const root: StorageRoot = {
        id: 2,
        name: 'Local Drive',
        protocol: 'local',
        enabled: false,
        max_depth: 3,
        enable_duplicate_detection: false,
        enable_metadata_extraction: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        last_scan_at: '2024-05-30T12:00:00Z',
      }
      expect(root.last_scan_at).toBe('2024-05-30T12:00:00Z')
    })
  })
})
