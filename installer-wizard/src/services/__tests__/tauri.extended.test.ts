import { describe, it, expect, vi, beforeEach } from 'vitest'
import { TauriService } from '../tauri'

describe('TauriService (extended)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('validateConfiguration edge cases', () => {
    it('rejects null input', () => {
      expect(TauriService.validateConfiguration(null)).toBe(false)
    })

    it('rejects undefined input', () => {
      expect(TauriService.validateConfiguration(undefined)).toBe(false)
    })

    it('rejects number input', () => {
      expect(TauriService.validateConfiguration(42)).toBe(false)
    })

    it('rejects string input', () => {
      expect(TauriService.validateConfiguration('not a config')).toBe(false)
    })

    it('rejects boolean input', () => {
      expect(TauriService.validateConfiguration(true)).toBe(false)
    })

    it('rejects empty object', () => {
      expect(TauriService.validateConfiguration({})).toBe(false)
    })

    it('rejects object with only accesses', () => {
      expect(TauriService.validateConfiguration({ accesses: [] })).toBe(false)
    })

    it('rejects object with only sources', () => {
      expect(TauriService.validateConfiguration({ sources: [] })).toBe(false)
    })

    it('accepts empty accesses and sources arrays', () => {
      expect(TauriService.validateConfiguration({
        accesses: [],
        sources: [],
      })).toBe(true)
    })

    it('validates access with all required string fields', () => {
      expect(TauriService.validateConfiguration({
        accesses: [{
          name: 'test',
          type: 'credentials',
          account: 'user',
          secret: 'pass',
        }],
        sources: [],
      })).toBe(true)
    })

    it('rejects access with numeric name', () => {
      expect(TauriService.validateConfiguration({
        accesses: [{
          name: 123,
          type: 'credentials',
          account: 'user',
          secret: 'pass',
        }],
        sources: [],
      })).toBe(false)
    })

    it('rejects access with missing name', () => {
      expect(TauriService.validateConfiguration({
        accesses: [{
          type: 'credentials',
          account: 'user',
          secret: 'pass',
        }],
        sources: [],
      })).toBe(false)
    })

    it('rejects access with missing type', () => {
      expect(TauriService.validateConfiguration({
        accesses: [{
          name: 'test',
          account: 'user',
          secret: 'pass',
        }],
        sources: [],
      })).toBe(false)
    })

    it('rejects access with missing account', () => {
      expect(TauriService.validateConfiguration({
        accesses: [{
          name: 'test',
          type: 'credentials',
          secret: 'pass',
        }],
        sources: [],
      })).toBe(false)
    })

    it('rejects access with missing secret', () => {
      expect(TauriService.validateConfiguration({
        accesses: [{
          name: 'test',
          type: 'credentials',
          account: 'user',
        }],
        sources: [],
      })).toBe(false)
    })

    it('validates source with all required string fields', () => {
      expect(TauriService.validateConfiguration({
        accesses: [],
        sources: [{
          type: 'samba',
          url: 'smb://host/share',
          access: 'cred_name',
        }],
      })).toBe(true)
    })

    it('rejects source with missing type', () => {
      expect(TauriService.validateConfiguration({
        accesses: [],
        sources: [{
          url: 'smb://host/share',
          access: 'cred_name',
        }],
      })).toBe(false)
    })

    it('rejects source with missing url', () => {
      expect(TauriService.validateConfiguration({
        accesses: [],
        sources: [{
          type: 'samba',
          access: 'cred_name',
        }],
      })).toBe(false)
    })

    it('rejects source with missing access', () => {
      expect(TauriService.validateConfiguration({
        accesses: [],
        sources: [{
          type: 'samba',
          url: 'smb://host/share',
        }],
      })).toBe(false)
    })

    it('validates multiple accesses and sources', () => {
      expect(TauriService.validateConfiguration({
        accesses: [
          { name: 'cred1', type: 'credentials', account: 'u1', secret: 's1' },
          { name: 'cred2', type: 'credentials', account: 'u2', secret: 's2' },
        ],
        sources: [
          { type: 'samba', url: 'smb://h1/s1', access: 'cred1' },
          { type: 'ftp', url: 'ftp://h2:21', access: 'cred2' },
        ],
      })).toBe(true)
    })

    it('rejects if any single access is invalid among valid ones', () => {
      expect(TauriService.validateConfiguration({
        accesses: [
          { name: 'valid', type: 'credentials', account: 'u', secret: 's' },
          { name: 'invalid' },
        ],
        sources: [],
      })).toBe(false)
    })

    it('rejects if any single source is invalid among valid ones', () => {
      expect(TauriService.validateConfiguration({
        accesses: [],
        sources: [
          { type: 'samba', url: 'smb://h/s', access: 'c' },
          { type: 'ftp' },
        ],
      })).toBe(false)
    })

    it('accepts accesses with empty string values', () => {
      // Empty strings are still valid typeof string
      expect(TauriService.validateConfiguration({
        accesses: [{
          name: '',
          type: '',
          account: '',
          secret: '',
        }],
        sources: [],
      })).toBe(true)
    })
  })

  describe('scanSMBShares', () => {
    it('calls invoke with correct host parameter', async () => {
      const mockShares = [
        { host: '192.168.1.100', share_name: 'media', path: '/media', writable: true },
      ]
      global.mockInvoke.mockResolvedValue(mockShares)

      const result = await TauriService.scanSMBShares('192.168.1.100')

      expect(global.mockInvoke).toHaveBeenCalledWith('scan_smb_shares', {
        host: '192.168.1.100',
      })
      expect(result).toEqual(mockShares)
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('Host unreachable'))

      await expect(
        TauriService.scanSMBShares('192.168.1.200')
      ).rejects.toThrow('SMB share scan failed')
    })
  })

  describe('browseSMBShare', () => {
    it('calls invoke with host, share, and path', async () => {
      const mockEntries = [
        { name: 'movie.mkv', path: '/media/movie.mkv', is_directory: false, size: 1024 },
      ]
      global.mockInvoke.mockResolvedValue(mockEntries)

      const result = await TauriService.browseSMBShare('192.168.1.100', 'media', '/movies')

      expect(global.mockInvoke).toHaveBeenCalledWith('browse_smb_share', {
        host: '192.168.1.100',
        share: 'media',
        path: '/movies',
      })
      expect(result).toEqual(mockEntries)
    })

    it('calls invoke without path when not provided', async () => {
      global.mockInvoke.mockResolvedValue([])

      await TauriService.browseSMBShare('192.168.1.100', 'media')

      expect(global.mockInvoke).toHaveBeenCalledWith('browse_smb_share', {
        host: '192.168.1.100',
        share: 'media',
        path: undefined,
      })
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('Access denied'))

      await expect(
        TauriService.browseSMBShare('192.168.1.100', 'private')
      ).rejects.toThrow('SMB share browsing failed')
    })
  })

  describe('testFTPConnection', () => {
    it('calls invoke with all parameters', async () => {
      global.mockInvoke.mockResolvedValue(true)

      const result = await TauriService.testFTPConnection(
        'ftp.example.com', 21, 'ftpuser', 'ftppass', '/data'
      )

      expect(global.mockInvoke).toHaveBeenCalledWith('test_ftp_connection', {
        host: 'ftp.example.com',
        port: 21,
        username: 'ftpuser',
        password: 'ftppass',
        path: '/data',
      })
      expect(result).toBe(true)
    })

    it('handles missing path parameter', async () => {
      global.mockInvoke.mockResolvedValue(true)

      await TauriService.testFTPConnection('ftp.example.com', 21, 'user', 'pass')

      expect(global.mockInvoke).toHaveBeenCalledWith('test_ftp_connection', {
        host: 'ftp.example.com',
        port: 21,
        username: 'user',
        password: 'pass',
        path: undefined,
      })
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('Connection refused'))

      await expect(
        TauriService.testFTPConnection('ftp.example.com', 21, 'u', 'p')
      ).rejects.toThrow('FTP connection test failed')
    })
  })

  describe('testNFSConnection', () => {
    it('calls invoke with all parameters', async () => {
      global.mockInvoke.mockResolvedValue(true)

      const result = await TauriService.testNFSConnection(
        '192.168.1.100', '/export/data', '/mnt/nfs', 'vers=3'
      )

      expect(global.mockInvoke).toHaveBeenCalledWith('test_nfs_connection', {
        host: '192.168.1.100',
        path: '/export/data',
        mountPoint: '/mnt/nfs',
        options: 'vers=3',
      })
      expect(result).toBe(true)
    })

    it('handles missing options', async () => {
      global.mockInvoke.mockResolvedValue(true)

      await TauriService.testNFSConnection('192.168.1.100', '/export', '/mnt')

      expect(global.mockInvoke).toHaveBeenCalledWith('test_nfs_connection', {
        host: '192.168.1.100',
        path: '/export',
        mountPoint: '/mnt',
        options: undefined,
      })
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('Mount failed'))

      await expect(
        TauriService.testNFSConnection('h', '/p', '/m')
      ).rejects.toThrow('NFS connection test failed')
    })
  })

  describe('testWebDAVConnection', () => {
    it('calls invoke with all parameters', async () => {
      global.mockInvoke.mockResolvedValue(true)

      const result = await TauriService.testWebDAVConnection(
        'https://webdav.example.com', 'user', 'pass', '/files'
      )

      expect(global.mockInvoke).toHaveBeenCalledWith('test_webdav_connection', {
        url: 'https://webdav.example.com',
        username: 'user',
        password: 'pass',
        path: '/files',
      })
      expect(result).toBe(true)
    })

    it('handles missing path', async () => {
      global.mockInvoke.mockResolvedValue(true)

      await TauriService.testWebDAVConnection('https://dav.com', 'u', 'p')

      expect(global.mockInvoke).toHaveBeenCalledWith('test_webdav_connection', {
        url: 'https://dav.com',
        username: 'u',
        password: 'p',
        path: undefined,
      })
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('SSL error'))

      await expect(
        TauriService.testWebDAVConnection('https://x.com', 'u', 'p')
      ).rejects.toThrow('WebDAV connection test failed')
    })
  })

  describe('testLocalConnection', () => {
    it('calls invoke with base path', async () => {
      global.mockInvoke.mockResolvedValue(true)

      const result = await TauriService.testLocalConnection('/home/user/media')

      expect(global.mockInvoke).toHaveBeenCalledWith('test_local_connection', {
        basePath: '/home/user/media',
      })
      expect(result).toBe(true)
    })

    it('returns false for inaccessible path', async () => {
      global.mockInvoke.mockResolvedValue(false)

      const result = await TauriService.testLocalConnection('/nonexistent')

      expect(result).toBe(false)
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('Permission denied'))

      await expect(
        TauriService.testLocalConnection('/root/secret')
      ).rejects.toThrow('Local connection test failed')
    })
  })

  describe('getDefaultConfigPath', () => {
    it('returns the default config path', async () => {
      global.mockInvoke.mockResolvedValue('/home/user/.catalogizer/config.json')

      const result = await TauriService.getDefaultConfigPath()

      expect(global.mockInvoke).toHaveBeenCalledWith('get_default_config_path')
      expect(result).toBe('/home/user/.catalogizer/config.json')
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('Home dir not found'))

      await expect(
        TauriService.getDefaultConfigPath()
      ).rejects.toThrow('Getting default config path failed')
    })
  })

  describe('loadConfiguration', () => {
    it('loads config from specific path', async () => {
      const mockConfig = {
        accesses: [{ name: 'a', type: 'credentials', account: 'u', secret: 's' }],
        sources: [{ type: 'samba', url: 'smb://h/s', access: 'a' }],
      }
      global.mockInvoke.mockResolvedValue(mockConfig)

      const result = await TauriService.loadConfiguration('/custom/config.json')

      expect(global.mockInvoke).toHaveBeenCalledWith('load_configuration', {
        filePath: '/custom/config.json',
      })
      expect(result).toEqual(mockConfig)
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('File not found'))

      await expect(
        TauriService.loadConfiguration('/missing/config.json')
      ).rejects.toThrow('Configuration loading failed')
    })
  })

  describe('saveConfiguration', () => {
    it('saves config to specific path', async () => {
      const config = { accesses: [], sources: [] }
      global.mockInvoke.mockResolvedValue(undefined)

      await TauriService.saveConfiguration('/output/config.json', config)

      expect(global.mockInvoke).toHaveBeenCalledWith('save_configuration', {
        filePath: '/output/config.json',
        config,
      })
    })

    it('throws descriptive error on failure', async () => {
      global.mockInvoke.mockRejectedValue(new Error('Disk full'))

      await expect(
        TauriService.saveConfiguration('/path', { accesses: [], sources: [] })
      ).rejects.toThrow('Configuration saving failed')
    })
  })
})
