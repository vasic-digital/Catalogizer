import axios from 'axios'
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  AuthStatus,
  ChangePasswordRequest,
  UpdateProfileRequest,
  User
} from '@/types/auth'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const authApi = {
  login: (data: LoginRequest): Promise<LoginResponse> =>
    api.post('/auth/login', data).then((res) => res.data),

  register: (data: RegisterRequest): Promise<User> =>
    api.post('/auth/register', data).then((res) => res.data),

  logout: (): Promise<void> =>
    api.post('/auth/logout').then(() => {/* no content */}),

  getProfile: (): Promise<User> =>
    api.get('/auth/profile').then((res) => res.data),

  updateProfile: (data: UpdateProfileRequest): Promise<User> =>
    api.put('/auth/profile', data).then((res) => res.data),

  changePassword: (data: ChangePasswordRequest): Promise<void> =>
    api.post('/auth/change-password', data).then(() => {/* no content */}),

  getAuthStatus: (): Promise<AuthStatus> =>
    api.get('/auth/status').then((res) => res.data),

  getPermissions: (): Promise<{ role: string; permissions: string[]; is_admin: boolean }> =>
    api.get('/auth/permissions').then((res) => res.data),

  getInitStatus: (): Promise<{ initialized: boolean; has_admin: boolean; user_count: number }> =>
    api.get('/auth/init-status').then((res) => res.data),
}

export default api