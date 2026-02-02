import { describe, it, expect, vi, beforeEach } from 'vitest'
import { invoke } from '@tauri-apps/api/core'
import { useConfigStore } from '../configStore'

const mockInvoke = vi.mocked(invoke)

describe('configStore', () => {
  beforeEach(() => {
    // Reset store state between tests
    useConfigStore.setState({
      serverUrl: null,
      theme: 'dark',
      autoStart: false,
      isLoading: false,
    })
    // Reset classList for theme tests
    document.documentElement.classList.remove('dark')
  })

  describe('initial state', () => {
    it('has correct default values', () => {
      const state = useConfigStore.getState()

      expect(state.serverUrl).toBeNull()
      expect(state.theme).toBe('dark')
      expect(state.autoStart).toBe(false)
      expect(state.isLoading).toBe(false)
    })
  })

  describe('loadConfig', () => {
    it('sets isLoading to true at start', async () => {
      let capturedLoading = false

      mockInvoke.mockImplementation(async () => {
        capturedLoading = useConfigStore.getState().isLoading
        return {
          server_url: 'http://localhost:8080',
          theme: 'dark',
          auto_start: false,
        } as any
      })

      await useConfigStore.getState().loadConfig()

      expect(capturedLoading).toBe(true)
    })

    it('sets state from config response', async () => {
      mockInvoke.mockResolvedValue({
        server_url: 'http://myserver:9090',
        theme: 'light',
        auto_start: true,
      } as any)

      await useConfigStore.getState().loadConfig()

      const state = useConfigStore.getState()
      expect(state.serverUrl).toBe('http://myserver:9090')
      expect(state.theme).toBe('light')
      expect(state.autoStart).toBe(true)
      expect(state.isLoading).toBe(false)
    })

    it('defaults serverUrl to null when server_url is empty', async () => {
      mockInvoke.mockResolvedValue({
        server_url: '',
        theme: 'dark',
        auto_start: false,
      } as any)

      await useConfigStore.getState().loadConfig()

      expect(useConfigStore.getState().serverUrl).toBeNull()
    })

    it('defaults theme to dark when not provided', async () => {
      mockInvoke.mockResolvedValue({
        server_url: 'http://localhost:8080',
        auto_start: false,
      } as any)

      await useConfigStore.getState().loadConfig()

      expect(useConfigStore.getState().theme).toBe('dark')
    })

    it('adds dark class to document when theme is dark', async () => {
      mockInvoke.mockResolvedValue({
        server_url: '',
        theme: 'dark',
        auto_start: false,
      } as any)

      await useConfigStore.getState().loadConfig()

      expect(document.documentElement.classList.contains('dark')).toBe(true)
    })

    it('removes dark class from document when theme is light', async () => {
      document.documentElement.classList.add('dark')

      mockInvoke.mockResolvedValue({
        server_url: '',
        theme: 'light',
        auto_start: false,
      } as any)

      await useConfigStore.getState().loadConfig()

      expect(document.documentElement.classList.contains('dark')).toBe(false)
    })

    it('sets isLoading to false on error', async () => {
      mockInvoke.mockRejectedValue(new Error('Storage error'))

      await useConfigStore.getState().loadConfig()

      expect(useConfigStore.getState().isLoading).toBe(false)
    })

    it('does not change other state on error', async () => {
      useConfigStore.setState({ serverUrl: 'http://existing:8080', theme: 'light' })
      mockInvoke.mockRejectedValue(new Error('Storage error'))

      await useConfigStore.getState().loadConfig()

      expect(useConfigStore.getState().serverUrl).toBe('http://existing:8080')
      expect(useConfigStore.getState().theme).toBe('light')
    })
  })

  describe('setServerUrl', () => {
    it('calls invoke with set_server_url command', async () => {
      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().setServerUrl('http://newserver:8080')

      expect(mockInvoke).toHaveBeenCalledWith('set_server_url', { url: 'http://newserver:8080' })
    })

    it('updates the store state on success', async () => {
      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().setServerUrl('http://newserver:8080')

      expect(useConfigStore.getState().serverUrl).toBe('http://newserver:8080')
    })

    it('throws and does not update state on error', async () => {
      mockInvoke.mockRejectedValue(new Error('Permission denied'))

      await expect(
        useConfigStore.getState().setServerUrl('http://fail:8080')
      ).rejects.toThrow('Permission denied')

      expect(useConfigStore.getState().serverUrl).toBeNull()
    })
  })

  describe('setTheme', () => {
    it('reads current config and updates with new theme', async () => {
      const existingConfig = {
        server_url: 'http://localhost:8080',
        theme: 'dark',
        auto_start: false,
      }

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return existingConfig as any
        return undefined as any
      })

      await useConfigStore.getState().setTheme('light')

      expect(mockInvoke).toHaveBeenCalledWith('update_config', {
        newConfig: { ...existingConfig, theme: 'light' },
      })
    })

    it('updates store theme on success', async () => {
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return { theme: 'dark', auto_start: false } as any
        return undefined as any
      })

      await useConfigStore.getState().setTheme('light')

      expect(useConfigStore.getState().theme).toBe('light')
    })

    it('adds dark class when setting dark theme', async () => {
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return { theme: 'light', auto_start: false } as any
        return undefined as any
      })

      await useConfigStore.getState().setTheme('dark')

      expect(document.documentElement.classList.contains('dark')).toBe(true)
    })

    it('removes dark class when setting light theme', async () => {
      document.documentElement.classList.add('dark')

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return { theme: 'dark', auto_start: false } as any
        return undefined as any
      })

      await useConfigStore.getState().setTheme('light')

      expect(document.documentElement.classList.contains('dark')).toBe(false)
    })

    it('throws on error and does not update state', async () => {
      mockInvoke.mockRejectedValue(new Error('Write error'))

      await expect(
        useConfigStore.getState().setTheme('light')
      ).rejects.toThrow('Write error')

      expect(useConfigStore.getState().theme).toBe('dark')
    })
  })

  describe('setAutoStart', () => {
    it('reads current config and updates with new auto_start value', async () => {
      const existingConfig = {
        server_url: 'http://localhost:8080',
        theme: 'dark',
        auto_start: false,
      }

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return existingConfig as any
        return undefined as any
      })

      await useConfigStore.getState().setAutoStart(true)

      expect(mockInvoke).toHaveBeenCalledWith('update_config', {
        newConfig: { ...existingConfig, auto_start: true },
      })
    })

    it('updates store autoStart on success', async () => {
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return { theme: 'dark', auto_start: false } as any
        return undefined as any
      })

      await useConfigStore.getState().setAutoStart(true)

      expect(useConfigStore.getState().autoStart).toBe(true)
    })

    it('throws on error and does not update state', async () => {
      mockInvoke.mockRejectedValue(new Error('Write error'))

      await expect(
        useConfigStore.getState().setAutoStart(true)
      ).rejects.toThrow('Write error')

      expect(useConfigStore.getState().autoStart).toBe(false)
    })
  })

  describe('resetConfig', () => {
    it('sends default config via update_config', async () => {
      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().resetConfig()

      expect(mockInvoke).toHaveBeenCalledWith('update_config', {
        newConfig: {
          server_url: undefined,
          auth_token: undefined,
          theme: 'dark',
          auto_start: false,
        },
      })
    })

    it('resets store state to defaults on success', async () => {
      useConfigStore.setState({
        serverUrl: 'http://custom:9090',
        theme: 'light',
        autoStart: true,
      })

      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().resetConfig()

      const state = useConfigStore.getState()
      expect(state.serverUrl).toBeNull()
      expect(state.theme).toBe('dark')
      expect(state.autoStart).toBe(false)
    })

    it('adds dark class to document on reset', async () => {
      document.documentElement.classList.remove('dark')
      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().resetConfig()

      expect(document.documentElement.classList.contains('dark')).toBe(true)
    })

    it('throws on error and does not reset state', async () => {
      useConfigStore.setState({
        serverUrl: 'http://custom:9090',
        theme: 'light',
        autoStart: true,
      })

      mockInvoke.mockRejectedValue(new Error('Write error'))

      await expect(
        useConfigStore.getState().resetConfig()
      ).rejects.toThrow('Write error')

      const state = useConfigStore.getState()
      expect(state.serverUrl).toBe('http://custom:9090')
      expect(state.theme).toBe('light')
      expect(state.autoStart).toBe(true)
    })
  })
})
