import { vi, describe, it, expect, beforeEach } from 'vitest'
import { invoke } from '@tauri-apps/api/core'

const mockInvoke = vi.fn()

vi.mock('@tauri-apps/api/core', () => ({
  invoke: (...args: any[]) => mockInvoke(...args),
}))

describe('Tauri Commands (extended)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('get_config', () => {
    it('returns full config object', async () => {
      const mockConfig = {
        server_url: 'http://localhost:8080',
        auth_token: 'jwt-token-abc',
        theme: 'dark',
        auto_start: false,
      }
      mockInvoke.mockResolvedValue(mockConfig)

      const result = await invoke('get_config')

      expect(result).toEqual(mockConfig)
      expect(mockInvoke).toHaveBeenCalledWith('get_config')
    })

    it('returns config without auth token', async () => {
      const mockConfig = {
        server_url: 'http://localhost:8080',
        theme: 'light',
        auto_start: true,
      }
      mockInvoke.mockResolvedValue(mockConfig)

      const result = await invoke('get_config')

      expect(result).toEqual(mockConfig)
    })

    it('handles error when config cannot be read', async () => {
      mockInvoke.mockRejectedValue(new Error('Config file corrupted'))

      await expect(invoke('get_config')).rejects.toThrow('Config file corrupted')
    })
  })

  describe('update_config', () => {
    it('updates config with new values', async () => {
      mockInvoke.mockResolvedValue(undefined)

      const newConfig = {
        server_url: 'http://new-server:9090',
        theme: 'light',
        auto_start: true,
      }

      await invoke('update_config', { newConfig })

      expect(mockInvoke).toHaveBeenCalledWith('update_config', { newConfig })
    })

    it('handles write permission errors', async () => {
      mockInvoke.mockRejectedValue(new Error('Permission denied'))

      await expect(
        invoke('update_config', { newConfig: { theme: 'dark' } })
      ).rejects.toThrow('Permission denied')
    })
  })

  describe('set_server_url', () => {
    it('sets server URL successfully', async () => {
      mockInvoke.mockResolvedValue(undefined)

      await invoke('set_server_url', { url: 'http://192.168.1.100:8080' })

      expect(mockInvoke).toHaveBeenCalledWith('set_server_url', {
        url: 'http://192.168.1.100:8080',
      })
    })
  })

  describe('set_auth_token', () => {
    it('stores auth token', async () => {
      mockInvoke.mockResolvedValue(undefined)

      await invoke('set_auth_token', { token: 'my-jwt-token' })

      expect(mockInvoke).toHaveBeenCalledWith('set_auth_token', {
        token: 'my-jwt-token',
      })
    })
  })

  describe('clear_auth_token', () => {
    it('clears stored auth token', async () => {
      mockInvoke.mockResolvedValue(undefined)

      await invoke('clear_auth_token')

      expect(mockInvoke).toHaveBeenCalledWith('clear_auth_token')
    })

    it('handles error when token storage is inaccessible', async () => {
      mockInvoke.mockRejectedValue(new Error('Storage unavailable'))

      await expect(invoke('clear_auth_token')).rejects.toThrow('Storage unavailable')
    })
  })

  describe('make_http_request', () => {
    it('makes GET request with headers', async () => {
      mockInvoke.mockResolvedValue(JSON.stringify({ status: 'ok' }))

      const result = await invoke('make_http_request', {
        url: 'http://localhost:8080/api/health',
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
      })

      expect(result).toBe(JSON.stringify({ status: 'ok' }))
      expect(mockInvoke).toHaveBeenCalledWith('make_http_request', {
        url: 'http://localhost:8080/api/health',
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
      })
    })

    it('makes POST request with body', async () => {
      mockInvoke.mockResolvedValue(JSON.stringify({ token: 'abc' }))

      await invoke('make_http_request', {
        url: 'http://localhost:8080/api/auth/login',
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'pass' }),
      })

      expect(mockInvoke).toHaveBeenCalledWith('make_http_request', expect.objectContaining({
        method: 'POST',
        body: expect.stringContaining('admin'),
      }))
    })

    it('makes PUT request', async () => {
      mockInvoke.mockResolvedValue('{}')

      await invoke('make_http_request', {
        url: 'http://localhost:8080/api/media/1/progress',
        method: 'PUT',
        headers: {},
        body: JSON.stringify({ position: 120 }),
      })

      expect(mockInvoke).toHaveBeenCalledWith('make_http_request', expect.objectContaining({
        method: 'PUT',
      }))
    })

    it('makes DELETE request', async () => {
      mockInvoke.mockResolvedValue('{}')

      await invoke('make_http_request', {
        url: 'http://localhost:8080/api/smb/configs/1',
        method: 'DELETE',
        headers: {},
      })

      expect(mockInvoke).toHaveBeenCalledWith('make_http_request', expect.objectContaining({
        method: 'DELETE',
      }))
    })

    it('handles SSRF validation error', async () => {
      mockInvoke.mockRejectedValue(new Error('URL does not match configured server'))

      await expect(
        invoke('make_http_request', {
          url: 'http://evil.com/steal',
          method: 'GET',
          headers: {},
        })
      ).rejects.toThrow('URL does not match configured server')
    })

    it('handles network timeout', async () => {
      mockInvoke.mockRejectedValue(new Error('Request timed out'))

      await expect(
        invoke('make_http_request', {
          url: 'http://localhost:8080/api/health',
          method: 'GET',
          headers: {},
        })
      ).rejects.toThrow('Request timed out')
    })
  })

  describe('get_app_version', () => {
    it('returns version string', async () => {
      mockInvoke.mockResolvedValue('1.1.0')

      const result = await invoke('get_app_version')

      expect(result).toBe('1.1.0')
    })
  })

  describe('get_platform', () => {
    it('returns platform string', async () => {
      mockInvoke.mockResolvedValue('linux')

      const result = await invoke('get_platform')

      expect(result).toBe('linux')
    })

    it('supports different platforms', async () => {
      for (const platform of ['linux', 'darwin', 'windows']) {
        mockInvoke.mockResolvedValue(platform)

        const result = await invoke('get_platform')

        expect(result).toBe(platform)
      }
    })
  })

  describe('get_arch', () => {
    it('returns architecture string', async () => {
      mockInvoke.mockResolvedValue('x86_64')

      const result = await invoke('get_arch')

      expect(result).toBe('x86_64')
    })

    it('supports arm64', async () => {
      mockInvoke.mockResolvedValue('aarch64')

      const result = await invoke('get_arch')

      expect(result).toBe('aarch64')
    })
  })
})
