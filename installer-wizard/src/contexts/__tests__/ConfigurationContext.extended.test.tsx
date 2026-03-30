import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { act } from 'react'
import { ConfigurationProvider, useConfiguration } from '../ConfigurationContext'
import type {
  Configuration,
  ConfigurationAccess,
  ConfigurationSource,
  NetworkHost,
  SMBConnectionConfig,
  FTPConnectionConfig,
  NFSConnectionConfig,
  WebDAVConnectionConfig,
  LocalConnectionConfig,
} from '../../types'

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <ConfigurationProvider>{children}</ConfigurationProvider>
)

describe('ConfigurationContext (extended edge cases)', () => {
  describe('useConfiguration outside provider', () => {
    it('throws when used outside ConfigurationProvider', () => {
      expect(() => {
        renderHook(() => useConfiguration())
      }).toThrow('useConfiguration must be used within a ConfigurationProvider')
    })
  })

  describe('setConfiguration', () => {
    it('replaces entire configuration', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const newConfig: Configuration = {
        accesses: [
          { name: 'a1', type: 'credentials', account: 'u1', secret: 's1' },
          { name: 'a2', type: 'credentials', account: 'u2', secret: 's2' },
        ],
        sources: [
          { type: 'samba', url: 'smb://h1/s1', access: 'a1' },
        ],
      }

      act(() => {
        result.current.setConfiguration(newConfig)
      })

      expect(result.current.state.configuration).toEqual(newConfig)
      expect(result.current.state.hasUnsavedChanges).toBe(false)
    })

    it('resets hasUnsavedChanges to false', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      // First make a change to set hasUnsavedChanges
      act(() => {
        result.current.addAccess({
          name: 'test', type: 'credentials', account: 'u', secret: 's',
        })
      })
      expect(result.current.state.hasUnsavedChanges).toBe(true)

      // Now setConfiguration should reset it
      act(() => {
        result.current.setConfiguration({ accesses: [], sources: [] })
      })
      expect(result.current.state.hasUnsavedChanges).toBe(false)
    })
  })

  describe('access CRUD operations', () => {
    it('adds multiple accesses sequentially', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const accesses: ConfigurationAccess[] = [
        { name: 'cred1', type: 'credentials', account: 'u1', secret: 's1' },
        { name: 'cred2', type: 'credentials', account: 'u2', secret: 's2' },
        { name: 'cred3', type: 'credentials', account: 'u3', secret: 's3' },
      ]

      act(() => {
        accesses.forEach(a => result.current.addAccess(a))
      })

      expect(result.current.state.configuration.accesses).toHaveLength(3)
      expect(result.current.state.configuration.accesses[2].name).toBe('cred3')
    })

    it('updateAccess at specific index does not affect other entries', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addAccess({ name: 'a', type: 'c', account: 'u1', secret: 's1' })
        result.current.addAccess({ name: 'b', type: 'c', account: 'u2', secret: 's2' })
        result.current.addAccess({ name: 'c', type: 'c', account: 'u3', secret: 's3' })
      })

      act(() => {
        result.current.updateAccess(1, {
          name: 'b-updated', type: 'c', account: 'u2-new', secret: 's2-new',
        })
      })

      expect(result.current.state.configuration.accesses[0].name).toBe('a')
      expect(result.current.state.configuration.accesses[1].name).toBe('b-updated')
      expect(result.current.state.configuration.accesses[2].name).toBe('c')
    })

    it('removeAccess at middle index shifts subsequent entries', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addAccess({ name: 'first', type: 'c', account: 'u1', secret: 's1' })
        result.current.addAccess({ name: 'second', type: 'c', account: 'u2', secret: 's2' })
        result.current.addAccess({ name: 'third', type: 'c', account: 'u3', secret: 's3' })
      })

      act(() => {
        result.current.removeAccess(1)
      })

      expect(result.current.state.configuration.accesses).toHaveLength(2)
      expect(result.current.state.configuration.accesses[0].name).toBe('first')
      expect(result.current.state.configuration.accesses[1].name).toBe('third')
    })

    it('removeAccess at last index removes last entry', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addAccess({ name: 'a', type: 'c', account: 'u', secret: 's' })
        result.current.addAccess({ name: 'b', type: 'c', account: 'u', secret: 's' })
      })

      act(() => {
        result.current.removeAccess(1)
      })

      expect(result.current.state.configuration.accesses).toHaveLength(1)
      expect(result.current.state.configuration.accesses[0].name).toBe('a')
    })
  })

  describe('source CRUD operations', () => {
    it('adds multiple sources sequentially', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addSource({ type: 'samba', url: 'smb://h1/s', access: 'c1' })
        result.current.addSource({ type: 'ftp', url: 'ftp://h2:21', access: 'c2' })
        result.current.addSource({ type: 'samba', url: 'smb://h3/m', access: 'c3' })
      })

      expect(result.current.state.configuration.sources).toHaveLength(3)
    })

    it('updateSource at specific index', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addSource({ type: 'samba', url: 'smb://old', access: 'c' })
      })

      act(() => {
        result.current.updateSource(0, { type: 'samba', url: 'smb://new', access: 'c' })
      })

      expect(result.current.state.configuration.sources[0].url).toBe('smb://new')
    })

    it('removeSource preserves remaining entries', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addSource({ type: 'samba', url: 'smb://h1', access: 'c1' })
        result.current.addSource({ type: 'ftp', url: 'ftp://h2', access: 'c2' })
      })

      act(() => {
        result.current.removeSource(0)
      })

      expect(result.current.state.configuration.sources).toHaveLength(1)
      expect(result.current.state.configuration.sources[0].type).toBe('ftp')
    })
  })

  describe('protocol config setters', () => {
    it('sets and clears SMB config', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const smbConfig: SMBConnectionConfig = {
        name: 'Test', host: '192.168.1.1', port: 445,
        share_name: 'share', username: 'u', password: 'p', enabled: true,
      }

      act(() => {
        result.current.setCurrentSMBConfig(smbConfig)
      })
      expect(result.current.state.currentSMBConfig).toEqual(smbConfig)

      act(() => {
        result.current.setCurrentSMBConfig(null)
      })
      expect(result.current.state.currentSMBConfig).toBeNull()
    })

    it('sets and clears FTP config', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const ftpConfig: FTPConnectionConfig = {
        name: 'FTP', host: '192.168.1.2', port: 21,
        username: 'ftp', password: 'ftp', enabled: true,
      }

      act(() => {
        result.current.setCurrentFTPConfig(ftpConfig)
      })
      expect(result.current.state.currentFTPConfig).toEqual(ftpConfig)

      act(() => {
        result.current.setCurrentFTPConfig(null)
      })
      expect(result.current.state.currentFTPConfig).toBeNull()
    })

    it('sets and clears NFS config', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const nfsConfig: NFSConnectionConfig = {
        name: 'NFS', host: '192.168.1.3', path: '/export',
        mount_point: '/mnt', enabled: true,
      }

      act(() => {
        result.current.setCurrentNFSConfig(nfsConfig)
      })
      expect(result.current.state.currentNFSConfig).toEqual(nfsConfig)

      act(() => {
        result.current.setCurrentNFSConfig(null)
      })
      expect(result.current.state.currentNFSConfig).toBeNull()
    })

    it('sets and clears WebDAV config', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const webdavConfig: WebDAVConnectionConfig = {
        name: 'DAV', url: 'https://dav.com', username: 'u',
        password: 'p', enabled: true,
      }

      act(() => {
        result.current.setCurrentWebDAVConfig(webdavConfig)
      })
      expect(result.current.state.currentWebDAVConfig).toEqual(webdavConfig)

      act(() => {
        result.current.setCurrentWebDAVConfig(null)
      })
      expect(result.current.state.currentWebDAVConfig).toBeNull()
    })

    it('sets and clears Local config', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const localConfig: LocalConnectionConfig = {
        name: 'Local', base_path: '/media', enabled: true,
      }

      act(() => {
        result.current.setCurrentLocalConfig(localConfig)
      })
      expect(result.current.state.currentLocalConfig).toEqual(localConfig)

      act(() => {
        result.current.setCurrentLocalConfig(null)
      })
      expect(result.current.state.currentLocalConfig).toBeNull()
    })
  })

  describe('selectedProtocol', () => {
    it('sets protocol to smb', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setSelectedProtocol('smb')
      })

      expect(result.current.state.selectedProtocol).toBe('smb')
    })

    it('sets protocol to ftp', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setSelectedProtocol('ftp')
      })

      expect(result.current.state.selectedProtocol).toBe('ftp')
    })

    it('sets protocol to nfs', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setSelectedProtocol('nfs')
      })

      expect(result.current.state.selectedProtocol).toBe('nfs')
    })

    it('sets protocol to webdav', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setSelectedProtocol('webdav')
      })

      expect(result.current.state.selectedProtocol).toBe('webdav')
    })

    it('sets protocol to local', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setSelectedProtocol('local')
      })

      expect(result.current.state.selectedProtocol).toBe('local')
    })

    it('clears protocol', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setSelectedProtocol('smb')
      })

      act(() => {
        result.current.setSelectedProtocol(null)
      })

      expect(result.current.state.selectedProtocol).toBeNull()
    })
  })

  describe('selectedHosts', () => {
    it('sets multiple hosts', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const hosts: NetworkHost[] = [
        { ip: '192.168.1.1', open_ports: [445], smb_shares: ['s1'] },
        { ip: '192.168.1.2', open_ports: [445, 139], smb_shares: ['s2', 's3'] },
      ]

      act(() => {
        result.current.setSelectedHosts(hosts)
      })

      expect(result.current.state.selectedHosts).toHaveLength(2)
      expect(result.current.state.selectedHosts[0].ip).toBe('192.168.1.1')
    })

    it('clears hosts with empty array', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setSelectedHosts([
          { ip: '10.0.0.1', open_ports: [], smb_shares: [] },
        ])
      })

      act(() => {
        result.current.setSelectedHosts([])
      })

      expect(result.current.state.selectedHosts).toHaveLength(0)
    })
  })

  describe('hasUnsavedChanges tracking', () => {
    it('starts as false', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      expect(result.current.state.hasUnsavedChanges).toBe(false)
    })

    it('becomes true after addAccess', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addAccess({ name: 'a', type: 'c', account: 'u', secret: 's' })
      })

      expect(result.current.state.hasUnsavedChanges).toBe(true)
    })

    it('becomes true after addSource', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.addSource({ type: 'samba', url: 'smb://h', access: 'a' })
      })

      expect(result.current.state.hasUnsavedChanges).toBe(true)
    })

    it('becomes true after updateAccess', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setConfiguration({
          accesses: [{ name: 'a', type: 'c', account: 'u', secret: 's' }],
          sources: [],
        })
      })

      expect(result.current.state.hasUnsavedChanges).toBe(false)

      act(() => {
        result.current.updateAccess(0, { name: 'b', type: 'c', account: 'u2', secret: 's2' })
      })

      expect(result.current.state.hasUnsavedChanges).toBe(true)
    })

    it('becomes true after removeAccess', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setConfiguration({
          accesses: [{ name: 'a', type: 'c', account: 'u', secret: 's' }],
          sources: [],
        })
      })

      act(() => {
        result.current.removeAccess(0)
      })

      expect(result.current.state.hasUnsavedChanges).toBe(true)
    })

    it('can be explicitly set', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setUnsavedChanges(true)
      })

      expect(result.current.state.hasUnsavedChanges).toBe(true)

      act(() => {
        result.current.setUnsavedChanges(false)
      })

      expect(result.current.state.hasUnsavedChanges).toBe(false)
    })
  })

  describe('error management', () => {
    it('sets and clears error', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setError('Something failed')
      })

      expect(result.current.state.error).toBe('Something failed')

      act(() => {
        result.current.clearError()
      })

      expect(result.current.state.error).toBeNull()
    })

    it('sets error to null directly', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setError('Error')
      })

      act(() => {
        result.current.setError(null)
      })

      expect(result.current.state.error).toBeNull()
    })
  })

  describe('loading state', () => {
    it('toggles loading', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.setLoading(true)
      })
      expect(result.current.state.isLoading).toBe(true)

      act(() => {
        result.current.setLoading(false)
      })
      expect(result.current.state.isLoading).toBe(false)
    })
  })

  describe('reset', () => {
    it('resets all state to initial values', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      // Modify everything
      act(() => {
        result.current.addAccess({ name: 'a', type: 'c', account: 'u', secret: 's' })
        result.current.addSource({ type: 'samba', url: 'smb://h', access: 'a' })
        result.current.setSelectedProtocol('smb')
        result.current.setSelectedHosts([{ ip: '1.2.3.4', open_ports: [], smb_shares: [] }])
        result.current.setLoading(true)
        result.current.setError('err')
        result.current.setCurrentSMBConfig({
          name: 'x', host: 'h', port: 445, share_name: 's',
          username: 'u', password: 'p', enabled: true,
        })
      })

      // Reset
      act(() => {
        result.current.reset()
      })

      const state = result.current.state
      expect(state.configuration.accesses).toHaveLength(0)
      expect(state.configuration.sources).toHaveLength(0)
      expect(state.currentSMBConfig).toBeNull()
      expect(state.currentFTPConfig).toBeNull()
      expect(state.currentNFSConfig).toBeNull()
      expect(state.currentWebDAVConfig).toBeNull()
      expect(state.currentLocalConfig).toBeNull()
      expect(state.selectedProtocol).toBeNull()
      expect(state.selectedHosts).toHaveLength(0)
      expect(state.isLoading).toBe(false)
      expect(state.error).toBeNull()
      expect(state.hasUnsavedChanges).toBe(false)
    })

    it('is idempotent', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      act(() => {
        result.current.reset()
        result.current.reset()
      })

      expect(result.current.state.configuration.accesses).toHaveLength(0)
      expect(result.current.state.isLoading).toBe(false)
    })
  })

  describe('unknown action type', () => {
    it('returns unchanged state for unknown dispatch type', () => {
      const { result } = renderHook(() => useConfiguration(), { wrapper })

      const stateBefore = result.current.state

      act(() => {
        result.current.dispatch({ type: 'UNKNOWN_ACTION' as any })
      })

      expect(result.current.state).toEqual(stateBefore)
    })
  })
})
