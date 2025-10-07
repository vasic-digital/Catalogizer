import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { act } from 'react'
import { ConfigurationProvider, useConfiguration } from '../ConfigurationContext'
import { Configuration, ConfigurationAccess, ConfigurationSource, NetworkHost } from '../../types'

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <ConfigurationProvider>{children}</ConfigurationProvider>
)

describe('ConfigurationContext', () => {
  it('initializes with correct default state', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    expect(result.current.state).toEqual({
      configuration: {
        accesses: [],
        sources: [],
      },
      currentSMBConfig: null,
      currentFTPConfig: null,
      currentNFSConfig: null,
      currentWebDAVConfig: null,
      currentLocalConfig: null,
      selectedProtocol: null,
      selectedHosts: [],
      isLoading: false,
      error: null,
      hasUnsavedChanges: false,
    })
  })

  it('sets configuration', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    const newConfig: Configuration = {
      accesses: [
        {
          name: 'test_user',
          type: 'credentials',
          account: 'username',
          secret: 'password',
        },
      ],
      sources: [
        {
          type: 'samba',
          url: 'smb://192.168.1.100/share',
          access: 'test_user',
        },
      ],
    }

    act(() => {
      result.current.setConfiguration(newConfig)
    })

    expect(result.current.state.configuration).toEqual(newConfig)
    expect(result.current.state.hasUnsavedChanges).toBe(false)
  })

  it('adds access credential', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    const newAccess: ConfigurationAccess = {
      name: 'new_user',
      type: 'credentials',
      account: 'new_username',
      secret: 'new_password',
    }

    act(() => {
      result.current.addAccess(newAccess)
    })

    expect(result.current.state.configuration.accesses).toContain(newAccess)
    expect(result.current.state.hasUnsavedChanges).toBe(true)
  })

  it('updates access credential', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    // Add initial access
    const initialAccess: ConfigurationAccess = {
      name: 'user1',
      type: 'credentials',
      account: 'username1',
      secret: 'password1',
    }

    act(() => {
      result.current.addAccess(initialAccess)
    })

    // Update the access
    const updatedAccess: ConfigurationAccess = {
      name: 'updated_user',
      type: 'credentials',
      account: 'updated_username',
      secret: 'updated_password',
    }

    act(() => {
      result.current.updateAccess(0, updatedAccess)
    })

    expect(result.current.state.configuration.accesses[0]).toEqual(updatedAccess)
    expect(result.current.state.hasUnsavedChanges).toBe(true)
  })

  it('removes access credential', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    // Add two access credentials
    const access1: ConfigurationAccess = {
      name: 'user1',
      type: 'credentials',
      account: 'username1',
      secret: 'password1',
    }

    const access2: ConfigurationAccess = {
      name: 'user2',
      type: 'credentials',
      account: 'username2',
      secret: 'password2',
    }

    act(() => {
      result.current.addAccess(access1)
      result.current.addAccess(access2)
    })

    expect(result.current.state.configuration.accesses).toHaveLength(2)

    // Remove first access
    act(() => {
      result.current.removeAccess(0)
    })

    expect(result.current.state.configuration.accesses).toHaveLength(1)
    expect(result.current.state.configuration.accesses[0]).toEqual(access2)
  })

  it('adds source', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    const newSource: ConfigurationSource = {
      type: 'samba',
      url: 'smb://192.168.1.100/share',
      access: 'test_user',
    }

    act(() => {
      result.current.addSource(newSource)
    })

    expect(result.current.state.configuration.sources).toContain(newSource)
    expect(result.current.state.hasUnsavedChanges).toBe(true)
  })

  it('updates source', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    // Add initial source
    const initialSource: ConfigurationSource = {
      type: 'samba',
      url: 'smb://192.168.1.100/share',
      access: 'user1',
    }

    act(() => {
      result.current.addSource(initialSource)
    })

    // Update the source
    const updatedSource: ConfigurationSource = {
      type: 'samba',
      url: 'smb://192.168.1.200/media',
      access: 'user2',
    }

    act(() => {
      result.current.updateSource(0, updatedSource)
    })

    expect(result.current.state.configuration.sources[0]).toEqual(updatedSource)
    expect(result.current.state.hasUnsavedChanges).toBe(true)
  })

  it('removes source', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    // Add two sources
    const source1: ConfigurationSource = {
      type: 'samba',
      url: 'smb://192.168.1.100/share',
      access: 'user1',
    }

    const source2: ConfigurationSource = {
      type: 'samba',
      url: 'smb://192.168.1.200/media',
      access: 'user2',
    }

    act(() => {
      result.current.addSource(source1)
      result.current.addSource(source2)
    })

    expect(result.current.state.configuration.sources).toHaveLength(2)

    // Remove first source
    act(() => {
      result.current.removeSource(0)
    })

    expect(result.current.state.configuration.sources).toHaveLength(1)
    expect(result.current.state.configuration.sources[0]).toEqual(source2)
  })

  it('sets selected hosts', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    const hosts: NetworkHost[] = [
      { ip: '192.168.1.100', hostname: 'host1', mac_address: '', vendor: '', open_ports: [445], smb_shares: [] },
      { ip: '192.168.1.200', hostname: 'host2', mac_address: '', vendor: '', open_ports: [445], smb_shares: [] }
    ]

    act(() => {
      result.current.setSelectedHosts(hosts)
    })

    expect(result.current.state.selectedHosts).toEqual(hosts)
  })

  it('manages loading state', () => {
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

  it('manages error state', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    const errorMessage = 'Test error'

    act(() => {
      result.current.setError(errorMessage)
    })

    expect(result.current.state.error).toBe(errorMessage)

    act(() => {
      result.current.clearError()
    })

    expect(result.current.state.error).toBe(null)
  })

  it('resets to initial state', () => {
    const { result } = renderHook(() => useConfiguration(), { wrapper })

    // Modify state
    act(() => {
      result.current.addAccess({
        name: 'test',
        type: 'credentials',
        account: 'test',
        secret: 'test',
      })
      result.current.setSelectedHosts([{ ip: '192.168.1.100', hostname: '', mac_address: '', vendor: '', open_ports: [], smb_shares: [] }])
      result.current.setLoading(true)
      result.current.setError('Test error')
    })

    // Reset
    act(() => {
      result.current.reset()
    })

    expect(result.current.state).toEqual({
      configuration: {
        accesses: [],
        sources: [],
      },
      currentSMBConfig: null,
      currentFTPConfig: null,
      currentNFSConfig: null,
      currentWebDAVConfig: null,
      currentLocalConfig: null,
      selectedProtocol: null,
      selectedHosts: [],
      isLoading: false,
      error: null,
      hasUnsavedChanges: false,
    })
  })
})