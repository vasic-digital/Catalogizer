import { HttpClient } from '../utils/http';
import { MediaItem, MediaSearchRequest, PaginatedResponse, MediaStats, PlaybackProgress, StreamInfo } from '../types';
export declare class MediaService {
    private http;
    constructor(http: HttpClient);
    /**
     * Search for media items
     * Backend: GET /api/v1/media/search
     */
    search(request?: MediaSearchRequest): Promise<PaginatedResponse<MediaItem>>;
    /**
     * Get a specific media item by ID
     * Backend: GET /api/v1/media/:id
     */
    getById(id: number): Promise<MediaItem>;
    /**
     * Get media statistics
     * Backend: GET /api/v1/media/stats
     */
    getStats(): Promise<MediaStats>;
    /**
     * Get recently added media
     */
    getRecentlyAdded(limit?: number): Promise<MediaItem[]>;
    /**
     * Get trending/popular media
     * Backend: GET /api/v1/recommendations/trending
     */
    getTrending(limit?: number): Promise<MediaItem[]>;
    /**
     * Get media by type
     */
    getByType(mediaType: string, limit?: number): Promise<MediaItem[]>;
    /**
     * Get user's favorite media
     * Backend: GET /api/v1/favorites
     */
    getFavorites(limit?: number): Promise<MediaItem[]>;
    /**
     * Add media item to favorites
     * Backend: POST /api/v1/favorites
     */
    addFavorite(entityType: string, entityId: number): Promise<void>;
    /**
     * Remove media item from favorites
     * Backend: DELETE /api/v1/favorites/:entity_type/:entity_id
     */
    removeFavorite(entityType: string, entityId: number): Promise<void>;
    /**
     * Check if media item is favorited
     * Backend: GET /api/v1/favorites/check/:entity_type/:entity_id
     */
    checkFavorite(entityType: string, entityId: number): Promise<{
        is_favorite: boolean;
    }>;
    /**
     * Update playback progress for a media item
     * Backend: PUT /api/v1/media/:id/progress
     */
    updateProgress(mediaId: number, progress: PlaybackProgress): Promise<void>;
    /**
     * Mark media as watched (100% progress)
     */
    markAsWatched(mediaId: number): Promise<void>;
    /**
     * Toggle favorite status for a media item
     * Backend: PUT /api/v1/media/:id/favorite
     */
    toggleFavorite(mediaId: number): Promise<{
        is_favorite: boolean;
    }>;
    /**
     * Get streaming URL for a media entity
     * Backend: GET /api/v1/entities/:id/stream
     */
    getStreamUrl(entityId: number): Promise<StreamInfo>;
    /**
     * Get download URL for a media entity
     * Backend: GET /api/v1/entities/:id/download
     */
    getDownloadUrl(entityId: number): Promise<{
        url: string;
        expires_at: string;
    }>;
    /**
     * Download a file by ID
     * Backend: GET /api/v1/download/file/:id
     */
    downloadFile(fileId: number): Promise<ArrayBuffer>;
    /**
     * Get asset (cover art, thumbnail) for a media entity
     * Backend: GET /api/v1/assets/by-entity/:type/:id
     */
    getEntityAsset(entityType: string, entityId: number): Promise<ArrayBuffer>;
    /**
     * Refresh metadata for a media entity
     * Backend: POST /api/v1/entities/:id/metadata/refresh
     */
    refreshMetadata(entityId: number): Promise<MediaItem>;
    /**
     * Get similar media items
     * Backend: GET /api/v1/recommendations/similar/:media_id
     */
    getSimilar(mediaId: number, limit?: number): Promise<MediaItem[]>;
    /**
     * Get personalized recommendations for a user
     * Backend: GET /api/v1/recommendations/personalized/:user_id
     */
    getRecommendations(userId: number, limit?: number): Promise<MediaItem[]>;
    /**
     * Get media entity children (e.g., seasons of a show)
     * Backend: GET /api/v1/entities/:id/children
     */
    getChildren(entityId: number): Promise<MediaItem[]>;
    /**
     * Get files associated with a media entity
     * Backend: GET /api/v1/entities/:id/files
     */
    getEntityFiles(entityId: number): Promise<unknown[]>;
    /**
     * Get entity metadata from external providers
     * Backend: GET /api/v1/entities/:id/metadata
     */
    getEntityMetadata(entityId: number): Promise<unknown>;
    /**
     * Update user metadata for an entity
     * Backend: PUT /api/v1/entities/:id/user-metadata
     */
    updateUserMetadata(entityId: number, metadata: Record<string, unknown>): Promise<void>;
}
//# sourceMappingURL=MediaService.d.ts.map