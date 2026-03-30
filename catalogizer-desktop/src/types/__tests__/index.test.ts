import { describe, it, expect } from 'vitest'
import type {
  MediaItem,
  ExternalMetadata,
  MediaVersion,
  MediaSearchRequest,
  MediaSearchResponse,
  MediaStats,
  User,
  LoginRequest,
  LoginResponse,
  AuthStatus,
  AppConfig,
  SMBConfig,
  SMBStatus,
  PlaybackProgress,
  DownloadJob,
  MediaType,
  QualityLevel,
  SortOption,
  SortOrder,
  Theme,
} from '../index'

describe('Type definitions', () => {
  describe('MediaItem', () => {
    it('accepts a complete media item', () => {
      const item: MediaItem = {
        id: 1,
        title: 'Test Movie',
        media_type: 'movie',
        year: 2023,
        description: 'A test movie',
        cover_image: '/cover.jpg',
        rating: 8.5,
        quality: '1080p',
        file_size: 4294967296,
        duration: 7200,
        directory_path: '/media/movies/test',
        smb_path: '//server/share/movies/test',
        created_at: '2023-01-01T00:00:00Z',
        updated_at: '2023-06-15T12:00:00Z',
        is_favorite: true,
        watch_progress: 0.5,
        last_watched: '2023-06-14T20:00:00Z',
      }

      expect(item.id).toBe(1)
      expect(item.title).toBe('Test Movie')
      expect(item.media_type).toBe('movie')
      expect(item.year).toBe(2023)
      expect(item.is_favorite).toBe(true)
      expect(item.watch_progress).toBe(0.5)
    })

    it('accepts a minimal media item with only required fields', () => {
      const item: MediaItem = {
        id: 2,
        title: 'Minimal Item',
        media_type: 'music',
        directory_path: '/media/music',
        created_at: '2023-01-01',
        updated_at: '2023-01-01',
      }

      expect(item.id).toBe(2)
      expect(item.year).toBeUndefined()
      expect(item.description).toBeUndefined()
      expect(item.cover_image).toBeUndefined()
      expect(item.rating).toBeUndefined()
      expect(item.quality).toBeUndefined()
      expect(item.file_size).toBeUndefined()
      expect(item.duration).toBeUndefined()
      expect(item.smb_path).toBeUndefined()
      expect(item.external_metadata).toBeUndefined()
      expect(item.versions).toBeUndefined()
      expect(item.is_favorite).toBeUndefined()
      expect(item.watch_progress).toBeUndefined()
      expect(item.last_watched).toBeUndefined()
    })

    it('supports external metadata array', () => {
      const metadata: ExternalMetadata = {
        id: 1,
        media_id: 1,
        provider: 'tmdb',
        external_id: 'tt1234567',
        title: 'External Title',
        description: 'External description',
        year: 2023,
        rating: 7.8,
        poster_url: 'https://image.tmdb.org/poster.jpg',
        backdrop_url: 'https://image.tmdb.org/backdrop.jpg',
        genres: ['Action', 'Drama'],
        cast: ['Actor A', 'Actor B'],
        crew: ['Director X'],
        metadata: { budget: '100000000' },
        last_updated: '2023-06-15T12:00:00Z',
      }

      const item: MediaItem = {
        id: 1,
        title: 'Movie',
        media_type: 'movie',
        directory_path: '/movies',
        created_at: '2023-01-01',
        updated_at: '2023-01-01',
        external_metadata: [metadata],
      }

      expect(item.external_metadata).toHaveLength(1)
      expect(item.external_metadata![0].provider).toBe('tmdb')
      expect(item.external_metadata![0].genres).toContain('Action')
    })

    it('supports media versions array', () => {
      const version: MediaVersion = {
        id: 1,
        media_id: 1,
        version: '1.0',
        quality: '1080p',
        file_path: '/media/movies/test/movie.mkv',
        file_size: 4294967296,
        codec: 'h264',
        resolution: '1920x1080',
        bitrate: 8000000,
        language: 'en',
        frame_rate: 24,
        audio_channels: 6,
        sample_rate: 48000,
      }

      const item: MediaItem = {
        id: 1,
        title: 'Movie',
        media_type: 'movie',
        directory_path: '/movies',
        created_at: '2023-01-01',
        updated_at: '2023-01-01',
        versions: [version],
      }

      expect(item.versions).toHaveLength(1)
      expect(item.versions![0].codec).toBe('h264')
      expect(item.versions![0].audio_channels).toBe(6)
    })
  })

  describe('MediaSearchRequest', () => {
    it('accepts an empty request', () => {
      const request: MediaSearchRequest = {}

      expect(request.query).toBeUndefined()
      expect(request.media_type).toBeUndefined()
      expect(request.limit).toBeUndefined()
    })

    it('accepts a fully populated request', () => {
      const request: MediaSearchRequest = {
        query: 'test',
        media_type: 'movie',
        year_min: 2000,
        year_max: 2023,
        rating_min: 7.0,
        quality: '1080p',
        sort_by: 'rating',
        sort_order: 'desc',
        limit: 50,
        offset: 0,
      }

      expect(request.query).toBe('test')
      expect(request.limit).toBe(50)
    })
  })

  describe('MediaSearchResponse', () => {
    it('represents a paginated result set', () => {
      const response: MediaSearchResponse = {
        items: [],
        total: 100,
        limit: 50,
        offset: 0,
      }

      expect(response.items).toHaveLength(0)
      expect(response.total).toBe(100)
    })
  })

  describe('MediaStats', () => {
    it('represents media collection statistics', () => {
      const stats: MediaStats = {
        total_items: 1500,
        by_type: { movie: 800, tv_show: 400, music: 300 },
        by_quality: { '1080p': 600, '4k': 200, '720p': 400, other: 300 },
        total_size: 5000000000000,
        recent_additions: 25,
      }

      expect(stats.total_items).toBe(1500)
      expect(stats.by_type.movie).toBe(800)
      expect(stats.by_quality['4k']).toBe(200)
    })
  })

  describe('User', () => {
    it('represents a complete user profile', () => {
      const user: User = {
        id: 1,
        username: 'admin',
        email: 'admin@example.com',
        first_name: 'Admin',
        last_name: 'User',
        is_admin: true,
        permissions: ['read', 'write', 'admin'],
        created_at: '2023-01-01T00:00:00Z',
        updated_at: '2023-06-15T12:00:00Z',
      }

      expect(user.username).toBe('admin')
      expect(user.is_admin).toBe(true)
      expect(user.permissions).toContain('admin')
    })

    it('works without optional permissions', () => {
      const user: User = {
        id: 2,
        username: 'viewer',
        email: 'viewer@example.com',
        first_name: 'View',
        last_name: 'Only',
        is_admin: false,
        created_at: '2023-01-01',
        updated_at: '2023-01-01',
      }

      expect(user.permissions).toBeUndefined()
      expect(user.is_admin).toBe(false)
    })
  })

  describe('Auth types', () => {
    it('represents a login request', () => {
      const request: LoginRequest = {
        username: 'admin',
        password: 'secret123',
      }

      expect(request.username).toBe('admin')
      expect(request.password).toBe('secret123')
    })

    it('represents a login response', () => {
      const response: LoginResponse = {
        token: 'jwt-token-here',
        refresh_token: 'refresh-token-here',
        expires_in: 3600,
        user: {
          id: 1,
          username: 'admin',
          email: 'admin@test.com',
          first_name: 'Admin',
          last_name: 'User',
          is_admin: true,
          created_at: '2023-01-01',
          updated_at: '2023-01-01',
        },
      }

      expect(response.token).toBe('jwt-token-here')
      expect(response.expires_in).toBe(3600)
      expect(response.user.username).toBe('admin')
    })

    it('represents an auth status', () => {
      const authenticated: AuthStatus = {
        authenticated: true,
        user: {
          id: 1,
          username: 'admin',
          email: 'a@b.com',
          first_name: 'A',
          last_name: 'B',
          is_admin: true,
          created_at: '',
          updated_at: '',
        },
        expires_at: '2023-12-31T23:59:59Z',
      }

      const unauthenticated: AuthStatus = {
        authenticated: false,
      }

      expect(authenticated.authenticated).toBe(true)
      expect(authenticated.user).toBeDefined()
      expect(unauthenticated.authenticated).toBe(false)
      expect(unauthenticated.user).toBeUndefined()
    })
  })

  describe('AppConfig', () => {
    it('represents full application configuration', () => {
      const config: AppConfig = {
        server_url: 'http://localhost:8080',
        auth_token: 'jwt-token',
        theme: 'dark',
        auto_start: true,
      }

      expect(config.server_url).toBe('http://localhost:8080')
      expect(config.theme).toBe('dark')
      expect(config.auto_start).toBe(true)
    })

    it('works with undefined optional fields', () => {
      const config: AppConfig = {
        theme: 'light',
        auto_start: false,
      }

      expect(config.server_url).toBeUndefined()
      expect(config.auth_token).toBeUndefined()
    })
  })

  describe('SMBConfig', () => {
    it('represents a complete SMB configuration', () => {
      const config: SMBConfig = {
        id: 1,
        name: 'Media Server',
        host: '192.168.1.100',
        port: 445,
        share_name: 'media',
        username: 'user',
        password: 'pass',
        domain: 'WORKGROUP',
        is_active: true,
        mount_point: '/mnt/media',
        created_at: '2023-01-01',
        updated_at: '2023-01-01',
      }

      expect(config.host).toBe('192.168.1.100')
      expect(config.port).toBe(445)
      expect(config.is_active).toBe(true)
      expect(config.domain).toBe('WORKGROUP')
    })
  })

  describe('SMBStatus', () => {
    it('represents SMB connection status', () => {
      const status: SMBStatus = {
        config_id: 1,
        is_connected: true,
        last_check: '2023-06-15T12:00:00Z',
        mount_point: '/mnt/media',
      }

      expect(status.is_connected).toBe(true)
      expect(status.error_message).toBeUndefined()
    })

    it('represents a disconnected status with error', () => {
      const status: SMBStatus = {
        config_id: 1,
        is_connected: false,
        last_check: '2023-06-15T12:00:00Z',
        error_message: 'Connection refused',
      }

      expect(status.is_connected).toBe(false)
      expect(status.error_message).toBe('Connection refused')
    })
  })

  describe('PlaybackProgress', () => {
    it('represents playback state', () => {
      const progress: PlaybackProgress = {
        media_id: 42,
        position: 3600,
        duration: 7200,
        timestamp: Date.now(),
      }

      expect(progress.position).toBe(3600)
      expect(progress.duration).toBe(7200)
      expect(progress.position / progress.duration).toBe(0.5)
    })
  })

  describe('DownloadJob', () => {
    it('represents a download in progress', () => {
      const job: DownloadJob = {
        id: 1,
        media_id: 42,
        status: 'downloading',
        progress: 0.65,
        file_path: '/downloads/movie.mkv',
        created_at: '2023-06-15T10:00:00Z',
        updated_at: '2023-06-15T10:15:00Z',
      }

      expect(job.status).toBe('downloading')
      expect(job.progress).toBe(0.65)
      expect(job.error_message).toBeUndefined()
    })

    it('represents a failed download', () => {
      const job: DownloadJob = {
        id: 2,
        media_id: 43,
        status: 'failed',
        progress: 0.3,
        file_path: '/downloads/other.mkv',
        error_message: 'Disk full',
        created_at: '2023-06-15T10:00:00Z',
        updated_at: '2023-06-15T10:05:00Z',
      }

      expect(job.status).toBe('failed')
      expect(job.error_message).toBe('Disk full')
    })

    it('supports all valid statuses', () => {
      const statuses: DownloadJob['status'][] = [
        'pending', 'downloading', 'completed', 'failed', 'paused',
      ]

      statuses.forEach(status => {
        const job: DownloadJob = {
          id: 1,
          media_id: 1,
          status,
          progress: 0,
          file_path: '/test',
          created_at: '',
          updated_at: '',
        }
        expect(job.status).toBe(status)
      })
    })
  })

  describe('Union types', () => {
    it('supports all MediaType values', () => {
      const types: MediaType[] = [
        'movie', 'tv_show', 'documentary', 'anime', 'music',
        'audiobook', 'podcast', 'game', 'software', 'ebook',
        'training', 'concert', 'youtube_video', 'sports', 'news', 'other',
      ]

      expect(types).toHaveLength(16)
    })

    it('supports all QualityLevel values', () => {
      const levels: QualityLevel[] = [
        'cam', 'ts', 'dvdrip', 'brrip', '720p', '1080p', '4k', 'hdr', 'dolby_vision',
      ]

      expect(levels).toHaveLength(9)
    })

    it('supports all SortOption values', () => {
      const options: SortOption[] = [
        'title', 'year', 'rating', 'updated_at', 'created_at', 'file_size', 'duration',
      ]

      expect(options).toHaveLength(7)
    })

    it('supports both SortOrder values', () => {
      const orders: SortOrder[] = ['asc', 'desc']

      expect(orders).toHaveLength(2)
    })

    it('supports all Theme values', () => {
      const themes: Theme[] = ['light', 'dark', 'system']

      expect(themes).toHaveLength(3)
    })
  })
})
