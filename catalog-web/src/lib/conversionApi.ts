import type { ConversionJob } from '@/types/conversion';

export const conversionApi = {
  async getConversionJobs(): Promise<ConversionJob[]> {
    // Mock implementation - would be replaced with actual API call
    return [
      {
        id: '1',
        sourceFile: {
          path: '/media/movies/sample.mkv',
          name: 'sample.mkv',
          format: 'mkv',
          size: 1073741824 // 1GB
        },
        targetFormat: 'mp4',
        quality: 'high',
        status: 'completed',
        progress: 100,
        startTime: '2023-12-09T10:00:00Z',
        endTime: '2023-12-09T10:15:00Z',
        outputFile: '/media/converted/sample.mp4',
        options: {
          resolution: '1080p',
          bitrate: 5000,
          framerate: 30,
          audioCodec: 'aac',
          videoCodec: 'h264'
        }
      },
      {
        id: '2',
        sourceFile: {
          path: '/media/movies/another-video.avi',
          name: 'another-video.avi',
          format: 'avi',
          size: 2147483648 // 2GB
        },
        targetFormat: 'mp4',
        quality: 'medium',
        status: 'processing',
        progress: 65,
        startTime: '2023-12-09T11:30:00Z',
        options: {
          resolution: '720p',
          bitrate: 2500,
          framerate: 30,
          audioCodec: 'aac',
          videoCodec: 'h264'
        }
      }
    ];
  },

  async startConversion(jobData: Omit<ConversionJob, 'id' | 'status' | 'progress'>): Promise<ConversionJob> {
    // Mock implementation
    const newJob: ConversionJob = {
      ...jobData,
      id: Date.now().toString(),
      status: 'pending',
      progress: 0
    };
    return newJob;
  },

  async cancelConversion(id: string): Promise<void> {
    // Mock implementation
    console.log(`Cancelling conversion job ${id}`);
  },

  async retryConversion(id: string): Promise<void> {
    // Mock implementation
    console.log(`Retrying conversion job ${id}`);
  },

  async downloadFile(path: string): Promise<void> {
    // Mock implementation - would trigger file download
    console.log(`Downloading file from ${path}`);
  }
};