"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.MediaService = void 0;
class MediaService {
    constructor(http) {
        this.http = http;
    }
    /**
     * Search for media items
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
     */
    async getById(id) {
        return this.http.get(`/media/${id}`);
    }
    /**
     * Get media statistics
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
     */
    async getTrending(limit = 20) {
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
     */
    async getFavorites(limit = 50) {
        return this.http.get(`/media/favorites?limit=${limit}`);
    }
    /**
     * Toggle favorite status for a media item
     */
    async toggleFavorite(mediaId) {
        return this.http.post(`/media/${mediaId}/favorite`);
    }
    /**
     * Get continue watching items (items with partial progress)
     */
    async getContinueWatching(limit = 20) {
        return this.http.get(`/media/continue-watching?limit=${limit}`);
    }
    /**
     * Update playback progress for a media item
     */
    async updateProgress(mediaId, progress) {
        return this.http.put(`/media/${mediaId}/progress`, progress);
    }
    /**
     * Get playback progress for a media item
     */
    async getProgress(mediaId) {
        try {
            return await this.http.get(`/media/${mediaId}/progress`);
        }
        catch (error) {
            // Return null if no progress found
            return null;
        }
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
     * Get streaming URL for a media item
     */
    async getStreamUrl(mediaId) {
        return this.http.get(`/media/${mediaId}/stream`);
    }
    /**
     * Get download URL for a media item
     */
    async getDownloadUrl(mediaId) {
        return this.http.get(`/media/${mediaId}/download`);
    }
    /**
     * Queue media for download
     */
    async queueDownload(mediaId) {
        return this.http.post(`/media/${mediaId}/download`);
    }
    /**
     * Get download jobs for the current user
     */
    async getDownloadJobs() {
        return this.http.get('/media/downloads');
    }
    /**
     * Get specific download job
     */
    async getDownloadJob(jobId) {
        return this.http.get(`/media/downloads/${jobId}`);
    }
    /**
     * Cancel a download job
     */
    async cancelDownload(jobId) {
        return this.http.post(`/media/downloads/${jobId}/cancel`);
    }
    /**
     * Pause a download job
     */
    async pauseDownload(jobId) {
        return this.http.post(`/media/downloads/${jobId}/pause`);
    }
    /**
     * Resume a download job
     */
    async resumeDownload(jobId) {
        return this.http.post(`/media/downloads/${jobId}/resume`);
    }
    /**
     * Get media thumbnail
     */
    async getThumbnail(mediaId, size) {
        const params = size ? `?size=${size}` : '';
        return this.http.downloadStream(`/media/${mediaId}/thumbnail${params}`);
    }
    /**
     * Get media poster
     */
    async getPoster(mediaId, size) {
        const params = size ? `?size=${size}` : '';
        return this.http.downloadStream(`/media/${mediaId}/poster${params}`);
    }
    /**
     * Get media backdrop
     */
    async getBackdrop(mediaId, size) {
        const params = size ? `?size=${size}` : '';
        return this.http.downloadStream(`/media/${mediaId}/backdrop${params}`);
    }
    /**
     * Update media metadata
     */
    async updateMetadata(mediaId, metadata) {
        return this.http.put(`/media/${mediaId}`, metadata);
    }
    /**
     * Delete a media item
     */
    async delete(mediaId, deleteFiles = false) {
        const params = deleteFiles ? '?delete_files=true' : '';
        return this.http.delete(`/media/${mediaId}${params}`);
    }
    /**
     * Refresh metadata for a media item
     */
    async refreshMetadata(mediaId) {
        return this.http.post(`/media/${mediaId}/refresh`);
    }
    /**
     * Get similar media items
     */
    async getSimilar(mediaId, limit = 10) {
        return this.http.get(`/media/${mediaId}/similar?limit=${limit}`);
    }
    /**
     * Get media recommendations for the user
     */
    async getRecommendations(limit = 20) {
        return this.http.get(`/media/recommendations?limit=${limit}`);
    }
    /**
     * Rate a media item
     */
    async rate(mediaId, rating) {
        return this.http.post(`/media/${mediaId}/rate`, { rating });
    }
    /**
     * Get user's rating for a media item
     */
    async getRating(mediaId) {
        try {
            return await this.http.get(`/media/${mediaId}/rating`);
        }
        catch (error) {
            return null;
        }
    }
}
exports.MediaService = MediaService;
//# sourceMappingURL=MediaService.js.map