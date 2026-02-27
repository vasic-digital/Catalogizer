/**
 * Example tests for Tauri commands
 */
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { invoke } from '@tauri-apps/api/core';

// Create local mock to avoid hoisting issues
const mockInvoke = vi.fn();

// Mock the Tauri API v2 - core module
vi.mock('@tauri-apps/api/core', () => ({
  invoke: (...args: any[]) => mockInvoke(...args),
}));

describe('Tauri Commands', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('get_app_info', () => {
    it('should return app info successfully', async () => {
      // Given
      const mockAppInfo = {
        name: 'Catalogizer Desktop',
        version: '1.0.0',
        author: 'Catalogizer Team',
      };
      mockInvoke.mockResolvedValue(mockAppInfo);

      // When
      const result = await invoke('get_app_info');

      // Then
      expect(result).toEqual(mockAppInfo);
      expect(mockInvoke).toHaveBeenCalledWith('get_app_info');
    });

    it('should handle errors when getting app info', async () => {
      // Given
      const errorMessage = 'Failed to get app info';
      mockInvoke.mockRejectedValue(new Error(errorMessage));

      // When & Then
      await expect(invoke('get_app_info')).rejects.toThrow(errorMessage);
    });
  });

  describe('get_system_info', () => {
    it('should return system info successfully', async () => {
      // Given
      const mockSystemInfo = {
        os: 'Linux',
        arch: 'x86_64',
        memory: 8192,
        cores: 8,
      };
      mockInvoke.mockResolvedValue(mockSystemInfo);

      // When
      const result = await invoke('get_system_info');

      // Then
      expect(result).toEqual(mockSystemInfo);
      expect(mockInvoke).toHaveBeenCalledWith('get_system_info');
    });
  });
});
