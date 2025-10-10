"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.WebSocketClient = void 0;
const events_1 = require("events");
const ws_1 = __importDefault(require("ws"));
class WebSocketClient extends events_1.EventEmitter {
    constructor(url) {
        super();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.isConnecting = false;
        this.shouldReconnect = true;
        this.url = url;
    }
    connect(authToken) {
        return new Promise((resolve, reject) => {
            if (this.isConnecting || (this.ws && this.ws.readyState === ws_1.default.OPEN)) {
                resolve();
                return;
            }
            this.isConnecting = true;
            this.authToken = authToken;
            const wsUrl = authToken ? `${this.url}?token=${authToken}` : this.url;
            this.ws = new ws_1.default(wsUrl);
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
    disconnect() {
        this.shouldReconnect = false;
        this.stopHeartbeat();
        if (this.ws) {
            this.ws.close();
            this.ws = undefined;
        }
    }
    isConnected() {
        return this.ws?.readyState === ws_1.default.OPEN;
    }
    send(message) {
        if (this.isConnected()) {
            this.ws.send(JSON.stringify(message));
        }
        else {
            throw new Error('WebSocket is not connected');
        }
    }
    setAuthToken(token) {
        this.authToken = token;
    }
    handleMessage(data) {
        try {
            const message = JSON.parse(data);
            switch (message.type) {
                case 'download_progress':
                    this.emit('download:progress', message.data);
                    break;
                case 'scan_progress':
                    this.emit('scan:progress', message.data);
                    break;
                case 'pong':
                    // Heartbeat response
                    break;
                default:
                    // Emit generic message event
                    this.emit('message', message);
                    break;
            }
        }
        catch (error) {
            console.error('Failed to parse WebSocket message:', error);
        }
    }
    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            if (this.isConnected()) {
                this.send({ type: 'ping', timestamp: new Date().toISOString() });
            }
        }, 30000); // Send ping every 30 seconds
    }
    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = undefined;
        }
    }
    scheduleReconnect() {
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
exports.WebSocketClient = WebSocketClient;
//# sourceMappingURL=websocket.js.map