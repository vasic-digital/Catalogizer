"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.MediaService = void 0;
class MediaService {
    constructor(http) {
        this.http = http;
    }
    /**
     * Search for media items
     * Backend: GET /api/v1/media/search
     */
    async search(request = {}) {
        const params = new URLSearchParams();
        // Build query parameters
        Object.entries(request).forEach(([key, value]) => {
            if (value !== undefined && value !== null) {
                params.append(key, value.toString());
            }
        });
        const query = params.toString();
        const endpoint = query ? `/media/search?${query}` : '/media/search';
        return this.http.get(endpoint);
    }
    /**
     * Get a specific media item by ID
     * Backend: GET /api/v1/media/:id
     */
    async getById(id) {
        return this.http.get(`/media/${id}`);
    }
    /**
     * Get media statistics
     * Backend: GET /api/v1/media/stats
     */
    async getStats() {
        return this.http.get('/media/stats');
    }
    /**
     * Get recently added media
     */
    async getRecentlyAdded(limit = 20) {
        const response = await this.search({
            sort_by: 'created_at',
            sort_order: 'desc',
            limit,
        });
        return response.items;
    }
    /**
     * Get trending/popular media
     * Backend: GET /api/v1/recommendations/trending
     */
    async getTrending(limit = 20) {
        return this.http.get(`/recommendations/trending?limit=${limit}`);
    }
    /**
     * Get media by type
     */
    async getByType(mediaType, limit = 20) {
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
     * Backend: GET /api/v1/favorites
     */
    async getFavorites(limit = 50) {
        return this.http.get(`/favorites?limit=${limit}`);
    }
    /**
     * Add media item to favorites
     * Backend: POST /api/v1/favorites
     */
    async addFavorite(entityType, entityId) {
        return this.http.post('/favorites', { entity_type: entityType, entity_id: entityId });
    }
    /**
     * Remove media item from favorites
     * Backend: DELETE /api/v1/favorites/:entity_type/:entity_id
     */
    async removeFavorite(entityType, entityId) {
        return this.http.delete(`/favorites/${entityType}/${entityId}`);
    }
    /**
     * Check if media item is favorited
     * Backend: GET /api/v1/favorites/check/:entity_type/:entity_id
     */
    async checkFavorite(entityType, entityId) {
        return this.http.get(`/favorites/check/${entityType}/${entityId}`);
    }
    /**
     * Update playback progress for a media item
     * Backend: PUT /api/v1/media/:id/progress
     */
    async updateProgress(mediaId, progress) {
        return this.http.put(`/media/${mediaId}/progress`, progress);
    }
    /**
     * Mark media as watched (100% progress)
     */
    async markAsWatched(mediaId) {
        const progress = {
            media_id: mediaId,
            position: 100,
            duration: 100,
            timestamp: Date.now(),
        };
        return this.updateProgress(mediaId, progress);
    }
    /**
     * Toggle favorite status for a media item
     * Backend: PUT /api/v1/media/:id/favorite
     */
    async toggleFavorite(mediaId) {
        return this.http.put(`/media/${mediaId}/favorite`);
    }
    /**
     * Get streaming URL for a media entity
     * Backend: GET /api/v1/entities/:id/stream
     */
    async getStreamUrl(entityId) {
        return this.http.get(`/entities/${entityId}/stream`);
    }
    /**
     * Get download URL for a media entity
     * Backend: GET /api/v1/entities/:id/download
     */
    async getDownloadUrl(entityId) {
        return this.http.get(`/entities/${entityId}/download`);
    }
    /**
     * Download a file by ID
     * Backend: GET /api/v1/download/file/:id
     */
    async downloadFile(fileId) {
        return this.http.downloadStream(`/download/file/${fileId}`);
    }
    /**
     * Get asset (cover art, thumbnail) for a media entity
     * Backend: GET /api/v1/assets/by-entity/:type/:id
     */
    async getEntityAsset(entityType, entityId) {
        return this.http.downloadStream(`/assets/by-entity/${entityType}/${entityId}`);
    }
    /**
     * Refresh metadata for a media entity
     * Backend: POST /api/v1/entities/:id/metadata/refresh
     */
    async refreshMetadata(entityId) {
        return this.http.post(`/entities/${entityId}/metadata/refresh`);
    }
    /**
     * Get similar media items
     * Backend: GET /api/v1/recommendations/similar/:media_id
     */
    async getSimilar(mediaId, limit = 10) {
        return this.http.get(`/recommendations/similar/${mediaId}?limit=${limit}`);
    }
    /**
     * Get personalized recommendations for a user
     * Backend: GET /api/v1/recommendations/personalized/:user_id
     */
    async getRecommendations(userId, limit = 20) {
        return this.http.get(`/recommendations/personalized/${userId}?limit=${limit}`);
    }
    /**
     * Get media entity children (e.g., seasons of a show)
     * Backend: GET /api/v1/entities/:id/children
     */
    async getChildren(entityId) {
        return this.http.get(`/entities/${entityId}/children`);
    }
    /**
     * Get files associated with a media entity
     * Backend: GET /api/v1/entities/:id/files
     */
    async getEntityFiles(entityId) {
        return this.http.get(`/entities/${entityId}/files`);
    }
    /**
     * Get entity metadata from external providers
     * Backend: GET /api/v1/entities/:id/metadata
     */
    async getEntityMetadata(entityId) {
        return this.http.get(`/entities/${entityId}/metadata`);
    }
    /**
     * Update user metadata for an entity
     * Backend: PUT /api/v1/entities/:id/user-metadata
     */
    async updateUserMetadata(entityId, metadata) {
        return this.http.put(`/entities/${entityId}/user-metadata`, metadata);
    }
}
exports.MediaService = MediaService;
//# sourceMappingURL=MediaService.js.map