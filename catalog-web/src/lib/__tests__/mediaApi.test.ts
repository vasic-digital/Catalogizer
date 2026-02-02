import { mediaApi } from '../mediaApi'
import apiDefault from '../api'

// Mock the api module
vi.mock('../api', async () => {
  const mockApi = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
  return {
    __esModule: true,
    default: mockApi,
    api: mockApi,
  }
})

const mockApi = vi.mocked(apiDefault)

describe('mediaApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('searchMedia', () => {
    it('calls GET /media/search with params and returns data', async () => {
      const mockResponse = {
        data: {
          items: [{ id: 1, title: 'Test Movie' }],
          total: 1,
          limit: 20,
          offset: 0,
        },
      }
      mockApi.get.mockResolvedValue(mockResponse)

      const params = { query: 'test', media_type: 'movie', limit: 20 }
      const result = await mediaApi.searchMedia(params)

      expect(mockApi.get).toHaveBeenCalledWith('/media/search', { params })
      expect(result).toEqual(mockResponse.data)
    })

    it('propagates errors on search failure', async () => {
      const error = new Error('Network error')
      mockApi.get.mockRejectedValue(error)

      await expect(mediaApi.searchMedia({ query: 'test' })).rejects.toThrow('Network error')
    })
  })

  describe('getMediaById', () => {
    it('calls GET /media/:id and returns the media item', async () => {
      const mockItem = { id: 42, title: 'Test Movie', media_type: 'movie' }
      mockApi.get.mockResolvedValue({ data: mockItem })

      const result = await mediaApi.getMediaById(42)

      expect(mockApi.get).toHaveBeenCalledWith('/media/42')
      expect(result).toEqual(mockItem)
    })

    it('propagates 404 error for non-existent media', async () => {
      const error = { response: { status: 404 }, message: 'Not found' }
      mockApi.get.mockRejectedValue(error)

      await expect(mediaApi.getMediaById(999)).rejects.toEqual(error)
    })
  })

  describe('getMediaByPath', () => {
    it('calls GET /media/by-path with path param', async () => {
      const mockItem = { id: 1, title: 'Test', directory_path: '/movies/test' }
      mockApi.get.mockResolvedValue({ data: mockItem })

      const result = await mediaApi.getMediaByPath('/movies/test')

      expect(mockApi.get).toHaveBeenCalledWith('/media/by-path', {
        params: { path: '/movies/test' },
      })
      expect(result).toEqual(mockItem)
    })
  })

  describe('analyzeDirectory', () => {
    it('calls POST /media/analyze with directory path', async () => {
      const mockResponse = { message: 'Analysis started', analysis_id: 'abc123' }
      mockApi.post.mockResolvedValue({ data: mockResponse })

      const result = await mediaApi.analyzeDirectory('/media/movies')

      expect(mockApi.post).toHaveBeenCalledWith('/media/analyze', {
        directory_path: '/media/movies',
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('getExternalMetadata', () => {
    it('calls GET /media/:id/metadata', async () => {
      const mockMetadata = [
        { id: 1, provider: 'tmdb', external_id: '12345', title: 'Test' },
      ]
      mockApi.get.mockResolvedValue({ data: mockMetadata })

      const result = await mediaApi.getExternalMetadata(42)

      expect(mockApi.get).toHaveBeenCalledWith('/media/42/metadata')
      expect(result).toEqual(mockMetadata)
    })
  })

  describe('refreshMetadata', () => {
    it('calls POST /media/:id/refresh', async () => {
      mockApi.post.mockResolvedValue({ data: { message: 'Metadata refreshed' } })

      const result = await mediaApi.refreshMetadata(42)

      expect(mockApi.post).toHaveBeenCalledWith('/media/42/refresh')
      expect(result).toEqual({ message: 'Metadata refreshed' })
    })
  })

  describe('getQualityInfo', () => {
    it('calls GET /media/:id/quality', async () => {
      const mockQuality = {
        overall_score: 85,
        resolution: '1080p',
        codec: 'h264',
        file_size: 1500000000,
      }
      mockApi.get.mockResolvedValue({ data: mockQuality })

      const result = await mediaApi.getQualityInfo(42)

      expect(mockApi.get).toHaveBeenCalledWith('/media/42/quality')
      expect(result).toEqual(mockQuality)
    })
  })

  describe('getMediaStats', () => {
    it('calls GET /media/stats and returns statistics', async () => {
      const mockStats = {
        total_items: 100,
        by_type: { movie: 50, tv_show: 30, music: 20 },
        by_quality: { '1080p': 60, '720p': 40 },
        total_size: 500000000000,
        recent_additions: 5,
      }
      mockApi.get.mockResolvedValue({ data: mockStats })

      const result = await mediaApi.getMediaStats()

      expect(mockApi.get).toHaveBeenCalledWith('/media/stats')
      expect(result).toEqual(mockStats)
    })
  })

  describe('getRecentMedia', () => {
    it('calls GET /media/recent with default limit', async () => {
      const mockItems = [{ id: 1, title: 'Recent Movie' }]
      mockApi.get.mockResolvedValue({ data: mockItems })

      const result = await mediaApi.getRecentMedia()

      expect(mockApi.get).toHaveBeenCalledWith('/media/recent', { params: { limit: 10 } })
      expect(result).toEqual(mockItems)
    })

    it('calls GET /media/recent with custom limit', async () => {
      mockApi.get.mockResolvedValue({ data: [] })

      await mediaApi.getRecentMedia(5)

      expect(mockApi.get).toHaveBeenCalledWith('/media/recent', { params: { limit: 5 } })
    })
  })

  describe('getPopularMedia', () => {
    it('calls GET /media/popular with default limit', async () => {
      mockApi.get.mockResolvedValue({ data: [] })

      await mediaApi.getPopularMedia()

      expect(mockApi.get).toHaveBeenCalledWith('/media/popular', { params: { limit: 10 } })
    })

    it('calls GET /media/popular with custom limit', async () => {
      mockApi.get.mockResolvedValue({ data: [] })

      await mediaApi.getPopularMedia(25)

      expect(mockApi.get).toHaveBeenCalledWith('/media/popular', { params: { limit: 25 } })
    })
  })

  describe('deleteMedia', () => {
    it('calls DELETE /media/:id', async () => {
      mockApi.delete.mockResolvedValue({})

      await mediaApi.deleteMedia(42)

      expect(mockApi.delete).toHaveBeenCalledWith('/media/42')
    })

    it('propagates errors on delete failure', async () => {
      mockApi.delete.mockRejectedValue(new Error('Forbidden'))

      await expect(mediaApi.deleteMedia(42)).rejects.toThrow('Forbidden')
    })
  })

  describe('updateMedia', () => {
    it('calls PUT /media/:id with partial data', async () => {
      const updatedItem = { id: 42, title: 'Updated Title', media_type: 'movie' }
      mockApi.put.mockResolvedValue({ data: updatedItem })

      const result = await mediaApi.updateMedia(42, { title: 'Updated Title' } as any)

      expect(mockApi.put).toHaveBeenCalledWith('/media/42', { title: 'Updated Title' })
      expect(result).toEqual(updatedItem)
    })
  })

  describe('Storage root management', () => {
    it('getStorageRoots calls GET /storage/roots and extracts data', async () => {
      const mockRoots = [{ id: 1, name: 'Movies', protocol: 'smb' }]
      mockApi.get.mockResolvedValue({ data: { data: mockRoots } })

      const result = await mediaApi.getStorageRoots()

      expect(mockApi.get).toHaveBeenCalledWith('/storage/roots')
      expect(result).toEqual(mockRoots)
    })

    it('getStorageRoot calls GET /storage/roots/:id', async () => {
      const mockRoot = { id: 1, name: 'Movies', protocol: 'smb' }
      mockApi.get.mockResolvedValue({ data: mockRoot })

      const result = await mediaApi.getStorageRoot(1)

      expect(mockApi.get).toHaveBeenCalledWith('/storage/roots/1')
      expect(result).toEqual(mockRoot)
    })

    it('createStorageRoot calls POST /storage/roots', async () => {
      const newRoot = { name: 'New Root', protocol: 'nfs', enabled: true, max_depth: 5, enable_duplicate_detection: false, enable_metadata_extraction: true }
      const created = { ...newRoot, id: 3, created_at: '2024-01-01', updated_at: '2024-01-01' }
      mockApi.post.mockResolvedValue({ data: created })

      const result = await mediaApi.createStorageRoot(newRoot as any)

      expect(mockApi.post).toHaveBeenCalledWith('/storage/roots', newRoot)
      expect(result).toEqual(created)
    })

    it('updateStorageRoot calls PUT /storage/roots/:id', async () => {
      const updates = { name: 'Updated Root' }
      mockApi.put.mockResolvedValue({ data: { id: 1, ...updates } })

      const result = await mediaApi.updateStorageRoot(1, updates as any)

      expect(mockApi.put).toHaveBeenCalledWith('/storage/roots/1', updates)
      expect(result.name).toBe('Updated Root')
    })

    it('deleteStorageRoot calls DELETE /storage/roots/:id', async () => {
      mockApi.delete.mockResolvedValue({})

      await mediaApi.deleteStorageRoot(1)

      expect(mockApi.delete).toHaveBeenCalledWith('/storage/roots/1')
    })

    it('testStorageRoot calls POST /storage/roots/:id/test', async () => {
      mockApi.post.mockResolvedValue({ data: { success: true, message: 'Connection OK' } })

      const result = await mediaApi.testStorageRoot(1)

      expect(mockApi.post).toHaveBeenCalledWith('/storage/roots/1/test')
      expect(result).toEqual({ success: true, message: 'Connection OK' })
    })
  })

  describe('downloadMedia', () => {
    let createObjectURLMock: vi.Mock
    let revokeObjectURLMock: vi.Mock

    beforeEach(() => {
      createObjectURLMock = vi.fn(() => 'blob:http://localhost/fake-url')
      revokeObjectURLMock = vi.fn()
      Object.defineProperty(window, 'URL', {
        value: {
          createObjectURL: createObjectURLMock,
          revokeObjectURL: revokeObjectURLMock,
        },
        writable: true,
      })
    })

    it('downloads a media file and triggers a link click', async () => {
      const blobData = new Blob(['file content'])
      mockApi.get.mockResolvedValue({ data: blobData })

      const clickMock = vi.fn()
      const appendChildMock = vi.spyOn(document.body, 'appendChild').mockImplementation((node) => node as any)
      const removeChildMock = vi.spyOn(document.body, 'removeChild').mockImplementation((node) => node as any)

      // Mock createElement to capture the anchor
      const originalCreateElement = document.createElement.bind(document)
      vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
        const el = originalCreateElement(tag)
        if (tag === 'a') {
          el.click = clickMock
        }
        return el
      })

      const media = {
        id: 1,
        title: 'Test Movie',
        media_type: 'movie',
        directory_path: '/movies/test-movie.mkv',
        storage_root_name: 'main',
        created_at: '2024-01-01',
        updated_at: '2024-01-01',
      }

      await mediaApi.downloadMedia(media as any)

      expect(mockApi.get).toHaveBeenCalledWith('/download', {
        params: { path: '/movies/test-movie.mkv', storage: 'main' },
        responseType: 'blob',
      })
      expect(clickMock).toHaveBeenCalled()
      expect(revokeObjectURLMock).toHaveBeenCalled()

      appendChildMock.mockRestore()
      removeChildMock.mockRestore()
      ;(document.createElement as vi.Mock).mockRestore()
    })
  })
})
