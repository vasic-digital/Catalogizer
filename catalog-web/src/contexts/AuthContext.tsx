import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi } from '@/lib/api'
import type { User, LoginRequest, RegisterRequest, ChangePasswordRequest, UpdateProfileRequest } from '@/types/auth'
import toast from 'react-hot-toast'

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  permissions: string[]
  isAdmin: boolean
  login: (data: LoginRequest) => Promise<any>
  register: (data: RegisterRequest) => Promise<any>
  logout: () => Promise<void>
  updateProfile: (data: UpdateProfileRequest) => Promise<any>
  changePassword: (data: ChangePasswordRequest) => Promise<void>
  hasPermission: (permission: string) => boolean
  canAccess: (resource: string, action: string) => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

interface AuthProviderProps {
  children: ReactNode
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [permissions, setPermissions] = useState<string[]>([])
  const queryClient = useQueryClient()

  const { data: authStatus, isLoading } = useQuery({
    queryKey: ['auth-status'],
    queryFn: authApi.getAuthStatus,
    retry: (failureCount, error: any) => {
      if (error?.response?.status === 401) return false
      return failureCount < 2
    },
    staleTime: 1000 * 60 * 5,
  })

  const { data: permissionsData } = useQuery({
    queryKey: ['permissions'],
    queryFn: authApi.getPermissions,
    enabled: !!user,
    staleTime: 1000 * 60 * 10,
  })

  useEffect(() => {
    if (authStatus?.authenticated && authStatus.user) {
      setUser(authStatus.user)
      setPermissions(authStatus.permissions || [])
    } else {
      setUser(null)
      setPermissions([])
    }
  }, [authStatus])

  useEffect(() => {
    if (permissionsData) {
      setPermissions(permissionsData.permissions || [])
    }
  }, [permissionsData])

  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      localStorage.setItem('auth_token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))
      setUser(data.user)
      queryClient.invalidateQueries({ queryKey: ['auth-status'] })
      queryClient.invalidateQueries({ queryKey: ['permissions'] })
      toast.success('Successfully logged in!')
    },
    onError: (error: any) => {
      const message = error?.response?.data?.error || 'Login failed'
      toast.error(message)
    },
  })

  const registerMutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: () => {
      toast.success('Registration successful! Please log in.')
    },
    onError: (error: any) => {
      const message = error?.response?.data?.error || 'Registration failed'
      toast.error(message)
    },
  })

  const logoutMutation = useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user')
      setUser(null)
      setPermissions([])
      queryClient.clear()
      toast.success('Successfully logged out!')
    },
    onError: () => {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user')
      setUser(null)
      setPermissions([])
      queryClient.clear()
    },
  })

  const updateProfileMutation = useMutation({
    mutationFn: authApi.updateProfile,
    onSuccess: (updatedUser) => {
      setUser(updatedUser)
      queryClient.invalidateQueries({ queryKey: ['auth-status'] })
      toast.success('Profile updated successfully!')
    },
    onError: (error: any) => {
      const message = error?.response?.data?.error || 'Profile update failed'
      toast.error(message)
    },
  })

  const changePasswordMutation = useMutation({
    mutationFn: authApi.changePassword,
    onSuccess: () => {
      toast.success('Password changed successfully!')
    },
    onError: (error: any) => {
      const message = error?.response?.data?.error || 'Password change failed'
      toast.error(message)
    },
  })

  const hasPermission = (permission: string): boolean => {
    if (user?.role === 'admin') return true
    return permissions.includes(permission)
  }

  const canAccess = (resource: string, action: string): boolean => {
    const permission = `${action}:${resource}`
    return hasPermission(permission) || hasPermission('admin:system')
  }

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    permissions,
    isAdmin: user?.role === 'admin',
    login: loginMutation.mutateAsync,
    register: registerMutation.mutateAsync,
    logout: logoutMutation.mutateAsync,
    updateProfile: updateProfileMutation.mutateAsync,
    changePassword: changePasswordMutation.mutateAsync,
    hasPermission,
    canAccess,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export default AuthContext