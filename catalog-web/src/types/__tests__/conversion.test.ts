import type { ConversionJob } from '../conversion'

describe('conversion types', () => {
  describe('ConversionJob interface', () => {
    it('has all required fields', () => {
      const job: ConversionJob = {
        id: 'job-1',
        sourceFile: {
          path: '/media/movies/movie.avi',
          name: 'movie.avi',
          format: 'avi',
          size: 1500000000,
        },
        targetFormat: 'mp4',
        quality: 'high',
        status: 'pending',
        progress: 0,
        options: {},
      }
      expect(job.id).toBe('job-1')
      expect(job.sourceFile.name).toBe('movie.avi')
      expect(job.targetFormat).toBe('mp4')
      expect(job.quality).toBe('high')
      expect(job.status).toBe('pending')
      expect(job.progress).toBe(0)
      expect(job.startTime).toBeUndefined()
      expect(job.endTime).toBeUndefined()
      expect(job.outputFile).toBeUndefined()
      expect(job.error).toBeUndefined()
    })

    it('accepts all quality values', () => {
      const qualities: ConversionJob['quality'][] = ['low', 'medium', 'high', 'ultra']
      expect(qualities).toHaveLength(4)
    })

    it('accepts all status values', () => {
      const statuses: ConversionJob['status'][] = [
        'pending', 'processing', 'completed', 'failed', 'cancelled',
      ]
      expect(statuses).toHaveLength(5)
    })

    it('accepts a completed job with all fields', () => {
      const job: ConversionJob = {
        id: 'job-2',
        sourceFile: {
          path: '/media/movies/old.avi',
          name: 'old.avi',
          format: 'avi',
          size: 2000000000,
        },
        targetFormat: 'mkv',
        quality: 'ultra',
        status: 'completed',
        progress: 100,
        startTime: '2024-06-01T10:00:00Z',
        endTime: '2024-06-01T10:15:00Z',
        outputFile: '/media/movies/old.mkv',
        options: {
          resolution: '1920x1080',
          bitrate: 8000,
          framerate: 24,
          audioCodec: 'aac',
          videoCodec: 'h265',
        },
      }
      expect(job.status).toBe('completed')
      expect(job.progress).toBe(100)
      expect(job.outputFile).toBe('/media/movies/old.mkv')
      expect(job.options.resolution).toBe('1920x1080')
      expect(job.options.bitrate).toBe(8000)
      expect(job.options.framerate).toBe(24)
      expect(job.options.audioCodec).toBe('aac')
      expect(job.options.videoCodec).toBe('h265')
    })

    it('accepts a failed job with error', () => {
      const job: ConversionJob = {
        id: 'job-3',
        sourceFile: {
          path: '/media/corrupt.avi',
          name: 'corrupt.avi',
          format: 'avi',
          size: 100,
        },
        targetFormat: 'mp4',
        quality: 'medium',
        status: 'failed',
        progress: 23,
        startTime: '2024-06-01T10:00:00Z',
        endTime: '2024-06-01T10:00:05Z',
        error: 'Invalid input file: corrupted header',
        options: {},
      }
      expect(job.status).toBe('failed')
      expect(job.error).toBe('Invalid input file: corrupted header')
      expect(job.progress).toBe(23)
    })

    it('options fields are all optional', () => {
      const job: ConversionJob = {
        id: 'job-4',
        sourceFile: { path: '/f', name: 'f', format: 'f', size: 0 },
        targetFormat: 'mp4',
        quality: 'low',
        status: 'pending',
        progress: 0,
        options: {},
      }
      expect(job.options.resolution).toBeUndefined()
      expect(job.options.bitrate).toBeUndefined()
      expect(job.options.framerate).toBeUndefined()
      expect(job.options.audioCodec).toBeUndefined()
      expect(job.options.videoCodec).toBeUndefined()
    })
  })
})
