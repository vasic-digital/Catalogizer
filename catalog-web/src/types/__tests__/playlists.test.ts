import { Music, Video, Image, FileText } from 'lucide-react'
import {
  flattenPlaylistItem,
  getMediaIconName,
  getMediaIcon,
  getMediaIconWithMap,
  PLAYLIST_TYPES,
} from '../playlists'
import type {
  Playlist,
  PlaylistItem,
  PlaylistCreateRequest,
  PlaylistUpdateRequest,
  SmartPlaylistRule,
  PlaylistResponse,
  PlaylistItemsResponse,
  PlaylistShareInfo,
  PlaylistAnalytics,
  PlaylistViewMode,
  PlaylistSortBy,
  CreatePlaylistRequest,
  UpdatePlaylistRequest,
} from '../playlists'

describe('playlists types and helpers', () => {
  describe('flattenPlaylistItem', () => {
    const baseItem: PlaylistItem = {
      id: 'pi-1',
      playlist_id: 'pl-1',
      media_id: 42,
      position: 0,
      media_item: {
        id: 42,
        title: 'Test Song',
        media_type: 'audio',
        year: 2024,
        cover_image: '/covers/song.jpg',
        duration: 240,
        rating: 4.5,
        quality: '320kbps',
        file_path: '/music/song.mp3',
        thumbnail_url: '/thumbs/song.jpg',
        artist: 'Test Artist',
        album: 'Test Album',
        description: 'A test song',
        file_size: 5000000,
      },
      added_at: '2024-06-01T12:00:00Z',
    }

    it('flattens all media_item properties to the top level', () => {
      const flattened = flattenPlaylistItem(baseItem)

      expect(flattened.item_id).toBe('42')
      expect(flattened.title).toBe('Test Song')
      expect(flattened.media_type).toBe('audio')
      expect(flattened.artist).toBe('Test Artist')
      expect(flattened.album).toBe('Test Album')
      expect(flattened.description).toBe('A test song')
      expect(flattened.duration).toBe(240)
      expect(flattened.quality).toBe('320kbps')
      expect(flattened.rating).toBe(4.5)
      expect(flattened.thumbnail_url).toBe('/thumbs/song.jpg')
      expect(flattened.file_path).toBe('/music/song.mp3')
      expect(flattened.file_size).toBe(5000000)
    })

    it('preserves the original media_item object', () => {
      const flattened = flattenPlaylistItem(baseItem)

      expect(flattened.media_item).toBe(baseItem.media_item)
      expect(flattened.media_item.id).toBe(42)
    })

    it('preserves the original PlaylistItem fields', () => {
      const flattened = flattenPlaylistItem(baseItem)

      expect(flattened.id).toBe('pi-1')
      expect(flattened.playlist_id).toBe('pl-1')
      expect(flattened.media_id).toBe(42)
      expect(flattened.position).toBe(0)
      expect(flattened.added_at).toBe('2024-06-01T12:00:00Z')
    })

    it('converts media_item.id to string for item_id', () => {
      const flattened = flattenPlaylistItem(baseItem)

      expect(typeof flattened.item_id).toBe('string')
      expect(flattened.item_id).toBe('42')
    })

    it('handles optional undefined fields in media_item', () => {
      const item: PlaylistItem = {
        id: 'pi-2',
        playlist_id: 'pl-1',
        media_id: 99,
        position: 1,
        media_item: {
          id: 99,
          title: 'Minimal Item',
          media_type: 'video',
        },
        added_at: '2024-06-01T12:00:00Z',
      }

      const flattened = flattenPlaylistItem(item)

      expect(flattened.title).toBe('Minimal Item')
      expect(flattened.artist).toBeUndefined()
      expect(flattened.album).toBeUndefined()
      expect(flattened.description).toBeUndefined()
      expect(flattened.duration).toBeUndefined()
      expect(flattened.quality).toBeUndefined()
      expect(flattened.rating).toBeUndefined()
      expect(flattened.thumbnail_url).toBeUndefined()
      expect(flattened.file_path).toBeUndefined()
      expect(flattened.file_size).toBeUndefined()
    })
  })

  describe('getMediaIconName', () => {
    it('returns "video" for video type', () => {
      expect(getMediaIconName('video')).toBe('video')
      expect(getMediaIconName('Video')).toBe('video')
      expect(getMediaIconName('VIDEO')).toBe('video')
    })

    it('returns "music" for audio type', () => {
      expect(getMediaIconName('audio')).toBe('music')
      expect(getMediaIconName('Audio')).toBe('music')
    })

    it('returns "music" for music type', () => {
      expect(getMediaIconName('music')).toBe('music')
      expect(getMediaIconName('Music')).toBe('music')
    })

    it('returns "image" for image type', () => {
      expect(getMediaIconName('image')).toBe('image')
      expect(getMediaIconName('Image')).toBe('image')
    })

    it('returns "document" for unknown types', () => {
      expect(getMediaIconName('document')).toBe('document')
      expect(getMediaIconName('pdf')).toBe('document')
      expect(getMediaIconName('unknown')).toBe('document')
      expect(getMediaIconName('')).toBe('document')
    })
  })

  describe('getMediaIcon', () => {
    const iconMap = {
      video: 'VideoIcon',
      music: 'MusicIcon',
      image: 'ImageIcon',
      document: 'DocumentIcon',
    }

    it('returns correct icon from icon map for video', () => {
      expect(getMediaIcon('video', iconMap)).toBe('VideoIcon')
    })

    it('returns correct icon from icon map for audio', () => {
      expect(getMediaIcon('audio', iconMap)).toBe('MusicIcon')
    })

    it('returns correct icon from icon map for music', () => {
      expect(getMediaIcon('music', iconMap)).toBe('MusicIcon')
    })

    it('returns correct icon from icon map for image', () => {
      expect(getMediaIcon('image', iconMap)).toBe('ImageIcon')
    })

    it('returns document icon for unknown media type', () => {
      expect(getMediaIcon('unknown', iconMap)).toBe('DocumentIcon')
    })

    it('returns document icon when specific icon is missing from map', () => {
      const partialMap = { document: 'FallbackIcon' }
      expect(getMediaIcon('video', partialMap)).toBe('FallbackIcon')
    })
  })

  describe('getMediaIconWithMap', () => {
    it('returns Music component for audio', () => {
      expect(getMediaIconWithMap('audio')).toBe(Music)
    })

    it('returns Music component for music', () => {
      expect(getMediaIconWithMap('music')).toBe(Music)
    })

    it('returns Video component for video', () => {
      expect(getMediaIconWithMap('video')).toBe(Video)
    })

    it('returns Image component for image', () => {
      expect(getMediaIconWithMap('image')).toBe(Image)
    })

    it('returns FileText component for unknown types', () => {
      expect(getMediaIconWithMap('document')).toBe(FileText)
      expect(getMediaIconWithMap('pdf')).toBe(FileText)
      expect(getMediaIconWithMap('unknown')).toBe(FileText)
    })
  })

  describe('PLAYLIST_TYPES', () => {
    it('contains all expected playlist types', () => {
      expect(PLAYLIST_TYPES).toContain('user_created')
      expect(PLAYLIST_TYPES).toContain('recently_played')
      expect(PLAYLIST_TYPES).toContain('most_played')
      expect(PLAYLIST_TYPES).toContain('favorites')
      expect(PLAYLIST_TYPES).toContain('watch_later')
      expect(PLAYLIST_TYPES).toContain('continue_watching')
      expect(PLAYLIST_TYPES).toContain('recommended')
    })

    it('has exactly 7 types', () => {
      expect(PLAYLIST_TYPES).toHaveLength(7)
    })

    it('is a readonly array', () => {
      // The array is created with 'as const', verify it is the expected tuple
      const types: readonly string[] = PLAYLIST_TYPES
      expect(Array.isArray(types)).toBe(true)
    })
  })

  describe('type aliases', () => {
    it('CreatePlaylistRequest is an alias for PlaylistCreateRequest', () => {
      // Verify at runtime that the types are compatible by creating objects
      const request: CreatePlaylistRequest = {
        name: 'Test',
        description: 'A test playlist',
        is_public: true,
      }
      const sameRequest: PlaylistCreateRequest = request
      expect(sameRequest.name).toBe('Test')
    })

    it('UpdatePlaylistRequest is an alias for PlaylistUpdateRequest', () => {
      const request: UpdatePlaylistRequest = {
        name: 'Updated',
        is_public: false,
      }
      const sameRequest: PlaylistUpdateRequest = request
      expect(sameRequest.name).toBe('Updated')
    })
  })

  describe('PlaylistViewMode type', () => {
    it('accepts grid and list values', () => {
      const gridMode: PlaylistViewMode = 'grid'
      const listMode: PlaylistViewMode = 'list'
      expect(gridMode).toBe('grid')
      expect(listMode).toBe('list')
    })
  })

  describe('PlaylistSortBy type', () => {
    it('accepts all valid sort values', () => {
      const sorts: PlaylistSortBy[] = [
        'name',
        'name_desc',
        'created_at',
        'updated_at',
        'duration',
        'item_count',
      ]
      expect(sorts).toHaveLength(6)
    })
  })

  describe('interface shape validation', () => {
    it('Playlist interface has required fields', () => {
      const playlist: Playlist = {
        id: 'pl-1',
        name: 'Test Playlist',
        user_id: 1,
        is_public: false,
        is_smart: false,
        item_count: 10,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      expect(playlist.id).toBe('pl-1')
      expect(playlist.name).toBe('Test Playlist')
      expect(playlist.user_id).toBe(1)
      expect(playlist.description).toBeUndefined()
      expect(playlist.thumbnail_url).toBeUndefined()
      expect(playlist.total_duration).toBeUndefined()
    })

    it('SmartPlaylistRule interface has required fields', () => {
      const rule: SmartPlaylistRule = {
        field: 'media_type',
        operator: 'equals',
        value: 'movie',
      }
      expect(rule.field).toBe('media_type')
      expect(rule.operator).toBe('equals')
      expect(rule.value).toBe('movie')
      expect(rule.condition).toBeUndefined()
    })

    it('SmartPlaylistRule supports array values', () => {
      const rule: SmartPlaylistRule = {
        field: 'genre',
        operator: 'in',
        value: ['action', 'comedy'],
        condition: 'and',
      }
      expect(rule.value).toEqual(['action', 'comedy'])
      expect(rule.condition).toBe('and')
    })

    it('PlaylistShareInfo interface has required fields', () => {
      const shareInfo: PlaylistShareInfo = {
        share_url: 'http://localhost/shared/abc',
        share_token: 'token123',
        permissions: {
          can_view: true,
          can_comment: false,
          can_download: true,
        },
      }
      expect(shareInfo.share_url).toBe('http://localhost/shared/abc')
      expect(shareInfo.permissions.can_view).toBe(true)
      expect(shareInfo.expires_at).toBeUndefined()
    })

    it('PlaylistAnalytics interface has required fields', () => {
      const analytics: PlaylistAnalytics = {
        playlist_id: 'pl-1',
        total_plays: 100,
        unique_viewers: 50,
        average_completion_rate: 0.75,
        popular_items: [
          { media_id: 1, title: 'Popular Song', play_count: 50 },
        ],
        viewing_stats: [
          { date: '2024-01-01', plays: 10, viewers: 5 },
        ],
      }
      expect(analytics.total_plays).toBe(100)
      expect(analytics.popular_items).toHaveLength(1)
      expect(analytics.viewing_stats).toHaveLength(1)
    })

    it('PlaylistResponse interface has required fields', () => {
      const response: PlaylistResponse = {
        playlists: [],
        total: 0,
        limit: 20,
        offset: 0,
      }
      expect(response.playlists).toEqual([])
      expect(response.total).toBe(0)
    })

    it('PlaylistItemsResponse interface has required fields', () => {
      const response: PlaylistItemsResponse = {
        items: [],
        total: 0,
        playlist: {
          id: 'pl-1',
          name: 'Test',
          user_id: 1,
          is_public: false,
          is_smart: false,
          item_count: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      }
      expect(response.items).toEqual([])
      expect(response.playlist.id).toBe('pl-1')
    })
  })
})
