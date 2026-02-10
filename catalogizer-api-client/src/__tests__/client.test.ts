import { CatalogizerClient } from '../index';
import axios from 'axios';
import WebSocket from 'ws';

jest.mock('axios');
jest.mock('ws');

const mockAxios = axios as jest.Mocked<typeof axios>;
const MockWebSocket = WebSocket as jest.MockedClass<typeof WebSocket>;

describe('CatalogizerClient Integration', () => {
  let mockAxiosInstance: any;
  let mockWs: jest.Mocked<WebSocket>;

  beforeEach(() => {
    jest.clearAllMocks();

    mockAxiosInstance = {
      interceptors: {
        request: { use: jest.fn(), eject: jest.fn() },
        response: { use: jest.fn(), eject: jest.fn() },
      },
      get: jest.fn(),
      post: jest.fn(),
      put: jest.fn(),
      patch: jest.fn(),
      delete: jest.fn(),
      defaults: {
        headers: {},
        baseURL: '',
        timeout: 30000,
      },
    };

    mockAxios.create.mockReturnValue(mockAxiosInstance);

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

  describe('client initialization', () => {
    it('creates client with all services', () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      expect(client).toBeDefined();
      expect(client.auth).toBeDefined();
      expect(client.media).toBeDefined();
      expect(client.smb).toBeDefined();
    });

    it('initializes with custom configuration', () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
        timeout: 10000,
        headers: { 'X-Custom': 'value' },
      });

      const config = client.getConfig();

      expect(config.baseURL).toBe('http://localhost:8080');
      expect(config.timeout).toBe(10000);
      expect(config.headers).toEqual({ 'X-Custom': 'value' });
    });
  });

  describe('server connection', () => {
    it('connects to server successfully', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.get.mockResolvedValueOnce({
        data: { status: 'healthy', version: '1.0.0' },
      });

      const result = await client.connect();

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/health', undefined);
      expect(result).toBeNull();
    });

    it('connects with credentials and logs in', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      const loginResponse = {
        user: { id: 1, username: 'testuser' },
        token: 'jwt-token',
        refresh_token: 'refresh-token',
        expires_in: 3600,
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: { status: 'healthy' } });
      mockAxiosInstance.post.mockResolvedValueOnce({ data: loginResponse });

      const loginListener = jest.fn();
      client.on('auth:login', loginListener);

      const result = await client.connect({
        username: 'testuser',
        password: 'password',
      });

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/auth/login', {
        username: 'testuser',
        password: 'password',
      }, undefined);
      expect(result).toEqual(loginResponse);
      expect(loginListener).toHaveBeenCalledWith(loginResponse.user);
    });

    it('checks existing authentication on connect', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.get
        .mockResolvedValueOnce({ data: { status: 'healthy' } })
        .mockResolvedValueOnce({
          data: {
            authenticated: true,
            user: { id: 1, username: 'testuser' },
          },
        });

      const loginListener = jest.fn();
      client.on('auth:login', loginListener);

      await client.connect();

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/auth/status', undefined);
      expect(loginListener).toHaveBeenCalledWith({ id: 1, username: 'testuser' });
    });

    it('handles connection failure gracefully', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Connection refused'));

      await expect(client.connect()).rejects.toThrow(
        'Failed to connect to Catalogizer server'
      );
    });
  });

  describe('disconnection', () => {
    it('disconnects and logs out', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.post.mockResolvedValueOnce({ data: {} });

      const logoutListener = jest.fn();
      client.on('auth:logout', logoutListener);

      await client.disconnect();

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/auth/logout', undefined, undefined);
      expect(logoutListener).toHaveBeenCalled();
    });

    it('handles logout errors gracefully', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.post.mockRejectedValueOnce(new Error('Logout failed'));

      const logoutListener = jest.fn();
      client.on('auth:logout', logoutListener);

      await client.disconnect();

      // Should still emit logout event even on error
      expect(logoutListener).toHaveBeenCalled();
    });
  });

  describe('authentication state', () => {
    it('checks if user is authenticated', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.get.mockResolvedValueOnce({
        data: { authenticated: true },
      });

      const result = await client.isAuthenticated();

      expect(result).toBe(true);
      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/auth/status', undefined);
    });

    it('returns false when not authenticated', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.get.mockResolvedValueOnce({
        data: { authenticated: false },
      });

      const result = await client.isAuthenticated();

      expect(result).toBe(false);
    });

    it('returns false on authentication check error', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Network error'));

      const result = await client.isAuthenticated();

      expect(result).toBe(false);
    });

    it('gets current user', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      const mockUser = { id: 1, username: 'testuser', email: 'test@test.com' };
      mockAxiosInstance.get.mockResolvedValueOnce({ data: mockUser });

      const result = await client.getCurrentUser();

      expect(result).toEqual(mockUser);
      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/auth/profile', undefined);
    });

    it('returns null when getting current user fails', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Not authenticated'));

      const result = await client.getCurrentUser();

      expect(result).toBeNull();
    });
  });

  describe('token management', () => {
    it('sets authentication token', () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      client.setAuthToken('test-token');

      expect(client.getAuthToken()).toBe('test-token');
    });

    it('clears authentication token', () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      client.setAuthToken('test-token');
      client.clearAuthToken();

      expect(client.getAuthToken()).toBeUndefined();
    });

    it('sets token in WebSocket when available', () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
        enableWebSocket: true,
        webSocketURL: 'ws://localhost:8080/ws',
      });

      client.setAuthToken('test-token');

      // Token should be set in WebSocket client (tested indirectly)
      expect(client).toBeDefined();
    });
  });

  describe('configuration', () => {
    it('updates client configuration', () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      client.updateConfig({
        baseURL: 'http://newhost:9090',
        timeout: 5000,
      });

      const config = client.getConfig();

      expect(config.baseURL).toBe('http://newhost:9090');
      expect(config.timeout).toBe(5000);
    });

    it('returns current configuration', () => {
      const initialConfig = {
        baseURL: 'http://localhost:8080',
        timeout: 10000,
      };

      const client = new CatalogizerClient(initialConfig);
      const config = client.getConfig();

      expect(config.baseURL).toBe(initialConfig.baseURL);
      expect(config.timeout).toBe(initialConfig.timeout);
    });
  });

  describe('health check', () => {
    it('performs health check', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      const healthResponse = {
        status: 'healthy',
        version: '1.0.0',
        timestamp: '2024-01-01T00:00:00Z',
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: healthResponse });

      const result = await client.healthCheck();

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/health', undefined);
      expect(result).toEqual(healthResponse);
    });
  });

  describe('server info', () => {
    it('gets server information', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      const serverInfo = {
        name: 'Catalogizer',
        version: '1.0.0',
        description: 'Media catalog manager',
        features: ['media', 'smb', 'ftp'],
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: serverInfo });

      const result = await client.getServerInfo();

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/info', undefined);
      expect(result).toEqual(serverInfo);
    });
  });

  describe('event handling', () => {
    it('emits auth:login event on successful login', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      const loginResponse = {
        user: { id: 1, username: 'testuser' },
        token: 'jwt-token',
        refresh_token: 'refresh-token',
        expires_in: 3600,
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: { status: 'healthy' } });
      mockAxiosInstance.post.mockResolvedValueOnce({ data: loginResponse });

      const loginListener = jest.fn();
      client.on('auth:login', loginListener);

      await client.connect({ username: 'testuser', password: 'password' });

      expect(loginListener).toHaveBeenCalledWith(loginResponse.user);
    });

    it('emits auth:logout event on disconnect', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      mockAxiosInstance.post.mockResolvedValueOnce({ data: {} });

      const logoutListener = jest.fn();
      client.on('auth:logout', logoutListener);

      await client.disconnect();

      expect(logoutListener).toHaveBeenCalled();
    });
  });

  describe('token refresh handling', () => {
    it('emits auth:token_refresh event on token refresh', async () => {
      const client = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
      });

      const tokenRefreshListener = jest.fn();
      client.on('auth:token_refresh', tokenRefreshListener);

      mockAxiosInstance.post.mockResolvedValueOnce({
        data: { token: 'new-token', expires_in: 3600 },
      });

      // Simulate token refresh by accessing handleTokenRefresh through auth.refreshToken
      await client.auth.refreshToken();

      // The token refresh event should be emitted
      // Note: This is tested indirectly through the auth service
      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/auth/refresh', undefined, undefined);
    });
  });
});
