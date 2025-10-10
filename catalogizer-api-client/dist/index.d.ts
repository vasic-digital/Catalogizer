import { EventEmitter } from 'events';
import { AuthService } from './services/AuthService';
import { MediaService } from './services/MediaService';
import { SMBService } from './services/SMBService';
import { ClientConfig, ClientEvents, User, LoginRequest, LoginResponse } from './types';
export declare class CatalogizerClient extends EventEmitter {
    private http;
    private ws?;
    private config;
    readonly auth: AuthService;
    readonly media: MediaService;
    readonly smb: SMBService;
    constructor(config: ClientConfig);
    /**
     * Initialize WebSocket connection
     */
    private initializeWebSocket;
    /**
     * Connect to the Catalogizer server
     */
    connect(credentials?: LoginRequest): Promise<LoginResponse | null>;
    /**
     * Disconnect from the server
     */
    disconnect(): Promise<void>;
    /**
     * Check if client is connected and authenticated
     */
    isAuthenticated(): Promise<boolean>;
    /**
     * Get current user information
     */
    getCurrentUser(): Promise<User | null>;
    /**
     * Set authentication token manually
     */
    setAuthToken(token: string): void;
    /**
     * Clear authentication token
     */
    clearAuthToken(): void;
    /**
     * Get current authentication token
     */
    getAuthToken(): string | undefined;
    /**
     * Update client configuration
     */
    updateConfig(newConfig: Partial<ClientConfig>): void;
    /**
     * Get current configuration
     */
    getConfig(): ClientConfig;
    /**
     * Check server health
     */
    healthCheck(): Promise<{
        status: string;
        version: string;
        timestamp: string;
    }>;
    /**
     * Get server information
     */
    getServerInfo(): Promise<{
        name: string;
        version: string;
        description: string;
        features: string[];
    }>;
    /**
     * Handle token refresh
     */
    private handleTokenRefresh;
    /**
     * Handle authentication errors
     */
    private handleAuthenticationError;
    on<K extends keyof ClientEvents>(event: K, listener: ClientEvents[K]): this;
    on(event: string | symbol, listener: (...args: any[]) => void): this;
    off<K extends keyof ClientEvents>(event: K, listener: ClientEvents[K]): this;
    off(event: string | symbol, listener: (...args: any[]) => void): this;
    emit<K extends keyof ClientEvents>(event: K, ...args: Parameters<ClientEvents[K]>): boolean;
    emit(event: string | symbol, ...args: any[]): boolean;
}
export * from './types';
export { HttpClient } from './utils/http';
export { WebSocketClient } from './utils/websocket';
export { AuthService } from './services/AuthService';
export { MediaService } from './services/MediaService';
export { SMBService } from './services/SMBService';
export default CatalogizerClient;
//# sourceMappingURL=index.d.ts.map