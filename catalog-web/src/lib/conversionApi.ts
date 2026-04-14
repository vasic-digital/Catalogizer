import type { ConversionJob } from '@/types/conversion';
import { api } from './api';

export const conversionApi = {
  async getConversionJobs(): Promise<ConversionJob[]> {
    const response = await api.get('/conversion/jobs');
    return response.data;
  },

  async startConversion(jobData: Omit<ConversionJob, 'id' | 'status' | 'progress'>): Promise<ConversionJob> {
    const response = await api.post('/conversion/jobs', jobData);
    return response.data;
  },

  async cancelConversion(id: string): Promise<void> {
    await api.delete(`/conversion/jobs/${id}`);
  },

  async retryConversion(id: string): Promise<void> {
    await api.post(`/conversion/jobs/${id}/retry`);
  },

  async downloadFile(id: string): Promise<void> {
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
  }
};
