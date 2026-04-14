import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { HttpClient } from '../utils/http';
import { CatalogizerError, AuthenticationError, NetworkError, ValidationError } from '../types';
import axios, { AxiosInstance } from 'axios';

vi.mock('axios');
const mockAxios = axios as unknown as { create: Mock };

describe('HttpClient', () => {
  let mockAxiosInstance: any;

  beforeEach(() => {
    vi.clearAllMocks();

    const requestUseMock = vi.fn();
    const responseUseMock = vi.fn();

    mockAxiosInstance = {
      interceptors: {
        request: { use: requestUseMock as any, eject: vi.fn() },
        response: { use: responseUseMock as any, eject: vi.fn() },
      },
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
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
      const requestUseMock = mockAxiosInstance.interceptors.request.use;
      const requestInterceptor = requestUseMock.mock.calls[0][0];

      const config = { headers: {} };
      const result = requestInterceptor(config);

      expect(result.headers.Authorization).toBe('Bearer test-token');
    });

    it('does not add auth header when token is not set', () => {
      const client = new HttpClient({ baseURL: 'http://localhost:8080' });

      const requestUseMock = mockAxiosInstance.interceptors.request.use;
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

    it('throws error using message field when error field is absent', async () => {
      const mockResponse = {
        data: {
          success: false,
          message: 'Fallback message',
          status: 422,
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      await expect(client.get('/test')).rejects.toThrow('Fallback message');
    });

    it('throws generic error when success is false with no error or message', async () => {
      const mockResponse = {
        data: {
          success: false,
          status: 500,
        },
      };
      mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

      await expect(client.get('/test')).rejects.toThrow('Request failed');
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
      const responseUseMock = mockAxiosInstance.interceptors.response.use;
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

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
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

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
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

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
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

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
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

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
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

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
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

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect((err as CatalogizerError).message).toBe('Request failed');
      }
    });

    it('handles unknown status codes with generic CatalogizerError', async () => {
      const error = {
        response: {
          status: 418,
          data: { message: "I'm a teapot" },
        },
      };

      const responseUseMock = mockAxiosInstance.interceptors.response.use;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(CatalogizerError);
        expect((err as CatalogizerError).message).toBe("I'm a teapot");
        expect((err as CatalogizerError).status).toBe(418);
        expect((err as CatalogizerError).code).toBeUndefined();
      }
    });

    it('handles request interceptor error rejection', async () => {
      const requestUseMock = mockAxiosInstance.interceptors.request.use;
      const errorHandler = requestUseMock.mock.calls[0][1];

      const error = new Error('Request setup failed');
      const result = errorHandler(error);

      await expect(result).rejects.toThrow('Request setup failed');
    });
  });

  describe('token refresh on 401', () => {
    let client: HttpClient;

    beforeEach(() => {
      client = new HttpClient({ baseURL: 'http://localhost:8080' });
    });

    it('has onTokenRefresh callback property', () => {
      expect(client.onTokenRefresh).toBeUndefined();

      const callback = vi.fn();
      client.onTokenRefresh = callback;

      expect(client.onTokenRefresh).toBe(callback);
    });

    it('has onAuthenticationError callback property', () => {
      expect(client.onAuthenticationError).toBeUndefined();

      const callback = vi.fn();
      client.onAuthenticationError = callback;

      expect(client.onAuthenticationError).toBe(callback);
    });

    it('retries request with new token when token refresh succeeds on 401', async () => {
      // Create an axios instance that is also callable (axios instances are callable)
      const callableInstance = Object.assign(
        vi.fn().mockResolvedValue({ data: { retried: true } }),
        {
          interceptors: {
            request: { use: vi.fn(), eject: vi.fn() },
            response: { use: vi.fn(), eject: vi.fn() },
          },
          get: vi.fn(),
          post: vi.fn(),
          put: vi.fn(),
          patch: vi.fn(),
          delete: vi.fn(),
          defaults: { headers: {}, baseURL: '', timeout: 30000 },
        }
      );
      mockAxios.create.mockReturnValue(callableInstance as any);

      const httpClient = new HttpClient({ baseURL: 'http://localhost:8080' });
      httpClient.onTokenRefresh = vi.fn().mockResolvedValue('refreshed-token');

      const responseUseMock = callableInstance.interceptors.response.use;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      const originalRequest = { headers: {} as any, _retry: false };
      const error = {
        response: { status: 401, data: { message: 'Unauthorized' } },
        config: originalRequest,
      };

      const result = await responseInterceptor(error);

      expect(httpClient.onTokenRefresh).toHaveBeenCalled();
      expect(originalRequest._retry).toBe(true);
      expect(originalRequest.headers.Authorization).toBe('Bearer refreshed-token');
      expect(callableInstance).toHaveBeenCalledWith(originalRequest);
      expect(result).toEqual({ data: { retried: true } });
    });

    it('calls onAuthenticationError when token refresh fails', async () => {
      const callableInstance = Object.assign(
        vi.fn(),
        {
          ...mockAxiosInstance,
          interceptors: {
            request: { use: vi.fn(), eject: vi.fn() },
            response: { use: vi.fn(), eject: vi.fn() },
          },
        }
      );
      mockAxios.create.mockReturnValue(callableInstance as any);

      const authErrorCallback = vi.fn();
      const client4 = new HttpClient({ baseURL: 'http://localhost:8080' });
      client4.onTokenRefresh = vi.fn().mockRejectedValue(new Error('Refresh failed'));
      client4.onAuthenticationError = authErrorCallback;

      const responseUseMock = callableInstance.interceptors.response.use;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      const originalRequest = { headers: {}, _retry: false };
      const error = {
        response: { status: 401, data: { message: 'Unauthorized' } },
        config: originalRequest,
      };

      try {
        await responseInterceptor(error);
      } catch (err) {
        // Expected to throw
      }

      expect(authErrorCallback).toHaveBeenCalled();
    });

    it('calls onAuthenticationError when token refresh returns null', async () => {
      const callableInstance = Object.assign(
        vi.fn(),
        {
          ...mockAxiosInstance,
          interceptors: {
            request: { use: vi.fn(), eject: vi.fn() },
            response: { use: vi.fn(), eject: vi.fn() },
          },
        }
      );
      mockAxios.create.mockReturnValue(callableInstance as any);

      const client5 = new HttpClient({ baseURL: 'http://localhost:8080' });
      client5.onTokenRefresh = vi.fn().mockResolvedValue(null);

      const responseUseMock = callableInstance.interceptors.response.use;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      const originalRequest = { headers: {}, _retry: false };
      const error = {
        response: { status: 401, data: { message: 'Unauthorized' } },
        config: originalRequest,
      };

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(AuthenticationError);
      }
    });

    it('does not retry 401 when request was already retried', async () => {
      const responseUseMock = mockAxiosInstance.interceptors.response.use;
      const responseInterceptor = responseUseMock.mock.calls[0][1];

      const originalRequest = { headers: {}, _retry: true };
      const error = {
        response: { status: 401, data: { message: 'Unauthorized' } },
        config: originalRequest,
      };

      try {
        await responseInterceptor(error);
      } catch (err) {
        expect(err).toBeInstanceOf(AuthenticationError);
      }
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
      const operation = vi.fn()
        .mockRejectedValue(new AuthenticationError('Unauthorized'));

      await expect(client.withRetry(operation, 3, 10)).rejects.toThrow(AuthenticationError);
      expect(operation).toHaveBeenCalledTimes(1);
    });

    it('does not retry validation errors', async () => {
      const operation = vi.fn()
        .mockRejectedValue(new ValidationError('Invalid data'));

      await expect(client.withRetry(operation, 3, 10)).rejects.toThrow(ValidationError);
      expect(operation).toHaveBeenCalledTimes(1);
    });

    it('succeeds on first attempt when operation succeeds', async () => {
      const operation = vi.fn().mockResolvedValue('Success');

      const result = await client.withRetry(operation, 3, 10);

      expect(operation).toHaveBeenCalledTimes(1);
      expect(result).toBe('Success');
    });

    it('retries and succeeds after transient failures', async () => {
      const operation = vi.fn()
        .mockRejectedValueOnce(new NetworkError('Connection reset'))
        .mockRejectedValueOnce(new NetworkError('Connection reset'))
        .mockResolvedValueOnce('Eventually succeeded');

      const result = await client.withRetry(operation, 3, 10);

      expect(operation).toHaveBeenCalledTimes(3);
      expect(result).toBe('Eventually succeeded');
    });

    it('throws last error after exhausting all retry attempts', async () => {
      const operation = vi.fn()
        .mockRejectedValueOnce(new NetworkError('Fail 1'))
        .mockRejectedValueOnce(new NetworkError('Fail 2'))
        .mockRejectedValueOnce(new NetworkError('Fail 3'));

      await expect(client.withRetry(operation, 3, 10)).rejects.toThrow('Fail 3');
      expect(operation).toHaveBeenCalledTimes(3);
    });

    it('uses config defaults when maxAttempts and delay are not provided', async () => {
      const operation = vi.fn()
        .mockRejectedValueOnce(new NetworkError('fail'))
        .mockResolvedValueOnce('ok');

      const result = await client.withRetry(operation);

      expect(operation).toHaveBeenCalledTimes(2);
      expect(result).toBe('ok');
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

    it('uploads a file with multipart/form-data', async () => {
      const mockResponse = { data: { id: 1, filename: 'test.jpg' } };
      mockAxiosInstance.post.mockResolvedValueOnce(mockResponse);

      const file = Buffer.from('fake-file-content');
      await client.uploadFile('/upload', file);

      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/upload',
        expect.any(FormData),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'multipart/form-data',
          }),
        })
      );
    });

    it('uploads a file with custom config headers', async () => {
      const mockResponse = { data: { id: 2, filename: 'doc.pdf' } };
      mockAxiosInstance.post.mockResolvedValueOnce(mockResponse);

      const file = Buffer.from('pdf-content');
      await client.uploadFile('/upload', file, {
        headers: { 'X-Upload-Token': 'abc123' },
      });

      expect(mockAxiosInstance.post).toHaveBeenCalledWith(
        '/upload',
        expect.any(FormData),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'multipart/form-data',
            'X-Upload-Token': 'abc123',
          }),
        })
      );
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
