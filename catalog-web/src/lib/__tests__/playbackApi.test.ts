import { getEntityProgress, listEntityHistory } from '../playbackApi'
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

describe('playbackApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getEntityProgress', () => {
    it('calls GET /entities/:id/progress and maps raw response to UiPlaybackProgress', async () => {
      mockApi.get.mockResolvedValue({
        data: {
          progress: {
            position_unit: 'seconds',
            duration_total: 7200,
            last_position: 3600,
            last_session_amount: 1800,
            total_reproductions: 5,
            aggregate_amount: 9000,
            last_session_ended_at: '2026-04-10T15:30:00Z',
          },
        },
      })

      const result = await getEntityProgress(42)

      expect(mockApi.get).toHaveBeenCalledWith('/entities/42/progress')
      expect(result).toEqual({
        mediaItemId: 42,
        positionUnit: 'seconds',
        durationTotal: 7200,
        lastPosition: 3600,
        lastSessionAmount: 1800,
        totalReproductions: 5,
        aggregateAmount: 9000,
        lastSessionEndedAt: '2026-04-10T15:30:00Z',
      })
    })

    it('returns null when backend responds with progress: null', async () => {
      mockApi.get.mockResolvedValue({
        data: { progress: null },
      })

      const result = await getEntityProgress(10)

      expect(mockApi.get).toHaveBeenCalledWith('/entities/10/progress')
      expect(result).toBeNull()
    })

    it('returns null when backend responds with empty data', async () => {
      mockApi.get.mockResolvedValue({ data: null })

      const result = await getEntityProgress(10)
      expect(result).toBeNull()
    })

    it('returns null on network error', async () => {
      mockApi.get.mockRejectedValue(new Error('Network error'))

      const result = await getEntityProgress(10)
      expect(result).toBeNull()
    })

    it('applies defaults for missing raw fields', async () => {
      mockApi.get.mockResolvedValue({
        data: {
          progress: {},
        },
      })

      const result = await getEntityProgress(7)

      expect(result).toEqual({
        mediaItemId: 7,
        positionUnit: 'seconds',
        durationTotal: null,
        lastPosition: 0,
        lastSessionAmount: 0,
        totalReproductions: 0,
        aggregateAmount: 0,
        lastSessionEndedAt: null,
      })
    })

    it('preserves explicit null for duration_total', async () => {
      mockApi.get.mockResolvedValue({
        data: {
          progress: {
            position_unit: 'pages',
            duration_total: null,
            last_position: 50,
            last_session_amount: 10,
            total_reproductions: 1,
            aggregate_amount: 50,
            last_session_ended_at: null,
          },
        },
      })

      const result = await getEntityProgress(3)

      expect(result).not.toBeNull()
      expect(result!.positionUnit).toBe('pages')
      expect(result!.durationTotal).toBeNull()
      expect(result!.lastPosition).toBe(50)
      expect(result!.lastSessionEndedAt).toBeNull()
    })
  })

  describe('listEntityHistory', () => {
    it('calls GET /entities/:id/history with default limit and maps sessions', async () => {
      mockApi.get.mockResolvedValue({
        data: {
          sessions: [
            {
              id: 1,
              position_unit: 'seconds',
              start_position: 0,
              end_position: 1800,
              total_amount: 1800,
              started_at: '2026-04-10T14:00:00Z',
              ended_at: '2026-04-10T14:30:00Z',
              completed: true,
            },
            {
              id: 2,
              position_unit: 'seconds',
              start_position: 1800,
              end_position: 3600,
              total_amount: 1800,
              started_at: '2026-04-10T15:00:00Z',
              ended_at: null,
              completed: false,
            },
          ],
        },
      })

      const result = await listEntityHistory(42)

      expect(mockApi.get).toHaveBeenCalledWith('/entities/42/history', {
        params: { limit: 50 },
      })
      expect(result).toHaveLength(2)
      expect(result[0]).toEqual({
        id: 1,
        positionUnit: 'seconds',
        startPosition: 0,
        endPosition: 1800,
        totalAmount: 1800,
        startedAt: '2026-04-10T14:00:00Z',
        endedAt: '2026-04-10T14:30:00Z',
        completed: true,
      })
      expect(result[1]).toEqual({
        id: 2,
        positionUnit: 'seconds',
        startPosition: 1800,
        endPosition: 3600,
        totalAmount: 1800,
        startedAt: '2026-04-10T15:00:00Z',
        endedAt: null,
        completed: false,
      })
    })

    it('passes custom limit parameter', async () => {
      mockApi.get.mockResolvedValue({ data: { sessions: [] } })

      await listEntityHistory(5, 200)

      expect(mockApi.get).toHaveBeenCalledWith('/entities/5/history', {
        params: { limit: 200 },
      })
    })

    it('returns empty array when backend responds with no sessions', async () => {
      mockApi.get.mockResolvedValue({ data: { sessions: [] } })

      const result = await listEntityHistory(10)
      expect(result).toEqual([])
    })

    it('returns empty array when sessions key is missing', async () => {
      mockApi.get.mockResolvedValue({ data: {} })

      const result = await listEntityHistory(10)
      expect(result).toEqual([])
    })

    it('returns empty array on network error', async () => {
      mockApi.get.mockRejectedValue(new Error('Network error'))

      const result = await listEntityHistory(10)
      expect(result).toEqual([])
    })

    it('applies defaults for missing raw session fields', async () => {
      mockApi.get.mockResolvedValue({
        data: {
          sessions: [{}],
        },
      })

      const result = await listEntityHistory(1)

      expect(result).toHaveLength(1)
      expect(result[0]).toEqual({
        id: 0,
        positionUnit: 'seconds',
        startPosition: 0,
        endPosition: 0,
        totalAmount: 0,
        startedAt: '',
        endedAt: null,
        completed: false,
      })
    })

    it('maps multiple sessions with different position units', async () => {
      mockApi.get.mockResolvedValue({
        data: {
          sessions: [
            {
              id: 10,
              position_unit: 'pages',
              start_position: 0,
              end_position: 100,
              total_amount: 100,
              started_at: '2026-01-01T00:00:00Z',
              ended_at: '2026-01-01T01:00:00Z',
              completed: true,
            },
            {
              id: 11,
              position_unit: 'events',
              start_position: 0,
              end_position: 1,
              total_amount: 1,
              started_at: '2026-01-02T00:00:00Z',
              ended_at: null,
              completed: false,
            },
          ],
        },
      })

      const result = await listEntityHistory(99)

      expect(result).toHaveLength(2)
      expect(result[0].positionUnit).toBe('pages')
      expect(result[0].completed).toBe(true)
      expect(result[1].positionUnit).toBe('events')
      expect(result[1].completed).toBe(false)
    })
  })
})
