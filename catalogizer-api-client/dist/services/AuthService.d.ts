import { HttpClient } from '../utils/http';
import { LoginRequest, LoginResponse, RegisterRequest, AuthStatus, User, UpdateProfileRequest, ChangePasswordRequest } from '../types';
export declare class AuthService {
    private http;
    constructor(http: HttpClient);
    /**
     * Authenticate user with username and password
     */
    login(credentials: LoginRequest): Promise<LoginResponse>;
    /**
     * Register a new user account
     */
    register(userData: RegisterRequest): Promise<User>;
    /**
     * Logout the current user
     */
    logout(): Promise<void>;
    /**
     * Get current authentication status
     */
    getStatus(): Promise<AuthStatus>;
    /**
     * Refresh the authentication token
     */
    refreshToken(): Promise<{
        token: string;
        expires_in: number;
    }>;
    /**
     * Get current user profile
     */
    getProfile(): Promise<User>;
    /**
     * Update user profile
     */
    updateProfile(updates: UpdateProfileRequest): Promise<User>;
    /**
     * Change user password
     */
    changePassword(passwordData: ChangePasswordRequest): Promise<void>;
    /**
     * Request password reset
     */
    requestPasswordReset(email: string): Promise<void>;
    /**
     * Reset password with token
     */
    resetPassword(token: string, newPassword: string): Promise<void>;
    /**
     * Verify email address
     */
    verifyEmail(token: string): Promise<void>;
    /**
     * Resend email verification
     */
    resendEmailVerification(): Promise<void>;
    /**
     * Check if a username is available
     */
    checkUsernameAvailability(username: string): Promise<{
        available: boolean;
    }>;
    /**
     * Check if an email is available
     */
    checkEmailAvailability(email: string): Promise<{
        available: boolean;
    }>;
    /**
     * Get user permissions
     */
    getPermissions(): Promise<string[]>;
    /**
     * Check if user has specific permission
     */
    hasPermission(permission: string): Promise<boolean>;
    /**
     * Generate API key for the current user
     */
    generateApiKey(name: string): Promise<{
        key: string;
        name: string;
        created_at: string;
    }>;
    /**
     * List user's API keys
     */
    listApiKeys(): Promise<Array<{
        id: number;
        name: string;
        created_at: string;
        last_used?: string;
    }>>;
    /**
     * Revoke an API key
     */
    revokeApiKey(keyId: number): Promise<void>;
}
//# sourceMappingURL=AuthService.d.ts.map