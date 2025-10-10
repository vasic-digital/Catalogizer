import { HttpClient } from '../utils/http';
import { SMBConfig, SMBStatus, CreateSMBConfigRequest } from '../types';
export declare class SMBService {
    private http;
    constructor(http: HttpClient);
    /**
     * Get all SMB configurations
     */
    getConfigs(): Promise<SMBConfig[]>;
    /**
     * Get a specific SMB configuration
     */
    getConfig(id: number): Promise<SMBConfig>;
    /**
     * Create a new SMB configuration
     */
    createConfig(config: CreateSMBConfigRequest): Promise<SMBConfig>;
    /**
     * Update an existing SMB configuration
     */
    updateConfig(id: number, updates: Partial<CreateSMBConfigRequest>): Promise<SMBConfig>;
    /**
     * Delete an SMB configuration
     */
    deleteConfig(id: number): Promise<void>;
    /**
     * Test connection to an SMB share
     */
    testConnection(config: CreateSMBConfigRequest): Promise<{
        success: boolean;
        message: string;
    }>;
    /**
     * Test existing SMB configuration
     */
    testExistingConfig(id: number): Promise<{
        success: boolean;
        message: string;
    }>;
    /**
     * Get status of all SMB connections
     */
    getStatus(): Promise<SMBStatus[]>;
    /**
     * Get status of a specific SMB connection
     */
    getConfigStatus(id: number): Promise<SMBStatus>;
    /**
     * Connect to an SMB share
     */
    connect(id: number): Promise<{
        success: boolean;
        message: string;
    }>;
    /**
     * Disconnect from an SMB share
     */
    disconnect(id: number): Promise<{
        success: boolean;
        message: string;
    }>;
    /**
     * Reconnect to an SMB share
     */
    reconnect(id: number): Promise<{
        success: boolean;
        message: string;
    }>;
    /**
     * Scan an SMB share for media files
     */
    scan(id: number, options?: {
        deep_scan?: boolean;
        update_metadata?: boolean;
        dry_run?: boolean;
    }): Promise<{
        job_id: number;
        message: string;
    }>;
    /**
     * Get scan job status
     */
    getScanStatus(jobId: number): Promise<{
        id: number;
        status: string;
        progress: number;
        found_items: number;
        processed_items: number;
        error_message?: string;
        created_at: string;
        updated_at: string;
    }>;
    /**
     * Cancel a scan job
     */
    cancelScan(jobId: number): Promise<void>;
    /**
     * Get list of scan jobs
     */
    getScanJobs(configId?: number): Promise<Array<{
        id: number;
        config_id: number;
        status: string;
        progress: number;
        found_items: number;
        processed_items: number;
        error_message?: string;
        created_at: string;
        updated_at: string;
    }>>;
    /**
     * Browse directories in an SMB share
     */
    browse(id: number, path?: string): Promise<{
        current_path: string;
        directories: Array<{
            name: string;
            path: string;
        }>;
        files: Array<{
            name: string;
            path: string;
            size: number;
            modified: string;
        }>;
    }>;
    /**
     * Enable/disable an SMB configuration
     */
    toggleConfig(id: number, isActive: boolean): Promise<SMBConfig>;
    /**
     * Get SMB share information
     */
    getShareInfo(id: number): Promise<{
        total_space: number;
        free_space: number;
        used_space: number;
        mount_point: string;
        share_name: string;
        server_name: string;
    }>;
    /**
     * Refresh connection to all active SMB shares
     */
    refreshAllConnections(): Promise<{
        refreshed: number;
        failed: number;
        results: Array<{
            config_id: number;
            success: boolean;
            message: string;
        }>;
    }>;
    /**
     * Get SMB connection logs
     */
    getLogs(id?: number, limit?: number): Promise<Array<{
        id: number;
        config_id: number;
        level: string;
        message: string;
        timestamp: string;
    }>>;
}
//# sourceMappingURL=SMBService.d.ts.map