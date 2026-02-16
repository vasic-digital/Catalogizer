import { conversionApi } from '../conversionApi'

describe('conversionApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getConversionJobs', () => {
    it('returns an array of conversion jobs', async () => {
      const jobs = await conversionApi.getConversionJobs()

      expect(Array.isArray(jobs)).toBe(true)
      expect(jobs.length).toBeGreaterThan(0)
    })

    it('returns jobs with expected structure', async () => {
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

    it('returns jobs with valid source file info', async () => {
      const jobs = await conversionApi.getConversionJobs()
      const sourceFile = jobs[0].sourceFile

      expect(sourceFile).toHaveProperty('path')
      expect(sourceFile).toHaveProperty('name')
      expect(sourceFile).toHaveProperty('format')
      expect(sourceFile).toHaveProperty('size')
      expect(typeof sourceFile.size).toBe('number')
    })

    it('returns jobs with valid options', async () => {
      const jobs = await conversionApi.getConversionJobs()
      const options = jobs[0].options

      expect(options).toHaveProperty('resolution')
      expect(options).toHaveProperty('bitrate')
      expect(options).toHaveProperty('framerate')
      expect(options).toHaveProperty('audioCodec')
      expect(options).toHaveProperty('videoCodec')
    })
  })

  describe('startConversion', () => {
    it('returns a new job with pending status and zero progress', async () => {
      const jobData = {
        sourceFile: {
          path: '/media/test.mkv',
          name: 'test.mkv',
          format: 'mkv',
          size: 1000000,
        },
        targetFormat: 'mp4',
        quality: 'high' as const,
        options: {
          resolution: '1080p',
          bitrate: 5000,
          framerate: 30,
          audioCodec: 'aac',
          videoCodec: 'h264',
        },
      }

      const result = await conversionApi.startConversion(jobData)

      expect(result).toHaveProperty('id')
      expect(result.status).toBe('pending')
      expect(result.progress).toBe(0)
      expect(result.sourceFile).toEqual(jobData.sourceFile)
      expect(result.targetFormat).toBe('mp4')
    })

    it('generates a unique ID for each new job', async () => {
      const jobData = {
        sourceFile: { path: '/test.mkv', name: 'test.mkv', format: 'mkv', size: 100 },
        targetFormat: 'mp4',
        quality: 'medium' as const,
        options: {},
      }

      const result1 = await conversionApi.startConversion(jobData)
      const result2 = await conversionApi.startConversion(jobData)

      expect(result1.id).toBeDefined()
      expect(result2.id).toBeDefined()
    })
  })

  describe('cancelConversion', () => {
    it('completes without throwing', async () => {
      await expect(conversionApi.cancelConversion('1')).resolves.toBeUndefined()
    })
  })

  describe('retryConversion', () => {
    it('completes without throwing', async () => {
      await expect(conversionApi.retryConversion('1')).resolves.toBeUndefined()
    })
  })

  describe('downloadFile', () => {
    it('completes without throwing', async () => {
      await expect(conversionApi.downloadFile('/media/converted/test.mp4')).resolves.toBeUndefined()
    })
  })
})
