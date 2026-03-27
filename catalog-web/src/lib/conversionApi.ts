import type { ConversionJob } from '@/types/conversion';
import { api } from './api';

const mockJobs: ConversionJob[] = [
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

function isFallbackError(error: unknown): boolean {
  const err = error as { response?: { status?: number } };
  return err.response?.status === 404;
}

export const conversionApi = {
  async getConversionJobs(): Promise<ConversionJob[]> {
    try {
      const response = await api.get('/conversion/jobs');
      return response.data;
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return mockJobs;
      }
      throw error;
    }
  },

  async startConversion(jobData: Omit<ConversionJob, 'id' | 'status' | 'progress'>): Promise<ConversionJob> {
    try {
      const response = await api.post('/conversion/jobs', jobData);
      return response.data;
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        const newJob: ConversionJob = {
          ...jobData,
          id: Date.now().toString(),
          status: 'pending',
          progress: 0
        };
        return newJob;
      }
      throw error;
    }
  },

  async cancelConversion(id: string): Promise<void> {
    try {
      await api.delete(`/conversion/jobs/${id}`);
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return;
      }
      throw error;
    }
  },

  async retryConversion(id: string): Promise<void> {
    try {
      await api.post(`/conversion/jobs/${id}/retry`);
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return;
      }
      throw error;
    }
  },

  async downloadFile(id: string): Promise<void> {
    try {
      const response = await api.get(`/conversion/jobs/${id}/download`, {
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(response.data);
      const link = document.createElement('a');
      link.href = url;
      link.download = `conversion-${id}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return;
      }
      throw error;
    }
  }
};
