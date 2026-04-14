import { conversionApi } from '../conversionApi'
import { api } from '../api'

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

const mockApi = vi.mocked(api)

describe('conversionApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getConversionJobs', () => {
    it('calls GET /conversion/jobs and returns jobs array', async () => {
      const mockJobs = [
        {
          id: '1',
          sourceFile: { path: '/media/test.mkv', name: 'test.mkv', format: 'mkv', size: 1000000 },
          targetFormat: 'mp4',
          quality: 'high',
          status: 'completed',
          progress: 100,
          options: { resolution: '1080p', bitrate: 5000, framerate: 30, audioCodec: 'aac', videoCodec: 'h264' },
        },
      ]
      mockApi.get.mockResolvedValue({ data: mockJobs })

      const result = await conversionApi.getConversionJobs()

      expect(mockApi.get).toHaveBeenCalledWith('/conversion/jobs')
      expect(result).toEqual(mockJobs)
    })

    it('returns jobs with expected structure', async () => {
      const mockJobs = [
        {
          id: '1',
          sourceFile: { path: '/media/test.mkv', name: 'test.mkv', format: 'mkv', size: 1073741824 },
          targetFormat: 'mp4',
          quality: 'high',
          status: 'completed',
          progress: 100,
          startTime: '2023-12-09T10:00:00Z',
          endTime: '2023-12-09T10:15:00Z',
          outputFile: '/media/converted/test.mp4',
          options: { resolution: '1080p', bitrate: 5000, framerate: 30, audioCodec: 'aac', videoCodec: 'h264' },
        },
      ]
      mockApi.get.mockResolvedValue({ data: mockJobs })

      const jobs = await conversionApi.getConversionJobs()
      const job = jobs[0]

      expect(job).toHaveProperty('id')
      expect(job).toHaveProperty('sourceFile')
      expect(job).toHaveProperty('targetFormat')
      expect(job).toHaveProperty('quality')
      expect(job).toHaveProperty('status')
      expect(job).toHaveProperty('progress')
      expect(job).toHaveProperty('options')
    })

    it('propagates errors', async () => {
      const error = new Error('Server error')
      ;(error as any).response = { status: 500 }
      mockApi.get.mockRejectedValue(error)

      await expect(conversionApi.getConversionJobs()).rejects.toThrow('Server error')
    })
  })

  describe('startConversion', () => {
    it('calls POST /conversion/jobs with job data and returns new job', async () => {
      const jobData = {
        sourceFile: { path: '/media/test.mkv', name: 'test.mkv', format: 'mkv', size: 1000000 },
        targetFormat: 'mp4',
        quality: 'high' as const,
        options: { resolution: '1080p', bitrate: 5000, framerate: 30, audioCodec: 'aac', videoCodec: 'h264' },
      }
      const created = { ...jobData, id: '3', status: 'pending', progress: 0 }
      mockApi.post.mockResolvedValue({ data: created })

      const result = await conversionApi.startConversion(jobData)

      expect(mockApi.post).toHaveBeenCalledWith('/conversion/jobs', jobData)
      expect(result).toEqual(created)
      expect(result.status).toBe('pending')
      expect(result.progress).toBe(0)
    })
  })

  describe('cancelConversion', () => {
    it('calls DELETE /conversion/jobs/:id', async () => {
      mockApi.delete.mockResolvedValue({})

      await conversionApi.cancelConversion('1')

      expect(mockApi.delete).toHaveBeenCalledWith('/conversion/jobs/1')
    })

    it('propagates errors', async () => {
      const error = new Error('Not found')
      ;(error as any).response = { status: 404 }
      mockApi.delete.mockRejectedValue(error)

      await expect(conversionApi.cancelConversion('999')).rejects.toThrow('Not found')
    })
  })

  describe('retryConversion', () => {
    it('calls POST /conversion/jobs/:id/retry', async () => {
      mockApi.post.mockResolvedValue({})

      await conversionApi.retryConversion('1')

      expect(mockApi.post).toHaveBeenCalledWith('/conversion/jobs/1/retry')
    })

    it('propagates errors', async () => {
      const error = new Error('Cannot retry')
      ;(error as any).response = { status: 400 }
      mockApi.post.mockRejectedValue(error)

      await expect(conversionApi.retryConversion('1')).rejects.toThrow('Cannot retry')
    })
  })

  describe('downloadFile', () => {
    it('calls GET /conversion/jobs/:id/download with blob response type', async () => {
      const blob = new Blob(['file content'], { type: 'video/mp4' })
      mockApi.get.mockResolvedValue({ data: blob })

      // Mock DOM APIs for download
      const mockCreateObjectURL = vi.fn(() => 'blob:mock-url')
      const mockRevokeObjectURL = vi.fn()
      const mockClick = vi.fn()
      const mockAppendChild = vi.fn()
      const mockRemoveChild = vi.fn()
      const mockCreateElement = vi.fn(() => ({
        href: '',
        download: '',
        click: mockClick,
      }))

      Object.defineProperty(window, 'URL', {
        value: { createObjectURL: mockCreateObjectURL, revokeObjectURL: mockRevokeObjectURL },
        writable: true,
      })
      vi.spyOn(document, 'createElement').mockImplementation(mockCreateElement as any)
      vi.spyOn(document.body, 'appendChild').mockImplementation(mockAppendChild)
      vi.spyOn(document.body, 'removeChild').mockImplementation(mockRemoveChild)

      await conversionApi.downloadFile('1')

      expect(mockApi.get).toHaveBeenCalledWith('/conversion/jobs/1/download', {
        responseType: 'blob',
      })
      expect(mockCreateObjectURL).toHaveBeenCalledWith(blob)
      expect(mockClick).toHaveBeenCalled()
      expect(mockRevokeObjectURL).toHaveBeenCalledWith('blob:mock-url')

      vi.restoreAllMocks()
    })

    it('propagates errors', async () => {
      const error = new Error('File not found')
      ;(error as any).response = { status: 404 }
      mockApi.get.mockRejectedValue(error)

      await expect(conversionApi.downloadFile('999')).rejects.toThrow('File not found')
    })
  })
})
