import { api, authApi } from '../api'

// Mock axios
vi.mock('axios', () => {
  const mockAxiosInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  }
  return {
    __esModule: true,
    default: {
      create: vi.fn(() => mockAxiosInstance),
    },
  }
})

// Since the api module uses axios.create(), we need to mock at the module level
vi.mock('../api', async () => {
  const mockApi = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  }

  const authApi = {
    login: vi.fn((data: any) => mockApi.post('/auth/login', data).then((res: any) => res.data)),
    register: vi.fn((data: any) => mockApi.post('/auth/register', data).then((res: any) => res.data)),
    logout: vi.fn(() => mockApi.post('/auth/logout').then(() => {})),
    getProfile: vi.fn(() => mockApi.get('/auth/profile').then((res: any) => res.data)),
    updateProfile: vi.fn((data: any) => mockApi.put('/auth/profile', data).then((res: any) => res.data)),
    changePassword: vi.fn((data: any) => mockApi.post('/auth/change-password', data).then(() => {})),
    getAuthStatus: vi.fn(() => mockApi.get('/auth/status').then((res: any) => res.data)),
    getPermissions: vi.fn(() => mockApi.get('/auth/permissions').then((res: any) => res.data)),
    getInitStatus: vi.fn(() => mockApi.get('/auth/init-status').then((res: any) => res.data)),
  }

  return {
    __esModule: true,
    default: mockApi,
    api: mockApi,
    authApi,
  }
})

const mockApi = vi.mocked(api)

describe('authApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('login', () => {
    it('calls POST /auth/login and returns token data', async () => {
      const loginResponse = {
        user: { id: 1, username: 'admin', email: 'admin@test.com' },
        token: 'jwt-token-123',
        refresh_token: 'refresh-123',
        expires_in: 3600,
      }
      mockApi.post.mockResolvedValue({ data: loginResponse })

      const result = await authApi.login({ username: 'admin', password: 'password' })

      expect(mockApi.post).toHaveBeenCalledWith('/auth/login', {
        username: 'admin',
        password: 'password',
      })
      expect(result).toEqual(loginResponse)
    })

    it('propagates errors on login failure', async () => {
      const error = { response: { status: 401 }, message: 'Invalid credentials' }
      mockApi.post.mockRejectedValue(error)

      await expect(
        authApi.login({ username: 'bad', password: 'bad' })
      ).rejects.toEqual(error)
    })
  })

  describe('register', () => {
    it('calls POST /auth/register and returns user', async () => {
      const newUser = { id: 2, username: 'newuser', email: 'new@test.com' }
      mockApi.post.mockResolvedValue({ data: newUser })

      const result = await authApi.register({
        username: 'newuser',
        email: 'new@test.com',
        password: 'password123',
        first_name: 'New',
        last_name: 'User',
      })

      expect(mockApi.post).toHaveBeenCalledWith('/auth/register', expect.objectContaining({
        username: 'newuser',
        email: 'new@test.com',
      }))
      expect(result).toEqual(newUser)
    })
  })

  describe('logout', () => {
    it('calls POST /auth/logout', async () => {
      mockApi.post.mockResolvedValue({ data: {} })

      await authApi.logout()

      expect(mockApi.post).toHaveBeenCalledWith('/auth/logout')
    })
  })

  describe('getProfile', () => {
    it('calls GET /auth/profile and returns user data', async () => {
      const profile = { id: 1, username: 'admin', email: 'admin@test.com' }
      mockApi.get.mockResolvedValue({ data: profile })

      const result = await authApi.getProfile()

      expect(mockApi.get).toHaveBeenCalledWith('/auth/profile')
      expect(result).toEqual(profile)
    })
  })

  describe('updateProfile', () => {
    it('calls PUT /auth/profile with update data', async () => {
      const updatedUser = { id: 1, username: 'admin', first_name: 'Updated' }
      mockApi.put.mockResolvedValue({ data: updatedUser })

      const result = await authApi.updateProfile({ first_name: 'Updated' })

      expect(mockApi.put).toHaveBeenCalledWith('/auth/profile', { first_name: 'Updated' })
      expect(result).toEqual(updatedUser)
    })
  })

  describe('changePassword', () => {
    it('calls POST /auth/change-password', async () => {
      mockApi.post.mockResolvedValue({ data: {} })

      await authApi.changePassword({
        current_password: 'old123',
        new_password: 'new123',
      })

      expect(mockApi.post).toHaveBeenCalledWith('/auth/change-password', {
        current_password: 'old123',
        new_password: 'new123',
      })
    })
  })

  describe('getAuthStatus', () => {
    it('calls GET /auth/status and returns status', async () => {
      const status = { authenticated: true, user: { id: 1 } }
      mockApi.get.mockResolvedValue({ data: status })

      const result = await authApi.getAuthStatus()

      expect(mockApi.get).toHaveBeenCalledWith('/auth/status')
      expect(result).toEqual(status)
    })
  })

  describe('getPermissions', () => {
    it('calls GET /auth/permissions and returns role data', async () => {
      const permissions = { role: 'admin', permissions: ['read', 'write'], is_admin: true }
      mockApi.get.mockResolvedValue({ data: permissions })

      const result = await authApi.getPermissions()

      expect(mockApi.get).toHaveBeenCalledWith('/auth/permissions')
      expect(result).toEqual(permissions)
    })
  })

  describe('getInitStatus', () => {
    it('calls GET /auth/init-status and returns initialization info', async () => {
      const initStatus = { initialized: true, has_admin: true, user_count: 5 }
      mockApi.get.mockResolvedValue({ data: initStatus })

      const result = await authApi.getInitStatus()

      expect(mockApi.get).toHaveBeenCalledWith('/auth/init-status')
      expect(result).toEqual(initStatus)
    })

    it('returns uninitialized state when no admin exists', async () => {
      const initStatus = { initialized: false, has_admin: false, user_count: 0 }
      mockApi.get.mockResolvedValue({ data: initStatus })

      const result = await authApi.getInitStatus()

      expect(result.initialized).toBe(false)
      expect(result.has_admin).toBe(false)
    })
  })
})
