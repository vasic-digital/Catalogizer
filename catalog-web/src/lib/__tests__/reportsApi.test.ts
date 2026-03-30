import { reportsApi } from '../reportsApi'
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

describe('reportsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getUsageReport', () => {
    it('calls GET /reports/usage without params and returns data', async () => {
      const mockReport = {
        period: '2024-01',
        total_users: 20,
        active_users: 10,
        total_media_accessed: 500,
        total_downloads: 50,
        storage_growth_bytes: 100000,
        top_media: [{ id: 1, title: 'Top Movie', access_count: 100 }],
        user_activity: [{ user_id: 1, username: 'admin', actions: 200 }],
      }
      mockApi.get.mockResolvedValue({ data: mockReport })

      const result = await reportsApi.getUsageReport()

      expect(mockApi.get).toHaveBeenCalledWith('/reports/usage', { params: undefined })
      expect(result).toEqual(mockReport)
    })

    it('calls GET /reports/usage with date range params', async () => {
      const params = { from: '2024-01-01', to: '2024-01-31' }
      mockApi.get.mockResolvedValue({ data: { period: '2024-01' } })

      await reportsApi.getUsageReport(params)

      expect(mockApi.get).toHaveBeenCalledWith('/reports/usage', { params })
    })

    it('propagates errors', async () => {
      mockApi.get.mockRejectedValue(new Error('Report failed'))

      await expect(reportsApi.getUsageReport()).rejects.toThrow('Report failed')
    })
  })

  describe('getPerformanceReport', () => {
    it('calls GET /reports/performance without params and returns data', async () => {
      const mockReport = {
        period: '2024-01',
        avg_response_time_ms: 45,
        p95_response_time_ms: 120,
        p99_response_time_ms: 350,
        total_requests: 50000,
        error_rate: 0.02,
        uptime_percentage: 99.95,
        slowest_endpoints: [{ path: '/api/v1/media/search', avg_ms: 200, count: 1000 }],
      }
      mockApi.get.mockResolvedValue({ data: mockReport })

      const result = await reportsApi.getPerformanceReport()

      expect(mockApi.get).toHaveBeenCalledWith('/reports/performance', { params: undefined })
      expect(result).toEqual(mockReport)
    })

    it('calls GET /reports/performance with date range params', async () => {
      const params = { from: '2024-01-01', to: '2024-01-31' }
      mockApi.get.mockResolvedValue({ data: { period: '2024-01' } })

      await reportsApi.getPerformanceReport(params)

      expect(mockApi.get).toHaveBeenCalledWith('/reports/performance', { params })
    })

    it('propagates errors', async () => {
      mockApi.get.mockRejectedValue(new Error('Server error'))

      await expect(reportsApi.getPerformanceReport()).rejects.toThrow('Server error')
    })
  })
})
