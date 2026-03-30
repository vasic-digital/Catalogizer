import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from '../AuthContext';
import { authApi } from '@/lib/api';

// Mock the API module completely
vi.mock('@/lib/api', async () => ({
  authApi: {
    getAuthStatus: vi.fn(),
    getPermissions: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    updateProfile: vi.fn(),
    changePassword: vi.fn(),
  }
}));

const mockAuthApi = vi.mocked(authApi);

// Mock react-hot-toast
vi.mock('react-hot-toast', async () => ({
  __esModule: true,
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Test component that uses auth context
function TestComponent() {
  const { user, isAuthenticated, isLoading, isAdmin, permissions, hasPermission, canAccess } = useAuth();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <div data-testid="auth-status">
        {isAuthenticated ? 'Authenticated' : 'Not authenticated'}
      </div>
      <div data-testid="user-info">
        {user ? `User: ${user.username}` : 'No user'}
      </div>
      <div data-testid="admin-status">
        {isAdmin ? 'Is Admin' : 'Not Admin'}
      </div>
      <div data-testid="permissions-count">
        {permissions.length}
      </div>
      <div data-testid="has-read-media">
        {hasPermission('read:media') ? 'yes' : 'no'}
      </div>
      <div data-testid="can-access-media-read">
        {canAccess('media', 'read') ? 'yes' : 'no'}
      </div>
    </div>
  );
}

// Helper function to create fresh query client for each test
function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
}

describe('AuthContext', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('provides initial loading state', () => {
    mockAuthApi.getAuthStatus.mockReturnValue(new Promise(() => { /* never resolves */ }));
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('throws error when useAuth is used outside provider', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {/* mock */});
    const queryClient = createQueryClient();

    expect(() => {
      render(
        <QueryClientProvider client={queryClient}>
          <TestComponent />
        </QueryClientProvider>
      );
    }).toThrow('useAuth must be used within an AuthProvider');

    consoleSpy.mockRestore();
  });

  it('shows unauthenticated state when auth status returns no user', async () => {
    mockAuthApi.getAuthStatus.mockResolvedValue({
      authenticated: false,
      user: null,
      permissions: [],
    });
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Not authenticated');
    });
    expect(screen.getByTestId('user-info')).toHaveTextContent('No user');
  });

  it('shows authenticated state when auth status returns a user', async () => {
    mockAuthApi.getAuthStatus.mockResolvedValue({
      authenticated: true,
      user: { id: 1, username: 'testuser', role: { name: 'User' }, role_id: 2 },
      permissions: ['read:media'],
    });
    mockAuthApi.getPermissions.mockResolvedValue({
      permissions: ['read:media', 'write:media'],
    });
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Authenticated');
    });
    expect(screen.getByTestId('user-info')).toHaveTextContent('User: testuser');
  });

  it('detects admin user by role name', async () => {
    mockAuthApi.getAuthStatus.mockResolvedValue({
      authenticated: true,
      user: { id: 1, username: 'admin', role: { name: 'Admin' }, role_id: 1 },
      permissions: ['admin:system'],
    });
    mockAuthApi.getPermissions.mockResolvedValue({
      permissions: ['admin:system'],
    });
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin-status')).toHaveTextContent('Is Admin');
    });
  });

  it('detects admin user by role_id=1', async () => {
    mockAuthApi.getAuthStatus.mockResolvedValue({
      authenticated: true,
      user: { id: 1, username: 'admin', role: { name: 'SomeRole' }, role_id: 1 },
      permissions: [],
    });
    mockAuthApi.getPermissions.mockResolvedValue({ permissions: [] });
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin-status')).toHaveTextContent('Is Admin');
    });
  });

  it('non-admin user is not flagged as admin', async () => {
    mockAuthApi.getAuthStatus.mockResolvedValue({
      authenticated: true,
      user: { id: 2, username: 'viewer', role: { name: 'Viewer' }, role_id: 3 },
      permissions: ['read:media'],
    });
    mockAuthApi.getPermissions.mockResolvedValue({ permissions: ['read:media'] });
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin-status')).toHaveTextContent('Not Admin');
    });
  });

  it('handles auth status API error gracefully', async () => {
    // Use a 401 error so the AuthProvider retry logic bails immediately
    // (non-401 errors trigger up to 2 retries with exponential backoff,
    // which can exceed the default waitFor timeout).
    const error401 = new Error('Unauthorized');
    (error401 as any).response = { status: 401 };
    mockAuthApi.getAuthStatus.mockRejectedValue(error401);
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <TestComponent />
        </AuthProvider>
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Not authenticated');
    });
  });

  it('renders children inside provider', () => {
    mockAuthApi.getAuthStatus.mockReturnValue(new Promise(() => { /* never resolves */ }));
    const queryClient = createQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <div data-testid="child">Child Content</div>
        </AuthProvider>
      </QueryClientProvider>
    );

    expect(screen.getByTestId('child')).toHaveTextContent('Child Content');
  });
});
