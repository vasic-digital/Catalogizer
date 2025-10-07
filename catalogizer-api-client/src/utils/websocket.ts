import { EventEmitter } from 'events';
import WebSocket from 'ws';
import {
  WebSocketMessage,
  DownloadProgressMessage,
  ScanProgressMessage,
  ClientEvents,
} from '../types';

export class WebSocketClient extends EventEmitter {
  private ws?: WebSocket;
  private url: string;
  private authToken?: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private heartbeatInterval?: NodeJS.Timeout;
  private isConnecting = false;
  private shouldReconnect = true;

  constructor(url: string) {
    super();
    this.url = url;
  }

  public connect(authToken?: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
        resolve();
        return;
      }

      this.isConnecting = true;
      this.authToken = authToken;

      const wsUrl = authToken ? `${this.url}?token=${authToken}` : this.url;
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.startHeartbeat();
        this.emit('connection:open');
        resolve();
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(typeof event.data === 'string' ? event.data : event.data.toString());
      };

      this.ws.onclose = (event) => {
        this.isConnecting = false;
        this.stopHeartbeat();
        this.emit('connection:close');

        if (this.shouldReconnect && !event.wasClean) {
          this.scheduleReconnect();
        }
      };

      this.ws.onerror = (error) => {
        this.isConnecting = false;
        this.emit('connection:error', error);
        reject(error);
      };
    });
  }

  public disconnect(): void {
    this.shouldReconnect = false;
    this.stopHeartbeat();

    if (this.ws) {
      this.ws.close();
      this.ws = undefined;
    }
  }

  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  public send(message: any): void {
    if (this.isConnected()) {
      this.ws!.send(JSON.stringify(message));
    } else {
      throw new Error('WebSocket is not connected');
    }
  }

  public setAuthToken(token: string): void {
    this.authToken = token;
  }

  private handleMessage(data: string): void {
    try {
      const message: WebSocketMessage = JSON.parse(data);

      switch (message.type) {
        case 'download_progress':
          this.emit('download:progress', (message as DownloadProgressMessage).data);
          break;

        case 'scan_progress':
          this.emit('scan:progress', (message as ScanProgressMessage).data);
          break;

        case 'pong':
          // Heartbeat response
          break;

        default:
          // Emit generic message event
          this.emit('message', message);
          break;
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  private startHeartbeat(): void {
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected()) {
        this.send({ type: 'ping', timestamp: new Date().toISOString() });
      }
    }, 30000); // Send ping every 30 seconds
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = undefined;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;

    setTimeout(() => {
      if (this.shouldReconnect) {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        this.connect(this.authToken).catch((error) => {
          console.error('Reconnection failed:', error);
        });
      }
    }, delay);
  }

  // Typed event methods
  public on<K extends keyof ClientEvents>(event: K, listener: ClientEvents[K]): this;
  public on(event: string | symbol, listener: (...args: any[]) => void): this;
  public on(event: any, listener: any): this {
    return super.on(event, listener);
  }

  public off<K extends keyof ClientEvents>(event: K, listener: ClientEvents[K]): this;
  public off(event: string | symbol, listener: (...args: any[]) => void): this;
  public off(event: any, listener: any): this {
    return super.off(event, listener);
  }

  public emit<K extends keyof ClientEvents>(event: K, ...args: Parameters<ClientEvents[K]>): boolean;
  public emit(event: string | symbol, ...args: any[]): boolean;
  public emit(event: any, ...args: any[]): boolean {
    return super.emit(event, ...args);
  }
}