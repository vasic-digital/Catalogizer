import { scansApi } from '../scansApi'
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

describe('scansApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('listScans', () => {
    it('calls GET /scans and returns data', async () => {
      const mockScans = [
        { job_id: 'scan-1', status: 'completed', files_found: 100, files_processed: 100, errors: 0, progress: 100, storage_root_id: 1 },
        { job_id: 'scan-2', status: 'running', files_found: 50, files_processed: 25, errors: 0, progress: 50, storage_root_id: 2 },
      ]
      mockApi.get.mockResolvedValue({ data: mockScans })

      const result = await scansApi.listScans()

      expect(mockApi.get).toHaveBeenCalledWith('/scans')
      expect(result).toEqual(mockScans)
    })

    it('propagates errors', async () => {
      mockApi.get.mockRejectedValue(new Error('Network error'))

      await expect(scansApi.listScans()).rejects.toThrow('Network error')
    })
  })

  describe('queueScan', () => {
    it('calls POST /scans with body and returns data', async () => {
      const request = { storage_root_id: 1, full_scan: true }
      const mockResponse = { job_id: 'scan-3', status: 'queued', storage_root_id: 1, files_found: 0, files_processed: 0, errors: 0, progress: 0 }
      mockApi.post.mockResolvedValue({ data: mockResponse })

      const result = await scansApi.queueScan(request)

      expect(mockApi.post).toHaveBeenCalledWith('/scans', request)
      expect(result).toEqual(mockResponse)
    })

    it('propagates errors on queue failure', async () => {
      mockApi.post.mockRejectedValue(new Error('Failed to queue'))

      await expect(scansApi.queueScan({ storage_root_id: 1 })).rejects.toThrow('Failed to queue')
    })
  })

  describe('getScanStatus', () => {
    it('calls GET /scans/:jobId and returns data', async () => {
      const mockScan = { job_id: 'scan-1', status: 'completed', files_found: 100, files_processed: 100, errors: 0, progress: 100, storage_root_id: 1 }
      mockApi.get.mockResolvedValue({ data: mockScan })

      const result = await scansApi.getScanStatus('scan-1')

      expect(mockApi.get).toHaveBeenCalledWith('/scans/scan-1')
      expect(result).toEqual(mockScan)
    })

    it('propagates errors on status check failure', async () => {
      mockApi.get.mockRejectedValue(new Error('Not found'))

      await expect(scansApi.getScanStatus('invalid-id')).rejects.toThrow('Not found')
    })
  })
})
