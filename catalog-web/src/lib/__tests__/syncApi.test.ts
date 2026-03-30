import { syncApi } from '../syncApi'
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

describe('syncApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('listEndpoints', () => {
    it('calls GET /sync/endpoints and returns data', async () => {
      const mockEndpoints = [
        { id: 1, name: 'Remote', url: 'https://remote.example.com', enabled: true, created_at: '2024-01-01', updated_at: '2024-01-01' },
      ]
      mockApi.get.mockResolvedValue({ data: mockEndpoints })

      const result = await syncApi.listEndpoints()

      expect(mockApi.get).toHaveBeenCalledWith('/sync/endpoints')
      expect(result).toEqual(mockEndpoints)
    })
  })

  describe('getEndpoint', () => {
    it('calls GET /sync/endpoints/:id and returns data', async () => {
      const mockEndpoint = { id: 1, name: 'Remote', url: 'https://remote.example.com', enabled: true, created_at: '2024-01-01', updated_at: '2024-01-01' }
      mockApi.get.mockResolvedValue({ data: mockEndpoint })

      const result = await syncApi.getEndpoint(1)

      expect(mockApi.get).toHaveBeenCalledWith('/sync/endpoints/1')
      expect(result).toEqual(mockEndpoint)
    })
  })

  describe('createEndpoint', () => {
    it('calls POST /sync/endpoints with body and returns data', async () => {
      const request = { name: 'New Endpoint', url: 'https://new.example.com' }
      const mockResponse = { id: 2, ...request, enabled: true, created_at: '2024-01-01', updated_at: '2024-01-01' }
      mockApi.post.mockResolvedValue({ data: mockResponse })

      const result = await syncApi.createEndpoint(request)

      expect(mockApi.post).toHaveBeenCalledWith('/sync/endpoints', request)
      expect(result).toEqual(mockResponse)
    })

    it('propagates errors on creation failure', async () => {
      mockApi.post.mockRejectedValue(new Error('Creation failed'))

      await expect(syncApi.createEndpoint({ name: 'x', url: 'y' })).rejects.toThrow('Creation failed')
    })
  })

  describe('updateEndpoint', () => {
    it('calls PUT /sync/endpoints/:id with body and returns data', async () => {
      const update = { name: 'Updated Name' }
      const mockResponse = { id: 1, name: 'Updated Name', url: 'https://example.com', enabled: true, created_at: '2024-01-01', updated_at: '2024-01-02' }
      mockApi.put.mockResolvedValue({ data: mockResponse })

      const result = await syncApi.updateEndpoint(1, update)

      expect(mockApi.put).toHaveBeenCalledWith('/sync/endpoints/1', update)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('deleteEndpoint', () => {
    it('calls DELETE /sync/endpoints/:id', async () => {
      mockApi.delete.mockResolvedValue({ data: null })

      await syncApi.deleteEndpoint(1)

      expect(mockApi.delete).toHaveBeenCalledWith('/sync/endpoints/1')
    })

    it('propagates errors on deletion failure', async () => {
      mockApi.delete.mockRejectedValue(new Error('Delete failed'))

      await expect(syncApi.deleteEndpoint(999)).rejects.toThrow('Delete failed')
    })
  })

  describe('startSync', () => {
    it('calls POST /sync/endpoints/:id/sync and returns session', async () => {
      const mockSession = { id: 1, endpoint_id: 1, status: 'running', started_at: '2024-01-01', items_synced: 0, items_failed: 0 }
      mockApi.post.mockResolvedValue({ data: mockSession })

      const result = await syncApi.startSync(1)

      expect(mockApi.post).toHaveBeenCalledWith('/sync/endpoints/1/sync')
      expect(result).toEqual(mockSession)
    })
  })

  describe('getSessions', () => {
    it('calls GET /sync/sessions and returns data', async () => {
      const mockSessions = [
        { id: 1, endpoint_id: 1, status: 'completed', started_at: '2024-01-01', items_synced: 100, items_failed: 0 },
      ]
      mockApi.get.mockResolvedValue({ data: mockSessions })

      const result = await syncApi.getSessions()

      expect(mockApi.get).toHaveBeenCalledWith('/sync/sessions')
      expect(result).toEqual(mockSessions)
    })
  })

  describe('getSession', () => {
    it('calls GET /sync/sessions/:id and returns data', async () => {
      const mockSession = { id: 1, endpoint_id: 1, status: 'completed', started_at: '2024-01-01', items_synced: 50, items_failed: 2 }
      mockApi.get.mockResolvedValue({ data: mockSession })

      const result = await syncApi.getSession(1)

      expect(mockApi.get).toHaveBeenCalledWith('/sync/sessions/1')
      expect(result).toEqual(mockSession)
    })
  })

  describe('getStatistics', () => {
    it('calls GET /sync/statistics and returns data', async () => {
      const mockStats = { total_syncs: 10, successful_syncs: 8, failed_syncs: 2, total_items_synced: 500 }
      mockApi.get.mockResolvedValue({ data: mockStats })

      const result = await syncApi.getStatistics()

      expect(mockApi.get).toHaveBeenCalledWith('/sync/statistics')
      expect(result).toEqual(mockStats)
    })
  })

  describe('scheduleSync', () => {
    it('calls POST /sync/schedules with body', async () => {
      const request = { endpoint_id: 1, cron_expression: '0 0 * * *' }
      mockApi.post.mockResolvedValue({ data: { message: 'Scheduled' } })

      const result = await syncApi.scheduleSync(request)

      expect(mockApi.post).toHaveBeenCalledWith('/sync/schedules', request)
      expect(result).toEqual({ message: 'Scheduled' })
    })
  })

  describe('cleanup', () => {
    it('calls POST /sync/cleanup with body', async () => {
      const request = { older_than_days: 30 }
      mockApi.post.mockResolvedValue({ data: { message: 'Cleaned up' } })

      const result = await syncApi.cleanup(request)

      expect(mockApi.post).toHaveBeenCalledWith('/sync/cleanup', request)
      expect(result).toEqual({ message: 'Cleaned up' })
    })

    it('calls POST /sync/cleanup without body', async () => {
      mockApi.post.mockResolvedValue({ data: { message: 'Cleaned up' } })

      const result = await syncApi.cleanup()

      expect(mockApi.post).toHaveBeenCalledWith('/sync/cleanup', undefined)
      expect(result).toEqual({ message: 'Cleaned up' })
    })
  })
})
