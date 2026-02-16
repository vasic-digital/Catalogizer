import { AuthService } from '../AuthService';
import { HttpClient } from '../../utils/http';

// Mock HttpClient
jest.mock('../../utils/http');

describe('AuthService', () => {
  let authService: AuthService;
  let mockHttp: jest.Mocked<HttpClient>;

  beforeEach(() => {
    jest.clearAllMocks();

    mockHttp = {
      get: jest.fn(),
      post: jest.fn(),
      put: jest.fn(),
      patch: jest.fn(),
      delete: jest.fn(),
      setAuthToken: jest.fn(),
      clearAuthToken: jest.fn(),
      getAuthToken: jest.fn(),
    } as any;

    authService = new AuthService(mockHttp);
  });

  describe('login', () => {
    it('sends credentials and returns login response', async () => {
      const loginResponse = {
        token: 'jwt-token',
        refresh_token: 'refresh-token',
        expires_in: 3600,
        user: { id: 1, username: 'testuser', email: 'test@test.com' },
      };
      mockHttp.post.mockResolvedValueOnce(loginResponse);

      const result = await authService.login({
        username: 'testuser',
        password: 'password123',
      });

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/login', {
        username: 'testuser',
        password: 'password123',
      });
      expect(result).toEqual(loginResponse);
    });

    it('sets auth token after successful login', async () => {
      const loginResponse = { token: 'new-jwt-token', refresh_token: 'r', expires_in: 3600, user: {} };
      mockHttp.post.mockResolvedValueOnce(loginResponse);

      await authService.login({ username: 'user', password: 'pass' });

      expect(mockHttp.setAuthToken).toHaveBeenCalledWith('new-jwt-token');
    });

    it('propagates errors on login failure', async () => {
      mockHttp.post.mockRejectedValueOnce(new Error('Invalid credentials'));

      await expect(authService.login({ username: 'wrong', password: 'wrong' }))
        .rejects.toThrow('Invalid credentials');
    });
  });

  describe('register', () => {
    it('sends registration data and returns user', async () => {
      const user = { id: 1, username: 'newuser', email: 'new@test.com' };
      mockHttp.post.mockResolvedValueOnce(user);

      const result = await authService.register({
        username: 'newuser',
        email: 'new@test.com',
        password: 'securepass',
        first_name: 'New',
        last_name: 'User',
      });

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/register', {
        username: 'newuser',
        email: 'new@test.com',
        password: 'securepass',
        first_name: 'New',
        last_name: 'User',
      });
      expect(result).toEqual(user);
    });
  });

  describe('logout', () => {
    it('calls logout endpoint and clears token', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);

      await authService.logout();

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/logout');
      expect(mockHttp.clearAuthToken).toHaveBeenCalled();
    });

    it('clears token even if logout request fails', async () => {
      mockHttp.post.mockRejectedValueOnce(new Error('Network error'));

      // logout uses try/finally (no catch), so the error propagates
      await expect(authService.logout()).rejects.toThrow('Network error');

      // But clearAuthToken should still be called via finally
      expect(mockHttp.clearAuthToken).toHaveBeenCalled();
    });
  });

  describe('getStatus', () => {
    it('returns authentication status', async () => {
      const status = { authenticated: true, user: { id: 1, username: 'testuser' } };
      mockHttp.get.mockResolvedValueOnce(status);

      const result = await authService.getStatus();

      expect(mockHttp.get).toHaveBeenCalledWith('/auth/status');
      expect(result).toEqual(status);
    });
  });

  describe('refreshToken', () => {
    it('refreshes token and updates auth header', async () => {
      const refreshResponse = { token: 'refreshed-token', expires_in: 7200 };
      mockHttp.post.mockResolvedValueOnce(refreshResponse);

      const result = await authService.refreshToken();

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/refresh');
      expect(mockHttp.setAuthToken).toHaveBeenCalledWith('refreshed-token');
      expect(result).toEqual(refreshResponse);
    });
  });

  describe('profile operations', () => {
    it('gets user profile', async () => {
      const profile = { id: 1, username: 'testuser', email: 'test@test.com' };
      mockHttp.get.mockResolvedValueOnce(profile);

      const result = await authService.getProfile();

      expect(mockHttp.get).toHaveBeenCalledWith('/auth/profile');
      expect(result).toEqual(profile);
    });

    it('updates user profile', async () => {
      const updatedUser = { id: 1, first_name: 'Updated', last_name: 'Name' };
      mockHttp.put.mockResolvedValueOnce(updatedUser);

      const result = await authService.updateProfile({ first_name: 'Updated' });

      expect(mockHttp.put).toHaveBeenCalledWith('/auth/profile', { first_name: 'Updated' });
      expect(result).toEqual(updatedUser);
    });
  });

  describe('password operations', () => {
    it('changes password', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);

      await authService.changePassword({
        current_password: 'oldpass',
        new_password: 'newpass',
      });

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/password', {
        current_password: 'oldpass',
        new_password: 'newpass',
      });
    });

    it('requests password reset', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);

      await authService.requestPasswordReset('test@test.com');

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/password/reset', { email: 'test@test.com' });
    });

    it('resets password with token', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);

      await authService.resetPassword('reset-token-123', 'newpassword');

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/password/reset/confirm', {
        token: 'reset-token-123',
        password: 'newpassword',
      });
    });
  });

  describe('email verification', () => {
    it('verifies email with token', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);

      await authService.verifyEmail('verify-token');

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/email/verify', { token: 'verify-token' });
    });

    it('resends email verification', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);

      await authService.resendEmailVerification();

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/email/verify/resend');
    });
  });

  describe('availability checks', () => {
    it('checks username availability', async () => {
      mockHttp.get.mockResolvedValueOnce({ available: true });

      const result = await authService.checkUsernameAvailability('newuser');

      expect(mockHttp.get).toHaveBeenCalledWith('/auth/username/check?username=newuser');
      expect(result).toEqual({ available: true });
    });

    it('checks email availability with URL encoding', async () => {
      mockHttp.get.mockResolvedValueOnce({ available: false });

      const result = await authService.checkEmailAvailability('user@example.com');

      expect(mockHttp.get).toHaveBeenCalledWith('/auth/email/check?email=user%40example.com');
      expect(result).toEqual({ available: false });
    });
  });

  describe('permissions', () => {
    it('gets user permissions', async () => {
      mockHttp.get.mockResolvedValueOnce(['read', 'write', 'admin:system']);

      const result = await authService.getPermissions();

      expect(mockHttp.get).toHaveBeenCalledWith('/auth/permissions');
      expect(result).toEqual(['read', 'write', 'admin:system']);
    });

    it('returns true for existing permission via hasPermission', async () => {
      mockHttp.get.mockResolvedValueOnce(['read', 'write']);

      const result = await authService.hasPermission('read');

      expect(result).toBe(true);
    });

    it('returns true for admin:system wildcard via hasPermission', async () => {
      mockHttp.get.mockResolvedValueOnce(['admin:system']);

      const result = await authService.hasPermission('any_permission');

      expect(result).toBe(true);
    });

    it('returns false for missing permission via hasPermission', async () => {
      mockHttp.get.mockResolvedValueOnce(['read']);

      const result = await authService.hasPermission('write');

      expect(result).toBe(false);
    });
  });

  describe('API keys', () => {
    it('generates an API key', async () => {
      const apiKey = { key: 'api-key-123', name: 'My Key', created_at: '2024-01-01T00:00:00Z' };
      mockHttp.post.mockResolvedValueOnce(apiKey);

      const result = await authService.generateApiKey('My Key');

      expect(mockHttp.post).toHaveBeenCalledWith('/auth/api-keys', { name: 'My Key' });
      expect(result).toEqual(apiKey);
    });

    it('lists API keys', async () => {
      const keys = [{ id: 1, name: 'Key 1', created_at: '2024-01-01T00:00:00Z' }];
      mockHttp.get.mockResolvedValueOnce(keys);

      const result = await authService.listApiKeys();

      expect(mockHttp.get).toHaveBeenCalledWith('/auth/api-keys');
      expect(result).toEqual(keys);
    });

    it('revokes an API key', async () => {
      mockHttp.delete.mockResolvedValueOnce(undefined);

      await authService.revokeApiKey(42);

      expect(mockHttp.delete).toHaveBeenCalledWith('/auth/api-keys/42');
    });
  });
});
