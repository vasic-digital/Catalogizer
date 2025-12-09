import type { SystemInfo, User, StorageInfo, Backup } from '@/types/admin';

export const adminApi = {
  async getSystemInfo(): Promise<SystemInfo> {
    // Mock implementation - would be replaced with actual API call
    return {
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
  },

  async getUsers(): Promise<User[]> {
    // Mock implementation
    return [
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
  },

  async getStorageInfo(): Promise<StorageInfo[]> {
    // Mock implementation
    return [
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
  },

  async getBackups(): Promise<Backup[]> {
    // Mock implementation
    return [
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
  },

  async createBackup(type: 'full' | 'incremental'): Promise<void> {
    // Mock implementation
    console.log(`Creating ${type} backup`);
  },

  async restoreBackup(id: string): Promise<void> {
    // Mock implementation
    console.log(`Restoring backup ${id}`);
  },

  async scanStorage(path: string): Promise<void> {
    // Mock implementation
    console.log(`Scanning storage ${path}`);
  },

  async updateUser(id: string, updates: Partial<User>): Promise<void> {
    // Mock implementation
    console.log(`Updating user ${id}:`, updates);
  }
};