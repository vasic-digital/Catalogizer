"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __exportStar = (this && this.__exportStar) || function(m, exports) {
    for (var p in m) if (p !== "default" && !Object.prototype.hasOwnProperty.call(exports, p)) __createBinding(exports, m, p);
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.SMBService = exports.MediaService = exports.AuthService = exports.WebSocketClient = exports.HttpClient = exports.CatalogizerClient = void 0;
const events_1 = require("events");
const http_1 = require("./utils/http");
const websocket_1 = require("./utils/websocket");
const AuthService_1 = require("./services/AuthService");
const MediaService_1 = require("./services/MediaService");
const SMBService_1 = require("./services/SMBService");
class CatalogizerClient extends events_1.EventEmitter {
    constructor(config) {
        super();
        this.config = config;
        this.http = new http_1.HttpClient(config);
        // Initialize services
        this.auth = new AuthService_1.AuthService(this.http);
        this.media = new MediaService_1.MediaService(this.http);
        this.smb = new SMBService_1.SMBService(this.http);
        // Set up HTTP client callbacks
        this.http.onTokenRefresh = this.handleTokenRefresh.bind(this);
        this.http.onAuthenticationError = this.handleAuthenticationError.bind(this);
        // Initialize WebSocket if enabled
        if (config.enableWebSocket) {
            this.initializeWebSocket();
        }
    }
    /**
     * Initialize WebSocket connection
     */
    initializeWebSocket() {
        if (!this.config.webSocketURL) {
            console.warn('WebSocket enabled but no WebSocket URL provided');
            return;
        }
        this.ws = new websocket_1.WebSocketClient(this.config.webSocketURL);
        // Proxy WebSocket events
        this.ws.on('connection:open', () => {
            this.emit('connection:open');
        });
        this.ws.on('connection:close', () => {
            this.emit('connection:close');
        });
        this.ws.on('connection:error', (error) => {
            this.emit('connection:error', error);
        });
        this.ws.on('download:progress', (progress) => {
            this.emit('download:progress', progress);
        });
        this.ws.on('scan:progress', (progress) => {
            this.emit('scan:progress', progress);
        });
    }
    /**
     * Connect to the Catalogizer server
     */
    async connect(credentials) {
        try {
            // Test server connection
            await this.http.get('/health');
            // If credentials provided, login
            if (credentials) {
                const loginResponse = await this.auth.login(credentials);
                this.emit('auth:login', loginResponse.user);
                // Connect WebSocket with auth token
                if (this.ws) {
                    await this.ws.connect(loginResponse.token);
                }
                return loginResponse;
            }
            // If no credentials, try to get current auth status
            try {
                const status = await this.auth.getStatus();
                if (status.authenticated && status.user) {
                    this.emit('auth:login', status.user);
                    // Connect WebSocket with existing token
                    if (this.ws) {
                        const token = this.http.getAuthToken();
                        if (token) {
                            await this.ws.connect(token);
                        }
                    }
                }
            }
            catch (error) {
                // Not authenticated, that's okay
            }
            return null;
        }
        catch (error) {
            throw new Error(`Failed to connect to Catalogizer server: ${error}`);
        }
    }
    /**
     * Disconnect from the server
     */
    async disconnect() {
        try {
            await this.auth.logout();
        }
        catch (error) {
            // Ignore logout errors
        }
        if (this.ws) {
            this.ws.disconnect();
        }
        this.emit('auth:logout');
    }
    /**
     * Check if client is connected and authenticated
     */
    async isAuthenticated() {
        try {
            const status = await this.auth.getStatus();
            return status.authenticated;
        }
        catch (error) {
            return false;
        }
    }
    /**
     * Get current user information
     */
    async getCurrentUser() {
        try {
            return await this.auth.getProfile();
        }
        catch (error) {
            return null;
        }
    }
    /**
     * Set authentication token manually
     */
    setAuthToken(token) {
        this.http.setAuthToken(token);
        if (this.ws) {
            this.ws.setAuthToken(token);
        }
    }
    /**
     * Clear authentication token
     */
    clearAuthToken() {
        this.http.clearAuthToken();
    }
    /**
     * Get current authentication token
     */
    getAuthToken() {
        return this.http.getAuthToken();
    }
    /**
     * Update client configuration
     */
    updateConfig(newConfig) {
        this.config = { ...this.config, ...newConfig };
        this.http.updateConfig(newConfig);
        // Reinitialize WebSocket if configuration changed
        if (newConfig.enableWebSocket !== undefined || newConfig.webSocketURL) {
            if (this.ws) {
                this.ws.disconnect();
                this.ws = undefined;
            }
            if (this.config.enableWebSocket) {
                this.initializeWebSocket();
            }
        }
    }
    /**
     * Get current configuration
     */
    getConfig() {
        return { ...this.config };
    }
    /**
     * Check server health
     */
    async healthCheck() {
        return this.http.get('/health');
    }
    /**
     * Get server information
     */
    async getServerInfo() {
        return this.http.get('/info');
    }
    /**
     * Handle token refresh
     */
    async handleTokenRefresh() {
        try {
            const response = await this.auth.refreshToken();
            this.emit('auth:token_refresh', response.token);
            return response.token;
        }
        catch (error) {
            return null;
        }
    }
    /**
     * Handle authentication errors
     */
    handleAuthenticationError() {
        this.emit('auth:logout');
    }
    on(event, listener) {
        return super.on(event, listener);
    }
    off(event, listener) {
        return super.off(event, listener);
    }
    emit(event, ...args) {
        return super.emit(event, ...args);
    }
}
exports.CatalogizerClient = CatalogizerClient;
// Export types and classes
__exportStar(require("./types"), exports);
var http_2 = require("./utils/http");
Object.defineProperty(exports, "HttpClient", { enumerable: true, get: function () { return http_2.HttpClient; } });
var websocket_2 = require("./utils/websocket");
Object.defineProperty(exports, "WebSocketClient", { enumerable: true, get: function () { return websocket_2.WebSocketClient; } });
var AuthService_2 = require("./services/AuthService");
Object.defineProperty(exports, "AuthService", { enumerable: true, get: function () { return AuthService_2.AuthService; } });
var MediaService_2 = require("./services/MediaService");
Object.defineProperty(exports, "MediaService", { enumerable: true, get: function () { return MediaService_2.MediaService; } });
var SMBService_2 = require("./services/SMBService");
Object.defineProperty(exports, "SMBService", { enumerable: true, get: function () { return SMBService_2.SMBService; } });
// Default export
exports.default = CatalogizerClient;
//# sourceMappingURL=index.js.map