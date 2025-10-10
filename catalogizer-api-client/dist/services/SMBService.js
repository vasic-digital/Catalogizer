"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SMBService = void 0;
class SMBService {
    constructor(http) {
        this.http = http;
    }
    /**
     * Get all SMB configurations
     */
    async getConfigs() {
        return this.http.get('/smb/configs');
    }
    /**
     * Get a specific SMB configuration
     */
    async getConfig(id) {
        return this.http.get(`/smb/configs/${id}`);
    }
    /**
     * Create a new SMB configuration
     */
    async createConfig(config) {
        return this.http.post('/smb/configs', config);
    }
    /**
     * Update an existing SMB configuration
     */
    async updateConfig(id, updates) {
        return this.http.put(`/smb/configs/${id}`, updates);
    }
    /**
     * Delete an SMB configuration
     */
    async deleteConfig(id) {
        return this.http.delete(`/smb/configs/${id}`);
    }
    /**
     * Test connection to an SMB share
     */
    async testConnection(config) {
        return this.http.post('/smb/test', config);
    }
    /**
     * Test existing SMB configuration
     */
    async testExistingConfig(id) {
        return this.http.post(`/smb/configs/${id}/test`);
    }
    /**
     * Get status of all SMB connections
     */
    async getStatus() {
        return this.http.get('/smb/status');
    }
    /**
     * Get status of a specific SMB connection
     */
    async getConfigStatus(id) {
        return this.http.get(`/smb/status/${id}`);
    }
    /**
     * Connect to an SMB share
     */
    async connect(id) {
        return this.http.post(`/smb/connect/${id}`);
    }
    /**
     * Disconnect from an SMB share
     */
    async disconnect(id) {
        return this.http.post(`/smb/disconnect/${id}`);
    }
    /**
     * Reconnect to an SMB share
     */
    async reconnect(id) {
        // Disconnect first, then connect
        await this.disconnect(id);
        return this.connect(id);
    }
    /**
     * Scan an SMB share for media files
     */
    async scan(id, options) {
        return this.http.post(`/smb/scan/${id}`, options || {});
    }
    /**
     * Get scan job status
     */
    async getScanStatus(jobId) {
        return this.http.get(`/smb/scan-jobs/${jobId}`);
    }
    /**
     * Cancel a scan job
     */
    async cancelScan(jobId) {
        return this.http.post(`/smb/scan-jobs/${jobId}/cancel`);
    }
    /**
     * Get list of scan jobs
     */
    async getScanJobs(configId) {
        const params = configId ? `?config_id=${configId}` : '';
        return this.http.get(`/smb/scan-jobs${params}`);
    }
    /**
     * Browse directories in an SMB share
     */
    async browse(id, path = '') {
        const params = path ? `?path=${encodeURIComponent(path)}` : '';
        return this.http.get(`/smb/configs/${id}/browse${params}`);
    }
    /**
     * Enable/disable an SMB configuration
     */
    async toggleConfig(id, isActive) {
        return this.http.patch(`/smb/configs/${id}`, { is_active: isActive });
    }
    /**
     * Get SMB share information
     */
    async getShareInfo(id) {
        return this.http.get(`/smb/configs/${id}/info`);
    }
    /**
     * Refresh connection to all active SMB shares
     */
    async refreshAllConnections() {
        return this.http.post('/smb/refresh-all');
    }
    /**
     * Get SMB connection logs
     */
    async getLogs(id, limit = 100) {
        const params = new URLSearchParams();
        if (id)
            params.append('config_id', id.toString());
        params.append('limit', limit.toString());
        const query = params.toString();
        return this.http.get(`/smb/logs?${query}`);
    }
}
exports.SMBService = SMBService;
//# sourceMappingURL=SMBService.js.map