package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_ValidatePassword(t *testing.T) {
	svc := &AuthService{}

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password with 8 characters",
			password: "Abcdefg1!",
			wantErr:  false,
		},
		{
			name:     "valid long password",
			password: "A_very_strong_password_123!",
			wantErr:  false,
		},
		{
			name:     "too short password",
			password: "short",
			wantErr:  true,
			errMsg:   "password must be at least 8 characters long",
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
			errMsg:   "password must be at least 8 characters long",
		},
		{
			name:     "exactly 7 characters",
			password: "1234567",
			wantErr:  true,
			errMsg:   "password must be at least 8 characters long",
		},
		{
			name:     "exactly 8 characters with all requirements",
			password: "Abcdefg1!",
			wantErr:  false,
		},
		{
			name:     "missing uppercase",
			password: "abcdefg1!",
			wantErr:  true,
			errMsg:   "password must contain at least one uppercase letter",
		},
		{
			name:     "missing lowercase",
			password: "ABCDEFG1!",
			wantErr:  true,
			errMsg:   "password must contain at least one lowercase letter",
		},
		{
			name:     "missing digit",
			password: "Abcdefg!!",
			wantErr:  true,
			errMsg:   "password must contain at least one digit",
		},
		{
			name:     "missing special character",
			password: "Abcdefg12",
			wantErr:  true,
			errMsg:   "password must contain at least one special character",
		},
		{
			name:     "too long password",
			password: "Abcdefg1!" + strings.Repeat("a", 120),
			wantErr:  true,
			errMsg:   "password must be at most 128 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePassword(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_HashAndVerifyPassword(t *testing.T) {
	svc := &AuthService{
		jwtSecret: []byte("test-secret"),
	}

	tests := []struct {
		name     string
		password string
		salt     string
	}{
		{
			name:     "basic password",
			password: "testpassword123",
			salt:     "somesalt123",
		},
		{
			name:     "password with special characters",
			password: "P@ssw0rd!#$%",
			salt:     "differentsalt",
		},
		{
			name:     "medium length password",
			password: "medium_password_here!",
			salt:     "longsalt12345678",
		},
		{
			name:     "empty salt",
			password: "password123",
			salt:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := svc.hashPassword(tt.password, tt.salt)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)

			// Verify correct password
			valid := svc.verifyPassword(tt.password, tt.salt, hash)
			assert.True(t, valid, "correct password should verify")

			// Verify wrong password
			invalid := svc.verifyPassword("wrongpassword", tt.salt, hash)
			assert.False(t, invalid, "wrong password should not verify")

			// Verify wrong salt
			invalidSalt := svc.verifyPassword(tt.password, "wrongsalt", hash)
			assert.False(t, invalidSalt, "wrong salt should not verify")
		})
	}
}

func TestAuthService_HashPasswordForUser(t *testing.T) {
	svc := &AuthService{
		jwtSecret: []byte("test-secret"),
	}

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "standard password",
			password: "mypassword123",
		},
		{
			name:     "complex password",
			password: "C0mpl3x!P@ss#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, salt, err := svc.HashPasswordForUser(tt.password)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEmpty(t, salt)

			// Verify the generated hash and salt work
			valid := svc.verifyPassword(tt.password, salt, hash)
			assert.True(t, valid)

			// Each call should generate a different salt
			hash2, salt2, err := svc.HashPasswordForUser(tt.password)
			require.NoError(t, err)
			assert.NotEqual(t, salt, salt2, "salts should be different")
			assert.NotEqual(t, hash, hash2, "hashes should be different due to different salts")
		})
	}
}

func TestAuthService_GenerateAndValidateJWT(t *testing.T) {
	secret := "test-jwt-secret-key-for-testing"
	svc := &AuthService{
		jwtSecret:  []byte(secret),
		jwtExpiry:  24 * time.Hour,
		refreshExp: 7 * 24 * time.Hour,
	}

	tests := []struct {
		name      string
		userID    int
		username  string
		roleID    int
		sessionID int
	}{
		{
			name:      "standard user",
			userID:    1,
			username:  "testuser",
			roleID:    1,
			sessionID: 100,
		},
		{
			name:      "admin user",
			userID:    42,
			username:  "admin",
			roleID:    2,
			sessionID: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				ID:       tt.userID,
				Username: tt.username,
				RoleID:   tt.roleID,
			}

			token, err := svc.generateJWT(user, tt.sessionID)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Validate the token
			claims, err := svc.validateToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.username, claims.Username)
			assert.Equal(t, tt.roleID, claims.RoleID)
			assert.Equal(t, fmt.Sprintf("%d", tt.sessionID), claims.SessionID)
		})
	}
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	svc := &AuthService{
		jwtSecret:  []byte("test-secret"),
		jwtExpiry:  24 * time.Hour,
		refreshExp: 7 * 24 * time.Hour,
	}

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "garbage token",
			token: "not.a.valid.jwt.token",
		},
		{
			name:  "malformed token",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.validateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	svc1 := &AuthService{
		jwtSecret:  []byte("secret-one"),
		jwtExpiry:  24 * time.Hour,
		refreshExp: 7 * 24 * time.Hour,
	}
	svc2 := &AuthService{
		jwtSecret:  []byte("secret-two"),
		jwtExpiry:  24 * time.Hour,
		refreshExp: 7 * 24 * time.Hour,
	}

	user := &models.User{
		ID:       1,
		Username: "testuser",
		RoleID:   1,
	}

	token, err := svc1.generateJWT(user, 1)
	require.NoError(t, err)

	// Should fail with wrong secret
	claims, err := svc2.validateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_GenerateSecureToken(t *testing.T) {
	svc := &AuthService{}

	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "short token",
			length: 8,
		},
		{
			name:   "standard token",
			length: 32,
		},
		{
			name:   "long token",
			length: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := svc.GenerateSecureToken(tt.length)
			require.NoError(t, err)
			assert.Len(t, token, tt.length*2) // hex encoding doubles the length

			// Generate another token and verify they are different
			token2, err := svc.GenerateSecureToken(tt.length)
			require.NoError(t, err)
			assert.NotEqual(t, token, token2, "tokens should be unique")
		})
	}
}

func TestAuthService_HashData(t *testing.T) {
	svc := &AuthService{}

	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple string",
			data: "hello world",
		},
		{
			name: "empty string",
			data: "",
		},
		{
			name: "complex data",
			data: "user:123:password:salt:extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := svc.HashData(tt.data)
			assert.NotEmpty(t, hash)
			assert.Len(t, hash, 64) // SHA-256 produces 64 hex characters

			// Same input should produce same hash
			hash2 := svc.HashData(tt.data)
			assert.Equal(t, hash, hash2, "same data should produce same hash")
		})
	}

	// Different inputs should produce different hashes
	hash1 := svc.HashData("input1")
	hash2 := svc.HashData("input2")
	assert.NotEqual(t, hash1, hash2, "different data should produce different hashes")
}

func TestAuthService_GenerateSessionToken(t *testing.T) {
	svc := &AuthService{}

	token, err := svc.generateSessionToken()
	require.NoError(t, err)
	assert.Len(t, token, 64) // 32 bytes = 64 hex chars

	// Uniqueness check
	token2, err := svc.generateSessionToken()
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestAuthService_GenerateRefreshToken(t *testing.T) {
	svc := &AuthService{}

	token, err := svc.generateRefreshToken()
	require.NoError(t, err)
	assert.Len(t, token, 64) // 32 bytes = 64 hex chars

	token2, err := svc.generateRefreshToken()
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestAuthService_GenerateSalt(t *testing.T) {
	svc := &AuthService{}

	salt, err := svc.generateSalt()
	require.NoError(t, err)
	assert.Len(t, salt, 32) // 16 bytes = 32 hex chars

	salt2, err := svc.generateSalt()
	require.NoError(t, err)
	assert.NotEqual(t, salt, salt2)
}
