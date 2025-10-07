import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AuthProvider, useAuth } from '../AuthContext';

// Mock axios
jest.mock('axios');
import axios from 'axios';

const mockAxios = axios as jest.Mocked<typeof axios>;

// Test component that uses auth context
function TestComponent() {
  const { user, login, logout, isAuthenticated, isLoading } = useAuth();

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
      <button onClick={() => login('test@example.com', 'password')}>
        Login
      </button>
      <button onClick={logout}>Logout</button>
    </div>
  );
}

describe('AuthContext', () => {
  beforeEach(() => {
    // Clear all mocks
    jest.clearAllMocks();
    // Clear localStorage
    localStorage.clear();
  });

  it('provides initial unauthenticated state', () => {
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    expect(screen.getByTestId('auth-status')).toHaveTextContent('Not authenticated');
    expect(screen.getByTestId('user-info')).toHaveTextContent('No user');
  });

  it('handles successful login', async () => {
    const user = userEvent.setup();

    mockAxios.post.mockResolvedValueOnce({
      data: {
        user: { id: 1, username: 'testuser', email: 'test@example.com' },
        token: 'fake-jwt-token',
        refreshToken: 'fake-refresh-token'
      }
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    const loginButton = screen.getByRole('button', { name: /login/i });
    await user.click(loginButton);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Authenticated');
      expect(screen.getByTestId('user-info')).toHaveTextContent('User: testuser');
    });

    expect(mockAxios.post).toHaveBeenCalledWith('/api/auth/login', {
      email: 'test@example.com',
      password: 'password'
    });
  });

  it('handles login failure', async () => {
    const user = userEvent.setup();

    mockAxios.post.mockRejectedValueOnce({
      response: { data: { message: 'Invalid credentials' } }
    });

    // Mock console.error to avoid test output pollution
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    const loginButton = screen.getByRole('button', { name: /login/i });
    await user.click(loginButton);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Not authenticated');
    });

    expect(consoleSpy).toHaveBeenCalled();
    consoleSpy.mockRestore();
  });

  it('handles logout', async () => {
    const user = userEvent.setup();

    // First login
    mockAxios.post.mockResolvedValueOnce({
      data: {
        user: { id: 1, username: 'testuser', email: 'test@example.com' },
        token: 'fake-jwt-token',
        refreshToken: 'fake-refresh-token'
      }
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    // Login first
    const loginButton = screen.getByRole('button', { name: /login/i });
    await user.click(loginButton);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Authenticated');
    });

    // Then logout
    mockAxios.post.mockResolvedValueOnce({});

    const logoutButton = screen.getByRole('button', { name: /logout/i });
    await user.click(logoutButton);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Not authenticated');
      expect(screen.getByTestId('user-info')).toHaveTextContent('No user');
    });

    expect(mockAxios.post).toHaveBeenCalledWith('/api/auth/logout');
  });

  it('loads user from localStorage on mount', () => {
    const storedUser = {
      id: 1,
      username: 'storeduser',
      email: 'stored@example.com'
    };
    const storedToken = 'stored-jwt-token';

    localStorage.setItem('auth_user', JSON.stringify(storedUser));
    localStorage.setItem('auth_token', storedToken);

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    expect(screen.getByTestId('auth-status')).toHaveTextContent('Authenticated');
    expect(screen.getByTestId('user-info')).toHaveTextContent('User: storeduser');
  });

  it('handles token refresh', async () => {
    mockAxios.post.mockResolvedValueOnce({
      data: {
        token: 'new-jwt-token',
        refreshToken: 'new-refresh-token'
      }
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    // Wait for component to mount and potentially refresh token
    await waitFor(() => {
      expect(mockAxios.post).toHaveBeenCalledWith('/api/auth/refresh');
    });
  });

  it('throws error when useAuth is used outside provider', () => {
    // Mock console.error for the error boundary
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      render(<TestComponent />);
    }).toThrow('useAuth must be used within an AuthProvider');

    consoleSpy.mockRestore();
  });

  it('handles network errors gracefully', async () => {
    const user = userEvent.setup();

    mockAxios.post.mockRejectedValueOnce(new Error('Network Error'));

    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    const loginButton = screen.getByRole('button', { name: /login/i });
    await user.click(loginButton);

    await waitFor(() => {
      expect(screen.getByTestId('auth-status')).toHaveTextContent('Not authenticated');
    });

    expect(consoleSpy).toHaveBeenCalled();
    consoleSpy.mockRestore();
  });
});