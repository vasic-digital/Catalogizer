import { EventEmitter } from 'events';
import { ClientEvents } from '../types';
export declare class WebSocketClient extends EventEmitter {
    private ws?;
    private url;
    private authToken?;
    private reconnectAttempts;
    private maxReconnectAttempts;
    private reconnectDelay;
    private heartbeatInterval?;
    private isConnecting;
    private shouldReconnect;
    constructor(url: string);
    connect(authToken?: string): Promise<void>;
    disconnect(): void;
    isConnected(): boolean;
    send(message: any): void;
    setAuthToken(token: string): void;
    private handleMessage;
    private startHeartbeat;
    private stopHeartbeat;
    private scheduleReconnect;
    on<K extends keyof ClientEvents>(event: K, listener: ClientEvents[K]): this;
    on(event: string | symbol, listener: (...args: any[]) => void): this;
    off<K extends keyof ClientEvents>(event: K, listener: ClientEvents[K]): this;
    off(event: string | symbol, listener: (...args: any[]) => void): this;
    emit<K extends keyof ClientEvents>(event: K, ...args: Parameters<ClientEvents[K]>): boolean;
    emit(event: string | symbol, ...args: any[]): boolean;
}
//# sourceMappingURL=websocket.d.ts.map