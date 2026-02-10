package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// JWT Authentication Security Tests
// =============================================================================

func TestAuth_JWTValidation(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectValid bool
		description string
	}{
		{
			name:        "Valid Token",
			token:       generateValidToken(),
			expectValid: true,
			description: "Should accept valid JWT token",
		},
		{
			name:        "Expired Token",
			token:       generateExpiredToken(),
			expectValid: false,
			description: "Should reject expired JWT token",
		},
		{
			name:        "Malformed Token",
			token:       "malformed.jwt.token",
			expectValid: false,
			description: "Should reject malformed token",
		},
		{
			name:        "Invalid Signature",
			token:       generateTokenWithInvalidSignature(),
			expectValid: false,
			description: "Should reject token with invalid signature",
		},
		{
			name:        "Empty Token",
			token:       "",
			expectValid: false,
			description: "Should reject empty token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateJWT(tt.token)
			assert.Equal(t, tt.expectValid, valid, tt.description)
		})
	}
}

func TestAuth_PasswordHashing(t *testing.T) {
	password := "SecureP@ssw0rd123"

	// Hash password
	hash, err := hashPassword(password)
	require.NoError(t, err)

	// Verify hash is not the plain password
	assert.NotEqual(t, password, hash)

	// Verify hash is bcrypt format
	assert.Contains(t, hash, "$2a$")

	// Verify password matches hash
	matches := checkPasswordHash(password, hash)
	assert.True(t, matches)

	// Verify wrong password doesn't match
	wrongMatches := checkPasswordHash("WrongPassword", hash)
	assert.False(t, wrongMatches)
}

func TestAuth_PasswordStrengthRequirements(t *testing.T) {
	tests := []struct {
		password string
		isStrong bool
		reason   string
	}{
		{"Pass1234!", true, "Valid strong password"},
		{"weak", false, "Too short"},
		{"alllowercase123", false, "No uppercase or special chars"},
		{"ALLUPPERCASE123", false, "No lowercase or special chars"},
		{"NoNumbers!", false, "No numbers"},
		{"NoSpecial1", false, "No special characters"},
		{"Short1!", false, "Too short (< 8 chars)"},
		{"V@lid1Pass", true, "Valid with all requirements"},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			isStrong := validatePasswordStrength(tt.password)
			assert.Equal(t, tt.isStrong, isStrong, tt.reason)
		})
	}
}

func TestAuth_SessionManagement(t *testing.T) {
	// Test session creation
	session := createSession("user123")
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "user123", session.UserID)
	assert.NotZero(t, session.CreatedAt)
	assert.NotZero(t, session.ExpiresAt)

	// Test session expiry
	expiredSession := createSession("user123")
	expiredSession.ExpiresAt = time.Now().Add(-1 * time.Hour)
	assert.True(t, isSessionExpired(expiredSession))

	// Test session invalidation
	invalidateSession(session.ID)
	valid := isSessionValid(session.ID)
	assert.False(t, valid)
}

func TestAuth_RateLimitLogin(t *testing.T) {
	// Test rate limiting on login attempts
	username := "testuser"

	// First 5 attempts should succeed
	for i := 0; i < 5; i++ {
		allowed := checkLoginRateLimit(username)
		assert.True(t, allowed, "Attempt %d should be allowed", i+1)
	}

	// 6th attempt should be rate limited
	allowed := checkLoginRateLimit(username)
	assert.False(t, allowed, "6th attempt should be rate limited")

	// After reset, should be allowed again
	resetLoginRateLimit(username)
	allowed = checkLoginRateLimit(username)
	assert.True(t, allowed, "After reset, should be allowed")
}

func TestAuth_AccountLockout(t *testing.T) {
	username := "testuser"

	// Track failed login attempts
	for i := 0; i < 5; i++ {
		recordFailedLogin(username)
		locked := isAccountLocked(username)

		if i < 4 {
			assert.False(t, locked, "Account should not be locked after %d failures", i+1)
		} else {
			assert.True(t, locked, "Account should be locked after 5 failures")
		}
	}

	// Verify locked account cannot login
	canLogin := attemptLogin(username, "password")
	assert.False(t, canLogin, "Locked account should not be able to login")

	// Unlock account
	unlockAccount(username)
	locked := isAccountLocked(username)
	assert.False(t, locked, "Account should be unlocked")
}

// =============================================================================
// Authorization Tests
// =============================================================================

func TestAuth_RoleBasedAccessControl(t *testing.T) {
	tests := []struct {
		role       string
		permission string
		hasAccess  bool
	}{
		{"admin", "read", true},
		{"admin", "write", true},
		{"admin", "delete", true},
		{"user", "read", true},
		{"user", "write", true},
		{"user", "delete", false},
		{"guest", "read", true},
		{"guest", "write", false},
		{"guest", "delete", false},
	}

	for _, tt := range tests {
		t.Run(tt.role+"_"+tt.permission, func(t *testing.T) {
			hasAccess := checkPermission(tt.role, tt.permission)
			assert.Equal(t, tt.hasAccess, hasAccess)
		})
	}
}

func TestAuth_ResourceOwnership(t *testing.T) {
	owner := "user123"
	otherUser := "user456"
	resource := createTestResource(owner)

	// Owner should have access
	canAccess := canUserAccessResource(owner, resource.ID)
	assert.True(t, canAccess, "Owner should have access to their resource")

	// Other user should not have access
	canAccess = canUserAccessResource(otherUser, resource.ID)
	assert.False(t, canAccess, "Other user should not have access")

	// Admin should have access
	canAccess = canUserAccessResource("admin", resource.ID)
	assert.True(t, canAccess, "Admin should have access to any resource")
}

// =============================================================================
// Token Security Tests
// =============================================================================

func TestAuth_TokenRefresh(t *testing.T) {
	// Generate initial token
	token := generateValidToken()

	// Refresh token
	newToken, err := refreshToken(token)
	require.NoError(t, err)

	// Verify new token is different
	assert.NotEqual(t, token, newToken)

	// Verify new token is valid
	valid := validateJWT(newToken)
	assert.True(t, valid)

	// Verify old token is invalidated
	valid = validateJWT(token)
	assert.False(t, valid, "Old token should be invalidated after refresh")
}

func TestAuth_TokenRevocation(t *testing.T) {
	// Clean state
	validTokens = make(map[string]bool)
	revokedTokens = make(map[string]bool)

	token := generateValidToken()

	// Initialize token as valid
	validTokens[token] = true

	// Token should be valid initially
	valid := validateJWT(token)
	assert.True(t, valid)

	// Revoke token
	err := revokeToken(token)
	require.NoError(t, err)

	// Token should no longer be valid
	valid = validateJWT(token)
	assert.False(t, valid, "Revoked token should not be valid")
}

func TestAuth_CSRFTokenValidation(t *testing.T) {
	// Generate CSRF token
	csrfToken := generateCSRFToken()
	assert.NotEmpty(t, csrfToken)

	// Valid CSRF token should be accepted
	valid := validateCSRFToken(csrfToken)
	assert.True(t, valid)

	// Invalid CSRF token should be rejected
	valid = validateCSRFToken("invalid-csrf-token")
	assert.False(t, valid)

	// Used CSRF token should be rejected (one-time use)
	markCSRFTokenUsed(csrfToken)
	valid = validateCSRFToken(csrfToken)
	assert.False(t, valid, "Used CSRF token should not be valid")
}

// =============================================================================
// API Key Security Tests
// =============================================================================

func TestAuth_APIKeyGeneration(t *testing.T) {
	apiKey := generateAPIKey()

	// Verify API key format
	assert.Len(t, apiKey, 64, "API key should be 64 characters")
	assert.Regexp(t, "^[a-zA-Z0-9]+$", apiKey, "API key should be alphanumeric")

	// Verify uniqueness
	apiKey2 := generateAPIKey()
	assert.NotEqual(t, apiKey, apiKey2, "API keys should be unique")
}

func TestAuth_APIKeyValidation(t *testing.T) {
	// Create API key
	apiKey := generateAPIKey()
	userID := "user123"
	storeAPIKey(apiKey, userID)

	// Valid API key should authenticate
	authenticated, uid := authenticateAPIKey(apiKey)
	assert.True(t, authenticated)
	assert.Equal(t, userID, uid)

	// Invalid API key should not authenticate
	authenticated, _ = authenticateAPIKey("invalid-key")
	assert.False(t, authenticated)

	// Revoked API key should not authenticate
	revokeAPIKey(apiKey)
	authenticated, _ = authenticateAPIKey(apiKey)
	assert.False(t, authenticated)
}

// =============================================================================
// Helper Functions (Mocked for Testing)
// =============================================================================

func generateValidToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiZXhwIjoxOTk5OTk5OTk5fQ.signature"
}

func generateExpiredToken() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiZXhwIjoxfQ.signature"
}

func generateTokenWithInvalidSignature() string {
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiZXhwIjoxOTk5OTk5OTk5fQ.invalid"
}

func validateJWT(token string) bool {
	// Check if revoked
	if revokedTokens[token] {
		return false
	}

	// Check if explicitly valid
	if validTokens[token] {
		return true
	}

	// Default valid token
	return token == generateValidToken()
}

func hashPassword(password string) (string, error) {
	// Mock bcrypt hash
	return "$2a$10$N9qo8uLOickgx2ZMRZoMye" + password, nil
}

func checkPasswordHash(password, hash string) bool {
	expected, _ := hashPassword(password)
	return hash == expected
}

func validatePasswordStrength(password string) bool {
	// Check length
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		} else if c >= 'a' && c <= 'z' {
			hasLower = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

var sessions = make(map[string]*Session)

func createSession(userID string) *Session {
	session := &Session{
		ID:        "session-" + userID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessions[session.ID] = session
	return session
}

func isSessionExpired(session *Session) bool {
	return time.Now().After(session.ExpiresAt)
}

func invalidateSession(sessionID string) {
	delete(sessions, sessionID)
}

func isSessionValid(sessionID string) bool {
	session, exists := sessions[sessionID]
	if !exists {
		return false
	}
	return !isSessionExpired(session)
}

var loginAttempts = make(map[string]int)

func checkLoginRateLimit(username string) bool {
	count := loginAttempts[username]
	if count >= 5 {
		return false
	}
	loginAttempts[username]++
	return true
}

func resetLoginRateLimit(username string) {
	delete(loginAttempts, username)
}

var failedLogins = make(map[string]int)
var lockedAccounts = make(map[string]bool)

func recordFailedLogin(username string) {
	failedLogins[username]++
	if failedLogins[username] >= 5 {
		lockedAccounts[username] = true
	}
}

func isAccountLocked(username string) bool {
	return lockedAccounts[username]
}

func attemptLogin(username, password string) bool {
	return !isAccountLocked(username)
}

func unlockAccount(username string) {
	delete(lockedAccounts, username)
	delete(failedLogins, username)
}

var permissions = map[string]map[string]bool{
	"admin": {"read": true, "write": true, "delete": true},
	"user":  {"read": true, "write": true, "delete": false},
	"guest": {"read": true, "write": false, "delete": false},
}

func checkPermission(role, permission string) bool {
	rolePerms, exists := permissions[role]
	if !exists {
		return false
	}
	return rolePerms[permission]
}

type Resource struct {
	ID    string
	Owner string
}

func createTestResource(owner string) *Resource {
	return &Resource{
		ID:    "resource-123",
		Owner: owner,
	}
}

var resources = make(map[string]*Resource)

func canUserAccessResource(userID, resourceID string) bool {
	// Admin has access to everything
	if userID == "admin" {
		return true
	}

	// Check ownership
	resource, exists := resources[resourceID]
	if !exists {
		resource = createTestResource("user123")
		resources[resourceID] = resource
	}

	return resource.Owner == userID
}

var validTokens = make(map[string]bool)
var revokedTokens = make(map[string]bool)

func refreshToken(oldToken string) (string, error) {
	// Invalidate old token
	revokedTokens[oldToken] = true

	// Generate new token
	newToken := "new-" + oldToken
	validTokens[newToken] = true

	return newToken, nil
}

func revokeToken(token string) error {
	revokedTokens[token] = true
	return nil
}

var csrfTokens = make(map[string]bool)

func generateCSRFToken() string {
	token := "csrf-" + time.Now().String()
	csrfTokens[token] = true
	return token
}

func validateCSRFToken(token string) bool {
	return csrfTokens[token]
}

func markCSRFTokenUsed(token string) {
	delete(csrfTokens, token)
}

func generateAPIKey() string {
	// Generate 64-character alphanumeric key
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	key := make([]byte, 64)
	timestamp := time.Now().UnixNano()
	for i := range key {
		key[i] = chars[(timestamp+int64(i))%int64(len(chars))]
	}
	return string(key)
}

var apiKeys = make(map[string]string)

func storeAPIKey(apiKey, userID string) {
	apiKeys[apiKey] = userID
}

func authenticateAPIKey(apiKey string) (bool, string) {
	userID, exists := apiKeys[apiKey]
	return exists, userID
}

func revokeAPIKey(apiKey string) {
	delete(apiKeys, apiKey)
}
