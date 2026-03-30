import { assetUrl, getAssetsByEntity, requestAsset } from '../assetApi'

// fetch is already mocked globally in test-setup.ts
const mockFetch = vi.mocked(global.fetch)

describe('assetApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('assetUrl', () => {
    it('returns the correct asset URL', () => {
      const url = assetUrl('asset-123')
      expect(url).toBe('/api/v1/assets/asset-123')
    })

    it('handles different asset IDs', () => {
      expect(assetUrl('abc')).toBe('/api/v1/assets/abc')
      expect(assetUrl('uuid-456-789')).toBe('/api/v1/assets/uuid-456-789')
    })
  })

  describe('getAssetsByEntity', () => {
    it('calls fetch with correct URL for entity type and ID', async () => {
      const mockAssets = [
        { id: 'a1', type: 'cover', status: 'ready', content_type: 'image/jpeg', size: 1024, entity_type: 'movie', entity_id: '42', created_at: '2024-01-01', updated_at: '2024-01-01' },
      ]
      mockFetch.mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue(mockAssets),
      } as unknown as Response)

      const result = await getAssetsByEntity('movie', 42)

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/assets/by-entity/movie/42',
        { headers: {} }
      )
      expect(result).toEqual(mockAssets)
    })

    it('passes Authorization header when token is provided', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue([]),
      } as unknown as Response)

      await getAssetsByEntity('tv_show', '10', 'my-token-123')

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/assets/by-entity/tv_show/10',
        { headers: { Authorization: 'Bearer my-token-123' } }
      )
    })

    it('returns empty array on non-ok response', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 404,
      } as unknown as Response)

      const result = await getAssetsByEntity('movie', 999)

      expect(result).toEqual([])
    })

    it('accepts string entity ID', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue([]),
      } as unknown as Response)

      await getAssetsByEntity('music_album', 'abc-123')

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/assets/by-entity/music_album/abc-123',
        { headers: {} }
      )
    })
  })

  describe('requestAsset', () => {
    it('calls fetch with POST method and correct body', async () => {
      const request = { type: 'cover', entity_type: 'movie', entity_id: '42' }
      mockFetch.mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ asset_id: 'new-asset-1' }),
      } as unknown as Response)

      const result = await requestAsset(request)

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/assets/request',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(request),
        }
      )
      expect(result).toEqual({ asset_id: 'new-asset-1' })
    })

    it('passes Authorization header when token is provided', async () => {
      const request = { type: 'thumbnail', entity_type: 'tv_show', entity_id: '10' }
      mockFetch.mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ asset_id: 'asset-2' }),
      } as unknown as Response)

      await requestAsset(request, 'my-token')

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/assets/request',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: 'Bearer my-token',
          },
          body: JSON.stringify(request),
        }
      )
    })

    it('throws error on non-ok response', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
      } as unknown as Response)

      const request = { type: 'cover', entity_type: 'movie', entity_id: '1' }

      await expect(requestAsset(request)).rejects.toThrow('Asset request failed: 500')
    })

    it('includes source_hint when provided', async () => {
      const request = { type: 'cover', entity_type: 'movie', entity_id: '1', source_hint: 'tmdb' }
      mockFetch.mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ asset_id: 'asset-3' }),
      } as unknown as Response)

      await requestAsset(request)

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/assets/request',
        expect.objectContaining({
          body: JSON.stringify(request),
        })
      )
    })
  })
})
