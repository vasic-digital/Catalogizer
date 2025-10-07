import { invoke } from '@tauri-apps/api/core'
import { open, save } from '@tauri-apps/plugin-dialog'
import { NetworkHost, SMBShare, FileEntry, Configuration } from '../types'

/**
 * Tauri service for interacting with the backend
 */
export class TauriService {
  /**
   * Scan the network for SMB hosts
   */
  static async scanNetwork(): Promise<NetworkHost[]> {
    try {
      return await invoke<NetworkHost[]>('scan_network')
    } catch (error) {
      console.error('Failed to scan network:', error)
      throw new Error(`Network scan failed: ${error}`)
    }
  }

  /**
   * Scan SMB shares on a specific host
   */
  static async scanSMBShares(host: string): Promise<SMBShare[]> {
    try {
      return await invoke<SMBShare[]>('scan_smb_shares', { host })
    } catch (error) {
      console.error('Failed to scan SMB shares:', error)
      throw new Error(`SMB share scan failed: ${error}`)
    }
  }

  /**
   * Browse files and directories in an SMB share
   */
  static async browseSMBShare(
    host: string,
    share: string,
    path?: string
  ): Promise<FileEntry[]> {
    try {
      return await invoke<FileEntry[]>('browse_smb_share', {
        host,
        share,
        path: path || undefined,
      })
    } catch (error) {
      console.error('Failed to browse SMB share:', error)
      throw new Error(`SMB share browsing failed: ${error}`)
    }
  }

  /**
   * Test SMB connection with credentials
   */
  static async testSMBConnection(
    host: string,
    share: string,
    username: string,
    password: string,
    domain?: string
  ): Promise<boolean> {
    try {
      return await invoke<boolean>('test_smb_connection', {
        host,
        share,
        username,
        password,
        domain: domain || undefined,
      })
    } catch (error) {
      console.error('Failed to test SMB connection:', error)
      throw new Error(`SMB connection test failed: ${error}`)
    }
  }

  /**
   * Test FTP connection with credentials
   */
  static async testFTPConnection(
    host: string,
    port: number,
    username: string,
    password: string,
    path?: string
  ): Promise<boolean> {
    try {
      return await invoke<boolean>('test_ftp_connection', {
        host,
        port,
        username,
        password,
        path: path || undefined,
      })
    } catch (error) {
      console.error('Failed to test FTP connection:', error)
      throw new Error(`FTP connection test failed: ${error}`)
    }
  }

  /**
   * Test NFS connection
   */
  static async testNFSConnection(
    host: string,
    path: string,
    mountPoint: string,
    options?: string
  ): Promise<boolean> {
    try {
      return await invoke<boolean>('test_nfs_connection', {
        host,
        path,
        mountPoint,
        options: options || undefined,
      })
    } catch (error) {
      console.error('Failed to test NFS connection:', error)
      throw new Error(`NFS connection test failed: ${error}`)
    }
  }

  /**
   * Test WebDAV connection with credentials
   */
  static async testWebDAVConnection(
    url: string,
    username: string,
    password: string,
    path?: string
  ): Promise<boolean> {
    try {
      return await invoke<boolean>('test_webdav_connection', {
        url,
        username,
        password,
        path: path || undefined,
      })
    } catch (error) {
      console.error('Failed to test WebDAV connection:', error)
      throw new Error(`WebDAV connection test failed: ${error}`)
    }
  }

  /**
   * Test local path accessibility
   */
  static async testLocalConnection(basePath: string): Promise<boolean> {
    try {
      return await invoke<boolean>('test_local_connection', {
        basePath,
      })
    } catch (error) {
      console.error('Failed to test local connection:', error)
      throw new Error(`Local connection test failed: ${error}`)
    }
  }

  /**
   * Load configuration from a file
   */
  static async loadConfiguration(filePath: string): Promise<Configuration> {
    try {
      return await invoke<Configuration>('load_configuration', { filePath })
    } catch (error) {
      console.error('Failed to load configuration:', error)
      throw new Error(`Configuration loading failed: ${error}`)
    }
  }

  /**
   * Save configuration to a file
   */
  static async saveConfiguration(
    filePath: string,
    config: Configuration
  ): Promise<void> {
    try {
      await invoke<void>('save_configuration', { filePath, config })
    } catch (error) {
      console.error('Failed to save configuration:', error)
      throw new Error(`Configuration saving failed: ${error}`)
    }
  }

  /**
   * Get the default configuration file path
   */
  static async getDefaultConfigPath(): Promise<string> {
    try {
      return await invoke<string>('get_default_config_path')
    } catch (error) {
      console.error('Failed to get default config path:', error)
      throw new Error(`Getting default config path failed: ${error}`)
    }
  }

  /**
   * Open file dialog to select configuration file
   */
  static async openConfigurationFile(): Promise<Configuration | null> {
    try {
      const filePath = await open({
        title: 'Open Configuration File',
        filters: [
          {
            name: 'JSON Configuration',
            extensions: ['json'],
          },
        ],
      })

      if (filePath) {
        return await this.loadConfiguration(filePath as string)
      }

      return null
    } catch (error) {
      console.error('Failed to open configuration file:', error)
      throw new Error(`Opening configuration file failed: ${error}`)
    }
  }

  /**
   * Save configuration file with dialog
   */
  static async saveConfigurationFile(config: Configuration): Promise<boolean> {
    try {
      const filePath = await save({
        title: 'Save Configuration File',
        defaultPath: 'catalogizer-config.json',
        filters: [
          {
            name: 'JSON Configuration',
            extensions: ['json'],
          },
        ],
      })

      if (filePath) {
        await this.saveConfiguration(filePath as string, config)
        return true
      }

      return false
    } catch (error) {
      console.error('Failed to save configuration file:', error)
      throw new Error(`Saving configuration file failed: ${error}`)
    }
  }

  /**
   * Validate configuration format
   */
  static validateConfiguration(config: any): config is Configuration {
    return (
      typeof config === 'object' &&
      config !== null &&
      Array.isArray(config.accesses) &&
      Array.isArray(config.sources) &&
      config.accesses.every((access: any) =>
        typeof access === 'object' &&
        typeof access.name === 'string' &&
        typeof access.type === 'string' &&
        typeof access.account === 'string' &&
        typeof access.secret === 'string'
      ) &&
      config.sources.every((source: any) =>
        typeof source === 'object' &&
        typeof source.type === 'string' &&
        typeof source.url === 'string' &&
        typeof source.access === 'string'
      )
    )
  }
}