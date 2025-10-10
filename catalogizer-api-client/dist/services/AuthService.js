"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.AuthService = void 0;
class AuthService {
    constructor(http) {
        this.http = http;
    }
    /**
     * Authenticate user with username and password
     */
    async login(credentials) {
        const response = await this.http.post('/auth/login', credentials);
        // Set the auth token for future requests
        this.http.setAuthToken(response.token);
        return response;
    }
    /**
     * Register a new user account
     */
    async register(userData) {
        return this.http.post('/auth/register', userData);
    }
    /**
     * Logout the current user
     */
    async logout() {
        try {
            await this.http.post('/auth/logout');
        }
        finally {
            // Always clear the token, even if the request fails
            this.http.clearAuthToken();
        }
    }
    /**
     * Get current authentication status
     */
    async getStatus() {
        return this.http.get('/auth/status');
    }
    /**
     * Refresh the authentication token
     */
    async refreshToken() {
        const response = await this.http.post('/auth/refresh');
        // Update the auth token
        this.http.setAuthToken(response.token);
        return response;
    }
    /**
     * Get current user profile
     */
    async getProfile() {
        return this.http.get('/auth/profile');
    }
    /**
     * Update user profile
     */
    async updateProfile(updates) {
        return this.http.put('/auth/profile', updates);
    }
    /**
     * Change user password
     */
    async changePassword(passwordData) {
        return this.http.post('/auth/password', passwordData);
    }
    /**
     * Request password reset
     */
    async requestPasswordReset(email) {
        return this.http.post('/auth/password/reset', { email });
    }
    /**
     * Reset password with token
     */
    async resetPassword(token, newPassword) {
        return this.http.post('/auth/password/reset/confirm', {
            token,
            password: newPassword,
        });
    }
    /**
     * Verify email address
     */
    async verifyEmail(token) {
        return this.http.post('/auth/email/verify', { token });
    }
    /**
     * Resend email verification
     */
    async resendEmailVerification() {
        return this.http.post('/auth/email/verify/resend');
    }
    /**
     * Check if a username is available
     */
    async checkUsernameAvailability(username) {
        return this.http.get(`/auth/username/check?username=${encodeURIComponent(username)}`);
    }
    /**
     * Check if an email is available
     */
    async checkEmailAvailability(email) {
        return this.http.get(`/auth/email/check?email=${encodeURIComponent(email)}`);
    }
    /**
     * Get user permissions
     */
    async getPermissions() {
        return this.http.get('/auth/permissions');
    }
    /**
     * Check if user has specific permission
     */
    async hasPermission(permission) {
        const permissions = await this.getPermissions();
        return permissions.includes(permission) || permissions.includes('admin:system');
    }
    /**
     * Generate API key for the current user
     */
    async generateApiKey(name) {
        return this.http.post('/auth/api-keys', { name });
    }
    /**
     * List user's API keys
     */
    async listApiKeys() {
        return this.http.get('/auth/api-keys');
    }
    /**
     * Revoke an API key
     */
    async revokeApiKey(keyId) {
        return this.http.delete(`/auth/api-keys/${keyId}`);
    }
}
exports.AuthService = AuthService;
//# sourceMappingURL=AuthService.js.map