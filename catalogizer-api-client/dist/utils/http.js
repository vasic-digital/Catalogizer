"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.HttpClient = void 0;
const axios_1 = __importDefault(require("axios"));
const types_1 = require("../types");
class HttpClient {
    constructor(config) {
        this.config = config;
        this.client = axios_1.default.create({
            baseURL: config.baseURL,
            timeout: config.timeout || 30000,
            headers: {
                'Content-Type': 'application/json',
                ...config.headers,
            },
        });
        this.setupInterceptors();
    }
    setupInterceptors() {
        // Request interceptor to add auth token
        this.client.interceptors.request.use((config) => {
            if (this.authToken) {
                config.headers.Authorization = `Bearer ${this.authToken}`;
            }
            return config;
        }, (error) => Promise.reject(error));
        // Response interceptor to handle errors
        this.client.interceptors.response.use((response) => response, async (error) => {
            const originalRequest = error.config;
            if (error.response?.status === 401 && !originalRequest._retry) {
                originalRequest._retry = true;
                // Try to refresh token if possible
                if (this.onTokenRefresh) {
                    try {
                        const newToken = await this.onTokenRefresh();
                        if (newToken) {
                            this.setAuthToken(newToken);
                            originalRequest.headers.Authorization = `Bearer ${newToken}`;
                            return this.client(originalRequest);
                        }
                    }
                    catch (refreshError) {
                        // Token refresh failed, user needs to login again
                        if (this.onAuthenticationError) {
                            this.onAuthenticationError();
                        }
                    }
                }
            }
            return Promise.reject(this.handleError(error));
        });
    }
    handleError(error) {
        if (!error.response) {
            // Network error
            return new types_1.NetworkError('Network connection failed');
        }
        const { status, data } = error.response;
        const message = data?.message || data?.error || error.message || 'Request failed';
        switch (status) {
            case 400:
                return new types_1.ValidationError(message);
            case 401:
                return new types_1.AuthenticationError(message);
            case 403:
                return new types_1.CatalogizerError('Access forbidden', status, 'FORBIDDEN');
            case 404:
                return new types_1.CatalogizerError('Resource not found', status, 'NOT_FOUND');
            case 500:
                return new types_1.CatalogizerError('Internal server error', status, 'SERVER_ERROR');
            default:
                return new types_1.CatalogizerError(message, status);
        }
    }
    setAuthToken(token) {
        this.authToken = token;
    }
    clearAuthToken() {
        this.authToken = undefined;
    }
    getAuthToken() {
        return this.authToken;
    }
    // HTTP methods
    async get(url, config) {
        const response = await this.client.get(url, config);
        return this.extractData(response);
    }
    async post(url, data, config) {
        const response = await this.client.post(url, data, config);
        return this.extractData(response);
    }
    async put(url, data, config) {
        const response = await this.client.put(url, data, config);
        return this.extractData(response);
    }
    async patch(url, data, config) {
        const response = await this.client.patch(url, data, config);
        return this.extractData(response);
    }
    async delete(url, config) {
        const response = await this.client.delete(url, config);
        return this.extractData(response);
    }
    extractData(response) {
        const { data: responseData } = response;
        // If the response has a success field and it's false, throw an error
        if (responseData.success === false) {
            throw new types_1.CatalogizerError(responseData.error || responseData.message || 'Request failed', responseData.status);
        }
        // Return the data field if it exists, otherwise return the entire response data
        return responseData.data !== undefined ? responseData.data : responseData;
    }
    // Stream methods for file downloads/uploads
    async downloadStream(url, config) {
        const response = await this.client.get(url, {
            ...config,
            responseType: 'arraybuffer',
        });
        return response.data;
    }
    async uploadFile(url, file, config) {
        const formData = new FormData();
        formData.append('file', file);
        return this.client.post(url, formData, {
            ...config,
            headers: {
                ...config?.headers,
                'Content-Type': 'multipart/form-data',
            },
        });
    }
    // Retry mechanism
    async withRetry(operation, maxAttempts = this.config.retryAttempts || 3, delay = this.config.retryDelay || 1000) {
        let lastError;
        for (let attempt = 1; attempt <= maxAttempts; attempt++) {
            try {
                return await operation();
            }
            catch (error) {
                lastError = error;
                // Don't retry on authentication or validation errors
                if (error instanceof types_1.AuthenticationError || error instanceof types_1.ValidationError) {
                    throw error;
                }
                if (attempt === maxAttempts) {
                    break;
                }
                // Wait before retrying
                await new Promise(resolve => setTimeout(resolve, delay * attempt));
            }
        }
        throw lastError;
    }
    // Update configuration
    updateConfig(newConfig) {
        this.config = { ...this.config, ...newConfig };
        if (newConfig.baseURL) {
            this.client.defaults.baseURL = newConfig.baseURL;
        }
        if (newConfig.timeout) {
            this.client.defaults.timeout = newConfig.timeout;
        }
        if (newConfig.headers) {
            this.client.defaults.headers = {
                ...this.client.defaults.headers,
                ...newConfig.headers,
            };
        }
    }
}
exports.HttpClient = HttpClient;
//# sourceMappingURL=http.js.map