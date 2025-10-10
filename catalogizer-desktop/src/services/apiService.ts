import { invoke } from '@tauri-apps/api/core';
import {
  MediaItem,
  MediaSearchRequest,
  MediaSearchResponse,
  MediaStats,
  LoginRequest,
  LoginResponse,
  AuthStatus,
  User,
  SMBConfig,
  SMBStatus,
  PlaybackProgress,
} from '../types';

class ApiService {
  private async makeRequest<T>(
    endpoint: string,
    options: {
      method?: string;
      body?: any;
      headers?: Record<string, string>;
    } = {}
  ): Promise<T> {
    const config = await invoke<any>('get_config');

    if (!config.server_url) {
      throw new Error('Server URL not configured');
    }

    const url = `${config.server_url}/api${endpoint}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    // Add auth token if available
    if (config.auth_token) {
      headers['Authorization'] = `Bearer ${config.auth_token}`;
    }

    const response = await invoke<string>('make_http_request', {
      url,
      method: options.method || 'GET',
      headers,
      body: options.body ? JSON.stringify(options.body) : undefined,
    });

    try {
      return JSON.parse(response);
    } catch (error) {
      throw new Error('Invalid response format');
    }
  }

  // Auth endpoints
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    return this.makeRequest<LoginResponse>('/auth/login', {
      method: 'POST',
      body: credentials,
    });
  }

  async logout(): Promise<void> {
    return this.makeRequest<void>('/auth/logout', {
      method: 'POST',
    });
  }

  async getAuthStatus(): Promise<AuthStatus> {
    return this.makeRequest<AuthStatus>('/auth/status');
  }

  async getProfile(): Promise<User> {
    return this.makeRequest<User>('/auth/profile');
  }

  // Media endpoints
  async searchMedia(request: MediaSearchRequest = {}): Promise<MediaSearchResponse> {
    const params = new URLSearchParams();

    Object.entries(request).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        params.append(key, value.toString());
      }
    });

    const query = params.toString();
    const endpoint = query ? `/media/search?${query}` : '/media/search';

    return this.makeRequest<MediaSearchResponse>(endpoint);
  }

  async getMediaById(id: number): Promise<MediaItem> {
    return this.makeRequest<MediaItem>(`/media/${id}`);
  }

  async getMediaStats(): Promise<MediaStats> {
    return this.makeRequest<MediaStats>('/media/stats');
  }

  async updateWatchProgress(mediaId: number, progress: PlaybackProgress): Promise<void> {
    return this.makeRequest<void>(`/media/${mediaId}/progress`, {
      method: 'PUT',
      body: progress,
    });
  }

  async toggleFavorite(mediaId: number): Promise<void> {
    return this.makeRequest<void>(`/media/${mediaId}/favorite`, {
      method: 'POST',
    });
  }

  async getMediaUrl(mediaId: number): Promise<{ url: string }> {
    return this.makeRequest<{ url: string }>(`/media/${mediaId}/stream`);
  }

  async downloadMedia(mediaId: number): Promise<{ job_id: number }> {
    return this.makeRequest<{ job_id: number }>(`/media/${mediaId}/download`, {
      method: 'POST',
    });
  }

  // SMB endpoints
  async getSMBConfigs(): Promise<SMBConfig[]> {
    return this.makeRequest<SMBConfig[]>('/smb/configs');
  }

  async createSMBConfig(config: Omit<SMBConfig, 'id' | 'created_at' | 'updated_at'>): Promise<SMBConfig> {
    return this.makeRequest<SMBConfig>('/smb/configs', {
      method: 'POST',
      body: config,
    });
  }

  async updateSMBConfig(id: number, config: Partial<SMBConfig>): Promise<SMBConfig> {
    return this.makeRequest<SMBConfig>(`/smb/configs/${id}`, {
      method: 'PUT',
      body: config,
    });
  }

  async deleteSMBConfig(id: number): Promise<void> {
    return this.makeRequest<void>(`/smb/configs/${id}`, {
      method: 'DELETE',
    });
  }

  async getSMBStatus(configId?: number): Promise<SMBStatus[]> {
    const endpoint = configId ? `/smb/status/${configId}` : '/smb/status';
    return this.makeRequest<SMBStatus[]>(endpoint);
  }

  async connectSMB(configId: number): Promise<void> {
    return this.makeRequest<void>(`/smb/connect/${configId}`, {
      method: 'POST',
    });
  }

  async disconnectSMB(configId: number): Promise<void> {
    return this.makeRequest<void>(`/smb/disconnect/${configId}`, {
      method: 'POST',
    });
  }

  async scanSMB(configId: number): Promise<{ job_id: number }> {
    return this.makeRequest<{ job_id: number }>(`/smb/scan/${configId}`, {
      method: 'POST',
    });
  }

  // System endpoints
  async getSystemInfo(): Promise<{
    version: string;
    platform: string;
    arch: string;
  }> {
    const [version, platform, arch] = await Promise.all([
      invoke<string>('get_app_version'),
      invoke<string>('get_platform'),
      invoke<string>('get_arch'),
    ]);

    return { version, platform, arch };
  }

  async healthCheck(): Promise<{ status: string; timestamp: string }> {
    return this.makeRequest<{ status: string; timestamp: string }>('/health');
  }
}

export const apiService = new ApiService();