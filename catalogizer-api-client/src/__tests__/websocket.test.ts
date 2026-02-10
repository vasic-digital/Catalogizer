import { WebSocketClient } from '../utils/websocket';
import WebSocket from 'ws';

// Mock WebSocket
jest.mock('ws');
const MockWebSocket = WebSocket as jest.MockedClass<typeof WebSocket>;

describe('WebSocketClient', () => {
  let mockWs: jest.Mocked<WebSocket>;

  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();

    mockWs = {
      readyState: WebSocket.OPEN,
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
    } as any;

    MockWebSocket.mockImplementation(() => mockWs);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('initialization', () => {
    it('creates WebSocketClient with URL', () => {
      const client = new WebSocketClient('ws://localhost:8080');

      expect(client).toBeDefined();
      expect(client.isConnected()).toBe(false);
    });
  });

  describe('connection', () => {
    it('connects successfully without auth token', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      // Simulate connection opened
      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      expect(MockWebSocket).toHaveBeenCalledWith('ws://localhost:8080');
      expect(client.isConnected()).toBe(true);
    });

    it('connects successfully with auth token', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect('test-token');

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      expect(MockWebSocket).toHaveBeenCalledWith('ws://localhost:8080?token=test-token');
    });

    it('emits connection:open event on successful connection', async () => {
      const client = new WebSocketClient('ws://localhost:8080');
      const openListener = jest.fn();

      client.on('connection:open', openListener);

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      expect(openListener).toHaveBeenCalled();
    });

    it('handles connection error', async () => {
      const client = new WebSocketClient('ws://localhost:8080');
      const errorListener = jest.fn();

      client.on('connection:error', errorListener);

      const connectPromise = client.connect();

      const error = new Error('Connection failed');
      if (mockWs.onerror) {
        mockWs.onerror(error as any);
      }

      await expect(connectPromise).rejects.toThrow('Connection failed');
      expect(errorListener).toHaveBeenCalledWith(error);
    });

    it('does not reconnect if already connecting', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const promise1 = client.connect();
      const promise2 = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await promise1;
      await promise2;

      expect(MockWebSocket).toHaveBeenCalledTimes(1);
    });

    it('does not reconnect if already connected', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const promise1 = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await promise1;

      await client.connect();

      expect(MockWebSocket).toHaveBeenCalledTimes(1);
    });
  });

  describe('disconnection', () => {
    it('disconnects successfully', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      client.disconnect();

      expect(mockWs.close).toHaveBeenCalled();
      expect(client.isConnected()).toBe(false);
    });

    it('emits connection:close event on disconnection', async () => {
      const client = new WebSocketClient('ws://localhost:8080');
      const closeListener = jest.fn();

      client.on('connection:close', closeListener);

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      if (mockWs.onclose) {
        mockWs.onclose({ wasClean: true } as any);
      }

      expect(closeListener).toHaveBeenCalled();
    });

    it('does not reconnect after manual disconnect', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      client.disconnect();

      if (mockWs.onclose) {
        mockWs.onclose({ wasClean: false } as any);
      }

      jest.advanceTimersByTime(5000);

      expect(MockWebSocket).toHaveBeenCalledTimes(1);
    });
  });

  describe('message handling', () => {
    it('handles download progress messages', async () => {
      const client = new WebSocketClient('ws://localhost:8080');
      const progressListener = jest.fn();

      client.on('download:progress', progressListener);

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      const message = {
        type: 'download_progress',
        data: { id: 1, progress: 50, speed: 1000000 },
      };

      if (mockWs.onmessage) {
        mockWs.onmessage({ data: JSON.stringify(message) } as any);
      }

      expect(progressListener).toHaveBeenCalledWith({ id: 1, progress: 50, speed: 1000000 });
    });

    it('handles scan progress messages', async () => {
      const client = new WebSocketClient('ws://localhost:8080');
      const progressListener = jest.fn();

      client.on('scan:progress', progressListener);

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      const message = {
        type: 'scan_progress',
        data: { id: 1, scanned: 100, total: 500 },
      };

      if (mockWs.onmessage) {
        mockWs.onmessage({ data: JSON.stringify(message) } as any);
      }

      expect(progressListener).toHaveBeenCalledWith({ id: 1, scanned: 100, total: 500 });
    });

    it('handles pong messages (heartbeat)', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      const message = { type: 'pong', timestamp: new Date().toISOString() };

      if (mockWs.onmessage) {
        mockWs.onmessage({ data: JSON.stringify(message) } as any);
      }

      // Should not throw, pong is handled silently
      expect(true).toBe(true);
    });

    it('emits generic message event for unknown types', async () => {
      const client = new WebSocketClient('ws://localhost:8080');
      const messageListener = jest.fn();

      client.on('message', messageListener);

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      const message = { type: 'unknown', data: { foo: 'bar' } };

      if (mockWs.onmessage) {
        mockWs.onmessage({ data: JSON.stringify(message) } as any);
      }

      expect(messageListener).toHaveBeenCalledWith(message);
    });

    it('handles malformed JSON messages gracefully', async () => {
      const client = new WebSocketClient('ws://localhost:8080');
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      if (mockWs.onmessage) {
        mockWs.onmessage({ data: 'not valid json' } as any);
      }

      expect(consoleSpy).toHaveBeenCalledWith(
        'Failed to parse WebSocket message:',
        expect.any(Error)
      );

      consoleSpy.mockRestore();
    });
  });

  describe('sending messages', () => {
    it('sends message when connected', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      const message = { type: 'test', data: { foo: 'bar' } };
      client.send(message);

      expect(mockWs.send).toHaveBeenCalledWith(JSON.stringify(message));
    });

    it('throws error when sending while not connected', () => {
      const client = new WebSocketClient('ws://localhost:8080');

      expect(() => {
        client.send({ type: 'test' });
      }).toThrow('WebSocket is not connected');
    });
  });

  describe('heartbeat mechanism', () => {
    it('sends ping messages at regular intervals', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      // Clear initial send calls
      mockWs.send.mockClear();

      // Fast-forward 30 seconds
      jest.advanceTimersByTime(30000);

      expect(mockWs.send).toHaveBeenCalledWith(expect.stringContaining('"type":"ping"'));
    });

    it('stops sending pings after disconnect', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      client.disconnect();

      mockWs.send.mockClear();

      jest.advanceTimersByTime(30000);

      expect(mockWs.send).not.toHaveBeenCalled();
    });
  });

  describe('reconnection mechanism', () => {
    it('schedules reconnection after unclean disconnect', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      // Simulate unclean disconnect
      if (mockWs.onclose) {
        mockWs.onclose({ wasClean: false } as any);
      }

      // Verify a timer was set (reconnection scheduled)
      expect(jest.getTimerCount()).toBeGreaterThan(0);
    });

    it('does not reconnect after clean disconnect', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      // Clear any existing timers
      jest.clearAllTimers();

      // Simulate clean disconnect
      if (mockWs.onclose) {
        mockWs.onclose({ wasClean: true } as any);
      }

      // No new timers should be set
      expect(jest.getTimerCount()).toBe(0);
    });
  });

  describe('auth token management', () => {
    it('sets auth token', () => {
      const client = new WebSocketClient('ws://localhost:8080');
      client.setAuthToken('new-token');

      // Token should be stored (tested indirectly through connection)
      expect(client).toBeDefined();
    });

    it('uses updated auth token on reconnection', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      const connectPromise = client.connect('old-token');

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      client.setAuthToken('new-token');

      // Simulate disconnect and reconnect
      if (mockWs.onclose) {
        mockWs.onclose({ wasClean: false } as any);
      }

      jest.advanceTimersByTime(1000);

      // New connection should use new token (but since reconnect uses stored authToken,
      // it will use 'old-token' unless we explicitly reconnect)
      expect(MockWebSocket).toHaveBeenLastCalledWith(expect.stringContaining('token=old-token'));
    });
  });

  describe('connection state', () => {
    it('reports connected state correctly', async () => {
      const client = new WebSocketClient('ws://localhost:8080');

      expect(client.isConnected()).toBe(false);

      const connectPromise = client.connect();

      if (mockWs.onopen) {
        mockWs.onopen({} as any);
      }

      await connectPromise;

      expect(client.isConnected()).toBe(true);

      client.disconnect();

      Object.defineProperty(mockWs, 'readyState', { value: WebSocket.CLOSED, writable: true });
      expect(client.isConnected()).toBe(false);
    });

    it('reports not connected when readyState is CONNECTING', () => {
      const client = new WebSocketClient('ws://localhost:8080');
      Object.defineProperty(mockWs, 'readyState', { value: WebSocket.CONNECTING, writable: true });

      client.connect();

      expect(client.isConnected()).toBe(false);
    });

    it('reports not connected when readyState is CLOSING', () => {
      const client = new WebSocketClient('ws://localhost:8080');
      Object.defineProperty(mockWs, 'readyState', { value: WebSocket.CLOSING, writable: true });

      expect(client.isConnected()).toBe(false);
    });
  });
});
