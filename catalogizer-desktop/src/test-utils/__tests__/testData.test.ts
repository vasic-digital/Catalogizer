import { describe, it, expect, beforeEach } from 'vitest'
import {
  generateMediaItems,
  generateUsers,
  createSuccessResponse,
  createErrorResponse,
  createLoadingResponse,
  mockTauriApi,
  resetTauriMocks,
  setupTauriSuccessResponse,
  setupTauriErrorResponse,
} from '../testData'

describe('testData generators', () => {
  describe('generateMediaItems', () => {
    it('generates default 10 items', () => {
      const items = generateMediaItems()

      expect(items).toHaveLength(10)
    })

    it('generates specified count of items', () => {
      const items = generateMediaItems(5)

      expect(items).toHaveLength(5)
    })

    it('generates zero items', () => {
      const items = generateMediaItems(0)

      expect(items).toHaveLength(0)
    })

    it('assigns sequential IDs starting from 1', () => {
      const items = generateMediaItems(3)

      expect(items[0].id).toBe(1)
      expect(items[1].id).toBe(2)
      expect(items[2].id).toBe(3)
    })

    it('cycles through media types', () => {
      const items = generateMediaItems(5)
      const types = items.map(i => i.type)

      expect(types).toContain('movie')
      expect(types).toContain('tv_show')
      expect(types).toContain('music_album')
      expect(types).toContain('game')
      expect(types).toContain('book')
    })

    it('generates items with valid titles', () => {
      const items = generateMediaItems(3)

      items.forEach(item => {
        expect(item.title).toBeTruthy()
        expect(typeof item.title).toBe('string')
        expect(item.title.length).toBeGreaterThan(0)
      })
    })

    it('generates items with valid years', () => {
      const items = generateMediaItems(20)

      items.forEach(item => {
        expect(item.year).toBeGreaterThanOrEqual(2010)
        expect(item.year).toBeLessThanOrEqual(2025)
      })
    })

    it('generates items with valid ratings', () => {
      const items = generateMediaItems(10)

      items.forEach(item => {
        expect(item.rating).toBeGreaterThanOrEqual(5.0)
        expect(item.rating).toBeLessThanOrEqual(10.0)
      })
    })

    it('generates items with overviews', () => {
      const items = generateMediaItems(3)

      items.forEach(item => {
        expect(item.overview).toBeTruthy()
        expect(item.overview.length).toBeGreaterThan(10)
      })
    })

    it('generates items with poster and backdrop paths', () => {
      const items = generateMediaItems(3)

      items.forEach(item => {
        expect(item.posterPath).toContain('/posters/')
        expect(item.backdropPath).toContain('/backdrops/')
      })
    })

    it('generates runtime only for movies and tv_shows', () => {
      const items = generateMediaItems(10)

      items.forEach(item => {
        if (item.type === 'movie' || item.type === 'tv_show') {
          expect(item.runtime).toBeDefined()
          expect(item.runtime).toBeGreaterThan(0)
        } else {
          expect(item.runtime).toBeUndefined()
        }
      })
    })

    it('generates items with genres arrays', () => {
      const items = generateMediaItems(5)

      items.forEach(item => {
        expect(Array.isArray(item.genres)).toBe(true)
        expect(item.genres.length).toBe(3)
      })
    })
  })

  describe('generateUsers', () => {
    it('generates default 5 users', () => {
      const users = generateUsers()

      expect(users).toHaveLength(5)
    })

    it('generates specified count of users', () => {
      const users = generateUsers(3)

      expect(users).toHaveLength(3)
    })

    it('assigns sequential IDs', () => {
      const users = generateUsers(3)

      expect(users[0].id).toBe(1)
      expect(users[1].id).toBe(2)
      expect(users[2].id).toBe(3)
    })

    it('generates unique usernames', () => {
      const users = generateUsers(5)
      const usernames = users.map(u => u.username)
      const uniqueUsernames = new Set(usernames)

      expect(uniqueUsernames.size).toBe(5)
    })

    it('generates valid email addresses', () => {
      const users = generateUsers(3)

      users.forEach(user => {
        expect(user.email).toContain('@')
        expect(user.email).toContain('example.com')
      })
    })

    it('generates valid date objects', () => {
      const users = generateUsers(3)

      users.forEach(user => {
        expect(user.createdAt).toBeInstanceOf(Date)
        expect(user.updatedAt).toBeInstanceOf(Date)
        expect(user.createdAt.getTime()).toBeLessThanOrEqual(user.updatedAt.getTime())
      })
    })
  })

  describe('API response factories', () => {
    it('creates a success response with data', () => {
      const data = { id: 1, name: 'test' }
      const response = createSuccessResponse(data)

      expect(response.success).toBe(true)
      expect(response.data).toEqual(data)
      expect(response.error).toBeUndefined()
    })

    it('creates a success response with array data', () => {
      const data = [1, 2, 3]
      const response = createSuccessResponse(data)

      expect(response.success).toBe(true)
      expect(response.data).toEqual([1, 2, 3])
    })

    it('creates an error response with message', () => {
      const response = createErrorResponse('Something went wrong')

      expect(response.success).toBe(false)
      expect(response.error).toBe('Something went wrong')
      expect(response.data).toBeUndefined()
    })

    it('creates a loading response', () => {
      const response = createLoadingResponse()

      expect(response.success).toBe(false)
      expect(response.message).toBe('Loading...')
    })
  })

  describe('Tauri mock utilities', () => {
    beforeEach(() => {
      resetTauriMocks()
    })

    it('provides mock invoke function', () => {
      expect(mockTauriApi.invoke).toBeDefined()
      expect(typeof mockTauriApi.invoke).toBe('function')
    })

    it('provides mock listen function', () => {
      expect(mockTauriApi.listen).toBeDefined()
      expect(typeof mockTauriApi.listen).toBe('function')
    })

    it('provides mock emit function', () => {
      expect(mockTauriApi.emit).toBeDefined()
      expect(typeof mockTauriApi.emit).toBe('function')
    })

    it('resets all mocks', () => {
      mockTauriApi.invoke.mockReturnValue('test')
      mockTauriApi.listen.mockReturnValue('test')
      mockTauriApi.emit.mockReturnValue('test')

      resetTauriMocks()

      expect(mockTauriApi.invoke.mock.calls).toHaveLength(0)
      expect(mockTauriApi.listen.mock.calls).toHaveLength(0)
      expect(mockTauriApi.emit.mock.calls).toHaveLength(0)
    })

    it('sets up success response', async () => {
      const data = { status: 'ok' }
      setupTauriSuccessResponse(data)

      const result = await mockTauriApi.invoke()

      expect(result).toEqual(data)
    })

    it('sets up error response', async () => {
      setupTauriErrorResponse('Failed operation')

      await expect(mockTauriApi.invoke()).rejects.toThrow('Failed operation')
    })
  })
})
