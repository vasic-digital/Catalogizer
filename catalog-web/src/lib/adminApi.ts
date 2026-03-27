import type { SystemInfo, User, StorageInfo, Backup } from '@/types/admin';
import { api } from './api';

const mockSystemInfo: SystemInfo = {
  version: '1.0.0',
  uptime: 86400,
  cpuUsage: 45,
  memoryUsage: 62,
  diskUsage: {
    total: 1073741824000, // 1TB
    used: 536870912000, // 500GB
    free: 536870912000 // 500GB
  },
  activeConnections: 12,
  totalRequests: 15420
};

const mockUsers: User[] = [
  {
    id: '1',
    username: 'admin',
    email: 'admin@catalogizer.local',
    role: 'admin',
    status: 'active',
    lastLogin: '2023-12-09T10:30:00Z',
    createdAt: '2023-01-01T00:00:00Z'
  },
  {
    id: '2',
    username: 'user1',
    email: 'user1@example.com',
    role: 'user',
    status: 'active',
    lastLogin: '2023-12-08T15:45:00Z',
    createdAt: '2023-02-15T10:30:00Z'
  }
];

const mockStorageInfo: StorageInfo[] = [
  {
    path: '/media/movies',
    totalSpace: 1073741824000,
    usedSpace: 751619276800,
    availableSpace: 322122547200,
    mediaCount: 1250,
    lastScan: '2023-12-09T09:00:00Z'
  },
  {
    path: '/media/tv',
    totalSpace: 536870912000,
    usedSpace: 268435456000,
    availableSpace: 268435456000,
    mediaCount: 850,
    lastScan: '2023-12-09T08:30:00Z'
  }
];

const mockBackups: Backup[] = [
  {
    id: '1',
    filename: 'catalogizer-backup-20231209-full.tar.gz',
    size: 1073741824, // 1GB
    createdAt: '2023-12-09T05:00:00Z',
    type: 'full',
    status: 'completed'
  },
  {
    id: '2',
    filename: 'catalogizer-backup-20231208-incremental.tar.gz',
    size: 268435456, // 256MB
    createdAt: '2023-12-08T05:00:00Z',
    type: 'incremental',
    status: 'completed'
  }
];

function isFallbackError(error: unknown): boolean {
  const err = error as { response?: { status?: number } };
  return err.response?.status === 404;
}

export const adminApi = {
  async getSystemInfo(): Promise<SystemInfo> {
    try {
      const response = await api.get('/admin/system-info');
      return response.data;
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return mockSystemInfo;
      }
      throw error;
    }
  },

  async getUsers(): Promise<User[]> {
    try {
      const response = await api.get('/admin/users');
      return response.data;
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return mockUsers;
      }
      throw error;
    }
  },

  async getStorageInfo(): Promise<StorageInfo[]> {
    try {
      const response = await api.get('/admin/storage');
      return response.data;
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return mockStorageInfo;
      }
      throw error;
    }
  },

  async getBackups(): Promise<Backup[]> {
    try {
      const response = await api.get('/admin/backups');
      return response.data;
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return mockBackups;
      }
      throw error;
    }
  },

  async createBackup(type: 'full' | 'incremental'): Promise<void> {
    try {
      await api.post('/admin/backups', { type });
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return;
      }
      throw error;
    }
  },

  async restoreBackup(id: string): Promise<void> {
    try {
      await api.post(`/admin/backups/${id}/restore`);
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return;
      }
      throw error;
    }
  },

  async scanStorage(path: string): Promise<void> {
    try {
      await api.post('/admin/storage/scan', { path });
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return;
      }
      throw error;
    }
  },

  async updateUser(id: string, updates: Partial<User>): Promise<void> {
    try {
      await api.put(`/admin/users/${id}`, updates);
    } catch (error: unknown) {
      if (isFallbackError(error)) {
        return;
      }
      throw error;
    }
  }
};
