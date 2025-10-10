import { HttpClient } from '../utils/http';
import { MediaItem, MediaSearchRequest, PaginatedResponse, MediaStats, PlaybackProgress, StreamInfo, DownloadJob } from '../types';
export declare class MediaService {
    private http;
    constructor(http: HttpClient);
    /**
     * Search for media items
     */
    search(request?: MediaSearchRequest): Promise<PaginatedResponse<MediaItem>>;
    /**
     * Get a specific media item by ID
     */
    getById(id: number): Promise<MediaItem>;
    /**
     * Get media statistics
     */
    getStats(): Promise<MediaStats>;
    /**
     * Get recently added media
     */
    getRecentlyAdded(limit?: number): Promise<MediaItem[]>;
    /**
     * Get trending/popular media
     */
    getTrending(limit?: number): Promise<MediaItem[]>;
    /**
     * Get media by type
     */
    getByType(mediaType: string, limit?: number): Promise<MediaItem[]>;
    /**
     * Get user's favorite media
     */
    getFavorites(limit?: number): Promise<MediaItem[]>;
    /**
     * Toggle favorite status for a media item
     */
    toggleFavorite(mediaId: number): Promise<{
        is_favorite: boolean;
    }>;
    /**
     * Get continue watching items (items with partial progress)
     */
    getContinueWatching(limit?: number): Promise<MediaItem[]>;
    /**
     * Update playback progress for a media item
     */
    updateProgress(mediaId: number, progress: PlaybackProgress): Promise<void>;
    /**
     * Get playback progress for a media item
     */
    getProgress(mediaId: number): Promise<PlaybackProgress | null>;
    /**
     * Mark media as watched (100% progress)
     */
    markAsWatched(mediaId: number): Promise<void>;
    /**
     * Get streaming URL for a media item
     */
    getStreamUrl(mediaId: number): Promise<StreamInfo>;
    /**
     * Get download URL for a media item
     */
    getDownloadUrl(mediaId: number): Promise<{
        url: string;
        expires_at: string;
    }>;
    /**
     * Queue media for download
     */
    queueDownload(mediaId: number): Promise<DownloadJob>;
    /**
     * Get download jobs for the current user
     */
    getDownloadJobs(): Promise<DownloadJob[]>;
    /**
     * Get specific download job
     */
    getDownloadJob(jobId: number): Promise<DownloadJob>;
    /**
     * Cancel a download job
     */
    cancelDownload(jobId: number): Promise<void>;
    /**
     * Pause a download job
     */
    pauseDownload(jobId: number): Promise<void>;
    /**
     * Resume a download job
     */
    resumeDownload(jobId: number): Promise<void>;
    /**
     * Get media thumbnail
     */
    getThumbnail(mediaId: number, size?: 'small' | 'medium' | 'large'): Promise<ArrayBuffer>;
    /**
     * Get media poster
     */
    getPoster(mediaId: number, size?: 'small' | 'medium' | 'large'): Promise<ArrayBuffer>;
    /**
     * Get media backdrop
     */
    getBackdrop(mediaId: number, size?: 'small' | 'medium' | 'large'): Promise<ArrayBuffer>;
    /**
     * Update media metadata
     */
    updateMetadata(mediaId: number, metadata: Partial<MediaItem>): Promise<MediaItem>;
    /**
     * Delete a media item
     */
    delete(mediaId: number, deleteFiles?: boolean): Promise<void>;
    /**
     * Refresh metadata for a media item
     */
    refreshMetadata(mediaId: number): Promise<MediaItem>;
    /**
     * Get similar media items
     */
    getSimilar(mediaId: number, limit?: number): Promise<MediaItem[]>;
    /**
     * Get media recommendations for the user
     */
    getRecommendations(limit?: number): Promise<MediaItem[]>;
    /**
     * Rate a media item
     */
    rate(mediaId: number, rating: number): Promise<{
        rating: number;
    }>;
    /**
     * Get user's rating for a media item
     */
    getRating(mediaId: number): Promise<{
        rating: number;
    } | null>;
}
//# sourceMappingURL=MediaService.d.ts.map