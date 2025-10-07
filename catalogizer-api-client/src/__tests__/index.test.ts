import { CatalogizerClient, CatalogizerError } from '../index';

// Mock axios
jest.mock('axios');
import axios from 'axios';
const mockAxios = axios as jest.Mocked<typeof axios>;

// Mock axios.create to return a mock instance with interceptors
const mockAxiosInstance = {
  interceptors: {
    request: { use: jest.fn() },
    response: { use: jest.fn() },
  },
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
  patch: jest.fn(),
};

mockAxios.create.mockReturnValue(mockAxiosInstance as any);

describe('CatalogizerClient', () => {
  let client: CatalogizerClient;

  beforeEach(() => {
    jest.clearAllMocks();
    client = new CatalogizerClient({
      baseURL: 'http://localhost:8080'
    });
  });

  describe('initialization', () => {
    it('creates client with custom config', () => {
      expect(client).toBeDefined();
    });
  });

  describe('media service', () => {
    describe('search', () => {
      it('searches media successfully', async () => {
        const mockResponse = {
          data: {
            items: [
              {
                id: 1,
                title: 'Test Movie',
                directory_path: '/media/movies/',
                media_type: 'movie'
              }
            ],
            total: 1,
            limit: 10,
            offset: 0
          }
        };

        // Mock the axios response structure
        const axiosResponse = { data: mockResponse.data };

        mockAxiosInstance.get.mockResolvedValueOnce(axiosResponse);

        const result = await client.media.search({
          query: 'test movie',
          media_type: 'movie',
          limit: 10
        });

        expect(mockAxiosInstance.get).toHaveBeenCalledWith('/media/search?query=test+movie&media_type=movie&limit=10', undefined);
        expect(result).toEqual(mockResponse.data);
      });

      it('handles search errors', async () => {
        mockAxiosInstance.get.mockRejectedValueOnce(new Error('Network error'));

        await expect(client.media.search({ query: 'test' })).rejects.toThrow('Network error');
      });
    });

    describe('getById', () => {
      it('gets media by id', async () => {
        const mockResponse = {
          data: {
            id: 1,
            title: 'Test Movie',
            directory_path: '/media/movies/',
            media_type: 'movie',
            file_size: 1000000
          }
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.get.mockResolvedValueOnce(axiosResponse);

        const result = await client.media.getById(1);

        expect(mockAxiosInstance.get).toHaveBeenCalledWith('/media/1', undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });

    describe('getStats', () => {
      it('gets media statistics', async () => {
        const mockResponse = {
          data: {
            total_items: 100,
            total_size: 1000000000,
            by_type: { movie: 50, tv: 30, music: 20 }
          }
        };

        mockAxiosInstance.get.mockResolvedValueOnce(mockResponse);

        const result = await client.media.getStats();

        expect(mockAxiosInstance.get).toHaveBeenCalledWith('/media/stats', undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });
  });

  describe('auth service', () => {
    describe('login', () => {
      it('logs in successfully', async () => {
        const mockResponse = {
          data: {
            user: { id: 1, username: 'testuser' },
            token: 'jwt-token',
            refresh_token: 'refresh-token',
            expires_in: 3600
          }
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.post.mockResolvedValueOnce(axiosResponse);

        const result = await client.auth.login({
          username: 'testuser',
          password: 'password'
        });

        expect(mockAxiosInstance.post).toHaveBeenCalledWith('/auth/login', {
          username: 'testuser',
          password: 'password'
        }, undefined);

        expect(result).toEqual(mockResponse.data);
      });

      it('handles login failure', async () => {
        const catalogizerError = new CatalogizerError('Invalid credentials', 401, 'AUTH_ERROR');
        mockAxiosInstance.post.mockRejectedValueOnce(catalogizerError);

        await expect(client.auth.login({
          username: 'wrong',
          password: 'wrong'
        })).rejects.toThrow(CatalogizerError);
      });
    });

    describe('refreshToken', () => {
      it('refreshes token', async () => {
        const mockResponse = {
          data: {
            token: 'new-jwt-token',
            expires_in: 3600
          }
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.post.mockResolvedValueOnce(axiosResponse);

        const result = await client.auth.refreshToken();

        expect(mockAxiosInstance.post).toHaveBeenCalledWith('/auth/refresh', undefined, undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });

    describe('logout', () => {
      it('logs out successfully', async () => {
        const axiosResponse = { data: {} };
        mockAxiosInstance.post.mockResolvedValueOnce(axiosResponse);

        await expect(client.auth.logout()).resolves.toBeUndefined();

        expect(mockAxiosInstance.post).toHaveBeenCalledWith('/auth/logout', undefined, undefined);
      });
    });

    describe('getStatus', () => {
      it('gets authentication status', async () => {
        const mockResponse = {
          data: {
            authenticated: true,
            user: { id: 1, username: 'testuser' }
          }
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.get.mockResolvedValueOnce(axiosResponse);

        const result = await client.auth.getStatus();

        expect(mockAxiosInstance.get).toHaveBeenCalledWith('/auth/status', undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });
  });

  describe('SMB service', () => {
    describe('getConfigs', () => {
      it('gets SMB configurations', async () => {
        const mockResponse = {
          data: [
            {
              id: 1,
              name: 'Media Server',
              host: '192.168.1.100',
              share_name: 'media',
              username: 'user'
            }
          ]
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.get.mockResolvedValueOnce(axiosResponse);

        const result = await client.smb.getConfigs();

        expect(mockAxiosInstance.get).toHaveBeenCalledWith('/smb/configs', undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });

    describe('getStatus', () => {
      it('gets SMB status', async () => {
        const mockResponse = {
          data: [
            {
              id: 1,
              config_id: 1,
              status: 'connected',
              last_connected: '2024-01-01T00:00:00Z'
            }
          ]
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.get.mockResolvedValueOnce(axiosResponse);

        const result = await client.smb.getStatus();

        expect(mockAxiosInstance.get).toHaveBeenCalledWith('/smb/status', undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });

    describe('connect', () => {
      it('connects to SMB share', async () => {
        const mockResponse = {
          data: { success: true, message: 'Connected successfully' }
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.post.mockResolvedValueOnce(axiosResponse);

        const result = await client.smb.connect(1);

        expect(mockAxiosInstance.post).toHaveBeenCalledWith('/smb/connect/1', undefined, undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });

    describe('createConfig', () => {
      it('creates SMB configuration', async () => {
        const config = {
          name: 'Test Share',
          host: '192.168.1.100',
          port: 445,
          share_name: 'media',
          username: 'user',
          password: 'pass',
          mount_point: '/mnt/media'
        };

        const mockResponse = {
          data: { id: 1, ...config }
        };

        const axiosResponse = { data: mockResponse.data };
        mockAxiosInstance.post.mockResolvedValueOnce(axiosResponse);

        const result = await client.smb.createConfig(config);

        expect(mockAxiosInstance.post).toHaveBeenCalledWith('/smb/configs', config, undefined);
        expect(result).toEqual(mockResponse.data);
      });
    });
  });

  describe('error handling', () => {
    it('handles network errors', async () => {
      mockAxiosInstance.get.mockRejectedValueOnce(new Error('Network Error'));

      await expect(client.media.search({ query: 'test' })).rejects.toThrow('Network Error');
    });

    it('handles HTTP errors', async () => {
      const catalogizerError = new CatalogizerError('Not found', 404, 'NOT_FOUND');
      mockAxiosInstance.get.mockRejectedValueOnce(catalogizerError);

      await expect(client.media.getById(999)).rejects.toThrow(CatalogizerError);
    });

    it('handles timeout errors', async () => {
      const networkError = new Error('Network connection failed');
      (networkError as any).code = 'ECONNABORTED';
      mockAxiosInstance.get.mockRejectedValueOnce(networkError);

      await expect(client.media.search({ query: 'test' })).rejects.toThrow('Network connection failed');
    });
  });

  describe('configuration', () => {
    it('uses custom timeout', () => {
      const customClient = new CatalogizerClient({
        baseURL: 'http://localhost:8080',
        timeout: 5000
      });

      expect(customClient).toBeDefined();
    });

    it('handles invalid base URL', () => {
      expect(() => {
        new CatalogizerClient({
          baseURL: 'invalid-url',
        });
      }).not.toThrow();
    });
  });
});