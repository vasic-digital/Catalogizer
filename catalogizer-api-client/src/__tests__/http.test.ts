import { HttpClient } from '../utils/http';
import { CatalogizerError, AuthenticationError, NetworkError, ValidationError } from '../types';
import axios, { AxiosInstance } from 'axios';

jest.mock('axios');
const mockAxios = axios as jest.Mocked<typeof axios>;

describe('HttpClient', () => {
  let mockAxiosInstance: jest.Mocked<AxiosInstance>;

  beforeEach(() => {
    jest.clearAllMocks();

    const requestUseMock = jest.fn();
    const responseUseMock = jest.fn();

    mockAxiosInstance = {
      interceptors: {
        request: { use: requestUseMock as any, eject: jest.fn() },
        response: { use: responseUseMock as any, eject: jest.fn() },
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
    } as any;

    mockAxios.create.mockReturnValue(mockAxiosInstance as any);
  });

  describe('initialization', () => {
    it('creates axios instance with default config', () => {
      new HttpClient({ baseURL: 'http://localhost:8080' });

      expect(mockAxios.create).toHaveBeenCalledWith({
        baseURL: 'http://localhost:8080',
        timeout: 30000,
        headers: {
          'Content-Type': 'application/json',
        },
      });
    });

    it('creates axios instance with custom timeout', () => {
      new HttpClient({
        baseURL: 'http://localhost:8080',
        timeout: 5000,
      });

      expect(mockAxios.create).toHaveBeenCalledWith(expect.objectContaining({
        timeout: 5000,
      }));
    });

    it('creates axios instance with custom headers', () => {
      new HttpClient({
        baseURL: 'http://localhost:8080',
        headers: { 'X-Custom-Header': 'value' },
      });

      expect(mockAxios.create).toHaveBeenCalledWith(expect.objectContaining({
        headers: {
          'Content-Type': 'application/json',
          'X-Custom-Header': 'value',
        },
      }));
    });

    it('sets up request and response interceptors', () => {
      new HttpClient({ baseURL: 'http://localhost:8080' });

      expect(mockAxiosInstance.interceptors.request.use).toHaveBeenCalled();
      expect(mockAxiosInstance.interceptors.response.use).toHaveBeenCalled();
    });
  });

  describe('authentication token management', () => {
    it('sets authentication token', () => {
      const client = new HttpClient({ baseURL: 'http://localhost:8080' });
      client.setAuthToken('test-token');

      expect(client.getAuthToken()).toBe('test-token');
    });

    it('clears authentication token', () => {
      const client = new HttpClient({ baseURL: 'http://localhost:8080' });
      client.setAuthToken('test-token');
      client.clearAuthToken();

      expect(client.getAuthToken()).toBeUndefined();
    });

    it('adds auth token to request headers', async () => {
      const client = new HttpClient({ baseURL: 'http://localhost:8080' });
      client.setAuthToken('test-token');

      // Get the request interceptor function
      const requestUseMock = mockAxiosInstance.interceptors.request.use as jest.Mock;
      const requestInterceptor = requestUseMock.mock.calls[0][0];

      const config = { headers: {} };
      const result = requestInterceptor(config);

      expect(result.headers.Authorization).toBe('Bearer test-token');
    });

    it('does not add auth header when token is not set', () => {
      const client = new HttpClient({ baseURL: 'http://localhost:8080' });

      const requestUseMock = mockAxiosInstance.interceptors.request.use as jest.Mock;
      const requestInterceptor = requestUseMock.mock.calls[0][0];
      const config = { headers: {} };
      const result = requestInterceptor(config);

      expect(result.headers.Authorization).toBeUndefined();
    });
  });

  describe('HTTP methods', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({ baseURL: 'http://localhost:8080' });
    });

    it('performs GET request', async () => {
      const mockData = { id: 1, name: 'test' };
      mockAxiosInstance.get.mockResolvedValueOnce({ data: mockData });

      const result = await client.get('/test');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/test', undefined);
      expect(result).toEqual(mockData);
    });

    it('performs GET request with config', async () => {
      const mockData = { items: [] };
      const config = { params: { page: 1 } };
      mockAxiosInstance.get.mockResolvedValueOnce({ data: mockData });

      const result = await client.get('/test', config);

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/test', config);
      expect(result).toEqual(mockData);
    });

    it('performs POST request', async () => {
      const mockData = { success: true };
      const payload = { name: 'test' };
      mockAxiosInstance.post.mockResolvedValueOnce({ data: mockData });

      const result = await client.post('/test', payload);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith('/test', payload, undefined);
      expect(result).toEqual(mockData);
    });

    it('performs PUT request', async () => {
      const mockData = { id: 1, updated: true };
      const payload = { name: 'updated' };
      mockAxiosInstance.put.mockResolvedValueOnce({ data: mockData });

      const result = await client.put('/test/1', payload);

      expect(mockAxiosInstance.put).toHaveBeenCalledWith('/test/1', payload, undefined);
      expect(result).toEqual(mockData);
    });

    it('performs PATCH request', async () => {
      const mockData = { id: 1, patched: true };
      const payload = { name: 'patched' };
      mockAxiosInstance.patch.mockResolvedValueOnce({ data: mockData });

      const result = await client.patch('/test/1', payload);

      expect(mockAxiosInstance.patch).toHaveBeenCalledWith('/test/1', payload, undefined);
      expect(result).toEqual(mockData);
    });

    it('performs DELETE request', async () => {
      mockAxiosInstance.delete.mockResolvedValueOnce({ data: {} });

      await client.delete('/test/1');

      expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/test/1', undefined);
    });
  });

  describe('response data extraction', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({ baseURL: 'http://localhost:8080' });
    });

    it('extracts data from response with data field', async () => {
      const mockResponse = {
        data: {
          success: true,
          data: { id: 1, name: 'test' },
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      const result = await client.get('/test');

      expect(result).toEqual({ id: 1, name: 'test' });
    });

    it('returns entire response when no data field', async () => {
      const mockResponse = {
        data: { id: 1, name: 'test' },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      const result = await client.get('/test');

      expect(result).toEqual({ id: 1, name: 'test' });
    });

    it('throws error when response has success: false', async () => {
      const mockResponse = {
        data: {
          success: false,
          error: 'Something went wrong',
          status: 400,
        },
      };
      mockAxiosInstance.get
        .mockResolvedValueOnce(mockResponse)
        .mockResolvedValueOnce(mockResponse);

      await expect(client.get('/test')).rejects.toThrow('Something went wrong');
    });
  });

  describe('error handling', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({ baseURL: 'http://localhost:8080' });
    });

    it('handles network errors', async () => {
      const networkError = new Error('Network Error');
      mockAxiosInstance.get.mockRejectedValueOnce(networkError);

      // Get the response interceptor error handler
      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(networkError);
      } catch (error) {
        expect(error).toBeInstanceOf(NetworkError);
        expect((error as NetworkError).message).toBe('Network connection failed');
      }
    });

    it('handles 400 validation errors', async () => {
      const error = {
        response: {
          status: 400,
          data: { message: 'Validation failed' },
        },
      };
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(ValidationError);
        expect((err as ValidationError).message).toBe('Validation failed');
      }
    });

    it('handles 401 authentication errors', async () => {
      const error = {
        response: {
          status: 401,
          data: { message: 'Unauthorized' },
        },
        config: {},
      };
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(AuthenticationError);
      }
    });

    it('handles 403 forbidden errors', async () => {
      const error = {
        response: {
          status: 403,
          data: { message: 'Forbidden' },
        },
      };
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(CatalogizerError);
        expect((err as CatalogizerError).code).toBe('FORBIDDEN');
      }
    });

    it('handles 404 not found errors', async () => {
      const error = {
        response: {
          status: 404,
          data: { message: 'Not found' },
        },
      };
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(CatalogizerError);
        expect((err as CatalogizerError).code).toBe('NOT_FOUND');
      }
    });

    it('handles 500 server errors', async () => {
      const error = {
        response: {
          status: 500,
          data: { message: 'Internal server error' },
        },
      };
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(CatalogizerError);
        expect((err as CatalogizerError).code).toBe('SERVER_ERROR');
      }
    });

    it('extracts error message from data.error field', async () => {
      const error = {
        response: {
          status: 400,
          data: { error: 'Custom error message' },
        },
      };
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect((err as CatalogizerError).message).toBe('Custom error message');
      }
    });

    it('uses default error message when none provided', async () => {
      const error = {
        response: {
          status: 400,
          data: {},
        },
        message: 'Request failed',
      };
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      const responseUseMock = mockAxiosInstance.interceptors.response.use as jest.Mock;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect((err as CatalogizerError).message).toBe('Request failed');
      }
    });
  });

  describe('token refresh on 401', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({ baseURL: 'http://localhost:8080' });
    });

    it('has onTokenRefresh callback property', () => {
      expect(client.onTokenRefresh).toBeUndefined();

      const callback = jest.fn();
      client.onTokenRefresh = callback;

      expect(client.onTokenRefresh).toBe(callback);
    });

    it('has onAuthenticationError callback property', () => {
      expect(client.onAuthenticationError).toBeUndefined();

      const callback = jest.fn();
      client.onAuthenticationError = callback;

      expect(client.onAuthenticationError).toBe(callback);
    });
  });

  describe('retry mechanism', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({
        baseURL: 'http://localhost:8080',
        retryAttempts: 3,
        retryDelay: 10,
      });
    });

    it('does not retry authentication errors', async () => {
      const operation = jest.fn()
        .mockRejectedValue(new AuthenticationError('Unauthorized'));

      await expect(client.withRetry(operation, 3, 10)).rejects.toThrow(AuthenticationError);
      expect(operation).toHaveBeenCalledTimes(1);
    });

    it('does not retry validation errors', async () => {
      const operation = jest.fn()
        .mockRejectedValue(new ValidationError('Invalid data'));

      await expect(client.withRetry(operation, 3, 10)).rejects.toThrow(ValidationError);
      expect(operation).toHaveBeenCalledTimes(1);
    });

    it('succeeds on first attempt when operation succeeds', async () => {
      const operation = jest.fn().mockResolvedValue('Success');

      const result = await client.withRetry(operation, 3, 10);

      expect(operation).toHaveBeenCalledTimes(1);
      expect(result).toBe('Success');
    });
  });

  describe('stream operations', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({ baseURL: 'http://localhost:8080' });
    });

    it('downloads stream as arraybuffer', async () => {
      const arrayBuffer = new ArrayBuffer(100);
      mockAxiosInstance.get.mockResolvedValueOnce({ data: arrayBuffer });

      const result = await client.downloadStream('/file');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/file', {
        responseType: 'arraybuffer',
      });
      expect(result).toBe(arrayBuffer);
    });

    it('downloads stream with custom config', async () => {
      const arrayBuffer = new ArrayBuffer(100);
      const config = { headers: { 'X-Custom': 'value' } };
      mockAxiosInstance.get.mockResolvedValueOnce({ data: arrayBuffer });

      await client.downloadStream('/file', config);

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/file', expect.objectContaining({
        responseType: 'arraybuffer',
        headers: { 'X-Custom': 'value' },
      }));
    });
  });

  describe('config updates', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({ baseURL: 'http://localhost:8080' });
    });

    it('updates base URL', () => {
      client.updateConfig({ baseURL: 'http://newhost:9090' });

      expect(mockAxiosInstance.defaults.baseURL).toBe('http://newhost:9090');
    });

    it('updates timeout', () => {
      client.updateConfig({ timeout: 5000 });

      expect(mockAxiosInstance.defaults.timeout).toBe(5000);
    });

    it('updates headers', () => {
      client.updateConfig({
        headers: { 'X-New-Header': 'value' },
      });

      expect(mockAxiosInstance.defaults.headers['X-New-Header']).toBe('value');
    });

    it('updates multiple config properties', () => {
      client.updateConfig({
        baseURL: 'http://newhost:9090',
        timeout: 10000,
        headers: { 'X-Custom': 'value' },
      });

      expect(mockAxiosInstance.defaults.baseURL).toBe('http://newhost:9090');
      expect(mockAxiosInstance.defaults.timeout).toBe(10000);
      expect(mockAxiosInstance.defaults.headers['X-Custom']).toBe('value');
    });
  });
});
