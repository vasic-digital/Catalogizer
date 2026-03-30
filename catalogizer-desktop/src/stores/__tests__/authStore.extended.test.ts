import { describe, it, expect, vi, beforeEach } from 'vitest'
import { invoke } from '@tauri-apps/api/core'
import { useAuthStore } from '../authStore'

const mockInvoke = vi.mocked(invoke)

describe('authStore (extended edge cases)', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      serverUrl: null,
    })
  })

  describe('setServerUrl edge cases', () => {
    it('sets URL with port number', () => {
      useAuthStore.getState().setServerUrl('http://192.168.1.100:9090')

      expect(useAuthStore.getState().serverUrl).toBe('http://192.168.1.100:9090')
    })

    it('sets HTTPS URL', () => {
      useAuthStore.getState().setServerUrl('https://secure-server.example.com')

      expect(useAuthStore.getState().serverUrl).toBe('https://secure-server.example.com')
    })

    it('sets empty string URL', () => {
      useAuthStore.getState().setServerUrl('')

      expect(useAuthStore.getState().serverUrl).toBe('')
    })
  })

  describe('setAuthToken edge cases', () => {
    it('sets isAuthenticated with any non-empty token', () => {
      useAuthStore.getState().setAuthToken('x')

      expect(useAuthStore.getState().isAuthenticated).toBe(true)
    })

    it('does not modify serverUrl', () => {
      useAuthStore.setState({ serverUrl: 'http://test:8080' })

      useAuthStore.getState().setAuthToken('some-token')

      expect(useAuthStore.getState().serverUrl).toBe('http://test:8080')
    })

    it('does not modify isLoading', () => {
      useAuthStore.setState({ isLoading: true })

      useAuthStore.getState().setAuthToken('token')

      expect(useAuthStore.getState().isLoading).toBe(true)
    })
  })

  describe('clearAuth edge cases', () => {
    it('is idempotent when already cleared', () => {
      useAuthStore.getState().clearAuth()
      useAuthStore.getState().clearAuth()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isAuthenticated).toBe(false)
      expect(state.error).toBeNull()
    })

    it('preserves isLoading state', () => {
      useAuthStore.setState({ isLoading: true, isAuthenticated: true })

      useAuthStore.getState().clearAuth()

      expect(useAuthStore.getState().isLoading).toBe(true)
    })
  })

  describe('login edge cases', () => {
    it('rejects when serverUrl is null', async () => {
      useAuthStore.setState({ serverUrl: null })

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'make_http_request') {
          throw new Error('Network error')
        }
        return undefined as any
      })

      await expect(
        useAuthStore.getState().login('admin', 'pass')
      ).rejects.toThrow()

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })

    it('handles JSON parse error from response', async () => {
      useAuthStore.setState({ serverUrl: 'http://localhost:8080' })

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'make_http_request') {
          return 'not-valid-json' as any
        }
        return undefined as any
      })

      await expect(
        useAuthStore.getState().login('admin', 'pass')
      ).rejects.toThrow()

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
      expect(useAuthStore.getState().isLoading).toBe(false)
    })

    it('clears error from previous failed login on new attempt', async () => {
      useAuthStore.setState({
        serverUrl: 'http://localhost:8080',
        error: 'Previous error',
      })

      const mockUser = {
        id: 1, username: 'admin', email: 'a@b.com',
        first_name: 'A', last_name: 'B', is_admin: true,
        created_at: '', updated_at: '',
      }

      let errorDuringRequest: string | null = 'not-cleared'
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'make_http_request') {
          errorDuringRequest = useAuthStore.getState().error
          return JSON.stringify({ token: 'new-token', user: mockUser })
        }
        return undefined as any
      })

      await useAuthStore.getState().login('admin', 'password')

      expect(errorDuringRequest).toBeNull()
    })

    it('handles set_auth_token invoke failure after successful login', async () => {
      useAuthStore.setState({ serverUrl: 'http://localhost:8080' })

      const mockUser = {
        id: 1, username: 'admin', email: 'a@b.com',
        first_name: 'A', last_name: 'B', is_admin: true,
        created_at: '', updated_at: '',
      }

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'make_http_request') {
          return JSON.stringify({ token: 'token', user: mockUser })
        }
        if (cmd === 'set_auth_token') {
          throw new Error('Storage full')
        }
        return undefined as any
      })

      await expect(
        useAuthStore.getState().login('admin', 'pass')
      ).rejects.toThrow('Storage full')

      expect(useAuthStore.getState().isLoading).toBe(false)
    })
  })

  describe('logout edge cases', () => {
    it('clears state even when not currently authenticated', async () => {
      useAuthStore.setState({
        user: null,
        isAuthenticated: false,
        serverUrl: 'http://localhost:8080',
      })

      mockInvoke.mockResolvedValue(undefined as any)

      await useAuthStore.getState().logout()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isAuthenticated).toBe(false)
    })

    it('calls clear_auth_token before the logout endpoint', async () => {
      useAuthStore.setState({
        isAuthenticated: true,
        serverUrl: 'http://localhost:8080',
      })

      const callOrder: string[] = []
      mockInvoke.mockImplementation(async (cmd) => {
        callOrder.push(cmd)
        return undefined as any
      })

      await useAuthStore.getState().logout()

      const clearIndex = callOrder.indexOf('clear_auth_token')
      const httpIndex = callOrder.indexOf('make_http_request')
      expect(clearIndex).toBeLessThan(httpIndex)
    })
  })

  describe('checkAuthStatus edge cases', () => {
    it('does not set authenticated when auth_token is empty string', async () => {
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') {
          return { server_url: 'http://localhost:8080', auth_token: '' }
        }
        return undefined as any
      })

      await useAuthStore.getState().checkAuthStatus()

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })

    it('does not set authenticated when server_url is empty string', async () => {
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') {
          return { server_url: '', auth_token: 'valid-token' }
        }
        return undefined as any
      })

      await useAuthStore.getState().checkAuthStatus()

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })

    it('clears auth token when server returns unauthenticated', async () => {
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') {
          return { server_url: 'http://localhost:8080', auth_token: 'expired' }
        }
        if (cmd === 'make_http_request') {
          return JSON.stringify({ authenticated: false })
        }
        return undefined as any
      })

      await useAuthStore.getState().checkAuthStatus()

      expect(mockInvoke).toHaveBeenCalledWith('clear_auth_token')
    })

    it('handles get_config returning null gracefully', async () => {
      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'get_config') {
          return {} as any
        }
        return undefined as any
      })

      await useAuthStore.getState().checkAuthStatus()

      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    })
  })

  describe('state isolation', () => {
    it('login does not affect serverUrl', async () => {
      useAuthStore.setState({ serverUrl: 'http://myserver:8080' })

      const mockUser = {
        id: 1, username: 'admin', email: 'a@b.com',
        first_name: 'A', last_name: 'B', is_admin: false,
        created_at: '', updated_at: '',
      }

      mockInvoke.mockImplementation(async (cmd) => {
        if (cmd === 'make_http_request') {
          return JSON.stringify({ token: 'tok', user: mockUser })
        }
        return undefined as any
      })

      await useAuthStore.getState().login('user', 'pass')

      expect(useAuthStore.getState().serverUrl).toBe('http://myserver:8080')
    })

    it('multiple rapid setServerUrl calls keep the last value', () => {
      useAuthStore.getState().setServerUrl('http://first:8080')
      useAuthStore.getState().setServerUrl('http://second:8080')
      useAuthStore.getState().setServerUrl('http://third:8080')

      expect(useAuthStore.getState().serverUrl).toBe('http://third:8080')
    })
  })
})
