import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SMBService } from '../SMBService';
import { HttpClient } from '../../utils/http';

// Mock HttpClient
vi.mock('../../utils/http');

describe('SMBService', () => {
  let smbService: SMBService;
  let mockHttp: any;

  beforeEach(() => {
    vi.clearAllMocks();

    mockHttp = {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      patch: vi.fn(),
      delete: vi.fn(),
    };

    smbService = new SMBService(mockHttp as any);
  });

  describe('config management', () => {
    it('gets all SMB configurations', async () => {
      const configs = [{ id: 1, name: 'Media NAS', host: '192.168.1.100' }];
      mockHttp.get.mockResolvedValueOnce(configs);

      const result = await smbService.getConfigs();

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/configs');
      expect(result).toEqual(configs);
    });

    it('gets a specific SMB configuration', async () => {
      const config = { id: 1, name: 'Media NAS', host: '192.168.1.100' };
      mockHttp.get.mockResolvedValueOnce(config);

      const result = await smbService.getConfig(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/configs/1');
      expect(result).toEqual(config);
    });

    it('creates a new SMB configuration', async () => {
      const newConfig = {
        name: 'New Share',
        host: '192.168.1.200',
        port: 445,
        share_name: 'media',
        username: 'admin',
        password: 'secret',
        mount_point: '/mnt/media',
      };
      const created = { id: 2, ...newConfig };
      mockHttp.post.mockResolvedValueOnce(created);

      const result = await smbService.createConfig(newConfig);

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/configs', newConfig);
      expect(result).toEqual(created);
    });

    it('updates an existing SMB configuration', async () => {
      const updated = { id: 1, name: 'Updated NAS' };
      mockHttp.put.mockResolvedValueOnce(updated);

      const result = await smbService.updateConfig(1, { name: 'Updated NAS' });

      expect(mockHttp.put).toHaveBeenCalledWith('/smb/configs/1', { name: 'Updated NAS' });
      expect(result).toEqual(updated);
    });

    it('deletes an SMB configuration', async () => {
      mockHttp.delete.mockResolvedValueOnce(undefined);

      await smbService.deleteConfig(1);

      expect(mockHttp.delete).toHaveBeenCalledWith('/smb/configs/1');
    });

    it('toggles config active state', async () => {
      const toggled = { id: 1, is_active: false };
      mockHttp.patch.mockResolvedValueOnce(toggled);

      const result = await smbService.toggleConfig(1, false);

      expect(mockHttp.patch).toHaveBeenCalledWith('/smb/configs/1', { is_active: false });
      expect(result).toEqual(toggled);
    });
  });

  describe('connection testing', () => {
    it('tests connection with new config', async () => {
      const testResult = { success: true, message: 'Connection successful' };
      mockHttp.post.mockResolvedValueOnce(testResult);

      const config = {
        name: 'Test',
        host: '192.168.1.100',
        port: 445,
        share_name: 'media',
        username: 'user',
        password: 'pass',
        mount_point: '/mnt/test',
      };
      const result = await smbService.testConnection(config);

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/test', config);
      expect(result).toEqual(testResult);
    });

    it('tests existing SMB configuration', async () => {
      const testResult = { success: true, message: 'OK' };
      mockHttp.post.mockResolvedValueOnce(testResult);

      const result = await smbService.testExistingConfig(1);

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/configs/1/test');
      expect(result).toEqual(testResult);
    });
  });

  describe('connection management', () => {
    it('connects to an SMB share', async () => {
      const connectResult = { success: true, message: 'Connected' };
      mockHttp.post.mockResolvedValueOnce(connectResult);

      const result = await smbService.connect(1);

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/connect/1');
      expect(result).toEqual(connectResult);
    });

    it('disconnects from an SMB share', async () => {
      const disconnectResult = { success: true, message: 'Disconnected' };
      mockHttp.post.mockResolvedValueOnce(disconnectResult);

      const result = await smbService.disconnect(1);

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/disconnect/1');
      expect(result).toEqual(disconnectResult);
    });

    it('reconnects by disconnecting then connecting', async () => {
      const disconnectResult = { success: true, message: 'Disconnected' };
      const connectResult = { success: true, message: 'Connected' };
      mockHttp.post
        .mockResolvedValueOnce(disconnectResult)
        .mockResolvedValueOnce(connectResult);

      const result = await smbService.reconnect(1);

      expect(mockHttp.post).toHaveBeenCalledTimes(2);
      expect(mockHttp.post).toHaveBeenNthCalledWith(1, '/smb/disconnect/1');
      expect(mockHttp.post).toHaveBeenNthCalledWith(2, '/smb/connect/1');
      expect(result).toEqual(connectResult);
    });

    it('refreshes all connections', async () => {
      const refreshResult = {
        refreshed: 2,
        failed: 0,
        results: [
          { config_id: 1, success: true, message: 'OK' },
          { config_id: 2, success: true, message: 'OK' },
        ],
      };
      mockHttp.post.mockResolvedValueOnce(refreshResult);

      const result = await smbService.refreshAllConnections();

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/refresh-all');
      expect(result).toEqual(refreshResult);
    });
  });

  describe('status', () => {
    it('gets all SMB connection statuses', async () => {
      const statuses = [{ config_id: 1, is_connected: true, last_check: '2024-01-01' }];
      mockHttp.get.mockResolvedValueOnce(statuses);

      const result = await smbService.getStatus();

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/status');
      expect(result).toEqual(statuses);
    });

    it('gets status for a specific config', async () => {
      const status = { config_id: 1, is_connected: false, error_message: 'Timeout' };
      mockHttp.get.mockResolvedValueOnce(status);

      const result = await smbService.getConfigStatus(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/status/1');
      expect(result).toEqual(status);
    });
  });

  describe('scanning', () => {
    it('starts a scan with options', async () => {
      const scanResult = { job_id: 42, message: 'Scan started' };
      mockHttp.post.mockResolvedValueOnce(scanResult);

      const result = await smbService.scan(1, { deep_scan: true, update_metadata: true });

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/scan/1', { deep_scan: true, update_metadata: true });
      expect(result).toEqual(scanResult);
    });

    it('starts a scan with default options', async () => {
      const scanResult = { job_id: 43, message: 'Scan started' };
      mockHttp.post.mockResolvedValueOnce(scanResult);

      await smbService.scan(1);

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/scan/1', {});
    });

    it('gets scan job status', async () => {
      const scanStatus = { id: 42, status: 'in_progress', progress: 60, found_items: 100, processed_items: 60 };
      mockHttp.get.mockResolvedValueOnce(scanStatus);

      const result = await smbService.getScanStatus(42);

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/scan-jobs/42');
      expect(result).toEqual(scanStatus);
    });

    it('cancels a scan job', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);

      await smbService.cancelScan(42);

      expect(mockHttp.post).toHaveBeenCalledWith('/smb/scan-jobs/42/cancel');
    });

    it('gets all scan jobs', async () => {
      const jobs = [{ id: 1, config_id: 1, status: 'completed' }];
      mockHttp.get.mockResolvedValueOnce(jobs);

      const result = await smbService.getScanJobs();

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/scan-jobs');
      expect(result).toEqual(jobs);
    });

    it('gets scan jobs for a specific config', async () => {
      const jobs = [{ id: 1, config_id: 5, status: 'completed' }];
      mockHttp.get.mockResolvedValueOnce(jobs);

      const result = await smbService.getScanJobs(5);

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/scan-jobs?config_id=5');
      expect(result).toEqual(jobs);
    });
  });

  describe('browsing', () => {
    it('browses root directory', async () => {
      const browseResult = {
        current_path: '/',
        directories: [{ name: 'movies', path: '/movies' }],
        files: [],
      };
      mockHttp.get.mockResolvedValueOnce(browseResult);

      const result = await smbService.browse(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/configs/1/browse');
      expect(result).toEqual(browseResult);
    });

    it('browses a specific path', async () => {
      const browseResult = {
        current_path: '/movies',
        directories: [],
        files: [{ name: 'movie.mp4', path: '/movies/movie.mp4', size: 1000000, modified: '2024-01-01' }],
      };
      mockHttp.get.mockResolvedValueOnce(browseResult);

      const result = await smbService.browse(1, '/movies');

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/configs/1/browse?path=%2Fmovies');
      expect(result).toEqual(browseResult);
    });
  });

  describe('share info', () => {
    it('gets SMB share info', async () => {
      const shareInfo = {
        total_space: 2000000000,
        free_space: 1000000000,
        used_space: 1000000000,
        mount_point: '/mnt/nas',
        share_name: 'media',
        server_name: 'nas-server',
      };
      mockHttp.get.mockResolvedValueOnce(shareInfo);

      const result = await smbService.getShareInfo(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/configs/1/info');
      expect(result).toEqual(shareInfo);
    });
  });

  describe('logs', () => {
    it('gets SMB logs with default params', async () => {
      const logs = [{ id: 1, config_id: 1, level: 'info', message: 'Connected', timestamp: '2024-01-01' }];
      mockHttp.get.mockResolvedValueOnce(logs);

      const result = await smbService.getLogs();

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/logs?limit=100');
      expect(result).toEqual(logs);
    });

    it('gets logs for specific config with custom limit', async () => {
      const logs: any[] = [];
      mockHttp.get.mockResolvedValueOnce(logs);

      await smbService.getLogs(3, 50);

      expect(mockHttp.get).toHaveBeenCalledWith('/smb/logs?config_id=3&limit=50');
    });
  });
});
