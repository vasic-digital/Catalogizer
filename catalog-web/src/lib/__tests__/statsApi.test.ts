import { statsApi } from '../statsApi'
import apiDefault from '../api'

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

describe('statsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getOverallStats', () => {
    it('calls GET /stats/overall and returns data', async () => {
      const mockStats = { total_files: 5000, total_size: 1000000000, total_directories: 200, total_storage_roots: 3 }
      mockApi.get.mockResolvedValue({ data: mockStats })

      const result = await statsApi.getOverallStats()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/overall')
      expect(result).toEqual(mockStats)
    })

    it('propagates errors', async () => {
      mockApi.get.mockRejectedValue(new Error('Server error'))

      await expect(statsApi.getOverallStats()).rejects.toThrow('Server error')
    })
  })

  describe('getFileTypeStats', () => {
    it('calls GET /stats/filetypes and returns data', async () => {
      const mockStats = [
        { extension: '.mp4', count: 100, total_size: 500000 },
        { extension: '.mkv', count: 50, total_size: 300000 },
      ]
      mockApi.get.mockResolvedValue({ data: mockStats })

      const result = await statsApi.getFileTypeStats()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/filetypes')
      expect(result).toEqual(mockStats)
    })
  })

  describe('getSizeDistribution', () => {
    it('calls GET /stats/sizes and returns data', async () => {
      const mockDist = [
        { range: '0-100MB', count: 200, total_size: 5000000 },
        { range: '100MB-1GB', count: 100, total_size: 50000000 },
      ]
      mockApi.get.mockResolvedValue({ data: mockDist })

      const result = await statsApi.getSizeDistribution()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/sizes')
      expect(result).toEqual(mockDist)
    })
  })

  describe('getDuplicateStats', () => {
    it('calls GET /stats/duplicates and returns data', async () => {
      const mockStats = { total_duplicate_groups: 10, total_duplicate_files: 25, wasted_space: 1000000 }
      mockApi.get.mockResolvedValue({ data: mockStats })

      const result = await statsApi.getDuplicateStats()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/duplicates')
      expect(result).toEqual(mockStats)
    })
  })

  describe('getTopDuplicateGroups', () => {
    it('calls GET /stats/duplicates/groups with default limit', async () => {
      const mockGroups = [{ hash: 'abc123', count: 3, file_size: 5000, paths: ['/a', '/b', '/c'] }]
      mockApi.get.mockResolvedValue({ data: mockGroups })

      const result = await statsApi.getTopDuplicateGroups()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/duplicates/groups', { params: { limit: 10 } })
      expect(result).toEqual(mockGroups)
    })

    it('calls GET /stats/duplicates/groups with custom limit', async () => {
      mockApi.get.mockResolvedValue({ data: [] })

      await statsApi.getTopDuplicateGroups(5)

      expect(mockApi.get).toHaveBeenCalledWith('/stats/duplicates/groups', { params: { limit: 5 } })
    })
  })

  describe('getAccessPatterns', () => {
    it('calls GET /stats/access and returns data', async () => {
      const mockPatterns = [
        { hour: 9, count: 50 },
        { hour: 14, count: 120 },
      ]
      mockApi.get.mockResolvedValue({ data: mockPatterns })

      const result = await statsApi.getAccessPatterns()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/access')
      expect(result).toEqual(mockPatterns)
    })
  })

  describe('getGrowthTrends', () => {
    it('calls GET /stats/growth with default days', async () => {
      const mockTrends = [{ date: '2024-01-01', files_added: 10, size_added: 5000, cumulative_files: 100, cumulative_size: 50000 }]
      mockApi.get.mockResolvedValue({ data: mockTrends })

      const result = await statsApi.getGrowthTrends()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/growth', { params: { days: 30 } })
      expect(result).toEqual(mockTrends)
    })

    it('calls GET /stats/growth with custom days', async () => {
      mockApi.get.mockResolvedValue({ data: [] })

      await statsApi.getGrowthTrends(7)

      expect(mockApi.get).toHaveBeenCalledWith('/stats/growth', { params: { days: 7 } })
    })
  })

  describe('getScanHistory', () => {
    it('calls GET /stats/scans with default limit', async () => {
      const mockHistory = [{ id: 1, storage_root_id: 1, storage_root_name: 'NAS', started_at: '2024-01-01', status: 'completed', files_found: 100, files_new: 10, files_updated: 5, errors: 0, duration_ms: 5000 }]
      mockApi.get.mockResolvedValue({ data: mockHistory })

      const result = await statsApi.getScanHistory()

      expect(mockApi.get).toHaveBeenCalledWith('/stats/scans', { params: { limit: 20 } })
      expect(result).toEqual(mockHistory)
    })

    it('calls GET /stats/scans with custom limit', async () => {
      mockApi.get.mockResolvedValue({ data: [] })

      await statsApi.getScanHistory(5)

      expect(mockApi.get).toHaveBeenCalledWith('/stats/scans', { params: { limit: 5 } })
    })
  })
})
