import { HttpClient } from '../utils/http';
import {
  SMBConfig,
  SMBStatus,
  CreateSMBConfigRequest,
} from '../types';

export class SMBService {
  constructor(private http: HttpClient) {}

  /**
   * Get all SMB configurations
   */
  public async getConfigs(): Promise<SMBConfig[]> {
    return this.http.get<SMBConfig[]>('/smb/configs');
  }

  /**
   * Get a specific SMB configuration
   */
  public async getConfig(id: number): Promise<SMBConfig> {
    return this.http.get<SMBConfig>(`/smb/configs/${id}`);
  }

  /**
   * Create a new SMB configuration
   */
  public async createConfig(config: CreateSMBConfigRequest): Promise<SMBConfig> {
    return this.http.post<SMBConfig>('/smb/configs', config);
  }

  /**
   * Update an existing SMB configuration
   */
  public async updateConfig(id: number, updates: Partial<CreateSMBConfigRequest>): Promise<SMBConfig> {
    return this.http.put<SMBConfig>(`/smb/configs/${id}`, updates);
  }

  /**
   * Delete an SMB configuration
   */
  public async deleteConfig(id: number): Promise<void> {
    return this.http.delete<void>(`/smb/configs/${id}`);
  }

  /**
   * Test connection to an SMB share
   */
  public async testConnection(config: CreateSMBConfigRequest): Promise<{ success: boolean; message: string }> {
    return this.http.post<{ success: boolean; message: string }>('/smb/test', config);
  }

  /**
   * Test existing SMB configuration
   */
  public async testExistingConfig(id: number): Promise<{ success: boolean; message: string }> {
    return this.http.post<{ success: boolean; message: string }>(`/smb/configs/${id}/test`);
  }

  /**
   * Get status of all SMB connections
   */
  public async getStatus(): Promise<SMBStatus[]> {
    return this.http.get<SMBStatus[]>('/smb/status');
  }

  /**
   * Get status of a specific SMB connection
   */
  public async getConfigStatus(id: number): Promise<SMBStatus> {
    return this.http.get<SMBStatus>(`/smb/status/${id}`);
  }

  /**
   * Connect to an SMB share
   */
  public async connect(id: number): Promise<{ success: boolean; message: string }> {
    return this.http.post<{ success: boolean; message: string }>(`/smb/connect/${id}`);
  }

  /**
   * Disconnect from an SMB share
   */
  public async disconnect(id: number): Promise<{ success: boolean; message: string }> {
    return this.http.post<{ success: boolean; message: string }>(`/smb/disconnect/${id}`);
  }

  /**
   * Reconnect to an SMB share
   */
  public async reconnect(id: number): Promise<{ success: boolean; message: string }> {
    // Disconnect first, then connect
    await this.disconnect(id);
    return this.connect(id);
  }

  /**
   * Scan an SMB share for media files
   */
  public async scan(id: number, options?: {
    deep_scan?: boolean;
    update_metadata?: boolean;
    dry_run?: boolean;
  }): Promise<{ job_id: number; message: string }> {
    return this.http.post<{ job_id: number; message: string }>(`/smb/scan/${id}`, options || {});
  }

  /**
   * Get scan job status
   */
  public async getScanStatus(jobId: number): Promise<{
    id: number;
    status: string;
    progress: number;
    found_items: number;
    processed_items: number;
    error_message?: string;
    created_at: string;
    updated_at: string;
  }> {
    return this.http.get<{
      id: number;
      status: string;
      progress: number;
      found_items: number;
      processed_items: number;
      error_message?: string;
      created_at: string;
      updated_at: string;
    }>(`/smb/scan-jobs/${jobId}`);
  }

  /**
   * Cancel a scan job
   */
  public async cancelScan(jobId: number): Promise<void> {
    return this.http.post<void>(`/smb/scan-jobs/${jobId}/cancel`);
  }

  /**
   * Get list of scan jobs
   */
  public async getScanJobs(configId?: number): Promise<Array<{
    id: number;
    config_id: number;
    status: string;
    progress: number;
    found_items: number;
    processed_items: number;
    error_message?: string;
    created_at: string;
    updated_at: string;
  }>> {
    const params = configId ? `?config_id=${configId}` : '';
    return this.http.get<Array<{
      id: number;
      config_id: number;
      status: string;
      progress: number;
      found_items: number;
      processed_items: number;
      error_message?: string;
      created_at: string;
      updated_at: string;
    }>>(`/smb/scan-jobs${params}`);
  }

  /**
   * Browse directories in an SMB share
   */
  public async browse(id: number, path = ''): Promise<{
    current_path: string;
    directories: Array<{ name: string; path: string }>;
    files: Array<{ name: string; path: string; size: number; modified: string }>;
  }> {
    const params = path ? `?path=${encodeURIComponent(path)}` : '';
    return this.http.get<{
      current_path: string;
      directories: Array<{ name: string; path: string }>;
      files: Array<{ name: string; path: string; size: number; modified: string }>;
    }>(`/smb/configs/${id}/browse${params}`);
  }

  /**
   * Enable/disable an SMB configuration
   */
  public async toggleConfig(id: number, isActive: boolean): Promise<SMBConfig> {
    return this.http.patch<SMBConfig>(`/smb/configs/${id}`, { is_active: isActive });
  }

  /**
   * Get SMB share information
   */
  public async getShareInfo(id: number): Promise<{
    total_space: number;
    free_space: number;
    used_space: number;
    mount_point: string;
    share_name: string;
    server_name: string;
  }> {
    return this.http.get<{
      total_space: number;
      free_space: number;
      used_space: number;
      mount_point: string;
      share_name: string;
      server_name: string;
    }>(`/smb/configs/${id}/info`);
  }

  /**
   * Refresh connection to all active SMB shares
   */
  public async refreshAllConnections(): Promise<{
    refreshed: number;
    failed: number;
    results: Array<{ config_id: number; success: boolean; message: string }>;
  }> {
    return this.http.post<{
      refreshed: number;
      failed: number;
      results: Array<{ config_id: number; success: boolean; message: string }>;
    }>('/smb/refresh-all');
  }

  /**
   * Get SMB connection logs
   */
  public async getLogs(id?: number, limit = 100): Promise<Array<{
    id: number;
    config_id: number;
    level: string;
    message: string;
    timestamp: string;
  }>> {
    const params = new URLSearchParams();
    if (id) params.append('config_id', id.toString());
    params.append('limit', limit.toString());

    const query = params.toString();
    return this.http.get<Array<{
      id: number;
      config_id: number;
      level: string;
      message: string;
      timestamp: string;
    }>>(`/smb/logs?${query}`);
  }
}