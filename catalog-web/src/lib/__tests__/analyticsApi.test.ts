import { analyticsApi } from '../analyticsApi'
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

describe('analyticsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('trackAccess', () => {
    it('calls POST /analytics/access with body and returns data', async () => {
      const event = { media_id: 1, action: 'play' as const, duration_ms: 5000 }
      mockApi.post.mockResolvedValue({ data: { message: 'Tracked' } })

      const result = await analyticsApi.trackAccess(event)

      expect(mockApi.post).toHaveBeenCalledWith('/analytics/access', event)
      expect(result).toEqual({ message: 'Tracked' })
    })

    it('propagates errors', async () => {
      mockApi.post.mockRejectedValue(new Error('Tracking failed'))

      await expect(analyticsApi.trackAccess({ media_id: 1, action: 'view' })).rejects.toThrow('Tracking failed')
    })
  })

  describe('trackEvent', () => {
    it('calls POST /analytics/event with body and returns data', async () => {
      const event = { event_type: 'search', event_data: { query: 'test' } }
      mockApi.post.mockResolvedValue({ data: { message: 'Event tracked' } })

      const result = await analyticsApi.trackEvent(event)

      expect(mockApi.post).toHaveBeenCalledWith('/analytics/event', event)
      expect(result).toEqual({ message: 'Event tracked' })
    })
  })

  describe('getUserAnalytics', () => {
    it('calls GET /analytics/user/:userId and returns data', async () => {
      const mockAnalytics = {
        user_id: 1,
        total_views: 100,
        total_plays: 50,
        total_downloads: 10,
        favorite_types: { movie: 30, music: 20 },
        recent_activity: [{ date: '2024-01-01', count: 5 }],
      }
      mockApi.get.mockResolvedValue({ data: mockAnalytics })

      const result = await analyticsApi.getUserAnalytics(1)

      expect(mockApi.get).toHaveBeenCalledWith('/analytics/user/1')
      expect(result).toEqual(mockAnalytics)
    })

    it('propagates errors for invalid user', async () => {
      mockApi.get.mockRejectedValue(new Error('User not found'))

      await expect(analyticsApi.getUserAnalytics(999)).rejects.toThrow('User not found')
    })
  })

  describe('getSystemAnalytics', () => {
    it('calls GET /analytics/system and returns data', async () => {
      const mockAnalytics = {
        total_events: 1000,
        active_users_24h: 5,
        popular_media: [{ media_id: 1, title: 'Top Movie', access_count: 50 }],
        events_by_type: { view: 500, play: 300, download: 200 },
      }
      mockApi.get.mockResolvedValue({ data: mockAnalytics })

      const result = await analyticsApi.getSystemAnalytics()

      expect(mockApi.get).toHaveBeenCalledWith('/analytics/system')
      expect(result).toEqual(mockAnalytics)
    })
  })

  describe('getMediaAnalytics', () => {
    it('calls GET /analytics/media/:mediaId and returns data', async () => {
      const mockAnalytics = {
        media_id: 42,
        view_count: 200,
        play_count: 150,
        download_count: 30,
        unique_users: 10,
      }
      mockApi.get.mockResolvedValue({ data: mockAnalytics })

      const result = await analyticsApi.getMediaAnalytics(42)

      expect(mockApi.get).toHaveBeenCalledWith('/analytics/media/42')
      expect(result).toEqual(mockAnalytics)
    })
  })

  describe('createReport', () => {
    it('calls POST /analytics/reports with body and returns data', async () => {
      const request = { report_type: 'usage', date_range: { from: '2024-01-01', to: '2024-01-31' } }
      mockApi.post.mockResolvedValue({ data: { report_id: 'rpt-123', message: 'Report created' } })

      const result = await analyticsApi.createReport(request)

      expect(mockApi.post).toHaveBeenCalledWith('/analytics/reports', request)
      expect(result).toEqual({ report_id: 'rpt-123', message: 'Report created' })
    })

    it('calls POST /analytics/reports without date range', async () => {
      const request = { report_type: 'performance' }
      mockApi.post.mockResolvedValue({ data: { report_id: 'rpt-456', message: 'Report created' } })

      const result = await analyticsApi.createReport(request)

      expect(mockApi.post).toHaveBeenCalledWith('/analytics/reports', request)
      expect(result.report_id).toBe('rpt-456')
    })
  })
})
