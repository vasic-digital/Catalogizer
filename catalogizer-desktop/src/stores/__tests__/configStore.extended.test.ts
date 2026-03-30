import { describe, it, expect, vi, beforeEach } from 'vitest'
import { invoke } from '@tauri-apps/api/core'
import { useConfigStore } from '../configStore'

const mockInvoke = vi.mocked(invoke)

describe('configStore (extended edge cases)', () => {
  beforeEach(() => {
    useConfigStore.setState({
      serverUrl: null,
      theme: 'dark',
      autoStart: false,
      isLoading: false,
    })
    document.documentElement.classList.remove('dark')
  })

  describe('loadConfig edge cases', () => {
    it('handles config with undefined theme gracefully', async () => {
      mockInvoke.mockResolvedValue({
        server_url: 'http://localhost:8080',
        theme: undefined,
        auto_start: false,
      } as any)

      await useConfigStore.getState().loadConfig()

      expect(useConfigStore.getState().theme).toBe('dark')
    })

    it('handles config with null auto_start', async () => {
      mockInvoke.mockResolvedValue({
        server_url: 'http://localhost:8080',
        theme: 'dark',
        auto_start: null,
      } as any)

      await useConfigStore.getState().loadConfig()

      expect(useConfigStore.getState().autoStart).toBe(false)
    })

    it('handles system theme value', async () => {
      mockInvoke.mockResolvedValue({
        server_url: 'http://localhost:8080',
        theme: 'system',
        auto_start: false,
      } as any)

      await useConfigStore.getState().loadConfig()

      expect(useConfigStore.getState().theme).toBe('system')
    })

    it('sets isLoading to false even when invoke returns unexpected data', async () => {
      mockInvoke.mockResolvedValue(null as any)

      // This should not throw but handle gracefully
      try {
        await useConfigStore.getState().loadConfig()
      } catch {
        // May throw due to null access
      }

      // isLoading should eventually be false after error handling
      const state = useConfigStore.getState()
      expect(state.isLoading).toBe(false)
    })
  })

  describe('setServerUrl edge cases', () => {
    it('sets URL with HTTPS', async () => {
      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().setServerUrl('https://secure.example.com')

      expect(useConfigStore.getState().serverUrl).toBe('https://secure.example.com')
    })

    it('sets URL with port', async () => {
      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().setServerUrl('http://192.168.0.241:8080')

      expect(useConfigStore.getState().serverUrl).toBe('http://192.168.0.241:8080')
    })

    it('does not update state if invoke fails', async () => {
      useConfigStore.setState({ serverUrl: 'http://original:8080' })
      mockInvoke.mockRejectedValue(new Error('Write failed'))

      await expect(
        useConfigStore.getState().setServerUrl('http://new:9090')
      ).rejects.toThrow('Write failed')

      expect(useConfigStore.getState().serverUrl).toBe('http://original:8080')
    })
  })

  describe('setTheme edge cases', () => {
    it('preserves server URL when changing theme', async () => {
      useConfigStore.setState({ serverUrl: 'http://server:8080' })

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') {
          return { server_url: 'http://server:8080', theme: 'dark', auto_start: false } as any
        }
        return undefined as any
      })

      await useConfigStore.getState().setTheme('light')

      expect(useConfigStore.getState().serverUrl).toBe('http://server:8080')
    })

    it('preserves autoStart when changing theme', async () => {
      useConfigStore.setState({ autoStart: true })

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') {
          return { theme: 'dark', auto_start: true } as any
        }
        return undefined as any
      })

      await useConfigStore.getState().setTheme('light')

      expect(useConfigStore.getState().autoStart).toBe(true)
    })

    it('switches from light to dark correctly', async () => {
      useConfigStore.setState({ theme: 'light' })
      document.documentElement.classList.remove('dark')

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return { theme: 'light', auto_start: false } as any
        return undefined as any
      })

      await useConfigStore.getState().setTheme('dark')

      expect(useConfigStore.getState().theme).toBe('dark')
      expect(document.documentElement.classList.contains('dark')).toBe(true)
    })
  })

  describe('setAutoStart edge cases', () => {
    it('toggles autoStart from true to false', async () => {
      useConfigStore.setState({ autoStart: true })

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return { theme: 'dark', auto_start: true } as any
        return undefined as any
      })

      await useConfigStore.getState().setAutoStart(false)

      expect(useConfigStore.getState().autoStart).toBe(false)
    })

    it('preserves theme when changing autoStart', async () => {
      useConfigStore.setState({ theme: 'light' })

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') return { theme: 'light', auto_start: false } as any
        return undefined as any
      })

      await useConfigStore.getState().setAutoStart(true)

      expect(useConfigStore.getState().theme).toBe('light')
    })

    it('does not update state on failure', async () => {
      useConfigStore.setState({ autoStart: false })
      mockInvoke.mockRejectedValue(new Error('Permission denied'))

      await expect(
        useConfigStore.getState().setAutoStart(true)
      ).rejects.toThrow('Permission denied')

      expect(useConfigStore.getState().autoStart).toBe(false)
    })
  })

  describe('resetConfig edge cases', () => {
    it('resets all fields to defaults', async () => {
      useConfigStore.setState({
        serverUrl: 'http://custom:9090',
        theme: 'light',
        autoStart: true,
        isLoading: true,
      })

      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().resetConfig()

      const state = useConfigStore.getState()
      expect(state.serverUrl).toBeNull()
      expect(state.theme).toBe('dark')
      expect(state.autoStart).toBe(false)
    })

    it('sends correct default config to invoke', async () => {
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

    it('applies dark class to document on reset', async () => {
      document.documentElement.classList.remove('dark')
      mockInvoke.mockResolvedValue(undefined as any)

      await useConfigStore.getState().resetConfig()

      expect(document.documentElement.classList.contains('dark')).toBe(true)
    })

    it('preserves state on reset failure', async () => {
      useConfigStore.setState({
        serverUrl: 'http://keep:8080',
        theme: 'light',
        autoStart: true,
      })

      mockInvoke.mockRejectedValue(new Error('Storage error'))

      await expect(useConfigStore.getState().resetConfig()).rejects.toThrow('Storage error')

      const state = useConfigStore.getState()
      expect(state.serverUrl).toBe('http://keep:8080')
      expect(state.theme).toBe('light')
      expect(state.autoStart).toBe(true)
    })
  })

  describe('concurrent operations', () => {
    it('handles sequential loadConfig calls', async () => {
      let callCount = 0
      mockInvoke.mockImplementation(async () => {
        callCount++
        return {
          server_url: `http://server-${callCount}:8080`,
          theme: 'dark',
          auto_start: false,
        } as any
      })

      await useConfigStore.getState().loadConfig()
      await useConfigStore.getState().loadConfig()

      // The second call's result should win
      expect(useConfigStore.getState().serverUrl).toBe('http://server-2:8080')
    })
  })
})
