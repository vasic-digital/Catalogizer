import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import axios from 'axios'
import { mediaApi } from './mediaApi'
import type { MediaItem, MediaSearchRequest } from '@/types/media'

// Mock axios
vi.mock('axios')
const mockedAxios = axios as jest.Mocked<typeof axios>

describe('mediaApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  describe('searchMedia', () => {
    it('should search media with parameters', async () => {
      const mockResponse = {
        data: {
          items: [
            {
              id: 1,
              title: 'Test Movie',
              media_type: 'movie',
              directory_path: '/movies/test.mp4',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
          ],
          total: 1,
          limit: 20,
          offset: 0,
        },
      }

      mockedAxios.get.mockResolvedValue(mockResponse)

      const params: MediaSearchRequest = {
        query: 'Test',
        media_type: 'movie',
        limit: 20,
        offset: 0,
      }

      const result = await mediaApi.searchMedia(params)

      expect(mockedAxios.get).toHaveBeenCalledWith('/media/search', { params })
      expect(result).toEqual(mockResponse.data)
      expect(result.items).toHaveLength(1)
      expect(result.items[0].title).toBe('Test Movie')
    })

    it('should handle search with all filter parameters', async () => {
      const mockResponse = {
        data: {
          items: [],
          total: 0,
          limit: 10,
          offset: 0,
        },
      }

      mockedAxios.get.mockResolvedValue(mockResponse)

      const params: MediaSearchRequest = {
        query: 'Matrix',
        media_type: 'movie',
        year_min: 1999,
        year_max: 1999,
        rating_min: 8.0,
        quality: '1080p',
        sort_by: 'rating',
        sort_order: 'desc',
        limit: 10,
        offset: 0,
      }

      await mediaApi.searchMedia(params)

      expect(mockedAxios.get).toHaveBeenCalledWith('/media/search', { params })
    })

    it('should handle search errors', async () => {
      mockedAxios.get.mockRejectedValue(new Error('Network error'))

      await expect(mediaApi.searchMedia({})).rejects.toThrow('Network error')
    })
  })

  describe('getMediaById', () => {
    it('should fetch media by ID', async () => {
      const mockMedia: MediaItem = {
        id: 123,
        title: 'The Matrix',
        media_type: 'movie',
        year: 1999,
        rating: 8.7,
        directory_path: '/movies/matrix.mp4',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      mockedAxios.get.mockResolvedValue({ data: mockMedia })

      const result = await mediaApi.getMediaById(123)

      expect(mockedAxios.get).toHaveBeenCalledWith('/media/123')
      expect(result).toEqual(mockMedia)
      expect(result.id).toBe(123)
      expect(result.title).toBe('The Matrix')
    })

    it('should handle 404 for non-existent media', async () => {
      mockedAxios.get.mockRejectedValue({ response: { status: 404 } })

      await expect(mediaApi.getMediaById(999)).rejects.toBeTruthy()
    })
  })

  describe('getMediaStats', () => {
    it('should fetch media statistics', async () => {
      const mockStats = {
        total_items: 1000,
        by_type: {
          movie: 500,
          tv_show: 300,
          music: 200,
        },
        by_quality: {
          '1080p': 600,
          '720p': 300,
          '4k': 100,
        },
        total_size: 1073741824000, // 1 TB
        recent_additions: 50,
      }

      mockedAxios.get.mockResolvedValue({ data: mockStats })

      const result = await mediaApi.getMediaStats()

      expect(mockedAxios.get).toHaveBeenCalledWith('/media/stats')
      expect(result).toEqual(mockStats)
      expect(result.total_items).toBe(1000)
      expect(result.by_type.movie).toBe(500)
    })
  })

  describe('deleteMedia', () => {
    it('should delete media by ID', async () => {
      mockedAxios.delete.mockResolvedValue({ data: null })

      await mediaApi.deleteMedia(123)

      expect(mockedAxios.delete).toHaveBeenCalledWith('/media/123')
    })

    it('should handle delete errors', async () => {
      mockedAxios.delete.mockRejectedValue(new Error('Delete failed'))

      await expect(mediaApi.deleteMedia(123)).rejects.toThrow('Delete failed')
    })
  })

  describe('updateMedia', () => {
    it('should update media metadata', async () => {
      const updatedMedia: MediaItem = {
        id: 123,
        title: 'Updated Title',
        media_type: 'movie',
        directory_path: '/movies/test.mp4',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      }

      mockedAxios.put.mockResolvedValue({ data: updatedMedia })

      const result = await mediaApi.updateMedia(123, { title: 'Updated Title' })

      expect(mockedAxios.put).toHaveBeenCalledWith('/media/123', { title: 'Updated Title' })
      expect(result.title).toBe('Updated Title')
    })
  })

  describe('downloadMedia', () => {
    let createElementSpy: jest.SpyInstance
    let appendChildSpy: jest.SpyInstance
    let removeChildSpy: jest.SpyInstance
    let clickSpy: jest.Mock

    beforeEach(() => {
      // Mock DOM methods
      const mockLink = {
        href: '',
        download: '',
        click: vi.fn(),
      } as unknown as HTMLAnchorElement

      createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue(mockLink)
      appendChildSpy = vi.spyOn(document.body, 'appendChild').mockImplementation(() => mockLink)
      removeChildSpy = vi.spyOn(document.body, 'removeChild').mockImplementation(() => mockLink)
      clickSpy = mockLink.click as jest.Mock

      // Mock URL methods
      global.URL.createObjectURL = vi.fn(() => 'blob:mock-url')
      global.URL.revokeObjectURL = vi.fn()

      // Mock Blob
      global.Blob = vi.fn(() => ({} as Blob)) as any
    })

    afterEach(() => {
      createElementSpy.mockRestore()
      appendChildSpy.mockRestore()
      removeChildSpy.mockRestore()
      vi.unstubAllGlobals()
    })

    it('should download media file', async () => {
      const media: MediaItem = {
        id: 123,
        title: 'Test Movie',
        media_type: 'movie',
        directory_path: '/movies/test.mp4',
        storage_root_name: 'main_storage',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      const mockBlob = new Blob(['test content'])
      mockedAxios.get.mockResolvedValue({ data: mockBlob })

      await mediaApi.downloadMedia(media)

      expect(mockedAxios.get).toHaveBeenCalledWith('/download', {
        params: {
          path: '/movies/test.mp4',
          storage: 'main_storage',
        },
        responseType: 'blob',
      })

      expect(createElementSpy).toHaveBeenCalledWith('a')
      expect(clickSpy).toHaveBeenCalled()
      expect(global.URL.createObjectURL).toHaveBeenCalled()
      expect(global.URL.revokeObjectURL).toHaveBeenCalled()
    })

    it('should extract filename from path', async () => {
      const media: MediaItem = {
        id: 123,
        title: 'Test Movie',
        media_type: 'movie',
        directory_path: '/movies/subfolder/test.mp4',
        storage_root_name: 'main_storage',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      mockedAxios.get.mockResolvedValue({ data: new Blob() })

      await mediaApi.downloadMedia(media)

      const mockLink = createElementSpy.mock.results[0].value
      expect(mockLink.download).toBe('test.mp4')
    })

    it('should use title as fallback filename', async () => {
      const media: MediaItem = {
        id: 123,
        title: 'Test Movie',
        media_type: 'movie',
        directory_path: '/',
        storage_root_name: 'main_storage',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      mockedAxios.get.mockResolvedValue({ data: new Blob() })

      await mediaApi.downloadMedia(media)

      const mockLink = createElementSpy.mock.results[0].value
      expect(mockLink.download).toBe('Test Movie.movie')
    })

    it('should handle download errors', async () => {
      const media: MediaItem = {
        id: 123,
        title: 'Test Movie',
        media_type: 'movie',
        directory_path: '/movies/test.mp4',
        storage_root_name: 'main_storage',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      mockedAxios.get.mockRejectedValue(new Error('Download failed'))

      await expect(mediaApi.downloadMedia(media)).rejects.toThrow('Download failed')
    })

    it('should cleanup DOM after download', async () => {
      const media: MediaItem = {
        id: 123,
        title: 'Test Movie',
        media_type: 'movie',
        directory_path: '/movies/test.mp4',
        storage_root_name: 'main_storage',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      mockedAxios.get.mockResolvedValue({ data: new Blob() })

      await mediaApi.downloadMedia(media)

      expect(removeChildSpy).toHaveBeenCalled()
      expect(global.URL.revokeObjectURL).toHaveBeenCalled()
    })
  })

  describe('Storage Root Management', () => {
    it('should fetch all storage roots', async () => {
      const mockRoots = [
        {
          id: 1,
          name: 'main_storage',
          protocol: 'smb',
          enabled: true,
          max_depth: 10,
          enable_duplicate_detection: true,
          enable_metadata_extraction: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ]

      mockedAxios.get.mockResolvedValue({ data: { data: mockRoots } })

      const result = await mediaApi.getStorageRoots()

      expect(mockedAxios.get).toHaveBeenCalledWith('/storage/roots')
      expect(result).toEqual(mockRoots)
    })

    it('should test storage root connection', async () => {
      const mockResponse = {
        success: true,
        message: 'Connection successful',
      }

      mockedAxios.post.mockResolvedValue({ data: mockResponse })

      const result = await mediaApi.testStorageRoot(1)

      expect(mockedAxios.post).toHaveBeenCalledWith('/storage/roots/1/test')
      expect(result.success).toBe(true)
    })
  })

  describe('Error handling', () => {
    it('should propagate network errors', async () => {
      mockedAxios.get.mockRejectedValue(new Error('Network error'))

      await expect(mediaApi.getMediaById(1)).rejects.toThrow('Network error')
    })

    it('should handle API errors with status codes', async () => {
      const error = {
        response: {
          status: 500,
          data: { message: 'Internal server error' },
        },
      }

      mockedAxios.get.mockRejectedValue(error)

      await expect(mediaApi.getMediaStats()).rejects.toMatchObject(error)
    })
  })
})
