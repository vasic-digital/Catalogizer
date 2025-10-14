import api from './api'
import type {
  MediaItem,
  MediaSearchRequest,
  MediaSearchResponse,
  ExternalMetadata,
  QualityInfo,
  StorageRoot
} from '@/types/media'

export const mediaApi = {
  searchMedia: (params: MediaSearchRequest): Promise<MediaSearchResponse> =>
    api.get('/media/search', { params }).then((res) => res.data),

  getMediaById: (id: number): Promise<MediaItem> =>
    api.get(`/media/${id}`).then((res) => res.data),

  getMediaByPath: (path: string): Promise<MediaItem> =>
    api.get('/media/by-path', { params: { path } }).then((res) => res.data),

  analyzeDirectory: (path: string): Promise<{ message: string; analysis_id: string }> =>
    api.post('/media/analyze', { directory_path: path }).then((res) => res.data),

  getExternalMetadata: (mediaId: number): Promise<ExternalMetadata[]> =>
    api.get(`/media/${mediaId}/metadata`).then((res) => res.data),

  refreshMetadata: (mediaId: number): Promise<{ message: string }> =>
    api.post(`/media/${mediaId}/refresh`).then((res) => res.data),

  getQualityInfo: (mediaId: number): Promise<QualityInfo> =>
    api.get(`/media/${mediaId}/quality`).then((res) => res.data),

  getMediaStats: (): Promise<{
    total_items: number
    by_type: Record<string, number>
    by_quality: Record<string, number>
    total_size: number
    recent_additions: number
  }> =>
    api.get('/media/stats').then((res) => res.data),

  getRecentMedia: (limit = 10): Promise<MediaItem[]> =>
    api.get('/media/recent', { params: { limit } }).then((res) => res.data),

  getPopularMedia: (limit = 10): Promise<MediaItem[]> =>
    api.get('/media/popular', { params: { limit } }).then((res) => res.data),

  deleteMedia: (id: number): Promise<void> =>
    api.delete(`/media/${id}`).then(() => {/* no content */}),

  updateMedia: (id: number, data: Partial<MediaItem>): Promise<MediaItem> =>
    api.put(`/media/${id}`, data).then((res) => res.data),

  // Storage root management
  getStorageRoots: (): Promise<StorageRoot[]> =>
    api.get('/storage/roots').then((res) => res.data.data),

  getStorageRoot: (id: number): Promise<StorageRoot> =>
    api.get(`/storage/roots/${id}`).then((res) => res.data),

  createStorageRoot: (data: Omit<StorageRoot, 'id' | 'created_at' | 'updated_at'>): Promise<StorageRoot> =>
    api.post('/storage/roots', data).then((res) => res.data),

  updateStorageRoot: (id: number, data: Partial<StorageRoot>): Promise<StorageRoot> =>
    api.put(`/storage/roots/${id}`, data).then((res) => res.data),

  deleteStorageRoot: (id: number): Promise<void> =>
    api.delete(`/storage/roots/${id}`).then(() => {/* no content */}),

  testStorageRoot: (id: number): Promise<{ success: boolean; message: string }> =>
    api.post(`/storage/roots/${id}/test`).then((res) => res.data),
}

export default mediaApi