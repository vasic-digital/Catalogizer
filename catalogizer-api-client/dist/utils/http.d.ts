import { AxiosRequestConfig } from 'axios';
import { ClientConfig } from '../types';
export declare class HttpClient {
    private client;
    private config;
    private authToken?;
    constructor(config: ClientConfig);
    private setupInterceptors;
    private handleError;
    setAuthToken(token: string): void;
    clearAuthToken(): void;
    getAuthToken(): string | undefined;
    onTokenRefresh?: () => Promise<string | null>;
    onAuthenticationError?: () => void;
    get<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
    post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
    put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
    patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
    delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
    private extractData;
    downloadStream(url: string, config?: AxiosRequestConfig): Promise<ArrayBuffer>;
    uploadFile(url: string, file: File | Buffer, config?: AxiosRequestConfig): Promise<any>;
    withRetry<T>(operation: () => Promise<T>, maxAttempts?: number, delay?: number): Promise<T>;
    updateConfig(newConfig: Partial<ClientConfig>): void;
}
//# sourceMappingURL=http.d.ts.map