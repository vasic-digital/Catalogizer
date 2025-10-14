import { HttpClient } from '../utils/http';
import {
  MediaItem,
  MediaSearchRequest,
  PaginatedResponse,
  MediaStats,
  PlaybackProgress,
  StreamInfo,
  DownloadJob,
} from '../types';

export class MediaService {
  constructor(private http: HttpClient) {}

  /**
   * Search for media items
   */
  public async search(request: MediaSearchRequest = {}): Promise<PaginatedResponse<MediaItem>> {
    const params = new URLSearchParams();

    // Build query parameters
    Object.entries(request).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        params.append(key, value.toString());
      }
    });

    const query = params.toString();
    const endpoint = query ? `/media/search?${query}` : '/media/search';

    return this.http.get<PaginatedResponse<MediaItem>>(endpoint);
  }

  /**
   * Get a specific media item by ID
   */
  public async getById(id: number): Promise<MediaItem> {
    return this.http.get<MediaItem>(`/media/${id}`);
  }

  /**
   * Get media statistics
   */
  public async getStats(): Promise<MediaStats> {
    return this.http.get<MediaStats>('/media/stats');
  }

  /**
   * Get recently added media
   */
  public async getRecentlyAdded(limit = 20): Promise<MediaItem[]> {
    const response = await this.search({
      sort_by: 'created_at',
      sort_order: 'desc',
      limit,
    });
    return response.items;
  }

  /**
   * Get trending/popular media
   */
  public async getTrending(limit = 20): Promise<MediaItem[]> {
    const response = await this.search({
      sort_by: 'rating',
      sort_order: 'desc',
      limit,
    });
    return response.items;
  }

  /**
   * Get media by type
   */
  public async getByType(mediaType: string, limit = 20): Promise<MediaItem[]> {
    const response = await this.search({
      media_type: mediaType,
      sort_by: 'updated_at',
      sort_order: 'desc',
      limit,
    });
    return response.items;
  }

  /**
   * Get user's favorite media
   */
  public async getFavorites(limit = 50): Promise<MediaItem[]> {
    return this.http.get<MediaItem[]>(`/media/favorites?limit=${limit}`);
  }

  /**
   * Toggle favorite status for a media item
   */
  public async toggleFavorite(mediaId: number): Promise<{ is_favorite: boolean }> {
    return this.http.post<{ is_favorite: boolean }>(`/media/${mediaId}/favorite`);
  }

  /**
   * Get continue watching items (items with partial progress)
   */
  public async getContinueWatching(limit = 20): Promise<MediaItem[]> {
    return this.http.get<MediaItem[]>(`/media/continue-watching?limit=${limit}`);
  }

  /**
   * Update playback progress for a media item
   */
  public async updateProgress(mediaId: number, progress: PlaybackProgress): Promise<void> {
    return this.http.put<void>(`/media/${mediaId}/progress`, progress);
  }

  /**
   * Get playback progress for a media item
   */
  public async getProgress(mediaId: number): Promise<PlaybackProgress | null> {
    try {
      return await this.http.get<PlaybackProgress>(`/media/${mediaId}/progress`);
    } catch (error) {
      // Return null if no progress found
      return null;
    }
  }

  /**
   * Mark media as watched (100% progress)
   */
  public async markAsWatched(mediaId: number): Promise<void> {
    const progress: PlaybackProgress = {
      media_id: mediaId,
      position: 100,
      duration: 100,
      timestamp: Date.now(),
    };
    return this.updateProgress(mediaId, progress);
  }

  /**
   * Get streaming URL for a media item
   */
  public async getStreamUrl(mediaId: number): Promise<StreamInfo> {
    return this.http.get<StreamInfo>(`/media/${mediaId}/stream`);
  }

  /**
   * Get download URL for a media item
   */
  public async getDownloadUrl(mediaId: number): Promise<{ url: string; expires_at: string }> {
    return this.http.get<{ url: string; expires_at: string }>(`/media/${mediaId}/download`);
  }

  /**
   * Queue media for download
   */
  public async queueDownload(mediaId: number): Promise<DownloadJob> {
    return this.http.post<DownloadJob>(`/media/${mediaId}/download`);
  }

  /**
   * Get download jobs for the current user
   */
  public async getDownloadJobs(): Promise<DownloadJob[]> {
    return this.http.get<DownloadJob[]>('/media/downloads');
  }

  /**
   * Get specific download job
   */
  public async getDownloadJob(jobId: number): Promise<DownloadJob> {
    return this.http.get<DownloadJob>(`/media/downloads/${jobId}`);
  }

  /**
   * Cancel a download job
   */
  public async cancelDownload(jobId: number): Promise<void> {
    return this.http.post<void>(`/media/downloads/${jobId}/cancel`);
  }

  /**
   * Pause a download job
   */
  public async pauseDownload(jobId: number): Promise<void> {
    return this.http.post<void>(`/media/downloads/${jobId}/pause`);
  }

  /**
   * Resume a download job
   */
  public async resumeDownload(jobId: number): Promise<void> {
    return this.http.post<void>(`/media/downloads/${jobId}/resume`);
  }

  /**
   * Get media thumbnail
   */
  public async getThumbnail(mediaId: number, size?: 'small' | 'medium' | 'large'): Promise<ArrayBuffer> {
    const params = size ? `?size=${size}` : '';
    return this.http.downloadStream(`/media/${mediaId}/thumbnail${params}`);
  }

  /**
   * Get media poster
   */
  public async getPoster(mediaId: number, size?: 'small' | 'medium' | 'large'): Promise<ArrayBuffer> {
    const params = size ? `?size=${size}` : '';
    return this.http.downloadStream(`/media/${mediaId}/poster${params}`);
  }

  /**
   * Get media backdrop
   */
  public async getBackdrop(mediaId: number, size?: 'small' | 'medium' | 'large'): Promise<ArrayBuffer> {
    const params = size ? `?size=${size}` : '';
    return this.http.downloadStream(`/media/${mediaId}/backdrop${params}`);
  }

  /**
   * Update media metadata
   */
  public async updateMetadata(mediaId: number, metadata: Partial<MediaItem>): Promise<MediaItem> {
    return this.http.put<MediaItem>(`/media/${mediaId}`, metadata);
  }

  /**
   * Delete a media item
   */
  public async delete(mediaId: number, deleteFiles = false): Promise<void> {
    const params = deleteFiles ? '?delete_files=true' : '';
    return this.http.delete<void>(`/media/${mediaId}${params}`);
  }

  /**
   * Refresh metadata for a media item
   */
  public async refreshMetadata(mediaId: number): Promise<MediaItem> {
    return this.http.post<MediaItem>(`/media/${mediaId}/refresh`);
  }

  /**
   * Get similar media items
   */
  public async getSimilar(mediaId: number, limit = 10): Promise<MediaItem[]> {
    return this.http.get<MediaItem[]>(`/media/${mediaId}/similar?limit=${limit}`);
  }

  /**
   * Get media recommendations for the user
   */
  public async getRecommendations(limit = 20): Promise<MediaItem[]> {
    return this.http.get<MediaItem[]>(`/media/recommendations?limit=${limit}`);
  }

  /**
   * Rate a media item
   */
  public async rate(mediaId: number, rating: number): Promise<{ rating: number }> {
    return this.http.post<{ rating: number }>(`/media/${mediaId}/rate`, { rating });
  }

  /**
   * Get user's rating for a media item
   */
  public async getRating(mediaId: number): Promise<{ rating: number } | null> {
    try {
      return await this.http.get<{ rating: number }>(`/media/${mediaId}/rating`);
    } catch (error) {
      return null;
    }
  }
}